# Validation Usage Example

This example demonstrates how to use the JSON validation functionality provided by the AsyncAPI Go Code Generator. The generated Go structs include built-in validation methods that allow you to validate data against the original AsyncAPI schema at runtime.

## Features Demonstrated

1. **Basic Struct Validation** - Validating Go struct instances using the `Validate()` method
2. **JSON Validation** - Validating raw JSON data using the `ValidateJSON()` method
3. **Constraint Validation** - Testing string length, numeric ranges, and other schema constraints
4. **Enum Validation** - Validating enum values with type safety
5. **EventBridge Validation** - Special validation for AWS EventBridge event structures
6. **Error Handling** - Comprehensive error handling and categorization
7. **Custom Validator Configuration** - Using strict vs permissive validation modes

## Running the Example

```bash
cd examples/validation_usage
go run main.go
```

## Expected Output

The example will demonstrate various validation scenarios:

```
=== AsyncAPI Go Code Generator - Validation Usage Example ===

1. Basic Struct Validation
✅ User struct validation passed!
✅ Invalid user correctly rejected: required property is missing

2. JSON Validation
✅ Valid JSON validation passed!
✅ Invalid JSON correctly rejected: required property is missing

3. Validation with Constraints
✅ String length validation passed!
✅ Short string correctly rejected: string length 2 is less than minimum 3
✅ Number range validation passed!
✅ High number correctly rejected: value 200 exceeds maximum 150

4. Enum Validation
✅ Valid enum value accepted!
✅ Invalid enum value correctly rejected: value must be one of: active, inactive, pending
✅ Generated enum types provide compile-time type safety!

5. EventBridge Validation
✅ EventBridge event structure validation passed!
✅ EventBridge event with detail schema validation passed!
✅ Invalid EventBridge event correctly rejected: required EventBridge field is missing or empty

6. Detailed Error Handling
Found 2 validation errors:
  1. Field 'userId': required property is missing
  2. Field 'profile.age': value -5 is less than minimum 0

Error categories:
  - Missing Required Field: required property is missing
  - Range Constraint: value -5 is less than minimum 0

7. Custom Validator Configuration
✅ Strict validator correctly rejected extra field: additional property not allowed in strict mode
✅ Permissive validator allowed extra field!
```

## Key Concepts

### Generated Validation Methods

Every generated struct includes two validation methods:

```go
// Validate validates the struct instance against its schema
func (s *UserStruct) Validate() *ValidationResult

// ValidateJSON validates raw JSON data against the schema
func (s *UserStruct) ValidateJSON(jsonData []byte) *ValidationResult
```

### Validation Result

The validation methods return a `ValidationResult` with detailed error information:

```go
type ValidationResult struct {
    Valid  bool               `json:"valid"`
    Errors []*ValidationError `json:"errors,omitempty"`
}

type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}
```

### Validator Configuration

You can create validators with different configurations:

```go
// Strict mode - rejects additional properties
strictValidator := generator.NewValidator(true)

// Permissive mode - allows additional properties (default)
permissiveValidator := generator.NewValidator(false)
```

### EventBridge Support

For AWS EventBridge events, use the specialized validator:

```go
validator := generator.NewEventBridgeValidator()

// Validate EventBridge event structure
result := validator.ValidateEventBridgeEvent(eventJSON)

// Validate with detail schema
result = validator.ValidateEventBridgeEventWithSchema(eventJSON, detailSchema)
```

## Real-World Usage

In a real application, you would:

1. Generate Go structs from your AsyncAPI specification
2. Use the generated validation methods to validate incoming data
3. Handle validation errors appropriately in your application logic
4. Configure validators based on your application's requirements

Example integration:

```go
func handleUserSignup(w http.ResponseWriter, r *http.Request) {
    var payload UserSignupPayload
    
    // Parse JSON request
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    // Validate against schema
    result := payload.Validate()
    if !result.Valid {
        // Return validation errors
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Validation failed",
            "details": result.Errors,
        })
        return
    }
    
    // Process valid payload
    processUserSignup(&payload)
    w.WriteHeader(http.StatusCreated)
}
```

## Benefits

- **Runtime Safety**: Catch schema violations at runtime
- **Type Safety**: Generated enum types provide compile-time safety
- **Detailed Errors**: Get specific field-level error messages
- **Flexible Configuration**: Choose strict or permissive validation modes
- **EventBridge Support**: Built-in support for AWS EventBridge event validation
- **Easy Integration**: Simple API that integrates seamlessly with existing code