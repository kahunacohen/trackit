-- name: ReadCategoryIdByName :one
SELECT id FROM categories WHERE name=?

-- name: ReadAllCategories :many
SELECT id, name FROM categories;
