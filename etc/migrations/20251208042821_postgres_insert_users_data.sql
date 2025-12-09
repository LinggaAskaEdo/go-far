-- +goose Up
INSERT INTO users (name, email, age) VALUES
    ('John Doe', 'john.doe@example.com', 30),
    ('Jane Smith', 'jane.smith@example.com', 25),
    ('Bob Johnson', 'bob.johnson@example.com', 35)
ON CONFLICT (email) DO NOTHING;

-- +goose Down
DELETE FROM users
WHERE (email, name, age) IN (
    ('john.doe@example.com', 'John Doe', 30),
    ('jane.smith@example.com', 'Jane Smith', 25),
    ('bob.johnson@example.com', 'Bob Johnson', 35)
)
