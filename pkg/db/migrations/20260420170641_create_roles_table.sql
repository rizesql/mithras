-- +goose Up

CREATE TABLE "role" (
  pk          bigint      GENERATED ALWAYS AS IDENTITY,
  name        text        NOT NULL,
  description text        NOT NULL DEFAULT '',
  created_at  timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT role_pk          PRIMARY KEY (pk),
  CONSTRAINT role_name_unique UNIQUE (name),
  CONSTRAINT role_name_format CHECK (
    name = upper(name) AND length(name) BETWEEN 1 AND 64
  )
);

CREATE TABLE "user_role" (
  user_pk bigint NOT NULL,
  role_pk bigint NOT NULL,

  granted_at timestamptz NOT NULL DEFAULT now(),
  granted_by bigint      REFERENCES "user" (pk) ON DELETE SET NULL,

  CONSTRAINT user_role_pk      PRIMARY KEY (user_pk, role_pk),
  CONSTRAINT user_role_user_fk FOREIGN KEY (user_pk) REFERENCES "user" (pk) ON DELETE CASCADE,
  CONSTRAINT user_role_role_fk FOREIGN KEY (role_pk) REFERENCES "role" (pk) ON DELETE RESTRICT
);

CREATE INDEX user_role_user_pk_idx ON "user_role" (user_pk);

-- Seed roles
INSERT INTO "role" (name, description) VALUES
  ('USER',    'Regular user with access to own resources'),
  ('ANALYST', 'Read access to assigned resources and reports'),
  ('MANAGER', 'Full access including user and role management');

-- +goose Down
DROP INDEX user_role_user_pk_idx;
DROP TABLE "user_role";
DROP TABLE "role";
