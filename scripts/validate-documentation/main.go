package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ValidationResult represents the result of a validation test
type ValidationResult struct {
	TestName    string
	Passed      bool
	Duration    time.Duration
	Output      string
	ErrorOutput string
}

// ValidationSuite represents a collection of validation tests
type ValidationSuite struct {
	Name    string
	Results []ValidationResult
}

func main() {
	fmt.Println("=== API Documentation Accuracy Validation ===\n")

	// Define validation test suites
	suites := []ValidationSuite{
		{
			Name: "Route Implementation vs Documentation",
			Results: []ValidationResult{
				runValidationTest("TestOpenAPIRouteCompleteness", "./internal/validation"),
			},
		},
		{
			Name: "Response Schema Validation",
			Results: []ValidationResult{
				runValidationTest("TestResponseSchemaValidation", "./internal/validation"),
			},
		},
		{
			Name: "Authentication Documentation",
			Results: []ValidationResult{
				runValidationTest("TestAuthenticationDocumentation", "./internal/validation"),
			},
		},
		{
			Name: "Documentation Completeness",
			Results: []ValidationResult{
				runValidationTest("TestDocumentationCompleteness", "./internal/validation"),
			},
		},
		{
			Name: "Existing OpenAPI Validation",
			Results: []ValidationResult{
				runValidationTest("TestOpenAPISchemaCompliance", "./internal/docs"),
				runValidationTest("TestSwaggerSpecificationCompleteness", "./internal/docs"),
			},
		},
	}

	// Run all validation suites
	totalTests := 0
	passedTests := 0

	for _, suite := range suites {
		fmt.Printf("Running %s:\n", suite.Name)
		fmt.Println(strings.Repeat("=", 60))

		for _, result := range suite.Results {
			totalTests++

			status := "❌ FAILED"
			if result.Passed {
				status = "✅ PASSED"
				passedTests++
			}

			fmt.Printf("  %s %s (%.2fs)\n", status, result.TestName, result.Duration.Seconds())

			if !result.Passed && result.ErrorOutput != "" {
				fmt.Printf("    Error: %s\n", strings.TrimSpace(result.ErrorOutput))
			}
		}

		fmt.Println()
	}

	// Generate summary report
	generateSummaryReport(suites, totalTests, passedTests)

	// Generate detailed validation report
	generateDetailedReport(suites)

	// Exit with appropriate code
	if passedTests == totalTests {
		fmt.Println("🎉 All documentation validation tests passed!")
		os.Exit(0)
	} else {
		fmt.Printf("⚠️  %d out of %d tests failed. Please review the issues above.\n", totalTests-passedTests, totalTests)
		os.Exit(1)
	}
}

func runValidationTest(testName, packagePath string) ValidationResult {
	start := time.Now()

	cmd := exec.Command("go", "test", "-v", packagePath, "-run", testName)

	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	result := ValidationResult{
		TestName: testName,
		Passed:   err == nil,
		Duration: duration,
		Output:   string(output),
	}

	if err != nil {
		result.ErrorOutput = err.Error()
		if len(output) > 0 {
			result.ErrorOutput = string(output)
		}
	}

	return result
}

func generateSummaryReport(suites []ValidationSuite, totalTests, passedTests int) {
	fmt.Println("=== Validation Summary ===")
	fmt.Printf("Total Tests: %d\n", totalTests)
	fmt.Printf("Passed: %d\n", passedTests)
	fmt.Printf("Failed: %d\n", totalTests-passedTests)
	fmt.Printf("Success Rate: %.1f%%\n", float64(passedTests)/float64(totalTests)*100)
	fmt.Println()

	// Suite-by-suite breakdown
	for _, suite := range suites {
		suitePassed := 0
		suiteTotal := len(suite.Results)

		for _, result := range suite.Results {
			if result.Passed {
				suitePassed++
			}
		}

		status := "✅"
		if suitePassed < suiteTotal {
			status = "❌"
		}

		fmt.Printf("%s %s: %d/%d passed\n", status, suite.Name, suitePassed, suiteTotal)
	}
	fmt.Println()
}

func generateDetailedReport(suites []ValidationSuite) {
	reportPath := "docs/validation-report.md"

	file, err := os.Create(reportPath)
	if err != nil {
		log.Printf("Failed to create detailed report: %v", err)
		return
	}
	defer file.Close()

	// Write report header
	fmt.Fprintf(file, "# API Documentation Validation Report\n\n")
	fmt.Fprintf(file, "Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	// Write summary
	totalTests := 0
	passedTests := 0
	for _, suite := range suites {
		for _, result := range suite.Results {
			totalTests++
			if result.Passed {
				passedTests++
			}
		}
	}

	fmt.Fprintf(file, "## Summary\n\n")
	fmt.Fprintf(file, "- **Total Tests**: %d\n", totalTests)
	fmt.Fprintf(file, "- **Passed**: %d\n", passedTests)
	fmt.Fprintf(file, "- **Failed**: %d\n", totalTests-passedTests)
	fmt.Fprintf(file, "- **Success Rate**: %.1f%%\n\n", float64(passedTests)/float64(totalTests)*100)

	// Write detailed results
	fmt.Fprintf(file, "## Detailed Results\n\n")

	for _, suite := range suites {
		fmt.Fprintf(file, "### %s\n\n", suite.Name)

		for _, result := range suite.Results {
			status := "❌ FAILED"
			if result.Passed {
				status = "✅ PASSED"
			}

			fmt.Fprintf(file, "#### %s %s\n\n", status, result.TestName)
			fmt.Fprintf(file, "- **Duration**: %.2fs\n", result.Duration.Seconds())

			if !result.Passed {
				fmt.Fprintf(file, "- **Error**: %s\n", strings.TrimSpace(result.ErrorOutput))
			}

			// Include relevant output (truncated)
			if len(result.Output) > 0 {
				fmt.Fprintf(file, "\n**Output:**\n```\n")
				output := result.Output
				if len(output) > 2000 {
					output = output[:2000] + "\n... (truncated)"
				}
				fmt.Fprintf(file, "%s\n```\n\n", output)
			}
		}
	}

	// Write recommendations
	fmt.Fprintf(file, "## Recommendations\n\n")

	hasFailures := passedTests < totalTests
	if hasFailures {
		fmt.Fprintf(file, "### Issues Found\n\n")
		fmt.Fprintf(file, "The following issues were identified in the documentation validation:\n\n")

		for _, suite := range suites {
			for _, result := range suite.Results {
				if !result.Passed {
					fmt.Fprintf(file, "- **%s**: %s\n", result.TestName, extractMainError(result.ErrorOutput))
				}
			}
		}

		fmt.Fprintf(file, "\n### Next Steps\n\n")
		fmt.Fprintf(file, "1. Review the failed tests above\n")
		fmt.Fprintf(file, "2. Update the OpenAPI specification to match implementation\n")
		fmt.Fprintf(file, "3. Ensure all endpoints have proper documentation\n")
		fmt.Fprintf(file, "4. Verify authentication requirements are correctly documented\n")
		fmt.Fprintf(file, "5. Re-run validation tests to confirm fixes\n\n")
	} else {
		fmt.Fprintf(file, "✅ All validation tests passed! The API documentation is accurate and complete.\n\n")
	}

	fmt.Printf("📄 Detailed report generated: %s\n", reportPath)
}

func extractMainError(errorOutput string) string {
	lines := strings.Split(errorOutput, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "FAIL:") || strings.Contains(line, "Error:") {
			return line
		}
	}

	// Return first non-empty line if no specific error found
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "=== RUN") && !strings.HasPrefix(line, "---") {
			return line
		}
	}

	return "Test failed"
}

// Additional validation functions

func validateOpenAPIFileExists() bool {
	_, err := os.Stat("docs/openapi-v3.yaml")
	return err == nil
}

func validateGeneratedDocsExist() bool {
	requiredFiles := []string{
		"docs/generated/api-documentation.html",
		"docs/generated/api-documentation.md",
		"docs/generated/api-types.ts",
		"docs/generated/developer-guide.md",
	}

	for _, file := range requiredFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return false
		}
	}

	return true
}

func runPreValidationChecks() []string {
	var issues []string

	if !validateOpenAPIFileExists() {
		issues = append(issues, "OpenAPI specification file (docs/openapi-v3.yaml) not found")
	}

	if !validateGeneratedDocsExist() {
		issues = append(issues, "Generated documentation files missing. Run 'make docs-generate' first")
	}

	// Check if routes.go exists
	if _, err := os.Stat("internal/server/routes/routes.go"); os.IsNotExist(err) {
		issues = append(issues, "Routes file (internal/server/routes/routes.go) not found")
	}

	return issues
}

func init() {
	// Run pre-validation checks
	issues := runPreValidationChecks()
	if len(issues) > 0 {
		fmt.Println("❌ Pre-validation checks failed:")
		for _, issue := range issues {
			fmt.Printf("  - %s\n", issue)
		}
		fmt.Println("\nPlease resolve these issues before running validation.")
		os.Exit(1)
	}
}
