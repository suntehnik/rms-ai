# Scripts Directory

This directory contains various validation and documentation generation scripts for the Product Requirements Management API.

## Directory Structure

Each script is organized in its own subdirectory to avoid Go package conflicts:

```
scripts/
├── README.md                           # This file
├── comprehensive-validation/           # Comprehensive OpenAPI validation
│   └── main.go
├── generate-api-docs/                  # API documentation generation
│   └── main.go
├── run-all-validation/                 # Run all validation tests
│   └── main.go
├── validate-api-completeness/          # API completeness validation
│   └── main.go
├── validate-documentation/             # Documentation accuracy validation
│   └── main.go
├── validate-openapi/                   # OpenAPI specification validation
│   └── main.go
├── validate-schemas/                   # Schema and parameter validation
│   └── main.go
└── verify-models/                      # Model verification
    └── main.go
```

## Available Scripts

### 1. Comprehensive Validation (`comprehensive-validation/`)
**Purpose**: Runs a complete validation suite for the OpenAPI specification
**Usage**: 
```bash
# Via Makefile (recommended)
make docs-validate-comprehensive

# Direct execution
go run scripts/comprehensive-validation/main.go
```
**Features**:
- Route documentation coverage
- Request/response schema validation
- Parameter definition validation
- Entity type coverage
- Authentication documentation
- Deletion system documentation
- Comment system documentation

### 2. API Documentation Generation (`generate-api-docs/`)
**Purpose**: Generates comprehensive API documentation in multiple formats
**Usage**:
```bash
# Generate all formats
make docs-generate

# Generate specific formats
make docs-generate-html
make docs-generate-markdown
make docs-generate-typescript
make docs-generate-json

# Direct execution
go run scripts/generate-api-docs/main.go -input=docs/openapi-v3.yaml -output=docs/generated -format=all
```
**Supported Formats**:
- HTML (interactive documentation)
- Markdown (developer-friendly)
- TypeScript (type definitions)
- JSON (machine-readable)

### 3. API Completeness Validation (`validate-api-completeness/`)
**Purpose**: Validates that all implemented routes are documented and vice versa
**Usage**:
```bash
go run scripts/validate-api-completeness/main.go
```
**Features**:
- Compares implemented routes vs documented routes
- Entity coverage analysis
- Missing documentation detection
- Extra documentation detection

### 4. Documentation Accuracy Validation (`validate-documentation/`)
**Purpose**: Validates documentation accuracy and consistency
**Usage**:
```bash
make docs-validate
# or
go run scripts/validate-documentation/main.go
```

### 5. OpenAPI Specification Validation (`validate-openapi/`)
**Purpose**: Validates the OpenAPI specification structure and syntax
**Usage**:
```bash
go run scripts/validate-openapi/main.go
```

### 6. Schema and Parameter Validation (`validate-schemas/`)
**Purpose**: Validates schema definitions and parameter consistency
**Usage**:
```bash
go run scripts/validate-schemas/main.go
```

### 7. Model Verification (`verify-models/`)
**Purpose**: Verifies that Go models match OpenAPI schema definitions
**Usage**:
```bash
go run scripts/verify-models/main.go
```

### 8. Run All Validation Tests (`run-all-validation/`)
**Purpose**: Executes all validation scripts in sequence
**Usage**:
```bash
go run scripts/run-all-validation/main.go
```

## Makefile Integration

The scripts are integrated with the project's Makefile for easy execution:

```bash
# Documentation validation
make docs-validate                    # Run documentation accuracy validation
make docs-validate-comprehensive     # Run comprehensive OpenAPI validation

# Documentation generation
make docs-generate                   # Generate all documentation formats
make docs-generate-html             # Generate HTML documentation
make docs-generate-markdown         # Generate Markdown documentation
make docs-generate-typescript       # Generate TypeScript definitions
make docs-generate-json             # Generate JSON documentation
```

## Development Guidelines

### Adding New Scripts

1. Create a new subdirectory under `scripts/`
2. Name the main file `main.go`
3. Use descriptive directory names with hyphens (e.g., `validate-new-feature/`)
4. Add the script to the Makefile if it should be part of the standard workflow
5. Update this README with documentation

### Script Structure

Each script should:
- Have a clear purpose and single responsibility
- Include comprehensive error handling
- Provide detailed output and progress information
- Use consistent formatting for results
- Include help/usage information

### Dependencies

Scripts should minimize external dependencies and use only:
- Go standard library
- Project internal packages (when necessary)
- Well-established third-party libraries (with justification)

## Troubleshooting

### "main redeclared" Error
This error occurs when trying to run multiple Go files with `main` functions from the same directory. The new structure prevents this by isolating each script in its own directory.

**Solution**: Use the specific script path:
```bash
# ✅ Correct
go run scripts/comprehensive-validation/main.go

# ❌ Incorrect (causes conflicts)
go run scripts/*.go
```

### Path Issues
Scripts are designed to run from the project root directory. If you encounter path-related errors, ensure you're running from the correct location:

```bash
# Run from project root
pwd  # Should show your project directory
go run scripts/script-name/main.go
```

## Contributing

When contributing new validation scripts:

1. Follow the established directory structure
2. Include comprehensive documentation
3. Add appropriate Makefile targets
4. Test the script thoroughly
5. Update this README

## Related Documentation

- [OpenAPI Specification](../docs/openapi-v3.yaml)
- [API Documentation](../docs/generated/)
- [Makefile](../Makefile)
- [Project README](../README.md)