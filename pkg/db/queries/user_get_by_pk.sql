-- name: GetUserByPk :one
SELECT * FROM "user"
WHERE pk = $1 LIMIT 1;
