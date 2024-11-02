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

