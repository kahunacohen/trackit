-- name: ReadSettingByName :one
SELECT "value" FROM settings WHERE name=?;

-- name: CreateSetting :exec
INSERT INTO settings ("name", "value") VALUES (?, ?);