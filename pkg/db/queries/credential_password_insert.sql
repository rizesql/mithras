-- name: InsertCredential :exec
INSERT INTO "credential_password" (
  user_id,
  secret
) VALUES  (
  @user_id,
  @secret
);
