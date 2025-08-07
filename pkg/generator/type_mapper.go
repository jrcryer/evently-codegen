package generator

import (
	"fmt"
	"strings"
)

// DefaultTypeMapper implements the TypeMapper interface
type DefaultTypeMapper struct {
	config *Config
}

// NewTypeMapper creates a new DefaultTypeMapper instance
func NewTypeMapper(config *Config) *DefaultTypeMapper {
	return &DefaultTypeMapper{
		config: config,
	}
}

// MapType maps AsyncAPI/JSON Schema types to Go types
func (tm *DefaultTypeMapper) MapType(schemaType string, format string) string {
	// Handle format-specific mappings first
	if format != "" {
		if goType := tm.mapFormatType(schemaType, format); goType != "" {
			return goType
		}
	}

	// Handle basic type mappings
	switch schemaType {
	case "string":
		return "string"
	case "number":
		return "float64"
	case "integer":
		return "int64"
	case "boolean":
		return "bool"
	case "array":
		return "[]any" // Will be refined by MapProperty
	case "object":
		return "any" // Will be refined by MapProperty for nested objects
	case "null":
		return "any"
	default:
		// Unknown type, default to any
		return "any"
	}
}

// mapFormatType handles format-specific type mappings
func (tm *DefaultTypeMapper) mapFormatType(schemaType, format string) string {
	switch schemaType {
	case "string":
		switch format {
		case "date-time":
			return "time.Time"
		case "date":
			return "time.Time"
		case "time":
			return "time.Time"
		case "email":
			return "string" // Could be a custom email type in the future
		case "hostname":
			return "string"
		case "ipv4":
			return "string"
		case "ipv6":
			return "string"
		case "uri":
			return "string"
		case "uri-reference":
			return "string"
		case "uuid":
			return "string"
		case "byte":
			return "[]byte"
		case "binary":
			return "[]byte"
		case "password":
			return "string"
		}
	case "integer":
		switch format {
		case "int32":
			return "int32"
		case "int64":
			return "int64"
		default:
			return "int64"
		}
	case "number":
		switch format {
		case "float":
			return "float32"
		case "double":
			return "float64"
		default:
			return "float64"
		}
	}
	return ""
}

// MapProperty maps a Property to a GoField
func (tm *DefaultTypeMapper) MapProperty(prop *Property) *GoField {
	if prop == nil {
		return nil
	}

	field := &GoField{
		Comment: prop.Description,
	}

	// Handle array types
	if prop.Type == "array" && prop.Items != nil {
		itemType := tm.getPropertyType(prop.Items)
		field.Type = "[]" + itemType
	} else if prop.Type == "object" && prop.Properties != nil {
		// For nested objects, we'll use a struct type name
		// This will be handled by the code generator to create nested structs
		field.Type = "struct{}" // Placeholder, will be replaced by actual struct name
	} else {
		// Handle primitive types
		field.Type = tm.MapType(prop.Type, prop.Format)
	}

	// Handle optional vs required fields
	field.Optional = !tm.isRequired(prop)
	if field.Optional && tm.config != nil && tm.config.UsePointers {
		// Make optional fields pointers if configured to do so
		if !strings.HasPrefix(field.Type, "[]") && !strings.HasPrefix(field.Type, "map[") {
			field.Type = "*" + field.Type
		}
	}

	return field
}

// getPropertyType returns the Go type for a property
func (tm *DefaultTypeMapper) getPropertyType(prop *Property) string {
	if prop == nil {
		return "any"
	}

	// Check for enum first
	if tm.isEnumProperty(prop) {
		// For getPropertyType, we can't generate a specific enum name
		// so we return the base type. This is mainly used for nested cases
		// where the specific enum type will be handled by MapPropertyWithContext
		return tm.getEnumBaseType(prop.Enum)
	}

	if prop.Type == "array" && prop.Items != nil {
		itemType := tm.getPropertyType(prop.Items)
		return "[]" + itemType
	}

	if prop.Type == "object" {
		return "any" // Will be refined for nested objects
	}

	return tm.MapType(prop.Type, prop.Format)
}

// isRequired determines if a property should be treated as required
// This is a simplified version - in practice, this would check against
// the parent schema's Required slice
func (tm *DefaultTypeMapper) isRequired(prop *Property) bool {
	// For now, assume all properties are optional unless explicitly marked
	// This will be refined when we have the full context from the parent schema
	return false
}

// ToPascalCase converts a string to PascalCase for Go naming conventions
func ToPascalCase(s string) string {
	if s == "" {
		return ""
	}

	// First, handle common separators
	words := strings.FieldsFunc(s, func(c rune) bool {
		return c == '_' || c == '-' || c == ' ' || c == '.'
	})

	// If no separators found, try to split camelCase
	if len(words) == 1 {
		camelWords := splitCamelCase(words[0])
		if len(camelWords) > 1 {
			words = camelWords
		}
	}

	if len(words) == 0 {
		return s
	}

	var result strings.Builder
	for _, word := range words {
		if len(word) > 0 {
			// Capitalize first letter and make rest lowercase
			result.WriteString(strings.ToUpper(string(word[0])))
			if len(word) > 1 {
				result.WriteString(strings.ToLower(word[1:]))
			}
		}
	}

	return result.String()
}

// splitCamelCase splits a camelCase string into words
func splitCamelCase(s string) []string {
	if s == "" {
		return nil
	}

	var words []string
	var currentWord strings.Builder

	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			// Found a capital letter, start a new word
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
		}
		currentWord.WriteRune(r)
	}

	// Add the last word
	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	return words
}

// ToJSONTag creates a JSON tag for a field name
func ToJSONTag(fieldName string) string {
	return fmt.Sprintf(`json:"%s"`, fieldName)
}

// isEnumProperty checks if a property has enum values defined
func (tm *DefaultTypeMapper) isEnumProperty(prop *Property) bool {
	return prop != nil && len(prop.Enum) > 0
}

// generateEnumTypeName creates a Go type name for an enum based on the field name
func (tm *DefaultTypeMapper) generateEnumTypeName(fieldName string) string {
	return ToPascalCase(fieldName) + "Enum"
}

// getEnumBaseType determines the base Go type for an enum based on its values
func (tm *DefaultTypeMapper) getEnumBaseType(enumValues []interface{}) string {
	if len(enumValues) == 0 {
		return "string" // default to string
	}

	// Check the type of the first non-nil value to determine the base type
	for _, value := range enumValues {
		if value == nil {
			continue
		}

		switch value.(type) {
		case string:
			return "string"
		case int, int32, int64, float64:
			// For numeric enums, we'll use int for integers and float64 for floats
			if _, ok := value.(float64); ok {
				// Check if it's actually an integer stored as float64 (common in JSON)
				if f, ok := value.(float64); ok && f == float64(int64(f)) {
					return "int"
				}
				return "float64"
			}
			return "int"
		case bool:
			return "bool"
		default:
			return "string" // fallback to string for unknown types
		}
	}

	return "string" // fallback if all values are nil
}

// MapPropertyWithContext maps a property with additional context about whether it's required
func (tm *DefaultTypeMapper) MapPropertyWithContext(prop *Property, fieldName string, required bool) *GoField {
	if prop == nil {
		return nil
	}

	field := &GoField{
		Comment: prop.Description,
	}

	// Check if this is an enum property first
	if tm.isEnumProperty(prop) {
		// Generate enum type name
		enumTypeName := tm.generateEnumTypeName(fieldName)
		field.Type = enumTypeName
		field.IsEnum = true
		field.EnumValues = prop.Enum
		field.EnumBaseType = tm.getEnumBaseType(prop.Enum)
	} else if prop.Type == "array" && prop.Items != nil {
		// Check if array items are enums
		if tm.isEnumProperty(prop.Items) {
			enumTypeName := tm.generateEnumTypeName(fieldName + "Item")
			field.Type = "[]" + enumTypeName
			field.IsEnum = false // The field itself is not an enum, but contains enums
			field.ArrayItemEnum = &EnumInfo{
				TypeName: enumTypeName,
				Values:   prop.Items.Enum,
				BaseType: tm.getEnumBaseType(prop.Items.Enum),
			}
		} else {
			itemType := tm.getPropertyType(prop.Items)
			field.Type = "[]" + itemType
		}
	} else if prop.Type == "object" && prop.Properties != nil {
		// For nested objects, we'll use a struct type name
		// This will be handled by the code generator to create nested structs
		field.Type = "struct{}" // Placeholder, will be replaced by actual struct name
	} else {
		// Handle primitive types
		field.Type = tm.MapType(prop.Type, prop.Format)
	}

	// Set the field name in PascalCase
	field.Name = ToPascalCase(fieldName)

	// Set JSON tag with original field name
	field.JSONTag = ToJSONTag(fieldName)

	// Set required status with context
	field.Optional = !required
	if field.Optional && tm.config != nil && tm.config.UsePointers {
		// Make optional fields pointers if configured to do so
		if !strings.HasPrefix(field.Type, "*") {
			field.Type = "*" + field.Type
		}
	}

	return field
}
