CREATE TABLE user_app (
    user_id             TEXT NOT NULL REFERENCES user(id),
    app_id              TEXT NOT NULL REFERENCES app(id),
    PRIMARY KEY (user_id, app_id)
);
INSERT INTO user_app SELECT owner_user_id as user_id, id as app_id FROM app WHERE TRUE;
ALTER TABLE app DROP COLUMN owner_user_id;
ALTER TABLE app DROP COLUMN credential_ids;
DELETE FROM user_credential WHERE id = user_id;
