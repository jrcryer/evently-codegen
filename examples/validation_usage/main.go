package main

import (
	"fmt"

	"github.com/jrcryer/evently-codegen/pkg/generator"
)

func main() {
	fmt.Println("=== AsyncAPI Go Code Generator - Validation Usage Example ===\n")

	// Example 1: Basic struct validation
	fmt.Println("1. Basic Struct Validation")
	basicStructValidationExample()

	// Example 2: JSON validation
	fmt.Println("\n2. JSON Validation")
	jsonValidationExample()

	// Example 3: Validation with constraints
	fmt.Println("\n3. Validation with Constraints")
	constraintValidationExample()

	// Example 4: Enum validation
	fmt.Println("\n4. Enum Validation")
	enumValidationExample()

	// Example 5: EventBridge validation
	fmt.Println("\n5. EventBridge Validation")
	eventBridgeValidationExample()

	// Example 6: Error handling
	fmt.Println("\n6. Detailed Error Handling")
	errorHandlingExample()

	// Example 7: Custom validator configuration
	fmt.Println("\n7. Custom Validator Configuration")
	customValidatorExample()
}

// basicStructValidationExample demonstrates validating a struct instance
func basicStructValidationExample() {
	// Create a sample user struct (simulating generated code)
	user := createSampleUser("user123", "john@example.com", "John", "Doe", 30)

	// Validate the struct using the generated Validate() method
	result := validateUserStruct(user)

	if result.Valid {
		fmt.Println("✅ User struct validation passed!")
	} else {
		fmt.Printf("❌ User struct validation failed: %v\n", result.Errors)
	}

	// Test with invalid data (missing required field)
	invalidUser := createSampleUser("", "john@example.com", "John", "Doe", 30) // Missing userId
	result = validateUserStruct(invalidUser)

	if !result.Valid {
		fmt.Printf("✅ Invalid user correctly rejected: %v\n", result.Errors[0].Message)
	}
}

// jsonValidationExample demonstrates validating JSON data
func jsonValidationExample() {
	// Valid JSON data
	validJSON := []byte(`{
		"userId": "user456",
		"email": "jane@example.com",
		"profile": {
			"firstName": "Jane",
			"lastName": "Smith",
			"age": 25
		}
	}`)

	// Validate JSON using the generated ValidateJSON() method
	result := validateUserJSON(validJSON)

	if result.Valid {
		fmt.Println("✅ Valid JSON validation passed!")
	} else {
		fmt.Printf("❌ Valid JSON validation failed: %v\n", result.Errors)
	}

	// Invalid JSON data (missing required field)
	invalidJSON := []byte(`{
		"email": "jane@example.com",
		"profile": {
			"firstName": "Jane",
			"age": 25
		}
	}`)

	result = validateUserJSON(invalidJSON)

	if !result.Valid {
		fmt.Printf("✅ Invalid JSON correctly rejected: %s\n", result.Errors[0].Message)
	}
}

// constraintValidationExample demonstrates validation with schema constraints
func constraintValidationExample() {
	// Create validator for testing constraints
	validator := generator.NewValidator(false)

	// Test string length constraints
	stringSchema := &generator.Property{
		Type:      "string",
		MinLength: intPtr(3),
		MaxLength: intPtr(10),
	}

	// Valid string
	result := validator.ValidateValue("hello", stringSchema, "testField")
	if result.Valid {
		fmt.Println("✅ String length validation passed!")
	}

	// Invalid string (too short)
	result = validator.ValidateValue("hi", stringSchema, "testField")
	if !result.Valid {
		fmt.Printf("✅ Short string correctly rejected: %s\n", result.Errors[0].Message)
	}

	// Test numeric constraints
	numberSchema := &generator.Property{
		Type:    "integer",
		Minimum: float64Ptr(0),
		Maximum: float64Ptr(150),
	}

	// Valid number
	result = validator.ValidateValue(25, numberSchema, "age")
	if result.Valid {
		fmt.Println("✅ Number range validation passed!")
	}

	// Invalid number (too high)
	result = validator.ValidateValue(200, numberSchema, "age")
	if !result.Valid {
		fmt.Printf("✅ High number correctly rejected: %s\n", result.Errors[0].Message)
	}
}

// enumValidationExample demonstrates enum validation
func enumValidationExample() {
	validator := generator.NewValidator(false)

	// Create enum schema
	enumSchema := &generator.Property{
		Type: "string",
		Enum: []interface{}{"active", "inactive", "pending"},
	}

	// Valid enum value
	result := validator.ValidateValue("active", enumSchema, "status")
	if result.Valid {
		fmt.Println("✅ Valid enum value accepted!")
	}

	// Invalid enum value
	result = validator.ValidateValue("unknown", enumSchema, "status")
	if !result.Valid {
		fmt.Printf("✅ Invalid enum value correctly rejected: %s\n", result.Errors[0].Message)
	}

	// Demonstrate generated enum type validation (simulated)
	fmt.Println("✅ Generated enum types provide compile-time type safety!")
}

// eventBridgeValidationExample demonstrates EventBridge-specific validation
func eventBridgeValidationExample() {
	validator := generator.NewEventBridgeValidator()

	// Valid EventBridge event
	validEventJSON := []byte(`{
		"version": "0",
		"id": "12345678-1234-1234-1234-123456789012",
		"detail-type": "User Signup",
		"source": "com.example.userservice",
		"account": "123456789012",
		"time": "2023-01-01T12:00:00Z",
		"region": "us-east-1",
		"detail": {
			"userId": "user123",
			"email": "user@example.com"
		}
	}`)

	result := validator.ValidateEventBridgeEvent(validEventJSON)
	if result.Valid {
		fmt.Println("✅ EventBridge event structure validation passed!")
	} else {
		fmt.Printf("❌ EventBridge validation failed: %v\n", result.Errors)
	}

	// Validate with detail schema
	detailSchema := &generator.MessageSchema{
		Type: "object",
		Properties: map[string]*generator.Property{
			"userId": {Type: "string"},
			"email":  {Type: "string"},
		},
		Required: []string{"userId", "email"},
	}

	result = validator.ValidateEventBridgeEventWithSchema(validEventJSON, detailSchema)
	if result.Valid {
		fmt.Println("✅ EventBridge event with detail schema validation passed!")
	} else {
		fmt.Printf("❌ EventBridge detail validation failed: %v\n", result.Errors)
	}

	// Invalid EventBridge event (missing required field)
	invalidEventJSON := []byte(`{
		"version": "0",
		"id": "12345678-1234-1234-1234-123456789012",
		"detail-type": "User Signup",
		"source": "com.example.userservice",
		"account": "123456789012",
		"time": "2023-01-01T12:00:00Z",
		"detail": {
			"userId": "user123",
			"email": "user@example.com"
		}
	}`)

	result = validator.ValidateEventBridgeEvent(invalidEventJSON)
	if !result.Valid {
		fmt.Printf("✅ Invalid EventBridge event correctly rejected: %s\n", result.Errors[0].Message)
	}
}

// errorHandlingExample demonstrates detailed error handling
func errorHandlingExample() {
	// Create complex invalid JSON to generate multiple errors
	invalidJSON := []byte(`{
		"email": "not-an-email",
		"profile": {
			"age": -5
		}
	}`)

	result := validateUserJSON(invalidJSON)

	if !result.Valid {
		fmt.Printf("Found %d validation errors:\n", len(result.Errors))
		for i, err := range result.Errors {
			fmt.Printf("  %d. Field '%s': %s\n", i+1, err.Field, err.Message)
		}
	}

	// Demonstrate error categorization
	fmt.Println("\nError categories:")
	for _, err := range result.Errors {
		category := categorizeError(err)
		fmt.Printf("  - %s: %s\n", category, err.Message)
	}
}

// customValidatorExample demonstrates different validator configurations
func customValidatorExample() {
	// Strict mode validator (rejects additional properties)
	strictValidator := generator.NewValidator(true)

	// Permissive mode validator (allows additional properties)
	permissiveValidator := generator.NewValidator(false)

	// JSON with additional property
	jsonWithExtra := []byte(`{
		"userId": "user123",
		"email": "user@example.com",
		"extraField": "this should not be here"
	}`)

	schema := createUserSchema()

	// Test with strict validator
	result := strictValidator.ValidateJSON(jsonWithExtra, schema)
	if !result.Valid {
		fmt.Printf("✅ Strict validator correctly rejected extra field: %s\n", result.Errors[0].Message)
	}

	// Test with permissive validator
	result = permissiveValidator.ValidateJSON(jsonWithExtra, schema)
	if result.Valid {
		fmt.Println("✅ Permissive validator allowed extra field!")
	}
}

// Helper functions to simulate generated code behavior

// createSampleUser creates a sample user struct
func createSampleUser(userId, email, firstName, lastName string, age int) map[string]interface{} {
	user := map[string]interface{}{
		"userId": userId,
		"email":  email,
	}

	if firstName != "" || lastName != "" || age > 0 {
		profile := make(map[string]interface{})
		if firstName != "" {
			profile["firstName"] = firstName
		}
		if lastName != "" {
			profile["lastName"] = lastName
		}
		if age > 0 {
			profile["age"] = age
		}
		user["profile"] = profile
	}

	return user
}

// validateUserStruct simulates the generated Validate() method
func validateUserStruct(user map[string]interface{}) *generator.ValidationResult {
	validator := generator.NewValidator(false)
	schema := &generator.Property{
		Type: "object",
		Properties: map[string]*generator.Property{
			"userId": {Type: "string"},
			"email":  {Type: "string"},
			"profile": {
				Type: "object",
				Properties: map[string]*generator.Property{
					"firstName": {Type: "string"},
					"lastName":  {Type: "string"},
					"age":       {Type: "integer"},
				},
			},
		},
		Required: []string{"userId", "email"},
	}

	return validator.ValidateValue(user, schema, "")
}

// validateUserJSON simulates the generated ValidateJSON() method
func validateUserJSON(jsonData []byte) *generator.ValidationResult {
	validator := generator.NewValidator(false)
	schema := createUserSchema()
	return validator.ValidateJSON(jsonData, schema)
}

// createUserSchema creates a user message schema
func createUserSchema() *generator.MessageSchema {
	return &generator.MessageSchema{
		Type: "object",
		Properties: map[string]*generator.Property{
			"userId": {Type: "string"},
			"email":  {Type: "string"},
			"profile": {
				Type: "object",
				Properties: map[string]*generator.Property{
					"firstName": {Type: "string"},
					"lastName":  {Type: "string"},
					"age":       {Type: "integer"},
				},
			},
		},
		Required: []string{"userId", "email"},
	}
}

// categorizeError categorizes validation errors
func categorizeError(err *generator.ValidationError) string {
	switch {
	case err.Field == "" && err.Message == "required property is missing":
		return "Missing Required Field"
	case err.Field != "" && err.Message == "required property is missing":
		return "Missing Required Field"
	case contains(err.Message, "expected"):
		return "Type Mismatch"
	case contains(err.Message, "length"):
		return "Length Constraint"
	case contains(err.Message, "minimum") || contains(err.Message, "maximum"):
		return "Range Constraint"
	case contains(err.Message, "one of"):
		return "Enum Constraint"
	default:
		return "Other"
	}
}

// Helper functions

func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
