-- +goose Up
CREATE DOMAIN password_reset_id AS text CHECK (is_valid_nanoid (value, 'rst'));

CREATE TABLE "password_reset" (
  pk         bigint            GENERATED ALWAYS AS IDENTITY,
  id         password_reset_id NOT NULL,
  user_pk    bigint            NOT NULL,

  token_hash bytea       NOT NULL,
  user_agent text,
  ip_addr    inet        NOT NULL,
  expires_at timestamptz NOT NULL,
  used_at    timestamptz,

  created_at timestamptz       NOT NULL DEFAULT now(),

  CONSTRAINT password_reset_pk               PRIMARY KEY (pk),
  CONSTRAINT password_reset_id_unique        UNIQUE (id),
  CONSTRAINT password_reset_hash_unique      UNIQUE (token_hash),
  CONSTRAINT password_reset_user_fk          FOREIGN KEY (user_pk) REFERENCES "user" (pk) ON DELETE CASCADE,
  CONSTRAINT password_reset_user_agent_check CHECK (
    char_length(user_agent) < 1024
  ),
  CONSTRAINT password_reset_max_expiry       CHECK (
    expires_at <= created_at + interval '1 hour'
  ),
  CONSTRAINT password_reset_expiry_check     CHECK (
    expires_at > created_at
  ),
  CONSTRAINT password_reset_used_check       CHECK (
    used_at IS NULL OR used_at >= created_at
  )
);

CREATE INDEX password_reset_user_pk_idx ON "password_reset" (user_pk)
  WHERE used_at IS NULL;

CREATE UNIQUE INDEX password_reset_active_id_idx ON "password_reset" (id)
  WHERE used_at IS NULL;

CREATE INDEX password_reset_token_hash_idx ON "password_reset" (token_hash)
  WHERE used_at IS NULL;

-- +goose Down
DROP INDEX password_reset_user_pk_idx;
DROP INDEX password_reset_active_id_idx;
DROP INDEX password_reset_token_hash_idx;
DROP TABLE "password_reset";
DROP DOMAIN password_reset_id;
