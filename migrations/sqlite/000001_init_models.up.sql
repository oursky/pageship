CREATE TABLE app (
    id                  TEXT NOT NULL PRIMARY KEY,
    created_at          TIMESTAMP NOT NULL,
    updated_at          TIMESTAMP NOT NULL,
    deleted_at          TIMESTAMP,
    config              TEXT NOT NULL
);

CREATE TABLE deployment (
    id                  TEXT NOT NULL PRIMARY KEY,
    created_at          TIMESTAMP NOT NULL,
    updated_at          TIMESTAMP NOT NULL,
    deleted_at          TIMESTAMP,
    name                TEXT NOT NULL,
    app_id              TEXT NOT NULL REFERENCES app(id),
    storage_key_prefix  TEXT NOT NULL,
    metadata            TEXT,
    uploaded_at         TIMESTAMP
);
CREATE INDEX deployment_order ON deployment(app_id, created_at) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX deployment_key ON deployment(app_id, name) WHERE deleted_at IS NULL;

CREATE TABLE site (
    id                  TEXT NOT NULL PRIMARY KEY,
    app_id              TEXT NOT NULL REFERENCES app(id),
    name                TEXT NOT NULL,
    created_at          TIMESTAMP NOT NULL,
    updated_at          TIMESTAMP NOT NULL,
    deleted_at          TIMESTAMP,
    deployment_id       TEXT REFERENCES deployment(id)
);
CREATE UNIQUE INDEX site_key ON site(app_id, name) WHERE deleted_at IS NULL;
