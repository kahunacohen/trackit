-- IMPORTANT: you must copy this file to cmd directory so that it can be
-- embedded in the go binary in order for the init command to also use it.
-- The Makefile does this for you, but if you change the schema.sql file in development,
-- you must manually copy it.
CREATE TABLE IF NOT EXISTS settings (
    "name" TEXT NOT NULL,
    value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    hash TEXT NOT NULL,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS accounts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    currency TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS transactions (
	id INTEGER PRIMARY KEY,
	account_id INTEGER,
	category_id INTEGER,
	counter_party TEXT NOT NULL,
    "description" TEXT,
	amount REAL NOT NULL,
	deposit REAL,
	withdrawl REAL,
    ignore_when_summing INTEGER NOT NULL DEFAULT 0 CHECK (ignore_when_summing IN (0, 1)),
	"date" TEXT NOT NULL,
	FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
	FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
	"name" TEXT UNIQUE NOT NULL
);

INSERT OR IGNORE INTO categories (name) VALUES
('Business Expenses'),
('Childcare'),
('Clothing'),
('Debt Payments'),
('Dining Out'),
('Donations'),
('Education'),
('Entertainment'),
('Gifts'),
('Groceries'),
('Healthcare'),
('Hobbies'),
('Home Improvement'),
('Household Supplies'),
('Insurance'),
('Investments'),
('Miscellaneous'),
('Mortgage/Rent'),
('Personal Care'),
('Pet Care'),
('Professional Development'),
('Savings'),
('Special Occasions'),
('Sports & Fitness'),
('Subscriptions'),
('Taxes'),
('Transportation'),
('Travel'),
('Utilities');

-- Drop the view if it exists because sqlite does not allow
-- the IF EXISTS with view creation.
DROP VIEW IF EXISTS transactions_view;

CREATE VIEW transactions_view AS
SELECT 
    accounts.id AS account_id,
    accounts.name AS account_name, 
    transactions.id AS transaction_id, 
	transactions.date AS date, 
    transactions.counter_party AS counter_party, 
    transactions.amount AS amount,
    transactions.ignore_when_summing as ignore_when_summing,
    transactions.description AS "description",
    categories.name AS category_name
FROM 
    transactions
LEFT JOIN 
    accounts ON transactions.account_id = accounts.id
LEFT JOIN 
    categories ON transactions.category_id = categories.id;

CREATE TABLE IF NOT EXISTS currency_codes (
    id INTEGER PRIMARY KEY,
    symbol TEXT NOT NULL UNIQUE,
    CHECK (LENGTH(symbol) = 3)
);

INSERT OR IGNORE INTO currency_codes (symbol) VALUES
    ('AUD'),
    ('CAD'),
    ('CHF'),
    ('CNY'),
    ('EUR'),
    ('GBP'),
    ('HKD'),
    ('ILS'),
    ('JPY'),
    ('NZD'),
    ('USD');


CREATE TABLE IF NOT EXISTS rates (
    id INTEGER PRIMARY KEY,
    rate NUMERIC NOT NULL,
    currency_code_from_id INTEGER NOT NULL,
    "month" TEXT NOT NULL,
    FOREIGN KEY (currency_code_from_id) REFERENCES currency_codes(id) ON DELETE CASCADE,
    CHECK (month LIKE '____-__' AND substr(month, 1, 4) BETWEEN '0000' AND '9999' AND substr(month, 6, 2) BETWEEN '01' AND '12')
);
INSERT OR IGNORE INTO rates (rate, currency_code_from_id, "month") VALUES
    (3.75, 11, '2024-09'),
    (3.75, 11, '2024-10'),
    (3.75, 11, '2024-11');