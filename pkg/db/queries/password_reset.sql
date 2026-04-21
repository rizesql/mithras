-- name: PasswordResetInsert :one
INSERT INTO "password_reset" (
  id,
  user_pk,
  token_hash,
  user_agent,
  ip_addr,
  expires_at
) VALUES (
  $1, $2, $3, $4, $5, $6
) RETURNING pk;

-- name: PasswordResetGetActive :one
SELECT
  pk,
  id,
  user_pk,
  token_hash,
  expires_at
FROM
  "password_reset"
WHERE
  id = $1
  AND used_at IS NULL
  AND expires_at > now()
LIMIT 1;

-- name: PasswordResetMarkUsed :exec
UPDATE "password_reset"
SET used_at = now()
WHERE pk = $1;

-- name: PasswordResetInvalidateSiblings :exec
UPDATE "password_reset"
SET used_at = now()
WHERE user_pk = $1 AND used_at IS NULL AND pk != $2;
