-- name: InsertCredential :exec
INSERT INTO "credential_password" (
  user_pk,
  secret
) VALUES  (
  @user_pk,
  @secret
);

-- name: UpdateCredentialByUserId :exec
UPDATE "credential_password"
SET
  secret = @secret,
  updated_at = now()
WHERE
  user_pk = @user_pk;
