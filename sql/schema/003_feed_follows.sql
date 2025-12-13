-- +goose Up
CREATE TABLE feed_follows (
    id uuid PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_id uuid REFERENCES users ON DELETE CASCADE NOT NULL,
    feed_id uuid REFERENCES feeds ON DELETE CASCADE NOT NULL,
    CONSTRAINT "user feed relation" UNIQUE (feed_id, user_id)
)
;

-- +goose Down
DROP TABLE feed_follows
;
