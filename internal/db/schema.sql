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
	id TEXT PRIMARY KEY,
	account_id INTEGER NOT NULL,
	category_id INTEGER,
	counter_party TEXT NOT NULL,
	amount REAL NOT NULL,
	deposit REAL,
	withdrawl REAL,
	date DATETIME NOT NULL,
	FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
	FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL
);

CREATE VIEW transactions_view AS
SELECT 
    accounts.id AS account_id,
    accounts.name AS account_name, 
    transactions.id AS transaction_id, 
	transactions.date AS date, 
    transactions.counter_party AS counter_party, 
    transactions.amount AS amount, 
    categories.name AS category_name
FROM 
    transactions
JOIN 
    accounts ON transactions.account_id = accounts.id
LEFT JOIN 
    categories ON transactions.category_id = categories.id;