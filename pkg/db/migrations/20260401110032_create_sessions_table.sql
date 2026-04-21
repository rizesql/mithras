-- +goose Up
CREATE DOMAIN session_id as text CHECK (is_valid_nanoid (value, 'ses'));

CREATE TABLE "session" (
  pk         bigint      GENERATED ALWAYS AS IDENTITY,
  id         session_id  NOT NULL,
  user_pk    bigint      NOT NULL,

  token_hash bytea       NOT NULL,
  user_agent text,
  ip_addr    inet        NOT NULL,
  expires_at timestamptz NOT NULL,
  revoked_at timestamptz,

  created_at timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT session_pk            PRIMARY KEY (pk),
  CONSTRAINT session_id_unique     UNIQUE (id),
  CONSTRAINT session_user_fk       FOREIGN KEY (user_pk) REFERENCES "user" (pk) ON DELETE CASCADE,
  CONSTRAINT session_user_agent    CHECK (
    char_length(user_agent) < 1024
  ),
  CONSTRAINT session_expires_check CHECK (
    expires_at > created_at
  ),
  CONSTRAINT session_revoked_check CHECK (
    revoked_at IS NULL OR revoked_at >= created_at
  )
);

CREATE UNIQUE INDEX session_token_hash_idx ON "session" (token_hash);

CREATE INDEX session_user_pk_ix ON "session" (user_pk)
  WHERE revoked_at is NULL;

-- +goose Down
DROP INDEX session_token_hash_idx;
DROP INDEX session_user_pk_ix;
DROP TABLE "session";
DROP DOMAIN session_id;
