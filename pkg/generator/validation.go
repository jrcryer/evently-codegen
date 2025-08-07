package generator

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// ValidationErrorWithValue extends ValidationError with the actual value that failed validation.
// This is useful for debugging and providing more context in error messages.
type ValidationErrorWithValue struct {
	*ValidationError
	// Value is the actual value that failed validation
	Value any `json:"value,omitempty"`
}

// ValidationResult contains the complete result of a validation operation.
// It includes the validation status and any errors that were encountered.
//
// Example usage:
//
//	result := validator.ValidateJSON(jsonData, schema)
//	if !result.Valid {
//	    for _, err := range result.Errors {
//	        fmt.Printf("Validation error in field '%s': %s\n", err.Field, err.Message)
//	    }
//	}
type ValidationResult struct {
	// Valid indicates whether the validation passed (true) or failed (false)
	Valid bool `json:"valid"`

	// Errors contains all validation errors found during validation.
	// This slice is empty when Valid is true.
	Errors []*ValidationError `json:"errors,omitempty"`
}

// AddError adds a validation error to the result
func (r *ValidationResult) AddError(field, message string, value any) {
	r.Valid = false
	r.Errors = append(r.Errors, &ValidationError{
		Field:   field,
		Message: message,
	})
}

// SchemaValidator implements the Validator interface and provides JSON schema validation
// capabilities for AsyncAPI schemas. It validates data against AsyncAPI/JSON Schema
// constraints including type validation, constraint validation, enum validation,
// and required field validation.
//
// The validator supports two modes:
//   - Strict mode: Rejects additional properties not defined in the schema
//   - Permissive mode: Allows additional properties (default behavior)
//
// Example usage:
//
//	// Create a permissive validator (allows extra properties)
//	validator := NewValidator(false)
//
//	// Create a strict validator (rejects extra properties)
//	strictValidator := NewValidator(true)
//
//	// Validate JSON data
//	result := validator.ValidateJSON(jsonData, schema)
//	if !result.Valid {
//	    for _, err := range result.Errors {
//	        fmt.Printf("Field '%s': %s\n", err.Field, err.Message)
//	    }
//	}
type SchemaValidator struct {
	// StrictMode determines whether additional properties are allowed.
	// When true, the validator will reject any properties not defined in the schema.
	// When false, additional properties are allowed (permissive mode).
	StrictMode bool
}

// NewValidator creates a new SchemaValidator instance with the specified strict mode setting.
//
// Parameters:
//   - strictMode: If true, the validator will reject additional properties not defined in the schema.
//     If false, additional properties are allowed (permissive mode).
//
// Returns:
//   - A Validator interface implementation that can be used to validate data against AsyncAPI schemas.
//
// Example:
//
//	// Create a permissive validator
//	validator := NewValidator(false)
//
//	// Create a strict validator
//	strictValidator := NewValidator(true)
func NewValidator(strictMode bool) Validator {
	return &SchemaValidator{
		StrictMode: strictMode,
	}
}

// ValidateJSON validates raw JSON data against a MessageSchema.
// This method first parses the JSON data and then validates it against the provided schema.
// It handles JSON parsing errors as well as schema validation errors.
//
// Parameters:
//   - jsonData: Raw JSON data as bytes to be validated
//   - schema: The MessageSchema to validate against
//
// Returns:
//   - ValidationResult containing validation status and any errors found
//
// The method performs the following validations:
//   - JSON syntax validation (parsing)
//   - Type validation (string, number, boolean, array, object)
//   - Constraint validation (min/max length, numeric ranges, patterns)
//   - Enum validation
//   - Required field validation
//   - Additional property validation (based on StrictMode)
//
// Example:
//
//	jsonData := []byte(`{"name": "John", "age": 30}`)
//	result := validator.ValidateJSON(jsonData, schema)
//	if !result.Valid {
//	    for _, err := range result.Errors {
//	        fmt.Printf("Error in %s: %s\n", err.Field, err.Message)
//	    }
//	}
func (v *SchemaValidator) ValidateJSON(jsonData []byte, schema *MessageSchema) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Parse JSON data
	var data any
	if err := json.Unmarshal(jsonData, &data); err != nil {
		result.AddError("", fmt.Sprintf("invalid JSON: %v", err), nil)
		return result
	}

	// Convert MessageSchema to Property for validation
	prop := v.messageSchemaToProperty(schema)
	return v.ValidateValue(data, prop, "")
}

// ValidateMessage validates a message payload against its MessageSchema.
// The data should be a Go value (typically map[string]any from JSON unmarshaling)
// that represents the message payload.
//
// Parameters:
//   - data: The data to validate (typically map[string]any, but can be any Go type)
//   - message: The MessageSchema to validate against
//
// Returns:
//   - ValidationResult containing validation status and any errors found
//
// This method is useful when you already have parsed data (e.g., from JSON unmarshaling)
// and want to validate it against a schema without re-parsing JSON.
//
// Example:
//
//	var data map[string]any
//	json.Unmarshal(jsonBytes, &data)
//	result := validator.ValidateMessage(data, schema)
func (v *SchemaValidator) ValidateMessage(data any, message *MessageSchema) *ValidationResult {
	prop := v.messageSchemaToProperty(message)
	return v.ValidateValue(data, prop, "")
}

// ValidateValue validates a Go value against a Property schema.
// This is the core validation method that performs detailed validation
// of individual values against their schema constraints.
//
// Parameters:
//   - value: The value to validate (can be any Go type)
//   - schema: The Property schema defining validation rules
//   - fieldPath: The path to this field for error reporting (e.g., "user.profile.age")
//
// Returns:
//   - ValidationResult containing validation status and any errors found
//
// The method validates:
//   - Type correctness (string, number, boolean, array, object)
//   - String constraints (minLength, maxLength, pattern)
//   - Numeric constraints (minimum, maximum, multipleOf)
//   - Array constraints (minItems, maxItems, uniqueItems)
//   - Object constraints (required properties, additional properties)
//   - Enum constraints
//   - Const constraints
//
// Example:
//
//	schema := &Property{Type: "string", MinLength: &[]int{3}[0]}
//	result := validator.ValidateValue("hi", schema, "name")
//	// result.Valid will be false because "hi" is shorter than 3 characters
func (v *SchemaValidator) ValidateValue(value any, schema *Property, fieldPath string) *ValidationResult {
	result := &ValidationResult{Valid: true}

	if schema == nil {
		return result
	}

	// Handle null values
	if value == nil {
		// Null is only valid if the schema allows it or has no type constraint
		if schema.Type != "" && schema.Type != "null" {
			// Check if null is allowed by enum or const before rejecting
			if len(schema.Enum) == 0 && schema.Const == nil {
				result.AddError(fieldPath, "value cannot be null", value)
				return result
			}
		}
		// Continue to validate enum/const constraints even for null values
	}

	// Validate based on schema type (skip for null values)
	if value != nil {
		switch schema.Type {
		case "string":
			v.validateString(value, schema, fieldPath, result)
		case "number", "integer":
			v.validateNumber(value, schema, fieldPath, result)
		case "boolean":
			v.validateBoolean(value, schema, fieldPath, result)
		case "array":
			v.validateArray(value, schema, fieldPath, result)
		case "object":
			v.validateObject(value, schema, fieldPath, result)
		case "":
			// No type specified, try to infer and validate
			v.validateAnyType(value, schema, fieldPath, result)
		default:
			result.AddError(fieldPath, fmt.Sprintf("unsupported schema type: %s", schema.Type), value)
		}
	}

	// Validate enum constraints
	if len(schema.Enum) > 0 {
		v.validateEnum(value, schema, fieldPath, result)
	}

	// Validate const constraints
	v.validateConst(value, schema, fieldPath, result)

	return result
}

// validateString validates string values
func (v *SchemaValidator) validateString(value any, schema *Property, fieldPath string, result *ValidationResult) {
	str, ok := value.(string)
	if !ok {
		result.AddError(fieldPath, fmt.Sprintf("expected string, got %T", value), value)
		return
	}

	// Length constraints
	if schema.MinLength != nil && len(str) < *schema.MinLength {
		result.AddError(fieldPath, fmt.Sprintf("string length %d is less than minimum %d", len(str), *schema.MinLength), value)
	}

	if schema.MaxLength != nil && len(str) > *schema.MaxLength {
		result.AddError(fieldPath, fmt.Sprintf("string length %d exceeds maximum %d", len(str), *schema.MaxLength), value)
	}

	// Pattern constraint
	if schema.Pattern != "" {
		matched, err := regexp.MatchString(schema.Pattern, str)
		if err != nil {
			result.AddError(fieldPath, fmt.Sprintf("invalid regex pattern: %v", err), value)
		} else if !matched {
			result.AddError(fieldPath, fmt.Sprintf("string does not match pattern: %s", schema.Pattern), value)
		}
	}
}

// validateNumber validates numeric values
func (v *SchemaValidator) validateNumber(value any, schema *Property, fieldPath string, result *ValidationResult) {
	var num float64
	var isInteger bool

	switch val := value.(type) {
	case int:
		num = float64(val)
		isInteger = true
	case int32:
		num = float64(val)
		isInteger = true
	case int64:
		num = float64(val)
		isInteger = true
	case float32:
		num = float64(val)
	case float64:
		num = val
	case json.Number:
		var err error
		if schema.Type == "integer" {
			intVal, err := val.Int64()
			if err != nil {
				result.AddError(fieldPath, fmt.Sprintf("expected integer, got %v", val), value)
				return
			}
			num = float64(intVal)
			isInteger = true
		} else {
			num, err = val.Float64()
			if err != nil {
				result.AddError(fieldPath, fmt.Sprintf("expected number, got %v", val), value)
				return
			}
		}
	default:
		result.AddError(fieldPath, fmt.Sprintf("expected number, got %T", value), value)
		return
	}

	// Integer type validation
	if schema.Type == "integer" && !isInteger && num != float64(int64(num)) {
		result.AddError(fieldPath, "expected integer value", value)
	}

	// Numeric constraints
	if schema.Minimum != nil && num < *schema.Minimum {
		result.AddError(fieldPath, fmt.Sprintf("value %g is less than minimum %g", num, *schema.Minimum), value)
	}

	if schema.Maximum != nil && num > *schema.Maximum {
		result.AddError(fieldPath, fmt.Sprintf("value %g exceeds maximum %g", num, *schema.Maximum), value)
	}

	if schema.ExclusiveMinimum != nil && num <= *schema.ExclusiveMinimum {
		result.AddError(fieldPath, fmt.Sprintf("value %g is not greater than exclusive minimum %g", num, *schema.ExclusiveMinimum), value)
	}

	if schema.ExclusiveMaximum != nil && num >= *schema.ExclusiveMaximum {
		result.AddError(fieldPath, fmt.Sprintf("value %g is not less than exclusive maximum %g", num, *schema.ExclusiveMaximum), value)
	}

	if schema.MultipleOf != nil && *schema.MultipleOf != 0 {
		if remainder := num / *schema.MultipleOf; remainder != float64(int64(remainder)) {
			result.AddError(fieldPath, fmt.Sprintf("value %g is not a multiple of %g", num, *schema.MultipleOf), value)
		}
	}
}

// validateBoolean validates boolean values
func (v *SchemaValidator) validateBoolean(value any, schema *Property, fieldPath string, result *ValidationResult) {
	if _, ok := value.(bool); !ok {
		result.AddError(fieldPath, fmt.Sprintf("expected boolean, got %T", value), value)
	}
}

// validateArray validates array values
func (v *SchemaValidator) validateArray(value any, schema *Property, fieldPath string, result *ValidationResult) {
	arr, ok := value.([]any)
	if !ok {
		// Try to handle different slice types
		rv := reflect.ValueOf(value)
		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			result.AddError(fieldPath, fmt.Sprintf("expected array, got %T", value), value)
			return
		}

		// Convert to []any
		arr = make([]any, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			arr[i] = rv.Index(i).Interface()
		}
	}

	// Length constraints
	if schema.MinItems != nil && len(arr) < *schema.MinItems {
		result.AddError(fieldPath, fmt.Sprintf("array length %d is less than minimum %d", len(arr), *schema.MinItems), value)
	}

	if schema.MaxItems != nil && len(arr) > *schema.MaxItems {
		result.AddError(fieldPath, fmt.Sprintf("array length %d exceeds maximum %d", len(arr), *schema.MaxItems), value)
	}

	// Unique items constraint
	if schema.UniqueItems != nil && *schema.UniqueItems {
		seen := make(map[string]bool)
		for i, item := range arr {
			key := fmt.Sprintf("%v", item)
			if seen[key] {
				itemPath := fmt.Sprintf("%s[%d]", fieldPath, i)
				result.AddError(itemPath, "duplicate item in array with uniqueItems constraint", item)
			}
			seen[key] = true
		}
	}

	// Validate items
	if schema.Items != nil {
		for i, item := range arr {
			itemPath := fmt.Sprintf("%s[%d]", fieldPath, i)
			itemResult := v.ValidateValue(item, schema.Items, itemPath)
			if !itemResult.Valid {
				result.Valid = false
				result.Errors = append(result.Errors, itemResult.Errors...)
			}
		}
	}
}

// validateObject validates object values
func (v *SchemaValidator) validateObject(value any, schema *Property, fieldPath string, result *ValidationResult) {
	obj, ok := value.(map[string]any)
	if !ok {
		result.AddError(fieldPath, fmt.Sprintf("expected object, got %T", value), value)
		return
	}

	// Property count constraints
	if schema.MinProperties != nil && len(obj) < *schema.MinProperties {
		result.AddError(fieldPath, fmt.Sprintf("object has %d properties, minimum is %d", len(obj), *schema.MinProperties), value)
	}

	if schema.MaxProperties != nil && len(obj) > *schema.MaxProperties {
		result.AddError(fieldPath, fmt.Sprintf("object has %d properties, maximum is %d", len(obj), *schema.MaxProperties), value)
	}

	// Validate required properties
	for _, required := range schema.Required {
		if _, exists := obj[required]; !exists {
			propPath := v.buildFieldPath(fieldPath, required)
			result.AddError(propPath, "required property is missing", nil)
		}
	}

	// Validate defined properties
	if schema.Properties != nil {
		for propName, propSchema := range schema.Properties {
			if propValue, exists := obj[propName]; exists {
				propPath := v.buildFieldPath(fieldPath, propName)
				propResult := v.ValidateValue(propValue, propSchema, propPath)
				if !propResult.Valid {
					result.Valid = false
					result.Errors = append(result.Errors, propResult.Errors...)
				}
			}
		}
	}

	// Handle additional properties
	if v.StrictMode {
		for propName := range obj {
			if schema.Properties == nil || schema.Properties[propName] == nil {
				propPath := v.buildFieldPath(fieldPath, propName)
				result.AddError(propPath, "additional property not allowed in strict mode", obj[propName])
			}
		}
	}
}

// validateAnyType validates values when no specific type is defined
func (v *SchemaValidator) validateAnyType(value any, schema *Property, fieldPath string, result *ValidationResult) {
	// Try to infer type and validate accordingly
	switch value.(type) {
	case string:
		v.validateString(value, schema, fieldPath, result)
	case int, int32, int64, float32, float64, json.Number:
		v.validateNumber(value, schema, fieldPath, result)
	case bool:
		v.validateBoolean(value, schema, fieldPath, result)
	case []any:
		v.validateArray(value, schema, fieldPath, result)
	case map[string]any:
		v.validateObject(value, schema, fieldPath, result)
	}
}

// validateEnum validates enum constraints
func (v *SchemaValidator) validateEnum(value any, schema *Property, fieldPath string, result *ValidationResult) {
	for _, enumValue := range schema.Enum {
		if reflect.DeepEqual(value, enumValue) {
			return
		}
	}

	enumStrs := make([]string, len(schema.Enum))
	for i, ev := range schema.Enum {
		enumStrs[i] = fmt.Sprintf("%v", ev)
	}
	result.AddError(fieldPath, fmt.Sprintf("value must be one of: %s", strings.Join(enumStrs, ", ")), value)
}

// validateConst validates const constraints
func (v *SchemaValidator) validateConst(value any, schema *Property, fieldPath string, result *ValidationResult) {
	// For this implementation, we validate const when:
	// 1. Const is not nil, OR
	// 2. Type is "null" (indicating null is the expected constant value)
	if schema.Const != nil || schema.Type == "null" {
		if !reflect.DeepEqual(value, schema.Const) {
			result.AddError(fieldPath, fmt.Sprintf("value must be: %v", schema.Const), value)
		}
	}
}

// buildFieldPath constructs a field path for error reporting
func (v *SchemaValidator) buildFieldPath(parentPath, fieldName string) string {
	if parentPath == "" {
		return fieldName
	}
	return parentPath + "." + fieldName
}

// messageSchemaToProperty converts a MessageSchema to a Property for validation
func (v *SchemaValidator) messageSchemaToProperty(schema *MessageSchema) *Property {
	return &Property{
		Ref:                  schema.Ref,
		ID:                   schema.ID,
		Schema:               schema.Schema,
		Title:                schema.Title,
		Description:          schema.Description,
		Default:              schema.Default,
		Examples:             schema.Examples,
		Type:                 schema.Type,
		Enum:                 schema.Enum,
		Const:                schema.Const,
		MultipleOf:           schema.MultipleOf,
		Maximum:              schema.Maximum,
		ExclusiveMaximum:     schema.ExclusiveMaximum,
		Minimum:              schema.Minimum,
		ExclusiveMinimum:     schema.ExclusiveMinimum,
		MaxLength:            schema.MaxLength,
		MinLength:            schema.MinLength,
		Pattern:              schema.Pattern,
		Format:               schema.Format,
		Items:                schema.Items,
		AdditionalItems:      schema.AdditionalItems,
		MaxItems:             schema.MaxItems,
		MinItems:             schema.MinItems,
		UniqueItems:          schema.UniqueItems,
		Properties:           schema.Properties,
		PatternProperties:    schema.PatternProperties,
		AdditionalProperties: schema.AdditionalProperties,
		Required:             schema.Required,
		PropertyNames:        schema.PropertyNames,
		MaxProperties:        schema.MaxProperties,
		MinProperties:        schema.MinProperties,
		AllOf:                schema.AllOf,
		AnyOf:                schema.AnyOf,
		OneOf:                schema.OneOf,
		Not:                  schema.Not,
		If:                   schema.If,
		Then:                 schema.Then,
		Else:                 schema.Else,
		ReadOnly:             schema.ReadOnly,
		WriteOnly:            schema.WriteOnly,
	}
}
