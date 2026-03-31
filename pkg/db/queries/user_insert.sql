-- name: InsertUser :exec
INSERT INTO "user" (
  id,
  name,
  email
) values (
  @id,
  @name,
  @email
);
