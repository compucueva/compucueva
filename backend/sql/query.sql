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

-- name: UpdateTaskStatus :one
UPDATE tasks
SET status = ?, updated_at=CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: UpdateTask :one
UPDATE tasks
SET title=?, description=?, priority=?, updated_at=CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteTask :execrows
DELETE FROM tasks
WHERE id=?;

-- name: AssignUserToTask :exec
INSERT INTO user_task (user_id, task_id) 
VALUES ( ?, ? );

-- name: GetTaskAssignees :many
SELECT u.id, u.username, u.name FROM users u
JOIN user_task t ON u.id = t.user_id
WHERE t.task_id = ?;

-- name: CreateUser :one
INSERT INTO users (
  username, name, role
) VALUES (
  ?, ?, ?
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
