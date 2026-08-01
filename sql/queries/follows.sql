-- name: CreateFeedFollow :one
WITH insert_feed_follow AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING *
)
SELECT
    insert_feed_follow.*,
    feed.name AS feed_name,
    users.name AS user_name
FROM insert_feed_follow
JOIN feed ON insert_feed_follow.feed_id = feed.id
JOIN users ON insert_feed_follow.user_id = users.id;