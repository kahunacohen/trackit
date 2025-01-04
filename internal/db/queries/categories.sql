-- name: ReadCategoryIdByName :one
SELECT id FROM categories WHERE "name"=?;

-- name: ReadAllCategories :many
SELECT * from categories ORDER BY "name";

-- name: CreateCategory :exec
INSERT OR IGNORE INTO categories ("name") VALUES (?);
