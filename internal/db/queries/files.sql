-- name: ReadHashFromFileName :one
SELECT hash FROM files where name=?;

-- name: CreateFile :exec
INSERT INTO files (name, hash) VALUES (?, ?);