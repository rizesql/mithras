-- name: RevokeSession :execrows
UPDATE "session"
SET revoked_at = now()
WHERE pk = @pk AND revoked_at IS NULL;
