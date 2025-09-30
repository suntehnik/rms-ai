-- Migration Rollback: Rename updated_at column back to last_modified
-- This rollback migration reverts the column rename from updated_at back to last_modified
-- and restores the original trigger names.

-- Rename columns back to original names
ALTER TABLE epics RENAME COLUMN updated_at TO last_modified;
ALTER TABLE user_stories RENAME COLUMN updated_at TO last_modified;
ALTER TABLE acceptance_criteria RENAME COLUMN updated_at TO last_modified;
ALTER TABLE requirements RENAME COLUMN updated_at TO last_modified;

-- Restore original trigger names
-- Drop new triggers with updated names
DROP TRIGGER IF EXISTS update_epics_updated_at ON epics;
DROP TRIGGER IF EXISTS update_user_stories_updated_at ON user_stories;
DROP TRIGGER IF EXISTS update_acceptance_criteria_updated_at ON acceptance_criteria;
DROP TRIGGER IF EXISTS update_requirements_updated_at ON requirements;

-- Recreate original triggers with last_modified names
CREATE TRIGGER update_epics_last_modified BEFORE UPDATE ON epics 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_stories_last_modified BEFORE UPDATE ON user_stories 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_acceptance_criteria_last_modified BEFORE UPDATE ON acceptance_criteria 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_requirements_last_modified BEFORE UPDATE ON requirements 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Note: Indexes are automatically renamed back by PostgreSQL when columns are renamed
-- The following indexes will be automatically reverted:
-- - idx_epics_updated_at -> idx_epics_last_modified (automatically handled)
-- - idx_user_stories_updated_at -> idx_user_stories_last_modified (automatically handled)
-- - idx_requirements_updated_at -> idx_requirements_last_modified (automatically handled)
-- 
-- Full-text search indexes will continue to work as they reference the column, not the name.