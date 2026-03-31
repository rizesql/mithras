-- +goose Up

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION is_valid_nanoid (val text, prefix text) returns boolean AS $$
  SELECT val ~ ('^' || prefix || '_[23456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]{20}$');
$$ language sql immutable;
-- +goose StatementEnd

CREATE DOMAIN user_id AS text CHECK (is_valid_nanoid (value, 'usr'));

CREATE TABLE "user" (
  pk         bigint      GENERATED ALWAYS AS IDENTITY,
  id         user_id     NOT NULL,
  name       text        NOT NULL,
  email      text        NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT user_pk PRIMARY KEY (pk),
  CONSTRAINT user_id_unique UNIQUE (id),
  CONSTRAINT user_email_check CHECK (
    email ~ '^[a-zA-Z0-9.!#$%&''*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*\.[a-zA-Z]{2,}$'
  ),
  CONSTRAINT user_name_check CHECK (char_length(name) BETWEEN 1 AND 255)
);

CREATE UNIQUE INDEX user_email_unique_idx ON "user" (lower(email));


CREATE TABLE "credential_password" (
  pk         bigint          GENERATED ALWAYS AS IDENTITY,
  user_id    user_id         NOT NULL,
  secret     text            NOT NULL,
  created_at timestamptz     NOT NULL DEFAULT now(),
  updated_at timestamptz     NOT NULL DEFAULT now(),
  CONSTRAINT credential_password_pk PRIMARY KEY (pk),
  CONSTRAINT credential_password_user_id_fk FOREIGN key (user_id) REFERENCES "user" (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE "credential_password";

DROP INDEX if EXISTS user_email_unique_idx;
DROP TABLE "user";

DROP DOMAIN user_id;

DROP FUNCTION is_valid_nanoid (text, text);
