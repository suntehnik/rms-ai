package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"product-requirements-management/internal/docs"
)

func main() {
	var (
		outputFile = flag.String("output", "docs-metrics.json", "Output file for metrics")
		format     = flag.String("format", "json", "Output format: json, text, or summary")
		verbose    = flag.Bool("verbose", false, "Verbose output")
	)
	flag.Parse()

	// Generate metrics
	fmt.Println("Generating documentation metrics...")
	metrics, err := docs.GenerateDocumentationMetrics()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating metrics: %v\n", err)
		os.Exit(1)
	}

	// Output based on format
	switch *format {
	case "json":
		if err := outputJSON(metrics, *outputFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing JSON output: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Metrics saved to %s\n", *outputFile)

	case "text":
		outputText(metrics, *verbose)

	case "summary":
		outputSummary(metrics)

	default:
		fmt.Fprintf(os.Stderr, "Unknown format: %s\n", *format)
		os.Exit(1)
	}
}

func outputJSON(metrics *docs.DocumentationMetrics, filename string) error {
	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func outputText(metrics *docs.DocumentationMetrics, verbose bool) {
	fmt.Println("=== Documentation Quality Metrics ===")
	fmt.Printf("Generated: %s\n\n", metrics.GeneratedAt.Format("2006-01-02 15:04:05"))

	// Overall Quality Score
	fmt.Printf("Overall Quality Score: %.1f%% (Grade: %s)\n\n",
		metrics.QualityScore.OverallScore, metrics.QualityScore.Grade)

	// Annotation Coverage
	fmt.Println("📝 Annotation Coverage:")
	fmt.Printf("  Total Handlers: %d\n", metrics.AnnotationCoverage.TotalHandlers)
	fmt.Printf("  Annotated: %d (%.1f%%)\n",
		metrics.AnnotationCoverage.AnnotatedHandlers,
		metrics.AnnotationCoverage.CoveragePercentage)
	fmt.Printf("  Missing Annotations: %d\n\n", metrics.AnnotationCoverage.MissingCount)

	// Endpoint Coverage
	fmt.Println("🌐 Endpoint Coverage:")
	fmt.Printf("  Total Endpoints: %d\n", metrics.EndpointCoverage.TotalEndpoints)
	fmt.Printf("  Documented: %d (%.1f%%)\n",
		metrics.EndpointCoverage.DocumentedEndpoints,
		metrics.EndpointCoverage.CoveragePercentage)

	if verbose {
		fmt.Println("  By HTTP Method:")
		for method, count := range metrics.EndpointCoverage.EndpointsByMethod {
			fmt.Printf("    %s: %d\n", method, count)
		}
	}
	fmt.Println()

	// Example Coverage
	fmt.Println("📋 Example Coverage:")
	fmt.Printf("  Handlers with Examples: %d (%.1f%%)\n",
		metrics.ExampleCoverage.HandlersWithExamples,
		metrics.ExampleCoverage.ExampleCoveragePercent)
	fmt.Printf("  Request Body Examples: %d\n", metrics.ExampleCoverage.RequestBodyExamples)
	fmt.Printf("  Query Parameter Examples: %d\n", metrics.ExampleCoverage.QueryParameterExamples)
	fmt.Println()

	// Tag Distribution
	if verbose && len(metrics.TagDistribution) > 0 {
		fmt.Println("🏷️  Tag Distribution:")
		for tag, count := range metrics.TagDistribution {
			fmt.Printf("  %s: %d endpoints\n", tag, count)
		}
		fmt.Println()
	}

	// Component Scores
	if verbose {
		fmt.Println("📊 Component Scores:")
		for component, score := range metrics.QualityScore.ComponentScores {
			fmt.Printf("  %s: %.1f%%\n", component, score)
		}
		fmt.Println()
	}

	// Recommendations
	if len(metrics.Recommendations) > 0 {
		fmt.Println("💡 Recommendations:")
		for i, rec := range metrics.Recommendations {
			fmt.Printf("  %d. %s\n", i+1, rec)
		}
		fmt.Println()
	}

	// Missing Annotations (if verbose)
	if verbose && len(metrics.MissingAnnotations) > 0 {
		fmt.Println("⚠️  Missing Annotations (Top 10):")
		count := len(metrics.MissingAnnotations)
		if count > 10 {
			count = 10
		}

		for i := 0; i < count; i++ {
			missing := metrics.MissingAnnotations[i]
			fmt.Printf("  %s:%s [%s] - %v\n",
				missing.File, missing.Handler, missing.Severity, missing.MissingFields)
		}

		if len(metrics.MissingAnnotations) > 10 {
			fmt.Printf("  ... and %d more\n", len(metrics.MissingAnnotations)-10)
		}
	}
}

func outputSummary(metrics *docs.DocumentationMetrics) {
	fmt.Printf("Documentation Quality: %.1f%% (%s)\n",
		metrics.QualityScore.OverallScore, metrics.QualityScore.Grade)
	fmt.Printf("Annotation Coverage: %.1f%% (%d/%d handlers)\n",
		metrics.AnnotationCoverage.CoveragePercentage,
		metrics.AnnotationCoverage.AnnotatedHandlers,
		metrics.AnnotationCoverage.TotalHandlers)
	fmt.Printf("Endpoint Coverage: %.1f%% (%d/%d endpoints)\n",
		metrics.EndpointCoverage.CoveragePercentage,
		metrics.EndpointCoverage.DocumentedEndpoints,
		metrics.EndpointCoverage.TotalEndpoints)

	if len(metrics.Recommendations) > 0 {
		fmt.Printf("Top Recommendation: %s\n", metrics.Recommendations[0])
	}
}
