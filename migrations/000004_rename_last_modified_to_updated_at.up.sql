-- Migration: Rename last_modified column to updated_at for consistency
-- This migration renames the last_modified column to updated_at in four core tables
-- to achieve naming consistency across the entire schema.

-- Rename columns in all affected tables
ALTER TABLE epics RENAME COLUMN last_modified TO updated_at;
ALTER TABLE user_stories RENAME COLUMN last_modified TO updated_at;
ALTER TABLE acceptance_criteria RENAME COLUMN last_modified TO updated_at;
ALTER TABLE requirements RENAME COLUMN last_modified TO updated_at;

-- Update trigger names for improved naming consistency
-- Drop existing triggers with old names
DROP TRIGGER IF EXISTS update_epics_last_modified ON epics;
DROP TRIGGER IF EXISTS update_user_stories_last_modified ON user_stories;
DROP TRIGGER IF EXISTS update_acceptance_criteria_last_modified ON acceptance_criteria;
DROP TRIGGER IF EXISTS update_requirements_last_modified ON requirements;

-- Create new triggers with updated names that reflect the new column name
CREATE TRIGGER update_epics_updated_at BEFORE UPDATE ON epics 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_stories_updated_at BEFORE UPDATE ON user_stories 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_acceptance_criteria_updated_at BEFORE UPDATE ON acceptance_criteria 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_requirements_updated_at BEFORE UPDATE ON requirements 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Note: Indexes are automatically renamed by PostgreSQL when columns are renamed
-- The following indexes will be automatically updated:
-- - idx_epics_last_modified -> idx_epics_updated_at (automatically handled)
-- - idx_user_stories_last_modified -> idx_user_stories_updated_at (automatically handled)
-- - idx_requirements_last_modified -> idx_requirements_updated_at (automatically handled)
-- 
-- Full-text search indexes that reference the column will continue to work
-- as they use the column reference, not the column name directly.