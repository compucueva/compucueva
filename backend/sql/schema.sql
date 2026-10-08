-- backend/schema.sql
CREATE TABLE users(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
);

CREATE TABLE tasks(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL,
  description TEXT NOT NULL,
  status TEXT DEFAULT 'TODO' NOT NULL,
  priority TEXT DEFAULT 'NORMAL' NOT NULL,
  assignee INTEGER,
  item_id INTEGER,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
  FOREIGN KEY(assignee) REFERENCES users(id)
);

CREATE TABLE user_task(
  user_id INTEGER,
  task_id INTEGER,
  PRIMARY KEY(user_id, task_id),
  FOREIGN KEY(user_id) REFERENCES users(id),
  FOREIGN KEY(task_id) REFERENCES tasks(id)
);
