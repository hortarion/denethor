-- name: RegisterApp :one
INSERT INTO apps (id, created_at, updated_at, app_name)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1
)
RETURNING *;

-- name: ListApps :many
SELECT * FROM apps;
