CREATE TABLE cert_data (
    key                 TEXT NOT NULL PRIMARY KEY,
    updated_at          TIMESTAMP NOT NULL,
    value               TEXT NOT NULL
);

CREATE TABLE cert_lock (
    name                TEXT NOT NULL PRIMARY KEY,
    holder              TEXT NOT NULL,
    release_at          TIMESTAMP NOT NULL
);
