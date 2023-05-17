CREATE TABLE app (
    id                  TEXT NOT NULL PRIMARY KEY,
    created_at          TIMESTAMP NOT NULL,
    updated_at          TIMESTAMP NOT NULL,
    deleted_at          TIMESTAMP
);

CREATE TABLE environment (
    app_id              TEXT NOT NULL REFERENCES app(id),
    name                TEXT NOT NULL,
    created_at          TIMESTAMP NOT NULL,
    updated_at          TIMESTAMP NOT NULL,
    deleted_at          TIMESTAMP,
    PRIMARY KEY (app_id, name)
);

CREATE TABLE deployment (
    id                  TEXT NOT NULL PRIMARY KEY,
    created_at          TIMESTAMP NOT NULL,
    updated_at          TIMESTAMP NOT NULL,
    deleted_at          TIMESTAMP,
    app_id              TEXT NOT NULL,
    environment_name    TEXT NOT NULL,
    status              TEXT NOT NULL,
    storage_key_prefix  TEXT NOT NULL,
    metadata            TEXT,
    FOREIGN KEY (app_id, environment_name) REFERENCES environment(app_id, name)
);

CREATE TABLE site (
    id                  TEXT NOT NULL PRIMARY KEY,
    app_id              TEXT NOT NULL,
    environment_name    TEXT NOT NULL,
    deployment_id       TEXT NOT NULL REFERENCES deployment(id),
    FOREIGN KEY (app_id, environment_name) REFERENCES environment(app_id, name)
);