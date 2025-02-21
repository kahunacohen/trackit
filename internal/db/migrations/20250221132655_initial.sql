-- Create "settings" table
CREATE TABLE `settings` (
 `name` text NOT NULL,
 `value` text NOT NULL
);
-- Create index "settings_name" to table: "settings"
CREATE UNIQUE INDEX `settings_name` ON `settings` (`name`);
-- Create "files" table
CREATE TABLE `files` (
 `id` integer NULL PRIMARY KEY AUTOINCREMENT,
 `hash` text NOT NULL,
 `name` text NOT NULL
);
-- Create index "files_hash" to table: "files"
CREATE UNIQUE INDEX `files_hash` ON `files` (`hash`);
-- Create index "files_name" to table: "files"
CREATE UNIQUE INDEX `files_name` ON `files` (`name`);
-- Create "accounts" table
CREATE TABLE `accounts` (
 `id` integer NULL PRIMARY KEY AUTOINCREMENT,
 `name` text NOT NULL,
 `currency` text NOT NULL
);
-- Create index "accounts_name" to table: "accounts"
CREATE UNIQUE INDEX `accounts_name` ON `accounts` (`name`);
-- Create "transactions" table
CREATE TABLE `transactions` (
 `id` integer NULL,
 `account_id` integer NULL,
 `category_id` integer NULL,
 `counter_party` text NOT NULL,
 `description` text NULL,
 `amount` real NOT NULL,
 `deposit` real NULL,
 `withdrawl` real NULL,
 `ignore_when_summing` integer NOT NULL DEFAULT 0,
 `date` text NOT NULL,
 PRIMARY KEY (`id`),
 CONSTRAINT `0` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
 CONSTRAINT `1` FOREIGN KEY (`account_id`) REFERENCES `accounts` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
 CHECK (ignore_when_summing IN (0, 1))
);
-- Create "categories" table
CREATE TABLE `categories` (
 `id` integer NULL PRIMARY KEY AUTOINCREMENT,
 `name` text NOT NULL
);
-- Create index "categories_name" to table: "categories"
CREATE UNIQUE INDEX `categories_name` ON `categories` (`name`);
-- Create "currency_codes" table
CREATE TABLE `currency_codes` (
 `id` integer NULL,
 `symbol` text NOT NULL,
 PRIMARY KEY (`id`),
 CHECK (LENGTH(symbol) = 3)
);
-- Create index "currency_codes_symbol" to table: "currency_codes"
CREATE UNIQUE INDEX `currency_codes_symbol` ON `currency_codes` (`symbol`);
-- Create "rates" table
CREATE TABLE `rates` (
 `id` integer NULL,
 `rate` numeric NOT NULL,
 `currency_code_from_id` integer NOT NULL,
 `month` text NOT NULL,
 PRIMARY KEY (`id`),
 CONSTRAINT `0` FOREIGN KEY (`currency_code_from_id`) REFERENCES `currency_codes` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
 CHECK (month LIKE '____-__' AND substr(month, 1, 4) BETWEEN '0000' AND '9999' AND substr(month, 6, 2) BETWEEN '01' AND '12')
);
