-- Drop all tables and sequences in reverse order

-- Drop triggers first
DROP TRIGGER IF EXISTS update_comments_updated_at ON comments;
DROP TRIGGER IF EXISTS update_relationship_types_updated_at ON relationship_types;
DROP TRIGGER IF EXISTS update_requirement_types_updated_at ON requirement_types;
DROP TRIGGER IF EXISTS update_requirements_last_modified ON requirements;
DROP TRIGGER IF EXISTS update_acceptance_criteria_last_modified ON acceptance_criteria;
DROP TRIGGER IF EXISTS update_user_stories_last_modified ON user_stories;
DROP TRIGGER IF EXISTS update_epics_last_modified ON epics;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS requirement_relationships;
DROP TABLE IF EXISTS requirements;
DROP TABLE IF EXISTS relationship_types;
DROP TABLE IF EXISTS requirement_types;
DROP TABLE IF EXISTS acceptance_criteria;
DROP TABLE IF EXISTS user_stories;
DROP TABLE IF EXISTS epics;
DROP TABLE IF EXISTS users;

-- Drop sequences
DROP SEQUENCE IF EXISTS requirement_ref_seq;
DROP SEQUENCE IF EXISTS acceptance_criteria_ref_seq;
DROP SEQUENCE IF EXISTS user_story_ref_seq;
DROP SEQUENCE IF EXISTS epic_ref_seq;

-- Drop UUID extension (only if no other tables use it)
-- DROP EXTENSION IF EXISTS "uuid-ossp";