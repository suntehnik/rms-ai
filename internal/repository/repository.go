package repository

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Repositories holds all repository instances
type Repositories struct {
	User                    UserRepository
	Epic                    EpicRepository
	UserStory               UserStoryRepository
	AcceptanceCriteria      AcceptanceCriteriaRepository
	Requirement             RequirementRepository
	RequirementType         RequirementTypeRepository
	RelationshipType        RelationshipTypeRepository
	RequirementRelationship RequirementRelationshipRepository
	Comment                 CommentRepository
	StatusModel             StatusModelRepository
	Status                  StatusRepository
	StatusTransition        StatusTransitionRepository
	PersonalAccessToken     PersonalAccessTokenRepository
	SteeringDocument        SteeringDocumentRepository
}

// NewRepositories creates a new instance of all repositories
func NewRepositories(db *gorm.DB, redis *redis.Client) *Repositories {
	return &Repositories{
		User:                    NewUserRepository(db),
		Epic:                    NewEpicRepository(db),
		UserStory:               NewUserStoryRepository(db, redis),
		AcceptanceCriteria:      NewAcceptanceCriteriaRepository(db),
		Requirement:             NewRequirementRepository(db),
		RequirementType:         NewRequirementTypeRepository(db),
		RelationshipType:        NewRelationshipTypeRepository(db),
		RequirementRelationship: NewRequirementRelationshipRepository(db),
		Comment:                 NewCommentRepository(db),
		StatusModel:             NewStatusModelRepository(db),
		Status:                  NewStatusRepository(db),
		StatusTransition:        NewStatusTransitionRepository(db),
		PersonalAccessToken:     NewPersonalAccessTokenRepository(db),
		SteeringDocument:        NewSteeringDocumentRepository(db),
	}
}

// WithTransaction executes a function within a database transaction
// This is a convenience method that can be used when multiple repositories need to work together
func (r *Repositories) WithTransaction(fn func(*Repositories) error) error {
	return r.User.WithTransaction(func(tx *gorm.DB) error {
		// Create new repository instances with the transaction
		txRepos := &Repositories{
			User:                    NewUserRepository(tx),
			Epic:                    NewEpicRepository(tx),
			UserStory:               NewUserStoryRepository(tx, nil),
			AcceptanceCriteria:      NewAcceptanceCriteriaRepository(tx),
			Requirement:             NewRequirementRepository(tx),
			RequirementType:         NewRequirementTypeRepository(tx),
			RelationshipType:        NewRelationshipTypeRepository(tx),
			RequirementRelationship: NewRequirementRelationshipRepository(tx),
			Comment:                 NewCommentRepository(tx),
			StatusModel:             NewStatusModelRepository(tx),
			Status:                  NewStatusRepository(tx),
			StatusTransition:        NewStatusTransitionRepository(tx),
			PersonalAccessToken:     NewPersonalAccessTokenRepository(tx),
			SteeringDocument:        NewSteeringDocumentRepository(tx),
		}
		return fn(txRepos)
	})
}
