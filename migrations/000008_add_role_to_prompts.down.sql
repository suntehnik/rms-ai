-- Rollback migration for adding role field to prompts table

-- Drop the index
DROP INDEX IF EXISTS idx_prompts_role;

-- Drop the check constraint
ALTER TABLE prompts DROP CONSTRAINT IF EXISTS check_prompt_role;

-- Drop the role column
ALTER TABLE prompts DROP COLUMN IF EXISTS role;