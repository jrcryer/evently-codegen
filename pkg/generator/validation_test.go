package generator

import (
	"strings"
	"testing"
)

func TestSchemaValidator_ValidateString(t *testing.T) {
	validator := NewValidator(false)

	tests := []struct {
		name     string
		value    any
		schema   *Property
		wantErr  bool
		errField string
	}{
		{
			name:  "valid string",
			value: "hello",
			schema: &Property{
				Type: "string",
			},
			wantErr: false,
		},
		{
			name:  "invalid type",
			value: 123,
			schema: &Property{
				Type: "string",
			},
			wantErr:  true,
			errField: "test",
		},
		{
			name:  "string too short",
			value: "hi",
			schema: &Property{
				Type:      "string",
				MinLength: intPtr(5),
			},
			wantErr:  true,
			errField: "test",
		},
		{
			name:  "string too long",
			value: "hello world",
			schema: &Property{
				Type:      "string",
				MaxLength: intPtr(5),
			},
			wantErr:  true,
			errField: "test",
		},
		{
			name:  "valid pattern",
			value: "test@example.com",
			schema: &Property{
				Type:    "string",
				Pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			},
			wantErr: false,
		},
		{
			name:  "invalid pattern",
			value: "invalid-email",
			schema: &Property{
				Type:    "string",
				Pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			},
			wantErr:  true,
			errField: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateValue(tt.value, tt.schema, "test")

			if tt.wantErr && result.Valid {
				t.Errorf("expected validation error, but validation passed")
			}

			if !tt.wantErr && !result.Valid {
				t.Errorf("expected validation to pass, but got errors: %v", result.Errors)
			}

			if tt.wantErr && len(result.Errors) > 0 && result.Errors[0].Field != tt.errField {
				t.Errorf("expected error field %s, got %s", tt.errField, result.Errors[0].Field)
			}
		})
	}
}

func TestSchemaValidator_ValidateNumber(t *testing.T) {
	validator := NewValidator(false)

	tests := []struct {
		name     string
		value    any
		schema   *Property
		wantErr  bool
		errField string
	}{
		{
			name:  "valid integer",
			value: 42,
			schema: &Property{
				Type: "integer",
			},
			wantErr: false,
		},
		{
			name:  "valid number",
			value: 3.14,
			schema: &Property{
				Type: "number",
			},
			wantErr: false,
		},
		{
			name:  "float as integer should fail",
			value: 3.14,
			schema: &Property{
				Type: "integer",
			},
			wantErr:  true,
			errField: "test",
		},
		{
			name:  "number below minimum",
			value: 5,
			schema: &Property{
				Type:    "number",
				Minimum: float64Ptr(10),
			},
			wantErr:  true,
			errField: "test",
		},
		{
			name:  "number above maximum",
			value: 15,
			schema: &Property{
				Type:    "number",
				Maximum: float64Ptr(10),
			},
			wantErr:  true,
			errField: "test",
		},
		{
			name:  "valid multiple",
			value: 15,
			schema: &Property{
				Type:       "number",
				MultipleOf: float64Ptr(5),
			},
			wantErr: false,
		},
		{
			name:  "invalid multiple",
			value: 17,
			schema: &Property{
				Type:       "number",
				MultipleOf: float64Ptr(5),
			},
			wantErr:  true,
			errField: "test",
		},
		{
			name:  "invalid type",
			value: "not a number",
			schema: &Property{
				Type: "number",
			},
			wantErr:  true,
			errField: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateValue(tt.value, tt.schema, "test")

			if tt.wantErr && result.Valid {
				t.Errorf("expected validation error, but validation passed")
			}

			if !tt.wantErr && !result.Valid {
				t.Errorf("expected validation to pass, but got errors: %v", result.Errors)
			}
		})
	}
}

func TestSchemaValidator_ValidateBoolean(t *testing.T) {
	validator := NewValidator(false)

	tests := []struct {
		name    string
		value   any
		schema  *Property
		wantErr bool
	}{
		{
			name:  "valid true",
			value: true,
			schema: &Property{
				Type: "boolean",
			},
			wantErr: false,
		},
		{
			name:  "valid false",
			value: false,
			schema: &Property{
				Type: "boolean",
			},
			wantErr: false,
		},
		{
			name:  "invalid type",
			value: "true",
			schema: &Property{
				Type: "boolean",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateValue(tt.value, tt.schema, "test")

			if tt.wantErr && result.Valid {
				t.Errorf("expected validation error, but validation passed")
			}

			if !tt.wantErr && !result.Valid {
				t.Errorf("expected validation to pass, but got errors: %v", result.Errors)
			}
		})
	}
}

func TestSchemaValidator_ValidateArray(t *testing.T) {
	validator := NewValidator(false)

	tests := []struct {
		name    string
		value   any
		schema  *Property
		wantErr bool
	}{
		{
			name:  "valid array",
			value: []any{"a", "b", "c"},
			schema: &Property{
				Type: "array",
				Items: &Property{
					Type: "string",
				},
			},
			wantErr: false,
		},
		{
			name:  "array too short",
			value: []any{"a"},
			schema: &Property{
				Type:     "array",
				MinItems: intPtr(2),
			},
			wantErr: true,
		},
		{
			name:  "array too long",
			value: []any{"a", "b", "c"},
			schema: &Property{
				Type:     "array",
				MaxItems: intPtr(2),
			},
			wantErr: true,
		},
		{
			name:  "unique items valid",
			value: []any{"a", "b", "c"},
			schema: &Property{
				Type:        "array",
				UniqueItems: boolPtr(true),
			},
			wantErr: false,
		},
		{
			name:  "unique items invalid",
			value: []any{"a", "b", "a"},
			schema: &Property{
				Type:        "array",
				UniqueItems: boolPtr(true),
			},
			wantErr: true,
		},
		{
			name:  "invalid item type",
			value: []any{"a", 123, "c"},
			schema: &Property{
				Type: "array",
				Items: &Property{
					Type: "string",
				},
			},
			wantErr: true,
		},
		{
			name:  "invalid type",
			value: "not an array",
			schema: &Property{
				Type: "array",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateValue(tt.value, tt.schema, "test")

			if tt.wantErr && result.Valid {
				t.Errorf("expected validation error, but validation passed")
			}

			if !tt.wantErr && !result.Valid {
				t.Errorf("expected validation to pass, but got errors: %v", result.Errors)
			}
		})
	}
}

func TestSchemaValidator_ValidateObject(t *testing.T) {
	validator := NewValidator(false)

	tests := []struct {
		name    string
		value   any
		schema  *Property
		wantErr bool
	}{
		{
			name: "valid object",
			value: map[string]any{
				"name": "John",
				"age":  30,
			},
			schema: &Property{
				Type: "object",
				Properties: map[string]*Property{
					"name": {Type: "string"},
					"age":  {Type: "number"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing required property",
			value: map[string]any{
				"age": 30,
			},
			schema: &Property{
				Type: "object",
				Properties: map[string]*Property{
					"name": {Type: "string"},
					"age":  {Type: "number"},
				},
				Required: []string{"name"},
			},
			wantErr: true,
		},
		{
			name: "invalid property type",
			value: map[string]any{
				"name": 123,
				"age":  30,
			},
			schema: &Property{
				Type: "object",
				Properties: map[string]*Property{
					"name": {Type: "string"},
					"age":  {Type: "number"},
				},
			},
			wantErr: true,
		},
		{
			name: "too few properties",
			value: map[string]any{
				"name": "John",
			},
			schema: &Property{
				Type:          "object",
				MinProperties: intPtr(2),
			},
			wantErr: true,
		},
		{
			name: "too many properties",
			value: map[string]any{
				"name": "John",
				"age":  30,
				"city": "NYC",
			},
			schema: &Property{
				Type:          "object",
				MaxProperties: intPtr(2),
			},
			wantErr: true,
		},
		{
			name:  "invalid type",
			value: "not an object",
			schema: &Property{
				Type: "object",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateValue(tt.value, tt.schema, "test")

			if tt.wantErr && result.Valid {
				t.Errorf("expected validation error, but validation passed")
			}

			if !tt.wantErr && !result.Valid {
				t.Errorf("expected validation to pass, but got errors: %v", result.Errors)
			}
		})
	}
}

func TestSchemaValidator_ValidateEnum(t *testing.T) {
	validator := NewValidator(false)

	tests := []struct {
		name    string
		value   any
		schema  *Property
		wantErr bool
	}{
		{
			name:  "valid enum value",
			value: "red",
			schema: &Property{
				Type: "string",
				Enum: []any{"red", "green", "blue"},
			},
			wantErr: false,
		},
		{
			name:  "invalid enum value",
			value: "yellow",
			schema: &Property{
				Type: "string",
				Enum: []any{"red", "green", "blue"},
			},
			wantErr: true,
		},
		{
			name:  "valid numeric enum",
			value: 2,
			schema: &Property{
				Type: "integer",
				Enum: []any{1, 2, 3},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateValue(tt.value, tt.schema, "test")

			if tt.wantErr && result.Valid {
				t.Errorf("expected validation error, but validation passed")
			}

			if !tt.wantErr && !result.Valid {
				t.Errorf("expected validation to pass, but got errors: %v", result.Errors)
			}
		})
	}
}

func TestSchemaValidator_ValidateJSON(t *testing.T) {
	validator := NewValidator(false)

	schema := &MessageSchema{
		Type: "object",
		Properties: map[string]*Property{
			"name": {Type: "string"},
			"age":  {Type: "integer", Minimum: float64Ptr(0)},
			"email": {
				Type:    "string",
				Pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			},
		},
		Required: []string{"name", "email"},
	}

	tests := []struct {
		name     string
		jsonData string
		wantErr  bool
	}{
		{
			name:     "valid JSON",
			jsonData: `{"name": "John", "age": 30, "email": "john@example.com"}`,
			wantErr:  false,
		},
		{
			name:     "missing required field",
			jsonData: `{"age": 30}`,
			wantErr:  true,
		},
		{
			name:     "invalid email pattern",
			jsonData: `{"name": "John", "email": "invalid-email"}`,
			wantErr:  true,
		},
		{
			name:     "negative age",
			jsonData: `{"name": "John", "age": -5, "email": "john@example.com"}`,
			wantErr:  true,
		},
		{
			name:     "invalid JSON syntax",
			jsonData: `{"name": "John", "age": 30,}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateJSON([]byte(tt.jsonData), schema)

			if tt.wantErr && result.Valid {
				t.Errorf("expected validation error, but validation passed")
			}

			if !tt.wantErr && !result.Valid {
				t.Errorf("expected validation to pass, but got errors: %v", result.Errors)
			}
		})
	}
}

func TestSchemaValidator_StrictMode(t *testing.T) {
	strictValidator := NewValidator(true)
	permissiveValidator := NewValidator(false)

	schema := &Property{
		Type: "object",
		Properties: map[string]*Property{
			"name": {Type: "string"},
		},
	}

	value := map[string]any{
		"name":  "John",
		"extra": "not allowed",
	}

	// Test strict mode
	strictResult := strictValidator.ValidateValue(value, schema, "test")
	if strictResult.Valid {
		t.Errorf("expected validation error in strict mode for additional property")
	}

	// Test permissive mode
	permissiveResult := permissiveValidator.ValidateValue(value, schema, "test")
	if !permissiveResult.Valid {
		t.Errorf("expected validation to pass in permissive mode, got errors: %v", permissiveResult.Errors)
	}
}

// Helper functions for creating pointers
func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

func boolPtr(b bool) *bool {
	return &b
}

func TestSchemaValidator_StringConstraints(t *testing.T) {
	validator := NewValidator(false)

	tests := []struct {
		name    string
		value   any
		schema  *Property
		wantErr bool
	}{
		{
			name:  "minLength constraint valid",
			value: "hello",
			schema: &Property{
				Type:      "string",
				MinLength: intPtr(3),
			},
			wantErr: false,
		},
		{
			name:  "minLength constraint invalid",
			value: "hi",
			schema: &Property{
				Type:      "string",
				MinLength: intPtr(5),
			},
			wantErr: true,
		},
		{
			name:  "maxLength constraint valid",
			value: "hello",
			schema: &Property{
				Type:      "string",
				MaxLength: intPtr(10),
			},
			wantErr: false,
		},
		{
			name:  "maxLength constraint invalid",
			value: "hello world",
			schema: &Property{
				Type:      "string",
				MaxLength: intPtr(5),
			},
			wantErr: true,
		},
		{
			name:  "pattern constraint valid email",
			value: "user@example.com",
			schema: &Property{
				Type:    "string",
				Pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			},
			wantErr: false,
		},
		{
			name:  "pattern constraint invalid email",
			value: "invalid-email",
			schema: &Property{
				Type:    "string",
				Pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			},
			wantErr: true,
		},
		{
			name:  "pattern constraint valid phone",
			value: "+1-555-123-4567",
			schema: &Property{
				Type:    "string",
				Pattern: `^\+\d{1,3}-\d{3}-\d{3}-\d{4}$`,
			},
			wantErr: false,
		},
		{
			name:  "pattern constraint invalid phone",
			value: "555-123-4567",
			schema: &Property{
				Type:    "string",
				Pattern: `^\+\d{1,3}-\d{3}-\d{3}-\d{4}$`,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateValue(tt.value, tt.schema, "test")

			if tt.wantErr && result.Valid {
				t.Errorf("expected validation error, but validation passed")
			}

			if !tt.wantErr && !result.Valid {
				t.Errorf("expected validation to pass, but got errors: %v", result.Errors)
			}
		})
	}
}

func TestSchemaValidator_NumericConstraints(t *testing.T) {
	validator := NewValidator(false)

	tests := []struct {
		name    string
		value   any
		schema  *Property
		wantErr bool
	}{
		{
			name:  "minimum constraint valid",
			value: 10,
			schema: &Property{
				Type:    "number",
				Minimum: float64Ptr(5),
			},
			wantErr: false,
		},
		{
			name:  "minimum constraint invalid",
			value: 3,
			schema: &Property{
				Type:    "number",
				Minimum: float64Ptr(5),
			},
			wantErr: true,
		},
		{
			name:  "maximum constraint valid",
			value: 8,
			schema: &Property{
				Type:    "number",
				Maximum: float64Ptr(10),
			},
			wantErr: false,
		},
		{
			name:  "maximum constraint invalid",
			value: 15,
			schema: &Property{
				Type:    "number",
				Maximum: float64Ptr(10),
			},
			wantErr: true,
		},
		{
			name:  "exclusiveMinimum constraint valid",
			value: 6,
			schema: &Property{
				Type:             "number",
				ExclusiveMinimum: float64Ptr(5),
			},
			wantErr: false,
		},
		{
			name:  "exclusiveMinimum constraint invalid equal",
			value: 5,
			schema: &Property{
				Type:             "number",
				ExclusiveMinimum: float64Ptr(5),
			},
			wantErr: true,
		},
		{
			name:  "exclusiveMinimum constraint invalid less",
			value: 4,
			schema: &Property{
				Type:             "number",
				ExclusiveMinimum: float64Ptr(5),
			},
			wantErr: true,
		},
		{
			name:  "exclusiveMaximum constraint valid",
			value: 9,
			schema: &Property{
				Type:             "number",
				ExclusiveMaximum: float64Ptr(10),
			},
			wantErr: false,
		},
		{
			name:  "exclusiveMaximum constraint invalid equal",
			value: 10,
			schema: &Property{
				Type:             "number",
				ExclusiveMaximum: float64Ptr(10),
			},
			wantErr: true,
		},
		{
			name:  "exclusiveMaximum constraint invalid greater",
			value: 11,
			schema: &Property{
				Type:             "number",
				ExclusiveMaximum: float64Ptr(10),
			},
			wantErr: true,
		},
		{
			name:  "multipleOf constraint valid",
			value: 15,
			schema: &Property{
				Type:       "number",
				MultipleOf: float64Ptr(5),
			},
			wantErr: false,
		},
		{
			name:  "multipleOf constraint invalid",
			value: 17,
			schema: &Property{
				Type:       "number",
				MultipleOf: float64Ptr(5),
			},
			wantErr: true,
		},
		{
			name:  "multipleOf constraint with decimals valid",
			value: 1.5,
			schema: &Property{
				Type:       "number",
				MultipleOf: float64Ptr(0.5),
			},
			wantErr: false,
		},
		{
			name:  "multipleOf constraint with decimals invalid",
			value: 1.7,
			schema: &Property{
				Type:       "number",
				MultipleOf: float64Ptr(0.5),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateValue(tt.value, tt.schema, "test")

			if tt.wantErr && result.Valid {
				t.Errorf("expected validation error, but validation passed")
			}

			if !tt.wantErr && !result.Valid {
				t.Errorf("expected validation to pass, but got errors: %v", result.Errors)
			}
		})
	}
}

func TestSchemaValidator_ArrayConstraints(t *testing.T) {
	validator := NewValidator(false)

	tests := []struct {
		name    string
		value   any
		schema  *Property
		wantErr bool
	}{
		{
			name:  "minItems constraint valid",
			value: []any{"a", "b", "c"},
			schema: &Property{
				Type:     "array",
				MinItems: intPtr(2),
			},
			wantErr: false,
		},
		{
			name:  "minItems constraint invalid",
			value: []any{"a"},
			schema: &Property{
				Type:     "array",
				MinItems: intPtr(2),
			},
			wantErr: true,
		},
		{
			name:  "maxItems constraint valid",
			value: []any{"a", "b"},
			schema: &Property{
				Type:     "array",
				MaxItems: intPtr(3),
			},
			wantErr: false,
		},
		{
			name:  "maxItems constraint invalid",
			value: []any{"a", "b", "c", "d"},
			schema: &Property{
				Type:     "array",
				MaxItems: intPtr(3),
			},
			wantErr: true,
		},
		{
			name:  "uniqueItems constraint valid",
			value: []any{"a", "b", "c"},
			schema: &Property{
				Type:        "array",
				UniqueItems: boolPtr(true),
			},
			wantErr: false,
		},
		{
			name:  "uniqueItems constraint invalid",
			value: []any{"a", "b", "a"},
			schema: &Property{
				Type:        "array",
				UniqueItems: boolPtr(true),
			},
			wantErr: true,
		},
		{
			name:  "uniqueItems constraint with numbers valid",
			value: []any{1, 2, 3},
			schema: &Property{
				Type:        "array",
				UniqueItems: boolPtr(true),
			},
			wantErr: false,
		},
		{
			name:  "uniqueItems constraint with numbers invalid",
			value: []any{1, 2, 1},
			schema: &Property{
				Type:        "array",
				UniqueItems: boolPtr(true),
			},
			wantErr: true,
		},
		{
			name:  "uniqueItems false allows duplicates",
			value: []any{"a", "b", "a"},
			schema: &Property{
				Type:        "array",
				UniqueItems: boolPtr(false),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateValue(tt.value, tt.schema, "test")

			if tt.wantErr && result.Valid {
				t.Errorf("expected validation error, but validation passed")
			}

			if !tt.wantErr && !result.Valid {
				t.Errorf("expected validation to pass, but got errors: %v", result.Errors)
			}
		})
	}
}

func TestSchemaValidator_EnumConstraints(t *testing.T) {
	validator := NewValidator(false)

	tests := []struct {
		name    string
		value   any
		schema  *Property
		wantErr bool
	}{
		{
			name:  "string enum valid",
			value: "red",
			schema: &Property{
				Type: "string",
				Enum: []any{"red", "green", "blue"},
			},
			wantErr: false,
		},
		{
			name:  "string enum invalid",
			value: "yellow",
			schema: &Property{
				Type: "string",
				Enum: []any{"red", "green", "blue"},
			},
			wantErr: true,
		},
		{
			name:  "integer enum valid",
			value: 2,
			schema: &Property{
				Type: "integer",
				Enum: []any{1, 2, 3},
			},
			wantErr: false,
		},
		{
			name:  "integer enum invalid",
			value: 4,
			schema: &Property{
				Type: "integer",
				Enum: []any{1, 2, 3},
			},
			wantErr: true,
		},
		{
			name:  "mixed type enum valid",
			value: "mixed",
			schema: &Property{
				Enum: []any{"mixed", 42, true},
			},
			wantErr: false,
		},
		{
			name:  "mixed type enum valid number",
			value: 42,
			schema: &Property{
				Enum: []any{"mixed", 42, true},
			},
			wantErr: false,
		},
		{
			name:  "mixed type enum valid boolean",
			value: true,
			schema: &Property{
				Enum: []any{"mixed", 42, true},
			},
			wantErr: false,
		},
		{
			name:  "mixed type enum invalid",
			value: "invalid",
			schema: &Property{
				Enum: []any{"mixed", 42, true},
			},
			wantErr: true,
		},
		{
			name:  "null enum valid",
			value: nil,
			schema: &Property{
				Enum: []any{nil, "value"},
			},
			wantErr: false,
		},
		{
			name:  "null enum invalid",
			value: nil,
			schema: &Property{
				Enum: []any{"value1", "value2"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateValue(tt.value, tt.schema, "test")

			if tt.wantErr && result.Valid {
				t.Errorf("expected validation error, but validation passed")
			}

			if !tt.wantErr && !result.Valid {
				t.Errorf("expected validation to pass, but got errors: %v", result.Errors)
			}
		})
	}
}

func TestSchemaValidator_ConstConstraints(t *testing.T) {
	validator := NewValidator(false)

	tests := []struct {
		name    string
		value   any
		schema  *Property
		wantErr bool
	}{
		{
			name:  "const string valid",
			value: "constant",
			schema: &Property{
				Type:  "string",
				Const: "constant",
			},
			wantErr: false,
		},
		{
			name:  "const string invalid",
			value: "different",
			schema: &Property{
				Type:  "string",
				Const: "constant",
			},
			wantErr: true,
		},
		{
			name:  "const number valid",
			value: 42,
			schema: &Property{
				Type:  "integer",
				Const: 42,
			},
			wantErr: false,
		},
		{
			name:  "const number invalid",
			value: 43,
			schema: &Property{
				Type:  "integer",
				Const: 42,
			},
			wantErr: true,
		},
		{
			name:  "const boolean valid",
			value: true,
			schema: &Property{
				Type:  "boolean",
				Const: true,
			},
			wantErr: false,
		},
		{
			name:  "const boolean invalid",
			value: false,
			schema: &Property{
				Type:  "boolean",
				Const: true,
			},
			wantErr: true,
		},
		{
			name:  "const null valid",
			value: nil,
			schema: &Property{
				Const: nil,
			},
			wantErr: false,
		},
		{
			name:  "const null invalid",
			value: "not null",
			schema: &Property{
				Type:  "null",
				Const: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateValue(tt.value, tt.schema, "test")

			if tt.wantErr && result.Valid {
				t.Errorf("expected validation error, but validation passed")
			}

			if !tt.wantErr && !result.Valid {
				t.Errorf("expected validation to pass, but got errors: %v", result.Errors)
			}
		})
	}
}

func TestSchemaValidator_ComplexConstraintCombinations(t *testing.T) {
	validator := NewValidator(false)

	tests := []struct {
		name    string
		value   any
		schema  *Property
		wantErr bool
	}{
		{
			name:  "string with multiple constraints valid",
			value: "hello@example.com",
			schema: &Property{
				Type:      "string",
				MinLength: intPtr(5),
				MaxLength: intPtr(50),
				Pattern:   `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			},
			wantErr: false,
		},
		{
			name:  "string with multiple constraints invalid length",
			value: "a@b.c",
			schema: &Property{
				Type:      "string",
				MinLength: intPtr(10),
				MaxLength: intPtr(50),
				Pattern:   `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			},
			wantErr: true,
		},
		{
			name:  "number with multiple constraints valid",
			value: 25,
			schema: &Property{
				Type:       "integer",
				Minimum:    float64Ptr(10),
				Maximum:    float64Ptr(100),
				MultipleOf: float64Ptr(5),
			},
			wantErr: false,
		},
		{
			name:  "number with multiple constraints invalid multiple",
			value: 23,
			schema: &Property{
				Type:       "integer",
				Minimum:    float64Ptr(10),
				Maximum:    float64Ptr(100),
				MultipleOf: float64Ptr(5),
			},
			wantErr: true,
		},
		{
			name:  "array with multiple constraints valid",
			value: []any{"apple", "banana", "cherry"},
			schema: &Property{
				Type:        "array",
				MinItems:    intPtr(2),
				MaxItems:    intPtr(5),
				UniqueItems: boolPtr(true),
				Items: &Property{
					Type:      "string",
					MinLength: intPtr(3),
				},
			},
			wantErr: false,
		},
		{
			name:  "array with multiple constraints invalid item",
			value: []any{"apple", "banana", "x"},
			schema: &Property{
				Type:        "array",
				MinItems:    intPtr(2),
				MaxItems:    intPtr(5),
				UniqueItems: boolPtr(true),
				Items: &Property{
					Type:      "string",
					MinLength: intPtr(3),
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateValue(tt.value, tt.schema, "test")

			if tt.wantErr && result.Valid {
				t.Errorf("expected validation error, but validation passed")
			}

			if !tt.wantErr && !result.Valid {
				t.Errorf("expected validation to pass, but got errors: %v", result.Errors)
			}
		})
	}
}
func TestSchemaValidator_RequiredFieldValidation(t *testing.T) {
	validator := NewValidator(false)

	tests := []struct {
		name    string
		value   any
		schema  *Property
		wantErr bool
		errMsg  string
	}{
		{
			name: "all required fields present",
			value: map[string]any{
				"name":  "John",
				"email": "john@example.com",
				"age":   30,
			},
			schema: &Property{
				Type: "object",
				Properties: map[string]*Property{
					"name":  {Type: "string"},
					"email": {Type: "string"},
					"age":   {Type: "integer"},
				},
				Required: []string{"name", "email"},
			},
			wantErr: false,
		},
		{
			name: "missing one required field",
			value: map[string]any{
				"name": "John",
				"age":  30,
			},
			schema: &Property{
				Type: "object",
				Properties: map[string]*Property{
					"name":  {Type: "string"},
					"email": {Type: "string"},
					"age":   {Type: "integer"},
				},
				Required: []string{"name", "email"},
			},
			wantErr: true,
			errMsg:  "required property is missing",
		},
		{
			name: "missing multiple required fields",
			value: map[string]any{
				"age": 30,
			},
			schema: &Property{
				Type: "object",
				Properties: map[string]*Property{
					"name":  {Type: "string"},
					"email": {Type: "string"},
					"age":   {Type: "integer"},
				},
				Required: []string{"name", "email"},
			},
			wantErr: true,
			errMsg:  "required property is missing",
		},
		{
			name: "no required fields",
			value: map[string]any{
				"age": 30,
			},
			schema: &Property{
				Type: "object",
				Properties: map[string]*Property{
					"name":  {Type: "string"},
					"email": {Type: "string"},
					"age":   {Type: "integer"},
				},
				Required: []string{},
			},
			wantErr: false,
		},
		{
			name:  "empty object with required fields",
			value: map[string]any{},
			schema: &Property{
				Type: "object",
				Properties: map[string]*Property{
					"name": {Type: "string"},
				},
				Required: []string{"name"},
			},
			wantErr: true,
			errMsg:  "required property is missing",
		},
		{
			name: "nested object with required fields",
			value: map[string]any{
				"user": map[string]any{
					"name": "John",
				},
			},
			schema: &Property{
				Type: "object",
				Properties: map[string]*Property{
					"user": {
						Type: "object",
						Properties: map[string]*Property{
							"name":  {Type: "string"},
							"email": {Type: "string"},
						},
						Required: []string{"name", "email"},
					},
				},
				Required: []string{"user"},
			},
			wantErr: true,
			errMsg:  "required property is missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateValue(tt.value, tt.schema, "test")

			if tt.wantErr && result.Valid {
				t.Errorf("expected validation error, but validation passed")
			}

			if !tt.wantErr && !result.Valid {
				t.Errorf("expected validation to pass, but got errors: %v", result.Errors)
			}

			if tt.wantErr && len(result.Errors) > 0 {
				found := false
				for _, err := range result.Errors {
					if strings.Contains(err.Message, tt.errMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error message containing '%s', but got: %v", tt.errMsg, result.Errors)
				}
			}
		})
	}
}

func TestSchemaValidator_AdditionalPropertiesValidation(t *testing.T) {
	tests := []struct {
		name       string
		value      any
		schema     *Property
		strictMode bool
		wantErr    bool
		errMsg     string
	}{
		{
			name: "additional properties allowed in permissive mode",
			value: map[string]any{
				"name":  "John",
				"extra": "allowed",
			},
			schema: &Property{
				Type: "object",
				Properties: map[string]*Property{
					"name": {Type: "string"},
				},
			},
			strictMode: false,
			wantErr:    false,
		},
		{
			name: "additional properties rejected in strict mode",
			value: map[string]any{
				"name":  "John",
				"extra": "not allowed",
			},
			schema: &Property{
				Type: "object",
				Properties: map[string]*Property{
					"name": {Type: "string"},
				},
			},
			strictMode: true,
			wantErr:    true,
			errMsg:     "additional property not allowed in strict mode",
		},
		{
			name: "multiple additional properties in strict mode",
			value: map[string]any{
				"name":   "John",
				"extra1": "not allowed",
				"extra2": "also not allowed",
			},
			schema: &Property{
				Type: "object",
				Properties: map[string]*Property{
					"name": {Type: "string"},
				},
			},
			strictMode: true,
			wantErr:    true,
			errMsg:     "additional property not allowed in strict mode",
		},
		{
			name: "no additional properties",
			value: map[string]any{
				"name": "John",
				"age":  30,
			},
			schema: &Property{
				Type: "object",
				Properties: map[string]*Property{
					"name": {Type: "string"},
					"age":  {Type: "integer"},
				},
			},
			strictMode: true,
			wantErr:    false,
		},
		{
			name: "empty schema allows all properties in permissive mode",
			value: map[string]any{
				"anything": "goes",
				"here":     123,
			},
			schema: &Property{
				Type: "object",
			},
			strictMode: false,
			wantErr:    false,
		},
		{
			name: "empty schema rejects all properties in strict mode",
			value: map[string]any{
				"anything": "rejected",
			},
			schema: &Property{
				Type: "object",
			},
			strictMode: true,
			wantErr:    true,
			errMsg:     "additional property not allowed in strict mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewValidator(tt.strictMode)
			result := validator.ValidateValue(tt.value, tt.schema, "test")

			if tt.wantErr && result.Valid {
				t.Errorf("expected validation error, but validation passed")
			}

			if !tt.wantErr && !result.Valid {
				t.Errorf("expected validation to pass, but got errors: %v", result.Errors)
			}

			if tt.wantErr && len(result.Errors) > 0 {
				found := false
				for _, err := range result.Errors {
					if strings.Contains(err.Message, tt.errMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error message containing '%s', but got: %v", tt.errMsg, result.Errors)
				}
			}
		})
	}
}

func TestSchemaValidator_RequiredFieldsWithAdditionalProperties(t *testing.T) {
	tests := []struct {
		name       string
		value      any
		schema     *Property
		strictMode bool
		wantErr    bool
		errCount   int
	}{
		{
			name: "missing required field with additional property in permissive mode",
			value: map[string]any{
				"extra": "allowed",
			},
			schema: &Property{
				Type: "object",
				Properties: map[string]*Property{
					"name": {Type: "string"},
				},
				Required: []string{"name"},
			},
			strictMode: false,
			wantErr:    true,
			errCount:   1, // Only missing required field error
		},
		{
			name: "missing required field with additional property in strict mode",
			value: map[string]any{
				"extra": "not allowed",
			},
			schema: &Property{
				Type: "object",
				Properties: map[string]*Property{
					"name": {Type: "string"},
				},
				Required: []string{"name"},
			},
			strictMode: true,
			wantErr:    true,
			errCount:   2, // Missing required field + additional property errors
		},
		{
			name: "valid required field with additional property in permissive mode",
			value: map[string]any{
				"name":  "John",
				"extra": "allowed",
			},
			schema: &Property{
				Type: "object",
				Properties: map[string]*Property{
					"name": {Type: "string"},
				},
				Required: []string{"name"},
			},
			strictMode: false,
			wantErr:    false,
			errCount:   0,
		},
		{
			name: "valid required field with additional property in strict mode",
			value: map[string]any{
				"name":  "John",
				"extra": "not allowed",
			},
			schema: &Property{
				Type: "object",
				Properties: map[string]*Property{
					"name": {Type: "string"},
				},
				Required: []string{"name"},
			},
			strictMode: true,
			wantErr:    true,
			errCount:   1, // Only additional property error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewValidator(tt.strictMode)
			result := validator.ValidateValue(tt.value, tt.schema, "test")

			if tt.wantErr && result.Valid {
				t.Errorf("expected validation error, but validation passed")
			}

			if !tt.wantErr && !result.Valid {
				t.Errorf("expected validation to pass, but got errors: %v", result.Errors)
			}

			if len(result.Errors) != tt.errCount {
				t.Errorf("expected %d errors, but got %d: %v", tt.errCount, len(result.Errors), result.Errors)
			}
		})
	}
}

func TestSchemaValidator_DescriptiveErrorMessages(t *testing.T) {
	validator := NewValidator(true)

	tests := []struct {
		name          string
		value         any
		schema        *Property
		expectedField string
		expectedMsg   string
	}{
		{
			name: "missing required field with field path",
			value: map[string]any{
				"user": map[string]any{
					"name": "John",
				},
			},
			schema: &Property{
				Type: "object",
				Properties: map[string]*Property{
					"user": {
						Type: "object",
						Properties: map[string]*Property{
							"name":  {Type: "string"},
							"email": {Type: "string"},
						},
						Required: []string{"email"},
					},
				},
			},
			expectedField: "test.user.email",
			expectedMsg:   "required property is missing",
		},
		{
			name: "additional property with field path",
			value: map[string]any{
				"user": map[string]any{
					"name":  "John",
					"extra": "not allowed",
				},
			},
			schema: &Property{
				Type: "object",
				Properties: map[string]*Property{
					"user": {
						Type: "object",
						Properties: map[string]*Property{
							"name": {Type: "string"},
						},
					},
				},
			},
			expectedField: "test.user.extra",
			expectedMsg:   "additional property not allowed in strict mode",
		},
		{
			name: "root level missing required field",
			value: map[string]any{
				"age": 30,
			},
			schema: &Property{
				Type: "object",
				Properties: map[string]*Property{
					"name": {Type: "string"},
					"age":  {Type: "integer"},
				},
				Required: []string{"name"},
			},
			expectedField: "test.name",
			expectedMsg:   "required property is missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateValue(tt.value, tt.schema, "test")

			if result.Valid {
				t.Errorf("expected validation error, but validation passed")
				return
			}

			found := false
			for _, err := range result.Errors {
				if err.Field == tt.expectedField && strings.Contains(err.Message, tt.expectedMsg) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("expected error with field '%s' and message containing '%s', but got: %v",
					tt.expectedField, tt.expectedMsg, result.Errors)
			}
		})
	}
}
