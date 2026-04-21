-- name: InsertSession :exec
INSERT INTO "session" (
  id,
  user_pk,
  token_hash,
  user_agent,
  ip_addr,
  expires_at
) VALUES (
  @id,
  @user_pk,
  @token_hash,
  @user_agent,
  @ip_addr,
  @expires_at
);
