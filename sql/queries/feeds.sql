-- name: CreateFeed :one
INSERT INTO feeds (name, url, user_id)
VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetAllFeeds :many
SELECT * FROM feeds;

-- name: DeleteFeeds :exec
DELETE FROM feeds;