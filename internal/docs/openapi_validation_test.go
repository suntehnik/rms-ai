package docs

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpenAPISchemaCompliance validates the generated OpenAPI spec against the official schema
func TestOpenAPISchemaCompliance(t *testing.T) {
	specPath := "../../docs/swagger.json"

	// Check if swagger.json exists
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Skip("swagger.json not found. Run 'make swagger' to generate it.")
	}

	// Read the OpenAPI specification
	specData, err := os.ReadFile(specPath)
	require.NoError(t, err, "Should be able to read swagger.json")

	// Parse the specification
	var spec map[string]interface{}
	err = json.Unmarshal(specData, &spec)
	require.NoError(t, err, "swagger.json should be valid JSON")

	// Use a simplified OpenAPI 2.0 schema validation
	t.Run("BasicStructureValidation", func(t *testing.T) {
		validateBasicOpenAPIStructure(t, spec)
	})

	t.Run("PathsValidation", func(t *testing.T) {
		validatePathsStructure(t, spec)
	})

	t.Run("DefinitionsValidation", func(t *testing.T) {
		validateDefinitionsStructure(t, spec)
	})

	t.Run("SecurityValidation", func(t *testing.T) {
		validateSecurityStructure(t, spec)
	})
}

// validateBasicOpenAPIStructure validates the basic structure of OpenAPI spec
func validateBasicOpenAPIStructure(t *testing.T, spec map[string]interface{}) {
	// Required fields for OpenAPI 2.0
	requiredFields := []string{"swagger", "info", "paths"}

	for _, field := range requiredFields {
		assert.Contains(t, spec, field, "OpenAPI spec should contain %s", field)
	}

	// Validate swagger version
	if swagger, exists := spec["swagger"]; exists {
		assert.Equal(t, "2.0", swagger, "Should use Swagger 2.0")
	}

	// Validate info object
	if info, exists := spec["info"]; exists {
		infoObj, ok := info.(map[string]interface{})
		require.True(t, ok, "Info should be an object")

		infoRequiredFields := []string{"title", "version"}
		for _, field := range infoRequiredFields {
			assert.Contains(t, infoObj, field, "Info should contain %s", field)
		}

		// Validate title and version are strings
		if title, exists := infoObj["title"]; exists {
			assert.IsType(t, "", title, "Title should be a string")
			assert.NotEmpty(t, title, "Title should not be empty")
		}

		if version, exists := infoObj["version"]; exists {
			assert.IsType(t, "", version, "Version should be a string")
			assert.NotEmpty(t, version, "Version should not be empty")
		}
	}
}

// validatePathsStructure validates the paths section of OpenAPI spec
func validatePathsStructure(t *testing.T, spec map[string]interface{}) {
	paths, exists := spec["paths"]
	if !exists {
		t.Skip("No paths defined in specification")
	}

	pathsObj, ok := paths.(map[string]interface{})
	require.True(t, ok, "Paths should be an object")

	assert.NotEmpty(t, pathsObj, "Should have at least one path defined")

	httpMethods := []string{"get", "post", "put", "patch", "delete", "head", "options"}

	for pathName, pathValue := range pathsObj {
		assert.True(t, strings.HasPrefix(pathName, "/"), "Path %s should start with /", pathName)

		pathObj, ok := pathValue.(map[string]interface{})
		require.True(t, ok, "Path %s should be an object", pathName)

		// Check each HTTP method in the path
		for methodName, methodValue := range pathObj {
			if contains(httpMethods, strings.ToLower(methodName)) {
				validateOperationObject(t, pathName, methodName, methodValue)
			}
		}
	}
}

// validateOperationObject validates an individual operation (HTTP method) in a path
func validateOperationObject(t *testing.T, pathName, methodName string, operation interface{}) {
	operationObj, ok := operation.(map[string]interface{})
	require.True(t, ok, "Operation %s %s should be an object", methodName, pathName)

	// Responses are required
	assert.Contains(t, operationObj, "responses",
		"Operation %s %s should have responses", methodName, pathName)

	// Validate responses
	if responses, exists := operationObj["responses"]; exists {
		responsesObj, ok := responses.(map[string]interface{})
		require.True(t, ok, "Responses should be an object")
		assert.NotEmpty(t, responsesObj, "Should have at least one response defined")

		// Check for at least one success response (2xx)
		hasSuccessResponse := false
		for statusCode := range responsesObj {
			if len(statusCode) == 3 && statusCode[0] == '2' {
				hasSuccessResponse = true
				break
			}
		}
		assert.True(t, hasSuccessResponse,
			"Operation %s %s should have at least one success response", methodName, pathName)
	}

	// Validate tags if present
	if tags, exists := operationObj["tags"]; exists {
		tagsArray, ok := tags.([]interface{})
		require.True(t, ok, "Tags should be an array")
		assert.NotEmpty(t, tagsArray, "Tags array should not be empty if present")

		for i, tag := range tagsArray {
			assert.IsType(t, "", tag, "Tag %d should be a string", i)
		}
	}

	// Validate parameters if present
	if parameters, exists := operationObj["parameters"]; exists {
		parametersArray, ok := parameters.([]interface{})
		require.True(t, ok, "Parameters should be an array")

		for i, param := range parametersArray {
			paramObj, ok := param.(map[string]interface{})
			require.True(t, ok, "Parameter %d should be an object", i)

			// Required fields for parameters
			assert.Contains(t, paramObj, "name", "Parameter %d should have name", i)
			assert.Contains(t, paramObj, "in", "Parameter %d should have 'in' field", i)
		}
	}
}

// validateDefinitionsStructure validates the definitions section
func validateDefinitionsStructure(t *testing.T, spec map[string]interface{}) {
	definitions, exists := spec["definitions"]
	if !exists {
		return // Definitions are optional
	}

	definitionsObj, ok := definitions.(map[string]interface{})
	require.True(t, ok, "Definitions should be an object")

	for defName, defValue := range definitionsObj {
		defObj, ok := defValue.(map[string]interface{})
		require.True(t, ok, "Definition %s should be an object", defName)

		// Should have type or properties or allOf
		hasValidStructure := false
		validFields := []string{"type", "properties", "allOf", "anyOf", "oneOf"}

		for _, field := range validFields {
			if _, exists := defObj[field]; exists {
				hasValidStructure = true
				break
			}
		}

		assert.True(t, hasValidStructure,
			"Definition %s should have type, properties, or composition fields", defName)

		// Validate properties if present
		if properties, exists := defObj["properties"]; exists {
			propertiesObj, ok := properties.(map[string]interface{})
			require.True(t, ok, "Properties should be an object")

			for propName, propValue := range propertiesObj {
				propObj, ok := propValue.(map[string]interface{})
				require.True(t, ok, "Property %s should be an object", propName)

				// Properties should have type or $ref
				hasType := false
				if _, exists := propObj["type"]; exists {
					hasType = true
				}
				if _, exists := propObj["$ref"]; exists {
					hasType = true
				}
				if _, exists := propObj["allOf"]; exists {
					hasType = true
				}

				assert.True(t, hasType,
					"Property %s in definition %s should have type or $ref", propName, defName)
			}
		}
	}
}

// validateSecurityStructure validates security definitions
func validateSecurityStructure(t *testing.T, spec map[string]interface{}) {
	securityDefs, exists := spec["securityDefinitions"]
	if !exists {
		return // Security definitions are optional
	}

	securityObj, ok := securityDefs.(map[string]interface{})
	require.True(t, ok, "Security definitions should be an object")

	for secName, secValue := range securityObj {
		secDefObj, ok := secValue.(map[string]interface{})
		require.True(t, ok, "Security definition %s should be an object", secName)

		// Required field: type
		assert.Contains(t, secDefObj, "type", "Security definition %s should have type", secName)

		if secType, exists := secDefObj["type"]; exists {
			typeStr, ok := secType.(string)
			require.True(t, ok, "Security type should be a string")

			validTypes := []string{"basic", "apiKey", "oauth2"}
			assert.Contains(t, validTypes, typeStr,
				"Security type should be one of: %v", validTypes)

			// Validate apiKey specific fields
			if typeStr == "apiKey" {
				assert.Contains(t, secDefObj, "name", "apiKey security should have name")
				assert.Contains(t, secDefObj, "in", "apiKey security should have 'in' field")

				if in, exists := secDefObj["in"]; exists {
					inStr, ok := in.(string)
					require.True(t, ok, "apiKey 'in' should be a string")
					validIn := []string{"query", "header", "formData"}
					assert.Contains(t, validIn, inStr,
						"apiKey 'in' should be one of: %v", validIn)
				}
			}
		}
	}
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// TestSwaggerSpecificationCompleteness checks for completeness of the API documentation
func TestSwaggerSpecificationCompleteness(t *testing.T) {
	specPath := "../../docs/swagger.json"

	// Check if swagger.json exists
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Skip("swagger.json not found. Run 'make swagger' to generate it.")
	}

	// Read and parse specification
	specData, err := os.ReadFile(specPath)
	require.NoError(t, err, "Should be able to read swagger.json")

	var spec map[string]interface{}
	err = json.Unmarshal(specData, &spec)
	require.NoError(t, err, "swagger.json should be valid JSON")

	t.Run("CoreEntityCoverage", func(t *testing.T) {
		paths, exists := spec["paths"]
		if !exists {
			t.Skip("No paths defined")
		}

		pathsObj := paths.(map[string]interface{})

		// Check for core entity endpoints
		coreEntities := []string{"epics", "user-stories", "acceptance-criteria", "requirements"}

		for _, entity := range coreEntities {
			hasEntityEndpoint := false

			for pathName := range pathsObj {
				if strings.Contains(pathName, entity) {
					hasEntityEndpoint = true
					break
				}
			}

			assert.True(t, hasEntityEndpoint,
				"Should have endpoints for %s entity", entity)
		}
	})

	t.Run("CRUDOperationCoverage", func(t *testing.T) {
		paths, exists := spec["paths"]
		if !exists {
			t.Skip("No paths defined")
		}

		pathsObj := paths.(map[string]interface{})

		// Check for CRUD operations
		hasCRUD := map[string]bool{
			"create": false, // POST
			"read":   false, // GET
			"update": false, // PUT/PATCH
			"delete": false, // DELETE
		}

		for _, pathValue := range pathsObj {
			pathObj := pathValue.(map[string]interface{})

			if _, exists := pathObj["post"]; exists {
				hasCRUD["create"] = true
			}
			if _, exists := pathObj["get"]; exists {
				hasCRUD["read"] = true
			}
			if _, exists := pathObj["put"]; exists || pathObj["patch"] != nil {
				hasCRUD["update"] = true
			}
			if _, exists := pathObj["delete"]; exists {
				hasCRUD["delete"] = true
			}
		}

		for operation, found := range hasCRUD {
			assert.True(t, found, "Should have %s operations documented", operation)
		}
	})

	t.Run("SearchEndpointCoverage", func(t *testing.T) {
		paths, exists := spec["paths"]
		if !exists {
			t.Skip("No paths defined")
		}

		pathsObj := paths.(map[string]interface{})

		hasSearchEndpoint := false
		for pathName := range pathsObj {
			if strings.Contains(pathName, "search") {
				hasSearchEndpoint = true
				break
			}
		}

		assert.True(t, hasSearchEndpoint, "Should have search endpoint documented")
	})
}
