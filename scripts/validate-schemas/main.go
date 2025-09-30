package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

func main() {
	fmt.Println("=== OpenAPI Schema and Parameter Validation ===\n")

	// Read OpenAPI specification
	content, err := os.ReadFile("docs/openapi-v3.yaml")
	if err != nil {
		log.Fatalf("Error reading OpenAPI spec: %v", err)
	}

	spec := string(content)

	// Validate schemas
	validateSchemas(spec)

	// Validate parameters
	validateParameters(spec)

	// Validate response formats
	validateResponseFormats(spec)

	// Validate authentication requirements
	validateAuthentication(spec)
}

func validateSchemas(spec string) {
	fmt.Println("1. Schema Validation:")
	fmt.Println(strings.Repeat("=", 50))

	// Required schemas for the API
	requiredSchemas := []string{
		// Core entity schemas
		"Epic", "UserStory", "AcceptanceCriteria", "Requirement",
		"User", "Comment", "RequirementType", "RelationshipType",
		"RequirementRelationship", "StatusModel", "Status", "StatusTransition",

		// Request schemas
		"CreateEpicRequest", "UpdateEpicRequest", "CreateUserStoryRequest", "UpdateUserStoryRequest",
		"CreateAcceptanceCriteriaRequest", "UpdateAcceptanceCriteriaRequest",
		"CreateRequirementRequest", "UpdateRequirementRequest", "CreateCommentRequest",
		"CreateInlineCommentRequest", "CreateRelationshipRequest",
		"LoginRequest", "ChangePasswordRequest", "CreateUserRequest", "UpdateUserRequest",
		"StatusChangeRequest", "AssignmentRequest",

		// Response schemas
		"LoginResponse", "EpicListResponse", "UserStoryListResponse",
		"AcceptanceCriteriaListResponse", "RequirementListResponse", "CommentListResponse",
		"UserListResponse", "SearchResponse", "SearchSuggestionsResponse",
		"RequirementTypeListResponse", "RelationshipTypeListResponse",
		"StatusModelListResponse", "StatusListResponse", "StatusTransitionListResponse",
		"ListResponse", "ErrorResponse", "HealthCheckResponse",

		// Deletion system schemas
		"DependencyInfo", "DependencyItem", "DeletionResult", "DeletedEntity",

		// Comment system schemas
		"InlineCommentValidationRequest", "InlineCommentPosition", "ValidationResponse",

		// Search schemas
		"SearchResult", "HierarchyNode", "EntityPath",
	}

	missingSchemas := []string{}
	for _, schema := range requiredSchemas {
		if !strings.Contains(spec, fmt.Sprintf("%s:", schema)) {
			missingSchemas = append(missingSchemas, schema)
		}
	}

	if len(missingSchemas) == 0 {
		fmt.Println("   ✅ All required schemas are defined")
	} else {
		fmt.Println("   ❌ Missing schemas:")
		for _, schema := range missingSchemas {
			fmt.Printf("      - %s\n", schema)
		}
	}

	// Check for unused schemas (basic check)
	fmt.Println("\n   Schema Usage Analysis:")
	definedSchemas := extractDefinedSchemas(spec)
	referencedSchemas := extractReferencedSchemas(spec)

	unusedCount := 0
	for _, schema := range definedSchemas {
		if !contains(referencedSchemas, schema) {
			if unusedCount == 0 {
				fmt.Println("   ⚠️  Potentially unused schemas:")
			}
			fmt.Printf("      - %s\n", schema)
			unusedCount++
		}
	}

	if unusedCount == 0 {
		fmt.Println("   ✅ All defined schemas are referenced")
	}
}

func validateParameters(spec string) {
	fmt.Println("\n2. Parameter Validation:")
	fmt.Println(strings.Repeat("=", 50))

	// Required parameters
	requiredParams := []string{
		"EntityIdParam", "LimitParam", "OffsetParam", "OrderByParam",
		"CreatorIdParam", "AssigneeIdParam", "PriorityParam", "IncludeParam",
	}

	missingParams := []string{}
	for _, param := range requiredParams {
		if !strings.Contains(spec, fmt.Sprintf("%s:", param)) {
			missingParams = append(missingParams, param)
		}
	}

	if len(missingParams) == 0 {
		fmt.Println("   ✅ All required parameters are defined")
	} else {
		fmt.Println("   ❌ Missing parameters:")
		for _, param := range missingParams {
			fmt.Printf("      - %s\n", param)
		}
	}

	// Validate parameter consistency
	fmt.Println("\n   Parameter Consistency Check:")

	// Check if EntityIdParam is used consistently
	entityIdUsage := strings.Count(spec, "$ref: '#/components/parameters/EntityIdParam'")

	if entityIdUsage > 0 {
		fmt.Printf("   ✅ EntityIdParam referenced %d times\n", entityIdUsage)
	} else {
		fmt.Println("   ❌ EntityIdParam not used consistently")
	}
}

func validateResponseFormats(spec string) {
	fmt.Println("\n3. Response Format Validation:")
	fmt.Println(strings.Repeat("=", 50))

	// Check for standard response references
	standardResponses := []string{
		"Unauthorized", "Forbidden", "NotFound", "ValidationError",
	}

	for _, response := range standardResponses {
		refPattern := fmt.Sprintf("$ref: '#/components/responses/%s'", response)
		if strings.Contains(spec, refPattern) {
			fmt.Printf("   ✅ %s response is used\n", response)
		} else {
			fmt.Printf("   ❌ %s response is not referenced\n", response)
		}
	}

	// Check ListResponse usage
	listResponseUsage := strings.Count(spec, "ListResponse")
	if listResponseUsage > 5 {
		fmt.Printf("   ✅ ListResponse pattern used %d times\n", listResponseUsage)
	} else {
		fmt.Println("   ❌ ListResponse pattern not used consistently")
	}

	// Check ErrorResponse usage
	errorResponseUsage := strings.Count(spec, "ErrorResponse")
	if errorResponseUsage > 10 {
		fmt.Printf("   ✅ ErrorResponse used %d times\n", errorResponseUsage)
	} else {
		fmt.Println("   ❌ ErrorResponse not used consistently")
	}
}

func validateAuthentication(spec string) {
	fmt.Println("\n4. Authentication Documentation:")
	fmt.Println(strings.Repeat("=", 50))

	// Check for security scheme definition
	if strings.Contains(spec, "BearerAuth:") {
		fmt.Println("   ✅ BearerAuth security scheme defined")
	} else {
		fmt.Println("   ❌ BearerAuth security scheme not found")
	}

	// Check for public endpoints (should have security: [])
	publicEndpoints := []string{"/auth/login", "/ready", "/live"}
	for _, endpoint := range publicEndpoints {
		if strings.Contains(spec, endpoint) {
			// Look for security: [] after the endpoint
			endpointIndex := strings.Index(spec, endpoint)
			if endpointIndex != -1 {
				// Check next 500 characters for security: []
				searchArea := spec[endpointIndex:min(endpointIndex+500, len(spec))]
				if strings.Contains(searchArea, "security: []") {
					fmt.Printf("   ✅ %s correctly marked as public\n", endpoint)
				} else {
					fmt.Printf("   ❌ %s should be marked as public (security: [])\n", endpoint)
				}
			}
		}
	}

	// Check for admin-only endpoints
	adminEndpoints := []string{"/auth/users", "/config/"}
	for _, endpoint := range adminEndpoints {
		if strings.Contains(spec, endpoint) {
			// Look for x-required-role: Administrator
			endpointIndex := strings.Index(spec, endpoint)
			if endpointIndex != -1 {
				searchArea := spec[endpointIndex:min(endpointIndex+1000, len(spec))]
				if strings.Contains(searchArea, "x-required-role: Administrator") {
					fmt.Printf("   ✅ %s correctly marked as admin-only\n", endpoint)
				} else {
					fmt.Printf("   ⚠️  %s should specify admin role requirement\n", endpoint)
				}
			}
		}
	}
}

func extractDefinedSchemas(spec string) []string {
	var schemas []string
	lines := strings.Split(spec, "\n")
	inSchemas := false

	for _, line := range lines {
		if strings.TrimSpace(line) == "schemas:" {
			inSchemas = true
			continue
		}

		if inSchemas {
			// Check if we've left the schemas section
			if strings.HasPrefix(line, "  responses:") || strings.HasPrefix(line, "  parameters:") {
				break
			}

			// Look for schema definitions (indented with 6 spaces)
			if strings.HasPrefix(line, "    ") && strings.HasSuffix(line, ":") {
				schemaName := strings.TrimSpace(strings.TrimSuffix(line, ":"))
				if schemaName != "" && !strings.Contains(schemaName, " ") {
					schemas = append(schemas, schemaName)
				}
			}
		}
	}

	return schemas
}

func extractReferencedSchemas(spec string) []string {
	var schemas []string

	// Find all $ref references to schemas
	refPattern := regexp.MustCompile(`\$ref:\s*['"]#/components/schemas/([^'"]+)['"]`)
	matches := refPattern.FindAllStringSubmatch(spec, -1)

	for _, match := range matches {
		if len(match) > 1 {
			schemas = append(schemas, match[1])
		}
	}

	return schemas
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
