-- Add status model system tables

-- Status Models table
CREATE TABLE status_models (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type VARCHAR(50) NOT NULL CHECK (entity_type IN ('epic', 'user_story', 'acceptance_criteria', 'requirement')),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(entity_type, name)
);

-- Create indexes for status_models
CREATE INDEX idx_status_models_entity_type ON status_models(entity_type);
CREATE INDEX idx_status_models_is_default ON status_models(is_default);
CREATE INDEX idx_status_models_name ON status_models(name);

-- Statuses table
CREATE TABLE statuses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    status_model_id UUID NOT NULL REFERENCES status_models(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    color VARCHAR(7), -- Hex color code
    is_initial BOOLEAN NOT NULL DEFAULT FALSE,
    is_final BOOLEAN NOT NULL DEFAULT FALSE,
    "order" INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(status_model_id, name)
);

-- Create indexes for statuses
CREATE INDEX idx_statuses_status_model ON statuses(status_model_id);
CREATE INDEX idx_statuses_name ON statuses(name);
CREATE INDEX idx_statuses_is_initial ON statuses(is_initial);
CREATE INDEX idx_statuses_is_final ON statuses(is_final);
CREATE INDEX idx_statuses_order ON statuses("order");

-- Status Transitions table
CREATE TABLE status_transitions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    status_model_id UUID NOT NULL REFERENCES status_models(id) ON DELETE CASCADE,
    from_status_id UUID NOT NULL REFERENCES statuses(id) ON DELETE CASCADE,
    to_status_id UUID NOT NULL REFERENCES statuses(id) ON DELETE CASCADE,
    name VARCHAR(255),
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(status_model_id, from_status_id, to_status_id)
);

-- Create indexes for status_transitions
CREATE INDEX idx_status_transitions_status_model ON status_transitions(status_model_id);
CREATE INDEX idx_status_transitions_from_status ON status_transitions(from_status_id);
CREATE INDEX idx_status_transitions_to_status ON status_transitions(to_status_id);

-- Add triggers for updated_at columns
CREATE TRIGGER update_status_models_updated_at BEFORE UPDATE ON status_models FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_statuses_updated_at BEFORE UPDATE ON statuses FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_status_transitions_updated_at BEFORE UPDATE ON status_transitions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default status models and statuses

-- Epic Status Model
INSERT INTO status_models (entity_type, name, description, is_default) VALUES
    ('epic', 'Default Epic Workflow', 'Default status workflow for epics', true);

-- Get the epic status model ID
DO $$
DECLARE
    epic_model_id UUID;
    backlog_status_id UUID;
    draft_status_id UUID;
    in_progress_status_id UUID;
    done_status_id UUID;
    cancelled_status_id UUID;
BEGIN
    SELECT id INTO epic_model_id FROM status_models WHERE entity_type = 'epic' AND is_default = true;
    
    -- Insert epic statuses
    INSERT INTO statuses (status_model_id, name, description, color, is_initial, is_final, "order") VALUES
        (epic_model_id, 'Backlog', 'Epic is in the backlog', '#6c757d', true, false, 1),
        (epic_model_id, 'Draft', 'Epic is being drafted', '#ffc107', false, false, 2),
        (epic_model_id, 'In Progress', 'Epic is in progress', '#007bff', false, false, 3),
        (epic_model_id, 'Done', 'Epic is completed', '#28a745', false, true, 4),
        (epic_model_id, 'Cancelled', 'Epic has been cancelled', '#dc3545', false, true, 5);
    
    -- Get status IDs for transitions
    SELECT id INTO backlog_status_id FROM statuses WHERE status_model_id = epic_model_id AND name = 'Backlog';
    SELECT id INTO draft_status_id FROM statuses WHERE status_model_id = epic_model_id AND name = 'Draft';
    SELECT id INTO in_progress_status_id FROM statuses WHERE status_model_id = epic_model_id AND name = 'In Progress';
    SELECT id INTO done_status_id FROM statuses WHERE status_model_id = epic_model_id AND name = 'Done';
    SELECT id INTO cancelled_status_id FROM statuses WHERE status_model_id = epic_model_id AND name = 'Cancelled';
    
    -- Insert default transitions (allow all transitions for now)
    INSERT INTO status_transitions (status_model_id, from_status_id, to_status_id, name) VALUES
        (epic_model_id, backlog_status_id, draft_status_id, 'Start Draft'),
        (epic_model_id, backlog_status_id, in_progress_status_id, 'Start Work'),
        (epic_model_id, backlog_status_id, cancelled_status_id, 'Cancel'),
        (epic_model_id, draft_status_id, backlog_status_id, 'Return to Backlog'),
        (epic_model_id, draft_status_id, in_progress_status_id, 'Start Work'),
        (epic_model_id, draft_status_id, cancelled_status_id, 'Cancel'),
        (epic_model_id, in_progress_status_id, done_status_id, 'Complete'),
        (epic_model_id, in_progress_status_id, cancelled_status_id, 'Cancel'),
        (epic_model_id, done_status_id, in_progress_status_id, 'Reopen'),
        (epic_model_id, cancelled_status_id, backlog_status_id, 'Reactivate');
END $$;

-- User Story Status Model
INSERT INTO status_models (entity_type, name, description, is_default) VALUES
    ('user_story', 'Default User Story Workflow', 'Default status workflow for user stories', true);

-- Get the user story status model ID and insert statuses
DO $$
DECLARE
    us_model_id UUID;
    backlog_status_id UUID;
    draft_status_id UUID;
    in_progress_status_id UUID;
    done_status_id UUID;
    cancelled_status_id UUID;
BEGIN
    SELECT id INTO us_model_id FROM status_models WHERE entity_type = 'user_story' AND is_default = true;
    
    -- Insert user story statuses
    INSERT INTO statuses (status_model_id, name, description, color, is_initial, is_final, "order") VALUES
        (us_model_id, 'Backlog', 'User story is in the backlog', '#6c757d', true, false, 1),
        (us_model_id, 'Draft', 'User story is being drafted', '#ffc107', false, false, 2),
        (us_model_id, 'In Progress', 'User story is in progress', '#007bff', false, false, 3),
        (us_model_id, 'Done', 'User story is completed', '#28a745', false, true, 4),
        (us_model_id, 'Cancelled', 'User story has been cancelled', '#dc3545', false, true, 5);
    
    -- Get status IDs for transitions
    SELECT id INTO backlog_status_id FROM statuses WHERE status_model_id = us_model_id AND name = 'Backlog';
    SELECT id INTO draft_status_id FROM statuses WHERE status_model_id = us_model_id AND name = 'Draft';
    SELECT id INTO in_progress_status_id FROM statuses WHERE status_model_id = us_model_id AND name = 'In Progress';
    SELECT id INTO done_status_id FROM statuses WHERE status_model_id = us_model_id AND name = 'Done';
    SELECT id INTO cancelled_status_id FROM statuses WHERE status_model_id = us_model_id AND name = 'Cancelled';
    
    -- Insert default transitions
    INSERT INTO status_transitions (status_model_id, from_status_id, to_status_id, name) VALUES
        (us_model_id, backlog_status_id, draft_status_id, 'Start Draft'),
        (us_model_id, backlog_status_id, in_progress_status_id, 'Start Work'),
        (us_model_id, backlog_status_id, cancelled_status_id, 'Cancel'),
        (us_model_id, draft_status_id, backlog_status_id, 'Return to Backlog'),
        (us_model_id, draft_status_id, in_progress_status_id, 'Start Work'),
        (us_model_id, draft_status_id, cancelled_status_id, 'Cancel'),
        (us_model_id, in_progress_status_id, done_status_id, 'Complete'),
        (us_model_id, in_progress_status_id, cancelled_status_id, 'Cancel'),
        (us_model_id, done_status_id, in_progress_status_id, 'Reopen'),
        (us_model_id, cancelled_status_id, backlog_status_id, 'Reactivate');
END $$;

-- Requirement Status Model
INSERT INTO status_models (entity_type, name, description, is_default) VALUES
    ('requirement', 'Default Requirement Workflow', 'Default status workflow for requirements', true);

-- Get the requirement status model ID and insert statuses
DO $$
DECLARE
    req_model_id UUID;
    draft_status_id UUID;
    active_status_id UUID;
    obsolete_status_id UUID;
BEGIN
    SELECT id INTO req_model_id FROM status_models WHERE entity_type = 'requirement' AND is_default = true;
    
    -- Insert requirement statuses
    INSERT INTO statuses (status_model_id, name, description, color, is_initial, is_final, "order") VALUES
        (req_model_id, 'Draft', 'Requirement is being drafted', '#ffc107', true, false, 1),
        (req_model_id, 'Active', 'Requirement is active', '#28a745', false, false, 2),
        (req_model_id, 'Obsolete', 'Requirement is obsolete', '#6c757d', false, true, 3);
    
    -- Get status IDs for transitions
    SELECT id INTO draft_status_id FROM statuses WHERE status_model_id = req_model_id AND name = 'Draft';
    SELECT id INTO active_status_id FROM statuses WHERE status_model_id = req_model_id AND name = 'Active';
    SELECT id INTO obsolete_status_id FROM statuses WHERE status_model_id = req_model_id AND name = 'Obsolete';
    
    -- Insert default transitions
    INSERT INTO status_transitions (status_model_id, from_status_id, to_status_id, name) VALUES
        (req_model_id, draft_status_id, active_status_id, 'Activate'),
        (req_model_id, draft_status_id, obsolete_status_id, 'Mark Obsolete'),
        (req_model_id, active_status_id, obsolete_status_id, 'Mark Obsolete'),
        (req_model_id, obsolete_status_id, active_status_id, 'Reactivate');
END $$;