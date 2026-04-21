-- +goose Up
CREATE DOMAIN key_id as text CHECK (is_valid_nanoid(value, 'key'));

CREATE TABLE "jws" (
  pk bigint GENERATED ALWAYS AS IDENTITY,
  id key_id NOT NULL,

  data       bytea       NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  rotates_at timestamptz NOT NULL,
  expires_at timestamptz NOT NULL,

  CONSTRAINT jws_pk              PRIMARY KEY (pk),
  CONSTRAINT jws_id_unique       UNIQUE (id),
  CONSTRAINT jws_lifecycle_order CHECK (
    created_at < rotates_at AND rotates_at < expires_at
  ),
  CONSTRAINT jws_min_grace       CHECK (
    expires_at >= rotates_at + interval '30 minutes'
  )
);

CREATE INDEX jws_expires_at_idx ON "jws" (expires_at);

-- +goose Down
DROP INDEX jws_expires_at_idx;
DROP TABLE "jws";
DROP DOMAIN key_id;
