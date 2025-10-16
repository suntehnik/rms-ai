package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Integration test for deletion service using real service interfaces
func TestDeletionService_Integration(t *testing.T) {
	// Create a logger for testing
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	// Test that we can create a deletion service with nil repositories
	// This tests the constructor and basic structure
	deletionService := NewDeletionService(
		nil, // epicRepo
		nil, // userStoryRepo
		nil, // acceptanceCriteriaRepo
		nil, // requirementRepo
		nil, // requirementRelationshipRepo
		nil, // commentRepo
		nil, // userRepo
		logger,
	)

	assert.NotNil(t, deletionService)
}

// Test the transaction ID generation
func TestDeletionService_GenerateTransactionID(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewDeletionService(
		nil, nil, nil, nil, nil, nil, nil, logger,
	)

	// Use type assertion to access private method for testing
	deletionSvc, ok := service.(*deletionService)
	assert.True(t, ok)

	transactionID1 := deletionSvc.generateTransactionID()
	transactionID2 := deletionSvc.generateTransactionID()

	assert.NotEmpty(t, transactionID1)
	assert.NotEmpty(t, transactionID2)
	assert.NotEqual(t, transactionID1, transactionID2)
	assert.Contains(t, transactionID1, "del_")
	assert.Contains(t, transactionID2, "del_")
}

// Test the audit logging
func TestDeletionService_LogAuditEntry(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewDeletionService(
		nil, nil, nil, nil, nil, nil, nil, logger,
	)

	// Use type assertion to access private method for testing
	deletionSvc, ok := service.(*deletionService)
	assert.True(t, ok)

	auditID := deletionSvc.logAuditEntry(
		"epic",
		uuid.New(),
		"EP-001",
		"DELETE",
		uuid.New(),
		"test_transaction_id",
		map[string]interface{}{"test": "data"},
	)

	assert.NotEqual(t, uuid.Nil, auditID)
}

// Test error constants
func TestDeletionService_ErrorConstants(t *testing.T) {
	assert.Equal(t, "deletion cancelled by user", ErrDeletionCancelled.Error())
	assert.Equal(t, "deletion validation failed", ErrDeletionValidationFailed.Error())
	assert.Equal(t, "deletion transaction failed", ErrDeletionTransactionFailed.Error())
}

// Test DeletionResult structure
func TestDeletionResult_Structure(t *testing.T) {
	entityID := uuid.New()
	userID := uuid.New()
	auditID := uuid.New()

	result := &DeletionResult{
		EntityType:     "epic",
		EntityID:       entityID,
		ReferenceID:    "EP-001",
		DeletedBy:      userID,
		CascadeDeleted: []CascadeDeletedEntity{},
		AuditLogID:     auditID,
		TransactionID:  "test_transaction",
	}

	assert.Equal(t, "epic", result.EntityType)
	assert.Equal(t, entityID, result.EntityID)
	assert.Equal(t, "EP-001", result.ReferenceID)
	assert.Equal(t, userID, result.DeletedBy)
	assert.Empty(t, result.CascadeDeleted)
	assert.Equal(t, auditID, result.AuditLogID)
	assert.Equal(t, "test_transaction", result.TransactionID)
}

// Test DependencyInfo structure
func TestDependencyInfo_Structure(t *testing.T) {
	depInfo := &DependencyInfo{
		CanDelete:             false,
		Dependencies:          []DependencyDetail{},
		CascadeDeleteCount:    5,
		CascadeDeleteEntities: []CascadeDeletePreview{},
		RequiresConfirmation:  true,
	}

	assert.False(t, depInfo.CanDelete)
	assert.Empty(t, depInfo.Dependencies)
	assert.Equal(t, 5, depInfo.CascadeDeleteCount)
	assert.Empty(t, depInfo.CascadeDeleteEntities)
	assert.True(t, depInfo.RequiresConfirmation)
}

// Test CascadeDeletedEntity structure
func TestCascadeDeletedEntity_Structure(t *testing.T) {
	entityID := uuid.New()

	entity := CascadeDeletedEntity{
		EntityType:  "user_story",
		EntityID:    entityID,
		ReferenceID: "US-001",
	}

	assert.Equal(t, "user_story", entity.EntityType)
	assert.Equal(t, entityID, entity.EntityID)
	assert.Equal(t, "US-001", entity.ReferenceID)
}

// Test DependencyDetail structure
func TestDependencyDetail_Structure(t *testing.T) {
	entityID := uuid.New()

	detail := DependencyDetail{
		EntityType:  "requirement",
		EntityID:    entityID,
		ReferenceID: "REQ-001",
		Title:       "Test Requirement",
		Reason:      "Has active relationships",
	}

	assert.Equal(t, "requirement", detail.EntityType)
	assert.Equal(t, entityID, detail.EntityID)
	assert.Equal(t, "REQ-001", detail.ReferenceID)
	assert.Equal(t, "Test Requirement", detail.Title)
	assert.Equal(t, "Has active relationships", detail.Reason)
}

// Test CascadeDeletePreview structure
func TestCascadeDeletePreview_Structure(t *testing.T) {
	entityID := uuid.New()

	preview := CascadeDeletePreview{
		EntityType:  "acceptance_criteria",
		EntityID:    entityID,
		ReferenceID: "AC-001",
		Title:       "Test Acceptance Criteria",
	}

	assert.Equal(t, "acceptance_criteria", preview.EntityType)
	assert.Equal(t, entityID, preview.EntityID)
	assert.Equal(t, "AC-001", preview.ReferenceID)
	assert.Equal(t, "Test Acceptance Criteria", preview.Title)
}

// Test AuditLog structure
func TestAuditLog_Structure(t *testing.T) {
	auditID := uuid.New()
	entityID := uuid.New()
	userID := uuid.New()

	auditLog := AuditLog{
		ID:            auditID,
		EntityType:    "epic",
		EntityID:      entityID,
		ReferenceID:   "EP-001",
		Operation:     "DELETE",
		PerformedBy:   userID,
		Details:       map[string]interface{}{"force": true},
		TransactionID: "test_transaction",
	}

	assert.Equal(t, auditID, auditLog.ID)
	assert.Equal(t, "epic", auditLog.EntityType)
	assert.Equal(t, entityID, auditLog.EntityID)
	assert.Equal(t, "EP-001", auditLog.ReferenceID)
	assert.Equal(t, "DELETE", auditLog.Operation)
	assert.Equal(t, userID, auditLog.PerformedBy)
	assert.Equal(t, map[string]interface{}{"force": true}, auditLog.Details)
	assert.Equal(t, "test_transaction", auditLog.TransactionID)
}

// Test min helper function
func TestMinFunction(t *testing.T) {
	assert.Equal(t, 5, min(5, 10))
	assert.Equal(t, 3, min(10, 3))
	assert.Equal(t, 7, min(7, 7))
	assert.Equal(t, 0, min(0, 5))
	assert.Equal(t, -1, min(-1, 5))
}
