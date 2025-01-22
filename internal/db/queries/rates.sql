-- name: ReadRateFromSymbols :one
SELECT 
    rates.rate 
FROM rates
INNER JOIN currency_codes AS from_currency ON from_currency.id = rates.currency_code_from_id
INNER JOIN currency_codes AS to_currency ON to_currency.id = rates.currency_code_to_id
WHERE from_currency.symbol =  sqlc.arg(fromSymbol) AND to_currency.symbol =  sqlc.arg(toSymbol);

-- name: CreateRate :exec
INSERT INTO rates (rate, currency_code_from_id, currency_code_to_id, "month")
VALUES (
    ?,
    (SELECT id FROM currency_codes WHERE currency_codes.symbol = sqlc.arg(fromSymbol)),
    (SELECT id FROM currency_codes WHERE currency_codes.symbol = sqlc.arg(toSymbol)),
    ?
);

-- name: DeleteRate :exec
DELETE FROM rates WHERE id=?;

-- name: UpdateRate :exec
UPDATE 
    rates 
SET 
    rate=?,
    currency_code_from_id=(SELECT currency_codes.id FROM currency_codes WHERE currency_codes.symbol = sqlc.arg(fromSymbol)),
    currency_code_to_id=(SELECT currency_codes.id FROM currency_codes WHERE currency_codes.symbol = sqlc.arg(toSymbol)),
    month=?;
