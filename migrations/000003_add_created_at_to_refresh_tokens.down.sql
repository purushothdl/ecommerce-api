-- migrations/000003_add_created_at_to_refresh_tokens.down.sql
ALTER TABLE refresh_tokens DROP COLUMN created_at;