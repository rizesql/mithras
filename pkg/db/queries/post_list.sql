-- name: ListPost :many
SELECT *
FROM post
WHERE pk > sqlc.arg(pagination_cursor)
ORDER BY pk ASC
LIMIT $1;
