-- name: CreateTransaction :exec
INSERT INTO transactions (account_id, date, amount, counter_party, category_id, ignore_when_summing) VALUES (?, ?, ?, ?, ?, ?);

-- name: ReadTransactionById :one
SELECT * from transactions_view WHERE transaction_id=?;

-- name: ReadNonCategorizedTransactions :many
SELECT * FROM transactions_view WHERE category_name IS NULL;

-- name: UpdateTransactionCategory :exec
UPDATE transactions SET category_id=? WHERE id=?;

-- name: UpdateTransactionIgnore :exec
UPDATE transactions SET ignore_when_summing=? WHERE id=?;

-- name: ReadTransactionsAggregation :one
SELECT COALESCE(category_name, 'uncategorized') AS category_name, SUM(amount) AS total_amount FROM transactions_view GROUP BY category_name ORDER BY total_amount;

-- name: ReadTransactionsWithSum :many
SELECT *, SUM(CASE WHEN NOT ignore_when_summing THEN amount ELSE 0 END) OVER () AS total_amount FROM transactions_view ORDER BY "date" DESC;

-- name: ReadTransactionsByAccountNameAndDateWithSum :many
SELECT *, SUM(CASE WHEN NOT ignore_when_summing THEN amount ELSE 0 END) OVER () AS total_amount FROM transactions_view WHERE account_name=? AND strftime('%Y-%m', "date") = ?;

-- name: ReadTransactionsByAccountNameWithSum :many
SELECT *, SUM(CASE WHEN NOT ignore_when_summing THEN amount ELSE 0 END) OVER () AS total_amount  FROM transactions_view WHERE account_name=?;

-- name: ReadTransactionsByDateWithSum :many
SELECT *, SUM(CASE WHEN NOT ignore_when_summing THEN amount ELSE 0 END) OVER () AS total_amount FROM transactions_view WHERE strftime('%Y-%m', "date") = ?;

-- name: AggregateTransactions :many
SELECT COALESCE(category_name, 'Uncategorized') AS category_name, SUM(CASE WHEN NOT ignore_when_summing THEN amount ELSE 0 END) AS total_amount FROM transactions_view WHERE ignore_when_summing = false GROUP BY category_name ORDER BY total_amount;

-- name: AggregateTransactionsByAccountName :many
SELECT COALESCE(category_name, 'Uncategorized') AS category_name, SUM(CASE WHEN NOT ignore_when_summing THEN amount ELSE 0 END) AS total_amount FROM transactions_view WHERE ignore_when_summing = false AND account_name=? GROUP BY category_name ORDER BY total_amount;

-- name: AggregateTransactionsByDate :many
SELECT COALESCE(category_name, 'Uncategorized') AS category_name, SUM(CASE WHEN NOT ignore_when_summing THEN amount ELSE 0 END) AS total_amount FROM transactions_view WHERE ignore_when_summing = false AND strftime('%Y-%m', "date")=? GROUP BY category_name ORDER BY total_amount;

-- name: AggregateTransactionsByAccountNameAndDate :many
SELECT COALESCE(category_name, 'Uncategorized') AS category_name, SUM(CASE WHEN NOT ignore_when_summing THEN amount ELSE 0 END) AS total_amount FROM transactions_view WHERE ignore_when_summing = false AND account_name=? AND strftime('%Y-%m', date)=? GROUP BY category_name ORDER BY total_amount;

-- name: SearchTransactionsWithSum :many
SELECT *, SUM(amount) OVER () AS total_amount FROM transactions_view WHERE counter_party LIKE '%' || :search_term || '%' ORDER BY "date" DESC;
