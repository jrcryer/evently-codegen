package generator

import (
	"fmt"
	"go/format"
	"sort"
	"strings"
)

// DefaultCodeGenerator implements the CodeGenerator interface
type DefaultCodeGenerator struct {
	typeMapper    TypeMapper
	config        *Config
	nestedStructs map[string]*GoStruct // Track nested structs to generate
	structNames   map[string]bool      // Track used struct names to avoid conflicts
	enumTypes     map[string]*EnumType // Track enum types to generate
}

// NewCodeGenerator creates a new DefaultCodeGenerator instance
func NewCodeGenerator(config *Config) *DefaultCodeGenerator {
	return &DefaultCodeGenerator{
		typeMapper:    NewTypeMapper(config),
		config:        config,
		nestedStructs: make(map[string]*GoStruct),
		structNames:   make(map[string]bool),
		enumTypes:     make(map[string]*EnumType),
	}
}

// GenerateTypes generates Go type definitions from message schemas
func (cg *DefaultCodeGenerator) GenerateTypes(messages map[string]*MessageSchema, config *Config) (*GenerateResult, error) {
	if config != nil {
		cg.config = config
		cg.typeMapper = NewTypeMapper(config)
	}

	// Reset nested structs and enum types tracking for each generation
	cg.nestedStructs = make(map[string]*GoStruct)
	cg.structNames = make(map[string]bool)
	cg.enumTypes = make(map[string]*EnumType)

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

		// Collect all struct codes for this schema (main + nested + enums)
		var allStructCodes []string
		allStructCodes = append(allStructCodes, structCode)

		// Add enum type definitions
		for _, enumType := range cg.enumTypes {
			enumCode, err := cg.generateEnumCode(enumType)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("failed to generate enum type %s: %w", enumType.Name, err))
				continue
			}
			allStructCodes = append(allStructCodes, enumCode)
		}

		// Add nested struct codes with validation methods
		for _, nestedStruct := range cg.nestedStructs {
			nestedCode, err := cg.generateStructCode(nestedStruct)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("failed to generate nested struct %s: %w", nestedStruct.Name, err))
				continue
			}

			// Generate validation methods for nested struct
			// Create a minimal schema for the nested struct
			nestedSchema := cg.createSchemaFromStruct(nestedStruct)
			validationMethods, err := cg.generateValidationMethods(nestedStruct, nestedSchema)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("failed to generate validation methods for nested struct %s: %w", nestedStruct.Name, err))
				continue
			}

			allStructCodes = append(allStructCodes, nestedCode+"\n"+validationMethods)
		}

		// Combine all struct codes for this schema
		combinedStructCode := strings.Join(allStructCodes, "\n\n")

		// Create a complete Go file with package declaration and imports
		fileContent, err := cg.createGoFile(combinedStructCode, name)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to create Go file for %s: %w", name, err))
			continue
		}

		// Use snake_case filename
		filename := toSnakeCase(name) + ".go"
		result.Files[filename] = fileContent

		// Clear nested structs and enum types for next iteration
		cg.nestedStructs = make(map[string]*GoStruct)
		cg.structNames = make(map[string]bool)
		cg.enumTypes = make(map[string]*EnumType)
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
				// Handle nested objects and arrays - we'll update the types later
				// For now, just keep the default types from the type mapper

				goStruct.Fields = append(goStruct.Fields, field)
			}
		}
	}

	// Reserve the main struct name first
	cg.structNames[structName] = true

	// Process nested objects and enum types to generate additional definitions and update field types
	if schema.Properties != nil {
		cg.processNestedObjectsAndUpdateFields(goStruct, structName, schema.Properties, schema.Required)
		cg.collectEnumTypes(goStruct.Fields)
	}

	// Generate the struct code
	structCode, err := cg.generateStructCode(goStruct)
	if err != nil {
		return "", err
	}

	// Generate enum types
	var enumCodes []string
	for _, enumType := range cg.enumTypes {
		enumCode, err := cg.generateEnumCode(enumType)
		if err != nil {
			return "", fmt.Errorf("failed to generate enum type %s: %w", enumType.Name, err)
		}
		enumCodes = append(enumCodes, enumCode)
	}

	// Generate validation methods
	validationMethods, err := cg.generateValidationMethods(goStruct, schema)
	if err != nil {
		return "", err
	}

	// Combine enum types, struct, and validation methods
	var allParts []string
	if len(enumCodes) > 0 {
		allParts = append(allParts, strings.Join(enumCodes, "\n"))
	}
	allParts = append(allParts, structCode)
	allParts = append(allParts, validationMethods)

	return strings.Join(allParts, "\n"), nil
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
				fmt.Fprintf(&builder, "// %s\n", comment)
			}
		}
	}

	// Add struct declaration
	fmt.Fprintf(&builder, "type %s struct {\n", goStruct.Name)

	// Add fields
	for _, field := range goStruct.Fields {
		// Add field comment if present and comments are enabled
		if field.Comment != "" && cg.config.IncludeComments {
			fmt.Fprintf(&builder, "\t// %s\n", field.Comment)
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
func (cg *DefaultCodeGenerator) createGoFile(structCode, _ string) (string, error) {
	var builder strings.Builder

	// Add package declaration with proper formatting
	packageName := cg.getPackageName()
	if !IsValidGoIdentifier(packageName) {
		return "", fmt.Errorf("invalid package name: %s", packageName)
	}
	fmt.Fprintf(&builder, "package %s\n\n", packageName)

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
		fmt.Fprintf(builder, "import \"%s\"\n\n", imports[0])
	} else {
		builder.WriteString("import (\n")
		for _, imp := range imports {
			fmt.Fprintf(builder, "\t\"%s\"\n", imp)
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

	// Check for validation types (always include if validation methods are present)
	if strings.Contains(structCode, "ValidationResult") || strings.Contains(structCode, "NewValidator") {
		// Note: In a real implementation, this would be the actual import path
		// For this example, we assume validation types are in the same package
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

// generateNestedStructName creates a unique name for a nested struct
func generateNestedStructName(parentName, fieldName string) string {
	// Convert field name to PascalCase and combine with parent name
	fieldPascal := ToPascalCase(fieldName)
	return parentName + fieldPascal
}

// ensureUniqueStructName ensures the struct name is unique, adding a suffix if needed
func (cg *DefaultCodeGenerator) ensureUniqueStructName(baseName string) string {
	if !cg.structNames[baseName] {
		cg.structNames[baseName] = true
		return baseName
	}

	// If name is taken, try with numeric suffixes
	counter := 2
	for {
		candidateName := fmt.Sprintf("%s%d", baseName, counter)
		if !cg.structNames[candidateName] {
			cg.structNames[candidateName] = true
			return candidateName
		}
		counter++
	}
}

// processNestedObjects recursively processes nested objects and generates struct definitions
func (cg *DefaultCodeGenerator) processNestedObjects(parentName string, properties map[string]*Property, required []string) {
	for propName, prop := range properties {
		if prop.Type == "object" && prop.Properties != nil {
			// Generate nested struct name
			nestedStructName := generateNestedStructName(parentName, propName)
			uniqueName := cg.ensureUniqueStructName(nestedStructName)

			// Create the nested struct
			nestedStruct := &GoStruct{
				Name:        uniqueName,
				PackageName: cg.getPackageName(),
				Fields:      []*GoField{},
				Comments:    []string{},
			}

			// Add description comment if available
			if prop.Description != "" && cg.config.IncludeComments {
				nestedStruct.Comments = append(nestedStruct.Comments, prop.Description)
			}

			// Process fields of the nested struct
			for nestedPropName, nestedProp := range prop.Properties {
				isRequired := cg.isPropertyRequired(nestedPropName, prop.Required)
				field := cg.typeMapper.MapPropertyWithContext(nestedProp, nestedPropName, isRequired)
				if field != nil {
					// Handle further nested objects
					if nestedProp.Type == "object" && nestedProp.Properties != nil {
						nestedNestedStructName := generateNestedStructName(uniqueName, nestedPropName)
						field.Type = cg.ensureUniqueStructName(nestedNestedStructName)
					}
					nestedStruct.Fields = append(nestedStruct.Fields, field)
				}
			}

			// Store the nested struct for generation
			cg.nestedStructs[uniqueName] = nestedStruct

			// Recursively process nested objects within this nested object
			cg.processNestedObjects(uniqueName, prop.Properties, prop.Required)
		} else if prop.Type == "array" && prop.Items != nil && prop.Items.Type == "object" && prop.Items.Properties != nil {
			// Handle arrays of objects - create a struct for the array item
			itemStructName := generateNestedStructName(parentName, strings.TrimSuffix(propName, "s")) // Remove 's' for singular
			if itemStructName == parentName+ToPascalCase(propName) {
				// If removing 's' didn't work, use "Item" suffix
				itemStructName = generateNestedStructName(parentName, propName+"Item")
			}
			uniqueName := cg.ensureUniqueStructName(itemStructName)

			// Create the item struct
			itemStruct := &GoStruct{
				Name:        uniqueName,
				PackageName: cg.getPackageName(),
				Fields:      []*GoField{},
				Comments:    []string{},
			}

			// Add description comment if available
			if prop.Items.Description != "" && cg.config.IncludeComments {
				itemStruct.Comments = append(itemStruct.Comments, prop.Items.Description)
			}

			// Process fields of the item struct
			for itemPropName, itemProp := range prop.Items.Properties {
				isRequired := cg.isPropertyRequired(itemPropName, prop.Items.Required)
				field := cg.typeMapper.MapPropertyWithContext(itemProp, itemPropName, isRequired)
				if field != nil {
					// Handle nested objects within array items
					if itemProp.Type == "object" && itemProp.Properties != nil {
						nestedStructName := generateNestedStructName(uniqueName, itemPropName)
						field.Type = cg.ensureUniqueStructName(nestedStructName)
					}
					itemStruct.Fields = append(itemStruct.Fields, field)
				}
			}

			// Store the item struct for generation
			cg.nestedStructs[uniqueName] = itemStruct

			// Recursively process nested objects within array items
			cg.processNestedObjects(uniqueName, prop.Items.Properties, prop.Items.Required)
		}
	}
}

// updateFieldTypes updates field types with the actual generated struct names
func (cg *DefaultCodeGenerator) updateFieldTypes(goStruct *GoStruct, parentName string, properties map[string]*Property) {
	for _, field := range goStruct.Fields {
		// Find the corresponding property
		var prop *Property
		for propName, p := range properties {
			if field.JSONTag == fmt.Sprintf(`json:"%s"`, propName) {
				prop = p
				break
			}
		}

		if prop == nil {
			continue
		}

		// Update nested object types
		if prop.Type == "object" && prop.Properties != nil {
			expectedName := generateNestedStructName(parentName, getFieldNameFromJSONTag(field.JSONTag))
			// Find the actual generated name
			for actualName := range cg.nestedStructs {
				if strings.HasPrefix(actualName, expectedName) {
					field.Type = actualName
					break
				}
			}
		} else if prop.Type == "array" && prop.Items != nil && prop.Items.Type == "object" && prop.Items.Properties != nil {
			propName := getFieldNameFromJSONTag(field.JSONTag)
			expectedName := generateNestedStructName(parentName, strings.TrimSuffix(propName, "s"))
			if expectedName == parentName+ToPascalCase(propName) {
				expectedName = generateNestedStructName(parentName, propName+"Item")
			}
			// Find the actual generated name
			for actualName := range cg.nestedStructs {
				if strings.HasPrefix(actualName, expectedName) {
					field.Type = "[]" + actualName
					break
				}
			}
		}
	}
}

// getFieldNameFromJSONTag extracts the field name from a JSON tag
func getFieldNameFromJSONTag(jsonTag string) string {
	// Extract field name from json:"fieldName"
	start := strings.Index(jsonTag, `"`) + 1
	end := strings.LastIndex(jsonTag, `"`)
	if start > 0 && end > start {
		return jsonTag[start:end]
	}
	return ""
}

// generateValidationMethods generates validation methods for a Go struct
func (cg *DefaultCodeGenerator) generateValidationMethods(goStruct *GoStruct, schema *MessageSchema) (string, error) {
	var builder strings.Builder

	// Generate Validate() method
	builder.WriteString(cg.generateValidateMethod(goStruct, schema))
	builder.WriteString("\n\n")

	// Generate ValidateJSON() method
	builder.WriteString(cg.generateValidateJSONMethod(goStruct, schema))

	return builder.String(), nil
}

// generateValidateMethod generates a Validate() method for the struct
func (cg *DefaultCodeGenerator) generateValidateMethod(goStruct *GoStruct, schema *MessageSchema) string {
	var builder strings.Builder

	// Method signature
	fmt.Fprintf(&builder, "// Validate validates the %s struct against its schema\n", goStruct.Name)
	fmt.Fprintf(&builder, "func (s *%s) Validate() *ValidationResult {\n", goStruct.Name)

	// Create validator instance
	builder.WriteString("\tvalidator := NewValidator(false) // Use permissive mode by default\n")

	// Convert struct to property for validation
	builder.WriteString("\tschema := &Property{\n")
	builder.WriteString("\t\tType: \"object\",\n")

	// Add properties
	if len(goStruct.Fields) > 0 {
		builder.WriteString("\t\tProperties: map[string]*Property{\n")
		for _, field := range goStruct.Fields {
			propName := cg.getPropertyNameFromField(field)
			propType := cg.getSchemaTypeFromGoType(field.Type)

			// Handle enum fields
			if field.IsEnum {
				fmt.Fprintf(&builder, "\t\t\t\"%s\": {Type: \"%s\", Enum: []interface{}{", propName, propType)
				for i, enumValue := range field.EnumValues {
					if i > 0 {
						builder.WriteString(", ")
					}
					switch enumValue.(type) {
					case string:
						fmt.Fprintf(&builder, "\"%v\"", enumValue)
					default:
						fmt.Fprintf(&builder, "%v", enumValue)
					}
				}
				builder.WriteString("}},\n")
			} else {
				fmt.Fprintf(&builder, "\t\t\t\"%s\": {Type: \"%s\"},\n", propName, propType)
			}
		}
		builder.WriteString("\t\t},\n")
	}

	// Add required fields
	if len(schema.Required) > 0 {
		builder.WriteString("\t\tRequired: []string{")
		for i, req := range schema.Required {
			if i > 0 {
				builder.WriteString(", ")
			}
			fmt.Fprintf(&builder, "\"%s\"", req)
		}
		builder.WriteString("},\n")
	}

	builder.WriteString("\t}\n\n")

	// Convert struct to map[string]any for validation
	builder.WriteString("\t// Convert struct to map for validation\n")
	builder.WriteString("\tdata := map[string]any{\n")
	for _, field := range goStruct.Fields {
		propName := cg.getPropertyNameFromField(field)
		fmt.Fprintf(&builder, "\t\t\"%s\": s.%s,\n", propName, field.Name)
	}
	builder.WriteString("\t}\n\n")

	// Perform validation
	builder.WriteString("\treturn validator.ValidateValue(data, schema, \"\")\n")
	builder.WriteString("}\n")

	return builder.String()
}

// generateValidateJSONMethod generates a ValidateJSON() method for the struct
func (cg *DefaultCodeGenerator) generateValidateJSONMethod(goStruct *GoStruct, schema *MessageSchema) string {
	var builder strings.Builder

	// Method signature
	fmt.Fprintf(&builder, "// ValidateJSON validates raw JSON data against the %s schema\n", goStruct.Name)
	fmt.Fprintf(&builder, "func (s *%s) ValidateJSON(jsonData []byte) *ValidationResult {\n", goStruct.Name)

	// Create validator instance
	builder.WriteString("\tvalidator := NewValidator(false) // Use permissive mode by default\n")

	// Create message schema for validation
	builder.WriteString("\tschema := &MessageSchema{\n")
	builder.WriteString("\t\tType: \"object\",\n")

	// Add properties
	if len(goStruct.Fields) > 0 {
		builder.WriteString("\t\tProperties: map[string]*Property{\n")
		for _, field := range goStruct.Fields {
			propName := cg.getPropertyNameFromField(field)
			propType := cg.getSchemaTypeFromGoType(field.Type)

			// Handle enum fields
			if field.IsEnum {
				fmt.Fprintf(&builder, "\t\t\t\"%s\": {Type: \"%s\", Enum: []interface{}{", propName, propType)
				for i, enumValue := range field.EnumValues {
					if i > 0 {
						builder.WriteString(", ")
					}
					switch enumValue.(type) {
					case string:
						fmt.Fprintf(&builder, "\"%v\"", enumValue)
					default:
						fmt.Fprintf(&builder, "%v", enumValue)
					}
				}
				builder.WriteString("}},\n")
			} else {
				fmt.Fprintf(&builder, "\t\t\t\"%s\": {Type: \"%s\"},\n", propName, propType)
			}
		}
		builder.WriteString("\t\t},\n")
	}

	// Add required fields
	if len(schema.Required) > 0 {
		builder.WriteString("\t\tRequired: []string{")
		for i, req := range schema.Required {
			if i > 0 {
				builder.WriteString(", ")
			}
			fmt.Fprintf(&builder, "\"%s\"", req)
		}
		builder.WriteString("},\n")
	}

	builder.WriteString("\t}\n\n")

	// Perform validation
	builder.WriteString("\treturn validator.ValidateJSON(jsonData, schema)\n")
	builder.WriteString("}\n")

	return builder.String()
}

// getPropertyNameFromField extracts the property name from a field's JSON tag
func (cg *DefaultCodeGenerator) getPropertyNameFromField(field *GoField) string {
	if field.JSONTag != "" {
		// Extract from json:"propertyName"
		start := strings.Index(field.JSONTag, `"`) + 1
		end := strings.LastIndex(field.JSONTag, `"`)
		if start > 0 && end > start {
			return field.JSONTag[start:end]
		}
	}
	// Fallback to field name in lowercase
	return strings.ToLower(field.Name)
}

// getSchemaTypeFromGoType maps Go types back to JSON schema types
func (cg *DefaultCodeGenerator) getSchemaTypeFromGoType(goType string) string {
	// Remove pointer indicators
	goType = strings.TrimPrefix(goType, "*")

	// Remove slice indicators
	if strings.HasPrefix(goType, "[]") {
		return "array"
	}

	switch goType {
	case "string":
		return "string"
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		return "integer"
	case "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	case "time.Time":
		return "string"
	case "interface{}", "any":
		return ""
	default:
		// Check if it's an enum type
		if strings.HasSuffix(goType, "Enum") {
			// Determine base type from enum type
			if enumType, exists := cg.enumTypes[goType]; exists {
				return enumType.BaseType
			}
			// Fallback: try to infer from name
			return "string" // Most enums are string-based
		}
		// If it's a custom type (struct), assume it's an object
		if goType != "" && goType[0] >= 'A' && goType[0] <= 'Z' {
			return "object"
		}
		return ""
	}
}

// createSchemaFromStruct creates a minimal MessageSchema from a GoStruct
func (cg *DefaultCodeGenerator) createSchemaFromStruct(goStruct *GoStruct) *MessageSchema {
	schema := &MessageSchema{
		Type:       "object",
		Properties: make(map[string]*Property),
		Required:   []string{}, // For nested structs, we don't have required field info
	}

	for _, field := range goStruct.Fields {
		propName := cg.getPropertyNameFromField(field)
		propType := cg.getSchemaTypeFromGoType(field.Type)

		schema.Properties[propName] = &Property{
			Type: propType,
		}
	}

	return schema
}

// processNestedObjectsAndUpdateFields processes nested objects and updates field types in one pass
func (cg *DefaultCodeGenerator) processNestedObjectsAndUpdateFields(goStruct *GoStruct, parentName string, properties map[string]*Property, required []string) {
	// Create a map to track field updates
	fieldUpdates := make(map[string]string)

	// Process each property
	for propName, prop := range properties {
		if prop.Type == "object" && prop.Properties != nil {
			// Generate nested struct
			nestedStructName := generateNestedStructName(parentName, propName)
			uniqueName := cg.ensureUniqueStructName(nestedStructName)

			// Create the nested struct
			nestedStruct := &GoStruct{
				Name:        uniqueName,
				PackageName: cg.getPackageName(),
				Fields:      []*GoField{},
				Comments:    []string{},
			}

			// Add description comment if available
			if prop.Description != "" && cg.config.IncludeComments {
				nestedStruct.Comments = append(nestedStruct.Comments, prop.Description)
			}

			// Process fields of the nested struct
			for nestedPropName, nestedProp := range prop.Properties {
				isRequired := cg.isPropertyRequired(nestedPropName, prop.Required)
				field := cg.typeMapper.MapPropertyWithContext(nestedProp, nestedPropName, isRequired)
				if field != nil {
					nestedStruct.Fields = append(nestedStruct.Fields, field)
				}
			}

			// Store the nested struct for generation
			cg.nestedStructs[uniqueName] = nestedStruct

			// Record the field update
			fieldUpdates[propName] = uniqueName

			// Recursively process nested objects within this nested object
			cg.processNestedObjectsAndUpdateFields(nestedStruct, uniqueName, prop.Properties, prop.Required)

		} else if prop.Type == "array" && prop.Items != nil && prop.Items.Type == "object" && prop.Items.Properties != nil {
			// Handle arrays of objects - create a struct for the array item
			itemStructName := generateNestedStructName(parentName, strings.TrimSuffix(propName, "s"))
			if itemStructName == parentName+ToPascalCase(propName) {
				itemStructName = generateNestedStructName(parentName, propName+"Item")
			}
			uniqueName := cg.ensureUniqueStructName(itemStructName)

			// Create the item struct
			itemStruct := &GoStruct{
				Name:        uniqueName,
				PackageName: cg.getPackageName(),
				Fields:      []*GoField{},
				Comments:    []string{},
			}

			// Add description comment if available
			if prop.Items.Description != "" && cg.config.IncludeComments {
				itemStruct.Comments = append(itemStruct.Comments, prop.Items.Description)
			}

			// Process fields of the item struct
			for itemPropName, itemProp := range prop.Items.Properties {
				isRequired := cg.isPropertyRequired(itemPropName, prop.Items.Required)
				field := cg.typeMapper.MapPropertyWithContext(itemProp, itemPropName, isRequired)
				if field != nil {
					itemStruct.Fields = append(itemStruct.Fields, field)
				}
			}

			// Store the item struct for generation
			cg.nestedStructs[uniqueName] = itemStruct

			// Record the field update
			fieldUpdates[propName] = "[]" + uniqueName

			// Recursively process nested objects within array items
			cg.processNestedObjectsAndUpdateFields(itemStruct, uniqueName, prop.Items.Properties, prop.Items.Required)
		}
	}

	// Update field types in the current struct
	for _, field := range goStruct.Fields {
		propName := getFieldNameFromJSONTag(field.JSONTag)
		if newType, exists := fieldUpdates[propName]; exists {
			field.Type = newType
		}
	}
}

// collectEnumTypes collects enum types from struct fields
func (cg *DefaultCodeGenerator) collectEnumTypes(fields []*GoField) {
	for _, field := range fields {
		if field.IsEnum {
			cg.addEnumType(field.Type, field.EnumBaseType, field.EnumValues, field.Comment)
		}
		if field.ArrayItemEnum != nil {
			cg.addEnumType(field.ArrayItemEnum.TypeName, field.ArrayItemEnum.BaseType, field.ArrayItemEnum.Values, "")
		}
	}
}

// addEnumType adds an enum type to the collection
func (cg *DefaultCodeGenerator) addEnumType(typeName, baseType string, values []interface{}, comment string) {
	if _, exists := cg.enumTypes[typeName]; exists {
		return // Already added
	}

	enumType := &EnumType{
		Name:     typeName,
		BaseType: baseType,
		Values:   make([]EnumValue, 0, len(values)),
		Comment:  comment,
	}

	for _, value := range values {
		enumValue := EnumValue{
			Name:  cg.generateEnumValueName(typeName, value),
			Value: value,
		}
		enumType.Values = append(enumType.Values, enumValue)
	}

	cg.enumTypes[typeName] = enumType
}

// generateEnumValueName generates a Go constant name for an enum value
func (cg *DefaultCodeGenerator) generateEnumValueName(typeName string, value interface{}) string {
	// Remove "Enum" suffix from type name for the constant prefix
	prefix := strings.TrimSuffix(typeName, "Enum")

	var valueName string
	switch v := value.(type) {
	case string:
		// Convert string value to PascalCase
		valueName = ToPascalCase(v)
	case int, int32, int64:
		valueName = fmt.Sprintf("%v", v)
	case float64:
		// Check if it's actually an integer stored as float64
		if v == float64(int64(v)) {
			valueName = fmt.Sprintf("%d", int64(v))
		} else {
			valueName = fmt.Sprintf("%g", v)
		}
	case bool:
		if v {
			valueName = "True"
		} else {
			valueName = "False"
		}
	default:
		valueName = fmt.Sprintf("%v", v)
	}

	return prefix + valueName
}

// generateEnumCode generates Go code for an enum type
func (cg *DefaultCodeGenerator) generateEnumCode(enumType *EnumType) (string, error) {
	var builder strings.Builder

	// Add enum type comment
	if enumType.Comment != "" && cg.config.IncludeComments {
		fmt.Fprintf(&builder, "// %s\n", enumType.Comment)
	}

	// Add type definition
	fmt.Fprintf(&builder, "type %s %s\n\n", enumType.Name, enumType.BaseType)

	// Add const block with enum values
	if len(enumType.Values) > 0 {
		builder.WriteString("const (\n")
		for i, value := range enumType.Values {
			if i == 0 {
				// First value uses iota or explicit value
				switch enumType.BaseType {
				case "string":
					fmt.Fprintf(&builder, "\t%s %s = %q\n", value.Name, enumType.Name, value.Value)
				case "int", "int32", "int64":
					fmt.Fprintf(&builder, "\t%s %s = %v\n", value.Name, enumType.Name, value.Value)
				case "float64", "float32":
					fmt.Fprintf(&builder, "\t%s %s = %v\n", value.Name, enumType.Name, value.Value)
				case "bool":
					fmt.Fprintf(&builder, "\t%s %s = %v\n", value.Name, enumType.Name, value.Value)
				default:
					fmt.Fprintf(&builder, "\t%s %s = %v\n", value.Name, enumType.Name, value.Value)
				}
			} else {
				// Subsequent values
				switch enumType.BaseType {
				case "string":
					fmt.Fprintf(&builder, "\t%s = %q\n", value.Name, value.Value)
				default:
					fmt.Fprintf(&builder, "\t%s = %v\n", value.Name, value.Value)
				}
			}
		}
		builder.WriteString(")\n\n")
	}

	// Add validation method for the enum
	builder.WriteString(cg.generateEnumValidationMethod(enumType))

	return builder.String(), nil
}

// generateEnumValidationMethod generates a validation method for an enum type
func (cg *DefaultCodeGenerator) generateEnumValidationMethod(enumType *EnumType) string {
	var builder strings.Builder

	// Generate IsValid method
	fmt.Fprintf(&builder, "// IsValid returns true if the %s value is valid\n", enumType.Name)
	fmt.Fprintf(&builder, "func (e %s) IsValid() bool {\n", enumType.Name)
	builder.WriteString("\tswitch e {\n")

	for _, value := range enumType.Values {
		fmt.Fprintf(&builder, "\tcase %s:\n", value.Name)
	}

	builder.WriteString("\t\treturn true\n")
	builder.WriteString("\tdefault:\n")
	builder.WriteString("\t\treturn false\n")
	builder.WriteString("\t}\n")
	builder.WriteString("}\n\n")

	// Generate String method for string enums
	if enumType.BaseType == "string" {
		fmt.Fprintf(&builder, "// String returns the string representation of %s\n", enumType.Name)
		fmt.Fprintf(&builder, "func (e %s) String() string {\n", enumType.Name)
		fmt.Fprintf(&builder, "\treturn string(e)\n")
		builder.WriteString("}\n\n")
	}

	// Generate Values method that returns all valid values
	fmt.Fprintf(&builder, "// %sValues returns all valid %s values\n", enumType.Name, enumType.Name)
	fmt.Fprintf(&builder, "func %sValues() []%s {\n", enumType.Name, enumType.Name)
	fmt.Fprintf(&builder, "\treturn []%s{\n", enumType.Name)

	for _, value := range enumType.Values {
		fmt.Fprintf(&builder, "\t\t%s,\n", value.Name)
	}

	builder.WriteString("\t}\n")
	builder.WriteString("}\n")

	return builder.String()
}
