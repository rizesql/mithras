-- name: RevokeUserSessions :exec
UPDATE "session"
SET
  revoked_at = @now
WHERE
  user_pk = @user_pk
AND
  revoked_at IS NULL;
