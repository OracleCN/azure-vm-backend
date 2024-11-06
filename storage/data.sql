-- auto-generated definition
create table virtual_machines
(
    id              INTEGER
        primary key autoincrement,
    vm_id           VARCHAR(128)                          not null
        unique,
    account_id      VARCHAR(32)                           not null
        references accounts (account_id),
    subscription_id VARCHAR(128)                          not null
        references subscriptions (subscription_id),
    name            VARCHAR(128)                          not null,
    resource_group  VARCHAR(128)                          not null,
    location        VARCHAR(64)                           not null,
    size            VARCHAR(64)                           not null,
    status          VARCHAR(32) default 'Running'         not null,
    state           VARCHAR(32)                           not null,
    private_ips     TEXT,
    public_ips      TEXT,
    os_type         VARCHAR(32),
    os_disk_size    INTEGER,
    data_disks      TEXT,
    tags            TEXT,
    sync_status     VARCHAR(32) default 'pending'         not null,
    last_sync_at    DATETIME,
    created_at      DATETIME    default CURRENT_TIMESTAMP not null,
    updated_at      DATETIME    default CURRENT_TIMESTAMP not null,
    deleted_at      DATETIME,
    created_time    DATETIME,
    power_state     VARCHAR(32),
    core            int,
    memory          int,
    os_image        VARCHAR(32),
    constraint idx_vm_account_subscription
        unique (account_id, subscription_id, name)
);

create index idx_vm_account_id
    on virtual_machines (account_id);

create index idx_vm_status
    on virtual_machines (status);

create index idx_vm_subscription_id
    on virtual_machines (subscription_id);

create index idx_vm_sync_status
    on virtual_machines (sync_status);

-- auto-generated definition
create table users
(
    id         INTEGER
        primary key autoincrement,
    user_id    VARCHAR(32)                        not null
        unique,
    nickname   VARCHAR(32)                        not null,
    password   VARCHAR(64)                        not null,
    email      VARCHAR(128)                       not null,
    created_at DATETIME default CURRENT_TIMESTAMP not null,
    updated_at DATETIME default CURRENT_TIMESTAMP not null,
    deleted_at DATETIME default NULL
);

create index idx_users_deleted_at
    on users (deleted_at);

create index idx_users_email
    on users (email);

create index idx_users_user_id
    on users (user_id);

-- auto-generated definition
create table subscriptions
(
    id                    INTEGER
        primary key autoincrement,
    account_id            VARCHAR(32)                        not null,
    subscription_id       VARCHAR(36)                        not null,
    display_name          VARCHAR(128)                       not null,
    state                 VARCHAR(32)                        not null,
    subscription_policies TEXT,
    authorization_source  VARCHAR(32),
    subscription_type     VARCHAR(32),
    spending_limit        VARCHAR(32),
    start_date            DATETIME,
    end_date              DATETIME,
    created_at            DATETIME default CURRENT_TIMESTAMP not null,
    updated_at            DATETIME default CURRENT_TIMESTAMP not null,
    deleted_at            DATETIME default NULL,
    unique (account_id, subscription_id)
);

create index idx_subs_account_id
    on subscriptions (account_id);

create index idx_subs_subscription_id
    on subscriptions (subscription_id);

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

