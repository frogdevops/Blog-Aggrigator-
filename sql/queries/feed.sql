-- name: CreateFeed :one
INSERT INTO feed (id, created_at, updated_at, name, url, user_id)
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6

       )
RETURNING *;

-- name: GetFeeds :many
SELECT feed.name AS feed_name, feed.url AS feed_url, users.name AS user_name
FROM feed
JOIN users ON feed.user_id = users.id;