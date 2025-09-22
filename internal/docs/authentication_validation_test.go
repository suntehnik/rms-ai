package docs

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// AuthenticationRequirement represents authentication info for an endpoint
type AuthenticationRequirement struct {
	Path         string
	Method       string
	RequiresAuth bool
	RequiresRole string
	IsPublic     bool
}

// TestAuthenticationDocumentationAccuracy validates authentication requirements
func TestAuthenticationDocumentationAccuracy(t *testing.T) {
	// Extract authentication requirements from routes.go
	routeAuthReqs, err := extractAuthenticationRequirementsFromRoutes("../../internal/server/routes/routes.go")
	require.NoError(t, err, "Should be able to extract auth requirements from routes")

	// Extract authentication documentation from OpenAPI spec
	specAuthReqs, err := extractAuthenticationFromOpenAPI("../../docs/openapi-v3.yaml")
	require.NoError(t, err, "Should be able to extract auth requirements from OpenAPI spec")

	t.Run("PublicEndpointsConsistent", func(t *testing.T) {
		validatePublicEndpointsConsistent(t, routeAuthReqs, specAuthReqs)
	})

	t.Run("AuthenticatedEndpointsConsistent", func(t *testing.T) {
		validateAuthenticatedEndpointsConsistent(t, routeAuthReqs, specAuthReqs)
	})

	t.Run("AdminEndpointsConsistent", func(t *testing.T) {
		validateAdminEndpointsConsistent(t, routeAuthReqs, specAuthReqs)
	})

	t.Run("SecuritySchemesComplete", func(t *testing.T) {
		validateSecuritySchemesComplete(t, specAuthReqs)
	})
}

// TestMiddlewareAuthenticationImplementation validates middleware implementation
func TestMiddlewareAuthenticationImplementation(t *testing.T) {
	// Check auth middleware implementation
	authMiddleware, err := analyzeAuthMiddleware("../../internal/auth/middleware.go")
	require.NoError(t, err, "Should be able to analyze auth middleware")

	t.Run("JWTValidationImplemented", func(t *testing.T) {
		validateJWTValidationImplemented(t, authMiddleware)
	})

	t.Run("RoleBasedAccessImplemented", func(t *testing.T) {
		validateRoleBasedAccessImplemented(t, authMiddleware)
	})

	t.Run("ErrorHandlingImplemented", func(t *testing.T) {
		validateAuthErrorHandlingImplemented(t, authMiddleware)
	})
}

// TestAuthenticationFlowDocumentation validates authentication flow documentation
func TestAuthenticationFlowDocumentation(t *testing.T) {
	specPath := "../../docs/openapi-v3.yaml"
	specData, err := os.ReadFile(specPath)
	require.NoError(t, err, "Should be able to read OpenAPI spec")

	var spec map[string]interface{}
	err = yaml.Unmarshal(specData, &spec)
	require.NoError(t, err, "OpenAPI spec should be valid YAML")

	t.Run("LoginEndpointDocumented", func(t *testing.T) {
		validateLoginEndpointDocumented(t, spec)
	})

	t.Run("TokenFormatDocumented", func(t *testing.T) {
		validateTokenFormatDocumented(t, spec)
	})

	t.Run("AuthenticationErrorsDocumented", func(t *testing.T) {
		validateAuthenticationErrorsDocumented(t, spec)
	})

	t.Run("SecurityRequirementsDocumented", func(t *testing.T) {
		validateSecurityRequirementsDocumented(t, spec)
	})
}

// Helper functions for extracting authentication requirements

func extractAuthenticationRequirementsFromRoutes(filename string) ([]AuthenticationRequirement, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var requirements []AuthenticationRequirement
	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Track current group and middleware
	currentGroup := ""
	var groupMiddleware []string
	inAuthGroup := false

	// Patterns for route definitions and middleware
	groupPattern := regexp.MustCompile(`^\s*(\w+)\s*:=\s*router\.Group\("([^"]+)"\)`)
	middlewarePattern := regexp.MustCompile(`^\s*(\w+)\.Use\(([^)]+)\)`)
	routePattern := regexp.MustCompile(`^\s*(\w+)\.(\w+)\("([^"]+)",\s*([^,)]+)(?:,\s*([^)]+))?\)`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip comments and empty lines
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "//") {
			continue
		}

		// Check for group definitions
		if matches := groupPattern.FindStringSubmatch(line); len(matches) > 2 {
			currentGroup = matches[2]
			groupMiddleware = []string{}
			inAuthGroup = strings.Contains(currentGroup, "auth")
		}

		// Check for middleware usage
		if matches := middlewarePattern.FindStringSubmatch(line); len(matches) > 2 {
			middleware := matches[2]
			groupMiddleware = append(groupMiddleware, middleware)
		}

		// Check for route definitions
		if matches := routePattern.FindStringSubmatch(line); len(matches) >= 4 {
			method := strings.ToUpper(matches[2])
			path := matches[3]
			handler := matches[4]

			// Determine full path
			fullPath := path
			if currentGroup != "" && !strings.HasPrefix(path, "/") {
				fullPath = currentGroup + "/" + path
			} else if currentGroup != "" {
				fullPath = currentGroup + path
			}

			// Normalize path
			fullPath = normalizeRoutePath(fullPath)

			// Determine authentication requirements
			req := AuthenticationRequirement{
				Path:   fullPath,
				Method: method,
			}

			// Check if it's a public endpoint
			req.IsPublic = isPublicEndpoint(fullPath)

			// Check for authentication middleware
			if !req.IsPublic {
				req.RequiresAuth = hasAuthMiddleware(groupMiddleware, handler, line)
				req.RequiresRole = extractRequiredRole(groupMiddleware, handler, line)
			}

			requirements = append(requirements, req)
		}
	}

	return requirements, scanner.Err()
}

func extractAuthenticationFromOpenAPI(filename string) (map[string]AuthenticationRequirement, error) {
	specData, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var spec map[string]interface{}
	err = yaml.Unmarshal(specData, &spec)
	if err != nil {
		return nil, err
	}

	requirements := make(map[string]AuthenticationRequirement)

	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		return requirements, nil
	}

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

			key := fmt.Sprintf("%s %s", strings.ToUpper(methodName), pathName)
			req := AuthenticationRequirement{
				Path:   pathName,
				Method: strings.ToUpper(methodName),
			}

			// Check security requirements
			if security, exists := methodMap["security"]; exists {
				securityArray, ok := security.([]interface{})
				if ok {
					req.IsPublic = len(securityArray) == 0
					req.RequiresAuth = len(securityArray) > 0
				}
			} else {
				// If no security specified, assume it requires auth (unless it's a known public endpoint)
				req.IsPublic = isPublicEndpoint(pathName)
				req.RequiresAuth = !req.IsPublic
			}

			// Check for role requirements
			if role, exists := methodMap["x-required-role"]; exists {
				if roleStr, ok := role.(string); ok {
					req.RequiresRole = roleStr
				}
			}

			requirements[key] = req
		}
	}

	return requirements, nil
}

func analyzeAuthMiddleware(filename string) (map[string]interface{}, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	analysis := make(map[string]interface{})

	// Look for middleware functions
	for _, decl := range node.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			if fn.Name.Name == "Middleware" || strings.Contains(fn.Name.Name, "Auth") {
				analysis[fn.Name.Name] = analyzeFunction(fn)
			}
		}
	}

	return analysis, nil
}

func analyzeFunction(fn *ast.FuncDecl) map[string]bool {
	features := map[string]bool{
		"jwt_validation":    false,
		"role_checking":     false,
		"error_handling":    false,
		"token_extraction":  false,
		"claims_validation": false,
	}

	// Walk the function body to find authentication-related code
	ast.Inspect(fn, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			if ident, ok := node.Fun.(*ast.Ident); ok {
				switch ident.Name {
				case "ParseWithClaims", "Parse":
					features["jwt_validation"] = true
				case "Valid":
					features["claims_validation"] = true
				}
			}
			if sel, ok := node.Fun.(*ast.SelectorExpr); ok {
				switch sel.Sel.Name {
				case "ParseWithClaims", "Parse":
					features["jwt_validation"] = true
				case "Valid":
					features["claims_validation"] = true
				}
			}
		case *ast.IfStmt:
			// Look for error handling patterns
			if containsErrorHandling(node) {
				features["error_handling"] = true
			}
		}
		return true
	})

	// Check function body for specific patterns
	if fn.Body != nil {
		bodyStr := fmt.Sprintf("%+v", fn.Body)
		if strings.Contains(bodyStr, "Authorization") || strings.Contains(bodyStr, "Bearer") {
			features["token_extraction"] = true
		}
		if strings.Contains(bodyStr, "Role") || strings.Contains(bodyStr, "role") {
			features["role_checking"] = true
		}
	}

	return features
}

func containsErrorHandling(ifStmt *ast.IfStmt) bool {
	if ifStmt.Cond != nil {
		condStr := fmt.Sprintf("%+v", ifStmt.Cond)
		return strings.Contains(condStr, "err") || strings.Contains(condStr, "error")
	}
	return false
}

// Validation functions

func validatePublicEndpointsConsistent(t *testing.T, routeReqs []AuthenticationRequirement, specReqs map[string]AuthenticationRequirement) {
	publicEndpoints := []string{"/auth/login", "/ready", "/live"}

	for _, endpoint := range publicEndpoints {
		// Find in route requirements
		routeReq := findRouteRequirement(routeReqs, endpoint)
		if routeReq != nil {
			assert.True(t, routeReq.IsPublic, "Route %s should be marked as public in implementation", endpoint)
		}

		// Find in spec requirements
		for _, method := range []string{"GET", "POST"} {
			key := fmt.Sprintf("%s %s", method, endpoint)
			if specReq, exists := specReqs[key]; exists {
				assert.True(t, specReq.IsPublic, "Endpoint %s should be marked as public in OpenAPI spec", key)
			}
		}
	}
}

func validateAuthenticatedEndpointsConsistent(t *testing.T, routeReqs []AuthenticationRequirement, specReqs map[string]AuthenticationRequirement) {
	// Check that non-public endpoints require authentication
	for _, routeReq := range routeReqs {
		if !routeReq.IsPublic {
			key := fmt.Sprintf("%s %s", routeReq.Method, routeReq.Path)
			if specReq, exists := specReqs[key]; exists {
				assert.True(t, specReq.RequiresAuth, "Authenticated endpoint %s should require auth in OpenAPI spec", key)
				assert.False(t, specReq.IsPublic, "Authenticated endpoint %s should not be public in OpenAPI spec", key)
			} else {
				t.Logf("⚠️  Endpoint %s found in routes but not in OpenAPI spec", key)
			}
		}
	}
}

func validateAdminEndpointsConsistent(t *testing.T, routeReqs []AuthenticationRequirement, specReqs map[string]AuthenticationRequirement) {
	adminPaths := []string{"/auth/users", "/api/v1/config/"}

	for _, routeReq := range routeReqs {
		isAdminPath := false
		for _, adminPath := range adminPaths {
			if strings.HasPrefix(routeReq.Path, adminPath) {
				isAdminPath = true
				break
			}
		}

		if isAdminPath {
			key := fmt.Sprintf("%s %s", routeReq.Method, routeReq.Path)
			if specReq, exists := specReqs[key]; exists {
				if routeReq.RequiresRole == "Administrator" {
					assert.Equal(t, "Administrator", specReq.RequiresRole,
						"Admin endpoint %s should require Administrator role in OpenAPI spec", key)
				}
			}
		}
	}
}

func validateSecuritySchemesComplete(t *testing.T, specReqs map[string]AuthenticationRequirement) {
	// This would be validated by reading the OpenAPI spec directly
	// For now, we'll check that we have authentication requirements documented
	hasAuthenticatedEndpoints := false
	for _, req := range specReqs {
		if req.RequiresAuth {
			hasAuthenticatedEndpoints = true
			break
		}
	}

	assert.True(t, hasAuthenticatedEndpoints, "Should have authenticated endpoints documented")
}

func validateJWTValidationImplemented(t *testing.T, authMiddleware map[string]interface{}) {
	hasJWTValidation := false
	for funcName, analysis := range authMiddleware {
		if analysisMap, ok := analysis.(map[string]bool); ok {
			if analysisMap["jwt_validation"] {
				hasJWTValidation = true
				t.Logf("✅ JWT validation implemented in %s", funcName)
			}
		}
	}

	assert.True(t, hasJWTValidation, "JWT validation should be implemented in auth middleware")
}

func validateRoleBasedAccessImplemented(t *testing.T, authMiddleware map[string]interface{}) {
	hasRoleChecking := false
	for funcName, analysis := range authMiddleware {
		if analysisMap, ok := analysis.(map[string]bool); ok {
			if analysisMap["role_checking"] {
				hasRoleChecking = true
				t.Logf("✅ Role-based access implemented in %s", funcName)
			}
		}
	}

	assert.True(t, hasRoleChecking, "Role-based access control should be implemented")
}

func validateAuthErrorHandlingImplemented(t *testing.T, authMiddleware map[string]interface{}) {
	hasErrorHandling := false
	for funcName, analysis := range authMiddleware {
		if analysisMap, ok := analysis.(map[string]bool); ok {
			if analysisMap["error_handling"] {
				hasErrorHandling = true
				t.Logf("✅ Error handling implemented in %s", funcName)
			}
		}
	}

	assert.True(t, hasErrorHandling, "Authentication error handling should be implemented")
}

func validateLoginEndpointDocumented(t *testing.T, spec map[string]interface{}) {
	paths, ok := spec["paths"].(map[string]interface{})
	require.True(t, ok, "Should have paths section")

	loginPath, exists := paths["/auth/login"]
	assert.True(t, exists, "Login endpoint should be documented")

	if exists {
		loginPathMap, ok := loginPath.(map[string]interface{})
		require.True(t, ok, "Login path should be an object")

		postMethod, exists := loginPathMap["post"]
		assert.True(t, exists, "Login should have POST method")

		if exists {
			postMethodMap, ok := postMethod.(map[string]interface{})
			require.True(t, ok, "POST method should be an object")

			// Should have request body
			assert.Contains(t, postMethodMap, "requestBody", "Login should have request body documented")

			// Should have responses
			assert.Contains(t, postMethodMap, "responses", "Login should have responses documented")

			// Should be marked as public (no security requirements)
			if security, exists := postMethodMap["security"]; exists {
				securityArray, ok := security.([]interface{})
				if ok {
					assert.Empty(t, securityArray, "Login endpoint should be public (empty security array)")
				}
			}
		}
	}
}

func validateTokenFormatDocumented(t *testing.T, spec map[string]interface{}) {
	components, ok := spec["components"].(map[string]interface{})
	require.True(t, ok, "Should have components section")

	securitySchemes, ok := components["securitySchemes"].(map[string]interface{})
	require.True(t, ok, "Should have security schemes")

	bearerAuth, exists := securitySchemes["BearerAuth"]
	assert.True(t, exists, "BearerAuth security scheme should be documented")

	if exists {
		bearerAuthMap, ok := bearerAuth.(map[string]interface{})
		require.True(t, ok, "BearerAuth should be an object")

		assert.Equal(t, "http", bearerAuthMap["type"], "Should be HTTP authentication")
		assert.Equal(t, "bearer", bearerAuthMap["scheme"], "Should use bearer scheme")
		assert.Equal(t, "JWT", bearerAuthMap["bearerFormat"], "Should specify JWT format")
	}
}

func validateAuthenticationErrorsDocumented(t *testing.T, spec map[string]interface{}) {
	components, ok := spec["components"].(map[string]interface{})
	require.True(t, ok, "Should have components section")

	responses, ok := components["responses"].(map[string]interface{})
	if !ok {
		t.Skip("No standard responses defined")
	}

	// Check for authentication error responses
	authErrors := []string{"Unauthorized", "Forbidden"}
	for _, errorType := range authErrors {
		if errorResponse, exists := responses[errorType]; exists {
			errorResponseMap, ok := errorResponse.(map[string]interface{})
			require.True(t, ok, "%s response should be an object", errorType)

			assert.Contains(t, errorResponseMap, "description", "%s response should have description", errorType)
		}
	}
}

func validateSecurityRequirementsDocumented(t *testing.T, spec map[string]interface{}) {
	paths, ok := spec["paths"].(map[string]interface{})
	require.True(t, ok, "Should have paths section")

	hasSecurityRequirements := false
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
						hasSecurityRequirements = true
					}
				}
			}
		}
	}

	assert.True(t, hasSecurityRequirements, "Should have security requirements documented for protected endpoints")
}

// Helper functions

func isPublicEndpoint(path string) bool {
	publicPaths := []string{"/auth/login", "/ready", "/live", "/docs", "/swagger"}
	for _, publicPath := range publicPaths {
		if path == publicPath || strings.HasPrefix(path, publicPath) {
			return true
		}
	}
	return false
}

func hasAuthMiddleware(middleware []string, handler, line string) bool {
	// Check if line contains auth middleware calls
	authPatterns := []string{
		"authService.Middleware()",
		"authService.RequireAdministrator()",
		"auth.Middleware",
		"RequireAuth",
	}

	for _, pattern := range authPatterns {
		if strings.Contains(line, pattern) {
			return true
		}
	}

	// Check middleware slice
	for _, mw := range middleware {
		if strings.Contains(mw, "auth") || strings.Contains(mw, "Auth") {
			return true
		}
	}

	return false
}

func extractRequiredRole(middleware []string, handler, line string) string {
	if strings.Contains(line, "RequireAdministrator") {
		return "Administrator"
	}

	for _, mw := range middleware {
		if strings.Contains(mw, "Administrator") {
			return "Administrator"
		}
	}

	return ""
}

func findRouteRequirement(requirements []AuthenticationRequirement, path string) *AuthenticationRequirement {
	for _, req := range requirements {
		if req.Path == path {
			return &req
		}
	}
	return nil
}
