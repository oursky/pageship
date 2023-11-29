BEGIN;

CREATE TABLE domain_association (
    id                  TEXT NOT NULL PRIMARY KEY,
    created_at          TIMESTAMPTZ NOT NULL,
    updated_at          TIMESTAMPTZ NOT NULL,
    deleted_at          TIMESTAMPTZ,
    domain              TEXT NOT NULL,
    app_id              TEXT NOT NULL REFERENCES app(id),
    site_name           TEXT NOT NULL
);
CREATE UNIQUE INDEX domain_name ON domain_association(domain) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX domain_mapping ON domain_association(app_id, site_name) WHERE deleted_at IS NULL;

COMMIT;
