-- name: CreatePluginExampleData :one
INSERT INTO plugin_example_data (
    name
) VALUES (?)
RETURNING *;

-- name: GetPluginExampleData :one
SELECT * FROM plugin_example_data
WHERE id = ? LIMIT 1;
