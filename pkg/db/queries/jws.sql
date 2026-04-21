-- name: InsertJWSKey :exec
INSERT INTO "jws" (
  id,
  data,
  rotates_at,
  expires_at
) VALUES (
  @id,
  @data,
  @rotates_at,
  @expires_at
);

-- name: GetActiveJWSKeys :many
SELECT
  j.id,
  j.data,
  j.rotates_at,
  j.expires_at
FROM
  "jws" j
WHERE
  j.expires_at > @now;

-- name: GetSigningKey :one
SELECT
  j.id,
  j.data,
  j.rotates_at,
  j.expires_at
FROM
  "jws" j
WHERE
  j.rotates_at > @now
ORDER BY
  j.created_at DESC
LIMIT 1;

-- name: PruneJWS :exec
DELETE FROM "jws"
WHERE expires_at < @now;
