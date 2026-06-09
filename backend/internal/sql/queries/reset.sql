-- name: ResetUsers :exec
DELETE FROM users;

-- name: ResetRefreshTokens :exec
DELETE FROM refresh_tokens;

-- name: ResetApps :exec
DELETE FROM apps;
