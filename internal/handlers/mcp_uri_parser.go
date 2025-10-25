package handlers

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// URISchemes defines the supported URI schemes for MCP resources
const (
	EpicURIScheme               = "epic"
	UserStoryURIScheme          = "user-story"
	RequirementURIScheme        = "requirement"
	AcceptanceCriteriaURIScheme = "acceptance-criteria"
	PromptURIScheme             = "prompt"
)

// URIParser handles parsing and validation of MCP resource URIs
type URIParser struct {
	referenceIDPattern *regexp.Regexp
	schemePrefixMap    map[string]string
}

// ParsedURI represents a parsed MCP resource URI
type ParsedURI struct {
	Scheme      string            `json:"scheme"`
	ReferenceID string            `json:"reference_id"`
	SubPath     string            `json:"sub_path,omitempty"`
	Parameters  map[string]string `json:"parameters,omitempty"`
}

// NewURIParser creates a new URI parser instance
func NewURIParser() *URIParser {
	return &URIParser{
		referenceIDPattern: regexp.MustCompile(`^(EP|US|REQ|AC|PROMPT)-\d+$`),
		schemePrefixMap: map[string]string{
			EpicURIScheme:               "EP",
			UserStoryURIScheme:          "US",
			RequirementURIScheme:        "REQ",
			AcceptanceCriteriaURIScheme: "AC",
			PromptURIScheme:             "PROMPT",
		},
	}
}

// Parse parses a URI string and returns a ParsedURI structure
func (p *URIParser) Parse(uri string) (*ParsedURI, error) {
	if uri == "" {
		return nil, fmt.Errorf("URI cannot be empty")
	}

	// Check for case sensitivity issues before parsing
	// Go's url.Parse automatically converts scheme to lowercase, but we want to reject non-lowercase schemes
	if schemeEnd := strings.Index(uri, "://"); schemeEnd != -1 {
		originalScheme := uri[:schemeEnd]
		if originalScheme != strings.ToLower(originalScheme) {
			return nil, fmt.Errorf("scheme must be lowercase: %s", originalScheme)
		}
	}

	// Parse the URI using Go's url package
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid URI format: %w", err)
	}

	// Validate scheme
	scheme := parsedURL.Scheme
	if !p.isValidScheme(scheme) {
		return nil, fmt.Errorf("unsupported URI scheme: %s", scheme)
	}

	// Extract reference ID from the host part
	referenceID := parsedURL.Host
	if referenceID == "" {
		return nil, fmt.Errorf("missing reference ID in URI")
	}

	// Validate reference ID format
	if !p.isValidReferenceID(referenceID) {
		return nil, fmt.Errorf("invalid reference ID format: %s", referenceID)
	}

	// Validate scheme and reference ID match
	if err := p.validateSchemeAndReferenceID(scheme, referenceID); err != nil {
		return nil, err
	}

	// Extract sub-path (remove leading slash)
	subPath := strings.TrimPrefix(parsedURL.Path, "/")

	// Parse query parameters
	parameters := make(map[string]string)
	for key, values := range parsedURL.Query() {
		if len(values) > 0 {
			parameters[key] = values[0] // Take first value if multiple
		}
	}

	return &ParsedURI{
		Scheme:      scheme,
		ReferenceID: referenceID,
		SubPath:     subPath,
		Parameters:  parameters,
	}, nil
}

// isValidScheme checks if the scheme is supported
func (p *URIParser) isValidScheme(scheme string) bool {
	_, exists := p.schemePrefixMap[scheme]
	return exists
}

// isValidReferenceID checks if the reference ID matches the expected pattern
func (p *URIParser) isValidReferenceID(id string) bool {
	return p.referenceIDPattern.MatchString(id)
}

// validateSchemeAndReferenceID ensures the reference ID prefix matches the scheme
func (p *URIParser) validateSchemeAndReferenceID(scheme, referenceID string) error {
	expectedPrefix, exists := p.schemePrefixMap[scheme]
	if !exists {
		return fmt.Errorf("unsupported scheme: %s", scheme)
	}

	if !strings.HasPrefix(referenceID, expectedPrefix+"-") {
		return fmt.Errorf("reference ID %s does not match scheme %s (expected prefix: %s)",
			referenceID, scheme, expectedPrefix)
	}

	return nil
}

// GetSupportedSchemes returns a list of all supported URI schemes
func (p *URIParser) GetSupportedSchemes() []string {
	schemes := make([]string, 0, len(p.schemePrefixMap))
	for scheme := range p.schemePrefixMap {
		schemes = append(schemes, scheme)
	}
	return schemes
}

// GetExpectedPrefix returns the expected reference ID prefix for a given scheme
func (p *URIParser) GetExpectedPrefix(scheme string) (string, error) {
	prefix, exists := p.schemePrefixMap[scheme]
	if !exists {
		return "", fmt.Errorf("unsupported scheme: %s", scheme)
	}
	return prefix, nil
}

// IsSubPathSupported checks if a sub-path is supported for the given scheme
func (p *URIParser) IsSubPathSupported(scheme, subPath string) bool {
	supportedSubPaths := map[string][]string{
		EpicURIScheme: {
			"hierarchy",
			"user-stories",
		},
		UserStoryURIScheme: {
			"requirements",
			"acceptance-criteria",
		},
		RequirementURIScheme: {
			"relationships",
		},
		AcceptanceCriteriaURIScheme: {
			// No sub-paths currently supported for acceptance criteria
		},
	}

	subPaths, exists := supportedSubPaths[scheme]
	if !exists {
		return false
	}

	for _, supportedPath := range subPaths {
		if supportedPath == subPath {
			return true
		}
	}

	return false
}

// BuildURI constructs a URI string from components
func (p *URIParser) BuildURI(scheme, referenceID, subPath string, parameters map[string]string) (string, error) {
	// Validate scheme
	if !p.isValidScheme(scheme) {
		return "", fmt.Errorf("unsupported scheme: %s", scheme)
	}

	// Validate reference ID
	if !p.isValidReferenceID(referenceID) {
		return "", fmt.Errorf("invalid reference ID format: %s", referenceID)
	}

	// Validate scheme and reference ID match
	if err := p.validateSchemeAndReferenceID(scheme, referenceID); err != nil {
		return "", err
	}

	// Build base URI
	uri := fmt.Sprintf("%s://%s", scheme, referenceID)

	// Add sub-path if provided
	if subPath != "" {
		uri += "/" + subPath
	}

	// Add parameters if provided
	if len(parameters) > 0 {
		values := url.Values{}
		for key, value := range parameters {
			values.Add(key, value)
		}
		uri += "?" + values.Encode()
	}

	return uri, nil
}
