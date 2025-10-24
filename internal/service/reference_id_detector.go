package service

import (
	"regexp"
	"strings"
)

// ReferenceIDPattern represents a detected reference ID pattern
type ReferenceIDPattern struct {
	IsReferenceID bool   `json:"is_reference_id"`
	EntityType    string `json:"entity_type,omitempty"` // "epic", "user_story", "requirement", "acceptance_criteria", "steering_document"
	Number        string `json:"number,omitempty"`      // The numeric part (e.g., "119" from "US-119")
	OriginalQuery string `json:"original_query"`
}

// ReferenceIDDetector provides reference ID pattern detection functionality
type ReferenceIDDetector struct {
	patterns map[string]*regexp.Regexp
}

// NewReferenceIDDetector creates a new reference ID detector
func NewReferenceIDDetector() *ReferenceIDDetector {
	return &ReferenceIDDetector{
		patterns: map[string]*regexp.Regexp{
			"epic":                regexp.MustCompile(`^(?i)EP-(\d+)$`),
			"user_story":          regexp.MustCompile(`^(?i)US-(\d+)$`),
			"requirement":         regexp.MustCompile(`^(?i)REQ-(\d+)$`),
			"acceptance_criteria": regexp.MustCompile(`^(?i)AC-(\d+)$`),
			"steering_document":   regexp.MustCompile(`^(?i)STD-(\d+)$`),
		},
	}
}

// DetectPattern analyzes a query string to determine if it matches a reference ID pattern
func (d *ReferenceIDDetector) DetectPattern(query string) ReferenceIDPattern {
	// Clean the query
	cleanQuery := strings.TrimSpace(query)

	// Return non-reference ID pattern for empty queries
	if cleanQuery == "" {
		return ReferenceIDPattern{
			IsReferenceID: false,
			OriginalQuery: query,
		}
	}

	// Check each pattern
	for entityType, pattern := range d.patterns {
		if matches := pattern.FindStringSubmatch(cleanQuery); matches != nil {
			return ReferenceIDPattern{
				IsReferenceID: true,
				EntityType:    entityType,
				Number:        matches[1], // The captured numeric part
				OriginalQuery: query,
			}
		}
	}

	// No pattern matched
	return ReferenceIDPattern{
		IsReferenceID: false,
		OriginalQuery: query,
	}
}

// IsValidReferenceID checks if a string is a valid reference ID for any entity type
func (d *ReferenceIDDetector) IsValidReferenceID(query string) bool {
	pattern := d.DetectPattern(query)
	return pattern.IsReferenceID
}

// GetEntityTypeFromReferenceID returns the entity type for a given reference ID
func (d *ReferenceIDDetector) GetEntityTypeFromReferenceID(query string) string {
	pattern := d.DetectPattern(query)
	if pattern.IsReferenceID {
		return pattern.EntityType
	}
	return ""
}
