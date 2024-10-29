CREATE TABLE users (
                       id INTEGER PRIMARY KEY AUTOINCREMENT,
                       user_id VARCHAR(32) UNIQUE NOT NULL,
                       nickname VARCHAR(32) NOT NULL,
                       password VARCHAR(64) NOT NULL,
                       email VARCHAR(128) NOT NULL,
                       created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                       updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                       deleted_at DATETIME DEFAULT NULL
);

-- 创建索引
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE INDEX idx_users_user_id ON users(user_id);
CREATE INDEX idx_users_email ON users(email);

-- 触发器：自动更新 updated_at 字段
CREATE TRIGGER trig_users_updated_at
    AFTER UPDATE ON users
BEGIN
    UPDATE users SET updated_at = CURRENT_TIMESTAMP
    WHERE id = NEW.id;
END;