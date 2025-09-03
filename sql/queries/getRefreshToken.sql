-- name: GetRefreshToken :one
select * from refresh_tokens
where token = $1;
