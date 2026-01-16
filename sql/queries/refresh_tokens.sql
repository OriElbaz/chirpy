-- name: MakeToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES ($1, NOW(), NOW(), $2, NOW() + INTERVAL '60 days', NULL)
RETURNING *;

-- name: GetToken :one
SELECT * FROM refresh_tokens
WHERE token = $1;

-- name: RevokeToken :one
UPDATE refresh_tokens
SET revoked_at = NOW(), updated_at = NOW()
WHERE refresh_tokens.user_id = $1
RETURNING *;