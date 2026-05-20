-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, username, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
    )
RETURNING *;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: DeleteUser :one
UPDATE users SET deleted_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CheckUserByName :one
SELECT EXISTS(SELECT 1 FROM users WHERE username = $1);
