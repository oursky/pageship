CREATE TABLE domain_verification (
    id                  TEXT NOT NULL PRIMARY KEY,
    created_at          TIMESTAMP NOT NULL,
    updated_at          TIMESTAMP NOT NULL,
    deleted_at          TIMESTAMP,
    domain              TEXT NOT NULL,
    domain_prefix       TEXT NOT NULL,
    value               TEXT NOT NULL,
    app_id              TEXT NOT NULL REFERENCES app(id),
    verified_at         TIMESTAMP,
    last_checked_at     TIMESTAMP,
    will_check_at       TIMESTAMP
);
CREATE UNIQUE INDEX app_domain_mapping ON domain_verification(app_id, domain) WHERE deleted_at IS NULL;
