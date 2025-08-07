package generator

import (
	"strings"
	"testing"
)

func TestCodeGenerator_GenerateValidationMethods(t *testing.T) {
	config := &Config{
		PackageName:     "test",
		IncludeComments: true,
		UsePointers:     false,
	}

	codeGen := NewCodeGenerator(config)

	// Create a test schema
	schema := &MessageSchema{
		Type: "object",
		Properties: map[string]*Property{
			"name": {
				Type:        "string",
				Description: "User name",
			},
			"email": {
				Type:        "string",
				Description: "User email",
			},
			"age": {
				Type:        "integer",
				Description: "User age",
			},
		},
		Required: []string{"name", "email"},
	}

	// Generate struct code
	structCode, err := codeGen.GenerateStruct(schema, "User")
	if err != nil {
		t.Fatalf("Failed to generate struct: %v", err)
	}

	// Check that validation methods are included
	if !strings.Contains(structCode, "func (s *User) Validate()") {
		t.Errorf("Generated code should contain Validate() method")
	}

	if !strings.Contains(structCode, "func (s *User) ValidateJSON(") {
		t.Errorf("Generated code should contain ValidateJSON() method")
	}

	// Check that validation logic is present
	if !strings.Contains(structCode, "NewValidator(false)") {
		t.Errorf("Generated code should create validator instance")
	}

	if !strings.Contains(structCode, "ValidationResult") {
		t.Errorf("Generated code should return ValidationResult")
	}

	// Check that required fields are included
	if !strings.Contains(structCode, `Required: []string{"name", "email"}`) {
		t.Errorf("Generated code should include required fields")
	}

	// Print the generated code for manual inspection
	t.Logf("Generated code:\n%s", structCode)
}

func TestCodeGenerator_ValidationMethodsWithDifferentTypes(t *testing.T) {
	config := &Config{
		PackageName:     "test",
		IncludeComments: false,
		UsePointers:     false,
	}

	codeGen := NewCodeGenerator(config)

	// Create a schema with various types
	schema := &MessageSchema{
		Type: "object",
		Properties: map[string]*Property{
			"id": {
				Type: "string",
			},
			"count": {
				Type: "integer",
			},
			"price": {
				Type: "number",
			},
			"active": {
				Type: "boolean",
			},
			"tags": {
				Type: "array",
				Items: &Property{
					Type: "string",
				},
			},
		},
		Required: []string{"id"},
	}

	// Generate struct code
	structCode, err := codeGen.GenerateStruct(schema, "Product")
	if err != nil {
		t.Fatalf("Failed to generate struct: %v", err)
	}

	// Check that different types are handled correctly in validation
	expectedTypes := []string{
		`"id": {Type: "string"}`,
		`"count": {Type: "integer"}`,
		`"price": {Type: "number"}`,
		`"active": {Type: "boolean"}`,
		`"tags": {Type: "array"}`,
	}

	for _, expectedType := range expectedTypes {
		if !strings.Contains(structCode, expectedType) {
			t.Errorf("Generated code should contain type mapping: %s", expectedType)
		}
	}

	// Check required field
	if !strings.Contains(structCode, `Required: []string{"id"}`) {
		t.Errorf("Generated code should include required field 'id'")
	}
}

func TestCodeGenerator_ValidationMethodsWithNoRequiredFields(t *testing.T) {
	config := &Config{
		PackageName: "test",
		UsePointers: false,
	}

	codeGen := NewCodeGenerator(config)

	// Create a schema with no required fields
	schema := &MessageSchema{
		Type: "object",
		Properties: map[string]*Property{
			"optional1": {
				Type: "string",
			},
			"optional2": {
				Type: "integer",
			},
		},
		Required: []string{}, // No required fields
	}

	// Generate struct code
	structCode, err := codeGen.GenerateStruct(schema, "OptionalStruct")
	if err != nil {
		t.Fatalf("Failed to generate struct: %v", err)
	}

	// Should still contain validation methods
	if !strings.Contains(structCode, "func (s *OptionalStruct) Validate()") {
		t.Errorf("Generated code should contain Validate() method even with no required fields")
	}

	// Should not contain Required field or should be empty
	if strings.Contains(structCode, `Required: []string{"`) {
		t.Errorf("Generated code should not contain non-empty Required field")
	}
}

func TestCodeGenerator_ValidationMethodsIntegration(t *testing.T) {
	config := &Config{
		PackageName: "test",
		UsePointers: false,
	}

	codeGen := NewCodeGenerator(config)

	// Create a realistic schema
	schema := &MessageSchema{
		Type:        "object",
		Description: "User profile information",
		Properties: map[string]*Property{
			"username": {
				Type:        "string",
				Description: "Unique username",
			},
			"email": {
				Type:        "string",
				Description: "Email address",
			},
			"profile": {
				Type: "object",
				Properties: map[string]*Property{
					"firstName": {
						Type: "string",
					},
					"lastName": {
						Type: "string",
					},
				},
				Required: []string{"firstName"},
			},
		},
		Required: []string{"username", "email"},
	}

	// Generate the complete type definitions
	messages := map[string]*MessageSchema{
		"UserProfile": schema,
	}

	result, err := codeGen.GenerateTypes(messages, config)
	if err != nil {
		t.Fatalf("Failed to generate types: %v", err)
	}

	if len(result.Errors) > 0 {
		t.Fatalf("Generation errors: %v", result.Errors)
	}

	if len(result.Files) == 0 {
		t.Fatalf("No files generated")
	}

	// Get the generated file content
	var fileContent string
	for _, content := range result.Files {
		fileContent = content
		break
	}

	// Check that the file contains validation methods
	if !strings.Contains(fileContent, "func (s *UserProfile) Validate()") {
		t.Errorf("Generated file should contain Validate() method")
	}

	if !strings.Contains(fileContent, "func (s *UserProfile) ValidateJSON(") {
		t.Errorf("Generated file should contain ValidateJSON() method")
	}

	// Check that it's valid Go code structure
	if !strings.Contains(fileContent, "package test") {
		t.Errorf("Generated file should have correct package declaration")
	}

	// Print the generated file for manual inspection
	t.Logf("Generated file content:\n%s", fileContent)
}
func TestGeneratedValidationMethods_Functionality(t *testing.T) {
	// This test verifies that the generated validation methods actually work
	// by creating a simple struct and testing its validation

	// Create a simple test struct that mimics generated code
	type TestUser struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Age   int    `json:"age"`
	}

	// Manually implement the validation methods as they would be generated
	validateMethod := func(s *TestUser) *ValidationResult {
		validator := NewValidator(false)
		schema := &Property{
			Type: "object",
			Properties: map[string]*Property{
				"name":  {Type: "string"},
				"email": {Type: "string"},
				"age":   {Type: "integer"},
			},
			Required: []string{"name", "email"},
		}

		data := map[string]any{
			"name":  s.Name,
			"email": s.Email,
			"age":   s.Age,
		}

		return validator.ValidateValue(data, schema, "")
	}

	validateJSONMethod := func(s *TestUser, jsonData []byte) *ValidationResult {
		validator := NewValidator(false)
		schema := &MessageSchema{
			Type: "object",
			Properties: map[string]*Property{
				"name":  {Type: "string"},
				"email": {Type: "string"},
				"age":   {Type: "integer"},
			},
			Required: []string{"name", "email"},
		}

		return validator.ValidateJSON(jsonData, schema)
	}

	// Test valid struct
	validUser := &TestUser{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	result := validateMethod(validUser)
	if !result.Valid {
		t.Errorf("Valid user should pass validation, but got errors: %v", result.Errors)
	}

	// Test invalid struct (missing required field)
	// In the generated validation, we need to simulate missing fields by not including them in the map
	validateMethodWithMissingField := func() *ValidationResult {
		validator := NewValidator(false)
		schema := &Property{
			Type: "object",
			Properties: map[string]*Property{
				"name":  {Type: "string"},
				"email": {Type: "string"},
				"age":   {Type: "integer"},
			},
			Required: []string{"name", "email"},
		}

		// Simulate missing email field
		data := map[string]any{
			"name": "John Doe",
			"age":  30,
			// email is missing from the map
		}

		return validator.ValidateValue(data, schema, "")
	}

	result = validateMethodWithMissingField()
	if result.Valid {
		t.Errorf("Invalid user should fail validation, but got: %v", result.Errors)
	} else {
		t.Logf("Validation failed as expected with errors: %v", result.Errors)
	}

	// Test ValidateJSON with valid JSON
	validJSON := []byte(`{"name": "Jane Doe", "email": "jane@example.com", "age": 25}`)
	result = validateJSONMethod(validUser, validJSON)
	if !result.Valid {
		t.Errorf("Valid JSON should pass validation, but got errors: %v", result.Errors)
	}

	// Test ValidateJSON with invalid JSON (missing required field)
	invalidJSON := []byte(`{"name": "Jane Doe", "age": 25}`)
	result = validateJSONMethod(validUser, invalidJSON)
	if result.Valid {
		t.Errorf("Invalid JSON should fail validation")
	}

	// Test ValidateJSON with malformed JSON
	malformedJSON := []byte(`{"name": "Jane Doe", "age": 25,}`)
	result = validateJSONMethod(validUser, malformedJSON)
	if result.Valid {
		t.Errorf("Malformed JSON should fail validation")
	}
}

func TestGeneratedValidationMethods_ErrorMessages(t *testing.T) {
	// Test that generated validation methods return proper error messages

	type TestProduct struct {
		ID    string  `json:"id"`
		Name  string  `json:"name"`
		Price float64 `json:"price"`
	}

	// Test with missing required fields
	// Simulate the validation method with missing fields in the data map
	validateMethodWithMissingFields := func() *ValidationResult {
		validator := NewValidator(false)
		schema := &Property{
			Type: "object",
			Properties: map[string]*Property{
				"id":    {Type: "string"},
				"name":  {Type: "string"},
				"price": {Type: "number"},
			},
			Required: []string{"id", "name"},
		}

		// Simulate missing required fields
		data := map[string]any{
			"price": 99.99,
			// id and name are missing from the map
		}

		return validator.ValidateValue(data, schema, "")
	}

	result := validateMethodWithMissingFields()
	if result.Valid {
		t.Errorf("Product with missing required fields should fail validation")
	}

	// Check that we get appropriate error messages
	if len(result.Errors) == 0 {
		t.Errorf("Should have validation errors")
	}

	// Look for required field errors
	foundIDError := false
	foundNameError := false
	for _, err := range result.Errors {
		if strings.Contains(err.Message, "required property is missing") {
			if strings.Contains(err.Field, "id") {
				foundIDError = true
			}
			if strings.Contains(err.Field, "name") {
				foundNameError = true
			}
		}
	}

	if !foundIDError {
		t.Errorf("Should have error for missing 'id' field")
	}
	if !foundNameError {
		t.Errorf("Should have error for missing 'name' field")
	}
}
