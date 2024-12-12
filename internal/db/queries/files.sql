-- name: GetHashFromFileName :one
SELECT hash FROM files where name=?;

-- name: CreateFile :exec
INSERT INTO files (name, hash) VALUES (?, ?);

-- name: UpdateFileHashByName :exec
UPDATE files set hash=? WHERE name=?;