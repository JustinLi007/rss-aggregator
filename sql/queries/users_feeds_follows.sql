-- name: FollowFeed :one
INSERT INTO users_feeds_follows(id, created_at, updated_at, user_id, feed_id)
VALUES($1, $2, $3, $4, $5)
RETURNING *;

-- name: UnfollowFeed :exec
DELETE FROM users_feeds_follows
WHERE id = $1;

-- name: GetFeedFollows :many
SELECT * FROM users_feeds_follows;
