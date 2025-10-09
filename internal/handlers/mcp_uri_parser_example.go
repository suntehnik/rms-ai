package handlers

import (
	"fmt"
)

// Example demonstrates how to use the URI parser in MCP resource handling
func ExampleURIParserUsage() {
	parser := NewURIParser()

	// Example URIs that would be received in MCP resources/read requests
	exampleURIs := []string{
		"epic://EP-001",
		"epic://EP-001/hierarchy",
		"user-story://US-042/requirements",
		"requirement://REQ-123/relationships",
		"acceptance-criteria://AC-005",
	}

	fmt.Println("MCP URI Parser Examples:")
	fmt.Println("========================")

	for _, uri := range exampleURIs {
		parsed, err := parser.Parse(uri)
		if err != nil {
			fmt.Printf("❌ Failed to parse URI '%s': %v\n", uri, err)
			continue
		}

		fmt.Printf("✅ URI: %s\n", uri)
		fmt.Printf("   Scheme: %s\n", parsed.Scheme)
		fmt.Printf("   Reference ID: %s\n", parsed.ReferenceID)
		if parsed.SubPath != "" {
			fmt.Printf("   Sub-path: %s\n", parsed.SubPath)
		}
		if len(parsed.Parameters) > 0 {
			fmt.Printf("   Parameters: %v\n", parsed.Parameters)
		}
		fmt.Println()
	}

	// Demonstrate scheme-to-prefix validation
	fmt.Println("Scheme-to-Prefix Validation Examples:")
	fmt.Println("====================================")

	validationExamples := []struct {
		uri   string
		valid bool
	}{
		{"epic://EP-001", true},
		{"epic://US-001", false}, // Wrong prefix for epic scheme
		{"user-story://US-001", true},
		{"user-story://EP-001", false}, // Wrong prefix for user-story scheme
		{"requirement://REQ-001", true},
		{"acceptance-criteria://AC-001", true},
	}

	for _, example := range validationExamples {
		_, err := parser.Parse(example.uri)
		status := "✅"
		if err != nil {
			status = "❌"
		}
		fmt.Printf("%s %s (expected: %v)\n", status, example.uri, example.valid)
	}

	// Demonstrate sub-path support
	fmt.Println("\nSupported Sub-paths:")
	fmt.Println("===================")

	subPathExamples := map[string][]string{
		"epic":                {"hierarchy", "user-stories"},
		"user-story":          {"requirements", "acceptance-criteria"},
		"requirement":         {"relationships"},
		"acceptance-criteria": {}, // No sub-paths supported
	}

	for scheme, subPaths := range subPathExamples {
		fmt.Printf("%s: ", scheme)
		if len(subPaths) == 0 {
			fmt.Println("(no sub-paths supported)")
		} else {
			fmt.Println(subPaths)
		}
	}
}

// ResourceHandlerExample shows how the URI parser would be integrated
// into the actual MCP resource handler
func ResourceHandlerExample(uri string) (*ParsedURI, error) {
	parser := NewURIParser()

	// Parse the URI from the MCP resources/read request
	parsed, err := parser.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid resource URI: %w", err)
	}

	// Validate sub-path if present
	if parsed.SubPath != "" {
		if !parser.IsSubPathSupported(parsed.Scheme, parsed.SubPath) {
			return nil, fmt.Errorf("unsupported sub-path '%s' for scheme '%s'",
				parsed.SubPath, parsed.Scheme)
		}
	}

	// At this point, the parsed URI can be used to:
	// 1. Route to the appropriate service based on scheme
	// 2. Fetch the entity using the reference ID
	// 3. Handle sub-path specific logic (hierarchy, relationships, etc.)
	// 4. Apply any query parameters for filtering/formatting

	return parsed, nil
}
