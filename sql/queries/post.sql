-- name: CreatePost :exec
INSERT INTO post (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8
       )
ON CONFLICT (url) DO NOTHING;

-- name: GetPostsForUser :many
SELECT post.*
FROM post
JOIN feed_follows ON post.feed_id = feed_follows.feed_id
WHERE feed_follows.user_id = $1
ORDER BY post.published_at DESC NULLS LAST
LIMIT $2;