-- backend/schema.sql
CREATE TABLE users(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL,
  password_hash TEXT NOT NULL,
  name TEXT NOT NULL, -- FULL NAME e.g. "Lionel Messi"
  role TEXT, -- ?
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE tasks(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL,
  description TEXT NOT NULL,
  status TEXT DEFAULT 'TODO' NOT NULL,
  priority TEXT DEFAULT 'NORMAL' NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE items(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  type TEXT NOT NULL, -- PC, FUENTE, MOBO, etc
  description TEXT,
  ingressed_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL, -- no me gusta la palabra ingressed, sujeto a cambios jajaj
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE user_task(
  user_id INTEGER,
  task_id INTEGER,
  PRIMARY KEY(user_id, task_id),
  FOREIGN KEY(user_id) REFERENCES users(id),
  FOREIGN KEY(task_id) REFERENCES tasks(id)
);

CREATE TABLE task_item(
  task_id INTEGER,
  item_id INTEGER,
  PRIMARY KEY(task_id, item_id),
  FOREIGN KEY(task_id) REFERENCES tasks(id),
  FOREIGN KEY(item_id) REFERENCES items(id)
);
