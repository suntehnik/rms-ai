-- Initial schema for Product Requirements Management System

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('Administrator', 'User', 'Commenter')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for users
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);

-- Reference ID sequences
CREATE SEQUENCE epic_ref_seq START 1;
CREATE SEQUENCE user_story_ref_seq START 1;
CREATE SEQUENCE acceptance_criteria_ref_seq START 1;
CREATE SEQUENCE requirement_ref_seq START 1;

-- Epics table
CREATE TABLE epics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reference_id VARCHAR(20) UNIQUE NOT NULL DEFAULT ('EP-' || LPAD(nextval('epic_ref_seq')::TEXT, 3, '0')),
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    assignee_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_modified TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    priority INTEGER NOT NULL CHECK (priority BETWEEN 1 AND 4),
    status VARCHAR(50) NOT NULL DEFAULT 'Backlog',
    title VARCHAR(500) NOT NULL,
    description TEXT
);

-- Create indexes for epics
CREATE INDEX idx_epics_creator ON epics(creator_id);
CREATE INDEX idx_epics_assignee ON epics(assignee_id);
CREATE INDEX idx_epics_status ON epics(status);
CREATE INDEX idx_epics_priority ON epics(priority);
CREATE INDEX idx_epics_reference ON epics(reference_id);
CREATE INDEX idx_epics_created_at ON epics(created_at);
CREATE INDEX idx_epics_last_modified ON epics(last_modified);

-- Full-text search index for epics
CREATE INDEX idx_epics_search ON epics USING gin(to_tsvector('english', reference_id || ' ' || title || ' ' || COALESCE(description, '')));

-- User Stories table
CREATE TABLE user_stories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reference_id VARCHAR(20) UNIQUE NOT NULL DEFAULT ('US-' || LPAD(nextval('user_story_ref_seq')::TEXT, 3, '0')),
    epic_id UUID NOT NULL REFERENCES epics(id) ON DELETE CASCADE,
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    assignee_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_modified TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    priority INTEGER NOT NULL CHECK (priority BETWEEN 1 AND 4),
    status VARCHAR(50) NOT NULL DEFAULT 'Backlog',
    title VARCHAR(500) NOT NULL,
    description TEXT
);

-- Create indexes for user_stories
CREATE INDEX idx_user_stories_epic ON user_stories(epic_id);
CREATE INDEX idx_user_stories_creator ON user_stories(creator_id);
CREATE INDEX idx_user_stories_assignee ON user_stories(assignee_id);
CREATE INDEX idx_user_stories_status ON user_stories(status);
CREATE INDEX idx_user_stories_priority ON user_stories(priority);
CREATE INDEX idx_user_stories_reference ON user_stories(reference_id);
CREATE INDEX idx_user_stories_created_at ON user_stories(created_at);
CREATE INDEX idx_user_stories_last_modified ON user_stories(last_modified);

-- Full-text search index for user_stories
CREATE INDEX idx_user_stories_search ON user_stories USING gin(to_tsvector('english', reference_id || ' ' || title || ' ' || COALESCE(description, '')));

-- Acceptance Criteria table
CREATE TABLE acceptance_criteria (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reference_id VARCHAR(20) UNIQUE NOT NULL DEFAULT ('AC-' || LPAD(nextval('acceptance_criteria_ref_seq')::TEXT, 3, '0')),
    user_story_id UUID NOT NULL REFERENCES user_stories(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_modified TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    description TEXT NOT NULL
);

-- Create indexes for acceptance_criteria
CREATE INDEX idx_acceptance_criteria_user_story ON acceptance_criteria(user_story_id);
CREATE INDEX idx_acceptance_criteria_author ON acceptance_criteria(author_id);
CREATE INDEX idx_acceptance_criteria_reference ON acceptance_criteria(reference_id);
CREATE INDEX idx_acceptance_criteria_created_at ON acceptance_criteria(created_at);

-- Requirement Types table (configurable dictionary)
CREATE TABLE requirement_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert default requirement types
INSERT INTO requirement_types (name, description) VALUES
    ('Functional', 'Functional requirements that describe what the system should do'),
    ('Non-Functional', 'Non-functional requirements that describe how the system should behave'),
    ('Business Rule', 'Business rules and constraints'),
    ('Interface', 'Interface and integration requirements'),
    ('Data', 'Data and information requirements');

-- Relationship Types table (configurable dictionary)
CREATE TABLE relationship_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert default relationship types
INSERT INTO relationship_types (name, description) VALUES
    ('depends_on', 'This requirement depends on another requirement'),
    ('blocks', 'This requirement blocks another requirement'),
    ('relates_to', 'This requirement is related to another requirement'),
    ('conflicts_with', 'This requirement conflicts with another requirement'),
    ('derives_from', 'This requirement is derived from another requirement');

-- Requirements table
CREATE TABLE requirements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reference_id VARCHAR(20) UNIQUE NOT NULL DEFAULT ('REQ-' || LPAD(nextval('requirement_ref_seq')::TEXT, 3, '0')),
    user_story_id UUID NOT NULL REFERENCES user_stories(id) ON DELETE CASCADE,
    acceptance_criteria_id UUID REFERENCES acceptance_criteria(id) ON DELETE SET NULL,
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    assignee_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_modified TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    priority INTEGER NOT NULL CHECK (priority BETWEEN 1 AND 4),
    status VARCHAR(50) NOT NULL DEFAULT 'Draft',
    type_id UUID NOT NULL REFERENCES requirement_types(id) ON DELETE RESTRICT,
    title VARCHAR(500) NOT NULL,
    description TEXT
);

-- Create indexes for requirements
CREATE INDEX idx_requirements_user_story ON requirements(user_story_id);
CREATE INDEX idx_requirements_acceptance_criteria ON requirements(acceptance_criteria_id);
CREATE INDEX idx_requirements_creator ON requirements(creator_id);
CREATE INDEX idx_requirements_assignee ON requirements(assignee_id);
CREATE INDEX idx_requirements_status ON requirements(status);
CREATE INDEX idx_requirements_priority ON requirements(priority);
CREATE INDEX idx_requirements_type ON requirements(type_id);
CREATE INDEX idx_requirements_reference ON requirements(reference_id);
CREATE INDEX idx_requirements_created_at ON requirements(created_at);
CREATE INDEX idx_requirements_last_modified ON requirements(last_modified);

-- Full-text search index for requirements
CREATE INDEX idx_requirements_search ON requirements USING gin(to_tsvector('english', reference_id || ' ' || title || ' ' || COALESCE(description, '')));

-- Requirement Relationships table
CREATE TABLE requirement_relationships (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_requirement_id UUID NOT NULL REFERENCES requirements(id) ON DELETE CASCADE,
    target_requirement_id UUID NOT NULL REFERENCES requirements(id) ON DELETE CASCADE,
    relationship_type_id UUID NOT NULL REFERENCES relationship_types(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    UNIQUE(source_requirement_id, target_requirement_id, relationship_type_id)
);

-- Create indexes for requirement_relationships
CREATE INDEX idx_req_relationships_source ON requirement_relationships(source_requirement_id);
CREATE INDEX idx_req_relationships_target ON requirement_relationships(target_requirement_id);
CREATE INDEX idx_req_relationships_type ON requirement_relationships(relationship_type_id);

-- Comments table
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type VARCHAR(50) NOT NULL CHECK (entity_type IN ('epic', 'user_story', 'acceptance_criteria', 'requirement')),
    entity_id UUID NOT NULL,
    parent_comment_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    content TEXT NOT NULL,
    is_resolved BOOLEAN DEFAULT FALSE,
    -- For inline comments
    linked_text TEXT,
    text_position_start INTEGER,
    text_position_end INTEGER
);

-- Create indexes for comments
CREATE INDEX idx_comments_entity ON comments(entity_type, entity_id);
CREATE INDEX idx_comments_parent ON comments(parent_comment_id);
CREATE INDEX idx_comments_author ON comments(author_id);
CREATE INDEX idx_comments_resolved ON comments(is_resolved);
CREATE INDEX idx_comments_created_at ON comments(created_at);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at columns
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_epics_last_modified BEFORE UPDATE ON epics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_stories_last_modified BEFORE UPDATE ON user_stories FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_acceptance_criteria_last_modified BEFORE UPDATE ON acceptance_criteria FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_requirements_last_modified BEFORE UPDATE ON requirements FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_requirement_types_updated_at BEFORE UPDATE ON requirement_types FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_relationship_types_updated_at BEFORE UPDATE ON relationship_types FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_comments_updated_at BEFORE UPDATE ON comments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();