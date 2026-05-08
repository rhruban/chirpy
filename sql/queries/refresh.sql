-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at)
VALUES (
  $1,
  NOW(),
  NOW(),
  $2,
  $3
)
RETURNING *;

-- name: GetRefreshByRefresh :one
SELECT * FROM refresh_tokens
WHERE token = $1;

-- name: RevokeRefresh :exec
UPDATE refresh_tokens SET updated_at = NOW(), revoked_at = NOW() where token = $1;
