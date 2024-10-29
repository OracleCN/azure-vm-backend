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

-- azure

-- Azure账号表
CREATE TABLE azure_accounts (
                                id INTEGER PRIMARY KEY AUTOINCREMENT,
                                account_id VARCHAR(32) UNIQUE NOT NULL,  -- 账号唯一标识符
                                user_id VARCHAR(32) NOT NULL,            -- 关联到users表的user_id
                                login_email VARCHAR(128) NOT NULL,       -- Azure登录邮箱
                                login_password VARCHAR(128) NOT NULL,    -- Azure登录密码（建议加密存储）
                                remark TEXT,                             -- 备注信息
                                app_id VARCHAR(128) NOT NULL,            -- Azure APP ID
                                app_password VARCHAR(256) NOT NULL,      -- Azure APP密码
                                tenant_id VARCHAR(128) NOT NULL,         -- Azure Tenant ID
                                subscription_id VARCHAR(128) NOT NULL,    -- Azure Subscription ID
                                subscription_status VARCHAR(32) NOT NULL DEFAULT 'normal',  -- 订阅状态：normal/error
                                created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                deleted_at DATETIME DEFAULT NULL,

    -- 检查约束
                                CONSTRAINT chk_subscription_status CHECK (subscription_status IN ('normal', 'error'))
);

-- 创建索引
CREATE INDEX idx_azure_accounts_user_id ON azure_accounts(user_id);
CREATE INDEX idx_azure_accounts_login_email ON azure_accounts(login_email);
CREATE INDEX idx_azure_accounts_deleted_at ON azure_accounts(deleted_at);
CREATE INDEX idx_azure_accounts_subscription_status ON azure_accounts(subscription_status);

-- 创建更新时间触发器
CREATE TRIGGER trig_azure_accounts_updated_at
    AFTER UPDATE ON azure_accounts
BEGIN
    UPDATE azure_accounts SET updated_at = CURRENT_TIMESTAMP
    WHERE id = NEW.id;
END;