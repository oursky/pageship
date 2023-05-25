BEGIN;

CREATE TABLE app (
    id                  TEXT NOT NULL PRIMARY KEY,
    created_at          TIMESTAMPTZ NOT NULL,
    updated_at          TIMESTAMPTZ NOT NULL,
    deleted_at          TIMESTAMPTZ,
    config              JSONB NOT NULL
);

CREATE TABLE deployment (
    id                  TEXT NOT NULL PRIMARY KEY,
    created_at          TIMESTAMPTZ NOT NULL,
    updated_at          TIMESTAMPTZ NOT NULL,
    deleted_at          TIMESTAMPTZ,
    name                TEXT NOT NULL,
    app_id              TEXT NOT NULL REFERENCES app(id),
    storage_key_prefix  TEXT NOT NULL,
    metadata            JSONB,
    uploaded_at         TIMESTAMPTZ,
    expire_at           TIMESTAMPTZ
);
CREATE INDEX deployment_order ON deployment(app_id, created_at) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX deployment_key ON deployment(app_id, name) WHERE deleted_at IS NULL;
CREATE INDEX deployment_expiry ON deployment(expire_at) WHERE deleted_at IS NULL;

CREATE TABLE site (
    id                  TEXT NOT NULL PRIMARY KEY,
    app_id              TEXT NOT NULL REFERENCES app(id),
    name                TEXT NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL,
    updated_at          TIMESTAMPTZ NOT NULL,
    deleted_at          TIMESTAMPTZ,
    deployment_id       TEXT REFERENCES deployment(id)
);
CREATE UNIQUE INDEX site_key ON site(app_id, name) WHERE deleted_at IS NULL;

CREATE TABLE "user" (
    id                  TEXT NOT NULL PRIMARY KEY,
    created_at          TIMESTAMPTZ NOT NULL,
    updated_at          TIMESTAMPTZ NOT NULL,
    deleted_at          TIMESTAMPTZ,
    name                TEXT
);
CREATE INDEX user_order ON "user"(name) WHERE deleted_at IS NULL;

CREATE TABLE user_credential (
    id                  TEXT NOT NULL PRIMARY KEY,
    created_at          TIMESTAMPTZ NOT NULL,
    updated_at          TIMESTAMPTZ NOT NULL,
    deleted_at          TIMESTAMPTZ,
    user_id             TEXT NOT NULL REFERENCES "user"(id),
    data                JSONB NOT NULL
);

CREATE TABLE user_app (
    user_id             TEXT NOT NULL REFERENCES "user"(id),
    app_id              TEXT NOT NULL REFERENCES app(id),
    PRIMARY KEY (user_id, app_id)
);

COMMIT;
