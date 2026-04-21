-- name: GetUserWithPassword :one
SELECT
  u.pk,
  u.id,
  u.status,
  u.failed_attempts,
  u.locked_until,
  cp.secret
FROM
  "user" u
JOIN
  "credential_password" cp ON u.pk = cp.user_pk
WHERE
  u.email = @email
LIMIT 1;
