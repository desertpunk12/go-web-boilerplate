-- name: CreateEmployee :one
INSERT INTO employees (id, first_name, last_name, email, department)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetEmployee :one
SELECT * FROM employees
WHERE id = $1 LIMIT 1;

-- name: ListEmployees :many
SELECT * FROM employees
ORDER BY last_name, first_name;

-- name: UpdateEmployee :one
UPDATE employees
SET first_name = $2, last_name = $3, email = $4, department = $5
WHERE id = $1
RETURNING *;

-- name: DeleteEmployee :exec
DELETE FROM employees
WHERE id = $1;
