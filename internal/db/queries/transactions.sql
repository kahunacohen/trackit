-- name: ReadTransactionByHash :one
SELECT * FROM transactions WHERE hash=?;

-- name: CreateTransaction :exec
INSERT INTO transactions (hash, account_id, date, amount, counter_party, category_id) VALUES (?, ?, ?, ?, ?, ?);

-- name: ReadAllTransactionsAggregation :one
SELECT COALESCE(category_name, 'uncategorized') AS category_name, SUM(amount) AS total_amount FROM transactions_view GROUP BY category_name ORDER BY total_amount;