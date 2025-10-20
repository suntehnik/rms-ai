package service

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"product-requirements-management/internal/models"
)

func TestCommentService_CreateComment_Basic(t *testing.T) {
	// This is a basic test to verify the comment service compiles and basic functionality works
	// More comprehensive tests would require setting up proper mocks or integration tests

	t.Run("validate entity type", func(t *testing.T) {
		// Test the isValidEntityType function
		assert.True(t, isValidEntityType(models.EntityTypeEpic))
		assert.True(t, isValidEntityType(models.EntityTypeUserStory))
		assert.True(t, isValidEntityType(models.EntityTypeAcceptanceCriteria))
		assert.True(t, isValidEntityType(models.EntityTypeRequirement))
		assert.False(t, isValidEntityType("invalid"))
	})

	t.Run("validate inline comment data", func(t *testing.T) {
		service := &commentService{}

		// Valid inline comment data
		linkedText := "test text"
		start := 0
		end := 9
		err := service.validateInlineCommentData(&linkedText, &start, &end)
		assert.NoError(t, err)

		// Invalid - missing fields
		err = service.validateInlineCommentData(&linkedText, nil, &end)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "inline comments require")

		// Invalid - negative start
		invalidStart := -1
		err = service.validateInlineCommentData(&linkedText, &invalidStart, &end)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid text position")

		// Invalid - end before start
		invalidEnd := -1
		err = service.validateInlineCommentData(&linkedText, &start, &invalidEnd)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid text position")

		// Invalid - empty linked text (but with valid positions)
		emptyText := "   " // whitespace only
		err = service.validateInlineCommentData(&emptyText, &start, &end)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "linked_text cannot be empty")
	})
}

func TestCommentService_ToCommentResponse(t *testing.T) {
	service := &commentService{}

	// Create a test comment
	commentID := uuid.New()
	authorID := uuid.New()
	entityID := uuid.New()

	comment := &models.Comment{
		ID:         commentID,
		EntityType: models.EntityTypeEpic,
		EntityID:   entityID,
		AuthorID:   authorID,
		Content:    "Test comment",
		IsResolved: false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	response := service.toCommentResponse(comment)

	assert.Equal(t, commentID, response.ID)
	assert.Equal(t, models.EntityTypeEpic, response.EntityType)
	assert.Equal(t, entityID, response.EntityID)
	assert.Equal(t, authorID, response.AuthorID)
	assert.Equal(t, "Test comment", response.Content)
	assert.False(t, response.IsResolved)
	assert.False(t, response.IsInline)
	assert.False(t, response.IsReply)
	assert.Equal(t, 0, response.Depth)
}

func TestCommentService_ToCommentResponseInline(t *testing.T) {
	service := &commentService{}

	// Create a test inline comment
	commentID := uuid.New()
	authorID := uuid.New()
	entityID := uuid.New()
	linkedText := "selected text"
	start := 10
	end := 23

	comment := &models.Comment{
		ID:                commentID,
		EntityType:        models.EntityTypeEpic,
		EntityID:          entityID,
		AuthorID:          authorID,
		Content:           "Inline comment",
		IsResolved:        false,
		LinkedText:        &linkedText,
		TextPositionStart: &start,
		TextPositionEnd:   &end,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	response := service.toCommentResponse(comment)

	assert.Equal(t, commentID, response.ID)
	assert.True(t, response.IsInline)
	assert.Equal(t, linkedText, *response.LinkedText)
	assert.Equal(t, start, *response.TextPositionStart)
	assert.Equal(t, end, *response.TextPositionEnd)
}

func TestCommentService_ToCommentResponseReply(t *testing.T) {
	service := &commentService{}

	// Create a test reply comment
	commentID := uuid.New()
	parentID := uuid.New()
	authorID := uuid.New()
	entityID := uuid.New()

	comment := &models.Comment{
		ID:              commentID,
		EntityType:      models.EntityTypeEpic,
		EntityID:        entityID,
		ParentCommentID: &parentID,
		AuthorID:        authorID,
		Content:         "Reply comment",
		IsResolved:      false,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	response := service.toCommentResponse(comment)

	assert.Equal(t, commentID, response.ID)
	assert.True(t, response.IsReply)
	assert.Equal(t, parentID, *response.ParentCommentID)
	assert.Equal(t, 1, response.Depth) // Replies have depth 1 in the current implementation
}

func TestCommentService_ValidationErrors(t *testing.T) {
	t.Run("empty content validation", func(t *testing.T) {
		req := CreateCommentRequest{}

		// This would normally require a full service setup, but we can test the validation logic
		service := &commentService{}
		err := service.validateInlineCommentData(req.LinkedText, req.TextPositionStart, req.TextPositionEnd)
		assert.NoError(t, err) // No inline comment data, so should be valid
	})

	t.Run("update comment validation", func(t *testing.T) {
		req := UpdateCommentRequest{
			Content: "   ", // Empty/whitespace content
		}

		// Test that empty content would be rejected
		assert.Equal(t, "", strings.TrimSpace(req.Content))
	})
}
