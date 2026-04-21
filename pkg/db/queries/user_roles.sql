-- name: GetUserRoles :many
SELECT
  r.name
FROM
  "role" r
JOIN
  "user_role" ur ON r.pk = ur.role_pk
WHERE
  ur.user_pk = $1;

-- name: AssignRole :exec
INSERT INTO "user_role" (user_pk, role_pk, granted_by)
VALUES ($1, (SELECT pk FROM "role" WHERE name = $2), $3);
