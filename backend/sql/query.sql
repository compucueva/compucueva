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
