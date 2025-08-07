package generator

import (
	"strings"
	"testing"
)

func TestNestedObjectHandling(t *testing.T) {
	tests := []struct {
		name     string
		schema   *MessageSchema
		expected []string // Expected struct names to be generated
	}{
		{
			name: "simple nested object",
			schema: &MessageSchema{
				Name: "User",
				Type: "object",
				Properties: map[string]*Property{
					"name": {
						Type: "string",
					},
					"address": {
						Type: "object",
						Properties: map[string]*Property{
							"street": {
								Type: "string",
							},
							"city": {
								Type: "string",
							},
						},
					},
				},
			},
			expected: []string{"User", "UserAddress"},
		},
		{
			name: "deeply nested objects",
			schema: &MessageSchema{
				Name: "Company",
				Type: "object",
				Properties: map[string]*Property{
					"name": {
						Type: "string",
					},
					"headquarters": {
						Type: "object",
						Properties: map[string]*Property{
							"address": {
								Type: "object",
								Properties: map[string]*Property{
									"street": {
										Type: "string",
									},
									"coordinates": {
										Type: "object",
										Properties: map[string]*Property{
											"lat": {
												Type: "number",
											},
											"lng": {
												Type: "number",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: []string{"Company", "CompanyHeadquarters", "CompanyHeadquartersAddress", "CompanyHeadquartersAddressCoordinates"},
		},
		{
			name: "nested object with array",
			schema: &MessageSchema{
				Name: "Order",
				Type: "object",
				Properties: map[string]*Property{
					"id": {
						Type: "string",
					},
					"items": {
						Type: "array",
						Items: &Property{
							Type: "object",
							Properties: map[string]*Property{
								"name": {
									Type: "string",
								},
								"price": {
									Type: "number",
								},
							},
						},
					},
				},
			},
			expected: []string{"Order", "OrderItem"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				PackageName:     "test",
				IncludeComments: true,
				UsePointers:     false,
			}

			cg := NewCodeGenerator(config)

			// Generate the struct code
			result, err := cg.GenerateTypes(map[string]*MessageSchema{
				tt.schema.Name: tt.schema,
			}, config)

			if err != nil {
				t.Fatalf("GenerateTypes failed: %v", err)
			}

			// Check that all expected struct names are present in the generated code
			for _, expectedStruct := range tt.expected {
				found := false
				for _, fileContent := range result.Files {
					if strings.Contains(fileContent, "type "+expectedStruct+" struct") {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected struct %s not found in generated code", expectedStruct)
					// Print generated code for debugging
					for filename, content := range result.Files {
						t.Logf("Generated file %s:\n%s", filename, content)
					}
				}
			}
		})
	}
}

func TestNestedObjectNaming(t *testing.T) {
	tests := []struct {
		name         string
		parentName   string
		fieldName    string
		expectedName string
	}{
		{
			name:         "simple nested",
			parentName:   "User",
			fieldName:    "address",
			expectedName: "UserAddress",
		},
		{
			name:         "camelCase field",
			parentName:   "User",
			fieldName:    "homeAddress",
			expectedName: "UserHomeAddress",
		},
		{
			name:         "snake_case field",
			parentName:   "User",
			fieldName:    "home_address",
			expectedName: "UserHomeAddress",
		},
		{
			name:         "kebab-case field",
			parentName:   "User",
			fieldName:    "home-address",
			expectedName: "UserHomeAddress",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateNestedStructName(tt.parentName, tt.fieldName)
			if result != tt.expectedName {
				t.Errorf("generateNestedStructName(%q, %q) = %q, expected %q",
					tt.parentName, tt.fieldName, result, tt.expectedName)
			}
		})
	}
}

func TestNestedObjectConflictResolution(t *testing.T) {
	// Test that nested structs with the same name get unique names
	schema := &MessageSchema{
		Name: "User",
		Type: "object",
		Properties: map[string]*Property{
			"homeAddress": {
				Type: "object",
				Properties: map[string]*Property{
					"street": {Type: "string"},
				},
			},
			"workAddress": {
				Type: "object",
				Properties: map[string]*Property{
					"street": {Type: "string"},
				},
			},
		},
	}

	config := &Config{
		PackageName:     "test",
		IncludeComments: true,
		UsePointers:     false,
	}

	cg := NewCodeGenerator(config)

	result, err := cg.GenerateTypes(map[string]*MessageSchema{
		schema.Name: schema,
	}, config)

	if err != nil {
		t.Fatalf("GenerateTypes failed: %v", err)
	}

	// Check that both address structs are generated with unique names
	expectedStructs := []string{"User", "UserHomeAddress", "UserWorkAddress"}

	for _, expectedStruct := range expectedStructs {
		found := false
		for _, fileContent := range result.Files {
			if strings.Contains(fileContent, "type "+expectedStruct+" struct") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected struct %s not found in generated code", expectedStruct)
			// Print generated code for debugging
			for filename, content := range result.Files {
				t.Logf("Generated file %s:\n%s", filename, content)
			}
		}
	}
}
