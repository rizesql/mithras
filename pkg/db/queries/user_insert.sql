-- name: InsertUser :one
INSERT INTO "user" (
  id,
  name,
  email
) values (
  @id,
  @name,
  @email
)
RETURNING pk;
