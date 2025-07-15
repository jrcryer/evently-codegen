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
		{"array", "", "[]interface{}"},
		{"object", "", "interface{}"},
		{"null", "", "interface{}"},
		{"unknown", "", "interface{}"},
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
		{"XMLHttpRequest", "Xmlhttprequest"}, // This is expected behavior for this simple implementation
		{"camelCase", "Camelcase"},
		{"PascalCase", "Pascalcase"},
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
			expected: "interface{}",
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
			expected: "interface{}",
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
