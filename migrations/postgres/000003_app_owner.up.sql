BEGIN;

ALTER TABLE app ADD COLUMN owner_user_id TEXT REFERENCES "user"(id);
ALTER TABLE app ADD COLUMN credential_ids JSONB;
UPDATE app SET owner_user_id = (SELECT user_id FROM user_app WHERE app_id = app.id LIMIT 1);
UPDATE app SET credential_ids = jsonb_build_array(owner_user_id);
ALTER TABLE app ALTER COLUMN owner_user_id SET NOT NULL;
ALTER TABLE app ALTER COLUMN credential_ids SET NOT NULL;

CREATE INDEX app_credentials ON app USING gin(credential_ids);

INSERT INTO user_credential (id, created_at, updated_at, deleted_at, user_id, data)
    SELECT
        id,
        NOW() as created_at,
        NOW() as updated_at,
        NULL AS deleted_at,
        id AS user_id,
        '{}' as data
    FROM "user" WHERE TRUE;

DROP TABLE user_app;

COMMIT;
