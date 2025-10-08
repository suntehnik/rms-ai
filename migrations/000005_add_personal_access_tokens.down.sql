-- Remove Personal Access Tokens table

-- Drop the trigger first
DROP TRIGGER IF EXISTS update_personal_access_tokens_updated_at ON personal_access_tokens;

-- Drop indexes
DROP INDEX IF EXISTS idx_pat_user_id;
DROP INDEX IF EXISTS idx_pat_prefix;
DROP INDEX IF EXISTS idx_pat_expires_at;
DROP INDEX IF EXISTS idx_pat_last_used_at;
DROP INDEX IF EXISTS idx_pat_created_at;

-- Drop the table
DROP TABLE IF EXISTS personal_access_tokens;