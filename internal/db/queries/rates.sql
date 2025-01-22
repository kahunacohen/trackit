-- name: ReadRateFromSymbols :one
SELECT 
    rates.rate 
FROM rates
INNER JOIN currency_codes AS from_currency ON from_currency.id = rates.currency_code_from_id
INNER JOIN currency_codes AS to_currency ON to_currency.id = rates.currency_code_to_id
WHERE from_currency.symbol =  sqlc.arg(fromSymbol) AND to_currency.symbol =  sqlc.arg(toSymbol);