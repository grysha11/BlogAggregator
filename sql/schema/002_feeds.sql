-- +goose Up
CREATE TABLE feeds (
    name VARCHAR(100) NOT NULL,
    url VARCHAR(150) UNIQUE NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE feeds;