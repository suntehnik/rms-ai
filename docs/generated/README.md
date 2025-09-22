# Product Requirements Management API - Generated Documentation

This directory contains comprehensive API documentation generated from the OpenAPI 3.0.3 specification. The documentation is available in multiple formats to support different use cases and development workflows.

## 📁 Documentation Files

### Interactive Documentation
- **[index.html](index.html)** - Documentation hub with links to all formats
- **[swagger-ui.html](swagger-ui.html)** - Interactive Swagger UI for API testing

### Reference Documentation
- **[api-documentation.html](api-documentation.html)** - Complete HTML reference
- **[api-documentation.md](api-documentation.md)** - Markdown format for wikis/README
- **[developer-guide.md](developer-guide.md)** - Comprehensive integration guide

### Development Resources
- **[api-types.ts](api-types.ts)** - TypeScript interface definitions
- **[api-documentation.json](api-documentation.json)** - JSON schema for tooling

## 🚀 Quick Start

### 1. Interactive Exploration
Open `swagger-ui.html` in your browser to:
- Test API endpoints with live requests
- View request/response examples
- Authenticate and make real API calls
- Validate request formats

### 2. Development Integration
Use `api-types.ts` for TypeScript projects:
```typescript
import { Epic, CreateEpicRequest, ApiClient } from './api-types';

const client: ApiClient = new RequirementsApiClient('http://localhost:8080');
const epic: Epic = await client.createEpic(epicRequest);
```

### 3. Documentation Integration
Include `api-documentation.md` in your project documentation or copy sections as needed.

## 🔄 Regenerating Documentation

The documentation is generated from the OpenAPI specification using automated tools:

```bash
# Generate all formats
make docs-generate

# Generate specific formats
make docs-generate-html      # HTML documentation
make docs-generate-markdown  # Markdown documentation
make docs-generate-typescript # TypeScript interfaces
make docs-generate-json      # JSON schema

# Generate interactive Swagger UI
make swagger
```

## 📊 Documentation Coverage

- **80+ API Endpoints**: Complete coverage of all implemented endpoints
- **15+ Entity Types**: Full documentation of all data models
- **5 Documentation Formats**: Multiple formats for different use cases
- **100% OpenAPI Coverage**: Generated directly from specification

## 🎯 Use Cases by Format

### HTML Documentation (`api-documentation.html`)
- **Best for**: Online reference, sharing with stakeholders
- **Features**: Searchable, styled, complete endpoint reference
- **Use when**: Need professional-looking documentation for presentations

### Markdown Documentation (`api-documentation.md`)
- **Best for**: README files, wikis, version control
- **Features**: Plain text, copy-paste friendly, platform independent
- **Use when**: Integrating into existing documentation systems

### TypeScript Interfaces (`api-types.ts`)
- **Best for**: Type-safe client development
- **Features**: Complete type definitions, IDE support, compile-time checking
- **Use when**: Building TypeScript/JavaScript clients

### JSON Schema (`api-documentation.json`)
- **Best for**: Automated tooling, testing frameworks
- **Features**: Machine-readable, structured data, programmatic access
- **Use when**: Building tools or automated testing

### Interactive Swagger UI (`swagger-ui.html`)
- **Best for**: API testing, development, debugging
- **Features**: Live testing, authentication, request building
- **Use when**: Developing against the API or debugging issues

## 🔧 Customization

### Modifying Templates
The documentation is generated using Go templates in `scripts/generate_api_documentation.go`. To customize:

1. Edit the template strings in the generation script
2. Regenerate documentation with `make docs-generate`
3. Templates support Go template syntax with custom functions

### Adding New Formats
To add new documentation formats:

1. Add a new generation function to `scripts/generate_api_documentation.go`
2. Add the format to the switch statement in `main()`
3. Add a new Makefile target for the format
4. Update this README with the new format information

## 📋 Validation

The generated documentation is validated against the OpenAPI specification:

```bash
# Validate OpenAPI specification
make swagger-validate

# Check documentation completeness
go run scripts/validate_openapi_completeness.go

# Generate documentation quality metrics
make docs-metrics
```

## 🔗 Integration Examples

### CI/CD Pipeline
```yaml
- name: Generate API Documentation
  run: |
    make docs-generate
    # Upload to documentation site
    aws s3 sync docs/generated/ s3://api-docs-bucket/
```

### Development Workflow
```bash
# After updating OpenAPI spec
make swagger                 # Update Swagger docs
make docs-generate          # Regenerate all formats
git add docs/generated/     # Commit updated docs
```

### Client SDK Generation
```bash
# Use OpenAPI spec for code generation
openapi-generator generate \
  -i docs/openapi-v3.yaml \
  -g typescript-fetch \
  -o generated-client/
```

## 📞 Support

- **Issues**: Report documentation issues in the main project repository
- **Updates**: Documentation is automatically updated when the OpenAPI spec changes
- **Contributions**: Improve templates and generation scripts via pull requests

## 📄 License

This documentation is generated from the Product Requirements Management API specification and follows the same license terms as the main project.

---

*Generated from OpenAPI specification version 1.0.0*  
*Last updated: Auto-generated*