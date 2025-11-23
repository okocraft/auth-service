CREATE TABLE IF NOT EXISTS users
(
    id         INT PRIMARY KEY AUTO_INCREMENT,
    uuid       BINARY(16)  NOT NULL UNIQUE,
    created_at DATETIME    NOT NULL
);

CREATE TABLE IF NOT EXISTS users_sub
(
    user_id    INT PRIMARY KEY REFERENCES users (id),
    sub        VARCHAR(255) NOT NULL UNIQUE,
    created_at DATETIME     NOT NULL
);

CREATE TABLE IF NOT EXISTS users_login_key
(
    user_id    INT PRIMARY KEY REFERENCES users (id),
    login_key  BIGINT   NOT NULL UNIQUE,
    created_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS users_refresh_tokens
(
    id         BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id    INT          NOT NULL REFERENCES users (id),
    jti        BINARY(16)   NOT NULL UNIQUE,
    login_id   BINARY(16)   NOT NULL,
    created_at DATETIME     NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_users_refresh_tokens_created_at ON users_refresh_tokens (created_at);

CREATE TABLE IF NOT EXISTS users_access_tokens
(
    id               BIGINT PRIMARY KEY AUTO_INCREMENT,
    refresh_token_id BIGINT     NOT NULL REFERENCES users_refresh_tokens (id) ON DELETE CASCADE,
    jti              BINARY(16) NOT NULL UNIQUE,
    created_at       DATETIME   NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_users_access_tokens_created_at ON users_access_tokens (created_at);

CREATE TABLE IF NOT EXISTS users_access_logs
(
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id     INT          NOT NULL REFERENCES users (id),
    action_type TINYINT      NOT NULL,
    login_id    BINARY(16)   NOT NULL UNIQUE,
    ip          BINARY(16)   NOT NULL,
    user_agent  VARCHAR(512) NOT NULL,
    created_at  DATETIME     NOT NULL
);
