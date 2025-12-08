-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: GetUserFeeds :many
SELECT f.* FROM feeds f
JOIN users u ON f.user_id = u.id
WHERE u.name = $1;

-- name: GetURLFeeds :one
SELECT * FROM feeds
WHERE url = $1;

-- name: GetAllFeeds :many
SELECT f.name, f.url, u.name user_name FROM feeds f
JOIN users u ON f.user_id = u.id;

-- name: MarkFeedFetched :one
UPDATE feeds f
SET created_at = CURRENT_TIMESTAMP,
    last_fetched_at = CURRENT_TIMESTAMP
WHERE f.id = $1
RETURNING *
;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1
;
