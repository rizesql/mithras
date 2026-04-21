-- name: GetSessionByTokenHash :one
SELECT
  s.pk,
  s.id,
  s.user_pk,
  s.expires_at,
  s.revoked_at,
  s.user_agent,
  s.ip_addr,
  u.id AS user_id,
  u.status AS user_status,
  u.locked_until AS user_locked_until
FROM
  "session" s
JOIN
  "user" u ON s.user_pk = u.pk
WHERE
  s.token_hash = @token_hash
LIMIT 1;
