ALTER TABLE app ADD COLUMN owner_user_id TEXT REFERENCES user(id);
ALTER TABLE app ADD COLUMN credential_ids TEXT DEFAULT '[]';
UPDATE app SET owner_user_id = (SELECT user_id FROM user_app WHERE app_id = app.id LIMIT 1);
UPDATE app SET credential_ids = json_array(owner_user_id);

INSERT INTO user_credential (id, created_at, updated_at, deleted_at, user_id, data)
    SELECT
        id,
        DATETIME('now') as created_at,
        DATETIME('now') as updated_at,
        NULL AS deleted_at,
        id AS user_id,
        '{}' as data
    FROM user WHERE TRUE;

DROP TABLE user_app;