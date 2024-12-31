-- name: CreateTransaction :exec
INSERT INTO transactions (account_id, date, amount, counter_party, category_id) VALUES (?, ?, ?, ?, ?);

-- name: ReadNonCategorizedTransactions :many
SELECT transaction_id AS id, "date", counter_party, account_name, amount FROM transactions_view WHERE category_name IS NULL;

-- name: UpdateTransactionCategory :exec
UPDATE transactions SET category_id=? WHERE id=?;

-- name: ReadAllTransactionsAggregation :one
SELECT COALESCE(category_name, 'uncategorized') AS category_name, SUM(amount) AS total_amount FROM transactions_view GROUP BY category_name ORDER BY total_amount;

-- name: ReadAllTransactions :many
SELECT transaction_id, "date", account_name, counter_party, amount, category_name FROM transactions_view ORDER BY "date";

-- name: ReadAllTransactionsByAccountNameAndDate :many
SELECT transaction_id, "date", account_name, counter_party, amount, category_name FROM transactions_view WHERE account_name=? AND strftime('%Y-%m', "date") = ?;

-- name: ReadAllTransactionsByAccountName :many
SELECT transaction_id, "date", account_name, counter_party, amount, category_name FROM transactions_view WHERE account_name=?;

-- name: ReadAllTransactionsByDate :many
SELECT transaction_id, "date", account_name, counter_party, amount, category_name FROM transactions_view WHERE strftime('%Y-%m', "date") = ?;

-- name: AggregateAllTransactions :many
SELECT COALESCE(category_name, 'uncategorized') AS category_name, SUM(amount) AS total_amount FROM transactions_view GROUP BY category_name ORDER BY total_amount;

-- name: AggregateAllTransactionsByAccountName :many
SELECT COALESCE(category_name, 'uncategorized') AS category_name, SUM(amount) AS total_amount FROM transactions_view WHERE account_name=? GROUP BY category_name ORDER BY total_amount;

-- name: AggregateAllTransactionsByDate :many
SELECT COALESCE(category_name, 'uncategorized') AS category_name, SUM(amount) AS total_amount FROM transactions_view WHERE strftime('%Y-%m', "date")=? GROUP BY category_name ORDER BY total_amount;