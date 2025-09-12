package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
)

var (
	ErrDeletionCancelled         = errors.New("deletion cancelled by user")
	ErrDeletionValidationFailed  = errors.New("deletion validation failed")
	ErrDeletionTransactionFailed = errors.New("deletion transaction failed")
)

// DeletionService defines the interface for comprehensive deletion operations
type DeletionService interface {
	// Epic deletion
	DeleteEpicWithValidation(id uuid.UUID, userID uuid.UUID, force bool) (*DeletionResult, error)

	// User Story deletion
	DeleteUserStoryWithValidation(id uuid.UUID, userID uuid.UUID, force bool) (*DeletionResult, error)

	// Acceptance Criteria deletion
	DeleteAcceptanceCriteriaWithValidation(id uuid.UUID, userID uuid.UUID, force bool) (*DeletionResult, error)

	// Requirement deletion
	DeleteRequirementWithValidation(id uuid.UUID, userID uuid.UUID, force bool) (*DeletionResult, error)

	// Dependency validation
	ValidateEpicDeletion(id uuid.UUID) (*DependencyInfo, error)
	ValidateUserStoryDeletion(id uuid.UUID) (*DependencyInfo, error)
	ValidateAcceptanceCriteriaDeletion(id uuid.UUID) (*DependencyInfo, error)
	ValidateRequirementDeletion(id uuid.UUID) (*DependencyInfo, error)
}

// DeletionResult represents the result of a deletion operation
type DeletionResult struct {
	EntityType     string                 `json:"entity_type"`
	EntityID       uuid.UUID              `json:"entity_id"`
	ReferenceID    string                 `json:"reference_id"`
	DeletedAt      time.Time              `json:"deleted_at"`
	DeletedBy      uuid.UUID              `json:"deleted_by"`
	CascadeDeleted []CascadeDeletedEntity `json:"cascade_deleted,omitempty"`
	AuditLogID     uuid.UUID              `json:"audit_log_id"`
	TransactionID  string                 `json:"transaction_id"`
}

// CascadeDeletedEntity represents an entity that was deleted as part of cascade
type CascadeDeletedEntity struct {
	EntityType  string    `json:"entity_type"`
	EntityID    uuid.UUID `json:"entity_id"`
	ReferenceID string    `json:"reference_id"`
}

// DependencyInfo represents information about dependencies that would prevent deletion
type DependencyInfo struct {
	CanDelete             bool                   `json:"can_delete"`
	Dependencies          []DependencyDetail     `json:"dependencies,omitempty"`
	CascadeDeleteCount    int                    `json:"cascade_delete_count"`
	CascadeDeleteEntities []CascadeDeletePreview `json:"cascade_delete_entities,omitempty"`
	RequiresConfirmation  bool                   `json:"requires_confirmation"`
}

// DependencyDetail represents a specific dependency that prevents deletion
type DependencyDetail struct {
	EntityType  string    `json:"entity_type"`
	EntityID    uuid.UUID `json:"entity_id"`
	ReferenceID string    `json:"reference_id"`
	Title       string    `json:"title"`
	Reason      string    `json:"reason"`
}

// CascadeDeletePreview represents entities that would be cascade deleted
type CascadeDeletePreview struct {
	EntityType  string    `json:"entity_type"`
	EntityID    uuid.UUID `json:"entity_id"`
	ReferenceID string    `json:"reference_id"`
	Title       string    `json:"title"`
}

// AuditLog represents an audit log entry for deletion operations
type AuditLog struct {
	ID            uuid.UUID              `json:"id"`
	EntityType    string                 `json:"entity_type"`
	EntityID      uuid.UUID              `json:"entity_id"`
	ReferenceID   string                 `json:"reference_id"`
	Operation     string                 `json:"operation"`
	PerformedBy   uuid.UUID              `json:"performed_by"`
	PerformedAt   time.Time              `json:"performed_at"`
	Details       map[string]interface{} `json:"details"`
	TransactionID string                 `json:"transaction_id"`
}

// deletionService implements DeletionService interface
type deletionService struct {
	// Repositories
	epicRepo                    repository.EpicRepository
	userStoryRepo               repository.UserStoryRepository
	acceptanceCriteriaRepo      repository.AcceptanceCriteriaRepository
	requirementRepo             repository.RequirementRepository
	requirementRelationshipRepo repository.RequirementRelationshipRepository
	commentRepo                 repository.CommentRepository
	userRepo                    repository.UserRepository

	// Logger
	logger *logrus.Logger
}

// NewDeletionService creates a new deletion service instance
func NewDeletionService(
	epicRepo repository.EpicRepository,
	userStoryRepo repository.UserStoryRepository,
	acceptanceCriteriaRepo repository.AcceptanceCriteriaRepository,
	requirementRepo repository.RequirementRepository,
	requirementRelationshipRepo repository.RequirementRelationshipRepository,
	commentRepo repository.CommentRepository,
	userRepo repository.UserRepository,
	logger *logrus.Logger,
) DeletionService {
	return &deletionService{
		epicRepo:                    epicRepo,
		userStoryRepo:               userStoryRepo,
		acceptanceCriteriaRepo:      acceptanceCriteriaRepo,
		requirementRepo:             requirementRepo,
		requirementRelationshipRepo: requirementRelationshipRepo,
		commentRepo:                 commentRepo,
		userRepo:                    userRepo,
		logger:                      logger,
	}
}

// generateTransactionID generates a unique transaction ID for audit logging
func (s *deletionService) generateTransactionID() string {
	return fmt.Sprintf("del_%d_%s", time.Now().Unix(), uuid.New().String()[:8])
}

// logAuditEntry creates an audit log entry for deletion operations
func (s *deletionService) logAuditEntry(entityType string, entityID uuid.UUID, referenceID string, operation string, performedBy uuid.UUID, transactionID string, details map[string]interface{}) uuid.UUID {
	auditID := uuid.New()

	s.logger.WithFields(logrus.Fields{
		"audit_id":       auditID,
		"entity_type":    entityType,
		"entity_id":      entityID,
		"reference_id":   referenceID,
		"operation":      operation,
		"performed_by":   performedBy,
		"transaction_id": transactionID,
		"details":        details,
	}).Info("Deletion audit log entry")

	return auditID
}

// DeleteEpicWithValidation deletes an epic with comprehensive validation and cascading
func (s *deletionService) DeleteEpicWithValidation(id uuid.UUID, userID uuid.UUID, force bool) (*DeletionResult, error) {
	transactionID := s.generateTransactionID()

	s.logger.WithFields(logrus.Fields{
		"epic_id":        id,
		"user_id":        userID,
		"force":          force,
		"transaction_id": transactionID,
	}).Info("Starting epic deletion with validation")

	// Get epic details for audit logging
	epic, err := s.epicRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrEpicNotFound
		}
		return nil, fmt.Errorf("failed to get epic: %w", err)
	}

	// Validate dependencies
	depInfo, err := s.ValidateEpicDeletion(id)
	if err != nil {
		return nil, fmt.Errorf("failed to validate epic deletion: %w", err)
	}

	// If not force delete and has dependencies, return validation error
	if !force && !depInfo.CanDelete {
		s.logger.WithFields(logrus.Fields{
			"epic_id":        id,
			"dependencies":   len(depInfo.Dependencies),
			"transaction_id": transactionID,
		}).Warn("Epic deletion blocked due to dependencies")
		return nil, ErrDeletionValidationFailed
	}

	// Perform deletion in transaction
	var result *DeletionResult
	err = s.epicRepo.WithTransaction(func(tx *gorm.DB) error {
		cascadeDeleted := []CascadeDeletedEntity{}

		if force {
			// Delete all user stories and their dependencies
			userStories, err := s.userStoryRepo.GetByEpic(id)
			if err != nil {
				return fmt.Errorf("failed to get user stories for epic: %w", err)
			}

			for _, userStory := range userStories {
				// Delete user story with cascade
				userStoryResult, err := s.deleteUserStoryInTransaction(tx, userStory.ID, userID, transactionID)
				if err != nil {
					return fmt.Errorf("failed to cascade delete user story %s: %w", userStory.ReferenceID, err)
				}

				// Add to cascade deleted list
				cascadeDeleted = append(cascadeDeleted, CascadeDeletedEntity{
					EntityType:  "user_story",
					EntityID:    userStory.ID,
					ReferenceID: userStory.ReferenceID,
				})

				// Add nested cascade deletions
				cascadeDeleted = append(cascadeDeleted, userStoryResult.CascadeDeleted...)
			}
		}

		// Delete comments associated with the epic
		if err := s.deleteCommentsInTransaction(tx, models.EntityTypeEpic, id, transactionID); err != nil {
			return fmt.Errorf("failed to delete epic comments: %w", err)
		}

		// Delete the epic itself
		if err := tx.Delete(&models.Epic{}, id).Error; err != nil {
			return fmt.Errorf("failed to delete epic: %w", err)
		}

		// Create audit log
		auditID := s.logAuditEntry("epic", id, epic.ReferenceID, "DELETE", userID, transactionID, map[string]interface{}{
			"force":         force,
			"cascade_count": len(cascadeDeleted),
			"title":         epic.Title,
		})

		result = &DeletionResult{
			EntityType:     "epic",
			EntityID:       id,
			ReferenceID:    epic.ReferenceID,
			DeletedAt:      time.Now(),
			DeletedBy:      userID,
			CascadeDeleted: cascadeDeleted,
			AuditLogID:     auditID,
			TransactionID:  transactionID,
		}

		return nil
	})

	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"epic_id":        id,
			"error":          err.Error(),
			"transaction_id": transactionID,
		}).Error("Epic deletion transaction failed")
		return nil, ErrDeletionTransactionFailed
	}

	s.logger.WithFields(logrus.Fields{
		"epic_id":        id,
		"cascade_count":  len(result.CascadeDeleted),
		"transaction_id": transactionID,
	}).Info("Epic deletion completed successfully")

	return result, nil
}

// DeleteUserStoryWithValidation deletes a user story with comprehensive validation and cascading
func (s *deletionService) DeleteUserStoryWithValidation(id uuid.UUID, userID uuid.UUID, force bool) (*DeletionResult, error) {
	transactionID := s.generateTransactionID()

	s.logger.WithFields(logrus.Fields{
		"user_story_id":  id,
		"user_id":        userID,
		"force":          force,
		"transaction_id": transactionID,
	}).Info("Starting user story deletion with validation")

	// Validate user story exists (done in ValidateUserStoryDeletion)

	// Validate dependencies
	depInfo, err := s.ValidateUserStoryDeletion(id)
	if err != nil {
		return nil, fmt.Errorf("failed to validate user story deletion: %w", err)
	}

	// If not force delete and has dependencies, return validation error
	if !force && !depInfo.CanDelete {
		s.logger.WithFields(logrus.Fields{
			"user_story_id":  id,
			"dependencies":   len(depInfo.Dependencies),
			"transaction_id": transactionID,
		}).Warn("User story deletion blocked due to dependencies")
		return nil, ErrDeletionValidationFailed
	}

	// Perform deletion in transaction
	var result *DeletionResult
	err = s.userStoryRepo.WithTransaction(func(tx *gorm.DB) error {
		userStoryResult, err := s.deleteUserStoryInTransaction(tx, id, userID, transactionID)
		if err != nil {
			return err
		}
		result = userStoryResult
		return nil
	})

	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"user_story_id":  id,
			"error":          err.Error(),
			"transaction_id": transactionID,
		}).Error("User story deletion transaction failed")
		return nil, ErrDeletionTransactionFailed
	}

	s.logger.WithFields(logrus.Fields{
		"user_story_id":  id,
		"cascade_count":  len(result.CascadeDeleted),
		"transaction_id": transactionID,
	}).Info("User story deletion completed successfully")

	return result, nil
}

// deleteUserStoryInTransaction deletes a user story within a transaction (helper method)
func (s *deletionService) deleteUserStoryInTransaction(tx *gorm.DB, id uuid.UUID, userID uuid.UUID, transactionID string) (*DeletionResult, error) {
	// Get user story details
	userStory, err := s.userStoryRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user story: %w", err)
	}

	cascadeDeleted := []CascadeDeletedEntity{}

	// Delete all acceptance criteria
	acceptanceCriteria, err := s.acceptanceCriteriaRepo.GetByUserStory(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get acceptance criteria for user story: %w", err)
	}

	for _, ac := range acceptanceCriteria {
		acResult, err := s.deleteAcceptanceCriteriaInTransaction(tx, ac.ID, userID, transactionID)
		if err != nil {
			return nil, fmt.Errorf("failed to cascade delete acceptance criteria %s: %w", ac.ReferenceID, err)
		}

		cascadeDeleted = append(cascadeDeleted, CascadeDeletedEntity{
			EntityType:  "acceptance_criteria",
			EntityID:    ac.ID,
			ReferenceID: ac.ReferenceID,
		})

		cascadeDeleted = append(cascadeDeleted, acResult.CascadeDeleted...)
	}

	// Delete all requirements
	requirements, err := s.requirementRepo.GetByUserStory(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get requirements for user story: %w", err)
	}

	for _, req := range requirements {
		reqResult, err := s.deleteRequirementInTransaction(tx, req.ID, userID, transactionID)
		if err != nil {
			return nil, fmt.Errorf("failed to cascade delete requirement %s: %w", req.ReferenceID, err)
		}

		cascadeDeleted = append(cascadeDeleted, CascadeDeletedEntity{
			EntityType:  "requirement",
			EntityID:    req.ID,
			ReferenceID: req.ReferenceID,
		})

		cascadeDeleted = append(cascadeDeleted, reqResult.CascadeDeleted...)
	}

	// Delete comments associated with the user story
	if err := s.deleteCommentsInTransaction(tx, models.EntityTypeUserStory, id, transactionID); err != nil {
		return nil, fmt.Errorf("failed to delete user story comments: %w", err)
	}

	// Delete the user story itself
	if err := tx.Delete(&models.UserStory{}, id).Error; err != nil {
		return nil, fmt.Errorf("failed to delete user story: %w", err)
	}

	// Create audit log
	auditID := s.logAuditEntry("user_story", id, userStory.ReferenceID, "DELETE", userID, transactionID, map[string]interface{}{
		"cascade_count": len(cascadeDeleted),
		"title":         userStory.Title,
		"epic_id":       userStory.EpicID,
	})

	return &DeletionResult{
		EntityType:     "user_story",
		EntityID:       id,
		ReferenceID:    userStory.ReferenceID,
		DeletedAt:      time.Now(),
		DeletedBy:      userID,
		CascadeDeleted: cascadeDeleted,
		AuditLogID:     auditID,
		TransactionID:  transactionID,
	}, nil
}

// DeleteAcceptanceCriteriaWithValidation deletes acceptance criteria with comprehensive validation and cascading
func (s *deletionService) DeleteAcceptanceCriteriaWithValidation(id uuid.UUID, userID uuid.UUID, force bool) (*DeletionResult, error) {
	transactionID := s.generateTransactionID()

	s.logger.WithFields(logrus.Fields{
		"acceptance_criteria_id": id,
		"user_id":                userID,
		"force":                  force,
		"transaction_id":         transactionID,
	}).Info("Starting acceptance criteria deletion with validation")

	// Validate acceptance criteria exists (done in ValidateAcceptanceCriteriaDeletion)

	// Validate dependencies
	depInfo, err := s.ValidateAcceptanceCriteriaDeletion(id)
	if err != nil {
		return nil, fmt.Errorf("failed to validate acceptance criteria deletion: %w", err)
	}

	// If not force delete and has dependencies, return validation error
	if !force && !depInfo.CanDelete {
		s.logger.WithFields(logrus.Fields{
			"acceptance_criteria_id": id,
			"dependencies":           len(depInfo.Dependencies),
			"transaction_id":         transactionID,
		}).Warn("Acceptance criteria deletion blocked due to dependencies")
		return nil, ErrDeletionValidationFailed
	}

	// Perform deletion in transaction
	var result *DeletionResult
	err = s.acceptanceCriteriaRepo.WithTransaction(func(tx *gorm.DB) error {
		acResult, err := s.deleteAcceptanceCriteriaInTransaction(tx, id, userID, transactionID)
		if err != nil {
			return err
		}
		result = acResult
		return nil
	})

	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"acceptance_criteria_id": id,
			"error":                  err.Error(),
			"transaction_id":         transactionID,
		}).Error("Acceptance criteria deletion transaction failed")
		return nil, ErrDeletionTransactionFailed
	}

	s.logger.WithFields(logrus.Fields{
		"acceptance_criteria_id": id,
		"cascade_count":          len(result.CascadeDeleted),
		"transaction_id":         transactionID,
	}).Info("Acceptance criteria deletion completed successfully")

	return result, nil
}

// deleteAcceptanceCriteriaInTransaction deletes acceptance criteria within a transaction (helper method)
func (s *deletionService) deleteAcceptanceCriteriaInTransaction(tx *gorm.DB, id uuid.UUID, userID uuid.UUID, transactionID string) (*DeletionResult, error) {
	// Get acceptance criteria details
	acceptanceCriteria, err := s.acceptanceCriteriaRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get acceptance criteria: %w", err)
	}

	cascadeDeleted := []CascadeDeletedEntity{}

	// Delete all requirements linked to this acceptance criteria
	requirements, err := s.requirementRepo.GetByAcceptanceCriteria(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get requirements for acceptance criteria: %w", err)
	}

	for _, req := range requirements {
		// Update requirement to remove acceptance criteria link instead of deleting
		if err := tx.Model(&models.Requirement{}).Where("id = ?", req.ID).Update("acceptance_criteria_id", nil).Error; err != nil {
			return nil, fmt.Errorf("failed to unlink requirement %s from acceptance criteria: %w", req.ReferenceID, err)
		}

		s.logger.WithFields(logrus.Fields{
			"requirement_id":         req.ID,
			"acceptance_criteria_id": id,
			"transaction_id":         transactionID,
		}).Info("Unlinked requirement from acceptance criteria")
	}

	// Delete comments associated with the acceptance criteria
	if err := s.deleteCommentsInTransaction(tx, models.EntityTypeAcceptanceCriteria, id, transactionID); err != nil {
		return nil, fmt.Errorf("failed to delete acceptance criteria comments: %w", err)
	}

	// Delete the acceptance criteria itself
	if err := tx.Delete(&models.AcceptanceCriteria{}, id).Error; err != nil {
		return nil, fmt.Errorf("failed to delete acceptance criteria: %w", err)
	}

	// Create audit log
	auditID := s.logAuditEntry("acceptance_criteria", id, acceptanceCriteria.ReferenceID, "DELETE", userID, transactionID, map[string]interface{}{
		"cascade_count": len(cascadeDeleted),
		"description":   acceptanceCriteria.Description,
		"user_story_id": acceptanceCriteria.UserStoryID,
		"unlinked_reqs": len(requirements),
	})

	return &DeletionResult{
		EntityType:     "acceptance_criteria",
		EntityID:       id,
		ReferenceID:    acceptanceCriteria.ReferenceID,
		DeletedAt:      time.Now(),
		DeletedBy:      userID,
		CascadeDeleted: cascadeDeleted,
		AuditLogID:     auditID,
		TransactionID:  transactionID,
	}, nil
}

// DeleteRequirementWithValidation deletes a requirement with comprehensive validation and cascading
func (s *deletionService) DeleteRequirementWithValidation(id uuid.UUID, userID uuid.UUID, force bool) (*DeletionResult, error) {
	transactionID := s.generateTransactionID()

	s.logger.WithFields(logrus.Fields{
		"requirement_id": id,
		"user_id":        userID,
		"force":          force,
		"transaction_id": transactionID,
	}).Info("Starting requirement deletion with validation")

	// Validate requirement exists (done in ValidateRequirementDeletion)

	// Validate dependencies
	depInfo, err := s.ValidateRequirementDeletion(id)
	if err != nil {
		return nil, fmt.Errorf("failed to validate requirement deletion: %w", err)
	}

	// If not force delete and has dependencies, return validation error
	if !force && !depInfo.CanDelete {
		s.logger.WithFields(logrus.Fields{
			"requirement_id": id,
			"dependencies":   len(depInfo.Dependencies),
			"transaction_id": transactionID,
		}).Warn("Requirement deletion blocked due to dependencies")
		return nil, ErrDeletionValidationFailed
	}

	// Perform deletion in transaction
	var result *DeletionResult
	err = s.requirementRepo.WithTransaction(func(tx *gorm.DB) error {
		reqResult, err := s.deleteRequirementInTransaction(tx, id, userID, transactionID)
		if err != nil {
			return err
		}
		result = reqResult
		return nil
	})

	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"requirement_id": id,
			"error":          err.Error(),
			"transaction_id": transactionID,
		}).Error("Requirement deletion transaction failed")
		return nil, ErrDeletionTransactionFailed
	}

	s.logger.WithFields(logrus.Fields{
		"requirement_id": id,
		"cascade_count":  len(result.CascadeDeleted),
		"transaction_id": transactionID,
	}).Info("Requirement deletion completed successfully")

	return result, nil
}

// deleteRequirementInTransaction deletes a requirement within a transaction (helper method)
func (s *deletionService) deleteRequirementInTransaction(tx *gorm.DB, id uuid.UUID, userID uuid.UUID, transactionID string) (*DeletionResult, error) {
	// Get requirement details
	requirement, err := s.requirementRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get requirement: %w", err)
	}

	cascadeDeleted := []CascadeDeletedEntity{}

	// Delete all requirement relationships (both source and target)
	relationships, err := s.requirementRelationshipRepo.GetByRequirement(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get relationships for requirement: %w", err)
	}

	for _, rel := range relationships {
		if err := tx.Delete(&models.RequirementRelationship{}, rel.ID).Error; err != nil {
			return nil, fmt.Errorf("failed to delete requirement relationship: %w", err)
		}

		cascadeDeleted = append(cascadeDeleted, CascadeDeletedEntity{
			EntityType:  "requirement_relationship",
			EntityID:    rel.ID,
			ReferenceID: fmt.Sprintf("REL-%s", rel.ID.String()[:8]),
		})

		s.logger.WithFields(logrus.Fields{
			"relationship_id": rel.ID,
			"requirement_id":  id,
			"transaction_id":  transactionID,
		}).Info("Deleted requirement relationship")
	}

	// Delete comments associated with the requirement
	if err := s.deleteCommentsInTransaction(tx, models.EntityTypeRequirement, id, transactionID); err != nil {
		return nil, fmt.Errorf("failed to delete requirement comments: %w", err)
	}

	// Delete the requirement itself
	if err := tx.Delete(&models.Requirement{}, id).Error; err != nil {
		return nil, fmt.Errorf("failed to delete requirement: %w", err)
	}

	// Create audit log
	auditID := s.logAuditEntry("requirement", id, requirement.ReferenceID, "DELETE", userID, transactionID, map[string]interface{}{
		"cascade_count":          len(cascadeDeleted),
		"title":                  requirement.Title,
		"user_story_id":          requirement.UserStoryID,
		"acceptance_criteria_id": requirement.AcceptanceCriteriaID,
		"deleted_relationships":  len(relationships),
	})

	return &DeletionResult{
		EntityType:     "requirement",
		EntityID:       id,
		ReferenceID:    requirement.ReferenceID,
		DeletedAt:      time.Now(),
		DeletedBy:      userID,
		CascadeDeleted: cascadeDeleted,
		AuditLogID:     auditID,
		TransactionID:  transactionID,
	}, nil
}

// deleteCommentsInTransaction deletes all comments for an entity within a transaction
func (s *deletionService) deleteCommentsInTransaction(tx *gorm.DB, entityType models.EntityType, entityID uuid.UUID, transactionID string) error {
	// Delete all comments for the entity
	if err := tx.Where("entity_type = ? AND entity_id = ?", entityType, entityID).Delete(&models.Comment{}).Error; err != nil {
		return fmt.Errorf("failed to delete comments for entity %s %s: %w", entityType, entityID, err)
	}

	s.logger.WithFields(logrus.Fields{
		"entity_type":    entityType,
		"entity_id":      entityID,
		"transaction_id": transactionID,
	}).Info("Deleted comments for entity")

	return nil
}

// ValidateEpicDeletion validates if an epic can be deleted and returns dependency information
func (s *deletionService) ValidateEpicDeletion(id uuid.UUID) (*DependencyInfo, error) {
	s.logger.WithFields(logrus.Fields{
		"epic_id": id,
	}).Debug("Validating epic deletion")

	// Check if epic exists
	epic, err := s.epicRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrEpicNotFound
		}
		return nil, fmt.Errorf("failed to get epic: %w", err)
	}
	_ = epic // Used for validation

	dependencies := []DependencyDetail{}
	cascadeEntities := []CascadeDeletePreview{}

	// Check for user stories
	userStories, err := s.userStoryRepo.GetByEpic(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stories for epic: %w", err)
	}

	if len(userStories) > 0 {
		for _, us := range userStories {
			dependencies = append(dependencies, DependencyDetail{
				EntityType:  "user_story",
				EntityID:    us.ID,
				ReferenceID: us.ReferenceID,
				Title:       us.Title,
				Reason:      "Epic contains user stories",
			})

			cascadeEntities = append(cascadeEntities, CascadeDeletePreview{
				EntityType:  "user_story",
				EntityID:    us.ID,
				ReferenceID: us.ReferenceID,
				Title:       us.Title,
			})

			// Add nested dependencies (acceptance criteria and requirements)
			acceptanceCriteria, err := s.acceptanceCriteriaRepo.GetByUserStory(us.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to get acceptance criteria for user story %s: %w", us.ReferenceID, err)
			}

			for _, ac := range acceptanceCriteria {
				cascadeEntities = append(cascadeEntities, CascadeDeletePreview{
					EntityType:  "acceptance_criteria",
					EntityID:    ac.ID,
					ReferenceID: ac.ReferenceID,
					Title:       fmt.Sprintf("AC: %s", ac.Description[:min(50, len(ac.Description))]),
				})
			}

			requirements, err := s.requirementRepo.GetByUserStory(us.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to get requirements for user story %s: %w", us.ReferenceID, err)
			}

			for _, req := range requirements {
				cascadeEntities = append(cascadeEntities, CascadeDeletePreview{
					EntityType:  "requirement",
					EntityID:    req.ID,
					ReferenceID: req.ReferenceID,
					Title:       req.Title,
				})
			}
		}
	}

	canDelete := len(dependencies) == 0
	requiresConfirmation := len(cascadeEntities) > 0

	return &DependencyInfo{
		CanDelete:             canDelete,
		Dependencies:          dependencies,
		CascadeDeleteCount:    len(cascadeEntities),
		CascadeDeleteEntities: cascadeEntities,
		RequiresConfirmation:  requiresConfirmation,
	}, nil
}

// ValidateUserStoryDeletion validates if a user story can be deleted and returns dependency information
func (s *deletionService) ValidateUserStoryDeletion(id uuid.UUID) (*DependencyInfo, error) {
	s.logger.WithFields(logrus.Fields{
		"user_story_id": id,
	}).Debug("Validating user story deletion")

	// Check if user story exists
	userStory, err := s.userStoryRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserStoryNotFound
		}
		return nil, fmt.Errorf("failed to get user story: %w", err)
	}
	_ = userStory // Used for validation

	dependencies := []DependencyDetail{}
	cascadeEntities := []CascadeDeletePreview{}

	// Check for acceptance criteria
	acceptanceCriteria, err := s.acceptanceCriteriaRepo.GetByUserStory(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get acceptance criteria for user story: %w", err)
	}

	if len(acceptanceCriteria) > 0 {
		for _, ac := range acceptanceCriteria {
			dependencies = append(dependencies, DependencyDetail{
				EntityType:  "acceptance_criteria",
				EntityID:    ac.ID,
				ReferenceID: ac.ReferenceID,
				Title:       fmt.Sprintf("AC: %s", ac.Description[:min(50, len(ac.Description))]),
				Reason:      "User story contains acceptance criteria",
			})

			cascadeEntities = append(cascadeEntities, CascadeDeletePreview{
				EntityType:  "acceptance_criteria",
				EntityID:    ac.ID,
				ReferenceID: ac.ReferenceID,
				Title:       fmt.Sprintf("AC: %s", ac.Description[:min(50, len(ac.Description))]),
			})
		}
	}

	// Check for requirements
	requirements, err := s.requirementRepo.GetByUserStory(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get requirements for user story: %w", err)
	}

	if len(requirements) > 0 {
		for _, req := range requirements {
			dependencies = append(dependencies, DependencyDetail{
				EntityType:  "requirement",
				EntityID:    req.ID,
				ReferenceID: req.ReferenceID,
				Title:       req.Title,
				Reason:      "User story contains requirements",
			})

			cascadeEntities = append(cascadeEntities, CascadeDeletePreview{
				EntityType:  "requirement",
				EntityID:    req.ID,
				ReferenceID: req.ReferenceID,
				Title:       req.Title,
			})
		}
	}

	canDelete := len(dependencies) == 0
	requiresConfirmation := len(cascadeEntities) > 0

	return &DependencyInfo{
		CanDelete:             canDelete,
		Dependencies:          dependencies,
		CascadeDeleteCount:    len(cascadeEntities),
		CascadeDeleteEntities: cascadeEntities,
		RequiresConfirmation:  requiresConfirmation,
	}, nil
}

// ValidateAcceptanceCriteriaDeletion validates if acceptance criteria can be deleted and returns dependency information
func (s *deletionService) ValidateAcceptanceCriteriaDeletion(id uuid.UUID) (*DependencyInfo, error) {
	s.logger.WithFields(logrus.Fields{
		"acceptance_criteria_id": id,
	}).Debug("Validating acceptance criteria deletion")

	// Check if acceptance criteria exists
	acceptanceCriteria, err := s.acceptanceCriteriaRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrAcceptanceCriteriaNotFound
		}
		return nil, fmt.Errorf("failed to get acceptance criteria: %w", err)
	}

	dependencies := []DependencyDetail{}
	cascadeEntities := []CascadeDeletePreview{}

	// Check if this is the last acceptance criteria for the user story
	count, err := s.acceptanceCriteriaRepo.CountByUserStory(acceptanceCriteria.UserStoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to count acceptance criteria for user story: %w", err)
	}

	if count <= 1 {
		dependencies = append(dependencies, DependencyDetail{
			EntityType:  "user_story",
			EntityID:    acceptanceCriteria.UserStoryID,
			ReferenceID: "", // Will be filled by getting user story
			Title:       "",
			Reason:      "User story must have at least one acceptance criteria",
		})
	}

	// Check for linked requirements (these will be unlinked, not deleted)
	requirements, err := s.requirementRepo.GetByAcceptanceCriteria(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get requirements for acceptance criteria: %w", err)
	}

	// Requirements are not dependencies since they will be unlinked, not deleted
	// But we should note them in the cascade entities for information
	for _, req := range requirements {
		cascadeEntities = append(cascadeEntities, CascadeDeletePreview{
			EntityType:  "requirement_unlink",
			EntityID:    req.ID,
			ReferenceID: req.ReferenceID,
			Title:       fmt.Sprintf("Unlink: %s", req.Title),
		})
	}

	canDelete := len(dependencies) == 0
	requiresConfirmation := len(cascadeEntities) > 0 || count <= 1

	return &DependencyInfo{
		CanDelete:             canDelete,
		Dependencies:          dependencies,
		CascadeDeleteCount:    len(cascadeEntities),
		CascadeDeleteEntities: cascadeEntities,
		RequiresConfirmation:  requiresConfirmation,
	}, nil
}

// ValidateRequirementDeletion validates if a requirement can be deleted and returns dependency information
func (s *deletionService) ValidateRequirementDeletion(id uuid.UUID) (*DependencyInfo, error) {
	s.logger.WithFields(logrus.Fields{
		"requirement_id": id,
	}).Debug("Validating requirement deletion")

	// Check if requirement exists
	requirement, err := s.requirementRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrRequirementNotFound
		}
		return nil, fmt.Errorf("failed to get requirement: %w", err)
	}
	_ = requirement // Used for validation

	dependencies := []DependencyDetail{}
	cascadeEntities := []CascadeDeletePreview{}

	// Check for requirement relationships
	relationships, err := s.requirementRelationshipRepo.GetByRequirement(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get relationships for requirement: %w", err)
	}

	if len(relationships) > 0 {
		for _, rel := range relationships {
			// Get the other requirement in the relationship
			var otherReqID uuid.UUID
			var relationshipDirection string

			if rel.SourceRequirementID == id {
				otherReqID = rel.TargetRequirementID
				relationshipDirection = "outgoing"
			} else {
				otherReqID = rel.SourceRequirementID
				relationshipDirection = "incoming"
			}

			otherReq, err := s.requirementRepo.GetByID(otherReqID)
			if err != nil {
				continue // Skip if other requirement doesn't exist
			}

			dependencies = append(dependencies, DependencyDetail{
				EntityType:  "requirement_relationship",
				EntityID:    rel.ID,
				ReferenceID: fmt.Sprintf("REL-%s", rel.ID.String()[:8]),
				Title:       fmt.Sprintf("%s relationship with %s", relationshipDirection, otherReq.ReferenceID),
				Reason:      "Requirement has active relationships",
			})

			cascadeEntities = append(cascadeEntities, CascadeDeletePreview{
				EntityType:  "requirement_relationship",
				EntityID:    rel.ID,
				ReferenceID: fmt.Sprintf("REL-%s", rel.ID.String()[:8]),
				Title:       fmt.Sprintf("Relationship with %s", otherReq.ReferenceID),
			})
		}
	}

	canDelete := len(dependencies) == 0
	requiresConfirmation := len(cascadeEntities) > 0

	return &DependencyInfo{
		CanDelete:             canDelete,
		Dependencies:          dependencies,
		CascadeDeleteCount:    len(cascadeEntities),
		CascadeDeleteEntities: cascadeEntities,
		RequiresConfirmation:  requiresConfirmation,
	}, nil
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
