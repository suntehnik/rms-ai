package docs

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateDocumentationMetrics tests the metrics generation functionality
func TestGenerateDocumentationMetrics(t *testing.T) {
	metrics, err := GenerateDocumentationMetrics()
	require.NoError(t, err, "Should be able to generate documentation metrics")
	require.NotNil(t, metrics, "Metrics should not be nil")

	t.Run("BasicStructure", func(t *testing.T) {
		assert.False(t, metrics.GeneratedAt.IsZero(), "Should have generation timestamp")
		assert.NotNil(t, metrics.TagDistribution, "Should have tag distribution")
		assert.NotNil(t, metrics.MissingAnnotations, "Should have missing annotations list")
		assert.NotNil(t, metrics.Recommendations, "Should have recommendations list")
	})

	t.Run("AnnotationCoverage", func(t *testing.T) {
		coverage := metrics.AnnotationCoverage
		assert.GreaterOrEqual(t, coverage.TotalHandlers, 0, "Should count handlers")
		assert.GreaterOrEqual(t, coverage.AnnotatedHandlers, 0, "Should count annotated handlers")
		assert.LessOrEqual(t, coverage.AnnotatedHandlers, coverage.TotalHandlers,
			"Annotated handlers should not exceed total")
		assert.GreaterOrEqual(t, coverage.CoveragePercentage, 0.0, "Coverage should be non-negative")
		assert.LessOrEqual(t, coverage.CoveragePercentage, 100.0, "Coverage should not exceed 100%")
	})

	t.Run("EndpointCoverage", func(t *testing.T) {
		coverage := metrics.EndpointCoverage
		assert.GreaterOrEqual(t, coverage.TotalEndpoints, 0, "Should count endpoints")
		assert.GreaterOrEqual(t, coverage.DocumentedEndpoints, 0, "Should count documented endpoints")
		assert.LessOrEqual(t, coverage.DocumentedEndpoints, coverage.TotalEndpoints,
			"Documented endpoints should not exceed total")
		assert.NotNil(t, coverage.EndpointsByMethod, "Should have method distribution")
		assert.NotNil(t, coverage.EndpointsByTag, "Should have tag distribution")
	})

	t.Run("ExampleCoverage", func(t *testing.T) {
		coverage := metrics.ExampleCoverage
		assert.GreaterOrEqual(t, coverage.HandlersWithExamples, 0, "Should count handlers with examples")
		assert.GreaterOrEqual(t, coverage.ExampleCoveragePercent, 0.0, "Example coverage should be non-negative")
		assert.LessOrEqual(t, coverage.ExampleCoveragePercent, 100.0, "Example coverage should not exceed 100%")
		assert.GreaterOrEqual(t, coverage.RequestBodyExamples, 0, "Should count request body examples")
		assert.GreaterOrEqual(t, coverage.ResponseExamples, 0, "Should count response examples")
		assert.GreaterOrEqual(t, coverage.QueryParameterExamples, 0, "Should count query parameter examples")
	})

	t.Run("QualityScore", func(t *testing.T) {
		quality := metrics.QualityScore
		assert.GreaterOrEqual(t, quality.OverallScore, 0.0, "Overall score should be non-negative")
		assert.LessOrEqual(t, quality.OverallScore, 100.0, "Overall score should not exceed 100%")
		assert.NotEmpty(t, quality.Grade, "Should have a grade")
		assert.NotNil(t, quality.ComponentScores, "Should have component scores")
		assert.NotNil(t, quality.ImprovementAreas, "Should have improvement areas")

		// Validate component scores
		expectedComponents := []string{"annotation_coverage", "endpoint_coverage", "example_coverage", "tag_distribution"}
		for _, component := range expectedComponents {
			score, exists := quality.ComponentScores[component]
			assert.True(t, exists, "Should have %s component score", component)
			assert.GreaterOrEqual(t, score, 0.0, "%s score should be non-negative", component)
			assert.LessOrEqual(t, score, 100.0, "%s score should not exceed 100%", component)
		}
	})

	t.Run("MissingAnnotations", func(t *testing.T) {
		for i, missing := range metrics.MissingAnnotations {
			assert.NotEmpty(t, missing.File, "Missing annotation %d should have file", i)
			assert.NotEmpty(t, missing.Handler, "Missing annotation %d should have handler", i)
			assert.NotEmpty(t, missing.MissingFields, "Missing annotation %d should have missing fields", i)
			assert.Contains(t, []string{"high", "medium", "low"}, missing.Severity,
				"Missing annotation %d should have valid severity", i)
		}
	})
}

// TestMetricsSerialization tests saving and loading metrics
func TestMetricsSerialization(t *testing.T) {
	// Generate metrics
	originalMetrics, err := GenerateDocumentationMetrics()
	require.NoError(t, err, "Should be able to generate metrics")

	// Save to temporary file
	tempFile := "test-metrics.json"
	defer os.Remove(tempFile)

	err = SaveMetricsToFile(originalMetrics, tempFile)
	require.NoError(t, err, "Should be able to save metrics to file")

	// Verify file exists and is valid JSON
	data, err := os.ReadFile(tempFile)
	require.NoError(t, err, "Should be able to read saved file")

	var jsonData map[string]interface{}
	err = json.Unmarshal(data, &jsonData)
	require.NoError(t, err, "Saved file should be valid JSON")

	// Load metrics back
	loadedMetrics, err := LoadMetricsFromFile(tempFile)
	require.NoError(t, err, "Should be able to load metrics from file")

	// Compare key fields
	assert.Equal(t, originalMetrics.AnnotationCoverage.TotalHandlers,
		loadedMetrics.AnnotationCoverage.TotalHandlers, "Total handlers should match")
	assert.Equal(t, originalMetrics.AnnotationCoverage.AnnotatedHandlers,
		loadedMetrics.AnnotationCoverage.AnnotatedHandlers, "Annotated handlers should match")
	assert.Equal(t, originalMetrics.QualityScore.Grade,
		loadedMetrics.QualityScore.Grade, "Grade should match")
}

// TestMetricsCalculations tests specific calculation logic
func TestMetricsCalculations(t *testing.T) {
	t.Run("PercentageCalculation", func(t *testing.T) {
		assert.Equal(t, 0.0, calculatePercentage(0, 0), "0/0 should be 0%")
		assert.Equal(t, 0.0, calculatePercentage(0, 10), "0/10 should be 0%")
		assert.Equal(t, 50.0, calculatePercentage(5, 10), "5/10 should be 50%")
		assert.Equal(t, 100.0, calculatePercentage(10, 10), "10/10 should be 100%")
	})

	t.Run("GradeCalculation", func(t *testing.T) {
		assert.Equal(t, "A", determineGrade(95.0), "95% should be grade A")
		assert.Equal(t, "A", determineGrade(90.0), "90% should be grade A")
		assert.Equal(t, "B", determineGrade(85.0), "85% should be grade B")
		assert.Equal(t, "B", determineGrade(80.0), "80% should be grade B")
		assert.Equal(t, "C", determineGrade(75.0), "75% should be grade C")
		assert.Equal(t, "C", determineGrade(70.0), "70% should be grade C")
		assert.Equal(t, "D", determineGrade(65.0), "65% should be grade D")
		assert.Equal(t, "D", determineGrade(60.0), "60% should be grade D")
		assert.Equal(t, "F", determineGrade(55.0), "55% should be grade F")
		assert.Equal(t, "F", determineGrade(0.0), "0% should be grade F")
	})

	t.Run("SeverityDetermination", func(t *testing.T) {
		// High severity cases
		assert.Equal(t, "high", determineSeverity("TestHandler", []string{"@Summary", "@Description", "@Tags"}))
		assert.Equal(t, "high", determineSeverity("TestHandler", []string{"@Summary"}))
		assert.Equal(t, "high", determineSeverity("TestHandler", []string{"@Description"}))

		// Medium severity cases
		assert.Equal(t, "medium", determineSeverity("TestHandler", []string{"@Accept"}))
		assert.Equal(t, "medium", determineSeverity("TestHandler", []string{"@Produce", "@Accept"}))
	})
}

// TestTagDistributionScore tests tag distribution scoring
func TestTagDistributionScore(t *testing.T) {
	t.Run("EmptyDistribution", func(t *testing.T) {
		score := calculateTagDistributionScore(map[string]int{}, 0)
		assert.Equal(t, 0.0, score, "Empty distribution should score 0")
	})

	t.Run("IdealDistribution", func(t *testing.T) {
		distribution := map[string]int{
			"epics": 5, "user-stories": 4, "requirements": 6,
			"acceptance-criteria": 3, "search": 2, "config": 2,
			"comments": 3, "auth": 1,
		}
		score := calculateTagDistributionScore(distribution, 26)
		assert.Equal(t, 100.0, score, "Ideal distribution should score 100%")
	})

	t.Run("PartialDistribution", func(t *testing.T) {
		distribution := map[string]int{
			"epics": 5, "user-stories": 4,
		}
		score := calculateTagDistributionScore(distribution, 9)
		assert.Greater(t, score, 0.0, "Partial distribution should have positive score")
		assert.Less(t, score, 100.0, "Partial distribution should not be perfect")
	})
}

// TestImprovementAreas tests improvement area identification
func TestImprovementAreas(t *testing.T) {
	t.Run("NoImprovementNeeded", func(t *testing.T) {
		componentScores := map[string]float64{
			"annotation_coverage": 95.0,
			"endpoint_coverage":   90.0,
			"example_coverage":    85.0,
			"tag_distribution":    80.0,
		}
		areas := identifyImprovementAreas(componentScores)
		assert.Empty(t, areas, "High scores should not generate improvement areas")
	})

	t.Run("MultipleImprovements", func(t *testing.T) {
		componentScores := map[string]float64{
			"annotation_coverage": 70.0,
			"endpoint_coverage":   75.0,
			"example_coverage":    60.0,
			"tag_distribution":    50.0,
		}
		areas := identifyImprovementAreas(componentScores)
		assert.Len(t, areas, 4, "Low scores should generate improvement areas for all components")
	})

	t.Run("SpecificImprovements", func(t *testing.T) {
		componentScores := map[string]float64{
			"annotation_coverage": 70.0,
			"endpoint_coverage":   90.0,
			"example_coverage":    85.0,
			"tag_distribution":    80.0,
		}
		areas := identifyImprovementAreas(componentScores)
		assert.Len(t, areas, 1, "Should identify only annotation coverage improvement")
		assert.Contains(t, areas[0], "Swagger annotations", "Should mention Swagger annotations")
	})
}

// BenchmarkMetricsGeneration benchmarks the metrics generation performance
func BenchmarkMetricsGeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateDocumentationMetrics()
		if err != nil {
			b.Fatalf("Error generating metrics: %v", err)
		}
	}
}
