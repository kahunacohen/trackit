-- name: CreateCurrencyCode :exec
INSERT INTO currency_codes (symbol) VALUES (?);

-- name: ReadCurrencyCodes :many
SELECT * FROM currency_codes ORDER BY symbol;

-- name: DeleteCurrencyCode :exec
DELETE FROM currency_codes WHERE symbol=?;

-- name: UpdateCurrencyCode :exec
UPDATE currency_codes SET "symbol"=sqlc.arg(newSymbol) WHERE id=(
    SELECT id FROM currency_codes WHERE currency_codes.symbol=sqlc.arg(oldSymbol)
);