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
		return "[]interface{}" // Will be refined by MapProperty
	case "object":
		return "interface{}" // Will be refined by MapProperty for nested objects
	case "null":
		return "interface{}"
	default:
		// Unknown type, default to interface{}
		return "interface{}"
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
		return "interface{}"
	}

	if prop.Type == "array" && prop.Items != nil {
		itemType := tm.getPropertyType(prop.Items)
		return "[]" + itemType
	}

	if prop.Type == "object" {
		return "interface{}" // Will be refined for nested objects
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

	// Handle common separators
	words := strings.FieldsFunc(s, func(c rune) bool {
		return c == '_' || c == '-' || c == ' ' || c == '.'
	})

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

// ToJSONTag creates a JSON tag for a field name
func ToJSONTag(fieldName string) string {
	return fmt.Sprintf(`json:"%s"`, fieldName)
}

// MapPropertyWithContext maps a property with additional context about whether it's required
func (tm *DefaultTypeMapper) MapPropertyWithContext(prop *Property, fieldName string, required bool) *GoField {
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
