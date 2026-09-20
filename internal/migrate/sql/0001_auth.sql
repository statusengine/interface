-- The interface's own tables. Everything this project writes lives behind
-- the sei_ prefix; the statusengine_* tables belong to the worker and are
-- only ever read.

CREATE TABLE IF NOT EXISTS sei_roles (
    id          INT UNSIGNED NOT NULL AUTO_INCREMENT,
    name        VARCHAR(64)  NOT NULL,
    description VARCHAR(255) NOT NULL DEFAULT '',
    permissions JSON         NOT NULL,
    is_system   TINYINT(1)   NOT NULL DEFAULT 0,
    created_at  BIGINT       NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uniq_role_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS sei_users (
    id            INT UNSIGNED NOT NULL AUTO_INCREMENT,
    username      VARCHAR(190) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name  VARCHAR(190) NOT NULL DEFAULT '',
    email         VARCHAR(190) NOT NULL DEFAULT '',
    role_id       INT UNSIGNED NOT NULL,
    is_active     TINYINT(1)   NOT NULL DEFAULT 1,
    created_at    BIGINT       NOT NULL,
    updated_at    BIGINT       NOT NULL,
    last_login_at BIGINT       NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uniq_username (username),
    KEY idx_role (role_id),
    CONSTRAINT fk_user_role FOREIGN KEY (role_id) REFERENCES sei_roles (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- id holds a hash of the session token, never the token itself: a dump of
-- this table must not be usable as a set of live credentials.
CREATE TABLE IF NOT EXISTS sei_sessions (
    id         CHAR(64)     NOT NULL,
    user_id    INT UNSIGNED NOT NULL,
    created_at BIGINT       NOT NULL,
    expires_at BIGINT       NOT NULL,
    user_agent VARCHAR(255) NOT NULL DEFAULT '',
    ip         VARCHAR(45)  NOT NULL DEFAULT '',
    PRIMARY KEY (id),
    KEY idx_user (user_id),
    KEY idx_expires (expires_at),
    CONSTRAINT fk_session_user FOREIGN KEY (user_id) REFERENCES sei_users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Every external command submission, including the ones that were refused.
-- username is denormalised so the trail survives the user being deleted.
CREATE TABLE IF NOT EXISTS sei_command_audit (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id      INT UNSIGNED    NULL,
    username     VARCHAR(190)    NOT NULL,
    action       VARCHAR(64)     NOT NULL,
    target       VARCHAR(512)    NOT NULL DEFAULT '',
    payload      JSON            NULL,
    http_status  SMALLINT        NOT NULL DEFAULT 0,
    response     VARCHAR(1024)   NOT NULL DEFAULT '',
    remote_ip    VARCHAR(45)     NOT NULL DEFAULT '',
    created_at   BIGINT          NOT NULL,
    PRIMARY KEY (id),
    KEY idx_created (created_at),
    KEY idx_user (user_id),
    KEY idx_action (action, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
