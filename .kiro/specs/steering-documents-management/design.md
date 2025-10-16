# Design Document

## Overview

Данный дизайн описывает реализацию системы управления steering документами как полноценной сущности в системе управления требованиями. Steering документы представляют собой инструкции, стандарты и нормы команды, которые могут быть связаны с эпиками для предоставления дополнительного контекста AI моделям при реализации.

## Architecture

### Database Schema

#### Основная таблица: steering_documents

```sql
CREATE TABLE steering_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reference_id VARCHAR(50) UNIQUE NOT NULL, -- STD-001, STD-002, etc.
    title VARCHAR(500) NOT NULL,
    description TEXT,
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_steering_documents_creator_id ON steering_documents(creator_id);
CREATE INDEX idx_steering_documents_reference_id ON steering_documents(reference_id);
CREATE INDEX idx_steering_documents_title ON steering_documents USING gin(to_tsvector('english', title));
CREATE INDEX idx_steering_documents_description ON steering_documents USING gin(to_tsvector('english', description));
```

#### Связующая таблица: epic_steering_documents

```sql
CREATE TABLE epic_steering_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    epic_id UUID NOT NULL REFERENCES epics(id) ON DELETE CASCADE,
    steering_document_id UUID NOT NULL REFERENCES steering_documents(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(epic_id, steering_document_id)
);

CREATE INDEX idx_epic_steering_documents_epic_id ON epic_steering_documents(epic_id);
CREATE INDEX idx_epic_steering_documents_steering_document_id ON epic_steering_documents(steering_document_id);
```

### GORM Models

#### SteeringDocument Model

```go
package models

import (
    "encoding/json"
    "time"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

// SteeringDocument represents a steering document in the system
// @Description Steering document contains instructions, standards and team norms that can be linked to epics for additional context
type SteeringDocument struct {
    // ID is the unique identifier for the steering document
    // @Description Unique UUID identifier for the steering document
    // @Example "123e4567-e89b-12d3-a456-426614174000"
    ID uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`

    // ReferenceID is the human-readable identifier for the steering document
    // @Description Human-readable reference identifier (auto-generated, format: STD-XXX)
    // @Example "STD-001"
    ReferenceID string `gorm:"uniqueIndex;not null" json:"reference_id"`

    // Title is the name/summary of the steering document
    // @Description Title or name of the steering document (required, max 500 characters)
    // @MaxLength 500
    // @Example "Code Review Standards"
    Title string `gorm:"not null" json:"title" validate:"required,max=500"`

    // Description provides detailed information about the steering document
    // @Description Detailed description of the steering document content (optional, max 50000 characters)
    // @MaxLength 50000
    // @Example "This document outlines the code review standards and practices for the development team..."
    Description *string `json:"description,omitempty" validate:"omitempty,max=50000"`

    // CreatorID is the UUID of the user who created the steering document
    // @Description UUID of the user who created this steering document
    // @Example "123e4567-e89b-12d3-a456-426614174001"
    CreatorID uuid.UUID `gorm:"not null" json:"creator_id"`

    // CreatedAt is the timestamp when the steering document was created
    // @Description Timestamp when the steering document was created (RFC3339 format)
    // @Example "2023-01-15T10:30:00Z"
    CreatedAt time.Time `json:"created_at"`

    // UpdatedAt is the timestamp when the steering document was last updated
    // @Description Timestamp when the steering document was last modified (RFC3339 format)
    // @Example "2023-01-16T14:45:30Z"
    UpdatedAt time.Time `json:"updated_at"`

    // Relationships - These fields are populated when explicitly requested and contain related entities

    // Creator contains the user information of who created the steering document
    // @Description User who created this steering document (included when explicitly preloaded)
    Creator User `gorm:"foreignKey:CreatorID;constraint:OnDelete:RESTRICT" json:"-"`

    // Epics contains all epics that are linked to this steering document
    // @Description List of epics linked to this steering document (populated when requested with ?include=epics)
    Epics []Epic `gorm:"many2many:epic_steering_documents;" json:"epics,omitempty"`
}

// BeforeCreate sets the ID if not already set
func (sd *SteeringDocument) BeforeCreate(tx *gorm.DB) error {
    if sd.ID == uuid.Nil {
        sd.ID = uuid.New()
    }
    // ReferenceID will be set by database default using the sequence
    return nil
}

// BeforeUpdate updates the UpdatedAt timestamp
func (sd *SteeringDocument) BeforeUpdate(tx *gorm.DB) error {
    sd.UpdatedAt = time.Now().UTC()
    return nil
}

// TableName returns the table name for the SteeringDocument model
func (SteeringDocument) TableName() string {
    return "steering_documents"
}

// MarshalJSON implements custom JSON marshaling for SteeringDocument
// This ensures that creator and epics objects are only included when they are actually populated
func (sd *SteeringDocument) MarshalJSON() ([]byte, error) {
    // Create a map to build the JSON response
    result := map[string]interface{}{
        "id":           sd.ID,
        "reference_id": sd.ReferenceID,
        "title":        sd.Title,
        "creator_id":   sd.CreatorID,
        "created_at":   sd.CreatedAt,
        "updated_at":   sd.UpdatedAt,
    }

    // Only include description if it's not nil
    if sd.Description != nil {
        result["description"] = *sd.Description
    }

    // Only include creator if it has been populated (has a username, indicating it was preloaded)
    if sd.Creator.Username != "" {
        result["creator"] = sd.Creator
    }

    // Only include epics if they have been populated
    if len(sd.Epics) > 0 {
        result["epics"] = sd.Epics
    }

    return json.Marshal(result)
}
```

#### Epic Model Extension

```go
// Добавить в существующую Epic модель в internal/models/epic.go:
type Epic struct {
    // ... существующие поля ...
    
    // SteeringDocuments содержит связанные steering документы
    // @Description List of steering documents linked to this epic (populated when requested with ?include=steering_documents)
    SteeringDocuments []SteeringDocument `gorm:"many2many:epic_steering_documents;" json:"steering_documents,omitempty"`
}

// Также обновить MarshalJSON метод в Epic модели:
func (e *Epic) MarshalJSON() ([]byte, error) {
    // ... существующий код ...
    
    // Only include steering_documents if they have been populated
    if len(e.SteeringDocuments) > 0 {
        result["steering_documents"] = e.SteeringDocuments
    }
    
    return json.Marshal(result)
}
```

#### Models Registration

```go
// Добавить в internal/models/models.go в функцию AllModels():
func AllModels() []interface{} {
    return []interface{}{
        &User{},
        &Epic{},
        &UserStory{},
        &AcceptanceCriteria{},
        &RequirementType{},
        &RelationshipType{},
        &Requirement{},
        &RequirementRelationship{},
        &Comment{},
        &StatusModel{},
        &Status{},
        &StatusTransition{},
        &PersonalAccessToken{},
        &SteeringDocument{}, // Добавить новую модель
    }
}
```

## Components and Interfaces

### Repository Layer

#### SteeringDocumentRepository Interface

```go
package repository

import (
    "github.com/google/uuid"
    "product-requirements-management/internal/models"
)

type SteeringDocumentRepository interface {
    Create(doc *models.SteeringDocument) error
    GetByID(id uuid.UUID) (*models.SteeringDocument, error)
    GetByReferenceID(referenceID string) (*models.SteeringDocument, error)
    Update(doc *models.SteeringDocument) error
    Delete(id uuid.UUID) error
    List(filters SteeringDocumentFilters) ([]models.SteeringDocument, int64, error)
    Search(query string) ([]models.SteeringDocument, error)
    GetByEpicID(epicID uuid.UUID) ([]models.SteeringDocument, error)
    LinkToEpic(steeringDocumentID, epicID uuid.UUID) error
    UnlinkFromEpic(steeringDocumentID, epicID uuid.UUID) error
}

type SteeringDocumentFilters struct {
    CreatorID *uuid.UUID
    Search    string
    Limit     int
    Offset    int
    OrderBy   string
}
```

#### SteeringDocumentRepository Implementation

```go
package repository

import (
    "fmt"
    "github.com/google/uuid"
    "gorm.io/gorm"
    "product-requirements-management/internal/models"
)

type steeringDocumentRepository struct {
    db *gorm.DB
}

func NewSteeringDocumentRepository(db *gorm.DB) SteeringDocumentRepository {
    return &steeringDocumentRepository{db: db}
}

func (r *steeringDocumentRepository) Create(doc *models.SteeringDocument) error {
    return r.db.Create(doc).Error
}

func (r *steeringDocumentRepository) GetByID(id uuid.UUID) (*models.SteeringDocument, error) {
    var doc models.SteeringDocument
    err := r.db.Preload("Creator").First(&doc, "id = ?", id).Error
    if err != nil {
        return nil, err
    }
    return &doc, nil
}

func (r *steeringDocumentRepository) GetByReferenceID(referenceID string) (*models.SteeringDocument, error) {
    var doc models.SteeringDocument
    err := r.db.Preload("Creator").First(&doc, "reference_id = ?", referenceID).Error
    if err != nil {
        return nil, err
    }
    return &doc, nil
}

func (r *steeringDocumentRepository) Update(doc *models.SteeringDocument) error {
    return r.db.Save(doc).Error
}

func (r *steeringDocumentRepository) Delete(id uuid.UUID) error {
    return r.db.Delete(&models.SteeringDocument{}, "id = ?", id).Error
}

func (r *steeringDocumentRepository) List(filters SteeringDocumentFilters) ([]models.SteeringDocument, int64, error) {
    var docs []models.SteeringDocument
    var total int64
    
    query := r.db.Model(&models.SteeringDocument{}).Preload("Creator")
    
    if filters.CreatorID != nil {
        query = query.Where("creator_id = ?", *filters.CreatorID)
    }
    
    if filters.Search != "" {
        query = query.Where("to_tsvector('english', title || ' ' || COALESCE(description, '')) @@ plainto_tsquery('english', ?)", filters.Search)
    }
    
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    
    if filters.OrderBy != "" {
        query = query.Order(filters.OrderBy)
    } else {
        query = query.Order("created_at DESC")
    }
    
    if filters.Limit > 0 {
        query = query.Limit(filters.Limit)
    }
    
    if filters.Offset > 0 {
        query = query.Offset(filters.Offset)
    }
    
    err := query.Find(&docs).Error
    return docs, total, err
}

func (r *steeringDocumentRepository) Search(query string) ([]models.SteeringDocument, error) {
    var docs []models.SteeringDocument
    err := r.db.Preload("Creator").
        Where("to_tsvector('english', title || ' ' || COALESCE(description, '')) @@ plainto_tsquery('english', ?)", query).
        Order("ts_rank(to_tsvector('english', title || ' ' || COALESCE(description, '')), plainto_tsquery('english', ?)) DESC", query).
        Find(&docs).Error
    return docs, err
}

func (r *steeringDocumentRepository) GetByEpicID(epicID uuid.UUID) ([]models.SteeringDocument, error) {
    var docs []models.SteeringDocument
    err := r.db.Preload("Creator").
        Joins("JOIN epic_steering_documents esd ON esd.steering_document_id = steering_documents.id").
        Where("esd.epic_id = ?", epicID).
        Find(&docs).Error
    return docs, err
}

func (r *steeringDocumentRepository) LinkToEpic(steeringDocumentID, epicID uuid.UUID) error {
    // Проверяем, что связь не существует
    var count int64
    r.db.Table("epic_steering_documents").
        Where("epic_id = ? AND steering_document_id = ?", epicID, steeringDocumentID).
        Count(&count)
    
    if count > 0 {
        return fmt.Errorf("link already exists")
    }
    
    return r.db.Exec("INSERT INTO epic_steering_documents (epic_id, steering_document_id) VALUES (?, ?)", 
        epicID, steeringDocumentID).Error
}

func (r *steeringDocumentRepository) UnlinkFromEpic(steeringDocumentID, epicID uuid.UUID) error {
    return r.db.Exec("DELETE FROM epic_steering_documents WHERE epic_id = ? AND steering_document_id = ?", 
        epicID, steeringDocumentID).Error
}
```

### Service Layer

#### SteeringDocumentService Interface

```go
package service

import (
    "github.com/google/uuid"
    "product-requirements-management/internal/models"
)

type SteeringDocumentService interface {
    CreateSteeringDocument(req CreateSteeringDocumentRequest) (*models.SteeringDocument, error)
    GetSteeringDocumentByID(id uuid.UUID) (*models.SteeringDocument, error)
    GetSteeringDocumentByReferenceID(referenceID string) (*models.SteeringDocument, error)
    UpdateSteeringDocument(id uuid.UUID, req UpdateSteeringDocumentRequest) (*models.SteeringDocument, error)
    DeleteSteeringDocument(id uuid.UUID) error
    ListSteeringDocuments(filters SteeringDocumentFilters) ([]models.SteeringDocument, int64, error)
    SearchSteeringDocuments(query string) ([]models.SteeringDocument, error)
    GetSteeringDocumentsByEpicID(epicID uuid.UUID) ([]models.SteeringDocument, error)
    LinkSteeringDocumentToEpic(steeringDocumentID, epicID uuid.UUID) error
    UnlinkSteeringDocumentFromEpic(steeringDocumentID, epicID uuid.UUID) error
}

type CreateSteeringDocumentRequest struct {
    Title       string     `json:"title" binding:"required,max=500"`
    Description *string    `json:"description,omitempty"`
    CreatorID   uuid.UUID  `json:"creator_id"`
}

type UpdateSteeringDocumentRequest struct {
    Title       *string `json:"title,omitempty"`
    Description *string `json:"description,omitempty"`
}

type SteeringDocumentFilters struct {
    CreatorID *uuid.UUID `json:"creator_id,omitempty"`
    Search    string     `json:"search,omitempty"`
    Limit     int        `json:"limit,omitempty"`
    Offset    int        `json:"offset,omitempty"`
    OrderBy   string     `json:"order_by,omitempty"`
}
```

#### SteeringDocumentService Implementation

```go
package service

import (
    "errors"
    "github.com/google/uuid"
    "product-requirements-management/internal/models"
    "product-requirements-management/internal/repository"
)

var (
    ErrSteeringDocumentNotFound = errors.New("steering document not found")
    ErrEpicNotFound            = errors.New("epic not found")
    ErrLinkAlreadyExists       = errors.New("link already exists")
)

type steeringDocumentService struct {
    steeringDocumentRepo repository.SteeringDocumentRepository
    epicRepo            repository.EpicRepository
    userRepo            repository.UserRepository
}

func NewSteeringDocumentService(
    steeringDocumentRepo repository.SteeringDocumentRepository,
    epicRepo repository.EpicRepository,
    userRepo repository.UserRepository,
) SteeringDocumentService {
    return &steeringDocumentService{
        steeringDocumentRepo: steeringDocumentRepo,
        epicRepo:            epicRepo,
        userRepo:            userRepo,
    }
}

func (s *steeringDocumentService) CreateSteeringDocument(req CreateSteeringDocumentRequest) (*models.SteeringDocument, error) {
    // Проверяем существование создателя
    _, err := s.userRepo.GetByID(req.CreatorID)
    if err != nil {
        return nil, ErrUserNotFound
    }
    
    doc := &models.SteeringDocument{
        Title:       req.Title,
        Description: req.Description,
        CreatorID:   req.CreatorID,
    }
    
    if err := s.steeringDocumentRepo.Create(doc); err != nil {
        return nil, err
    }
    
    return s.steeringDocumentRepo.GetByID(doc.ID)
}

func (s *steeringDocumentService) GetSteeringDocumentByID(id uuid.UUID) (*models.SteeringDocument, error) {
    return s.steeringDocumentRepo.GetByID(id)
}

func (s *steeringDocumentService) GetSteeringDocumentByReferenceID(referenceID string) (*models.SteeringDocument, error) {
    return s.steeringDocumentRepo.GetByReferenceID(referenceID)
}

func (s *steeringDocumentService) UpdateSteeringDocument(id uuid.UUID, req UpdateSteeringDocumentRequest) (*models.SteeringDocument, error) {
    doc, err := s.steeringDocumentRepo.GetByID(id)
    if err != nil {
        return nil, ErrSteeringDocumentNotFound
    }
    
    if req.Title != nil {
        doc.Title = *req.Title
    }
    
    if req.Description != nil {
        doc.Description = req.Description
    }
    
    if err := s.steeringDocumentRepo.Update(doc); err != nil {
        return nil, err
    }
    
    return s.steeringDocumentRepo.GetByID(id)
}

func (s *steeringDocumentService) DeleteSteeringDocument(id uuid.UUID) error {
    _, err := s.steeringDocumentRepo.GetByID(id)
    if err != nil {
        return ErrSteeringDocumentNotFound
    }
    
    return s.steeringDocumentRepo.Delete(id)
}

func (s *steeringDocumentService) ListSteeringDocuments(filters SteeringDocumentFilters) ([]models.SteeringDocument, int64, error) {
    repoFilters := repository.SteeringDocumentFilters{
        CreatorID: filters.CreatorID,
        Search:    filters.Search,
        Limit:     filters.Limit,
        Offset:    filters.Offset,
        OrderBy:   filters.OrderBy,
    }
    
    return s.steeringDocumentRepo.List(repoFilters)
}

func (s *steeringDocumentService) SearchSteeringDocuments(query string) ([]models.SteeringDocument, error) {
    return s.steeringDocumentRepo.Search(query)
}

func (s *steeringDocumentService) GetSteeringDocumentsByEpicID(epicID uuid.UUID) ([]models.SteeringDocument, error) {
    // Проверяем существование эпика
    _, err := s.epicRepo.GetByID(epicID)
    if err != nil {
        return nil, ErrEpicNotFound
    }
    
    return s.steeringDocumentRepo.GetByEpicID(epicID)
}

func (s *steeringDocumentService) LinkSteeringDocumentToEpic(steeringDocumentID, epicID uuid.UUID) error {
    // Проверяем существование документа
    _, err := s.steeringDocumentRepo.GetByID(steeringDocumentID)
    if err != nil {
        return ErrSteeringDocumentNotFound
    }
    
    // Проверяем существование эпика
    _, err = s.epicRepo.GetByID(epicID)
    if err != nil {
        return ErrEpicNotFound
    }
    
    return s.steeringDocumentRepo.LinkToEpic(steeringDocumentID, epicID)
}

func (s *steeringDocumentService) UnlinkSteeringDocumentFromEpic(steeringDocumentID, epicID uuid.UUID) error {
    return s.steeringDocumentRepo.UnlinkFromEpic(steeringDocumentID, epicID)
}
```

### Handler Layer

#### REST API Handler

```go
package handlers

import (
    "errors"
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    
    "product-requirements-management/internal/auth"
    "product-requirements-management/internal/service"
)

type SteeringDocumentHandler struct {
    steeringDocumentService service.SteeringDocumentService
}

func NewSteeringDocumentHandler(steeringDocumentService service.SteeringDocumentService) *SteeringDocumentHandler {
    return &SteeringDocumentHandler{
        steeringDocumentService: steeringDocumentService,
    }
}

// CreateSteeringDocument handles POST /api/v1/steering-documents
func (h *SteeringDocumentHandler) CreateSteeringDocument(c *gin.Context) {
    var req service.CreateSteeringDocumentRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": gin.H{
                "code":    "VALIDATION_ERROR",
                "message": "Invalid request body: " + err.Error(),
            },
        })
        return
    }
    
    // Получаем ID текущего пользователя
    creatorID, ok := auth.GetCurrentUserID(c)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{
            "error": gin.H{
                "code":    "AUTHENTICATION_REQUIRED",
                "message": "User authentication required",
            },
        })
        return
    }
    
    req.CreatorID = uuid.MustParse(creatorID)
    
    doc, err := h.steeringDocumentService.CreateSteeringDocument(req)
    if err != nil {
        switch {
        case errors.Is(err, service.ErrUserNotFound):
            c.JSON(http.StatusBadRequest, gin.H{
                "error": gin.H{
                    "code":    "ENTITY_NOT_FOUND",
                    "message": "Creator not found",
                },
            })
        default:
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": gin.H{
                    "code":    "INTERNAL_ERROR",
                    "message": "Failed to create steering document",
                },
            })
        }
        return
    }
    
    c.JSON(http.StatusCreated, doc)
}

// GetSteeringDocument handles GET /api/v1/steering-documents/:id
func (h *SteeringDocumentHandler) GetSteeringDocument(c *gin.Context) {
    idParam := c.Param("id")
    
    // Пытаемся парсить как UUID
    if id, err := uuid.Parse(idParam); err == nil {
        doc, err := h.steeringDocumentService.GetSteeringDocumentByID(id)
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{
                "error": gin.H{
                    "code":    "ENTITY_NOT_FOUND",
                    "message": "Steering document not found",
                },
            })
            return
        }
        c.JSON(http.StatusOK, doc)
        return
    }
    
    // Пытаемся найти по reference_id
    doc, err := h.steeringDocumentService.GetSteeringDocumentByReferenceID(idParam)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "error": gin.H{
                "code":    "ENTITY_NOT_FOUND",
                "message": "Steering document not found",
            },
        })
        return
    }
    
    c.JSON(http.StatusOK, doc)
}

// UpdateSteeringDocument handles PUT /api/v1/steering-documents/:id
func (h *SteeringDocumentHandler) UpdateSteeringDocument(c *gin.Context) {
    idParam := c.Param("id")
    
    var id uuid.UUID
    var err error
    
    // Пытаемся парсить как UUID
    if id, err = uuid.Parse(idParam); err != nil {
        // Пытаемся найти по reference_id
        doc, err := h.steeringDocumentService.GetSteeringDocumentByReferenceID(idParam)
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{
                "error": gin.H{
                    "code":    "ENTITY_NOT_FOUND",
                    "message": "Steering document not found",
                },
            })
            return
        }
        id = doc.ID
    }
    
    var req service.UpdateSteeringDocumentRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": gin.H{
                "code":    "VALIDATION_ERROR",
                "message": "Invalid request body: " + err.Error(),
            },
        })
        return
    }
    
    doc, err := h.steeringDocumentService.UpdateSteeringDocument(id, req)
    if err != nil {
        switch {
        case errors.Is(err, service.ErrSteeringDocumentNotFound):
            c.JSON(http.StatusNotFound, gin.H{
                "error": gin.H{
                    "code":    "ENTITY_NOT_FOUND",
                    "message": "Steering document not found",
                },
            })
        default:
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": gin.H{
                    "code":    "INTERNAL_ERROR",
                    "message": "Failed to update steering document",
                },
            })
        }
        return
    }
    
    c.JSON(http.StatusOK, doc)
}

// DeleteSteeringDocument handles DELETE /api/v1/steering-documents/:id
func (h *SteeringDocumentHandler) DeleteSteeringDocument(c *gin.Context) {
    idParam := c.Param("id")
    
    var id uuid.UUID
    var err error
    
    // Пытаемся парсить как UUID
    if id, err = uuid.Parse(idParam); err != nil {
        // Пытаемся найти по reference_id
        doc, err := h.steeringDocumentService.GetSteeringDocumentByReferenceID(idParam)
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{
                "error": gin.H{
                    "code":    "ENTITY_NOT_FOUND",
                    "message": "Steering document not found",
                },
            })
            return
        }
        id = doc.ID
    }
    
    if err := h.steeringDocumentService.DeleteSteeringDocument(id); err != nil {
        switch {
        case errors.Is(err, service.ErrSteeringDocumentNotFound):
            c.JSON(http.StatusNotFound, gin.H{
                "error": gin.H{
                    "code":    "ENTITY_NOT_FOUND",
                    "message": "Steering document not found",
                },
            })
        default:
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": gin.H{
                    "code":    "INTERNAL_ERROR",
                    "message": "Failed to delete steering document",
                },
            })
        }
        return
    }
    
    c.JSON(http.StatusNoContent, nil)
}

// ListSteeringDocuments handles GET /api/v1/steering-documents
func (h *SteeringDocumentHandler) ListSteeringDocuments(c *gin.Context) {
    var filters service.SteeringDocumentFilters
    
    // Парсим query параметры
    if creatorIDStr := c.Query("creator_id"); creatorIDStr != "" {
        if creatorID, err := uuid.Parse(creatorIDStr); err == nil {
            filters.CreatorID = &creatorID
        }
    }
    
    filters.Search = c.Query("search")
    filters.OrderBy = c.Query("order_by")
    
    if limitStr := c.Query("limit"); limitStr != "" {
        if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
            filters.Limit = limit
        }
    }
    
    if offsetStr := c.Query("offset"); offsetStr != "" {
        if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
            filters.Offset = offset
        }
    }
    
    docs, total, err := h.steeringDocumentService.ListSteeringDocuments(filters)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": gin.H{
                "code":    "INTERNAL_ERROR",
                "message": "Failed to list steering documents",
            },
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "data":        docs,
        "total_count": total,
        "limit":       filters.Limit,
        "offset":      filters.Offset,
    })
}

// LinkSteeringDocumentToEpic handles POST /api/v1/epics/:epic_id/steering-documents/:doc_id
func (h *SteeringDocumentHandler) LinkSteeringDocumentToEpic(c *gin.Context) {
    epicIDParam := c.Param("epic_id")
    docIDParam := c.Param("doc_id")
    
    // Парсим epic_id
    var epicID uuid.UUID
    var err error
    if epicID, err = uuid.Parse(epicIDParam); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": gin.H{
                "code":    "VALIDATION_ERROR",
                "message": "Invalid epic_id format",
            },
        })
        return
    }
    
    // Парсим doc_id
    var docID uuid.UUID
    if docID, err = uuid.Parse(docIDParam); err != nil {
        // Пытаемся найти по reference_id
        doc, err := h.steeringDocumentService.GetSteeringDocumentByReferenceID(docIDParam)
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{
                "error": gin.H{
                    "code":    "ENTITY_NOT_FOUND",
                    "message": "Steering document not found",
                },
            })
            return
        }
        docID = doc.ID
    }
    
    if err := h.steeringDocumentService.LinkSteeringDocumentToEpic(docID, epicID); err != nil {
        switch {
        case errors.Is(err, service.ErrSteeringDocumentNotFound):
            c.JSON(http.StatusNotFound, gin.H{
                "error": gin.H{
                    "code":    "ENTITY_NOT_FOUND",
                    "message": "Steering document not found",
                },
            })
        case errors.Is(err, service.ErrEpicNotFound):
            c.JSON(http.StatusNotFound, gin.H{
                "error": gin.H{
                    "code":    "ENTITY_NOT_FOUND",
                    "message": "Epic not found",
                },
            })
        case errors.Is(err, service.ErrLinkAlreadyExists):
            c.JSON(http.StatusConflict, gin.H{
                "error": gin.H{
                    "code":    "CONFLICT",
                    "message": "Link already exists",
                },
            })
        default:
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": gin.H{
                    "code":    "INTERNAL_ERROR",
                    "message": "Failed to link steering document to epic",
                },
            })
        }
        return
    }
    
    c.JSON(http.StatusCreated, gin.H{
        "message": "Successfully linked steering document to epic",
    })
}

// UnlinkSteeringDocumentFromEpic handles DELETE /api/v1/epics/:epic_id/steering-documents/:doc_id
func (h *SteeringDocumentHandler) UnlinkSteeringDocumentFromEpic(c *gin.Context) {
    epicIDParam := c.Param("epic_id")
    docIDParam := c.Param("doc_id")
    
    // Парсим epic_id
    var epicID uuid.UUID
    var err error
    if epicID, err = uuid.Parse(epicIDParam); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": gin.H{
                "code":    "VALIDATION_ERROR",
                "message": "Invalid epic_id format",
            },
        })
        return
    }
    
    // Парсим doc_id
    var docID uuid.UUID
    if docID, err = uuid.Parse(docIDParam); err != nil {
        // Пытаемся найти по reference_id
        doc, err := h.steeringDocumentService.GetSteeringDocumentByReferenceID(docIDParam)
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{
                "error": gin.H{
                    "code":    "ENTITY_NOT_FOUND",
                    "message": "Steering document not found",
                },
            })
            return
        }
        docID = doc.ID
    }
    
    if err := h.steeringDocumentService.UnlinkSteeringDocumentFromEpic(docID, epicID); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": gin.H{
                "code":    "INTERNAL_ERROR",
                "message": "Failed to unlink steering document from epic",
            },
        })
        return
    }
    
    c.JSON(http.StatusNoContent, nil)
}

// GetEpicSteeringDocuments handles GET /api/v1/epics/:id/steering-documents
func (h *SteeringDocumentHandler) GetEpicSteeringDocuments(c *gin.Context) {
    epicIDParam := c.Param("id")
    
    var epicID uuid.UUID
    var err error
    if epicID, err = uuid.Parse(epicIDParam); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": gin.H{
                "code":    "VALIDATION_ERROR",
                "message": "Invalid epic_id format",
            },
        })
        return
    }
    
    docs, err := h.steeringDocumentService.GetSteeringDocumentsByEpicID(epicID)
    if err != nil {
        switch {
        case errors.Is(err, service.ErrEpicNotFound):
            c.JSON(http.StatusNotFound, gin.H{
                "error": gin.H{
                    "code":    "ENTITY_NOT_FOUND",
                    "message": "Epic not found",
                },
            })
        default:
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": gin.H{
                    "code":    "INTERNAL_ERROR",
                    "message": "Failed to get epic steering documents",
                },
            })
        }
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "data": docs,
    })
}
```

#### MCP Tool Schemas Extension

```go
// Добавить в internal/handlers/mcp_tool_schemas.go в функцию GetSupportedTools():

{
    Name:        "list_steering_documents",
    Title:       "List Steering Documents",
    Description: "List all steering documents with optional filtering and pagination",
    InputSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "creator_id": map[string]interface{}{
                "type":        "string",
                "description": "UUID of the creator to filter by (optional)",
                "format":      "uuid",
            },
            "search": map[string]interface{}{
                "type":        "string",
                "description": "Search query string (optional)",
            },
            "limit": map[string]interface{}{
                "type":        "integer",
                "description": "Maximum number of results to return (default: 50, max: 100)",
                "minimum":     1,
                "maximum":     100,
                "default":     50,
            },
            "offset": map[string]interface{}{
                "type":        "integer",
                "description": "Number of results to skip for pagination (default: 0)",
                "minimum":     0,
                "default":     0,
            },
        },
        "required": []string{},
    },
},
{
    Name:        "create_steering_document",
    Title:       "Create Steering Document",
    Description: "Create a new steering document. The creator is automatically set to the authenticated user.",
    InputSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "title": map[string]interface{}{
                "type":        "string",
                "description": "Title of the steering document (required, max 500 characters)",
                "maxLength":   500,
            },
            "description": map[string]interface{}{
                "type":        "string",
                "description": "Detailed description of the steering document (optional, max 50000 characters)",
                "maxLength":   50000,
            },
        },
        "required": []string{"title"},
    },
},
{
    Name:        "get_steering_document",
    Title:       "Get Steering Document",
    Description: "Get a steering document by ID or reference ID",
    InputSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "steering_document_id": map[string]interface{}{
                "type":        "string",
                "description": "UUID or reference ID (STD-XXX) of the steering document",
            },
        },
        "required": []string{"steering_document_id"},
    },
},
{
    Name:        "update_steering_document",
    Title:       "Update Steering Document",
    Description: "Update an existing steering document",
    InputSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "steering_document_id": map[string]interface{}{
                "type":        "string",
                "description": "UUID or reference ID (STD-XXX) of the steering document to update",
            },
            "title": map[string]interface{}{
                "type":        "string",
                "description": "New title of the steering document (optional, max 500 characters)",
                "maxLength":   500,
            },
            "description": map[string]interface{}{
                "type":        "string",
                "description": "New description of the steering document (optional, max 50000 characters)",
                "maxLength":   50000,
            },
        },
        "required": []string{"steering_document_id"},
    },
},
{
    Name:        "link_steering_to_epic",
    Title:       "Link Steering Document to Epic",
    Description: "Create a link between a steering document and an epic",
    InputSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "steering_document_id": map[string]interface{}{
                "type":        "string",
                "description": "UUID or reference ID (STD-XXX) of the steering document",
            },
            "epic_id": map[string]interface{}{
                "type":        "string",
                "description": "UUID or reference ID (EP-XXX) of the epic",
            },
        },
        "required": []string{"steering_document_id", "epic_id"},
    },
},
{
    Name:        "unlink_steering_from_epic",
    Title:       "Unlink Steering Document from Epic",
    Description: "Remove a link between a steering document and an epic",
    InputSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "steering_document_id": map[string]interface{}{
                "type":        "string",
                "description": "UUID or reference ID (STD-XXX) of the steering document",
            },
            "epic_id": map[string]interface{}{
                "type":        "string",
                "description": "UUID or reference ID (EP-XXX) of the epic",
            },
        },
        "required": []string{"steering_document_id", "epic_id"},
    },
},
{
    Name:        "get_epic_steering_documents",
    Title:       "Get Epic Steering Documents",
    Description: "Get all steering documents linked to a specific epic",
    InputSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "epic_id": map[string]interface{}{
                "type":        "string",
                "description": "UUID or reference ID (EP-XXX) of the epic",
            },
        },
        "required": []string{"epic_id"},
    },
},
```

#### MCP Tools Handler Extension

```go
// Добавить в internal/handlers/mcp_tools_handler.go

// Добавить steeringDocumentService в ToolsHandler struct:
type ToolsHandler struct {
    epicService             service.EpicService
    userStoryService        service.UserStoryService
    requirementService      service.RequirementService
    searchService           service.SearchServiceInterface
    steeringDocumentService service.SteeringDocumentService // Добавить новый сервис
}

// Обновить NewToolsHandler конструктор:
func NewToolsHandler(
    epicService service.EpicService,
    userStoryService service.UserStoryService,
    requirementService service.RequirementService,
    searchService service.SearchServiceInterface,
    steeringDocumentService service.SteeringDocumentService, // Добавить параметр
) *ToolsHandler {
    return &ToolsHandler{
        epicService:             epicService,
        userStoryService:        userStoryService,
        requirementService:      requirementService,
        searchService:           searchService,
        steeringDocumentService: steeringDocumentService, // Инициализировать
    }
}

// Добавить в HandleToolsCall switch statement:
case "list_steering_documents":
    return h.handleListSteeringDocuments(ctx, arguments)
case "create_steering_document":
    return h.handleCreateSteeringDocument(ctx, arguments)
case "get_steering_document":
    return h.handleGetSteeringDocument(ctx, arguments)
case "update_steering_document":
    return h.handleUpdateSteeringDocument(ctx, arguments)
case "link_steering_to_epic":
    return h.handleLinkSteeringToEpic(ctx, arguments)
case "unlink_steering_from_epic":
    return h.handleUnlinkSteeringFromEpic(ctx, arguments)
case "get_epic_steering_documents":
    return h.handleGetEpicSteeringDocuments(ctx, arguments)

// Добавить методы обработки:

func (h *ToolsHandler) handleListSteeringDocuments(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    var filters service.SteeringDocumentFilters
    
    if creatorIDStr, ok := args["creator_id"].(string); ok && creatorIDStr != "" {
        if creatorID, err := uuid.Parse(creatorIDStr); err == nil {
            filters.CreatorID = &creatorID
        }
    }
    
    if search, ok := args["search"].(string); ok {
        filters.Search = search
    }
    
    if limitFloat, ok := args["limit"].(float64); ok {
        filters.Limit = int(limitFloat)
    }
    
    if offsetFloat, ok := args["offset"].(float64); ok {
        filters.Offset = int(offsetFloat)
    }
    
    docs, total, err := h.steeringDocumentService.ListSteeringDocuments(filters)
    if err != nil {
        return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to list steering documents: %v", err))
    }
    
    jsonData, err := json.MarshalIndent(map[string]interface{}{
        "documents":   docs,
        "total_count": total,
        "limit":       filters.Limit,
        "offset":      filters.Offset,
    }, "", "  ")
    if err != nil {
        return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to marshal steering documents: %v", err))
    }
    
    return &ToolResponse{
        Content: []ContentItem{
            {
                Type: "text",
                Text: fmt.Sprintf("Found %d steering documents", total),
            },
            {
                Type: "text",
                Text: string(jsonData),
            },
        },
    }, nil
}

func (h *ToolsHandler) handleCreateSteeringDocument(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    user, err := getUserFromContext(ctx)
    if err != nil {
        return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user from context: %v", err))
    }
    
    title, ok := args["title"].(string)
    if !ok || title == "" {
        return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'title' argument")
    }
    
    description, _ := args["description"].(string)
    
    req := service.CreateSteeringDocumentRequest{
        Title:       title,
        Description: &description,
        CreatorID:   user.ID,
    }
    
    doc, err := h.steeringDocumentService.CreateSteeringDocument(req)
    if err != nil {
        return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to create steering document: %v", err))
    }
    
    return &ToolResponse{
        Content: []ContentItem{
            {
                Type: "text",
                Text: fmt.Sprintf("Successfully created steering document %s: %s", doc.ReferenceID, doc.Title),
            },
            {
                Type: "text",
                Text: fmt.Sprintf("Steering document data: %+v", doc),
            },
        },
    }, nil
}

func (h *ToolsHandler) handleGetSteeringDocument(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    docIDStr, ok := args["steering_document_id"].(string)
    if !ok || docIDStr == "" {
        return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'steering_document_id' argument")
    }
    
    var doc *models.SteeringDocument
    var err error
    
    // Пытаемся парсить как UUID
    if docID, parseErr := uuid.Parse(docIDStr); parseErr == nil {
        doc, err = h.steeringDocumentService.GetSteeringDocumentByID(docID)
    } else {
        // Пытаемся найти по reference_id
        doc, err = h.steeringDocumentService.GetSteeringDocumentByReferenceID(docIDStr)
    }
    
    if err != nil {
        return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get steering document: %v", err))
    }
    
    jsonData, err := json.MarshalIndent(doc, "", "  ")
    if err != nil {
        return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to marshal steering document: %v", err))
    }
    
    return &ToolResponse{
        Content: []ContentItem{
            {
                Type: "text",
                Text: fmt.Sprintf("Steering document %s: %s", doc.ReferenceID, doc.Title),
            },
            {
                Type: "text",
                Text: string(jsonData),
            },
        },
    }, nil
}

func (h *ToolsHandler) handleUpdateSteeringDocument(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    docIDStr, ok := args["steering_document_id"].(string)
    if !ok || docIDStr == "" {
        return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'steering_document_id' argument")
    }
    
    var docID uuid.UUID
    var err error
    
    // Пытаемся парсить как UUID
    if docID, err = uuid.Parse(docIDStr); err != nil {
        // Пытаемся найти по reference_id
        doc, err := h.steeringDocumentService.GetSteeringDocumentByReferenceID(docIDStr)
        if err != nil {
            return nil, jsonrpc.NewInvalidParamsError("Invalid 'steering_document_id': not a valid UUID or reference ID")
        }
        docID = doc.ID
    }
    
    req := service.UpdateSteeringDocumentRequest{}
    
    if title, ok := args["title"].(string); ok && title != "" {
        req.Title = &title
    }
    
    if description, ok := args["description"].(string); ok {
        req.Description = &description
    }
    
    doc, err := h.steeringDocumentService.UpdateSteeringDocument(docID, req)
    if err != nil {
        return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to update steering document: %v", err))
    }
    
    return &ToolResponse{
        Content: []ContentItem{
            {
                Type: "text",
                Text: fmt.Sprintf("Successfully updated steering document %s: %s", doc.ReferenceID, doc.Title),
            },
            {
                Type: "text",
                Text: fmt.Sprintf("Steering document data: %+v", doc),
            },
        },
    }, nil
}

func (h *ToolsHandler) handleLinkSteeringToEpic(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    docIDStr, ok := args["steering_document_id"].(string)
    if !ok || docIDStr == "" {
        return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'steering_document_id' argument")
    }
    
    epicIDStr, ok := args["epic_id"].(string)
    if !ok || epicIDStr == "" {
        return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'epic_id' argument")
    }
    
    // Парсим steering_document_id
    var docID uuid.UUID
    var err error
    if docID, err = uuid.Parse(docIDStr); err != nil {
        // Пытаемся найти по reference_id
        doc, err := h.steeringDocumentService.GetSteeringDocumentByReferenceID(docIDStr)
        if err != nil {
            return nil, jsonrpc.NewInvalidParamsError("Invalid 'steering_document_id': not a valid UUID or reference ID")
        }
        docID = doc.ID
    }
    
    // Парсим epic_id
    var epicID uuid.UUID
    if epicID, err = uuid.Parse(epicIDStr); err != nil {
        // Пытаемся найти по reference_id
        epic, err := h.epicService.GetEpicByReferenceID(epicIDStr)
        if err != nil {
            return nil, jsonrpc.NewInvalidParamsError("Invalid 'epic_id': not a valid UUID or reference ID")
        }
        epicID = epic.ID
    }
    
    if err := h.steeringDocumentService.LinkSteeringDocumentToEpic(docID, epicID); err != nil {
        return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to link steering document to epic: %v", err))
    }
    
    return &ToolResponse{
        Content: []ContentItem{
            {
                Type: "text",
                Text: fmt.Sprintf("Successfully linked steering document %s to epic %s", docIDStr, epicIDStr),
            },
        },
    }, nil
}

func (h *ToolsHandler) handleUnlinkSteeringFromEpic(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    docIDStr, ok := args["steering_document_id"].(string)
    if !ok || docIDStr == "" {
        return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'steering_document_id' argument")
    }
    
    epicIDStr, ok := args["epic_id"].(string)
    if !ok || epicIDStr == "" {
        return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'epic_id' argument")
    }
    
    // Парсим steering_document_id
    var docID uuid.UUID
    var err error
    if docID, err = uuid.Parse(docIDStr); err != nil {
        // Пытаемся найти по reference_id
        doc, err := h.steeringDocumentService.GetSteeringDocumentByReferenceID(docIDStr)
        if err != nil {
            return nil, jsonrpc.NewInvalidParamsError("Invalid 'steering_document_id': not a valid UUID or reference ID")
        }
        docID = doc.ID
    }
    
    // Парсим epic_id
    var epicID uuid.UUID
    if epicID, err = uuid.Parse(epicIDStr); err != nil {
        // Пытаемся найти по reference_id
        epic, err := h.epicService.GetEpicByReferenceID(epicIDStr)
        if err != nil {
            return nil, jsonrpc.NewInvalidParamsError("Invalid 'epic_id': not a valid UUID or reference ID")
        }
        epicID = epic.ID
    }
    
    if err := h.steeringDocumentService.UnlinkSteeringDocumentFromEpic(docID, epicID); err != nil {
        return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to unlink steering document from epic: %v", err))
    }
    
    return &ToolResponse{
        Content: []ContentItem{
            {
                Type: "text",
                Text: fmt.Sprintf("Successfully unlinked steering document %s from epic %s", docIDStr, epicIDStr),
            },
        },
    }, nil
}

func (h *ToolsHandler) handleGetEpicSteeringDocuments(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    epicIDStr, ok := args["epic_id"].(string)
    if !ok || epicIDStr == "" {
        return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'epic_id' argument")
    }
    
    var epicID uuid.UUID
    var err error
    
    // Пытаемся парсить как UUID
    if epicID, err = uuid.Parse(epicIDStr); err != nil {
        // Пытаемся найти по reference_id
        epic, err := h.epicService.GetEpicByReferenceID(epicIDStr)
        if err != nil {
            return nil, jsonrpc.NewInvalidParamsError("Invalid 'epic_id': not a valid UUID or reference ID")
        }
        epicID = epic.ID
    }
    
    docs, err := h.steeringDocumentService.GetSteeringDocumentsByEpicID(epicID)
    if err != nil {
        return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get epic steering documents: %v", err))
    }
    
    jsonData, err := json.MarshalIndent(map[string]interface{}{
        "epic_id":   epicIDStr,
        "documents": docs,
        "count":     len(docs),
    }, "", "  ")
    if err != nil {
        return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to marshal epic steering documents: %v", err))
    }
    
    return &ToolResponse{
        Content: []ContentItem{
            {
                Type: "text",
                Text: fmt.Sprintf("Found %d steering documents for epic %s", len(docs), epicIDStr),
            },
            {
                Type: "text",
                Text: string(jsonData),
            },
        },
    }, nil
}
```

## Data Models

### Database Migration

#### Migration File: 000006_add_steering_documents.up.sql

```sql
-- Migration: Add steering documents tables

-- Create sequence for steering document reference IDs
CREATE SEQUENCE steering_document_ref_seq START 1;

-- Function to get next steering document reference ID
CREATE OR REPLACE FUNCTION get_next_steering_document_ref_id() RETURNS VARCHAR(20) AS $
BEGIN
    RETURN 'STD-' || LPAD(nextval('steering_document_ref_seq')::TEXT, 3, '0');
END;
$ LANGUAGE plpgsql;

-- Create steering_documents table
CREATE TABLE steering_documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reference_id VARCHAR(20) UNIQUE NOT NULL DEFAULT get_next_steering_document_ref_id(),
    title VARCHAR(500) NOT NULL,
    description TEXT,
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for steering_documents
CREATE INDEX idx_steering_documents_creator_id ON steering_documents(creator_id);
CREATE INDEX idx_steering_documents_reference_id ON steering_documents(reference_id);
CREATE INDEX idx_steering_documents_uuid ON steering_documents(id);
CREATE INDEX idx_steering_documents_created_at ON steering_documents(created_at);
CREATE INDEX idx_steering_documents_updated_at ON steering_documents(updated_at);

-- Full-text search index for steering_documents
CREATE INDEX idx_steering_documents_search ON steering_documents USING gin(to_tsvector('english', reference_id || ' ' || title || ' ' || COALESCE(description, '')));

-- Create epic_steering_documents junction table
CREATE TABLE epic_steering_documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    epic_id UUID NOT NULL REFERENCES epics(id) ON DELETE CASCADE,
    steering_document_id UUID NOT NULL REFERENCES steering_documents(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(epic_id, steering_document_id)
);

-- Create indexes for epic_steering_documents
CREATE INDEX idx_epic_steering_documents_epic_id ON epic_steering_documents(epic_id);
CREATE INDEX idx_epic_steering_documents_steering_document_id ON epic_steering_documents(steering_document_id);
CREATE INDEX idx_epic_steering_documents_uuid ON epic_steering_documents(id);

-- Create trigger for updated_at column
CREATE TRIGGER update_steering_documents_updated_at 
    BEFORE UPDATE ON steering_documents 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

#### Migration File: 000006_add_steering_documents.down.sql

```sql
-- Drop steering documents tables and sequences

-- Drop triggers
DROP TRIGGER IF EXISTS update_steering_documents_updated_at ON steering_documents;

-- Drop tables
DROP TABLE IF EXISTS epic_steering_documents;
DROP TABLE IF EXISTS steering_documents;

-- Drop function and sequence
DROP FUNCTION IF EXISTS get_next_steering_document_ref_id();
DROP SEQUENCE IF EXISTS steering_document_ref_seq;
```

## Error Handling

### Service Layer Errors

```go
var (
    ErrSteeringDocumentNotFound = errors.New("steering document not found")
    ErrEpicNotFound            = errors.New("epic not found")
    ErrLinkAlreadyExists       = errors.New("link already exists")
    ErrInvalidTitle            = errors.New("invalid title")
    ErrInvalidDescription      = errors.New("invalid description")
    ErrUnauthorizedAccess      = errors.New("unauthorized access")
)
```

### HTTP Error Responses

```go
// Стандартные коды ошибок:
// 400 - VALIDATION_ERROR (неверные данные запроса)
// 401 - AUTHENTICATION_REQUIRED (требуется аутентификация)
// 403 - INSUFFICIENT_PERMISSIONS (недостаточно прав)
// 404 - ENTITY_NOT_FOUND (сущность не найдена)
// 409 - CONFLICT (конфликт, например, связь уже существует)
// 500 - INTERNAL_ERROR (внутренняя ошибка сервера)
```

### MCP Error Responses

```go
// Используются стандартные JSON-RPC 2.0 ошибки:
// jsonrpc.NewInvalidParamsError() - неверные параметры
// jsonrpc.NewInternalError() - внутренняя ошибка
// jsonrpc.NewMethodNotFoundError() - метод не найден
```

## Testing Strategy

### Unit Tests

- Тестирование GORM модели SteeringDocument
- Тестирование service layer логики
- Тестирование repository layer с SQLite
- Тестирование MCP tools handlers

### Integration Tests

- Тестирование REST API endpoints
- Тестирование связей между эпиками и steering документами
- Тестирование full-text search функциональности
- Тестирование с PostgreSQL database

### E2E Tests

- Тестирование полного workflow создания и связывания документов
- Тестирование MCP integration
- Тестирование прав доступа и аутентификации

## Security Considerations

### Authentication & Authorization

- Все REST API endpoints требуют JWT аутентификации
- MCP tools используют PAT аутентификацию
- Проверка прав доступа на уровне service layer
- Пользователи могут редактировать только свои документы (кроме Administrator)

### Data Validation

- Валидация на уровне GORM модели
- Валидация на уровне service layer
- Санитизация входных данных
- Защита от SQL injection через GORM

### Access Control

- Administrator: полный доступ ко всем операциям
- User: может создавать, читать и редактировать свои документы
- Commenter: только чтение документов

## Performance Considerations

### Database Optimization

- Индексы на часто используемые поля (creator_id, reference_id)
- Full-text search индексы для поиска
- Оптимизированные JOIN запросы для связей

### Caching Strategy

- Кэширование часто запрашиваемых документов
- Кэширование результатов поиска
- Инвалидация кэша при обновлениях

### API Performance

- Пагинация для списков документов
- Lazy loading связанных сущностей
- Оптимизация JSON serialization

Этот дизайн обеспечивает полную интеграцию steering документов в существующую архитектуру системы управления требованиями, следуя установленным паттернам и обеспечивая все необходимые функции для управления документами и их связями с эпиками.