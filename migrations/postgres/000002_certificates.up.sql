CREATE TABLE cert_data (
    key                 TEXT NOT NULL PRIMARY KEY,
    updated_at          TIMESTAMPTZ NOT NULL,
    value               TEXT NOT NULL
);
