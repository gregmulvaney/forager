-- name: CreatePlugin :one
INSERT INTO plugins (
    name, path, hash, version
) VALUES ( ?, ?, ?, ? ) 
RETURNING *;

-- name: GetPlugin :one 
SELECT * FROM plugins
WHERE id = ? LIMIT 1;

-- name: ListPlugins :many
SELECT * FROM plugins
ORDER BY name;

-- name: UpdatePlugin :one
UPDATE plugins
SET name = ?, path = ?, hash = ?, updated_at = ?
WHERE id = ?
RETURNING *;

-- name: DeletePlugin :exec
DELETE FROM plugins
WHERE id = ?;

