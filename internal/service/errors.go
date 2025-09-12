package service

import "errors"

// Common service errors used across multiple services
var (
	// Requirement Type errors
	ErrRequirementTypeNotFound = errors.New("requirement type not found")

	// Relationship Type errors
	ErrRelationshipTypeNotFound = errors.New("relationship type not found")

	// Status transition errors
	ErrInvalidStatusTransition = errors.New("invalid status transition")
)
