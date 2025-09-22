package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
)

func main() {
	fmt.Println("=== API Documentation Completeness Validation ===\n")

	// Define expected routes based on routes.go analysis
	implementedRoutes := getImplementedRoutes()

	// Read OpenAPI spec to check documented routes
	documentedRoutes, err := getDocumentedRoutes()
	if err != nil {
		log.Fatalf("Error reading OpenAPI spec: %v", err)
	}

	// Validate completeness
	validateRouteCompleteness(implementedRoutes, documentedRoutes)
}

func getImplementedRoutes() map[string][]string {
	routes := make(map[string][]string)

	// Health endpoints
	routes["/ready"] = []string{"GET"}
	routes["/live"] = []string{"GET"}

	// Auth endpoints
	routes["/auth/login"] = []string{"POST"}
	routes["/auth/profile"] = []string{"GET"}
	routes["/auth/change-password"] = []string{"POST"}
	routes["/auth/users"] = []string{"POST", "GET"}
	routes["/auth/users/{id}"] = []string{"GET", "PUT", "DELETE"}

	// Search endpoints
	routes["/api/v1/search"] = []string{"GET"}
	routes["/api/v1/search/suggestions"] = []string{"GET"}

	// Hierarchy endpoints
	routes["/api/v1/hierarchy"] = []string{"GET"}
	routes["/api/v1/hierarchy/epics/{id}"] = []string{"GET"}
	routes["/api/v1/hierarchy/user-stories/{id}"] = []string{"GET"}
	routes["/api/v1/hierarchy/path/{entity_type}/{id}"] = []string{"GET"}

	// Epic endpoints
	routes["/api/v1/epics"] = []string{"POST", "GET"}
	routes["/api/v1/epics/{id}"] = []string{"GET", "PUT", "DELETE"}
	routes["/api/v1/epics/{id}/user-stories"] = []string{"GET", "POST"}
	routes["/api/v1/epics/{id}/status"] = []string{"PATCH"}
	routes["/api/v1/epics/{id}/assign"] = []string{"PATCH"}
	routes["/api/v1/epics/{id}/validate-deletion"] = []string{"GET"}
	routes["/api/v1/epics/{id}/delete"] = []string{"DELETE"}
	routes["/api/v1/epics/{id}/comments"] = []string{"GET", "POST"}
	routes["/api/v1/epics/{id}/comments/inline"] = []string{"POST"}
	routes["/api/v1/epics/{id}/comments/inline/visible"] = []string{"GET"}
	routes["/api/v1/epics/{id}/comments/inline/validate"] = []string{"POST"}

	// User Story endpoints
	routes["/api/v1/user-stories"] = []string{"POST", "GET"}
	routes["/api/v1/user-stories/{id}"] = []string{"GET", "PUT", "DELETE"}
	routes["/api/v1/user-stories/{id}/acceptance-criteria"] = []string{"GET", "POST"}
	routes["/api/v1/user-stories/{id}/requirements"] = []string{"GET", "POST"}
	routes["/api/v1/user-stories/{id}/status"] = []string{"PATCH"}
	routes["/api/v1/user-stories/{id}/assign"] = []string{"PATCH"}
	routes["/api/v1/user-stories/{id}/validate-deletion"] = []string{"GET"}
	routes["/api/v1/user-stories/{id}/delete"] = []string{"DELETE"}
	routes["/api/v1/user-stories/{id}/comments"] = []string{"GET", "POST"}
	routes["/api/v1/user-stories/{id}/comments/inline"] = []string{"POST"}
	routes["/api/v1/user-stories/{id}/comments/inline/visible"] = []string{"GET"}
	routes["/api/v1/user-stories/{id}/comments/inline/validate"] = []string{"POST"}

	// Acceptance Criteria endpoints
	routes["/api/v1/acceptance-criteria"] = []string{"GET"}
	routes["/api/v1/acceptance-criteria/{id}"] = []string{"GET", "PUT", "DELETE"}
	routes["/api/v1/acceptance-criteria/{id}/validate-deletion"] = []string{"GET"}
	routes["/api/v1/acceptance-criteria/{id}/delete"] = []string{"DELETE"}
	routes["/api/v1/acceptance-criteria/{id}/comments"] = []string{"GET", "POST"}
	routes["/api/v1/acceptance-criteria/{id}/comments/inline"] = []string{"POST"}
	routes["/api/v1/acceptance-criteria/{id}/comments/inline/visible"] = []string{"GET"}
	routes["/api/v1/acceptance-criteria/{id}/comments/inline/validate"] = []string{"POST"}

	// Requirement endpoints
	routes["/api/v1/requirements"] = []string{"POST", "GET"}
	routes["/api/v1/requirements/search"] = []string{"GET"}
	routes["/api/v1/requirements/{id}"] = []string{"GET", "PUT", "DELETE"}
	routes["/api/v1/requirements/{id}/relationships"] = []string{"GET"}
	routes["/api/v1/requirements/{id}/status"] = []string{"PATCH"}
	routes["/api/v1/requirements/{id}/assign"] = []string{"PATCH"}
	routes["/api/v1/requirements/relationships"] = []string{"POST"}
	routes["/api/v1/requirements/{id}/validate-deletion"] = []string{"GET"}
	routes["/api/v1/requirements/{id}/delete"] = []string{"DELETE"}
	routes["/api/v1/requirements/{id}/comments"] = []string{"GET", "POST"}
	routes["/api/v1/requirements/{id}/comments/inline"] = []string{"POST"}
	routes["/api/v1/requirements/{id}/comments/inline/visible"] = []string{"GET"}
	routes["/api/v1/requirements/{id}/comments/inline/validate"] = []string{"POST"}

	// Requirement Relationship endpoints
	routes["/api/v1/requirement-relationships/{id}"] = []string{"DELETE"}

	// Configuration endpoints
	routes["/api/v1/config/requirement-types"] = []string{"POST", "GET"}
	routes["/api/v1/config/requirement-types/{id}"] = []string{"GET", "PUT", "DELETE"}
	routes["/api/v1/config/relationship-types"] = []string{"POST", "GET"}
	routes["/api/v1/config/relationship-types/{id}"] = []string{"GET", "PUT", "DELETE"}
	routes["/api/v1/config/status-models"] = []string{"POST", "GET"}
	routes["/api/v1/config/status-models/{id}"] = []string{"GET", "PUT", "DELETE"}
	routes["/api/v1/config/status-models/default/{entity_type}"] = []string{"GET"}
	routes["/api/v1/config/status-models/{id}/statuses"] = []string{"GET"}
	routes["/api/v1/config/status-models/{id}/transitions"] = []string{"GET"}
	routes["/api/v1/config/statuses"] = []string{"POST"}
	routes["/api/v1/config/statuses/{id}"] = []string{"GET", "PUT", "DELETE"}
	routes["/api/v1/config/status-transitions"] = []string{"POST"}
	routes["/api/v1/config/status-transitions/{id}"] = []string{"GET", "PUT", "DELETE"}

	// General deletion confirmation
	routes["/api/v1/deletion/confirm"] = []string{"GET"}

	// Comment endpoints
	routes["/api/v1/comments/{id}"] = []string{"GET", "PUT", "DELETE"}
	routes["/api/v1/comments/{id}/resolve"] = []string{"POST"}
	routes["/api/v1/comments/{id}/unresolve"] = []string{"POST"}
	routes["/api/v1/comments/status/{status}"] = []string{"GET"}
	routes["/api/v1/comments/{id}/replies"] = []string{"GET", "POST"}

	return routes
}

func getDocumentedRoutes() (map[string][]string, error) {
	content, err := os.ReadFile("docs/openapi-v3.yaml")
	if err != nil {
		return nil, err
	}

	routes := make(map[string][]string)
	lines := strings.Split(string(content), "\n")

	inPaths := false
	currentPath := ""

	pathPattern := regexp.MustCompile(`^\s*(/[^:]*):$`)
	methodPattern := regexp.MustCompile(`^\s*(get|post|put|patch|delete|head|options):$`)

	for _, line := range lines {
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
			if _, exists := routes[currentPath]; !exists {
				routes[currentPath] = []string{}
			}
		}

		// Check for method definition
		if currentPath != "" {
			if matches := methodPattern.FindStringSubmatch(line); len(matches) > 1 {
				method := strings.ToUpper(matches[1])
				routes[currentPath] = append(routes[currentPath], method)
			}
		}
	}

	return routes, nil
}

func validateRouteCompleteness(implemented, documented map[string][]string) {
	fmt.Println("1. Routes Missing from OpenAPI Documentation:")
	fmt.Println(strings.Repeat("=", 60))

	missingCount := 0
	var sortedPaths []string
	for path := range implemented {
		sortedPaths = append(sortedPaths, path)
	}
	sort.Strings(sortedPaths)

	for _, path := range sortedPaths {
		implMethods := implemented[path]
		docMethods := documented[path]

		if len(docMethods) == 0 {
			fmt.Printf("   ❌ %s - Path not documented (methods: %s)\n", path, strings.Join(implMethods, ", "))
			missingCount++
		} else {
			for _, method := range implMethods {
				if !contains(docMethods, method) {
					fmt.Printf("   ❌ %s %s - Method not documented\n", method, path)
					missingCount++
				}
			}
		}
	}

	if missingCount == 0 {
		fmt.Println("   ✅ All implemented routes are documented")
	}

	fmt.Println("\n2. Documented Routes Without Implementation:")
	fmt.Println(strings.Repeat("=", 60))

	extraCount := 0
	var sortedDocPaths []string
	for path := range documented {
		sortedDocPaths = append(sortedDocPaths, path)
	}
	sort.Strings(sortedDocPaths)

	for _, path := range sortedDocPaths {
		docMethods := documented[path]
		implMethods := implemented[path]

		if len(implMethods) == 0 {
			fmt.Printf("   ❌ %s - Path documented but not implemented (methods: %s)\n", path, strings.Join(docMethods, ", "))
			extraCount++
		} else {
			for _, method := range docMethods {
				if !contains(implMethods, method) {
					fmt.Printf("   ❌ %s %s - Method documented but not implemented\n", method, path)
					extraCount++
				}
			}
		}
	}

	if extraCount == 0 {
		fmt.Println("   ✅ All documented routes have implementations")
	}

	fmt.Println("\n3. Entity Coverage Analysis:")
	fmt.Println(strings.Repeat("=", 60))
	analyzeEntityCoverage(implemented, documented)

	fmt.Println("\n4. Summary:")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("   Total implemented routes: %d\n", countTotalRoutes(implemented))
	fmt.Printf("   Total documented routes: %d\n", countTotalRoutes(documented))
	fmt.Printf("   Missing documentation: %d\n", missingCount)
	fmt.Printf("   Extra documentation: %d\n", extraCount)

	if missingCount == 0 && extraCount == 0 {
		fmt.Println("\n   ✅ OpenAPI specification is complete and accurate!")
	} else {
		fmt.Println("\n   ❌ OpenAPI specification needs updates")
	}
}

func analyzeEntityCoverage(implemented, documented map[string][]string) {
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

			implMethods := implemented[path]
			docMethods := documented[path]

			hasImpl := contains(implMethods, method)
			hasDoc := contains(docMethods, method)

			status := getStatus(hasImpl, hasDoc)
			fmt.Printf("     %s %s %s\n", status, method, path)
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
			{fmt.Sprintf("/api/v1/%s/{id}/comments/inline/visible", entity), "GET", "Get Visible Inline Comments"},
			{fmt.Sprintf("/api/v1/%s/{id}/comments/inline/validate", entity), "POST", "Validate Inline Comments"},
		}

		for _, op := range specialOps {
			implMethods := implemented[op.path]
			docMethods := documented[op.path]

			hasImpl := contains(implMethods, op.method)
			hasDoc := contains(docMethods, op.method)

			if hasImpl || hasDoc {
				status := getStatus(hasImpl, hasDoc)
				fmt.Printf("     %s %s %s - %s\n", status, op.method, op.path, op.desc)
			}
		}
	}

	fmt.Println("\n   Legend:")
	fmt.Println("     ✅ Implemented and documented")
	fmt.Println("     📝 Implemented but not documented")
	fmt.Println("     🔧 Documented but not implemented")
	fmt.Println("     ❌ Neither implemented nor documented")
}

func getStatus(implemented, documented bool) string {
	if implemented && documented {
		return "✅"
	} else if implemented && !documented {
		return "📝"
	} else if !implemented && documented {
		return "🔧"
	}
	return "❌"
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func countTotalRoutes(routes map[string][]string) int {
	total := 0
	for _, methods := range routes {
		total += len(methods)
	}
	return total
}
