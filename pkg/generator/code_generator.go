package generator

import (
	"fmt"
	"go/format"
	"sort"
	"strings"
)

// DefaultCodeGenerator implements the CodeGenerator interface
type DefaultCodeGenerator struct {
	typeMapper TypeMapper
	config     *Config
}

// NewCodeGenerator creates a new DefaultCodeGenerator instance
func NewCodeGenerator(config *Config) *DefaultCodeGenerator {
	return &DefaultCodeGenerator{
		typeMapper: NewTypeMapper(config),
		config:     config,
	}
}

// GenerateTypes generates Go type definitions from message schemas
func (cg *DefaultCodeGenerator) GenerateTypes(messages map[string]*MessageSchema, config *Config) (*GenerateResult, error) {
	if config != nil {
		cg.config = config
		cg.typeMapper = NewTypeMapper(config)
	}

	result := &GenerateResult{
		Files:  make(map[string]string),
		Errors: []error{},
	}

	// Generate structs for each message schema
	for name, schema := range messages {
		structCode, err := cg.GenerateStruct(schema, name)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to generate struct for %s: %w", name, err))
			continue
		}

		// Create a complete Go file with package declaration and imports
		fileContent, err := cg.createGoFile(structCode, name)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to create Go file for %s: %w", name, err))
			continue
		}

		// Use snake_case filename
		filename := toSnakeCase(name) + ".go"
		result.Files[filename] = fileContent
	}

	return result, nil
}

// GenerateStruct generates a Go struct from a message schema
func (cg *DefaultCodeGenerator) GenerateStruct(schema *MessageSchema, name string) (string, error) {
	if schema == nil {
		return "", fmt.Errorf("schema cannot be nil")
	}

	// Convert name to PascalCase for struct name
	structName := ToPascalCase(name)
	if structName == "" {
		structName = "GeneratedStruct"
	}

	goStruct := &GoStruct{
		Name:        structName,
		PackageName: cg.getPackageName(),
		Fields:      []*GoField{},
		Comments:    []string{},
	}

	// Add struct comment from schema description (if comments are enabled)
	if cg.config.IncludeComments {
		if schema.Description != "" {
			goStruct.Comments = append(goStruct.Comments, schema.Description)
		}
		if schema.Title != "" && schema.Title != schema.Description {
			if len(goStruct.Comments) > 0 {
				goStruct.Comments = append(goStruct.Comments, "")
			}
			goStruct.Comments = append(goStruct.Comments, "Title: "+schema.Title)
		}
	}

	// Generate fields from properties
	if schema.Properties != nil {
		// Sort property names for consistent output
		propertyNames := make([]string, 0, len(schema.Properties))
		for propName := range schema.Properties {
			propertyNames = append(propertyNames, propName)
		}
		sort.Strings(propertyNames)

		for _, propName := range propertyNames {
			prop := schema.Properties[propName]
			required := cg.isPropertyRequired(propName, schema.Required)

			field := cg.typeMapper.MapPropertyWithContext(prop, propName, required)
			if field != nil {
				// Handle nested objects
				if prop.Type == "object" && prop.Properties != nil {
					nestedStructName := ToPascalCase(propName)
					field.Type = nestedStructName
					// TODO: Generate nested struct definitions
				}

				goStruct.Fields = append(goStruct.Fields, field)
			}
		}
	}

	// Generate the struct code
	return cg.generateStructCode(goStruct)
}

// generateStructCode generates the actual Go struct code
func (cg *DefaultCodeGenerator) generateStructCode(goStruct *GoStruct) (string, error) {
	var builder strings.Builder

	// Add struct comments
	if len(goStruct.Comments) > 0 {
		for _, comment := range goStruct.Comments {
			if comment == "" {
				builder.WriteString("//\n")
			} else {
				builder.WriteString(fmt.Sprintf("// %s\n", comment))
			}
		}
	}

	// Add struct declaration
	builder.WriteString(fmt.Sprintf("type %s struct {\n", goStruct.Name))

	// Add fields
	for _, field := range goStruct.Fields {
		// Add field comment if present and comments are enabled
		if field.Comment != "" && cg.config.IncludeComments {
			builder.WriteString(fmt.Sprintf("\t// %s\n", field.Comment))
		}

		// Add field declaration
		fieldLine := fmt.Sprintf("\t%s %s", field.Name, field.Type)

		// Add JSON tag
		if field.JSONTag != "" {
			fieldLine += fmt.Sprintf(" `%s`", field.JSONTag)
		}

		builder.WriteString(fieldLine + "\n")
	}

	builder.WriteString("}\n")

	return builder.String(), nil
}

// createGoFile creates a complete Go file with package declaration and imports
func (cg *DefaultCodeGenerator) createGoFile(structCode, structName string) (string, error) {
	var builder strings.Builder

	// Add package declaration with proper formatting
	packageName := cg.getPackageName()
	if !IsValidGoIdentifier(packageName) {
		return "", fmt.Errorf("invalid package name: %s", packageName)
	}
	builder.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	// Add imports if needed
	imports := cg.getRequiredImports(structCode)
	if len(imports) > 0 {
		cg.writeImports(&builder, imports)
	}

	// Add struct code
	builder.WriteString(structCode)

	// Format the Go code using go/format
	return cg.formatGoCode(builder.String())
}

// writeImports writes the import section with proper formatting
func (cg *DefaultCodeGenerator) writeImports(builder *strings.Builder, imports []string) {
	if len(imports) == 1 {
		builder.WriteString(fmt.Sprintf("import \"%s\"\n\n", imports[0]))
	} else {
		builder.WriteString("import (\n")
		for _, imp := range imports {
			builder.WriteString(fmt.Sprintf("\t\"%s\"\n", imp))
		}
		builder.WriteString(")\n\n")
	}
}

// formatGoCode formats Go source code using go/format
func (cg *DefaultCodeGenerator) formatGoCode(code string) (string, error) {
	formatted, err := format.Source([]byte(code))
	if err != nil {
		return "", &GenerationError{
			Schema:  "code formatting",
			Message: fmt.Sprintf("failed to format Go code: %v", err),
		}
	}
	return string(formatted), nil
}

// getRequiredImports analyzes the struct code and returns required imports
func (cg *DefaultCodeGenerator) getRequiredImports(structCode string) []string {
	var imports []string
	importSet := make(map[string]bool)

	// Check for time.Time usage
	if strings.Contains(structCode, "time.Time") {
		importSet["time"] = true
	}

	// Check for other common types that require imports
	if strings.Contains(structCode, "json.") {
		importSet["encoding/json"] = true
	}
	if strings.Contains(structCode, "fmt.") {
		importSet["fmt"] = true
	}
	if strings.Contains(structCode, "url.URL") {
		importSet["net/url"] = true
	}
	if strings.Contains(structCode, "uuid.UUID") {
		importSet["github.com/google/uuid"] = true
	}

	// Convert set to slice
	for imp := range importSet {
		imports = append(imports, imp)
	}

	// Sort imports for consistent output
	sort.Strings(imports)
	return imports
}

// getPackageName returns the package name from config or default
func (cg *DefaultCodeGenerator) getPackageName() string {
	if cg.config != nil && cg.config.PackageName != "" {
		return cg.config.PackageName
	}
	return "generated"
}

// isPropertyRequired checks if a property is in the required list
func (cg *DefaultCodeGenerator) isPropertyRequired(propName string, required []string) bool {
	for _, req := range required {
		if req == propName {
			return true
		}
	}
	return false
}

// toSnakeCase converts a string to snake_case
func toSnakeCase(s string) string {
	if s == "" {
		return ""
	}

	var result strings.Builder
	for i, r := range s {
		if i > 0 && (r >= 'A' && r <= 'Z') {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}

	return strings.ToLower(result.String())
}

// IsValidGoIdentifier checks if a string is a valid Go identifier
func IsValidGoIdentifier(s string) bool {
	if s == "" {
		return false
	}

	// Check first character (must be letter or underscore)
	first := rune(s[0])
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}

	// Check remaining characters (must be letter, digit, or underscore)
	for _, r := range s[1:] {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}

	// Check if it's a Go keyword
	return !isGoKeyword(s)
}

// isGoKeyword checks if a string is a Go keyword
func isGoKeyword(s string) bool {
	keywords := map[string]bool{
		"break": true, "case": true, "chan": true, "const": true, "continue": true,
		"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
		"func": true, "go": true, "goto": true, "if": true, "import": true,
		"interface": true, "map": true, "package": true, "range": true, "return": true,
		"select": true, "struct": true, "switch": true, "type": true, "var": true,
	}
	return keywords[s]
}
