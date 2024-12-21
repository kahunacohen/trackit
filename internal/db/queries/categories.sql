-- name: ReadCategoryIdByName :one
SELECT id FROM categories WHERE "name"=?;

-- name: ReadAllCategories :many
SELECT * from categories ORDER BY "name";
