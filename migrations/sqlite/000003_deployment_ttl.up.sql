ALTER TABLE deployment ADD COLUMN expire_at TIMESTAMP;
CREATE INDEX deployment_expiry ON deployment(expire_at) WHERE deleted_at IS NULL;
