-- name: InsertAuthorizationCode :one
INSERT INTO authorization_code (
  code,
  user_pk,
  client_id,
  redirect_uri,
  challenge,
  expires_at
) VALUES (
  @code,
  @user_pk,
  @client_id,
  @redirect_uri,
  @challenge,
  @expires_at
) RETURNING
  code,
  client_id,
  redirect_uri,
  challenge,
  expires_at;

-- name: ConsumeAuthorizationCode :one
UPDATE authorization_code
SET
  used_at = now()
WHERE
  code = @code
  AND used_at IS NULL
  AND expires_at > now()
RETURNING
  code,
  user_pk,
  client_id,
  redirect_uri,
  challenge;
