-- +goose Up

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION is_valid_nanoid (val text, prefix text) returns boolean AS $$
  SELECT val ~ ('^' || prefix || '_[23456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]{20}$');
$$ language sql immutable;
-- +goose StatementEnd

CREATE DOMAIN user_id AS text CHECK (is_valid_nanoid (value, 'usr'));

CREATE TYPE user_status AS enum (
  'active',
  'suspended',
  'locked'
);

CREATE TABLE "user" (
  pk    bigint  GENERATED ALWAYS AS IDENTITY,
  id    user_id NOT NULL,
  name  text    NOT NULL,
  email text    NOT NULL,

  status          user_status NOT NULL DEFAULT 'active',
  failed_attempts integer     NOT NULL DEFAULT 0,
  locked_until    timestamptz,

  last_login_at timestamptz,
  created_at    timestamptz NOT NULL DEFAULT now(),
  updated_at    timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT user_pk                    PRIMARY KEY (pk),
  CONSTRAINT user_id_unique             UNIQUE (id),
  CONSTRAINT user_email_check           CHECK (
    email ~ '^[a-zA-Z0-9.!#$%&''*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*\.[a-zA-Z]{2,}$'
  ),
  CONSTRAINT user_name_check            CHECK (
    char_length(name) BETWEEN 1 AND 255
  ),
  CONSTRAINT user_failed_attempts_check CHECK (
    failed_attempts >= 0
  )
);

CREATE UNIQUE INDEX user_email_unique_idx ON "user" (lower(email));


CREATE TABLE "credential_password" (
  pk      bigint GENERATED ALWAYS AS IDENTITY,
  user_pk bigint NOT NULL,
  secret  text   NOT NULL,

  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT credential_password_pk          PRIMARY KEY (pk),
  CONSTRAINT credential_password_user_unique UNIQUE (user_pk),
  CONSTRAINT credential_password_user_fk     FOREIGN key (user_pk) REFERENCES "user" (pk) ON DELETE CASCADE
);

CREATE TABLE "password_history" (
  pk      bigint GENERATED ALWAYS AS IDENTITY,
  user_pk bigint NOT NULL,
  secret  text   NOT NULL,

  created_at timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT password_history_pk      PRIMARY KEY (pk),
  CONSTRAINT password_history_user_fk FOREIGN KEY (user_pk) REFERENCES "user" (pk) ON DELETE CASCADE
);

CREATE INDEX password_history_user_pk_idx ON "password_history" (user_pk, created_at DESC);

-- +goose Down
DROP INDEX password_history_user_pk_idx;
DROP TABLE "password_history";

DROP TABLE "credential_password";

DROP INDEX user_email_unique_idx;
DROP TABLE "user";

DROP TYPE user_status;
DROP DOMAIN user_id;
DROP FUNCTION is_valid_nanoid(text, text);
