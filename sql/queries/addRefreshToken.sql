-- name: AddRefreshToken :exec
INSERT INTO refresh_tokens (
  token, created_at, updated_at, user_id, experies_at  
) VALUES ( $1, $2, $3, $4, $5 );
