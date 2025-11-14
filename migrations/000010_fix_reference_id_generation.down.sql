-- Revert to original simple functions with LPAD(3)

CREATE OR REPLACE FUNCTION get_next_epic_ref_id() RETURNS VARCHAR(20) AS $$
BEGIN
    RETURN 'EP-' || LPAD(nextval('epic_ref_seq')::TEXT, 3, '0');
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION get_next_user_story_ref_id() RETURNS VARCHAR(20) AS $$
BEGIN
    RETURN 'US-' || LPAD(nextval('user_story_ref_seq')::TEXT, 3, '0');
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION get_next_acceptance_criteria_ref_id() RETURNS VARCHAR(20) AS $$
BEGIN
    RETURN 'AC-' || LPAD(nextval('acceptance_criteria_ref_seq')::TEXT, 3, '0');
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION get_next_requirement_ref_id() RETURNS VARCHAR(20) AS $$
BEGIN
    RETURN 'REQ-' || LPAD(nextval('requirement_ref_seq')::TEXT, 3, '0');
END;
$$ LANGUAGE plpgsql;
