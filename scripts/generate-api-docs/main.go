package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// OpenAPI specification structures
type OpenAPISpec struct {
	OpenAPI    string                `yaml:"openapi" json:"openapi"`
	Info       Info                  `yaml:"info" json:"info"`
	Servers    []Server              `yaml:"servers" json:"servers"`
	Paths      map[string]PathItem   `yaml:"paths" json:"paths"`
	Components Components            `yaml:"components" json:"components"`
	Security   []map[string][]string `yaml:"security" json:"security"`
}

type Info struct {
	Title       string  `yaml:"title" json:"title"`
	Description string  `yaml:"description" json:"description"`
	Version     string  `yaml:"version" json:"version"`
	Contact     Contact `yaml:"contact" json:"contact"`
	License     License `yaml:"license" json:"license"`
}

type Contact struct {
	Name  string `yaml:"name" json:"name"`
	URL   string `yaml:"url" json:"url"`
	Email string `yaml:"email" json:"email"`
}

type License struct {
	Name string `yaml:"name" json:"name"`
	URL  string `yaml:"url" json:"url"`
}

type Server struct {
	URL         string `yaml:"url" json:"url"`
	Description string `yaml:"description" json:"description"`
}

type PathItem struct {
	Get    *Operation `yaml:"get,omitempty" json:"get,omitempty"`
	Post   *Operation `yaml:"post,omitempty" json:"post,omitempty"`
	Put    *Operation `yaml:"put,omitempty" json:"put,omitempty"`
	Delete *Operation `yaml:"delete,omitempty" json:"delete,omitempty"`
	Patch  *Operation `yaml:"patch,omitempty" json:"patch,omitempty"`
}

type Operation struct {
	Tags        []string              `yaml:"tags,omitempty" json:"tags,omitempty"`
	Summary     string                `yaml:"summary,omitempty" json:"summary,omitempty"`
	Description string                `yaml:"description,omitempty" json:"description,omitempty"`
	Parameters  []Parameter           `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	RequestBody *RequestBody          `yaml:"requestBody,omitempty" json:"requestBody,omitempty"`
	Responses   map[string]Response   `yaml:"responses,omitempty" json:"responses,omitempty"`
	Security    []map[string][]string `yaml:"security,omitempty" json:"security,omitempty"`
}

type Parameter struct {
	Name        string      `yaml:"name" json:"name"`
	In          string      `yaml:"in" json:"in"`
	Description string      `yaml:"description,omitempty" json:"description,omitempty"`
	Required    bool        `yaml:"required,omitempty" json:"required,omitempty"`
	Schema      interface{} `yaml:"schema,omitempty" json:"schema,omitempty"`
}

type RequestBody struct {
	Description string                     `yaml:"description,omitempty" json:"description,omitempty"`
	Required    bool                       `yaml:"required,omitempty" json:"required,omitempty"`
	Content     map[string]MediaTypeObject `yaml:"content,omitempty" json:"content,omitempty"`
}

type MediaTypeObject struct {
	Schema interface{} `yaml:"schema,omitempty" json:"schema,omitempty"`
}

type Response struct {
	Description string                     `yaml:"description" json:"description"`
	Content     map[string]MediaTypeObject `yaml:"content,omitempty" json:"content,omitempty"`
}

type Components struct {
	Schemas         map[string]interface{} `yaml:"schemas,omitempty" json:"schemas,omitempty"`
	Parameters      map[string]Parameter   `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	Responses       map[string]Response    `yaml:"responses,omitempty" json:"responses,omitempty"`
	SecuritySchemes map[string]interface{} `yaml:"securitySchemes,omitempty" json:"securitySchemes,omitempty"`
}

// Documentation generation structures
type EndpointDoc struct {
	Method      string
	Path        string
	Summary     string
	Description string
	Tags        []string
	Parameters  []Parameter
	RequestBody *RequestBody
	Responses   map[string]Response
	Security    []map[string][]string
}

type TagGroup struct {
	Name      string
	Endpoints []EndpointDoc
}

func main() {
	var (
		inputFile = flag.String("input", "docs/openapi-v3.yaml", "Input OpenAPI specification file")
		outputDir = flag.String("output", "docs/generated", "Output directory for generated documentation")
		format    = flag.String("format", "all", "Output format: html, markdown, typescript, json, all")
		verbose   = flag.Bool("verbose", false, "Enable verbose output")
	)
	flag.Parse()

	if *verbose {
		log.Printf("Generating API documentation from %s to %s", *inputFile, *outputDir)
	}

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Read and parse OpenAPI specification
	spec, err := loadOpenAPISpec(*inputFile)
	if err != nil {
		log.Fatalf("Failed to load OpenAPI specification: %v", err)
	}

	// Generate documentation in requested formats
	switch *format {
	case "html":
		if err := generateHTMLDocs(spec, *outputDir, *verbose); err != nil {
			log.Fatalf("Failed to generate HTML documentation: %v", err)
		}
	case "markdown":
		if err := generateMarkdownDocs(spec, *outputDir, *verbose); err != nil {
			log.Fatalf("Failed to generate Markdown documentation: %v", err)
		}
	case "typescript":
		if err := generateTypeScriptDocs(spec, *outputDir, *verbose); err != nil {
			log.Fatalf("Failed to generate TypeScript documentation: %v", err)
		}
	case "json":
		if err := generateJSONDocs(spec, *outputDir, *verbose); err != nil {
			log.Fatalf("Failed to generate JSON documentation: %v", err)
		}
	case "all":
		if err := generateAllDocs(spec, *outputDir, *verbose); err != nil {
			log.Fatalf("Failed to generate documentation: %v", err)
		}
	default:
		log.Fatalf("Unknown format: %s. Use html, markdown, typescript, json, or all", *format)
	}

	if *verbose {
		log.Printf("Documentation generation completed successfully")
	}
}

func loadOpenAPISpec(filename string) (*OpenAPISpec, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var spec OpenAPISpec
	if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
		if err := yaml.Unmarshal(data, &spec); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
	} else {
		if err := json.Unmarshal(data, &spec); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	}

	return &spec, nil
}

func generateAllDocs(spec *OpenAPISpec, outputDir string, verbose bool) error {
	if err := generateHTMLDocs(spec, outputDir, verbose); err != nil {
		return err
	}
	if err := generateMarkdownDocs(spec, outputDir, verbose); err != nil {
		return err
	}
	if err := generateTypeScriptDocs(spec, outputDir, verbose); err != nil {
		return err
	}
	if err := generateJSONDocs(spec, outputDir, verbose); err != nil {
		return err
	}
	return nil
}

func generateHTMLDocs(spec *OpenAPISpec, outputDir string, verbose bool) error {
	if verbose {
		log.Printf("Generating HTML documentation...")
	}

	endpoints := extractEndpoints(spec)
	tagGroups := groupEndpointsByTag(endpoints)

	htmlTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Info.Title}} - API Documentation</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 20px; background: #f8f9fa; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { border-bottom: 2px solid #e9ecef; padding-bottom: 20px; margin-bottom: 30px; }
        .title { color: #2c3e50; margin: 0 0 10px 0; }
        .description { color: #6c757d; line-height: 1.6; }
        .version { background: #007bff; color: white; padding: 4px 12px; border-radius: 20px; font-size: 0.9em; display: inline-block; margin-top: 10px; }
        .tag-section { margin: 30px 0; }
        .tag-title { color: #495057; border-bottom: 1px solid #dee2e6; padding-bottom: 10px; margin-bottom: 20px; }
        .endpoint { background: #f8f9fa; border: 1px solid #dee2e6; border-radius: 6px; margin: 15px 0; overflow: hidden; }
        .endpoint-header { padding: 15px 20px; background: white; border-bottom: 1px solid #dee2e6; }
        .method { display: inline-block; padding: 4px 12px; border-radius: 4px; font-weight: bold; font-size: 0.9em; margin-right: 10px; }
        .method.get { background: #d4edda; color: #155724; }
        .method.post { background: #d1ecf1; color: #0c5460; }
        .method.put { background: #fff3cd; color: #856404; }
        .method.delete { background: #f8d7da; color: #721c24; }
        .method.patch { background: #e2e3e5; color: #383d41; }
        .path { font-family: 'Monaco', 'Menlo', monospace; font-weight: bold; }
        .summary { margin: 10px 0 5px 0; font-weight: 600; }
        .endpoint-description { color: #6c757d; margin: 5px 0; }
        .endpoint-details { padding: 20px; }
        .parameters, .responses { margin: 15px 0; }
        .param-table, .response-table { width: 100%; border-collapse: collapse; margin: 10px 0; }
        .param-table th, .param-table td, .response-table th, .response-table td { 
            padding: 8px 12px; text-align: left; border-bottom: 1px solid #dee2e6; 
        }
        .param-table th, .response-table th { background: #f8f9fa; font-weight: 600; }
        .required { color: #dc3545; font-weight: bold; }
        .code { font-family: 'Monaco', 'Menlo', monospace; background: #f8f9fa; padding: 2px 6px; border-radius: 3px; }
        .toc { background: #f8f9fa; padding: 20px; border-radius: 6px; margin: 20px 0; }
        .toc ul { list-style: none; padding-left: 0; }
        .toc li { margin: 5px 0; }
        .toc a { text-decoration: none; color: #007bff; }
        .toc a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 class="title">{{.Info.Title}}</h1>
            <p class="description">{{.Info.Description}}</p>
            <span class="version">Version {{.Info.Version}}</span>
        </div>

        <div class="toc">
            <h3>Table of Contents</h3>
            <ul>
                {{range .TagGroups}}
                <li><a href="#{{.Name}}">{{.Name}}</a></li>
                {{end}}
            </ul>
        </div>

        {{range .TagGroups}}
        <div class="tag-section" id="{{.Name}}">
            <h2 class="tag-title">{{.Name}}</h2>
            {{range .Endpoints}}
            <div class="endpoint">
                <div class="endpoint-header">
                    <span class="method {{.Method}}">{{.Method | upper}}</span>
                    <span class="path">{{.Path}}</span>
                    <div class="summary">{{.Summary}}</div>
                    {{if .Description}}<div class="endpoint-description">{{.Description}}</div>{{end}}
                </div>
                <div class="endpoint-details">
                    {{if .Parameters}}
                    <div class="parameters">
                        <h4>Parameters</h4>
                        <table class="param-table">
                            <thead>
                                <tr>
                                    <th>Name</th>
                                    <th>Location</th>
                                    <th>Type</th>
                                    <th>Required</th>
                                    <th>Description</th>
                                </tr>
                            </thead>
                            <tbody>
                                {{range .Parameters}}
                                <tr>
                                    <td class="code">{{.Name}}</td>
                                    <td>{{.In}}</td>
                                    <td>{{.Schema}}</td>
                                    <td>{{if .Required}}<span class="required">Yes</span>{{else}}No{{end}}</td>
                                    <td>{{.Description}}</td>
                                </tr>
                                {{end}}
                            </tbody>
                        </table>
                    </div>
                    {{end}}
                    {{if .Responses}}
                    <div class="responses">
                        <h4>Responses</h4>
                        <table class="response-table">
                            <thead>
                                <tr>
                                    <th>Status Code</th>
                                    <th>Description</th>
                                </tr>
                            </thead>
                            <tbody>
                                {{range $code, $response := .Responses}}
                                <tr>
                                    <td class="code">{{$code}}</td>
                                    <td>{{$response.Description}}</td>
                                </tr>
                                {{end}}
                            </tbody>
                        </table>
                    </div>
                    {{end}}
                </div>
            </div>
            {{end}}
        </div>
        {{end}}
    </div>
</body>
</html>`

	tmpl, err := template.New("html").Funcs(template.FuncMap{
		"upper": strings.ToUpper,
	}).Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	data := struct {
		*OpenAPISpec
		TagGroups []TagGroup
	}{
		OpenAPISpec: spec,
		TagGroups:   tagGroups,
	}

	outputFile := filepath.Join(outputDir, "api-documentation.html")
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create HTML file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}

	if verbose {
		log.Printf("HTML documentation generated: %s", outputFile)
	}

	return nil
}
func generateMarkdownDocs(spec *OpenAPISpec, outputDir string, verbose bool) error {
	if verbose {
		log.Printf("Generating Markdown documentation...")
	}

	endpoints := extractEndpoints(spec)
	tagGroups := groupEndpointsByTag(endpoints)

	markdownTemplate := "# {{.Info.Title}}\n\n" +
		"{{.Info.Description}}\n\n" +
		"**Version:** {{.Info.Version}}\n\n" +
		"## Table of Contents\n\n" +
		"{{range .TagGroups}}" +
		"- [{{.Name}}](#{{.Name | lower | replace \" \" \"-\"}})\n" +
		"{{end}}\n\n" +
		"## Base URLs\n\n" +
		"{{range .Servers}}" +
		"- **{{.Description}}**: {{.URL}}\n" +
		"{{end}}\n\n" +
		"## Authentication\n\n" +
		"This API uses JWT Bearer token authentication. Include the token in the Authorization header:\n\n" +
		"```\nAuthorization: Bearer <your_jwt_token>\n```\n\n" +
		"{{range .TagGroups}}" +
		"## {{.Name}}\n\n" +
		"{{range .Endpoints}}" +
		"### {{.Method | upper}} {{.Path}}\n\n" +
		"{{.Summary}}\n\n" +
		"{{if .Description}}" +
		"{{.Description}}\n" +
		"{{end}}\n\n" +
		"{{if .Parameters}}" +
		"#### Parameters\n\n" +
		"| Name | Location | Type | Required | Description |\n" +
		"|------|----------|------|----------|-------------|\n" +
		"{{range .Parameters}}" +
		"| {{.Name}} | {{.In}} | {{.Schema}} | {{if .Required}}**Yes**{{else}}No{{end}} | {{.Description}} |\n" +
		"{{end}}" +
		"{{end}}\n\n" +
		"{{if .RequestBody}}" +
		"#### Request Body\n\n" +
		"{{.RequestBody.Description}}\n\n" +
		"{{if .RequestBody.Required}}**Required:** Yes{{else}}**Required:** No{{end}}\n" +
		"{{end}}\n\n" +
		"#### Responses\n\n" +
		"| Status Code | Description |\n" +
		"|-------------|-------------|\n" +
		"{{range $code, $response := .Responses}}" +
		"| {{$code}} | {{$response.Description}} |\n" +
		"{{end}}\n\n" +
		"{{if .Security}}" +
		"#### Security\n\n" +
		"{{range .Security}}" +
		"{{range $scheme, $scopes := .}}" +
		"- **{{$scheme}}**: {{if $scopes}}{{join $scopes \", \"}}{{else}}No specific scopes{{end}}\n" +
		"{{end}}" +
		"{{end}}" +
		"{{end}}\n\n" +
		"---\n\n" +
		"{{end}}" +
		"{{end}}\n\n" +
		"## Error Handling\n\n" +
		"### Standard Error Response Format\n\n" +
		"```json\n{\n  \"error\": {\n    \"code\": \"ERROR_CODE\",\n    \"message\": \"Human readable error message\"\n  }\n}\n```\n\n" +
		"### Common Error Codes\n\n" +
		"- **VALIDATION_ERROR**: Request validation failed\n" +
		"- **AUTHENTICATION_REQUIRED**: JWT token required\n" +
		"- **INSUFFICIENT_PERMISSIONS**: User lacks required permissions\n" +
		"- **ENTITY_NOT_FOUND**: Requested entity doesn't exist\n" +
		"- **DELETION_CONFLICT**: Entity has dependencies preventing deletion\n" +
		"- **INTERNAL_ERROR**: Server-side error\n\n" +
		"---\n\n" +
		"*Generated from OpenAPI specification version {{.Info.Version}}*"

	tmpl, err := template.New("markdown").Funcs(template.FuncMap{
		"upper":   strings.ToUpper,
		"lower":   strings.ToLower,
		"replace": strings.ReplaceAll,
		"join":    strings.Join,
	}).Parse(markdownTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse Markdown template: %w", err)
	}

	data := struct {
		*OpenAPISpec
		TagGroups []TagGroup
	}{
		OpenAPISpec: spec,
		TagGroups:   tagGroups,
	}

	outputFile := filepath.Join(outputDir, "api-documentation.md")
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create Markdown file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute Markdown template: %w", err)
	}

	if verbose {
		log.Printf("Markdown documentation generated: %s", outputFile)
	}

	return nil
}

func generateTypeScriptDocs(spec *OpenAPISpec, outputDir string, verbose bool) error {
	if verbose {
		log.Printf("Generating TypeScript documentation...")
	}

	// Generate TypeScript interfaces from OpenAPI schemas
	tsContent := `// Generated TypeScript interfaces from OpenAPI specification
// Version: ` + spec.Info.Version + `
// Generated on: ` + fmt.Sprintf("%s", "auto-generated") + `

/**
 * ` + spec.Info.Title + `
 * ` + spec.Info.Description + `
 */

// Base API Configuration
export interface ApiConfig {
  baseUrl: string;
  apiKey?: string;
  timeout?: number;
}

// Standard API Response wrapper
export interface ApiResponse<T = any> {
  data?: T;
  error?: {
    code: string;
    message: string;
  };
}

// Pagination wrapper for list responses
export interface ListResponse<T> {
  data: T[];
  total_count: number;
  limit: number;
  offset: number;
}

// Error response format
export interface ErrorResponse {
  error: {
    code: string;
    message: string;
  };
}

// Authentication types
export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  expires_at: string;
  user: User;
}

export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
}

// User types
export interface User {
  id: string;
  username: string;
  email: string;
  role: 'Administrator' | 'User' | 'Commenter';
  created_at: string;
  updated_at: string;
}

export interface CreateUserRequest {
  username: string;
  email: string;
  password: string;
  role: 'Administrator' | 'User' | 'Commenter';
}

export interface UpdateUserRequest {
  username?: string;
  email?: string;
  role?: 'Administrator' | 'User' | 'Commenter';
}

// Epic types
export type EpicStatus = 'Backlog' | 'Draft' | 'In Progress' | 'Done' | 'Cancelled';
export type Priority = 1 | 2 | 3 | 4;

export interface Epic {
  id: string;
  reference_id: string;
  title: string;
  description?: string;
  status: EpicStatus;
  priority: Priority;
  creator_id: string;
  assignee_id?: string;
  created_at: string;
  updated_at: string;
  
  // Optional populated fields
  creator?: User;
  assignee?: User;
  user_stories?: UserStory[];
  comments?: Comment[];
}

export interface CreateEpicRequest {
  title: string;
  description?: string;
  priority: Priority;
  creator_id: string;
  assignee_id?: string;
}

export interface UpdateEpicRequest {
  title?: string;
  description?: string;
  priority?: Priority;
  assignee_id?: string;
}

// User Story types
export type UserStoryStatus = 'Backlog' | 'Draft' | 'In Progress' | 'Done' | 'Cancelled';

export interface UserStory {
  id: string;
  reference_id: string;
  title: string;
  description?: string;
  status: UserStoryStatus;
  priority: Priority;
  epic_id: string;
  creator_id: string;
  assignee_id?: string;
  created_at: string;
  updated_at: string;
  
  // Optional populated fields
  epic?: Epic;
  creator?: User;
  assignee?: User;
  acceptance_criteria?: AcceptanceCriteria[];
  requirements?: Requirement[];
  comments?: Comment[];
}

export interface CreateUserStoryRequest {
  title: string;
  description?: string;
  priority: Priority;
  epic_id: string;
  assignee_id?: string;
}

export interface UpdateUserStoryRequest {
  title?: string;
  description?: string;
  priority?: Priority;
  assignee_id?: string;
}

// Acceptance Criteria types
export interface AcceptanceCriteria {
  id: string;
  reference_id: string;
  description: string;
  user_story_id: string;
  author_id: string;
  created_at: string;
  updated_at: string;
  
  // Optional populated fields
  user_story?: UserStory;
  author?: User;
  requirements?: Requirement[];
  comments?: Comment[];
}

export interface CreateAcceptanceCriteriaRequest {
  description: string;
  user_story_id: string;
}

export interface UpdateAcceptanceCriteriaRequest {
  description?: string;
}

// Requirement types
export type RequirementStatus = 'Draft' | 'Active' | 'Obsolete';

export interface Requirement {
  id: string;
  reference_id: string;
  title: string;
  description?: string;
  status: RequirementStatus;
  priority: Priority;
  user_story_id: string;
  acceptance_criteria_id?: string;
  type_id: string;
  creator_id: string;
  assignee_id?: string;
  created_at: string;
  updated_at: string;
  
  // Optional populated fields
  user_story?: UserStory;
  acceptance_criteria?: AcceptanceCriteria;
  type?: RequirementType;
  creator?: User;
  assignee?: User;
  source_relationships?: RequirementRelationship[];
  target_relationships?: RequirementRelationship[];
  comments?: Comment[];
}

export interface CreateRequirementRequest {
  title: string;
  description?: string;
  priority: Priority;
  user_story_id: string;
  acceptance_criteria_id?: string;
  type_id: string;
  assignee_id?: string;
}

export interface UpdateRequirementRequest {
  title?: string;
  description?: string;
  priority?: Priority;
  assignee_id?: string;
}

// Comment types
export type EntityType = 'epic' | 'user_story' | 'acceptance_criteria' | 'requirement';

export interface Comment {
  id: string;
  content: string;
  entity_type: EntityType;
  entity_id: string;
  author_id: string;
  parent_comment_id?: string;
  is_resolved: boolean;
  linked_text?: string;
  text_position_start?: number;
  text_position_end?: number;
  created_at: string;
  updated_at: string;
  
  // Optional populated fields
  author?: User;
  parent_comment?: Comment;
  replies?: Comment[];
}

export interface CreateCommentRequest {
  content: string;
  parent_comment_id?: string;
}

export interface CreateInlineCommentRequest {
  content: string;
  linked_text: string;
  text_position_start: number;
  text_position_end: number;
}

export interface UpdateCommentRequest {
  content: string;
}

// Inline comment validation types
export interface InlineCommentValidationRequest {
  comments: InlineCommentPosition[];
}

export interface InlineCommentPosition {
  comment_id: string;
  text_position_start: number;
  text_position_end: number;
}

export interface ValidationResponse {
  valid: boolean;
  errors: string[];
}

// Configuration types
export interface RequirementType {
  id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface RelationshipType {
  id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

// Steering Document types
export interface SteeringDocument {
  id: string;
  reference_id: string;
  title: string;
  description?: string;
  creator_id: string;
  created_at: string;
  updated_at: string;
  
  // Optional populated fields
  creator?: User;
  epics?: Epic[];
}

export interface CreateSteeringDocumentRequest {
  title: string;
  description?: string;
}

export interface UpdateSteeringDocumentRequest {
  title?: string;
  description?: string;
}

export interface SteeringDocumentFilters {
  creator_id?: string;
  search?: string;
  order_by?: string;
  limit?: number;
  offset?: number;
}

export interface RequirementRelationship {
  id: string;
  source_requirement_id: string;
  target_requirement_id: string;
  relationship_type_id: string;
  created_by: string;
  created_at: string;
  
  // Optional populated fields
  source_requirement?: Requirement;
  target_requirement?: Requirement;
  relationship_type?: RelationshipType;
  creator?: User;
}

export interface CreateRelationshipRequest {
  source_requirement_id: string;
  target_requirement_id: string;
  relationship_type_id: string;
}

// Deletion workflow types
export interface DependencyInfo {
  can_delete: boolean;
  dependencies: DependencyItem[];
  warnings: string[];
}

export interface DependencyItem {
  entity_type: string;
  entity_id: string;
  reference_id: string;
  title: string;
  dependency_type: string;
}

export interface DeletionResult {
  success: boolean;
  deleted_entities: DeletedEntity[];
  message: string;
}

export interface DeletedEntity {
  entity_type: string;
  entity_id: string;
  reference_id: string;
}

// Search types
export interface SearchResult {
  entity_type: EntityType;
  entity_id: string;
  reference_id: string;
  title: string;
  description?: string;
  highlight?: string;
  rank: number;
}

export interface SearchResponse {
  results: SearchResult[];
  total_count: number;
  query: string;
  entity_types: string[];
  limit: number;
  offset: number;
}

export interface SearchSuggestionsResponse {
  titles: string[];
  reference_ids: string[];
  statuses: string[];
}

// Hierarchy types
export interface HierarchyNode {
  entity_type: EntityType;
  entity_id: string;
  reference_id: string;
  title: string;
  status: string;
  children?: HierarchyNode[];
}

export interface EntityPath {
  entity_type: EntityType;
  entity_id: string;
  reference_id: string;
  title: string;
}

// Status management types
export interface StatusChangeRequest {
  status: string;
}

export interface AssignmentRequest {
  assignee_id?: string;
}

// Health check types
export interface HealthCheckResponse {
  status: string;
  reason?: string;
}

// List response types
export interface UserListResponse extends ListResponse<User> {}
export interface EpicListResponse extends ListResponse<Epic> {}
export interface UserStoryListResponse extends ListResponse<UserStory> {}
export interface AcceptanceCriteriaListResponse extends ListResponse<AcceptanceCriteria> {}
export interface RequirementListResponse extends ListResponse<Requirement> {}
export interface CommentListResponse extends ListResponse<Comment> {}
export interface RequirementTypeListResponse extends ListResponse<RequirementType> {}
export interface RelationshipTypeListResponse extends ListResponse<RelationshipType> {}
export interface SteeringDocumentListResponse extends ListResponse<SteeringDocument> {}

// API Client interface
export interface ApiClient {
  // Authentication
  login(credentials: LoginRequest): Promise<LoginResponse>;
  getProfile(): Promise<User>;
  changePassword(request: ChangePasswordRequest): Promise<void>;
  
  // User management (Admin only)
  createUser(user: CreateUserRequest): Promise<User>;
  getUsers(params?: { limit?: number; offset?: number }): Promise<UserListResponse>;
  getUser(id: string): Promise<User>;
  updateUser(id: string, user: UpdateUserRequest): Promise<User>;
  deleteUser(id: string): Promise<void>;
  
  // Epics
  createEpic(epic: CreateEpicRequest): Promise<Epic>;
  getEpics(params?: {
    creator_id?: string;
    assignee_id?: string;
    status?: EpicStatus;
    priority?: Priority;
    limit?: number;
    offset?: number;
    include?: string;
  }): Promise<EpicListResponse>;
  getEpic(id: string): Promise<Epic>;
  updateEpic(id: string, epic: UpdateEpicRequest): Promise<Epic>;
  deleteEpic(id: string): Promise<void>;
  changeEpicStatus(id: string, status: StatusChangeRequest): Promise<Epic>;
  assignEpic(id: string, assignment: AssignmentRequest): Promise<Epic>;
  validateEpicDeletion(id: string): Promise<DependencyInfo>;
  deleteEpicComprehensive(id: string): Promise<DeletionResult>;
  
  // User Stories
  createUserStory(userStory: CreateUserStoryRequest): Promise<UserStory>;
  getUserStories(params?: {
    epic_id?: string;
    creator_id?: string;
    assignee_id?: string;
    status?: UserStoryStatus;
    priority?: Priority;
    limit?: number;
    offset?: number;
    include?: string;
  }): Promise<UserStoryListResponse>;
  getUserStory(id: string): Promise<UserStory>;
  updateUserStory(id: string, userStory: UpdateUserStoryRequest): Promise<UserStory>;
  deleteUserStory(id: string): Promise<void>;
  
  // Requirements
  createRequirement(requirement: CreateRequirementRequest): Promise<Requirement>;
  getRequirements(params?: {
    user_story_id?: string;
    acceptance_criteria_id?: string;
    type_id?: string;
    creator_id?: string;
    assignee_id?: string;
    status?: RequirementStatus;
    priority?: Priority;
    limit?: number;
    offset?: number;
    include?: string;
  }): Promise<RequirementListResponse>;
  getRequirement(id: string): Promise<Requirement>;
  updateRequirement(id: string, requirement: UpdateRequirementRequest): Promise<Requirement>;
  deleteRequirement(id: string): Promise<void>;
  
  // Comments
  getComments(entityType: EntityType, entityId: string, params?: {
    limit?: number;
    offset?: number;
  }): Promise<CommentListResponse>;
  createComment(entityType: EntityType, entityId: string, comment: CreateCommentRequest): Promise<Comment>;
  createInlineComment(entityType: EntityType, entityId: string, comment: CreateInlineCommentRequest): Promise<Comment>;
  updateComment(id: string, comment: UpdateCommentRequest): Promise<Comment>;
  deleteComment(id: string): Promise<void>;
  resolveComment(id: string): Promise<Comment>;
  unresolveComment(id: string): Promise<Comment>;
  
  // Steering Documents
  createSteeringDocument(doc: CreateSteeringDocumentRequest): Promise<SteeringDocument>;
  getSteeringDocuments(params?: SteeringDocumentFilters): Promise<SteeringDocumentListResponse>;
  getSteeringDocument(id: string): Promise<SteeringDocument>;
  updateSteeringDocument(id: string, doc: UpdateSteeringDocumentRequest): Promise<SteeringDocument>;
  deleteSteeringDocument(id: string): Promise<void>;
  getEpicSteeringDocuments(epicId: string): Promise<SteeringDocument[]>;
  linkSteeringDocumentToEpic(epicId: string, docId: string): Promise<void>;
  unlinkSteeringDocumentFromEpic(epicId: string, docId: string): Promise<void>;
  
  // Search
  search(params: {
    q: string;
    entity_types?: string;
    limit?: number;
    offset?: number;
  }): Promise<SearchResponse>;
  getSearchSuggestions(params: {
    query: string;
    limit?: number;
  }): Promise<SearchSuggestionsResponse>;
  
  // Health checks
  readinessCheck(): Promise<HealthCheckResponse>;
  livenessCheck(): Promise<HealthCheckResponse>;
}

// HTTP client configuration
export interface HttpClientConfig {
  baseURL: string;
  timeout?: number;
  headers?: Record<string, string>;
  interceptors?: {
    request?: (config: any) => any;
    response?: (response: any) => any;
    error?: (error: any) => any;
  };
}

// Utility types for API parameters
export type QueryParams = Record<string, string | number | boolean | undefined>;
export type PathParams = Record<string, string>;
export type RequestHeaders = Record<string, string>;

// API endpoint definitions
export const API_ENDPOINTS = {
  // Authentication
  LOGIN: '/auth/login',
  PROFILE: '/auth/profile',
  CHANGE_PASSWORD: '/auth/change-password',
  
  // User management
  USERS: '/auth/users',
  USER: '/auth/users/{id}',
  
  // Epics
  EPICS: '/api/v1/epics',
  EPIC: '/api/v1/epics/{id}',
  EPIC_USER_STORIES: '/api/v1/epics/{id}/user-stories',
  EPIC_STATUS: '/api/v1/epics/{id}/status',
  EPIC_ASSIGN: '/api/v1/epics/{id}/assign',
  EPIC_VALIDATE_DELETION: '/api/v1/epics/{id}/validate-deletion',
  EPIC_DELETE: '/api/v1/epics/{id}/delete',
  EPIC_COMMENTS: '/api/v1/epics/{id}/comments',
  
  // User Stories
  USER_STORIES: '/api/v1/user-stories',
  USER_STORY: '/api/v1/user-stories/{id}',
  USER_STORY_ACCEPTANCE_CRITERIA: '/api/v1/user-stories/{id}/acceptance-criteria',
  USER_STORY_REQUIREMENTS: '/api/v1/user-stories/{id}/requirements',
  USER_STORY_STATUS: '/api/v1/user-stories/{id}/status',
  USER_STORY_ASSIGN: '/api/v1/user-stories/{id}/assign',
  
  // Requirements
  REQUIREMENTS: '/api/v1/requirements',
  REQUIREMENT: '/api/v1/requirements/{id}',
  REQUIREMENT_RELATIONSHIPS: '/api/v1/requirements/{id}/relationships',
  REQUIREMENT_STATUS: '/api/v1/requirements/{id}/status',
  REQUIREMENT_ASSIGN: '/api/v1/requirements/{id}/assign',
  
  // Search
  SEARCH: '/api/v1/search',
  SEARCH_SUGGESTIONS: '/api/v1/search/suggestions',
  
  // Steering Documents
  STEERING_DOCUMENTS: '/api/v1/steering-documents',
  STEERING_DOCUMENT: '/api/v1/steering-documents/{id}',
  EPIC_STEERING_DOCUMENTS: '/api/v1/epics/{id}/steering-documents',
  EPIC_STEERING_DOCUMENT_LINK: '/api/v1/epics/{epic_id}/steering-documents/{doc_id}',
  
  // Health
  READY: '/ready',
  LIVE: '/live',
} as const;

export type ApiEndpoint = typeof API_ENDPOINTS[keyof typeof API_ENDPOINTS];
`

	outputFile := filepath.Join(outputDir, "api-types.ts")
	if err := ioutil.WriteFile(outputFile, []byte(tsContent), 0644); err != nil {
		return fmt.Errorf("failed to write TypeScript file: %w", err)
	}

	if verbose {
		log.Printf("TypeScript documentation generated: %s", outputFile)
	}

	return nil
}

func generateJSONDocs(spec *OpenAPISpec, outputDir string, verbose bool) error {
	if verbose {
		log.Printf("Generating JSON documentation...")
	}

	endpoints := extractEndpoints(spec)
	tagGroups := groupEndpointsByTag(endpoints)

	// Create comprehensive JSON documentation
	jsonDoc := map[string]interface{}{
		"info":             spec.Info,
		"servers":          spec.Servers,
		"tag_groups":       tagGroups,
		"endpoints":        endpoints,
		"schemas":          spec.Components.Schemas,
		"security_schemes": spec.Components.SecuritySchemes,
		"metadata": map[string]interface{}{
			"generated_at":    fmt.Sprintf("%s", "auto-generated"),
			"total_endpoints": len(endpoints),
			"total_tags":      len(tagGroups),
			"openapi_version": spec.OpenAPI,
		},
	}

	jsonData, err := json.MarshalIndent(jsonDoc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	outputFile := filepath.Join(outputDir, "api-documentation.json")
	if err := ioutil.WriteFile(outputFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	if verbose {
		log.Printf("JSON documentation generated: %s", outputFile)
	}

	return nil
}

func extractEndpoints(spec *OpenAPISpec) []EndpointDoc {
	var endpoints []EndpointDoc

	for path, pathItem := range spec.Paths {
		if pathItem.Get != nil {
			endpoints = append(endpoints, EndpointDoc{
				Method:      "get",
				Path:        path,
				Summary:     pathItem.Get.Summary,
				Description: pathItem.Get.Description,
				Tags:        pathItem.Get.Tags,
				Parameters:  pathItem.Get.Parameters,
				RequestBody: pathItem.Get.RequestBody,
				Responses:   pathItem.Get.Responses,
				Security:    pathItem.Get.Security,
			})
		}
		if pathItem.Post != nil {
			endpoints = append(endpoints, EndpointDoc{
				Method:      "post",
				Path:        path,
				Summary:     pathItem.Post.Summary,
				Description: pathItem.Post.Description,
				Tags:        pathItem.Post.Tags,
				Parameters:  pathItem.Post.Parameters,
				RequestBody: pathItem.Post.RequestBody,
				Responses:   pathItem.Post.Responses,
				Security:    pathItem.Post.Security,
			})
		}
		if pathItem.Put != nil {
			endpoints = append(endpoints, EndpointDoc{
				Method:      "put",
				Path:        path,
				Summary:     pathItem.Put.Summary,
				Description: pathItem.Put.Description,
				Tags:        pathItem.Put.Tags,
				Parameters:  pathItem.Put.Parameters,
				RequestBody: pathItem.Put.RequestBody,
				Responses:   pathItem.Put.Responses,
				Security:    pathItem.Put.Security,
			})
		}
		if pathItem.Delete != nil {
			endpoints = append(endpoints, EndpointDoc{
				Method:      "delete",
				Path:        path,
				Summary:     pathItem.Delete.Summary,
				Description: pathItem.Delete.Description,
				Tags:        pathItem.Delete.Tags,
				Parameters:  pathItem.Delete.Parameters,
				RequestBody: pathItem.Delete.RequestBody,
				Responses:   pathItem.Delete.Responses,
				Security:    pathItem.Delete.Security,
			})
		}
		if pathItem.Patch != nil {
			endpoints = append(endpoints, EndpointDoc{
				Method:      "patch",
				Path:        path,
				Summary:     pathItem.Patch.Summary,
				Description: pathItem.Patch.Description,
				Tags:        pathItem.Patch.Tags,
				Parameters:  pathItem.Patch.Parameters,
				RequestBody: pathItem.Patch.RequestBody,
				Responses:   pathItem.Patch.Responses,
				Security:    pathItem.Patch.Security,
			})
		}
	}

	return endpoints
}

func groupEndpointsByTag(endpoints []EndpointDoc) []TagGroup {
	tagMap := make(map[string][]EndpointDoc)

	for _, endpoint := range endpoints {
		if len(endpoint.Tags) == 0 {
			tagMap["Untagged"] = append(tagMap["Untagged"], endpoint)
		} else {
			for _, tag := range endpoint.Tags {
				tagMap[tag] = append(tagMap[tag], endpoint)
			}
		}
	}

	var tagGroups []TagGroup
	var tagNames []string
	for tagName := range tagMap {
		tagNames = append(tagNames, tagName)
	}
	sort.Strings(tagNames)

	for _, tagName := range tagNames {
		tagGroups = append(tagGroups, TagGroup{
			Name:      tagName,
			Endpoints: tagMap[tagName],
		})
	}

	return tagGroups
}
