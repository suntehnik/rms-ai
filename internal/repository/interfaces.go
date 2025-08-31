package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	
	"product-requirements-management/internal/models"
)

// Type aliases for models to make interfaces cleaner
type (
	User                      = models.User
	Epic                      = models.Epic
	UserStory                 = models.UserStory
	AcceptanceCriteria        = models.AcceptanceCriteria
	Requirement               = models.Requirement
	RequirementType           = models.RequirementType
	RelationshipType          = models.RelationshipType
	RequirementRelationship   = models.RequirementRelationship
	Comment                   = models.Comment
	StatusModel               = models.StatusModel
	Status                    = models.Status
	StatusTransition          = models.StatusTransition
	EpicStatus                = models.EpicStatus
	UserStoryStatus           = models.UserStoryStatus
	RequirementStatus         = models.RequirementStatus
	Priority                  = models.Priority
	EntityType                = models.EntityType
)

// Repository defines the common interface for all repositories
type Repository[T any] interface {
	Create(entity *T) error
	GetByID(id uuid.UUID) (*T, error)
	GetByReferenceID(referenceID string) (*T, error)
	Update(entity *T) error
	Delete(id uuid.UUID) error
	List(filters map[string]interface{}, orderBy string, limit, offset int) ([]T, error)
	Count(filters map[string]interface{}) (int64, error)
	Exists(id uuid.UUID) (bool, error)
	ExistsByReferenceID(referenceID string) (bool, error)
	WithTransaction(fn func(*gorm.DB) error) error
	GetDB() *gorm.DB
}

// UserRepository defines user-specific repository operations
type UserRepository interface {
	Repository[User]
	GetByUsername(username string) (*User, error)
	GetByEmail(email string) (*User, error)
	ExistsByUsername(username string) (bool, error)
	ExistsByEmail(email string) (bool, error)
}

// EpicRepository defines epic-specific repository operations
type EpicRepository interface {
	Repository[Epic]
	GetWithUserStories(id uuid.UUID) (*Epic, error)
	GetByCreator(creatorID uuid.UUID) ([]Epic, error)
	GetByAssignee(assigneeID uuid.UUID) ([]Epic, error)
	GetByStatus(status EpicStatus) ([]Epic, error)
	GetByPriority(priority Priority) ([]Epic, error)
	HasUserStories(id uuid.UUID) (bool, error)
}

// UserStoryRepository defines user story-specific repository operations
type UserStoryRepository interface {
	Repository[UserStory]
	GetWithAcceptanceCriteria(id uuid.UUID) (*UserStory, error)
	GetWithRequirements(id uuid.UUID) (*UserStory, error)
	GetByEpic(epicID uuid.UUID) ([]UserStory, error)
	GetByCreator(creatorID uuid.UUID) ([]UserStory, error)
	GetByAssignee(assigneeID uuid.UUID) ([]UserStory, error)
	GetByStatus(status UserStoryStatus) ([]UserStory, error)
	GetByPriority(priority Priority) ([]UserStory, error)
	HasAcceptanceCriteria(id uuid.UUID) (bool, error)
	HasRequirements(id uuid.UUID) (bool, error)
}

// AcceptanceCriteriaRepository defines acceptance criteria-specific repository operations
type AcceptanceCriteriaRepository interface {
	Repository[AcceptanceCriteria]
	GetByUserStory(userStoryID uuid.UUID) ([]AcceptanceCriteria, error)
	GetByAuthor(authorID uuid.UUID) ([]AcceptanceCriteria, error)
	HasRequirements(id uuid.UUID) (bool, error)
	CountByUserStory(userStoryID uuid.UUID) (int64, error)
}

// RequirementRepository defines requirement-specific repository operations
type RequirementRepository interface {
	Repository[Requirement]
	GetWithRelationships(id uuid.UUID) (*Requirement, error)
	GetByUserStory(userStoryID uuid.UUID) ([]Requirement, error)
	GetByAcceptanceCriteria(acceptanceCriteriaID uuid.UUID) ([]Requirement, error)
	GetByCreator(creatorID uuid.UUID) ([]Requirement, error)
	GetByAssignee(assigneeID uuid.UUID) ([]Requirement, error)
	GetByStatus(status RequirementStatus) ([]Requirement, error)
	GetByPriority(priority Priority) ([]Requirement, error)
	GetByType(typeID uuid.UUID) ([]Requirement, error)
	HasRelationships(id uuid.UUID) (bool, error)
	SearchByText(searchText string) ([]Requirement, error)
}

// RequirementTypeRepository defines requirement type-specific repository operations
type RequirementTypeRepository interface {
	Repository[RequirementType]
	GetByName(name string) (*RequirementType, error)
	ExistsByName(name string) (bool, error)
}

// RelationshipTypeRepository defines relationship type-specific repository operations
type RelationshipTypeRepository interface {
	Repository[RelationshipType]
	GetByName(name string) (*RelationshipType, error)
	ExistsByName(name string) (bool, error)
}

// RequirementRelationshipRepository defines requirement relationship-specific repository operations
type RequirementRelationshipRepository interface {
	Repository[RequirementRelationship]
	GetBySourceRequirement(sourceID uuid.UUID) ([]RequirementRelationship, error)
	GetByTargetRequirement(targetID uuid.UUID) ([]RequirementRelationship, error)
	GetByRequirement(requirementID uuid.UUID) ([]RequirementRelationship, error)
	GetByType(typeID uuid.UUID) ([]RequirementRelationship, error)
	ExistsRelationship(sourceID, targetID, typeID uuid.UUID) (bool, error)
}

// CommentRepository defines comment-specific repository operations
type CommentRepository interface {
	Repository[Comment]
	GetByEntity(entityType EntityType, entityID uuid.UUID) ([]Comment, error)
	GetByAuthor(authorID uuid.UUID) ([]Comment, error)
	GetByParent(parentID uuid.UUID) ([]Comment, error)
	GetThreaded(entityType EntityType, entityID uuid.UUID) ([]Comment, error)
	GetByStatus(isResolved bool) ([]Comment, error)
	GetInlineComments(entityType EntityType, entityID uuid.UUID) ([]Comment, error)
}

// StatusModelRepository defines status model-specific repository operations
type StatusModelRepository interface {
	Create(statusModel *StatusModel) error
	GetByID(id uuid.UUID) (*StatusModel, error)
	GetByEntityTypeAndName(entityType EntityType, name string) (*StatusModel, error)
	GetDefaultByEntityType(entityType EntityType) (*StatusModel, error)
	Update(statusModel *StatusModel) error
	Delete(id uuid.UUID) error
	List(filters map[string]interface{}, orderBy string, limit, offset int) ([]StatusModel, error)
	ListByEntityType(entityType EntityType) ([]StatusModel, error)
	Exists(id uuid.UUID) (bool, error)
	ExistsByEntityTypeAndName(entityType EntityType, name string) (bool, error)
}

// StatusRepository defines status-specific repository operations
type StatusRepository interface {
	Create(status *Status) error
	GetByID(id uuid.UUID) (*Status, error)
	GetByStatusModelID(statusModelID uuid.UUID) ([]Status, error)
	GetByName(statusModelID uuid.UUID, name string) (*Status, error)
	Update(status *Status) error
	Delete(id uuid.UUID) error
	List(filters map[string]interface{}, orderBy string, limit, offset int) ([]Status, error)
	Exists(id uuid.UUID) (bool, error)
	ExistsByName(statusModelID uuid.UUID, name string) (bool, error)
}

// StatusTransitionRepository defines status transition-specific repository operations
type StatusTransitionRepository interface {
	Create(transition *StatusTransition) error
	GetByID(id uuid.UUID) (*StatusTransition, error)
	GetByStatusModelID(statusModelID uuid.UUID) ([]StatusTransition, error)
	GetByFromStatus(fromStatusID uuid.UUID) ([]StatusTransition, error)
	Update(transition *StatusTransition) error
	Delete(id uuid.UUID) error
	List(filters map[string]interface{}, orderBy string, limit, offset int) ([]StatusTransition, error)
	Exists(id uuid.UUID) (bool, error)
	ExistsByTransition(statusModelID, fromStatusID, toStatusID uuid.UUID) (bool, error)
}