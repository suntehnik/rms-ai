package validation

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// RouteInfo represents a route extracted from routes.go
type RouteInfo struct {
	Method string
	Path   string
	Line   int
}

// OpenAPIPathInfo represents a path from OpenAPI specification
type OpenAPIPathInfo struct {
	Path    string
	Methods []string
	Line    int
}

// TestOpenAPIRouteCompleteness validates that all implemented routes are documented
func TestOpenAPIRouteCompleteness(t *testing.T) {
	// Extract routes from routes.go
	routes, err := extractRoutesFromRoutesFile("../../internal/server/routes/routes.go")
	require.NoError(t, err, "Should be able to extract routes from routes.go")

	// Extract paths from OpenAPI specification
	openAPIPaths, err := extractOpenAPIPathsFromYAML("../../docs/openapi-v3.yaml")
	require.NoError(t, err, "Should be able to extract paths from OpenAPI spec")

	// Validate completeness
	t.Run("ImplementedRoutesDocumented", func(t *testing.T) {
		validateImplementedRoutesDocumented(t, routes, openAPIPaths)
	})

	t.Run("DocumentedRoutesImplemented", func(t *testing.T) {
		validateDocumentedRoutesImplemented(t, routes, openAPIPaths)
	})

	t.Run("EntityCoverageAnalysis", func(t *testing.T) {
		analyzeEntityCoverageInTests(t, routes, openAPIPaths)
	})
}

// TestResponseSchemaValidation validates response format consistency
func TestResponseSchemaValidation(t *testing.T) {
	specPath := "../../docs/openapi-v3.yaml"

	// Read OpenAPI specification
	specData, err := os.ReadFile(specPath)
	require.NoError(t, err, "Should be able to read OpenAPI spec")

	var spec map[string]interface{}
	err = yaml.Unmarshal(specData, &spec)
	require.NoError(t, err, "OpenAPI spec should be valid YAML")

	t.Run("StandardResponseFormats", func(t *testing.T) {
		validateStandardResponseFormats(t, spec)
	})

	t.Run("ListResponseConsistency", func(t *testing.T) {
		validateListResponseConsistency(t, spec)
	})

	t.Run("ErrorResponseConsistency", func(t *testing.T) {
		validateErrorResponseConsistency(t, spec)
	})

	t.Run("RequiredSchemasPresent", func(t *testing.T) {
		validateRequiredSchemasPresent(t, spec)
	})
}

// TestAuthenticationDocumentation validates authentication requirements documentation
func TestAuthenticationDocumentation(t *testing.T) {
	specPath := "../../docs/openapi-v3.yaml"

	specData, err := os.ReadFile(specPath)
	require.NoError(t, err, "Should be able to read OpenAPI spec")

	var spec map[string]interface{}
	err = yaml.Unmarshal(specData, &spec)
	require.NoError(t, err, "OpenAPI spec should be valid YAML")

	t.Run("SecuritySchemesDefined", func(t *testing.T) {
		validateSecuritySchemesDefined(t, spec)
	})

	t.Run("PublicEndpointsMarked", func(t *testing.T) {
		validatePublicEndpointsMarked(t, spec)
	})

	t.Run("AdminEndpointsMarked", func(t *testing.T) {
		validateAdminEndpointsMarked(t, spec)
	})

	t.Run("AuthenticationRequirementsConsistent", func(t *testing.T) {
		validateAuthenticationRequirementsConsistent(t, spec)
	})
}

// TestDocumentationCompleteness implements automated completeness checks
func TestDocumentationCompleteness(t *testing.T) {
	specPath := "../../docs/openapi-v3.yaml"

	specData, err := os.ReadFile(specPath)
	require.NoError(t, err, "Should be able to read OpenAPI spec")

	var spec map[string]interface{}
	err = yaml.Unmarshal(specData, &spec)
	require.NoError(t, err, "OpenAPI spec should be valid YAML")

	t.Run("AllEndpointsHaveDescriptions", func(t *testing.T) {
		validateAllEndpointsHaveDescriptions(t, spec)
	})

	t.Run("AllSchemasHaveDescriptions", func(t *testing.T) {
		validateAllSchemasHaveDescriptions(t, spec)
	})

	t.Run("ParameterDocumentationComplete", func(t *testing.T) {
		validateParameterDocumentationComplete(t, spec)
	})

	t.Run("ResponseDocumentationComplete", func(t *testing.T) {
		validateResponseDocumentationComplete(t, spec)
	})
}

// Helper functions for route extraction

func extractRoutesFromRoutesFile(filename string) ([]RouteInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var routes []RouteInfo
	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Regex patterns for different route definitions
	routePatterns := []*regexp.Regexp{
		regexp.MustCompile(`^\s*(\w+)\.(\w+)\("([^"]+)",\s*\w+\.\w+\)`),                 // group.METHOD("path", handler)
		regexp.MustCompile(`^\s*router\.(\w+)\("([^"]+)",\s*\w+\)`),                     // router.METHOD("path", handler)
		regexp.MustCompile(`^\s*(\w+)\.(\w+)\("([^"]+)",\s*\w+\.\w+\(\),\s*\w+\.\w+\)`), // with middleware
		regexp.MustCompile(`^\s*v1\.(\w+)\("([^"]+)",\s*\w+\.\w+\)`),                    // v1.METHOD("path", handler)
	}

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip comments and empty lines
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "//") {
			continue
		}

		// Try each pattern
		for _, pattern := range routePatterns {
			matches := pattern.FindStringSubmatch(line)
			if len(matches) >= 3 {
				var method, path string

				if len(matches) == 4 && matches[1] != "" {
					// Pattern with group name
					method = strings.ToUpper(matches[2])
					path = matches[3]
				} else if len(matches) == 3 {
					// Direct router pattern
					method = strings.ToUpper(matches[1])
					path = matches[2]
				}

				if method != "" && path != "" {
					routes = append(routes, RouteInfo{
						Method: method,
						Path:   normalizeRoutePath(path),
						Line:   lineNum,
					})
				}
				break
			}
		}
	}

	return routes, scanner.Err()
}

func extractOpenAPIPathsFromYAML(filename string) (map[string]OpenAPIPathInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	paths := make(map[string]OpenAPIPathInfo)
	scanner := bufio.NewScanner(file)
	lineNum := 0
	inPaths := false
	currentPath := ""

	pathPattern := regexp.MustCompile(`^\s*(/[^:]*):`)
	methodPattern := regexp.MustCompile(`^\s*(get|post|put|patch|delete|head|options):`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check if we're in the paths section
		if strings.TrimSpace(line) == "paths:" {
			inPaths = true
			continue
		}

		// Check if we've left the paths section
		if inPaths && strings.HasPrefix(line, "components:") {
			break
		}

		if !inPaths {
			continue
		}

		// Check for path definition
		if matches := pathPattern.FindStringSubmatch(line); len(matches) > 1 {
			currentPath = matches[1]
			if _, exists := paths[currentPath]; !exists {
				paths[currentPath] = OpenAPIPathInfo{
					Path:    currentPath,
					Methods: []string{},
					Line:    lineNum,
				}
			}
		}

		// Check for method definition
		if currentPath != "" {
			if matches := methodPattern.FindStringSubmatch(line); len(matches) > 1 {
				method := strings.ToUpper(matches[1])
				pathInfo := paths[currentPath]
				pathInfo.Methods = append(pathInfo.Methods, method)
				paths[currentPath] = pathInfo
			}
		}
	}

	return paths, scanner.Err()
}

func normalizeRoutePath(path string) string {
	// Convert Gin route parameters to OpenAPI format
	// :id -> {id}
	normalized := regexp.MustCompile(`:(\w+)`).ReplaceAllString(path, "{$1}")

	// Add /api/v1 prefix if not present and not a health/auth route
	if !strings.HasPrefix(normalized, "/api/v1") &&
		!strings.HasPrefix(normalized, "/auth") &&
		!strings.HasPrefix(normalized, "/ready") &&
		!strings.HasPrefix(normalized, "/live") &&
		!strings.HasPrefix(normalized, "/docs") &&
		!strings.HasPrefix(normalized, "/swagger") {
		normalized = "/api/v1" + normalized
	}

	return normalized
}

// Validation functions

func validateImplementedRoutesDocumented(t *testing.T, routes []RouteInfo, openAPIPaths map[string]OpenAPIPathInfo) {
	// Group routes by path
	routesByPath := make(map[string][]string)
	for _, route := range routes {
		// Skip health check and documentation routes
		if strings.HasPrefix(route.Path, "/ready") ||
			strings.HasPrefix(route.Path, "/live") ||
			strings.HasPrefix(route.Path, "/docs") ||
			strings.HasPrefix(route.Path, "/swagger") {
			continue
		}

		routesByPath[route.Path] = append(routesByPath[route.Path], route.Method)
	}

	missingCount := 0
	var sortedPaths []string
	for path := range routesByPath {
		sortedPaths = append(sortedPaths, path)
	}
	sort.Strings(sortedPaths)

	for _, path := range sortedPaths {
		methods := routesByPath[path]
		if openAPIPath, exists := openAPIPaths[path]; exists {
			// Check if all methods are documented
			for _, method := range methods {
				if !containsString(openAPIPath.Methods, method) {
					t.Errorf("Method %s %s is implemented but not documented", method, path)
					missingCount++
				}
			}
		} else {
			t.Errorf("Path %s is implemented but not documented (methods: %s)", path, strings.Join(methods, ", "))
			missingCount++
		}
	}

	if missingCount == 0 {
		t.Logf("✅ All %d implemented routes are documented", len(sortedPaths))
	} else {
		t.Errorf("❌ %d routes are missing documentation", missingCount)
	}
}

func validateDocumentedRoutesImplemented(t *testing.T, routes []RouteInfo, openAPIPaths map[string]OpenAPIPathInfo) {
	// Group routes by path
	routesByPath := make(map[string][]string)
	for _, route := range routes {
		routesByPath[route.Path] = append(routesByPath[route.Path], route.Method)
	}

	extraCount := 0
	var sortedPaths []string
	for path := range openAPIPaths {
		sortedPaths = append(sortedPaths, path)
	}
	sort.Strings(sortedPaths)

	for _, path := range sortedPaths {
		openAPIPath := openAPIPaths[path]
		if routeMethods, exists := routesByPath[path]; exists {
			// Check if all documented methods are implemented
			for _, method := range openAPIPath.Methods {
				if !containsString(routeMethods, method) {
					t.Errorf("Method %s %s is documented but not implemented", method, path)
					extraCount++
				}
			}
		} else {
			t.Errorf("Path %s is documented but not implemented (methods: %s)", path, strings.Join(openAPIPath.Methods, ", "))
			extraCount++
		}
	}

	if extraCount == 0 {
		t.Logf("✅ All %d documented routes have implementations", len(sortedPaths))
	} else {
		t.Errorf("❌ %d documented routes are missing implementations", extraCount)
	}
}

func analyzeEntityCoverageInTests(t *testing.T, routes []RouteInfo, openAPIPaths map[string]OpenAPIPathInfo) {
	entities := []string{"epics", "user-stories", "acceptance-criteria", "requirements"}

	// Group routes by path
	routesByPath := make(map[string][]string)
	for _, route := range routes {
		routesByPath[route.Path] = append(routesByPath[route.Path], route.Method)
	}

	for _, entity := range entities {
		t.Run(fmt.Sprintf("Entity_%s", entity), func(t *testing.T) {
			// Standard CRUD operations
			crudOps := map[string]string{
				"POST":   fmt.Sprintf("/api/v1/%s", entity),
				"GET":    fmt.Sprintf("/api/v1/%s", entity),
				"GET_ID": fmt.Sprintf("/api/v1/%s/{id}", entity),
				"PUT":    fmt.Sprintf("/api/v1/%s/{id}", entity),
				"DELETE": fmt.Sprintf("/api/v1/%s/{id}", entity),
			}

			for op, path := range crudOps {
				method := strings.Split(op, "_")[0]
				implemented := false
				documented := false

				if routeMethods, exists := routesByPath[path]; exists {
					implemented = containsString(routeMethods, method)
				}

				if openAPIPath, exists := openAPIPaths[path]; exists {
					documented = containsString(openAPIPath.Methods, method)
				}

				if implemented && documented {
					t.Logf("✅ %s %s - Implemented and documented", method, path)
				} else if implemented && !documented {
					t.Errorf("📝 %s %s - Implemented but not documented", method, path)
				} else if !implemented && documented {
					t.Errorf("🔧 %s %s - Documented but not implemented", method, path)
				}
			}

			// Special operations
			specialOps := []struct {
				path   string
				method string
				desc   string
			}{
				{fmt.Sprintf("/api/v1/%s/{id}/validate-deletion", entity), "GET", "Validate Deletion"},
				{fmt.Sprintf("/api/v1/%s/{id}/delete", entity), "DELETE", "Comprehensive Delete"},
				{fmt.Sprintf("/api/v1/%s/{id}/comments", entity), "GET", "Get Comments"},
				{fmt.Sprintf("/api/v1/%s/{id}/comments", entity), "POST", "Create Comment"},
				{fmt.Sprintf("/api/v1/%s/{id}/comments/inline", entity), "POST", "Create Inline Comment"},
			}

			for _, op := range specialOps {
				implemented := false
				documented := false

				if routeMethods, exists := routesByPath[op.path]; exists {
					implemented = containsString(routeMethods, op.method)
				}

				if openAPIPath, exists := openAPIPaths[op.path]; exists {
					documented = containsString(openAPIPath.Methods, op.method)
				}

				if implemented || documented {
					if implemented && documented {
						t.Logf("✅ %s %s - %s", op.method, op.path, op.desc)
					} else if implemented && !documented {
						t.Errorf("📝 %s %s - %s (not documented)", op.method, op.path, op.desc)
					} else if !implemented && documented {
						t.Errorf("🔧 %s %s - %s (not implemented)", op.method, op.path, op.desc)
					}
				}
			}
		})
	}
}

func validateStandardResponseFormats(t *testing.T, spec map[string]interface{}) {
	components, ok := spec["components"].(map[string]interface{})
	if !ok {
		t.Fatal("OpenAPI spec should have components section")
	}

	schemas, ok := components["schemas"].(map[string]interface{})
	if !ok {
		t.Fatal("Components should have schemas section")
	}

	// Check for standard response schemas
	standardResponses := []string{
		"ListResponse", "ErrorResponse", "HealthCheckResponse",
		"DependencyInfo", "DeletionResult", "ValidationResponse",
	}

	for _, response := range standardResponses {
		assert.Contains(t, schemas, response, "Should have %s schema defined", response)
	}

	// Validate ListResponse structure
	if listResponse, exists := schemas["ListResponse"]; exists {
		listResponseMap, ok := listResponse.(map[string]interface{})
		require.True(t, ok, "ListResponse should be an object")

		properties, ok := listResponseMap["properties"].(map[string]interface{})
		require.True(t, ok, "ListResponse should have properties")

		requiredFields := []string{"data", "total_count", "limit", "offset"}
		for _, field := range requiredFields {
			assert.Contains(t, properties, field, "ListResponse should have %s field", field)
		}
	}
}

func validateListResponseConsistency(t *testing.T, spec map[string]interface{}) {
	components, ok := spec["components"].(map[string]interface{})
	if !ok {
		return
	}

	schemas, ok := components["schemas"].(map[string]interface{})
	if !ok {
		return
	}

	// Find all list response schemas
	listResponseSchemas := []string{
		"EpicListResponse", "UserStoryListResponse", "AcceptanceCriteriaListResponse",
		"RequirementListResponse", "CommentListResponse", "UserListResponse",
		"RequirementTypeListResponse", "RelationshipTypeListResponse",
		"StatusModelListResponse", "StatusListResponse", "StatusTransitionListResponse",
	}

	for _, schemaName := range listResponseSchemas {
		if schema, exists := schemas[schemaName]; exists {
			schemaMap, ok := schema.(map[string]interface{})
			require.True(t, ok, "%s should be an object", schemaName)

			// Check if it uses allOf with ListResponse
			if allOf, exists := schemaMap["allOf"]; exists {
				allOfArray, ok := allOf.([]interface{})
				require.True(t, ok, "%s allOf should be an array", schemaName)

				hasListResponseRef := false
				for _, item := range allOfArray {
					itemMap, ok := item.(map[string]interface{})
					if ok {
						if ref, exists := itemMap["$ref"]; exists {
							if refStr, ok := ref.(string); ok && strings.Contains(refStr, "ListResponse") {
								hasListResponseRef = true
								break
							}
						}
					}
				}

				assert.True(t, hasListResponseRef, "%s should reference ListResponse in allOf", schemaName)
			}
		}
	}
}

func validateErrorResponseConsistency(t *testing.T, spec map[string]interface{}) {
	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		return
	}

	errorStatusCodes := []string{"400", "401", "403", "404", "409", "500"}

	for pathName, pathValue := range paths {
		pathMap, ok := pathValue.(map[string]interface{})
		if !ok {
			continue
		}

		for methodName, methodValue := range pathMap {
			if !isHTTPMethodString(methodName) {
				continue
			}

			methodMap, ok := methodValue.(map[string]interface{})
			if !ok {
				continue
			}

			responses, ok := methodMap["responses"].(map[string]interface{})
			if !ok {
				continue
			}

			// Check error responses use standard format
			for statusCode, responseValue := range responses {
				if containsString(errorStatusCodes, statusCode) {
					responseMap, ok := responseValue.(map[string]interface{})
					if !ok {
						continue
					}

					content, ok := responseMap["content"].(map[string]interface{})
					if !ok {
						continue
					}

					jsonContent, ok := content["application/json"].(map[string]interface{})
					if !ok {
						continue
					}

					schema, ok := jsonContent["schema"].(map[string]interface{})
					if !ok {
						continue
					}

					// Check if it references ErrorResponse or has proper error structure
					if ref, exists := schema["$ref"]; exists {
						if refStr, ok := ref.(string); ok {
							assert.True(t, strings.Contains(refStr, "ErrorResponse"),
								"Error response in %s %s should reference ErrorResponse schema", methodName, pathName)
						}
					}
				}
			}
		}
	}
}

func validateRequiredSchemasPresent(t *testing.T, spec map[string]interface{}) {
	components, ok := spec["components"].(map[string]interface{})
	if !ok {
		t.Fatal("OpenAPI spec should have components section")
	}

	schemas, ok := components["schemas"].(map[string]interface{})
	if !ok {
		t.Fatal("Components should have schemas section")
	}

	// Core entity schemas
	coreSchemas := []string{
		"Epic", "UserStory", "AcceptanceCriteria", "Requirement",
		"User", "Comment", "RequirementType", "RelationshipType",
		"RequirementRelationship", "StatusModel", "Status", "StatusTransition",
	}

	// Request schemas
	requestSchemas := []string{
		"CreateEpicRequest", "UpdateEpicRequest", "CreateUserStoryRequest", "UpdateUserStoryRequest",
		"CreateAcceptanceCriteriaRequest", "UpdateAcceptanceCriteriaRequest",
		"CreateRequirementRequest", "UpdateRequirementRequest", "CreateCommentRequest",
		"CreateInlineCommentRequest", "CreateRelationshipRequest",
		"LoginRequest", "ChangePasswordRequest", "StatusChangeRequest", "AssignmentRequest",
	}

	// Response schemas
	responseSchemas := []string{
		"LoginResponse", "EpicListResponse", "UserStoryListResponse",
		"AcceptanceCriteriaListResponse", "RequirementListResponse", "CommentListResponse",
		"UserListResponse", "SearchResponse", "SearchSuggestionsResponse",
	}

	// Deletion system schemas
	deletionSchemas := []string{
		"DependencyInfo", "DependencyItem", "DeletionResult", "DeletedEntity",
	}

	// Comment system schemas
	commentSchemas := []string{
		"InlineCommentValidationRequest", "InlineCommentPosition", "ValidationResponse",
	}

	allRequiredSchemas := append(coreSchemas, requestSchemas...)
	allRequiredSchemas = append(allRequiredSchemas, responseSchemas...)
	allRequiredSchemas = append(allRequiredSchemas, deletionSchemas...)
	allRequiredSchemas = append(allRequiredSchemas, commentSchemas...)

	missingSchemas := []string{}
	for _, schema := range allRequiredSchemas {
		if _, exists := schemas[schema]; !exists {
			missingSchemas = append(missingSchemas, schema)
		}
	}

	if len(missingSchemas) > 0 {
		t.Errorf("Missing required schemas: %v", missingSchemas)
	} else {
		t.Logf("✅ All %d required schemas are present", len(allRequiredSchemas))
	}
}

func validateSecuritySchemesDefined(t *testing.T, spec map[string]interface{}) {
	components, ok := spec["components"].(map[string]interface{})
	if !ok {
		t.Fatal("OpenAPI spec should have components section")
	}

	securitySchemes, ok := components["securitySchemes"].(map[string]interface{})
	if !ok {
		t.Fatal("Components should have securitySchemes section")
	}

	// Check for BearerAuth security scheme
	assert.Contains(t, securitySchemes, "BearerAuth", "Should have BearerAuth security scheme")

	if bearerAuth, exists := securitySchemes["BearerAuth"]; exists {
		bearerAuthMap, ok := bearerAuth.(map[string]interface{})
		require.True(t, ok, "BearerAuth should be an object")

		assert.Equal(t, "http", bearerAuthMap["type"], "BearerAuth should be http type")
		assert.Equal(t, "bearer", bearerAuthMap["scheme"], "BearerAuth should use bearer scheme")
		assert.Equal(t, "JWT", bearerAuthMap["bearerFormat"], "BearerAuth should specify JWT format")
	}
}

func validatePublicEndpointsMarked(t *testing.T, spec map[string]interface{}) {
	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		return
	}

	publicEndpoints := []string{"/auth/login", "/ready", "/live"}

	for _, endpoint := range publicEndpoints {
		if pathValue, exists := paths[endpoint]; exists {
			pathMap, ok := pathValue.(map[string]interface{})
			if !ok {
				continue
			}

			for methodName, methodValue := range pathMap {
				if !isHTTPMethodString(methodName) {
					continue
				}

				methodMap, ok := methodValue.(map[string]interface{})
				if !ok {
					continue
				}

				// Check if security is empty array or not present (indicating public)
				if security, exists := methodMap["security"]; exists {
					securityArray, ok := security.([]interface{})
					if ok {
						assert.Empty(t, securityArray, "Public endpoint %s %s should have empty security array", methodName, endpoint)
					}
				}
			}
		}
	}
}

func validateAdminEndpointsMarked(t *testing.T, spec map[string]interface{}) {
	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		return
	}

	// Check admin endpoints have proper role requirements
	for pathName, pathValue := range paths {
		if strings.Contains(pathName, "/auth/users") || strings.Contains(pathName, "/config/") {
			pathMap, ok := pathValue.(map[string]interface{})
			if !ok {
				continue
			}

			for methodName, methodValue := range pathMap {
				if !isHTTPMethodString(methodName) {
					continue
				}

				methodMap, ok := methodValue.(map[string]interface{})
				if !ok {
					continue
				}

				// Check for x-required-role extension
				if role, exists := methodMap["x-required-role"]; exists {
					assert.Equal(t, "Administrator", role, "Admin endpoint %s %s should require Administrator role", methodName, pathName)
				} else {
					t.Logf("⚠️  Admin endpoint %s %s should specify required role", methodName, pathName)
				}
			}
		}
	}
}

func validateAuthenticationRequirementsConsistent(t *testing.T, spec map[string]interface{}) {
	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		return
	}

	publicEndpoints := []string{"/auth/login", "/ready", "/live"}

	for pathName, pathValue := range paths {
		pathMap, ok := pathValue.(map[string]interface{})
		if !ok {
			continue
		}

		isPublicEndpoint := false
		for _, publicPath := range publicEndpoints {
			if pathName == publicPath {
				isPublicEndpoint = true
				break
			}
		}

		for methodName, methodValue := range pathMap {
			if !isHTTPMethodString(methodName) {
				continue
			}

			methodMap, ok := methodValue.(map[string]interface{})
			if !ok {
				continue
			}

			if security, exists := methodMap["security"]; exists {
				securityArray, ok := security.([]interface{})
				if ok {
					if isPublicEndpoint {
						assert.Empty(t, securityArray, "Public endpoint %s %s should have empty security", methodName, pathName)
					} else {
						assert.NotEmpty(t, securityArray, "Protected endpoint %s %s should have security requirements", methodName, pathName)
					}
				}
			} else if !isPublicEndpoint {
				// Non-public endpoints should have security requirements
				t.Logf("⚠️  Protected endpoint %s %s should have security requirements", methodName, pathName)
			}
		}
	}
}

func validateAllEndpointsHaveDescriptions(t *testing.T, spec map[string]interface{}) {
	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		return
	}

	missingDescriptions := 0

	for pathName, pathValue := range paths {
		pathMap, ok := pathValue.(map[string]interface{})
		if !ok {
			continue
		}

		for methodName, methodValue := range pathMap {
			if !isHTTPMethodString(methodName) {
				continue
			}

			methodMap, ok := methodValue.(map[string]interface{})
			if !ok {
				continue
			}

			// Check for summary and description
			if _, exists := methodMap["summary"]; !exists {
				t.Errorf("Endpoint %s %s missing summary", methodName, pathName)
				missingDescriptions++
			}

			if _, exists := methodMap["description"]; !exists {
				t.Errorf("Endpoint %s %s missing description", methodName, pathName)
				missingDescriptions++
			}
		}
	}

	if missingDescriptions == 0 {
		t.Logf("✅ All endpoints have descriptions")
	}
}

func validateAllSchemasHaveDescriptions(t *testing.T, spec map[string]interface{}) {
	components, ok := spec["components"].(map[string]interface{})
	if !ok {
		return
	}

	schemas, ok := components["schemas"].(map[string]interface{})
	if !ok {
		return
	}

	missingDescriptions := 0

	for schemaName, schemaValue := range schemas {
		schemaMap, ok := schemaValue.(map[string]interface{})
		if !ok {
			continue
		}

		if _, exists := schemaMap["description"]; !exists {
			t.Errorf("Schema %s missing description", schemaName)
			missingDescriptions++
		}
	}

	if missingDescriptions == 0 {
		t.Logf("✅ All schemas have descriptions")
	}
}

func validateParameterDocumentationComplete(t *testing.T, spec map[string]interface{}) {
	components, ok := spec["components"].(map[string]interface{})
	if !ok {
		return
	}

	parameters, ok := components["parameters"].(map[string]interface{})
	if !ok {
		return
	}

	// Check that common parameters are defined
	commonParameters := []string{
		"EntityIdParam", "LimitParam", "OffsetParam", "OrderByParam",
		"CreatorIdParam", "AssigneeIdParam", "PriorityParam", "IncludeParam",
	}

	for _, param := range commonParameters {
		if paramValue, exists := parameters[param]; exists {
			paramMap, ok := paramValue.(map[string]interface{})
			if ok {
				assert.Contains(t, paramMap, "description", "Parameter %s should have description", param)
				assert.Contains(t, paramMap, "schema", "Parameter %s should have schema", param)
			}
		}
	}
}

func validateResponseDocumentationComplete(t *testing.T, spec map[string]interface{}) {
	components, ok := spec["components"].(map[string]interface{})
	if !ok {
		return
	}

	responses, ok := components["responses"].(map[string]interface{})
	if !ok {
		return
	}

	// Check that standard responses are defined
	standardResponses := []string{
		"Unauthorized", "Forbidden", "NotFound", "ValidationError", "DeletionConflict",
	}

	for _, response := range standardResponses {
		if responseValue, exists := responses[response]; exists {
			responseMap, ok := responseValue.(map[string]interface{})
			if ok {
				assert.Contains(t, responseMap, "description", "Response %s should have description", response)
			}
		}
	}
}

// Helper functions

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func isHTTPMethodString(method string) bool {
	httpMethods := []string{"get", "post", "put", "patch", "delete", "head", "options"}
	method = strings.ToLower(method)

	for _, m := range httpMethods {
		if method == m {
			return true
		}
	}

	return false
}
