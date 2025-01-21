-- name: CreateCurrencyCode :exec
INSERT INTO currency_codes (symbol) VALUES (?);

-- name: ReadCurrencyCodes :many
SELECT * FROM currency_codes;