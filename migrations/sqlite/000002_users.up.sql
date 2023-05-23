CREATE TABLE user (
    id                  TEXT NOT NULL PRIMARY KEY,
    created_at          TIMESTAMP NOT NULL,
    updated_at          TIMESTAMP NOT NULL,
    deleted_at          TIMESTAMP,
    name                TEXT
);
CREATE INDEX user_order ON user(name) WHERE deleted_at IS NULL;

CREATE TABLE user_credential (
    id                  TEXT NOT NULL PRIMARY KEY,
    created_at          TIMESTAMP NOT NULL,
    updated_at          TIMESTAMP NOT NULL,
    deleted_at          TIMESTAMP,
    user_id             TEXT NOT NULL REFERENCES user(id),
    data                TEXT NOT NULL
);

CREATE TABLE user_app (
    user_id             TEXT NOT NULL REFERENCES user(id),
    app_id              TEXT NOT NULL REFERENCES app(id),
    PRIMARY KEY (user_id, app_id)
);
