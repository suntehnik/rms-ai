package docs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// CompletenessMetrics tracks documentation completeness metrics
type CompletenessMetrics struct {
	TotalEndpoints            int
	DocumentedEndpoints       int
	EndpointsWithSummary      int
	EndpointsWithDescription  int
	EndpointsWithExamples     int
	TotalSchemas              int
	SchemasWithDescription    int
	TotalParameters           int
	ParametersWithDescription int
	MissingDocumentation      []string
	QualityIssues             []string
}

// TestDocumentationCompletenessMetrics provides comprehensive completeness analysis
func TestDocumentationCompletenessMetrics(t *testing.T) {
	specPath := "../../docs/openapi-v3.yaml"
	specData, err := os.ReadFile(specPath)
	require.NoError(t, err, "Should be able to read OpenAPI spec")

	var spec map[string]interface{}
	err = yaml.Unmarshal(specData, &spec)
	require.NoError(t, err, "OpenAPI spec should be valid YAML")

	metrics := analyzeCompletenessMetrics(spec)

	t.Run("EndpointCompleteness", func(t *testing.T) {
		validateEndpointCompleteness(t, metrics)
	})

	t.Run("SchemaCompleteness", func(t *testing.T) {
		validateSchemaCompleteness(t, metrics)
	})

	t.Run("ParameterCompleteness", func(t *testing.T) {
		validateParameterCompleteness(t, metrics)
	})

	t.Run("OverallQuality", func(t *testing.T) {
		validateOverallQuality(t, metrics)
	})

	// Report metrics
	reportCompletenessMetrics(t, metrics)
}

// TestCRUDOperationCompleteness validates CRUD operation coverage
func TestCRUDOperationCompleteness(t *testing.T) {
	specPath := "../../docs/openapi-v3.yaml"
	specData, err := os.ReadFile(specPath)
	require.NoError(t, err, "Should be able to read OpenAPI spec")

	var spec map[string]interface{}
	err = yaml.Unmarshal(specData, &spec)
	require.NoError(t, err, "OpenAPI spec should be valid YAML")

	entities := []string{"epics", "user-stories", "acceptance-criteria", "requirements"}

	for _, entity := range entities {
		t.Run(fmt.Sprintf("CRUD_%s", entity), func(t *testing.T) {
			validateEntityCRUDCompleteness(t, spec, entity)
		})
	}
}

// TestSpecialOperationCompleteness validates special operation coverage
func TestSpecialOperationCompleteness(t *testing.T) {
	specPath := "../../docs/openapi-v3.yaml"
	specData, err := os.ReadFile(specPath)
	require.NoError(t, err, "Should be able to read OpenAPI spec")

	var spec map[string]interface{}
	err = yaml.Unmarshal(specData, &spec)
	require.NoError(t, err, "OpenAPI spec should be valid YAML")

	t.Run("DeletionWorkflowCompleteness", func(t *testing.T) {
		validateDeletionWorkflowCompleteness(t, spec)
	})

	t.Run("CommentSystemCompleteness", func(t *testing.T) {
		validateCommentSystemCompleteness(t, spec)
	})

	t.Run("SearchFunctionalityCompleteness", func(t *testing.T) {
		validateSearchFunctionalityCompleteness(t, spec)
	})

	t.Run("ConfigurationManagementCompleteness", func(t *testing.T) {
		validateConfigurationManagementCompleteness(t, spec)
	})
}

// TestDocumentationConsistency validates consistency across the documentation
func TestDocumentationConsistency(t *testing.T) {
	specPath := "../../docs/openapi-v3.yaml"
	specData, err := os.ReadFile(specPath)
	require.NoError(t, err, "Should be able to read OpenAPI spec")

	var spec map[string]interface{}
	err = yaml.Unmarshal(specData, &spec)
	require.NoError(t, err, "OpenAPI spec should be valid YAML")

	t.Run("ResponseFormatConsistency", func(t *testing.T) {
		validateResponseFormatConsistency(t, spec)
	})

	t.Run("ParameterNamingConsistency", func(t *testing.T) {
		validateParameterNamingConsistency(t, spec)
	})

	t.Run("ErrorResponseConsistency", func(t *testing.T) {
		validateErrorResponseConsistencyInCompleteness(t, spec)
	})

	t.Run("TagConsistency", func(t *testing.T) {
		validateTagConsistency(t, spec)
	})
}

// TestDocumentationQualityStandards validates documentation quality standards
func TestDocumentationQualityStandards(t *testing.T) {
	specPath := "../../docs/openapi-v3.yaml"
	specData, err := os.ReadFile(specPath)
	require.NoError(t, err, "Should be able to read OpenAPI spec")

	var spec map[string]interface{}
	err = yaml.Unmarshal(specData, &spec)
	require.NoError(t, err, "OpenAPI spec should be valid YAML")

	t.Run("DescriptionQuality", func(t *testing.T) {
		validateDescriptionQuality(t, spec)
	})

	t.Run("SchemaValidationRules", func(t *testing.T) {
		validateSchemaValidationRules(t, spec)
	})

	t.Run("SecurityDocumentationQuality", func(t *testing.T) {
		validateSecurityDocumentationQuality(t, spec)
	})
}

// TestGeneratedDocumentationFiles validates generated documentation files
func TestGeneratedDocumentationFiles(t *testing.T) {
	docsDir := "../../docs/generated"

	t.Run("RequiredFilesExist", func(t *testing.T) {
		validateRequiredDocumentationFilesExist(t, docsDir)
	})

	t.Run("FileContentQuality", func(t *testing.T) {
		validateGeneratedFileContentQuality(t, docsDir)
	})

	t.Run("CrossReferenceConsistency", func(t *testing.T) {
		validateCrossReferenceConsistency(t, docsDir)
	})
}

// Analysis functions

func analyzeCompletenessMetrics(spec map[string]interface{}) CompletenessMetrics {
	metrics := CompletenessMetrics{
		MissingDocumentation: []string{},
		QualityIssues:        []string{},
	}

	// Analyze paths
	if paths, ok := spec["paths"].(map[string]interface{}); ok {
		for pathName, pathValue := range paths {
			pathMap, ok := pathValue.(map[string]interface{})
			if !ok {
				continue
			}

			for methodName, methodValue := range pathMap {
				if !isHTTPMethodString(methodName) {
					continue
				}

				metrics.TotalEndpoints++
				methodMap, ok := methodValue.(map[string]interface{})
				if !ok {
					continue
				}

				metrics.DocumentedEndpoints++

				// Check for summary
				if _, exists := methodMap["summary"]; exists {
					metrics.EndpointsWithSummary++
				} else {
					metrics.MissingDocumentation = append(metrics.MissingDocumentation,
						fmt.Sprintf("Missing summary: %s %s", strings.ToUpper(methodName), pathName))
				}

				// Check for description
				if _, exists := methodMap["description"]; exists {
					metrics.EndpointsWithDescription++
				} else {
					metrics.MissingDocumentation = append(metrics.MissingDocumentation,
						fmt.Sprintf("Missing description: %s %s", strings.ToUpper(methodName), pathName))
				}

				// Check for examples in responses
				if hasResponseExamples(methodMap) {
					metrics.EndpointsWithExamples++
				}
			}
		}
	}

	// Analyze schemas
	if components, ok := spec["components"].(map[string]interface{}); ok {
		if schemas, ok := components["schemas"].(map[string]interface{}); ok {
			for schemaName, schemaValue := range schemas {
				metrics.TotalSchemas++
				schemaMap, ok := schemaValue.(map[string]interface{})
				if !ok {
					continue
				}

				if _, exists := schemaMap["description"]; exists {
					metrics.SchemasWithDescription++
				} else {
					metrics.MissingDocumentation = append(metrics.MissingDocumentation,
						fmt.Sprintf("Missing schema description: %s", schemaName))
				}
			}
		}

		// Analyze parameters
		if parameters, ok := components["parameters"].(map[string]interface{}); ok {
			for paramName, paramValue := range parameters {
				metrics.TotalParameters++
				paramMap, ok := paramValue.(map[string]interface{})
				if !ok {
					continue
				}

				if _, exists := paramMap["description"]; exists {
					metrics.ParametersWithDescription++
				} else {
					metrics.MissingDocumentation = append(metrics.MissingDocumentation,
						fmt.Sprintf("Missing parameter description: %s", paramName))
				}
			}
		}
	}

	return metrics
}

func hasResponseExamples(methodMap map[string]interface{}) bool {
	responses, ok := methodMap["responses"].(map[string]interface{})
	if !ok {
		return false
	}

	for _, responseValue := range responses {
		responseMap, ok := responseValue.(map[string]interface{})
		if !ok {
			continue
		}

		if content, exists := responseMap["content"]; exists {
			contentMap, ok := content.(map[string]interface{})
			if !ok {
				continue
			}

			for _, mediaTypeValue := range contentMap {
				mediaTypeMap, ok := mediaTypeValue.(map[string]interface{})
				if !ok {
					continue
				}

				if _, exists := mediaTypeMap["example"]; exists {
					return true
				}
				if _, exists := mediaTypeMap["examples"]; exists {
					return true
				}
			}
		}
	}

	return false
}

// Validation functions

func validateEndpointCompleteness(t *testing.T, metrics CompletenessMetrics) {
	if metrics.TotalEndpoints > 0 {
		summaryPercentage := float64(metrics.EndpointsWithSummary) / float64(metrics.TotalEndpoints) * 100
		descriptionPercentage := float64(metrics.EndpointsWithDescription) / float64(metrics.TotalEndpoints) * 100
		examplePercentage := float64(metrics.EndpointsWithExamples) / float64(metrics.TotalEndpoints) * 100

		assert.GreaterOrEqual(t, summaryPercentage, 95.0, "At least 95%% of endpoints should have summaries")
		assert.GreaterOrEqual(t, descriptionPercentage, 90.0, "At least 90%% of endpoints should have descriptions")
		assert.GreaterOrEqual(t, examplePercentage, 70.0, "At least 70%% of endpoints should have examples")

		t.Logf("Endpoint completeness: Summary %.1f%%, Description %.1f%%, Examples %.1f%%",
			summaryPercentage, descriptionPercentage, examplePercentage)
	}
}

func validateSchemaCompleteness(t *testing.T, metrics CompletenessMetrics) {
	if metrics.TotalSchemas > 0 {
		descriptionPercentage := float64(metrics.SchemasWithDescription) / float64(metrics.TotalSchemas) * 100
		assert.GreaterOrEqual(t, descriptionPercentage, 85.0, "At least 85%% of schemas should have descriptions")

		t.Logf("Schema completeness: Description %.1f%%", descriptionPercentage)
	}
}

func validateParameterCompleteness(t *testing.T, metrics CompletenessMetrics) {
	if metrics.TotalParameters > 0 {
		descriptionPercentage := float64(metrics.ParametersWithDescription) / float64(metrics.TotalParameters) * 100
		assert.GreaterOrEqual(t, descriptionPercentage, 90.0, "At least 90%% of parameters should have descriptions")

		t.Logf("Parameter completeness: Description %.1f%%", descriptionPercentage)
	}
}

func validateOverallQuality(t *testing.T, metrics CompletenessMetrics) {
	// Calculate overall quality score
	totalItems := metrics.TotalEndpoints*2 + metrics.TotalSchemas + metrics.TotalParameters // *2 for summary and description
	documentedItems := metrics.EndpointsWithSummary + metrics.EndpointsWithDescription +
		metrics.SchemasWithDescription + metrics.ParametersWithDescription

	if totalItems > 0 {
		qualityScore := float64(documentedItems) / float64(totalItems) * 100
		assert.GreaterOrEqual(t, qualityScore, 85.0, "Overall documentation quality should be at least 85%%")

		t.Logf("Overall documentation quality: %.1f%%", qualityScore)
	}

	// Check for critical missing documentation
	criticalMissing := 0
	for _, missing := range metrics.MissingDocumentation {
		if strings.Contains(missing, "POST") || strings.Contains(missing, "GET") {
			criticalMissing++
		}
	}

	assert.LessOrEqual(t, criticalMissing, 5, "Should have no more than 5 critical missing documentation items")
}

func validateEntityCRUDCompleteness(t *testing.T, spec map[string]interface{}, entity string) {
	paths, ok := spec["paths"].(map[string]interface{})
	require.True(t, ok, "Should have paths section")

	crudOperations := map[string]string{
		"CREATE": fmt.Sprintf("/api/v1/%s", entity),
		"LIST":   fmt.Sprintf("/api/v1/%s", entity),
		"GET":    fmt.Sprintf("/api/v1/%s/{id}", entity),
		"UPDATE": fmt.Sprintf("/api/v1/%s/{id}", entity),
		"DELETE": fmt.Sprintf("/api/v1/%s/{id}", entity),
	}

	crudMethods := map[string]string{
		"CREATE": "post",
		"LIST":   "get",
		"GET":    "get",
		"UPDATE": "put",
		"DELETE": "delete",
	}

	for operation, path := range crudOperations {
		method := crudMethods[operation]

		if pathValue, exists := paths[path]; exists {
			pathMap, ok := pathValue.(map[string]interface{})
			if ok {
				if _, exists := pathMap[method]; exists {
					t.Logf("✅ %s %s operation documented for %s", operation, method, entity)
				} else {
					t.Errorf("❌ %s %s operation missing for %s", operation, method, entity)
				}
			}
		} else {
			t.Errorf("❌ Path %s missing for %s %s operation", path, entity, operation)
		}
	}
}

func validateDeletionWorkflowCompleteness(t *testing.T, spec map[string]interface{}) {
	paths, ok := spec["paths"].(map[string]interface{})
	require.True(t, ok, "Should have paths section")

	entities := []string{"epics", "user-stories", "acceptance-criteria", "requirements"}

	for _, entity := range entities {
		validatePath := fmt.Sprintf("/api/v1/%s/{id}/validate-deletion", entity)
		deletePath := fmt.Sprintf("/api/v1/%s/{id}/delete", entity)

		// Check validation endpoint
		if pathValue, exists := paths[validatePath]; exists {
			pathMap, ok := pathValue.(map[string]interface{})
			if ok {
				if _, exists := pathMap["get"]; exists {
					t.Logf("✅ Deletion validation documented for %s", entity)
				} else {
					t.Errorf("❌ Deletion validation GET method missing for %s", entity)
				}
			}
		} else {
			t.Errorf("❌ Deletion validation path missing for %s", entity)
		}

		// Check comprehensive deletion endpoint
		if pathValue, exists := paths[deletePath]; exists {
			pathMap, ok := pathValue.(map[string]interface{})
			if ok {
				if _, exists := pathMap["delete"]; exists {
					t.Logf("✅ Comprehensive deletion documented for %s", entity)
				} else {
					t.Errorf("❌ Comprehensive deletion DELETE method missing for %s", entity)
				}
			}
		} else {
			t.Errorf("❌ Comprehensive deletion path missing for %s", entity)
		}
	}
}

func validateCommentSystemCompleteness(t *testing.T, spec map[string]interface{}) {
	paths, ok := spec["paths"].(map[string]interface{})
	require.True(t, ok, "Should have paths section")

	entities := []string{"epics", "user-stories", "acceptance-criteria", "requirements"}

	commentOperations := []struct {
		pathTemplate string
		method       string
		description  string
	}{
		{"/api/v1/%s/{id}/comments", "get", "Get Comments"},
		{"/api/v1/%s/{id}/comments", "post", "Create Comment"},
		{"/api/v1/%s/{id}/comments/inline", "post", "Create Inline Comment"},
		{"/api/v1/%s/{id}/comments/inline/visible", "get", "Get Visible Inline Comments"},
		{"/api/v1/%s/{id}/comments/inline/validate", "post", "Validate Inline Comments"},
	}

	for _, entity := range entities {
		for _, op := range commentOperations {
			path := fmt.Sprintf(op.pathTemplate, entity)

			if pathValue, exists := paths[path]; exists {
				pathMap, ok := pathValue.(map[string]interface{})
				if ok {
					if _, exists := pathMap[op.method]; exists {
						t.Logf("✅ %s documented for %s", op.description, entity)
					} else {
						t.Errorf("❌ %s %s method missing for %s", op.description, op.method, entity)
					}
				}
			} else {
				t.Errorf("❌ %s path missing for %s", op.description, entity)
			}
		}
	}
}

func validateSearchFunctionalityCompleteness(t *testing.T, spec map[string]interface{}) {
	paths, ok := spec["paths"].(map[string]interface{})
	require.True(t, ok, "Should have paths section")

	searchEndpoints := []struct {
		path        string
		method      string
		description string
	}{
		{"/api/v1/search", "get", "Global Search"},
		{"/api/v1/search/suggestions", "get", "Search Suggestions"},
		{"/api/v1/hierarchy", "get", "Hierarchy Root"},
		{"/api/v1/hierarchy/epics/{id}", "get", "Epic Hierarchy"},
		{"/api/v1/hierarchy/user-stories/{id}", "get", "User Story Hierarchy"},
		{"/api/v1/hierarchy/path/{entity_type}/{id}", "get", "Entity Path"},
	}

	for _, endpoint := range searchEndpoints {
		if pathValue, exists := paths[endpoint.path]; exists {
			pathMap, ok := pathValue.(map[string]interface{})
			if ok {
				if _, exists := pathMap[endpoint.method]; exists {
					t.Logf("✅ %s documented", endpoint.description)
				} else {
					t.Errorf("❌ %s %s method missing", endpoint.description, endpoint.method)
				}
			}
		} else {
			t.Errorf("❌ %s path missing", endpoint.description)
		}
	}
}

func validateConfigurationManagementCompleteness(t *testing.T, spec map[string]interface{}) {
	paths, ok := spec["paths"].(map[string]interface{})
	require.True(t, ok, "Should have paths section")

	configEntities := []string{"requirement-types", "relationship-types", "status-models", "statuses", "status-transitions"}

	for _, entity := range configEntities {
		basePath := fmt.Sprintf("/api/v1/config/%s", entity)
		idPath := fmt.Sprintf("/api/v1/config/%s/{id}", entity)

		// Check base CRUD operations
		crudOps := map[string]string{
			"post": "Create",
			"get":  "List",
		}

		if pathValue, exists := paths[basePath]; exists {
			pathMap, ok := pathValue.(map[string]interface{})
			if ok {
				for method, operation := range crudOps {
					if _, exists := pathMap[method]; exists {
						t.Logf("✅ %s %s documented", operation, entity)
					} else {
						t.Errorf("❌ %s %s method missing", operation, entity)
					}
				}
			}
		} else {
			t.Errorf("❌ Base path missing for %s", entity)
		}

		// Check ID-based operations
		idOps := map[string]string{
			"get":    "Get",
			"put":    "Update",
			"delete": "Delete",
		}

		if pathValue, exists := paths[idPath]; exists {
			pathMap, ok := pathValue.(map[string]interface{})
			if ok {
				for method, operation := range idOps {
					if _, exists := pathMap[method]; exists {
						t.Logf("✅ %s %s by ID documented", operation, entity)
					} else {
						t.Errorf("❌ %s %s by ID method missing", operation, entity)
					}
				}
			}
		} else {
			t.Errorf("❌ ID path missing for %s", entity)
		}
	}
}

func validateResponseFormatConsistency(t *testing.T, spec map[string]interface{}) {
	paths, ok := spec["paths"].(map[string]interface{})
	require.True(t, ok, "Should have paths section")

	listEndpoints := []string{}
	errorResponses := []string{}

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

			// Check if it's a list endpoint
			if methodName == "get" && !strings.Contains(pathName, "{id}") {
				listEndpoints = append(listEndpoints, fmt.Sprintf("%s %s", strings.ToUpper(methodName), pathName))
			}

			// Check error responses
			if responses, exists := methodMap["responses"]; exists {
				responsesMap, ok := responses.(map[string]interface{})
				if ok {
					for statusCode := range responsesMap {
						if strings.HasPrefix(statusCode, "4") || strings.HasPrefix(statusCode, "5") {
							errorResponses = append(errorResponses, fmt.Sprintf("%s %s - %s", strings.ToUpper(methodName), pathName, statusCode))
						}
					}
				}
			}
		}
	}

	t.Logf("Found %d list endpoints and %d error responses", len(listEndpoints), len(errorResponses))

	// Basic consistency check - should have reasonable numbers
	assert.Greater(t, len(listEndpoints), 5, "Should have multiple list endpoints")
	assert.Greater(t, len(errorResponses), 10, "Should have multiple error responses documented")
}

func validateParameterNamingConsistency(t *testing.T, spec map[string]interface{}) {
	components, ok := spec["components"].(map[string]interface{})
	if !ok {
		return
	}

	parameters, ok := components["parameters"].(map[string]interface{})
	if !ok {
		return
	}

	// Check for consistent parameter naming
	expectedParams := []string{"EntityIdParam", "LimitParam", "OffsetParam", "OrderByParam"}

	for _, param := range expectedParams {
		if _, exists := parameters[param]; exists {
			t.Logf("✅ Standard parameter %s defined", param)
		} else {
			t.Errorf("❌ Standard parameter %s missing", param)
		}
	}
}

func validateErrorResponseConsistencyInCompleteness(t *testing.T, spec map[string]interface{}) {
	components, ok := spec["components"].(map[string]interface{})
	if !ok {
		return
	}

	responses, ok := components["responses"].(map[string]interface{})
	if !ok {
		return
	}

	// Check for standard error responses
	standardErrors := []string{"Unauthorized", "Forbidden", "NotFound", "ValidationError"}

	for _, errorType := range standardErrors {
		if _, exists := responses[errorType]; exists {
			t.Logf("✅ Standard error response %s defined", errorType)
		} else {
			t.Errorf("❌ Standard error response %s missing", errorType)
		}
	}
}

func validateTagConsistency(t *testing.T, spec map[string]interface{}) {
	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		return
	}

	tagUsage := make(map[string]int)

	for _, pathValue := range paths {
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

			if tags, exists := methodMap["tags"]; exists {
				tagsArray, ok := tags.([]interface{})
				if ok {
					for _, tag := range tagsArray {
						if tagStr, ok := tag.(string); ok {
							tagUsage[tagStr]++
						}
					}
				}
			}
		}
	}

	// Check that we have reasonable tag usage
	assert.Greater(t, len(tagUsage), 3, "Should have multiple tags for organization")

	for tag, count := range tagUsage {
		t.Logf("Tag '%s' used %d times", tag, count)
	}
}

func validateDescriptionQuality(t *testing.T, spec map[string]interface{}) {
	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		return
	}

	shortDescriptions := 0
	totalDescriptions := 0

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

			if description, exists := methodMap["description"]; exists {
				totalDescriptions++
				if descStr, ok := description.(string); ok {
					if len(descStr) < 20 {
						shortDescriptions++
						t.Logf("⚠️  Short description in %s %s: %s", strings.ToUpper(methodName), pathName, descStr)
					}
				}
			}
		}
	}

	if totalDescriptions > 0 {
		shortPercentage := float64(shortDescriptions) / float64(totalDescriptions) * 100
		assert.LessOrEqual(t, shortPercentage, 20.0, "No more than 20%% of descriptions should be too short")
	}
}

func validateSchemaValidationRules(t *testing.T, spec map[string]interface{}) {
	components, ok := spec["components"].(map[string]interface{})
	if !ok {
		return
	}

	schemas, ok := components["schemas"].(map[string]interface{})
	if !ok {
		return
	}

	schemasWithValidation := 0
	totalSchemas := len(schemas)

	for _, schemaValue := range schemas {
		schemaMap, ok := schemaValue.(map[string]interface{})
		if !ok {
			continue
		}

		hasValidation := false

		// Check for validation rules
		validationFields := []string{"required", "minLength", "maxLength", "minimum", "maximum", "pattern", "enum"}
		for _, field := range validationFields {
			if _, exists := schemaMap[field]; exists {
				hasValidation = true
				break
			}
		}

		// Check properties for validation
		if properties, exists := schemaMap["properties"]; exists {
			propertiesMap, ok := properties.(map[string]interface{})
			if ok {
				for _, propValue := range propertiesMap {
					propMap, ok := propValue.(map[string]interface{})
					if ok {
						for _, field := range validationFields {
							if _, exists := propMap[field]; exists {
								hasValidation = true
								break
							}
						}
					}
				}
			}
		}

		if hasValidation {
			schemasWithValidation++
		}
	}

	if totalSchemas > 0 {
		validationPercentage := float64(schemasWithValidation) / float64(totalSchemas) * 100
		t.Logf("Schemas with validation rules: %.1f%%", validationPercentage)
		assert.GreaterOrEqual(t, validationPercentage, 60.0, "At least 60%% of schemas should have validation rules")
	}
}

func validateSecurityDocumentationQuality(t *testing.T, spec map[string]interface{}) {
	components, ok := spec["components"].(map[string]interface{})
	if !ok {
		return
	}

	securitySchemes, ok := components["securitySchemes"].(map[string]interface{})
	if !ok {
		t.Error("Security schemes should be documented")
		return
	}

	assert.NotEmpty(t, securitySchemes, "Should have security schemes defined")

	// Check BearerAuth specifically
	if bearerAuth, exists := securitySchemes["BearerAuth"]; exists {
		bearerAuthMap, ok := bearerAuth.(map[string]interface{})
		require.True(t, ok, "BearerAuth should be an object")

		requiredFields := []string{"type", "scheme", "bearerFormat", "description"}
		for _, field := range requiredFields {
			if _, exists := bearerAuthMap[field]; exists {
				t.Logf("✅ BearerAuth has %s field", field)
			} else {
				t.Errorf("❌ BearerAuth missing %s field", field)
			}
		}
	}
}

func validateRequiredDocumentationFilesExist(t *testing.T, docsDir string) {
	requiredFiles := []string{
		"README.md",
		"index.html",
		"api-documentation.html",
		"api-documentation.md",
		"api-types.ts",
		"developer-guide.md",
		"swagger-ui.html",
	}

	for _, file := range requiredFiles {
		filePath := filepath.Join(docsDir, file)
		if _, err := os.Stat(filePath); err == nil {
			t.Logf("✅ Required file exists: %s", file)
		} else {
			t.Errorf("❌ Required file missing: %s", file)
		}
	}
}

func validateGeneratedFileContentQuality(t *testing.T, docsDir string) {
	// Check that generated files have reasonable content
	files := []string{"api-documentation.md", "developer-guide.md"}

	for _, file := range files {
		filePath := filepath.Join(docsDir, file)
		if content, err := os.ReadFile(filePath); err == nil {
			contentStr := string(content)

			// Basic quality checks
			assert.Greater(t, len(contentStr), 1000, "File %s should have substantial content", file)
			assert.Contains(t, contentStr, "API", "File %s should contain API documentation", file)

			// Check for common documentation elements
			if strings.Contains(file, "developer-guide") {
				assert.Contains(t, contentStr, "authentication", "Developer guide should mention authentication")
				assert.Contains(t, contentStr, "example", "Developer guide should have examples")
			}
		}
	}
}

func validateCrossReferenceConsistency(t *testing.T, docsDir string) {
	// This would check that cross-references between files are valid
	// For now, just check that TypeScript file has reasonable content

	tsFilePath := filepath.Join(docsDir, "api-types.ts")
	if content, err := os.ReadFile(tsFilePath); err == nil {
		contentStr := string(content)

		// Check for key interfaces
		keyInterfaces := []string{"Epic", "UserStory", "AcceptanceCriteria", "Requirement", "ListResponse", "ErrorResponse"}

		for _, interfaceName := range keyInterfaces {
			if strings.Contains(contentStr, fmt.Sprintf("interface %s", interfaceName)) {
				t.Logf("✅ TypeScript interface %s defined", interfaceName)
			} else {
				t.Errorf("❌ TypeScript interface %s missing", interfaceName)
			}
		}
	}
}

func reportCompletenessMetrics(t *testing.T, metrics CompletenessMetrics) {
	t.Logf("\n=== Documentation Completeness Report ===")
	t.Logf("Endpoints: %d total, %d documented", metrics.TotalEndpoints, metrics.DocumentedEndpoints)
	t.Logf("  - With summaries: %d (%.1f%%)", metrics.EndpointsWithSummary,
		float64(metrics.EndpointsWithSummary)/float64(metrics.TotalEndpoints)*100)
	t.Logf("  - With descriptions: %d (%.1f%%)", metrics.EndpointsWithDescription,
		float64(metrics.EndpointsWithDescription)/float64(metrics.TotalEndpoints)*100)
	t.Logf("  - With examples: %d (%.1f%%)", metrics.EndpointsWithExamples,
		float64(metrics.EndpointsWithExamples)/float64(metrics.TotalEndpoints)*100)

	t.Logf("Schemas: %d total, %d with descriptions (%.1f%%)", metrics.TotalSchemas, metrics.SchemasWithDescription,
		float64(metrics.SchemasWithDescription)/float64(metrics.TotalSchemas)*100)

	t.Logf("Parameters: %d total, %d with descriptions (%.1f%%)", metrics.TotalParameters, metrics.ParametersWithDescription,
		float64(metrics.ParametersWithDescription)/float64(metrics.TotalParameters)*100)

	if len(metrics.MissingDocumentation) > 0 {
		t.Logf("\nMissing Documentation (%d items):", len(metrics.MissingDocumentation))
		sort.Strings(metrics.MissingDocumentation)
		for i, missing := range metrics.MissingDocumentation {
			if i < 10 { // Limit output
				t.Logf("  - %s", missing)
			} else if i == 10 {
				t.Logf("  ... and %d more", len(metrics.MissingDocumentation)-10)
				break
			}
		}
	}

	if len(metrics.QualityIssues) > 0 {
		t.Logf("\nQuality Issues (%d items):", len(metrics.QualityIssues))
		for i, issue := range metrics.QualityIssues {
			if i < 5 { // Limit output
				t.Logf("  - %s", issue)
			} else if i == 5 {
				t.Logf("  ... and %d more", len(metrics.QualityIssues)-5)
				break
			}
		}
	}
}
