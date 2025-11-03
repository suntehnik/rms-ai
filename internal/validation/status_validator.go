package validation

import (
	"strings"

	"product-requirements-management/internal/models"
)

// StatusValidator defines the interface for validating entity status values
type StatusValidator interface {
	// ValidateEpicStatus validates if the provided status is valid for epics
	ValidateEpicStatus(status string) error

	// ValidateUserStoryStatus validates if the provided status is valid for user stories
	ValidateUserStoryStatus(status string) error

	// ValidateRequirementStatus validates if the provided status is valid for requirements
	ValidateRequirementStatus(status string) error
}

// statusValidator implements the StatusValidator interface
type statusValidator struct{}

// NewStatusValidator creates a new instance of StatusValidator
func NewStatusValidator() StatusValidator {
	return &statusValidator{}
}

// ValidateEpicStatus validates if the provided status is valid for epics
func (v *statusValidator) ValidateEpicStatus(status string) error {
	if status == "" {
		return NewStatusValidationError("epic", status, GetValidEpicStatuses())
	}

	// Convert to EpicStatus type for validation (case-insensitive)
	epicStatus := models.EpicStatus(normalizeStatus(status))

	validStatuses := []models.EpicStatus{
		models.EpicStatusBacklog,
		models.EpicStatusDraft,
		models.EpicStatusInProgress,
		models.EpicStatusDone,
		models.EpicStatusCancelled,
	}

	for _, validStatus := range validStatuses {
		if epicStatus == validStatus {
			return nil
		}
	}

	return NewStatusValidationError("epic", status, GetValidEpicStatuses())
}

// ValidateUserStoryStatus validates if the provided status is valid for user stories
func (v *statusValidator) ValidateUserStoryStatus(status string) error {
	if status == "" {
		return NewStatusValidationError("user story", status, GetValidUserStoryStatuses())
	}

	// Convert to UserStoryStatus type for validation (case-insensitive)
	userStoryStatus := models.UserStoryStatus(normalizeStatus(status))

	validStatuses := []models.UserStoryStatus{
		models.UserStoryStatusBacklog,
		models.UserStoryStatusDraft,
		models.UserStoryStatusInProgress,
		models.UserStoryStatusDone,
		models.UserStoryStatusCancelled,
	}

	for _, validStatus := range validStatuses {
		if userStoryStatus == validStatus {
			return nil
		}
	}

	return NewStatusValidationError("user story", status, GetValidUserStoryStatuses())
}

// ValidateRequirementStatus validates if the provided status is valid for requirements
func (v *statusValidator) ValidateRequirementStatus(status string) error {
	if status == "" {
		return NewStatusValidationError("requirement", status, GetValidRequirementStatuses())
	}

	// Convert to RequirementStatus type for validation (case-insensitive)
	requirementStatus := models.RequirementStatus(normalizeStatus(status))

	validStatuses := []models.RequirementStatus{
		models.RequirementStatusDraft,
		models.RequirementStatusActive,
		models.RequirementStatusObsolete,
	}

	for _, validStatus := range validStatuses {
		if requirementStatus == validStatus {
			return nil
		}
	}

	return NewStatusValidationError("requirement", status, GetValidRequirementStatuses())
}

// normalizeStatus normalizes the status string for case-insensitive comparison
// It preserves the original casing of valid statuses while allowing case-insensitive input
func normalizeStatus(status string) string {
	// Trim whitespace
	status = strings.TrimSpace(status)

	// Handle common case variations
	switch strings.ToLower(status) {
	case "backlog":
		return "Backlog"
	case "draft":
		return "Draft"
	case "in progress", "inprogress", "in_progress":
		return "In Progress"
	case "done":
		return "Done"
	case "cancelled", "canceled":
		return "Cancelled"
	case "active":
		return "Active"
	case "obsolete":
		return "Obsolete"
	default:
		// Return as-is if no normalization rule matches
		return status
	}
}

// Helper functions to get valid status lists for error messages

func GetValidEpicStatuses() []string {
	return []string{
		string(models.EpicStatusBacklog),
		string(models.EpicStatusDraft),
		string(models.EpicStatusInProgress),
		string(models.EpicStatusDone),
		string(models.EpicStatusCancelled),
	}
}

func GetValidUserStoryStatuses() []string {
	return []string{
		string(models.UserStoryStatusBacklog),
		string(models.UserStoryStatusDraft),
		string(models.UserStoryStatusInProgress),
		string(models.UserStoryStatusDone),
		string(models.UserStoryStatusCancelled),
	}
}

func GetValidRequirementStatuses() []string {
	return []string{
		string(models.RequirementStatusDraft),
		string(models.RequirementStatusActive),
		string(models.RequirementStatusObsolete),
	}
}
