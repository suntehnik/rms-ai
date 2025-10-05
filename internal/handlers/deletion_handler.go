package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"product-requirements-management/internal/service"
)

// DeletionHandler handles HTTP requests for deletion operations
type DeletionHandler struct {
	deletionService service.DeletionService
	logger          *logrus.Logger
}

// NewDeletionHandler creates a new deletion handler
func NewDeletionHandler(deletionService service.DeletionService, logger *logrus.Logger) *DeletionHandler {
	return &DeletionHandler{
		deletionService: deletionService,
		logger:          logger,
	}
}

// DeleteEpicRequest represents the request body for epic deletion
type DeleteEpicRequest struct {
	Force bool `json:"force" binding:"omitempty"`
}

// DeleteUserStoryRequest represents the request body for user story deletion
type DeleteUserStoryRequest struct {
	Force bool `json:"force" binding:"omitempty"`
}

// DeleteAcceptanceCriteriaRequest represents the request body for acceptance criteria deletion
type DeleteAcceptanceCriteriaRequest struct {
	Force bool `json:"force" binding:"omitempty"`
}

// DeleteRequirementRequest represents the request body for requirement deletion
type DeleteRequirementRequest struct {
	Force bool `json:"force" binding:"omitempty"`
}

// ValidateEpicDeletion validates if an epic can be deleted
//
//	@Summary		Validate epic deletion
//	@Description	Validates if an epic can be deleted and returns dependency information
//	@Tags			deletion
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Epic ID (UUID or reference ID)"
//	@Success		200	{object}	service.DependencyInfo
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/api/epics/{id}/validate-deletion [get]
func (h *DeletionHandler) ValidateEpicDeletion(c *gin.Context) {
	idParam := c.Param("id")

	// Try to parse as UUID first, then as reference ID
	var epicID uuid.UUID
	var err error

	if epicID, err = uuid.Parse(idParam); err != nil {
		// If not a valid UUID, treat as reference ID and look up the epic
		// This would require an epic service to look up by reference ID
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_ID",
				Message: "Invalid epic ID format",
			},
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"epic_id": epicID,
		"action":  "validate_deletion",
	}).Info("Validating epic deletion")

	depInfo, err := h.deletionService.ValidateEpicDeletion(epicID)
	if err != nil {
		if err == service.ErrEpicNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: ErrorDetail{
					Code:    "EPIC_NOT_FOUND",
					Message: "Epic not found",
				},
			})
			return
		}

		h.logger.WithFields(logrus.Fields{
			"epic_id": epicID,
			"error":   err.Error(),
		}).Error("Failed to validate epic deletion")

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{
				Code:    "VALIDATION_FAILED",
				Message: "Failed to validate epic deletion",
			},
		})
		return
	}

	c.JSON(http.StatusOK, depInfo)
}

// DeleteEpic deletes an epic with validation and cascading
//
//	@Summary		Delete epic
//	@Description	Deletes an epic with comprehensive validation and cascading deletion
//	@Tags			deletion
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string				true	"Epic ID (UUID or reference ID)"
//	@Param			request	body		DeleteEpicRequest	false	"Deletion options"
//	@Success		200		{object}	service.DeletionResult
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		409		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/api/epics/{id}/delete [delete]
func (h *DeletionHandler) DeleteEpic(c *gin.Context) {
	idParam := c.Param("id")

	// Try to parse as UUID first
	var epicID uuid.UUID
	var err error

	if epicID, err = uuid.Parse(idParam); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_ID",
				Message: "Invalid epic ID format",
			},
		})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: ErrorDetail{
				Code:    "UNAUTHORIZED",
				Message: "User not authenticated",
			},
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_USER_ID",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	var req DeleteEpicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// If no body provided, use default values
		req.Force = false
	}

	h.logger.WithFields(logrus.Fields{
		"epic_id": epicID,
		"user_id": userUUID,
		"force":   req.Force,
		"action":  "delete",
	}).Info("Deleting epic")

	result, err := h.deletionService.DeleteEpicWithValidation(epicID, userUUID, req.Force)
	if err != nil {
		switch err {
		case service.ErrEpicNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: ErrorDetail{
					Code:    "EPIC_NOT_FOUND",
					Message: "Epic not found",
				},
			})
		case service.ErrDeletionValidationFailed:
			c.JSON(http.StatusConflict, ErrorResponse{
				Error: ErrorDetail{
					Code:    "DELETION_BLOCKED",
					Message: "Epic cannot be deleted due to dependencies. Use force=true to override.",
				},
			})
		case service.ErrDeletionTransactionFailed:
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: ErrorDetail{
					Code:    "DELETION_FAILED",
					Message: "Failed to delete epic due to transaction error",
				},
			})
		default:
			h.logger.WithFields(logrus.Fields{
				"epic_id": epicID,
				"error":   err.Error(),
			}).Error("Failed to delete epic")

			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: ErrorDetail{
					Code:    "DELETION_FAILED",
					Message: "Failed to delete epic",
				},
			})
		}
		return
	}

	h.logger.WithFields(logrus.Fields{
		"epic_id":        epicID,
		"transaction_id": result.TransactionID,
		"cascade_count":  len(result.CascadeDeleted),
	}).Info("Epic deleted successfully")

	c.JSON(http.StatusOK, result)
}

// ValidateUserStoryDeletion validates if a user story can be deleted
//
//	@Summary		Validate user story deletion
//	@Description	Validates if a user story can be deleted and returns dependency information
//	@Tags			deletion
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"User Story ID (UUID or reference ID)"
//	@Success		200	{object}	service.DependencyInfo
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/api/user-stories/{id}/validate-deletion [get]
func (h *DeletionHandler) ValidateUserStoryDeletion(c *gin.Context) {
	idParam := c.Param("id")

	var userStoryID uuid.UUID
	var err error

	if userStoryID, err = uuid.Parse(idParam); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_ID",
				Message: "Invalid user story ID format",
			},
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_story_id": userStoryID,
		"action":        "validate_deletion",
	}).Info("Validating user story deletion")

	depInfo, err := h.deletionService.ValidateUserStoryDeletion(userStoryID)
	if err != nil {
		if err == service.ErrUserStoryNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: ErrorDetail{
					Code:    "USER_STORY_NOT_FOUND",
					Message: "User story not found",
				},
			})
			return
		}

		h.logger.WithFields(logrus.Fields{
			"user_story_id": userStoryID,
			"error":         err.Error(),
		}).Error("Failed to validate user story deletion")

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{
				Code:    "VALIDATION_FAILED",
				Message: "Failed to validate user story deletion",
			},
		})
		return
	}

	c.JSON(http.StatusOK, depInfo)
}

// DeleteUserStory deletes a user story with validation and cascading
//
//	@Summary		Delete user story
//	@Description	Deletes a user story with comprehensive validation and cascading deletion
//	@Tags			deletion
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"User Story ID (UUID or reference ID)"
//	@Param			request	body		DeleteUserStoryRequest	false	"Deletion options"
//	@Success		200		{object}	service.DeletionResult
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		409		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/api/user-stories/{id}/delete [delete]
func (h *DeletionHandler) DeleteUserStory(c *gin.Context) {
	idParam := c.Param("id")

	var userStoryID uuid.UUID
	var err error

	if userStoryID, err = uuid.Parse(idParam); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_ID",
				Message: "Invalid user story ID format",
			},
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: ErrorDetail{
				Code:    "UNAUTHORIZED",
				Message: "User not authenticated",
			},
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_USER_ID",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	var req DeleteUserStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Force = false
	}

	h.logger.WithFields(logrus.Fields{
		"user_story_id": userStoryID,
		"user_id":       userUUID,
		"force":         req.Force,
		"action":        "delete",
	}).Info("Deleting user story")

	result, err := h.deletionService.DeleteUserStoryWithValidation(userStoryID, userUUID, req.Force)
	if err != nil {
		switch err {
		case service.ErrUserStoryNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: ErrorDetail{
					Code:    "USER_STORY_NOT_FOUND",
					Message: "User story not found",
				},
			})
		case service.ErrDeletionValidationFailed:
			c.JSON(http.StatusConflict, ErrorResponse{
				Error: ErrorDetail{
					Code:    "DELETION_BLOCKED",
					Message: "User story cannot be deleted due to dependencies. Use force=true to override.",
				},
			})
		case service.ErrDeletionTransactionFailed:
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: ErrorDetail{
					Code:    "DELETION_FAILED",
					Message: "Failed to delete user story due to transaction error",
				},
			})
		default:
			h.logger.WithFields(logrus.Fields{
				"user_story_id": userStoryID,
				"error":         err.Error(),
			}).Error("Failed to delete user story")

			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: ErrorDetail{
					Code:    "DELETION_FAILED",
					Message: "Failed to delete user story",
				},
			})
		}
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_story_id":  userStoryID,
		"transaction_id": result.TransactionID,
		"cascade_count":  len(result.CascadeDeleted),
	}).Info("User story deleted successfully")

	c.JSON(http.StatusOK, result)
}

// ValidateAcceptanceCriteriaDeletion validates if acceptance criteria can be deleted
//
//	@Summary		Validate acceptance criteria deletion
//	@Description	Validates if acceptance criteria can be deleted and returns dependency information
//	@Tags			deletion
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Acceptance Criteria ID (UUID or reference ID)"
//	@Success		200	{object}	service.DependencyInfo
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/api/acceptance-criteria/{id}/validate-deletion [get]
func (h *DeletionHandler) ValidateAcceptanceCriteriaDeletion(c *gin.Context) {
	idParam := c.Param("id")

	var acceptanceCriteriaID uuid.UUID
	var err error

	if acceptanceCriteriaID, err = uuid.Parse(idParam); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_ID",
				Message: "Invalid acceptance criteria ID format",
			},
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"acceptance_criteria_id": acceptanceCriteriaID,
		"action":                 "validate_deletion",
	}).Info("Validating acceptance criteria deletion")

	depInfo, err := h.deletionService.ValidateAcceptanceCriteriaDeletion(acceptanceCriteriaID)
	if err != nil {
		if err == service.ErrAcceptanceCriteriaNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: ErrorDetail{
					Code:    "ACCEPTANCE_CRITERIA_NOT_FOUND",
					Message: "Acceptance criteria not found",
				},
			})
			return
		}

		h.logger.WithFields(logrus.Fields{
			"acceptance_criteria_id": acceptanceCriteriaID,
			"error":                  err.Error(),
		}).Error("Failed to validate acceptance criteria deletion")

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{
				Code:    "VALIDATION_FAILED",
				Message: "Failed to validate acceptance criteria deletion",
			},
		})
		return
	}

	c.JSON(http.StatusOK, depInfo)
}

// DeleteAcceptanceCriteria deletes acceptance criteria with validation and cascading
//
//	@Summary		Delete acceptance criteria
//	@Description	Deletes acceptance criteria with comprehensive validation and cascading deletion
//	@Tags			deletion
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Acceptance Criteria ID (UUID or reference ID)"
//	@Param			request	body		DeleteAcceptanceCriteriaRequest	false	"Deletion options"
//	@Success		200		{object}	service.DeletionResult
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		409		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/api/acceptance-criteria/{id}/delete [delete]
func (h *DeletionHandler) DeleteAcceptanceCriteria(c *gin.Context) {
	idParam := c.Param("id")

	var acceptanceCriteriaID uuid.UUID
	var err error

	if acceptanceCriteriaID, err = uuid.Parse(idParam); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_ID",
				Message: "Invalid acceptance criteria ID format",
			},
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: ErrorDetail{
				Code:    "UNAUTHORIZED",
				Message: "User not authenticated",
			},
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_USER_ID",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	var req DeleteAcceptanceCriteriaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Force = false
	}

	h.logger.WithFields(logrus.Fields{
		"acceptance_criteria_id": acceptanceCriteriaID,
		"user_id":                userUUID,
		"force":                  req.Force,
		"action":                 "delete",
	}).Info("Deleting acceptance criteria")

	result, err := h.deletionService.DeleteAcceptanceCriteriaWithValidation(acceptanceCriteriaID, userUUID, req.Force)
	if err != nil {
		switch err {
		case service.ErrAcceptanceCriteriaNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: ErrorDetail{
					Code:    "ACCEPTANCE_CRITERIA_NOT_FOUND",
					Message: "Acceptance criteria not found",
				},
			})
		case service.ErrDeletionValidationFailed:
			c.JSON(http.StatusConflict, ErrorResponse{
				Error: ErrorDetail{
					Code:    "DELETION_BLOCKED",
					Message: "Acceptance criteria cannot be deleted due to dependencies. Use force=true to override.",
				},
			})
		case service.ErrDeletionTransactionFailed:
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: ErrorDetail{
					Code:    "DELETION_FAILED",
					Message: "Failed to delete acceptance criteria due to transaction error",
				},
			})
		default:
			h.logger.WithFields(logrus.Fields{
				"acceptance_criteria_id": acceptanceCriteriaID,
				"error":                  err.Error(),
			}).Error("Failed to delete acceptance criteria")

			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: ErrorDetail{
					Code:    "DELETION_FAILED",
					Message: "Failed to delete acceptance criteria",
				},
			})
		}
		return
	}

	h.logger.WithFields(logrus.Fields{
		"acceptance_criteria_id": acceptanceCriteriaID,
		"transaction_id":         result.TransactionID,
		"cascade_count":          len(result.CascadeDeleted),
	}).Info("Acceptance criteria deleted successfully")

	c.JSON(http.StatusOK, result)
}

// ValidateRequirementDeletion validates if a requirement can be deleted
//
//	@Summary		Validate requirement deletion
//	@Description	Validates if a requirement can be deleted and returns dependency information
//	@Tags			deletion
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Requirement ID (UUID or reference ID)"
//	@Success		200	{object}	service.DependencyInfo
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/api/requirements/{id}/validate-deletion [get]
func (h *DeletionHandler) ValidateRequirementDeletion(c *gin.Context) {
	idParam := c.Param("id")

	var requirementID uuid.UUID
	var err error

	if requirementID, err = uuid.Parse(idParam); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_ID",
				Message: "Invalid requirement ID format",
			},
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"requirement_id": requirementID,
		"action":         "validate_deletion",
	}).Info("Validating requirement deletion")

	depInfo, err := h.deletionService.ValidateRequirementDeletion(requirementID)
	if err != nil {
		if err == service.ErrRequirementNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: ErrorDetail{
					Code:    "REQUIREMENT_NOT_FOUND",
					Message: "Requirement not found",
				},
			})
			return
		}

		h.logger.WithFields(logrus.Fields{
			"requirement_id": requirementID,
			"error":          err.Error(),
		}).Error("Failed to validate requirement deletion")

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{
				Code:    "VALIDATION_FAILED",
				Message: "Failed to validate requirement deletion",
			},
		})
		return
	}

	c.JSON(http.StatusOK, depInfo)
}

// DeleteRequirement deletes a requirement with validation and cascading
//
//	@Summary		Delete requirement
//	@Description	Deletes a requirement with comprehensive validation and cascading deletion
//	@Tags			deletion
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Requirement ID (UUID or reference ID)"
//	@Param			request	body		DeleteRequirementRequest	false	"Deletion options"
//	@Success		200		{object}	service.DeletionResult
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		409		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/api/requirements/{id}/delete [delete]
func (h *DeletionHandler) DeleteRequirement(c *gin.Context) {
	idParam := c.Param("id")

	var requirementID uuid.UUID
	var err error

	if requirementID, err = uuid.Parse(idParam); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_ID",
				Message: "Invalid requirement ID format",
			},
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: ErrorDetail{
				Code:    "UNAUTHORIZED",
				Message: "User not authenticated",
			},
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_USER_ID",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	var req DeleteRequirementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Force = false
	}

	h.logger.WithFields(logrus.Fields{
		"requirement_id": requirementID,
		"user_id":        userUUID,
		"force":          req.Force,
		"action":         "delete",
	}).Info("Deleting requirement")

	result, err := h.deletionService.DeleteRequirementWithValidation(requirementID, userUUID, req.Force)
	if err != nil {
		switch err {
		case service.ErrRequirementNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: ErrorDetail{
					Code:    "REQUIREMENT_NOT_FOUND",
					Message: "Requirement not found",
				},
			})
		case service.ErrDeletionValidationFailed:
			c.JSON(http.StatusConflict, ErrorResponse{
				Error: ErrorDetail{
					Code:    "DELETION_BLOCKED",
					Message: "Requirement cannot be deleted due to dependencies. Use force=true to override.",
				},
			})
		case service.ErrDeletionTransactionFailed:
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: ErrorDetail{
					Code:    "DELETION_FAILED",
					Message: "Failed to delete requirement due to transaction error",
				},
			})
		default:
			h.logger.WithFields(logrus.Fields{
				"requirement_id": requirementID,
				"error":          err.Error(),
			}).Error("Failed to delete requirement")

			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: ErrorDetail{
					Code:    "DELETION_FAILED",
					Message: "Failed to delete requirement",
				},
			})
		}
		return
	}

	h.logger.WithFields(logrus.Fields{
		"requirement_id": requirementID,
		"transaction_id": result.TransactionID,
		"cascade_count":  len(result.CascadeDeleted),
	}).Info("Requirement deleted successfully")

	c.JSON(http.StatusOK, result)
}

// GetDeletionConfirmation provides a confirmation dialog with dependency information
//
//	@Summary		Get deletion confirmation
//	@Description	Provides detailed information for user confirmation before deletion
//	@Tags			deletion
//	@Accept			json
//	@Produce		json
//	@Param			entity_type	query		string	true	"Entity type (epic, user_story, acceptance_criteria, requirement)"
//	@Param			id			query		string	true	"Entity ID (UUID)"
//	@Success		200			{object}	service.DependencyInfo
//	@Failure		400			{object}	ErrorResponse
//	@Failure		401			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/api/deletion/confirm [get]
func (h *DeletionHandler) GetDeletionConfirmation(c *gin.Context) {
	entityType := c.Query("entity_type")
	idParam := c.Query("id")

	if entityType == "" || idParam == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "MISSING_PARAMETERS",
				Message: "entity_type and id parameters are required",
			},
		})
		return
	}

	entityID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_ID",
				Message: "Invalid entity ID format",
			},
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"entity_type": entityType,
		"entity_id":   entityID,
		"action":      "get_confirmation",
	}).Info("Getting deletion confirmation")

	var depInfo *service.DependencyInfo

	switch entityType {
	case "epic":
		depInfo, err = h.deletionService.ValidateEpicDeletion(entityID)
	case "user_story":
		depInfo, err = h.deletionService.ValidateUserStoryDeletion(entityID)
	case "acceptance_criteria":
		depInfo, err = h.deletionService.ValidateAcceptanceCriteriaDeletion(entityID)
	case "requirement":
		depInfo, err = h.deletionService.ValidateRequirementDeletion(entityID)
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_ENTITY_TYPE",
				Message: "Invalid entity type. Must be one of: epic, user_story, acceptance_criteria, requirement",
			},
		})
		return
	}

	if err != nil {
		switch err {
		case service.ErrEpicNotFound, service.ErrUserStoryNotFound,
			service.ErrAcceptanceCriteriaNotFound, service.ErrRequirementNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: ErrorDetail{
					Code:    "ENTITY_NOT_FOUND",
					Message: "Entity not found",
				},
			})
		default:
			h.logger.WithFields(logrus.Fields{
				"entity_type": entityType,
				"entity_id":   entityID,
				"error":       err.Error(),
			}).Error("Failed to validate deletion")

			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: ErrorDetail{
					Code:    "VALIDATION_FAILED",
					Message: "Failed to validate deletion",
				},
			})
		}
		return
	}

	c.JSON(http.StatusOK, depInfo)
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail represents error details
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
