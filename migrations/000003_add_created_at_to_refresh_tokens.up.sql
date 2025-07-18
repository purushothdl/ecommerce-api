-- migrations/000003_add_created_at_to_refresh_tokens.up.sql
ALTER TABLE refresh_tokens 
ADD COLUMN created_at timestamp(0) with time zone NOT NULL DEFAULT NOW();

-- Backfill existing records with estimated created_at
UPDATE refresh_tokens 
SET created_at = expires_at - INTERVAL '7 days';
