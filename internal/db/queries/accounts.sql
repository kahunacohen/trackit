-- name: ReadAccountIdByName :one
SELECT id FROM accounts where "name"=?;

-- name: CreateAccount :one
INSERT INTO accounts ("name", currency) VALUES (?, ?) RETURNING id;

-- name: ReadLastInsertedAccountID :one
SELECT last_insert_rowid();



