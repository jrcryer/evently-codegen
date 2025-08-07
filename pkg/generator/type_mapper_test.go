package generator

import (
	"testing"
)

func TestNewTypeMapper(t *testing.T) {
	config := &Config{UsePointers: true}
	mapper := NewTypeMapper(config)

	if mapper == nil {
		t.Fatal("NewTypeMapper returned nil")
	}

	if mapper.config != config {
		t.Error("TypeMapper config not set correctly")
	}
}

func TestMapType_BasicTypes(t *testing.T) {
	mapper := NewTypeMapper(&Config{})

	tests := []struct {
		schemaType string
		format     string
		expected   string
	}{
		{"string", "", "string"},
		{"number", "", "float64"},
		{"integer", "", "int64"},
		{"boolean", "", "bool"},
		{"array", "", "[]any"},
		{"object", "", "any"},
		{"null", "", "any"},
		{"unknown", "", "any"},
	}

	for _, test := range tests {
		result := mapper.MapType(test.schemaType, test.format)
		if result != test.expected {
			t.Errorf("MapType(%q, %q) = %q, expected %q",
				test.schemaType, test.format, result, test.expected)
		}
	}
}

func TestMapType_FormatSpecific(t *testing.T) {
	mapper := NewTypeMapper(&Config{})

	tests := []struct {
		schemaType string
		format     string
		expected   string
	}{
		// String formats
		{"string", "date-time", "time.Time"},
		{"string", "date", "time.Time"},
		{"string", "time", "time.Time"},
		{"string", "email", "string"},
		{"string", "hostname", "string"},
		{"string", "ipv4", "string"},
		{"string", "ipv6", "string"},
		{"string", "uri", "string"},
		{"string", "uri-reference", "string"},
		{"string", "uuid", "string"},
		{"string", "byte", "[]byte"},
		{"string", "binary", "[]byte"},
		{"string", "password", "string"},

		// Integer formats
		{"integer", "int32", "int32"},
		{"integer", "int64", "int64"},
		{"integer", "unknown", "int64"},

		// Number formats
		{"number", "float", "float32"},
		{"number", "double", "float64"},
		{"number", "unknown", "float64"},
	}

	for _, test := range tests {
		result := mapper.MapType(test.schemaType, test.format)
		if result != test.expected {
			t.Errorf("MapType(%q, %q) = %q, expected %q",
				test.schemaType, test.format, result, test.expected)
		}
	}
}

func TestMapProperty_BasicTypes(t *testing.T) {
	mapper := NewTypeMapper(&Config{UsePointers: false})

	tests := []struct {
		name     string
		prop     *Property
		expected *GoField
	}{
		{
			name: "string property",
			prop: &Property{
				Type:        "string",
				Description: "A string field",
			},
			expected: &GoField{
				Type:     "string",
				Comment:  "A string field",
				Optional: true,
			},
		},
		{
			name: "integer property",
			prop: &Property{
				Type: "integer",
			},
			expected: &GoField{
				Type:     "int64",
				Optional: true,
			},
		},
		{
			name: "boolean property",
			prop: &Property{
				Type: "boolean",
			},
			expected: &GoField{
				Type:     "bool",
				Optional: true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := mapper.MapProperty(test.prop)

			if result == nil {
				t.Fatal("MapProperty returned nil")
			}

			if result.Type != test.expected.Type {
				t.Errorf("Type = %q, expected %q", result.Type, test.expected.Type)
			}

			if result.Comment != test.expected.Comment {
				t.Errorf("Comment = %q, expected %q", result.Comment, test.expected.Comment)
			}

			if result.Optional != test.expected.Optional {
				t.Errorf("Optional = %v, expected %v", result.Optional, test.expected.Optional)
			}
		})
	}
}

func TestMapProperty_ArrayTypes(t *testing.T) {
	mapper := NewTypeMapper(&Config{})

	tests := []struct {
		name     string
		prop     *Property
		expected string
	}{
		{
			name: "array of strings",
			prop: &Property{
				Type: "array",
				Items: &Property{
					Type: "string",
				},
			},
			expected: "[]string",
		},
		{
			name: "array of integers",
			prop: &Property{
				Type: "array",
				Items: &Property{
					Type: "integer",
				},
			},
			expected: "[]int64",
		},
		{
			name: "nested array",
			prop: &Property{
				Type: "array",
				Items: &Property{
					Type: "array",
					Items: &Property{
						Type: "string",
					},
				},
			},
			expected: "[][]string",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := mapper.MapProperty(test.prop)

			if result == nil {
				t.Fatal("MapProperty returned nil")
			}

			if result.Type != test.expected {
				t.Errorf("Type = %q, expected %q", result.Type, test.expected)
			}
		})
	}
}

func TestMapProperty_WithPointers(t *testing.T) {
	mapper := NewTypeMapper(&Config{UsePointers: true})

	prop := &Property{
		Type: "string",
	}

	result := mapper.MapProperty(prop)

	if result == nil {
		t.Fatal("MapProperty returned nil")
	}

	// Should be a pointer since it's optional and UsePointers is true
	if result.Type != "*string" {
		t.Errorf("Type = %q, expected %q", result.Type, "*string")
	}
}

func TestMapProperty_NilProperty(t *testing.T) {
	mapper := NewTypeMapper(&Config{})

	result := mapper.MapProperty(nil)

	if result != nil {
		t.Error("MapProperty with nil property should return nil")
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"hello", "Hello"},
		{"hello_world", "HelloWorld"},
		{"hello-world", "HelloWorld"},
		{"hello world", "HelloWorld"},
		{"hello.world", "HelloWorld"},
		{"user_id", "UserId"},
		{"API_KEY", "ApiKey"},
		{"XMLHttpRequest", "XMLHttpRequest"}, // Preserve acronyms and existing PascalCase
		{"camelCase", "CamelCase"},
		{"PascalCase", "PascalCase"},
	}

	for _, test := range tests {
		result := ToPascalCase(test.input)
		if result != test.expected {
			t.Errorf("ToPascalCase(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestToJSONTag(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"fieldName", `json:"fieldName"`},
		{"user_id", `json:"user_id"`},
		{"", `json:""`},
	}

	for _, test := range tests {
		result := ToJSONTag(test.input)
		if result != test.expected {
			t.Errorf("ToJSONTag(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestMapPropertyWithContext(t *testing.T) {
	mapper := NewTypeMapper(&Config{UsePointers: true})

	prop := &Property{
		Type:        "string",
		Description: "User name",
	}

	// Test required field
	result := mapper.MapPropertyWithContext(prop, "user_name", true)

	if result == nil {
		t.Fatal("MapPropertyWithContext returned nil")
	}

	if result.Name != "UserName" {
		t.Errorf("Name = %q, expected %q", result.Name, "UserName")
	}

	if result.JSONTag != `json:"user_name"` {
		t.Errorf("JSONTag = %q, expected %q", result.JSONTag, `json:"user_name"`)
	}

	if result.Optional {
		t.Error("Required field should not be optional")
	}

	if result.Type != "string" {
		t.Errorf("Required field should not be pointer, got %q", result.Type)
	}

	// Test optional field
	result = mapper.MapPropertyWithContext(prop, "user_name", false)

	if !result.Optional {
		t.Error("Optional field should be marked as optional")
	}

	if result.Type != "*string" {
		t.Errorf("Optional field should be pointer when UsePointers=true, got %q", result.Type)
	}
}

func TestMapPropertyWithContext_NilProperty(t *testing.T) {
	mapper := NewTypeMapper(&Config{})

	result := mapper.MapPropertyWithContext(nil, "field", true)

	if result != nil {
		t.Error("MapPropertyWithContext with nil property should return nil")
	}
}

func TestGetPropertyType(t *testing.T) {
	mapper := NewTypeMapper(&Config{})

	tests := []struct {
		name     string
		prop     *Property
		expected string
	}{
		{
			name:     "nil property",
			prop:     nil,
			expected: "any",
		},
		{
			name: "string property",
			prop: &Property{
				Type: "string",
			},
			expected: "string",
		},
		{
			name: "array property",
			prop: &Property{
				Type: "array",
				Items: &Property{
					Type: "integer",
				},
			},
			expected: "[]int64",
		},
		{
			name: "object property",
			prop: &Property{
				Type: "object",
			},
			expected: "any",
		},
		{
			name: "enum property",
			prop: &Property{
				Type: "string",
				Enum: []interface{}{"red", "green", "blue"},
			},
			expected: "string",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := mapper.getPropertyType(test.prop)
			if result != test.expected {
				t.Errorf("getPropertyType() = %q, expected %q", result, test.expected)
			}
		})
	}
}

// Benchmark tests
func BenchmarkMapType(b *testing.B) {
	mapper := NewTypeMapper(&Config{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mapper.MapType("string", "date-time")
	}
}

func BenchmarkToPascalCase(b *testing.B) {
	input := "hello_world_test_field"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ToPascalCase(input)
	}
}

func TestIsEnumProperty(t *testing.T) {
	mapper := NewTypeMapper(&Config{})

	tests := []struct {
		name     string
		prop     *Property
		expected bool
	}{
		{
			name:     "nil property",
			prop:     nil,
			expected: false,
		},
		{
			name: "property without enum",
			prop: &Property{
				Type: "string",
			},
			expected: false,
		},
		{
			name: "property with empty enum",
			prop: &Property{
				Type: "string",
				Enum: []interface{}{},
			},
			expected: false,
		},
		{
			name: "property with string enum",
			prop: &Property{
				Type: "string",
				Enum: []interface{}{"red", "green", "blue"},
			},
			expected: true,
		},
		{
			name: "property with numeric enum",
			prop: &Property{
				Type: "integer",
				Enum: []interface{}{1, 2, 3},
			},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := mapper.isEnumProperty(test.prop)
			if result != test.expected {
				t.Errorf("isEnumProperty() = %v, expected %v", result, test.expected)
			}
		})
	}
}

func TestGenerateEnumTypeName(t *testing.T) {
	mapper := NewTypeMapper(&Config{})

	tests := []struct {
		fieldName string
		expected  string
	}{
		{"status", "StatusEnum"},
		{"user_role", "UserRoleEnum"},
		{"color-code", "ColorCodeEnum"},
		{"API_KEY", "ApiKeyEnum"},
		{"", "Enum"},
	}

	for _, test := range tests {
		result := mapper.generateEnumTypeName(test.fieldName)
		if result != test.expected {
			t.Errorf("generateEnumTypeName(%q) = %q, expected %q", test.fieldName, result, test.expected)
		}
	}
}

func TestGetEnumBaseType(t *testing.T) {
	mapper := NewTypeMapper(&Config{})

	tests := []struct {
		name       string
		enumValues []interface{}
		expected   string
	}{
		{
			name:       "empty enum",
			enumValues: []interface{}{},
			expected:   "string",
		},
		{
			name:       "string enum",
			enumValues: []interface{}{"red", "green", "blue"},
			expected:   "string",
		},
		{
			name:       "integer enum",
			enumValues: []interface{}{1, 2, 3},
			expected:   "int",
		},
		{
			name:       "float enum (stored as float64)",
			enumValues: []interface{}{1.5, 2.5, 3.5},
			expected:   "float64",
		},
		{
			name:       "integer stored as float64",
			enumValues: []interface{}{1.0, 2.0, 3.0},
			expected:   "int",
		},
		{
			name:       "boolean enum",
			enumValues: []interface{}{true, false},
			expected:   "bool",
		},
		{
			name:       "mixed types (first non-nil wins)",
			enumValues: []interface{}{nil, "red", 1},
			expected:   "string",
		},
		{
			name:       "all nil values",
			enumValues: []interface{}{nil, nil, nil},
			expected:   "string",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := mapper.getEnumBaseType(test.enumValues)
			if result != test.expected {
				t.Errorf("getEnumBaseType() = %q, expected %q", result, test.expected)
			}
		})
	}
}

func TestMapPropertyWithContext_Enums(t *testing.T) {
	mapper := NewTypeMapper(&Config{UsePointers: false})

	tests := []struct {
		name      string
		prop      *Property
		fieldName string
		required  bool
		expected  *GoField
	}{
		{
			name: "string enum property",
			prop: &Property{
				Type:        "string",
				Enum:        []interface{}{"red", "green", "blue"},
				Description: "Color enum",
			},
			fieldName: "color",
			required:  true,
			expected: &GoField{
				Name:         "Color",
				Type:         "ColorEnum",
				JSONTag:      `json:"color"`,
				Comment:      "Color enum",
				Optional:     false,
				IsEnum:       true,
				EnumValues:   []interface{}{"red", "green", "blue"},
				EnumBaseType: "string",
			},
		},
		{
			name: "integer enum property",
			prop: &Property{
				Type: "integer",
				Enum: []interface{}{1, 2, 3},
			},
			fieldName: "priority",
			required:  false,
			expected: &GoField{
				Name:         "Priority",
				Type:         "PriorityEnum",
				JSONTag:      `json:"priority"`,
				Optional:     true,
				IsEnum:       true,
				EnumValues:   []interface{}{1, 2, 3},
				EnumBaseType: "int",
			},
		},
		{
			name: "array of enum items",
			prop: &Property{
				Type: "array",
				Items: &Property{
					Type: "string",
					Enum: []interface{}{"tag1", "tag2", "tag3"},
				},
			},
			fieldName: "tags",
			required:  true,
			expected: &GoField{
				Name:     "Tags",
				Type:     "[]TagsItemEnum",
				JSONTag:  `json:"tags"`,
				Optional: false,
				IsEnum:   false,
				ArrayItemEnum: &EnumInfo{
					TypeName: "TagsItemEnum",
					Values:   []interface{}{"tag1", "tag2", "tag3"},
					BaseType: "string",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := mapper.MapPropertyWithContext(test.prop, test.fieldName, test.required)

			if result == nil {
				t.Fatal("MapPropertyWithContext returned nil")
			}

			if result.Name != test.expected.Name {
				t.Errorf("Name = %q, expected %q", result.Name, test.expected.Name)
			}

			if result.Type != test.expected.Type {
				t.Errorf("Type = %q, expected %q", result.Type, test.expected.Type)
			}

			if result.JSONTag != test.expected.JSONTag {
				t.Errorf("JSONTag = %q, expected %q", result.JSONTag, test.expected.JSONTag)
			}

			if result.Comment != test.expected.Comment {
				t.Errorf("Comment = %q, expected %q", result.Comment, test.expected.Comment)
			}

			if result.Optional != test.expected.Optional {
				t.Errorf("Optional = %v, expected %v", result.Optional, test.expected.Optional)
			}

			if result.IsEnum != test.expected.IsEnum {
				t.Errorf("IsEnum = %v, expected %v", result.IsEnum, test.expected.IsEnum)
			}

			if test.expected.IsEnum {
				if result.EnumBaseType != test.expected.EnumBaseType {
					t.Errorf("EnumBaseType = %q, expected %q", result.EnumBaseType, test.expected.EnumBaseType)
				}

				if len(result.EnumValues) != len(test.expected.EnumValues) {
					t.Errorf("EnumValues length = %d, expected %d", len(result.EnumValues), len(test.expected.EnumValues))
				}
			}

			if test.expected.ArrayItemEnum != nil {
				if result.ArrayItemEnum == nil {
					t.Fatal("ArrayItemEnum is nil, expected non-nil")
				}

				if result.ArrayItemEnum.TypeName != test.expected.ArrayItemEnum.TypeName {
					t.Errorf("ArrayItemEnum.TypeName = %q, expected %q", result.ArrayItemEnum.TypeName, test.expected.ArrayItemEnum.TypeName)
				}

				if result.ArrayItemEnum.BaseType != test.expected.ArrayItemEnum.BaseType {
					t.Errorf("ArrayItemEnum.BaseType = %q, expected %q", result.ArrayItemEnum.BaseType, test.expected.ArrayItemEnum.BaseType)
				}
			}
		})
	}
}

func BenchmarkMapProperty(b *testing.B) {
	mapper := NewTypeMapper(&Config{UsePointers: true})
	prop := &Property{
		Type:        "string",
		Format:      "date-time",
		Description: "A timestamp field",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mapper.MapProperty(prop)
	}
}
