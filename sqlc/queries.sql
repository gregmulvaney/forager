-- name: CreatePlugin :one
INSERT INTO plugins (
    name, path
) VALUES ( ?,? )
RETURNING *;

-- name: GetPlugin :one
SELECT * FROM plugins
WHERE id = ? LIMIT 1;

-- name: ListPlugins :many
SELECT * FROM plugins
ORDER BY name;

-- name: UpdatePlugin :one
UPDATE plugins
set name = ?, path = ?, updated_at = ?
WHERE id = ?
RETURNING *;

-- name: DeletePlugin :exec
DELETE FROM plugins
WHERE id = ?;
