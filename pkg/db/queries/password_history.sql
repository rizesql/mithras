-- name: InsertPasswordHistory :exec
INSERT INTO "password_history" (user_pk, secret)
VALUES ($1, $2);

-- name: GetRecentPasswordHashes :many
SELECT
  secret
FROM
  "password_history"
WHERE
  user_pk = $1
ORDER BY
  created_at DESC
LIMIT $2;
