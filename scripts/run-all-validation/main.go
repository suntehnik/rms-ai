package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// TestResult represents the result of a validation test
type TestResult struct {
	Name     string
	Passed   bool
	Duration time.Duration
	Output   string
	Error    string
}

// TestSuite represents a collection of related tests
type TestSuite struct {
	Name   string
	Tests  []TestResult
	Passed int
	Total  int
}

func main() {
	fmt.Println("=== Comprehensive API Documentation Validation ===\n")

	// Define all validation test suites
	suites := []TestSuite{
		{
			Name: "Route Implementation vs Documentation",
			Tests: []TestResult{
				runTest("TestOpenAPIRouteCompleteness", "./internal/validation"),
			},
		},
		{
			Name: "Response Schema Validation",
			Tests: []TestResult{
				runTest("TestResponseSchemaValidation", "./internal/validation"),
			},
		},
		{
			Name: "Authentication Documentation",
			Tests: []TestResult{
				runTest("TestAuthenticationDocumentation", "./internal/validation"),
			},
		},
		{
			Name: "Documentation Completeness",
			Tests: []TestResult{
				runTest("TestDocumentationCompleteness", "./internal/validation"),
			},
		},
		{
			Name: "Existing OpenAPI Validation",
			Tests: []TestResult{
				runTest("TestOpenAPISchemaCompliance", "./internal/docs"),
				runTest("TestSwaggerSpecificationCompleteness", "./internal/docs"),
			},
		},
		{
			Name: "Legacy Validation Scripts",
			Tests: []TestResult{
				runLegacyScript("validate_api_completeness.go"),
				runLegacyScript("validate_openapi_completeness.go"),
				runLegacyScript("validate_schemas_and_parameters.go"),
			},
		},
	}

	// Calculate totals for each suite
	for i := range suites {
		suite := &suites[i]
		suite.Total = len(suite.Tests)
		suite.Passed = 0
		for _, test := range suite.Tests {
			if test.Passed {
				suite.Passed++
			}
		}
	}

	// Display results
	displayResults(suites)

	// Generate comprehensive report
	generateReport(suites)

	// Calculate overall statistics
	totalTests := 0
	passedTests := 0
	for _, suite := range suites {
		totalTests += suite.Total
		passedTests += suite.Passed
	}

	// Exit with appropriate code
	if passedTests == totalTests {
		fmt.Println("🎉 All documentation validation tests passed!")
		os.Exit(0)
	} else {
		fmt.Printf("⚠️  %d out of %d tests failed. Check the validation report for details.\n", totalTests-passedTests, totalTests)
		os.Exit(1)
	}
}

func runTest(testName, packagePath string) TestResult {
	start := time.Now()

	cmd := exec.Command("go", "test", "-v", packagePath, "-run", testName)
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	result := TestResult{
		Name:     testName,
		Passed:   err == nil,
		Duration: duration,
		Output:   string(output),
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result
}

func runLegacyScript(scriptName string) TestResult {
	start := time.Now()

	scriptPath := filepath.Join("scripts", scriptName)
	cmd := exec.Command("go", "run", scriptPath)
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	result := TestResult{
		Name:     strings.TrimSuffix(scriptName, ".go"),
		Passed:   err == nil,
		Duration: duration,
		Output:   string(output),
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result
}

func displayResults(suites []TestSuite) {
	for _, suite := range suites {
		fmt.Printf("📋 %s:\n", suite.Name)
		fmt.Println(strings.Repeat("=", 60))

		for _, test := range suite.Tests {
			status := "❌ FAILED"
			if test.Passed {
				status = "✅ PASSED"
			}

			fmt.Printf("  %s %s (%.2fs)\n", status, test.Name, test.Duration.Seconds())

			if !test.Passed && test.Error != "" {
				// Show first few lines of error for quick overview
				errorLines := strings.Split(test.Error, "\n")
				for i, line := range errorLines {
					if i >= 3 { // Limit to first 3 lines
						fmt.Printf("    ... (see full report for details)\n")
						break
					}
					if strings.TrimSpace(line) != "" {
						fmt.Printf("    %s\n", strings.TrimSpace(line))
					}
				}
			}
		}

		fmt.Printf("\n  Suite Summary: %d/%d passed (%.1f%%)\n\n",
			suite.Passed, suite.Total, float64(suite.Passed)/float64(suite.Total)*100)
	}
}

func generateReport(suites []TestSuite) {
	reportPath := "docs/validation-report.md"

	file, err := os.Create(reportPath)
	if err != nil {
		log.Printf("Failed to create validation report: %v", err)
		return
	}
	defer file.Close()

	// Write report header
	fmt.Fprintf(file, "# Comprehensive API Documentation Validation Report\n\n")
	fmt.Fprintf(file, "Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	// Calculate overall statistics
	totalTests := 0
	passedTests := 0
	for _, suite := range suites {
		totalTests += suite.Total
		passedTests += suite.Passed
	}

	// Write executive summary
	fmt.Fprintf(file, "## Executive Summary\n\n")
	fmt.Fprintf(file, "- **Total Tests**: %d\n", totalTests)
	fmt.Fprintf(file, "- **Passed**: %d\n", passedTests)
	fmt.Fprintf(file, "- **Failed**: %d\n", totalTests-passedTests)
	fmt.Fprintf(file, "- **Success Rate**: %.1f%%\n\n", float64(passedTests)/float64(totalTests)*100)

	if passedTests == totalTests {
		fmt.Fprintf(file, "🎉 **All validation tests passed!** The API documentation is accurate and complete.\n\n")
	} else {
		fmt.Fprintf(file, "⚠️ **Issues Found**: %d tests failed. Review the detailed results below.\n\n", totalTests-passedTests)
	}

	// Write suite summaries
	fmt.Fprintf(file, "## Test Suite Summary\n\n")
	fmt.Fprintf(file, "| Suite | Passed | Total | Success Rate |\n")
	fmt.Fprintf(file, "|-------|--------|-------|-------------|\n")

	for _, suite := range suites {
		successRate := float64(suite.Passed) / float64(suite.Total) * 100
		status := "✅"
		if suite.Passed < suite.Total {
			status = "❌"
		}
		fmt.Fprintf(file, "| %s %s | %d | %d | %.1f%% |\n",
			status, suite.Name, suite.Passed, suite.Total, successRate)
	}
	fmt.Fprintf(file, "\n")

	// Write detailed results
	fmt.Fprintf(file, "## Detailed Test Results\n\n")

	for _, suite := range suites {
		fmt.Fprintf(file, "### %s\n\n", suite.Name)

		for _, test := range suite.Tests {
			status := "❌ FAILED"
			if test.Passed {
				status = "✅ PASSED"
			}

			fmt.Fprintf(file, "#### %s %s\n\n", status, test.Name)
			fmt.Fprintf(file, "- **Duration**: %.2fs\n", test.Duration.Seconds())

			if !test.Passed {
				fmt.Fprintf(file, "- **Error**: %s\n", strings.TrimSpace(test.Error))

				// Include relevant output (truncated for readability)
				if len(test.Output) > 0 {
					fmt.Fprintf(file, "\n**Output:**\n```\n")
					output := test.Output
					if len(output) > 3000 {
						output = output[:3000] + "\n... (truncated for readability)"
					}
					fmt.Fprintf(file, "%s\n```\n\n", output)
				}
			} else {
				fmt.Fprintf(file, "- **Status**: Test passed successfully\n")

				// Include summary output for passed tests
				if len(test.Output) > 0 {
					lines := strings.Split(test.Output, "\n")
					summaryLines := []string{}
					for _, line := range lines {
						if strings.Contains(line, "✅") || strings.Contains(line, "PASS") {
							summaryLines = append(summaryLines, strings.TrimSpace(line))
						}
					}
					if len(summaryLines) > 0 {
						fmt.Fprintf(file, "\n**Summary:**\n```\n")
						for _, line := range summaryLines {
							if len(summaryLines) <= 10 || len(line) <= 100 {
								fmt.Fprintf(file, "%s\n", line)
							}
						}
						fmt.Fprintf(file, "```\n\n")
					}
				}
			}
		}
	}

	// Write recommendations
	fmt.Fprintf(file, "## Recommendations\n\n")

	hasFailures := passedTests < totalTests
	if hasFailures {
		fmt.Fprintf(file, "### Issues to Address\n\n")

		for _, suite := range suites {
			if suite.Passed < suite.Total {
				fmt.Fprintf(file, "#### %s\n\n", suite.Name)

				for _, test := range suite.Tests {
					if !test.Passed {
						fmt.Fprintf(file, "- **%s**: %s\n", test.Name, extractMainError(test.Error, test.Output))
					}
				}
				fmt.Fprintf(file, "\n")
			}
		}

		fmt.Fprintf(file, "### Action Items\n\n")
		fmt.Fprintf(file, "1. **Route Documentation**: Update OpenAPI specification to match actual route implementations\n")
		fmt.Fprintf(file, "2. **Schema Validation**: Ensure all response schemas are properly defined and consistent\n")
		fmt.Fprintf(file, "3. **Authentication**: Verify authentication requirements are correctly documented\n")
		fmt.Fprintf(file, "4. **Completeness**: Add missing descriptions, examples, and parameter documentation\n")
		fmt.Fprintf(file, "5. **Testing**: Re-run validation tests after making corrections\n\n")

		fmt.Fprintf(file, "### Commands to Fix Issues\n\n")
		fmt.Fprintf(file, "```bash\n")
		fmt.Fprintf(file, "# Update OpenAPI specification\n")
		fmt.Fprintf(file, "make swagger\n\n")
		fmt.Fprintf(file, "# Generate updated documentation\n")
		fmt.Fprintf(file, "make docs-generate\n\n")
		fmt.Fprintf(file, "# Re-run validation\n")
		fmt.Fprintf(file, "make docs-validate-all\n")
		fmt.Fprintf(file, "```\n\n")
	} else {
		fmt.Fprintf(file, "### Maintenance\n\n")
		fmt.Fprintf(file, "✅ All validation tests are currently passing. To maintain documentation quality:\n\n")
		fmt.Fprintf(file, "1. Run validation tests regularly: `make docs-validate-all`\n")
		fmt.Fprintf(file, "2. Update documentation when adding new endpoints\n")
		fmt.Fprintf(file, "3. Ensure new schemas are properly documented\n")
		fmt.Fprintf(file, "4. Maintain authentication documentation accuracy\n\n")
	}

	// Write validation commands reference
	fmt.Fprintf(file, "## Validation Commands Reference\n\n")
	fmt.Fprintf(file, "| Command | Description |\n")
	fmt.Fprintf(file, "|---------|-------------|\n")
	fmt.Fprintf(file, "| `make docs-validate` | Run comprehensive validation script |\n")
	fmt.Fprintf(file, "| `make docs-validate-routes` | Validate route implementation vs documentation |\n")
	fmt.Fprintf(file, "| `make docs-validate-schemas` | Validate response schema consistency |\n")
	fmt.Fprintf(file, "| `make docs-validate-auth` | Validate authentication documentation |\n")
	fmt.Fprintf(file, "| `make docs-validate-completeness` | Validate documentation completeness |\n")
	fmt.Fprintf(file, "| `make docs-validate-all` | Run all validation tests |\n")
	fmt.Fprintf(file, "| `go run scripts/run_all_validation_tests.go` | Run this comprehensive validation |\n\n")

	fmt.Printf("📄 Comprehensive validation report generated: %s\n", reportPath)
}

func extractMainError(errorOutput, testOutput string) string {
	// Try to extract meaningful error from error output first
	if errorOutput != "" {
		lines := strings.Split(errorOutput, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "FAIL:") || strings.Contains(line, "Error:") {
				return line
			}
		}
	}

	// Try to extract from test output
	if testOutput != "" {
		lines := strings.Split(testOutput, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "❌") && len(line) < 200 {
				return line
			}
		}
	}

	// Return generic message if no specific error found
	return "Test failed - see detailed output above"
}

func init() {
	// Ensure we're in the right directory
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		fmt.Println("❌ This script must be run from the project root directory")
		os.Exit(1)
	}

	// Check for required files
	requiredFiles := []string{
		"docs/openapi-v3.yaml",
		"internal/server/routes/routes.go",
	}

	for _, file := range requiredFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Printf("❌ Required file not found: %s\n", file)
			fmt.Println("Please ensure the project is properly set up before running validation.")
			os.Exit(1)
		}
	}
}
