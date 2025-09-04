-- name: SetIsRedChirpyUser :exec
update users
set is_chirpy_red = $1
where id = $2;
