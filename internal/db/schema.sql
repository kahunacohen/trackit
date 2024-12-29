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
	account_id INTEGER NOT NULL,
	category_id INTEGER,
	counter_party TEXT NOT NULL,
	amount REAL NOT NULL,
	deposit REAL,
	withdrawl REAL,
    ignore_when_summing INTEGER NOT NULL DEFAULT 0 CHECK (ignore_when_summing IN (0, 1)),
	"date" DATETIME NOT NULL,
	FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
	FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL
);

INSERT INTO categories (name) VALUES
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
('Professional Development'),
('Savings'),
('Special Occasions'),
('Sports & Fitness'),
('Subscriptions'),
('Taxes'),
('Transportation'),
('Travel'),
('Utilities');

CREATE VIEW transactions_view AS
SELECT 
    accounts.id AS account_id,
    accounts.name AS account_name, 
    transactions.id AS transaction_id, 
	transactions.date AS date, 
    transactions.counter_party AS counter_party, 
    transactions.amount AS amount,
    transactions.ignore_when_summing as ignore_when_summing,
    categories.name AS category_name
FROM 
    transactions
JOIN 
    accounts ON transactions.account_id = accounts.id
LEFT JOIN 
    categories ON transactions.category_id = categories.id;