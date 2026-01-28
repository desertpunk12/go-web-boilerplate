-- name: CreateUser :one
INSERT INTO users (id, name, email, username, password)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY name;

-- name: UpdateUser :one
UPDATE users
SET name = $2, email = $3, username = $4, password = $5
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
