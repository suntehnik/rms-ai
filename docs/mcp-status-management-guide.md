# MCP Status Management Guide

## Overview

The MCP API provides comprehensive status management capabilities for epics, user stories, and requirements. This guide covers how to effectively use status parameters in MCP tools to manage entity workflows and lifecycle transitions.

## Table of Contents

1. [Status Management Overview](#status-management-overview)
2. [Entity Status Workflows](#entity-status-workflows)
3. [Using Status Parameters](#using-status-parameters)
4. [Error Handling](#error-handling)
5. [Best Practices](#best-practices)
6. [Common Workflows](#common-workflows)
7. [Troubleshooting](#troubleshooting)

## Status Management Overview

### Supported Entities

Status management is available for three entity types:

| Entity Type | Update Tool | Status Parameter | Valid Statuses |
|-------------|-------------|------------------|----------------|
| Epic | `update_epic` | `status` | Backlog, Draft, In Progress, Done, Cancelled |
| User Story | `update_user_story` | `status` | Backlog, Draft, In Progress, Done, Cancelled |
| Requirement | `update_requirement` | `status` | Draft, Active, Obsolete |

### Key Features

- **Optional Parameter**: Status is always optional in update tools
- **Validation**: All status values are validated against allowed enums
- **Case Sensitive**: Status values must match exactly (case-sensitive)
- **Immediate Effect**: Status changes take effect immediately
- **Audit Trail**: All status changes are logged with timestamps

## Entity Status Workflows

### Epic Status Workflow

```
┌─────────┐    ┌───────┐    ┌─────────────┐    ┌──────┐
│ Backlog │───▶│ Draft │───▶│ In Progress │───▶│ Done │
└─────────┘    └───────┘    └─────────────┘    └──────┘
     │             │               │
     ▼             ▼               ▼
┌───────────────────────────────────────────────────────┐
│                  Cancelled                            │
└───────────────────────────────────────────────────────┘
```

**Status Descriptions:**
- **Backlog**: Initial state for new epics, awaiting prioritization
- **Draft**: Epic is being planned and refined
- **In Progress**: Epic is actively being worked on
- **Done**: Epic is completed successfully
- **Cancelled**: Epic is cancelled and will not be completed

### User Story Status Workflow

```
┌─────────┐    ┌───────┐    ┌─────────────┐    ┌──────┐
│ Backlog │───▶│ Draft │───▶│ In Progress │───▶│ Done │
└─────────┘    └───────┘    └─────────────┘    └──────┘
     │             │               │
     ▼             ▼               ▼
┌───────────────────────────────────────────────────────┐
│                  Cancelled                            │
└───────────────────────────────────────────────────────┘
```

**Status Descriptions:**
- **Backlog**: Initial state for new user stories
- **Draft**: User story is being refined and detailed
- **In Progress**: User story is actively being developed
- **Done**: User story is completed and tested
- **Cancelled**: User story is cancelled

### Requirement Status Workflow

```
┌───────┐    ┌────────┐    ┌──────────┐
│ Draft │───▶│ Active │───▶│ Obsolete │
└───────┘    └────────┘    └──────────┘
```

**Status Descriptions:**
- **Draft**: Initial state for new requirements, under review
- **Active**: Requirement is approved and being implemented
- **Obsolete**: Requirement is no longer valid or needed

## Using Status Parameters

### Basic Status Update

Update an entity's status using the appropriate update tool:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "update_epic",
    "arguments": {
      "epic_id": "EP-001",
      "status": "In Progress"
    }
  }
}
```

### Combined Updates

You can update status along with other fields in a single request:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "update_user_story",
    "arguments": {
      "user_story_id": "US-001",
      "title": "Updated User Story Title",
      "priority": 1,
      "status": "In Progress",
      "assignee_id": "123e4567-e89b-12d3-a456-426614174000"
    }
  }
}
```

### Status-Only Updates

For workflow management, you often only need to update status:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "update_requirement",
    "arguments": {
      "requirement_id": "REQ-001",
      "status": "Active"
    }
  }
}
```

## Error Handling

### Invalid Status Values

When an invalid status is provided, the API returns a validation error:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": "Invalid status 'InvalidStatus' for epic. Valid statuses are: Backlog, Draft, In Progress, Done, Cancelled"
  }
}
```

### Entity Not Found

If the entity doesn't exist, you'll receive an entity not found error:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": "Epic not found"
  }
}
```

### Case Sensitivity

Status values are case-sensitive. Using incorrect casing will result in validation errors:

```json
// ❌ Incorrect - will fail
{
  "status": "in progress"  // Should be "In Progress"
}

// ✅ Correct
{
  "status": "In Progress"
}
```

## Best Practices

### 1. Validate Status Values Client-Side

Implement client-side validation to catch errors early:

```python
def validate_epic_status(status: str) -> bool:
    valid_statuses = ["Backlog", "Draft", "In Progress", "Done", "Cancelled"]
    return status in valid_statuses

def validate_user_story_status(status: str) -> bool:
    valid_statuses = ["Backlog", "Draft", "In Progress", "Done", "Cancelled"]
    return status in valid_statuses

def validate_requirement_status(status: str) -> bool:
    valid_statuses = ["Draft", "Active", "Obsolete"]
    return status in valid_statuses
```

### 2. Implement Workflow Logic

Consider implementing workflow validation to ensure logical status transitions:

```python
def can_transition_epic_status(current_status: str, new_status: str) -> bool:
    """Check if epic status transition is logically valid"""
    valid_transitions = {
        "Backlog": ["Draft", "Cancelled"],
        "Draft": ["In Progress", "Cancelled"],
        "In Progress": ["Done", "Cancelled"],
        "Done": [],  # Terminal state
        "Cancelled": []  # Terminal state
    }
    
    allowed = valid_transitions.get(current_status, [])
    return new_status in allowed
```

### 3. Handle Status Updates Atomically

When updating multiple related entities, consider the order of operations:

```python
def complete_user_story_workflow(user_story_id: str):
    """Complete a user story and check if epic can be completed"""
    
    # 1. Complete the user story
    update_user_story_status(user_story_id, "Done")
    
    # 2. Check if all user stories in epic are complete
    epic_id = get_epic_for_user_story(user_story_id)
    if all_user_stories_complete(epic_id):
        update_epic_status(epic_id, "Done")
```

### 4. Provide User Feedback

Always provide clear feedback about status changes:

```python
def update_with_feedback(entity_type: str, entity_id: str, status: str):
    try:
        result = update_entity_status(entity_type, entity_id, status)
        if result["success"]:
            print(f"✅ {result['message']}")
            return result
        else:
            print(f"❌ Failed to update {entity_type} {entity_id}: {result['error']}")
            return result
    except Exception as e:
        print(f"❌ Unexpected error: {str(e)}")
        return {"success": False, "error": str(e)}
```

### 5. Log Status Changes

Implement logging for audit trails:

```python
import logging

def log_status_change(entity_type: str, entity_id: str, old_status: str, new_status: str, user_id: str):
    logging.info(f"Status change: {entity_type} {entity_id} from '{old_status}' to '{new_status}' by user {user_id}")
```

## Common Workflows

### 1. Feature Development Workflow

```python
def execute_feature_development_workflow(epic_id: str, user_story_ids: List[str], requirement_ids: List[str]):
    """Execute a complete feature development workflow"""
    
    # Phase 1: Planning
    print("📋 Phase 1: Planning")
    update_entity_status("epic", epic_id, "Draft")
    
    for us_id in user_story_ids:
        update_entity_status("user_story", us_id, "Draft")
    
    for req_id in requirement_ids:
        update_entity_status("requirement", req_id, "Draft")
    
    # Phase 2: Development
    print("🔨 Phase 2: Development")
    update_entity_status("epic", epic_id, "In Progress")
    
    for us_id in user_story_ids:
        update_entity_status("user_story", us_id, "In Progress")
    
    for req_id in requirement_ids:
        update_entity_status("requirement", req_id, "Active")
    
    # Phase 3: Completion
    print("✅ Phase 3: Completion")
    for us_id in user_story_ids:
        update_entity_status("user_story", us_id, "Done")
    
    update_entity_status("epic", epic_id, "Done")
```

### 2. Requirement Lifecycle Management

```python
def manage_requirement_lifecycle(requirement_id: str):
    """Manage complete requirement lifecycle"""
    
    # Start as draft
    update_entity_status("requirement", requirement_id, "Draft")
    print("📝 Requirement created in Draft status")
    
    # Review and approve
    if requirement_approved(requirement_id):
        update_entity_status("requirement", requirement_id, "Active")
        print("✅ Requirement approved and activated")
    
    # Mark obsolete when no longer needed
    if requirement_obsolete(requirement_id):
        update_entity_status("requirement", requirement_id, "Obsolete")
        print("🗑️ Requirement marked as obsolete")
```

### 3. Sprint Planning Workflow

```python
def plan_sprint(user_story_ids: List[str]):
    """Move user stories from backlog to in progress for sprint"""
    
    print("🏃 Starting Sprint Planning")
    
    for us_id in user_story_ids:
        # Move from Backlog to In Progress
        result = update_entity_status("user_story", us_id, "In Progress")
        
        if result["success"]:
            print(f"✅ {us_id} added to sprint")
        else:
            print(f"❌ Failed to add {us_id} to sprint: {result['error']}")
```

## Troubleshooting

### Common Issues

#### 1. Case Sensitivity Errors

**Problem**: Status update fails with validation error
**Cause**: Incorrect casing in status value
**Solution**: Use exact case-sensitive values

```python
# ❌ Wrong
update_entity_status("epic", "EP-001", "in progress")

# ✅ Correct  
update_entity_status("epic", "EP-001", "In Progress")
```

#### 2. Invalid Status for Entity Type

**Problem**: Using requirement status for epic
**Cause**: Mixing up status values between entity types
**Solution**: Use correct status values for each entity type

```python
# ❌ Wrong - "Active" is not valid for epics
update_entity_status("epic", "EP-001", "Active")

# ✅ Correct - Use epic-specific status
update_entity_status("epic", "EP-001", "In Progress")
```

#### 3. Entity Not Found

**Problem**: Status update fails with "not found" error
**Cause**: Invalid entity ID or reference ID
**Solution**: Verify entity exists and ID is correct

```python
# Verify entity exists before updating
def safe_status_update(entity_type: str, entity_id: str, status: str):
    if entity_exists(entity_type, entity_id):
        return update_entity_status(entity_type, entity_id, status)
    else:
        return {"success": False, "error": f"{entity_type} {entity_id} not found"}
```

### Debugging Tips

1. **Check Entity Existence**: Verify the entity exists before updating status
2. **Validate Status Values**: Use client-side validation to catch errors early
3. **Check Permissions**: Ensure user has permission to update the entity
4. **Review Error Messages**: API error messages provide specific guidance
5. **Test with Simple Cases**: Start with basic status updates before complex workflows

### Getting Help

If you encounter issues with status management:

1. Check the error message for specific guidance
2. Verify status values against the valid enums
3. Ensure entity IDs are correct (UUID or reference ID format)
4. Review the API logs for detailed error information
5. Test with a simple status update to isolate the issue

## Summary

Status management in the MCP API provides powerful workflow capabilities:

- **Three Entity Types**: Epic, User Story, and Requirement status management
- **Flexible Updates**: Status can be updated alone or with other fields
- **Validation**: Comprehensive validation with helpful error messages
- **Workflow Support**: Enables complete development lifecycle management
- **Audit Trail**: All changes are logged with timestamps

By following the best practices and examples in this guide, you can effectively implement status management workflows that enhance your requirements management processes.