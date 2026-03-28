-- +goose Up
-- +goose StatementBegin
CREATE DOMAIN post_id AS varchar(24) CHECK (VALUE LIKE 'pst_%');

CREATE TABLE post (
    pk bigint GENERATED ALWAYS AS IDENTITY,
    id post_id NOT NULL,
    title text,
    body text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz,

    CONSTRAINT post_pk PRIMARY KEY(pk),
    CONSTRAINT post_id_unique UNIQUE(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE post;
DROP DOMAIN post_id;
-- +goose StatementEnd
