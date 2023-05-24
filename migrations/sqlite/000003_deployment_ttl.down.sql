ALTER TABLE deployment DROP COLUMN expire_at TIMESTAMP;
DROP INDEX deployment_expiry;
