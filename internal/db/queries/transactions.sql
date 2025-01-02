-- name: CreateTransaction :exec
INSERT INTO transactions (account_id, date, amount, counter_party, category_id) VALUES (?, ?, ?, ?, ?);

-- name: ReadTransactionById :one
SELECT * from transactions_view WHERE transaction_id=?;

-- name: ReadNonCategorizedTransactions :many
SELECT transaction_id AS id, "date", counter_party, account_name, amount FROM transactions_view WHERE category_name IS NULL;

-- name: UpdateTransactionCategory :exec
UPDATE transactions SET category_id=? WHERE id=?;

-- name: UpdateTransactionIgnore :exec
UPDATE transactions SET ignore_when_summing=? WHERE id=?;

-- name: ReadTransactionsAggregation :one
SELECT COALESCE(category_name, 'uncategorized') AS category_name, SUM(amount) AS total_amount FROM transactions_view GROUP BY category_name ORDER BY total_amount;

-- name: ReadTransactions :many
SELECT transaction_id, "date", account_name, counter_party, amount, ignore_when_summing, category_name FROM transactions_view ORDER BY "date" DESC;

-- name: ReadTransactionsByAccountNameAndDate :many
SELECT transaction_id, "date", account_name, counter_party, amount, ignore_when_summing, category_name FROM transactions_view WHERE account_name=? AND strftime('%Y-%m', "date") = ?;

-- name: ReadTransactionsByAccountName :many
SELECT transaction_id, "date", account_name, counter_party, amount, ignore_when_summing, category_name FROM transactions_view WHERE account_name=?;

-- name: ReadTransactionsByDate :many
SELECT transaction_id, "date", account_name, counter_party, amount, ignore_when_summing, category_name FROM transactions_view WHERE strftime('%Y-%m', "date") = ?;

-- name: AggregateTransactions :many
SELECT COALESCE(category_name, 'Uncategorized') AS category_name, SUM(amount) AS total_amount FROM transactions_view WHERE ignore_when_summing = false GROUP BY category_name ORDER BY total_amount;

-- name: AggregateTransactionsByAccountName :many
SELECT COALESCE(category_name, 'Uncategorized') AS category_name, SUM(amount) AS total_amount FROM transactions_view WHERE account_name=? GROUP BY category_name ORDER BY total_amount;

-- name: AggregateTransactionsByDate :many
SELECT COALESCE(category_name, 'Uncategorized') AS category_name, SUM(amount) AS total_amount FROM transactions_view WHERE strftime('%Y-%m', "date")=? GROUP BY category_name ORDER BY total_amount;

-- name: AggregateTransactionsByAccountNameAndDate :many
SELECT COALESCE(category_name, 'Uncategorized') AS category_name, SUM(amount) AS total_amount FROM transactions_view WHERE account_name=? AND strftime('%Y-%m', date)=? GROUP BY category_name ORDER BY total_amount;