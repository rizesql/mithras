-- +goose Up
CREATE DOMAIN authorization_code_id AS text CHECK (is_valid_nanoid(value, 'auc'));

CREATE TABLE authorization_code (
  pk        bigint                GENERATED ALWAYS AS IDENTITY,
  code      authorization_code_id NOT NULL,
  user_pk   bigint                NOT NULL REFERENCES "user"(pk) ON DELETE CASCADE,
  client_id text                  NOT NULL,

  redirect_uri text   NOT NULL,
  scopes       text[] NOT NULL DEFAULT '{}',
  challenge    text   NOT NULL,

  created_at timestamptz NOT NULL DEFAULT now(),
  expires_at timestamptz NOT NULL,
  used_at    timestamptz,

  CONSTRAINT authorization_code_pk         PRIMARY KEY (pk),
  CONSTRAINT authorization_code_unique     UNIQUE (code),
  CONSTRAINT authorization_code_lifetime   CHECK (
    expires_at <= created_at + interval '10 minutes'
  ),
  CONSTRAINT authorization_code_min_ttl    CHECK (
    expires_at > created_at
  ),
  CONSTRAINT authorization_code_used_order CHECK (
    used_at IS NULL OR used_at >= created_at
  ),
  CONSTRAINT authorization_code_challenge  CHECK (
    length(challenge) BETWEEN 43 AND 128
  )
);

CREATE INDEX authorization_code_expires_at_idx ON authorization_code (expires_at)
  WHERE used_at IS NULL;

CREATE INDEX authorization_code_user_pk_idx ON authorization_code (user_pk)
  WHERE used_at IS NULL;

-- +goose Down
DROP INDEX authorization_code_expires_at_idx;
DROP INDEX authorization_code_user_pk_idx;
DROP TABLE authorization_code;
DROP DOMAIN authorization_code_id;
