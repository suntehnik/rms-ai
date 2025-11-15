-- Revert reference ID generation functions to previous version
-- Note: This will restore the old functions but will NOT revert sequence values

-- Restore old epic reference ID function (with advisory locks)
CREATE OR REPLACE FUNCTION get_next_epic_ref_id() RETURNS VARCHAR(20) AS $
DECLARE
    next_id BIGINT;
    lock_acquired BOOLEAN;
BEGIN
    -- Try to acquire advisory lock
    lock_acquired := pg_try_advisory_xact_lock(2147483647);
    
    IF NOT lock_acquired THEN
        -- Fallback to UUID-based reference ID if lock cannot be acquired
        RETURN 'EP-' || SUBSTRING(gen_random_uuid()::TEXT FROM 1 FOR 8);
    END IF;
    
    -- Get next value from sequence
    next_id := nextval('epic_ref_seq');
    
    -- Format with zero-padding (EP-001, EP-002, etc.)
    RETURN 'EP-' || LPAD(next_id::TEXT, 3, '0');
END;
$ LANGUAGE plpgsql;

-- Restore old user story reference ID function (with advisory locks)
CREATE OR REPLACE FUNCTION get_next_user_story_ref_id() RETURNS VARCHAR(20) AS $
DECLARE
    next_id BIGINT;
    lock_acquired BOOLEAN;
BEGIN
    lock_acquired := pg_try_advisory_xact_lock(2147483646);
    
    IF NOT lock_acquired THEN
        RETURN 'US-' || SUBSTRING(gen_random_uuid()::TEXT FROM 1 FOR 8);
    END IF;
    
    next_id := nextval('user_story_ref_seq');
    RETURN 'US-' || LPAD(next_id::TEXT, 3, '0');
END;
$ LANGUAGE plpgsql;

-- Restore old acceptance criteria reference ID function (with advisory locks)
CREATE OR REPLACE FUNCTION get_next_acceptance_criteria_ref_id() RETURNS VARCHAR(20) AS $
DECLARE
    next_id BIGINT;
    lock_acquired BOOLEAN;
BEGIN
    lock_acquired := pg_try_advisory_xact_lock(2147483644);
    
    IF NOT lock_acquired THEN
        RETURN 'AC-' || SUBSTRING(gen_random_uuid()::TEXT FROM 1 FOR 8);
    END IF;
    
    next_id := nextval('acceptance_criteria_ref_seq');
    RETURN 'AC-' || LPAD(next_id::TEXT, 3, '0');
END;
$ LANGUAGE plpgsql;

-- Restore old requirement reference ID function (with advisory locks)
CREATE OR REPLACE FUNCTION get_next_requirement_ref_id() RETURNS VARCHAR(20) AS $
DECLARE
    next_id BIGINT;
    lock_acquired BOOLEAN;
BEGIN
    lock_acquired := pg_try_advisory_xact_lock(2147483645);
    
    IF NOT lock_acquired THEN
        RETURN 'REQ-' || SUBSTRING(gen_random_uuid()::TEXT FROM 1 FOR 8);
    END IF;
    
    next_id := nextval('requirement_ref_seq');
    RETURN 'REQ-' || LPAD(next_id::TEXT, 3, '0');
END;
$ LANGUAGE plpgsql;
