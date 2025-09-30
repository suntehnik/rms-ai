package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
)

// RouteInfo represents a route from the routes.go file
type RouteInfo struct {
	Method string
	Path   string
	Line   int
}

// OpenAPIPath represents a path from the OpenAPI specification
type OpenAPIPath struct {
	Path    string
	Methods []string
	Line    int
}

func main() {
	fmt.Println("=== OpenAPI Specification Completeness Validation ===\n")

	// Extract routes from routes.go
	routes, err := extractRoutesFromFile("internal/server/routes/routes.go")
	if err != nil {
		log.Fatalf("Error extracting routes: %v", err)
	}

	// Extract paths from OpenAPI specification
	openAPIPaths, err := extractOpenAPIPaths("docs/openapi-v3.yaml")
	if err != nil {
		log.Fatalf("Error extracting OpenAPI paths: %v", err)
	}

	// Validate completeness
	validateCompleteness(routes, openAPIPaths)
}

func extractRoutesFromFile(filename string) ([]RouteInfo, error) {
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
						Path:   path,
						Line:   lineNum,
					})
				}
				break
			}
		}
	}

	return routes, scanner.Err()
}

func extractOpenAPIPaths(filename string) (map[string]OpenAPIPath, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	paths := make(map[string]OpenAPIPath)
	scanner := bufio.NewScanner(file)
	lineNum := 0
	inPaths := false
	currentPath := ""

	pathPattern := regexp.MustCompile(`^\s*(/[^:]*):$`)
	methodPattern := regexp.MustCompile(`^\s*(get|post|put|patch|delete|head|options):$`)

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
				paths[currentPath] = OpenAPIPath{
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

func validateCompleteness(routes []RouteInfo, openAPIPaths map[string]OpenAPIPath) {
	// Normalize routes for comparison
	normalizedRoutes := make(map[string][]string) // path -> methods
	for _, route := range routes {
		normalizedPath := normalizeRoutePath(route.Path)
		if normalizedPath == "" {
			continue // Skip health checks and other non-API routes
		}

		if _, exists := normalizedRoutes[normalizedPath]; !exists {
			normalizedRoutes[normalizedPath] = []string{}
		}
		normalizedRoutes[normalizedPath] = append(normalizedRoutes[normalizedPath], route.Method)
	}

	// Check for missing documentation
	fmt.Println("1. Routes missing from OpenAPI specification:")
	fmt.Println("=" + strings.Repeat("=", 50))
	missingCount := 0

	var sortedPaths []string
	for path := range normalizedRoutes {
		sortedPaths = append(sortedPaths, path)
	}
	sort.Strings(sortedPaths)

	for _, path := range sortedPaths {
		methods := normalizedRoutes[path]
		if openAPIPath, exists := openAPIPaths[path]; exists {
			// Check if all methods are documented
			for _, method := range methods {
				if !contains(openAPIPath.Methods, method) {
					fmt.Printf("   %s %s - Method not documented\n", method, path)
					missingCount++
				}
			}
		} else {
			fmt.Printf("   Path not documented: %s (methods: %s)\n", path, strings.Join(methods, ", "))
			missingCount++
		}
	}

	if missingCount == 0 {
		fmt.Println("   ✓ All implemented routes are documented")
	}

	// Check for documented endpoints without implementation
	fmt.Println("\n2. OpenAPI paths without implementation:")
	fmt.Println("=" + strings.Repeat("=", 50))
	extraCount := 0

	var sortedOpenAPIPaths []string
	for path := range openAPIPaths {
		sortedOpenAPIPaths = append(sortedOpenAPIPaths, path)
	}
	sort.Strings(sortedOpenAPIPaths)

	for _, path := range sortedOpenAPIPaths {
		openAPIPath := openAPIPaths[path]
		if routeMethods, exists := normalizedRoutes[path]; exists {
			// Check if all documented methods are implemented
			for _, method := range openAPIPath.Methods {
				if !contains(routeMethods, method) {
					fmt.Printf("   %s %s - Documented but not implemented\n", method, path)
					extraCount++
				}
			}
		} else {
			fmt.Printf("   Path documented but not implemented: %s (methods: %s)\n", path, strings.Join(openAPIPath.Methods, ", "))
			extraCount++
		}
	}

	if extraCount == 0 {
		fmt.Println("   ✓ All documented paths have implementations")
	}

	// Summary
	fmt.Println("\n3. Summary:")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Printf("   Total implemented routes: %d\n", len(routes))
	fmt.Printf("   Total documented paths: %d\n", len(openAPIPaths))
	fmt.Printf("   Missing documentation: %d\n", missingCount)
	fmt.Printf("   Extra documentation: %d\n", extraCount)

	if missingCount == 0 && extraCount == 0 {
		fmt.Println("\n   ✅ OpenAPI specification is complete and accurate!")
	} else {
		fmt.Println("\n   ❌ OpenAPI specification needs updates")
	}

	// Entity coverage analysis
	fmt.Println("\n4. Entity Coverage Analysis:")
	fmt.Println("=" + strings.Repeat("=", 50))
	analyzeEntityCoverage(normalizedRoutes, openAPIPaths)
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

	// Skip health check and documentation routes
	if strings.HasPrefix(normalized, "/ready") ||
		strings.HasPrefix(normalized, "/live") ||
		strings.HasPrefix(normalized, "/docs") ||
		strings.HasPrefix(normalized, "/swagger") {
		return ""
	}

	return normalized
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func analyzeEntityCoverage(routes map[string][]string, openAPIPaths map[string]OpenAPIPath) {
	entities := []string{"epics", "user-stories", "acceptance-criteria", "requirements"}

	for _, entity := range entities {
		fmt.Printf("\n   %s:\n", strings.Title(strings.ReplaceAll(entity, "-", " ")))

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

			if routeMethods, exists := routes[path]; exists {
				implemented = contains(routeMethods, method)
			}

			if openAPIPath, exists := openAPIPaths[path]; exists {
				documented = contains(openAPIPath.Methods, method)
			}

			status := "❌"
			if implemented && documented {
				status = "✅"
			} else if implemented && !documented {
				status = "📝" // Implemented but not documented
			} else if !implemented && documented {
				status = "🔧" // Documented but not implemented
			}

			fmt.Printf("     %s %s %s\n", status, method, path)
		}

		// Special operations
		specialOps := []string{
			fmt.Sprintf("/api/v1/%s/{id}/validate-deletion", entity),
			fmt.Sprintf("/api/v1/%s/{id}/delete", entity),
			fmt.Sprintf("/api/v1/%s/{id}/comments", entity),
			fmt.Sprintf("/api/v1/%s/{id}/comments/inline", entity),
		}

		for _, path := range specialOps {
			implemented := len(routes[path]) > 0
			documented := len(openAPIPaths[path].Methods) > 0

			status := "❌"
			if implemented && documented {
				status = "✅"
			} else if implemented && !documented {
				status = "📝"
			} else if !implemented && documented {
				status = "🔧"
			}

			if implemented || documented {
				fmt.Printf("     %s Special: %s\n", status, path)
			}
		}
	}

	fmt.Println("\n   Legend:")
	fmt.Println("     ✅ Implemented and documented")
	fmt.Println("     📝 Implemented but not documented")
	fmt.Println("     🔧 Documented but not implemented")
	fmt.Println("     ❌ Neither implemented nor documented")
}
