-- name: UpdateUser :one
update users
set email = $1, hashed_password = $2
where id = $3
returning *;
