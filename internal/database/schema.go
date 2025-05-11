package database

const Schema = `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    login TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS expressions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    expression TEXT NOT NULL,
    status TEXT NOT NULL,
    result REAL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    expression_id INTEGER NOT NULL,
    arg1 TEXT NOT NULL,
    arg2 TEXT NOT NULL,
    operation TEXT NOT NULL,
    operation_time INTEGER NOT NULL,
    completed INTEGER DEFAULT 0,
    result REAL,
    FOREIGN KEY (expression_id) REFERENCES expressions(id)
);

CREATE TABLE IF NOT EXISTS results (
    id TEXT PRIMARY KEY,
    expression_id INTEGER NOT NULL,
    task_id INTEGER,
    value REAL,
    completed INTEGER DEFAULT 0,
    FOREIGN KEY (expression_id) REFERENCES expressions(id),
    FOREIGN KEY (task_id) REFERENCES tasks(id)
);
`
