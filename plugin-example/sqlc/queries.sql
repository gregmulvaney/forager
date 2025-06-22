-- name: CreateExample :one
INSERT INTO plugin_example (
    name
) VALUES ( ? )
    RETURNING *;
