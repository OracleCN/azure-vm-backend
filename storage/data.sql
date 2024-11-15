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

-- vm_regions表
CREATE TABLE IF NOT EXISTS vm_regions (
                                          id INTEGER PRIMARY KEY AUTOINCREMENT,
                                          created_at DATETIME NOT NULL,
                                          updated_at DATETIME NOT NULL,
                                          deleted_at DATETIME,
                                          region_id VARCHAR(64) NOT NULL UNIQUE,
                                          display_name VARCHAR(128) NOT NULL,
                                          name VARCHAR(64) NOT NULL,
                                          location VARCHAR(64) NOT NULL,
                                          region_type VARCHAR(32) NOT NULL,
                                          category VARCHAR(32) NOT NULL,
                                          available BOOLEAN NOT NULL DEFAULT 1,
                                          enabled BOOLEAN NOT NULL DEFAULT 1,
                                          last_sync_at DATETIME NOT NULL,
                                          paired_region VARCHAR(64),
                                          resource_types TEXT,
                                          metadata TEXT
);

CREATE INDEX idx_vm_regions_deleted_at ON vm_regions(deleted_at);
CREATE INDEX idx_vm_regions_name ON vm_regions(name);
CREATE INDEX idx_vm_regions_location ON vm_regions(location);



-- vm_images表
CREATE TABLE IF NOT EXISTS vm_images (
                                         id INTEGER PRIMARY KEY AUTOINCREMENT,
                                         created_at DATETIME NOT NULL,
                                         updated_at DATETIME NOT NULL,
                                         deleted_at DATETIME,
                                         image_id VARCHAR(128) NOT NULL UNIQUE,
                                         name VARCHAR(128) NOT NULL,
                                         region_id VARCHAR(64) NOT NULL,
                                         publisher VARCHAR(128) NOT NULL,
                                         offer VARCHAR(128) NOT NULL,
                                         sku VARCHAR(128) NOT NULL,
                                         version VARCHAR(64) NOT NULL,
                                         os_type VARCHAR(32) NOT NULL,
                                         category VARCHAR(32) NOT NULL,
                                         available BOOLEAN NOT NULL DEFAULT 1,
                                         description TEXT,
                                         last_sync_at DATETIME NOT NULL,
                                         requirements TEXT,
                                         features TEXT,
                                         metadata TEXT
);

CREATE INDEX idx_vm_images_deleted_at ON vm_images(deleted_at);
CREATE INDEX idx_vm_images_region_id ON vm_images(region_id);
CREATE INDEX idx_vm_images_os_type ON vm_images(os_type);
CREATE INDEX idx_vm_images_publisher ON vm_images(publisher);



-- vm_sizes表
CREATE TABLE IF NOT EXISTS vm_sizes (
                                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                                        created_at DATETIME NOT NULL,
                                        updated_at DATETIME NOT NULL,
                                        deleted_at DATETIME,
                                        size_id VARCHAR(64) NOT NULL UNIQUE,
                                        name VARCHAR(64) NOT NULL,
                                        region_id VARCHAR(64) NOT NULL,
                                        category VARCHAR(32) NOT NULL,
                                        family VARCHAR(32) NOT NULL,
                                        number_of_cores INTEGER NOT NULL,
                                        memory_in_gb FLOAT NOT NULL,
                                        max_data_disks INTEGER NOT NULL,
                                        os_disk_size_in_gb INTEGER NOT NULL,
                                        available BOOLEAN NOT NULL DEFAULT 1,
                                        enabled BOOLEAN NOT NULL DEFAULT 1,
                                        last_sync_at DATETIME NOT NULL,
                                        network_bandwidth VARCHAR(32),
                                        temporary_disk_size_in_gb INTEGER,
                                        price_per_hour DECIMAL(10,4),
                                        accelerated_networking BOOLEAN DEFAULT 0,
                                        capabilities TEXT,
                                        metadata TEXT
);

CREATE INDEX idx_vm_sizes_deleted_at ON vm_sizes(deleted_at);
CREATE INDEX idx_vm_sizes_region_id ON vm_sizes(region_id);
CREATE INDEX idx_vm_sizes_category ON vm_sizes(category);
CREATE INDEX idx_vm_sizes_family ON vm_sizes(family);