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

-- name: GetFeed :one
SELECT * from feed WHERE url = $1;

-- name: GetFeedFollowsForUser :many
SELECT
    feed_follows.*,
    feed.name AS feed_name,
    users.name AS user_name
FROM feed_follows
         JOIN feed ON feed_follows.feed_id = feed.id
         JOIN users ON feed_follows.user_id = users.id
WHERE users.name = $1;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows
    USING feed
WHERE feed_follows.feed_id = feed.id
  AND feed_follows.user_id = $1
  AND feed.url = $2;

-- name: MarkFeedFetched :exec
UPDATE feed
SET updated_at = NOW(), last_fetch_at = NOW()
WHERE id = $1;

-- name: GetNextFeedToFetch :one
SELECT * FROM feed
ORDER BY last_fetch_at ASC NULLS FIRST
LIMIT 1;