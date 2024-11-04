-- auto-generated definition
create table accounts
(
    id                  INTEGER
        primary key autoincrement,
    account_id          VARCHAR(32)                           not null
        unique,
    user_id             VARCHAR(32)                           not null,
    login_email         VARCHAR(128)                          not null,
    login_password      VARCHAR(128)                          not null,
    remark              TEXT,
    app_id              VARCHAR(128)                          not null,
    password            VARCHAR(256)                          not null,
    tenant              VARCHAR(128)                          not null,
    display_name        VARCHAR(128)                          not null,
    subscription_status VARCHAR(32) default 'normal'          not null,
    created_at          DATETIME    default CURRENT_TIMESTAMP not null,
    updated_at          DATETIME    default CURRENT_TIMESTAMP not null,
    deleted_at          DATETIME    default NULL,
    vm_count            integer     default 0,
    constraint chk_subscription_status
        check (subscription_status IN ('normal', 'error'))
);

create index idx_azure_accounts_deleted_at
    on accounts (deleted_at);

create index idx_azure_accounts_login_email
    on accounts (login_email);

create index idx_azure_accounts_subscription_status
    on accounts (subscription_status);

create index idx_azure_accounts_user_id
    on accounts (user_id);

CREATE TABLE subscriptions (
                               id INTEGER PRIMARY KEY AUTOINCREMENT,
                               account_id VARCHAR(32) NOT NULL,           -- 关联的账户ID
                               subscription_id VARCHAR(36) NOT NULL,      -- Azure 订阅ID
                               display_name VARCHAR(128) NOT NULL,        -- 显示名称
                               state VARCHAR(32) NOT NULL,                -- 订阅状态
                               subscription_policies TEXT,                -- 订阅策略(JSON)
                               authorization_source VARCHAR(32),          -- 授权来源
                               subscription_type VARCHAR(32),             -- 订阅类型(Student/FreeTrial/PayAsYouGo等)
                               spending_limit VARCHAR(32),                -- 消费限制(On/Off)
                               start_date DATETIME,                       -- 订阅开始时间
                               end_date DATETIME,                         -- 订阅结束时间
                               created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
                               updated_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
                               deleted_at DATETIME DEFAULT NULL,
                               UNIQUE(account_id, subscription_id)
);

CREATE INDEX idx_subs_account_id ON subscriptions(account_id);
CREATE INDEX idx_subs_subscription_id ON subscriptions(subscription_id);