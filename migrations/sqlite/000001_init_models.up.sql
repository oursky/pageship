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
    status              TEXT NOT NULL,
    storage_key_prefix  TEXT NOT NULL,
    metadata            TEXT
);

CREATE TABLE site_deployment (
    id                  TEXT NOT NULL PRIMARY KEY,
    app_id              TEXT NOT NULL,
    site_name           TEXT NOT NULL,
    deployment_id       TEXT NOT NULL REFERENCES deployment(id),
    FOREIGN KEY (app_id, site_name) REFERENCES site(app_id, name)
);
