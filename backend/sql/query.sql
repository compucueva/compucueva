-- backend/query.sql

-- name: CreateTask :one
INSERT INTO tasks (
  title, description, status
) VALUES (
  ?, ?, 'TODO'
)
RETURNING *;

-- name: GetTask :one
SELECT * FROM tasks
WHERE id = ? LIMIT 1;

-- name: ListTasks :many
SELECT * FROM tasks
ORDER BY created_at DESC;

-- name: CreateUser :one
INSERT INTO users (
  username, password_hash, name, role
) VALUES (
  ?, ?, ?, ?
)
RETURNING *;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = ?;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = ?;

-- name: GetUsersByRole :many
SELECT * FROM users WHERE role = ?;

-- name: GetUsersWithTasks :many
SELECT * FROM users u
WHERE EXISTS (
  SELECT 1 FROM user_task ut WHERE ut.user_id = u.id
);

-- name: GetUsersWithoutTasks :many
SELECT * FROM users u
WHERE NOT EXISTS (
  SELECT 1 FROM user_task ut WHERE ut.user_id = u.id
);
