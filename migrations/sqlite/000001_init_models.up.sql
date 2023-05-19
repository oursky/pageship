CREATE TABLE app (
    id                  TEXT NOT NULL PRIMARY KEY,
    created_at          TIMESTAMP NOT NULL,
    updated_at          TIMESTAMP NOT NULL,
    deleted_at          TIMESTAMP,
    config              TEXT NOT NULL
);

CREATE TABLE site (
    id                  TEXT NOT NULL PRIMARY KEY,
    app_id              TEXT NOT NULL REFERENCES app(id),
    name                TEXT NOT NULL,
    created_at          TIMESTAMP NOT NULL,
    updated_at          TIMESTAMP NOT NULL,
    deleted_at          TIMESTAMP
);
CREATE UNIQUE INDEX site_key ON site(app_id, name) WHERE deleted_at IS NULL;

CREATE TABLE deployment (
    id                  TEXT NOT NULL PRIMARY KEY,
    created_at          TIMESTAMP NOT NULL,
    updated_at          TIMESTAMP NOT NULL,
    deleted_at          TIMESTAMP,
    app_id              TEXT NOT NULL,
    site_id             TEXT NOT NULL REFERENCES site(id),
    storage_key_prefix  TEXT NOT NULL,
    metadata            TEXT,
    uploaded_at         TIMESTAMP
);

CREATE TABLE site_deployment (
    site_id             TEXT NOT NULL PRIMARY KEY REFERENCES site(id) ON DELETE CASCADE,
    deployment_id       TEXT NOT NULL REFERENCES deployment(id)
);
