-- name: ReadRate :one
SELECT rates.rate FROM rates WHERE currency_codes.symbol = ? AND curr = ?;