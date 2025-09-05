-- name: GetChirps :many
SELECT * FROM chirps 
WHERE ($1::uuid IS NULL OR user_id = $1)
ORDER BY created_at ASC;
