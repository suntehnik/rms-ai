package docs

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// DocumentationMetrics represents comprehensive documentation quality metrics
type DocumentationMetrics struct {
	GeneratedAt        time.Time                 `json:"generated_at"`
	AnnotationCoverage AnnotationCoverageMetrics `json:"annotation_coverage"`
	EndpointCoverage   EndpointCoverageMetrics   `json:"endpoint_coverage"`
	ExampleCoverage    ExampleCoverageMetrics    `json:"example_coverage"`
	QualityScore       QualityScoreMetrics       `json:"quality_score"`
	TagDistribution    map[string]int            `json:"tag_distribution"`
	MissingAnnotations []MissingAnnotationInfo   `json:"missing_annotations"`
	Recommendations    []string                  `json:"recommendations"`
}

// AnnotationCoverageMetrics tracks Swagger annotation coverage
type AnnotationCoverageMetrics struct {
	TotalHandlers      int     `json:"total_handlers"`
	AnnotatedHandlers  int     `json:"annotated_handlers"`
	CoveragePercentage float64 `json:"coverage_percentage"`
	MissingCount       int     `json:"missing_count"`
}

// EndpointCoverageMetrics tracks API endpoint documentation coverage
type EndpointCoverageMetrics struct {
	TotalEndpoints      int            `json:"total_endpoints"`
	DocumentedEndpoints int            `json:"documented_endpoints"`
	CoveragePercentage  float64        `json:"coverage_percentage"`
	EndpointsByMethod   map[string]int `json:"endpoints_by_method"`
	EndpointsByTag      map[string]int `json:"endpoints_by_tag"`
}

// ExampleCoverageMetrics tracks example documentation coverage
type ExampleCoverageMetrics struct {
	HandlersWithExamples   int     `json:"handlers_with_examples"`
	ExampleCoveragePercent float64 `json:"example_coverage_percent"`
	RequestBodyExamples    int     `json:"request_body_examples"`
	ResponseExamples       int     `json:"response_examples"`
	QueryParameterExamples int     `json:"query_parameter_examples"`
}

// QualityScoreMetrics provides overall quality assessment
type QualityScoreMetrics struct {
	OverallScore     float64            `json:"overall_score"`
	ComponentScores  map[string]float64 `json:"component_scores"`
	Grade            string             `json:"grade"`
	ImprovementAreas []string           `json:"improvement_areas"`
}

// MissingAnnotationInfo details missing annotations for specific handlers
type MissingAnnotationInfo struct {
	File          string   `json:"file"`
	Handler       string   `json:"handler"`
	MissingFields []string `json:"missing_fields"`
	Severity      string   `json:"severity"`
}

// GenerateDocumentationMetrics analyzes the codebase and generates comprehensive metrics
func GenerateDocumentationMetrics() (*DocumentationMetrics, error) {
	metrics := &DocumentationMetrics{
		GeneratedAt:        time.Now(),
		TagDistribution:    make(map[string]int),
		MissingAnnotations: []MissingAnnotationInfo{},
		Recommendations:    []string{},
	}

	// Analyze handler files for annotation coverage
	if err := analyzeAnnotationCoverage(metrics); err != nil {
		return nil, fmt.Errorf("failed to analyze annotation coverage: %w", err)
	}

	// Analyze OpenAPI specification for endpoint coverage
	if err := analyzeEndpointCoverage(metrics); err != nil {
		return nil, fmt.Errorf("failed to analyze endpoint coverage: %w", err)
	}

	// Analyze example coverage
	if err := analyzeExampleCoverage(metrics); err != nil {
		return nil, fmt.Errorf("failed to analyze example coverage: %w", err)
	}

	// Calculate quality scores
	calculateQualityScores(metrics)

	// Generate recommendations
	generateRecommendations(metrics)

	return metrics, nil
}

// analyzeAnnotationCoverage analyzes Swagger annotation coverage in handler files
func analyzeAnnotationCoverage(metrics *DocumentationMetrics) error {
	handlerDir := findHandlersDirectory()

	handlerFiles, err := getGoFilesRecursive(handlerDir)
	if err != nil {
		return err
	}

	var totalHandlers, annotatedHandlers int
	var missingAnnotations []MissingAnnotationInfo

	for _, file := range handlerFiles {
		handlers, err := extractHandlerFunctionsWithDetails(file)
		if err != nil {
			continue // Skip files that can't be parsed
		}

		for _, handler := range handlers {
			totalHandlers++

			hasAnnotations, missing := checkSwaggerAnnotationsDetailed(file, handler)
			if hasAnnotations {
				annotatedHandlers++

				// Extract tags for distribution analysis
				extractTagsFromHandler(handler, metrics.TagDistribution)
			} else {
				severity := determineSeverity(handler.Name, missing)
				missingAnnotations = append(missingAnnotations, MissingAnnotationInfo{
					File:          filepath.Base(file),
					Handler:       handler.Name,
					MissingFields: missing,
					Severity:      severity,
				})
			}
		}
	}

	metrics.AnnotationCoverage = AnnotationCoverageMetrics{
		TotalHandlers:      totalHandlers,
		AnnotatedHandlers:  annotatedHandlers,
		CoveragePercentage: calculatePercentage(annotatedHandlers, totalHandlers),
		MissingCount:       len(missingAnnotations),
	}

	metrics.MissingAnnotations = missingAnnotations

	return nil
}

// analyzeEndpointCoverage analyzes endpoint coverage from OpenAPI specification
func analyzeEndpointCoverage(metrics *DocumentationMetrics) error {
	specPath := findSwaggerSpecPath()

	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		// If swagger.json doesn't exist, set zero values
		metrics.EndpointCoverage = EndpointCoverageMetrics{
			EndpointsByMethod: make(map[string]int),
			EndpointsByTag:    make(map[string]int),
		}
		return nil
	}

	specData, err := os.ReadFile(specPath)
	if err != nil {
		return err
	}

	var spec map[string]interface{}
	if err := json.Unmarshal(specData, &spec); err != nil {
		return err
	}

	endpointsByMethod := make(map[string]int)
	endpointsByTag := make(map[string]int)
	totalEndpoints := 0
	documentedEndpoints := 0

	if paths, exists := spec["paths"]; exists {
		pathsObj := paths.(map[string]interface{})

		for _, pathValue := range pathsObj {
			pathObj := pathValue.(map[string]interface{})

			for method, methodValue := range pathObj {
				if isHTTPMethodLower(method) {
					totalEndpoints++
					endpointsByMethod[strings.ToUpper(method)]++

					if methodObj, ok := methodValue.(map[string]interface{}); ok {
						// Check if endpoint is properly documented
						if hasRequiredDocumentation(methodObj) {
							documentedEndpoints++
						}

						// Extract tags
						if tags, exists := methodObj["tags"]; exists {
							if tagsArray, ok := tags.([]interface{}); ok {
								for _, tag := range tagsArray {
									if tagStr, ok := tag.(string); ok {
										endpointsByTag[tagStr]++
									}
								}
							}
						}
					}
				}
			}
		}
	}

	metrics.EndpointCoverage = EndpointCoverageMetrics{
		TotalEndpoints:      totalEndpoints,
		DocumentedEndpoints: documentedEndpoints,
		CoveragePercentage:  calculatePercentage(documentedEndpoints, totalEndpoints),
		EndpointsByMethod:   endpointsByMethod,
		EndpointsByTag:      endpointsByTag,
	}

	return nil
}

// analyzeExampleCoverage analyzes example coverage in documentation
func analyzeExampleCoverage(metrics *DocumentationMetrics) error {
	// Count handlers with examples
	handlerDir := findHandlersDirectory()
	handlerFiles, err := getGoFilesRecursive(handlerDir)
	if err != nil {
		return err
	}

	var handlersWithExamples int
	totalHandlers := metrics.AnnotationCoverage.TotalHandlers

	for _, file := range handlerFiles {
		handlers, err := extractHandlerFunctionsWithDetails(file)
		if err != nil {
			continue
		}

		for _, handler := range handlers {
			if hasExamples(handler) {
				handlersWithExamples++
			}
		}
	}

	// Count example data
	examples := GetExampleData()
	requestBodies := GetExampleRequestBodies()
	queryParams := GetExampleQueryParameters()

	metrics.ExampleCoverage = ExampleCoverageMetrics{
		HandlersWithExamples:   handlersWithExamples,
		ExampleCoveragePercent: calculatePercentage(handlersWithExamples, totalHandlers),
		RequestBodyExamples:    len(requestBodies),
		ResponseExamples:       countResponseExamples(examples),
		QueryParameterExamples: len(queryParams),
	}

	return nil
}

// calculateQualityScores calculates overall quality scores
func calculateQualityScores(metrics *DocumentationMetrics) {
	componentScores := make(map[string]float64)

	// Annotation coverage score (40% weight)
	componentScores["annotation_coverage"] = metrics.AnnotationCoverage.CoveragePercentage

	// Endpoint coverage score (30% weight)
	componentScores["endpoint_coverage"] = metrics.EndpointCoverage.CoveragePercentage

	// Example coverage score (20% weight)
	componentScores["example_coverage"] = metrics.ExampleCoverage.ExampleCoveragePercent

	// Tag distribution score (10% weight) - based on how well endpoints are categorized
	tagScore := calculateTagDistributionScore(metrics.TagDistribution, metrics.EndpointCoverage.TotalEndpoints)
	componentScores["tag_distribution"] = tagScore

	// Calculate weighted overall score
	overallScore := (componentScores["annotation_coverage"]*0.4 +
		componentScores["endpoint_coverage"]*0.3 +
		componentScores["example_coverage"]*0.2 +
		componentScores["tag_distribution"]*0.1)

	// Determine grade
	grade := determineGrade(overallScore)

	// Identify improvement areas
	improvementAreas := identifyImprovementAreas(componentScores)

	metrics.QualityScore = QualityScoreMetrics{
		OverallScore:     overallScore,
		ComponentScores:  componentScores,
		Grade:            grade,
		ImprovementAreas: improvementAreas,
	}
}

// generateRecommendations generates actionable recommendations
func generateRecommendations(metrics *DocumentationMetrics) {
	var recommendations []string

	// Annotation coverage recommendations
	if metrics.AnnotationCoverage.CoveragePercentage < 90 {
		recommendations = append(recommendations,
			fmt.Sprintf("Improve annotation coverage from %.1f%% to 90%% by adding Swagger annotations to %d handlers",
				metrics.AnnotationCoverage.CoveragePercentage,
				metrics.AnnotationCoverage.MissingCount))
	}

	// Endpoint coverage recommendations
	if metrics.EndpointCoverage.CoveragePercentage < 95 {
		recommendations = append(recommendations,
			"Ensure all API endpoints have complete documentation including summary, description, and response schemas")
	}

	// Example coverage recommendations
	if metrics.ExampleCoverage.ExampleCoveragePercent < 80 {
		recommendations = append(recommendations,
			"Add more examples to handler documentation to improve developer experience")
	}

	// Tag distribution recommendations
	if len(metrics.TagDistribution) < 5 {
		recommendations = append(recommendations,
			"Consider organizing endpoints into more logical tag groups for better API navigation")
	}

	// High-priority missing annotations
	highPriorityMissing := 0
	for _, missing := range metrics.MissingAnnotations {
		if missing.Severity == "high" {
			highPriorityMissing++
		}
	}

	if highPriorityMissing > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Address %d high-priority missing annotations first", highPriorityMissing))
	}

	metrics.Recommendations = recommendations
}

// Helper functions

func getGoFilesRecursive(dir string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

type HandlerFunctionDetailed struct {
	Name     string
	Comments []string
	IsPublic bool
}

func extractHandlerFunctionsWithDetails(filename string) ([]HandlerFunctionDetailed, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var handlers []HandlerFunctionDetailed

	for _, decl := range node.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			if isHandlerFunctionDetailed(fn) {
				handler := HandlerFunctionDetailed{
					Name:     fn.Name.Name,
					IsPublic: fn.Name.IsExported(),
				}

				if fn.Doc != nil {
					for _, comment := range fn.Doc.List {
						handler.Comments = append(handler.Comments, comment.Text)
					}
				}

				handlers = append(handlers, handler)
			}
		}
	}

	return handlers, nil
}

func isHandlerFunctionDetailed(fn *ast.FuncDecl) bool {
	if fn.Type.Params == nil || len(fn.Type.Params.List) == 0 {
		return false
	}

	for _, param := range fn.Type.Params.List {
		if starExpr, ok := param.Type.(*ast.StarExpr); ok {
			if selectorExpr, ok := starExpr.X.(*ast.SelectorExpr); ok {
				if ident, ok := selectorExpr.X.(*ast.Ident); ok {
					if ident.Name == "gin" && selectorExpr.Sel.Name == "Context" {
						return true
					}
				}
			}
		}
	}

	return false
}

func checkSwaggerAnnotationsDetailed(filename string, handler HandlerFunctionDetailed) (bool, []string) {
	var missing []string

	commentText := strings.Join(handler.Comments, "\n")

	requiredAnnotations := []string{
		"@Summary",
		"@Description",
		"@Tags",
		"@Accept",
		"@Produce",
	}

	for _, annotation := range requiredAnnotations {
		if !strings.Contains(commentText, annotation) {
			missing = append(missing, annotation)
		}
	}

	return len(missing) == 0, missing
}

func extractTagsFromHandler(handler HandlerFunctionDetailed, tagDistribution map[string]int) {
	commentText := strings.Join(handler.Comments, "\n")

	tagRegex := regexp.MustCompile(`@Tags\s+([^\n\r]+)`)
	matches := tagRegex.FindAllStringSubmatch(commentText, -1)

	for _, match := range matches {
		if len(match) > 1 {
			tags := strings.Split(strings.TrimSpace(match[1]), ",")
			for _, tag := range tags {
				tag = strings.TrimSpace(tag)
				tagDistribution[tag]++
			}
		}
	}
}

func determineSeverity(handlerName string, missing []string) string {
	// Public handlers with many missing annotations are high priority
	if len(missing) >= 3 {
		return "high"
	}

	// Handlers missing critical annotations
	for _, annotation := range missing {
		if annotation == "@Summary" || annotation == "@Description" {
			return "high"
		}
	}

	return "medium"
}

func isHTTPMethodLower(method string) bool {
	httpMethods := []string{"get", "post", "put", "patch", "delete", "head", "options"}
	method = strings.ToLower(method)

	for _, m := range httpMethods {
		if method == m {
			return true
		}
	}

	return false
}

func hasRequiredDocumentation(methodObj map[string]interface{}) bool {
	requiredFields := []string{"summary", "responses"}

	for _, field := range requiredFields {
		if _, exists := methodObj[field]; !exists {
			return false
		}
	}

	return true
}

func hasExamples(handler HandlerFunctionDetailed) bool {
	commentText := strings.Join(handler.Comments, "\n")
	return strings.Contains(strings.ToLower(commentText), "example")
}

func countResponseExamples(examples *ExampleData) int {
	// Count non-nil example entities
	count := 0
	if examples.Epic.ID != "" {
		count++
	}
	if examples.UserStory.ID != "" {
		count++
	}
	if examples.AcceptanceCriteria.ID != "" {
		count++
	}
	if examples.Requirement.ID != "" {
		count++
	}
	if examples.Comment.ID != "" {
		count++
	}
	return count
}

func calculatePercentage(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator) * 100
}

func calculateTagDistributionScore(tagDistribution map[string]int, totalEndpoints int) float64 {
	if len(tagDistribution) == 0 || totalEndpoints == 0 {
		return 0
	}

	// Score based on how well endpoints are distributed across tags
	// More tags with balanced distribution = higher score
	idealTagCount := 8 // Ideal number of tag categories
	tagCount := len(tagDistribution)

	// Base score from tag count (up to ideal)
	tagCountScore := float64(tagCount) / float64(idealTagCount) * 100
	if tagCountScore > 100 {
		tagCountScore = 100
	}

	return tagCountScore
}

func determineGrade(score float64) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}

func identifyImprovementAreas(componentScores map[string]float64) []string {
	var areas []string

	for component, score := range componentScores {
		if score < 80 {
			switch component {
			case "annotation_coverage":
				areas = append(areas, "Add missing Swagger annotations to handlers")
			case "endpoint_coverage":
				areas = append(areas, "Complete endpoint documentation in OpenAPI spec")
			case "example_coverage":
				areas = append(areas, "Add more examples to improve developer experience")
			case "tag_distribution":
				areas = append(areas, "Organize endpoints into logical tag groups")
			}
		}
	}

	return areas
}

// SaveMetricsToFile saves metrics to a JSON file
func SaveMetricsToFile(metrics *DocumentationMetrics, filename string) error {
	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// LoadMetricsFromFile loads metrics from a JSON file
func LoadMetricsFromFile(filename string) (*DocumentationMetrics, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var metrics DocumentationMetrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return nil, err
	}

	return &metrics, nil
}

// findHandlersDirectory finds the correct path to the handlers directory
func findHandlersDirectory() string {
	// Try different possible paths
	possiblePaths := []string{
		"internal/handlers",       // From project root
		"../../internal/handlers", // From internal/docs (test context)
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Default fallback
	return "internal/handlers"
}

// findSwaggerSpecPath finds the correct path to the swagger specification
func findSwaggerSpecPath() string {
	// Try different possible paths
	possiblePaths := []string{
		"docs/swagger.json",       // From project root
		"../../docs/swagger.json", // From internal/docs (test context)
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Default fallback
	return "docs/swagger.json"
}
