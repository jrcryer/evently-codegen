package generator

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEventBridgeValidator_ValidateEventBridgeEvent(t *testing.T) {
	validator := NewEventBridgeValidator()

	tests := []struct {
		name      string
		eventJSON string
		wantErr   bool
		errFields []string
	}{
		{
			name: "valid EventBridge event",
			eventJSON: `{
				"version": "0",
				"id": "12345678-1234-1234-1234-123456789012",
				"detail-type": "User Registration",
				"source": "com.example.userservice",
				"account": "123456789012",
				"time": "2023-01-01T12:00:00Z",
				"region": "us-east-1",
				"detail": {
					"userId": "user123",
					"email": "user@example.com"
				}
			}`,
			wantErr: false,
		},
		{
			name: "missing required fields",
			eventJSON: `{
				"version": "0",
				"id": "12345678-1234-1234-1234-123456789012"
			}`,
			wantErr:   true,
			errFields: []string{"detail-type", "source", "account", "time", "region", "detail"},
		},
		{
			name: "invalid version",
			eventJSON: `{
				"version": "1",
				"id": "12345678-1234-1234-1234-123456789012",
				"detail-type": "User Registration",
				"source": "com.example.userservice",
				"account": "123456789012",
				"time": "2023-01-01T12:00:00Z",
				"region": "us-east-1",
				"detail": {"test": "data"}
			}`,
			wantErr:   true,
			errFields: []string{"version"},
		},
		{
			name: "invalid source format",
			eventJSON: `{
				"version": "0",
				"id": "12345678-1234-1234-1234-123456789012",
				"detail-type": "User Registration",
				"source": "invalid source with spaces",
				"account": "123456789012",
				"time": "2023-01-01T12:00:00Z",
				"region": "us-east-1",
				"detail": {"test": "data"}
			}`,
			wantErr:   true,
			errFields: []string{"source"},
		},
		{
			name: "invalid account format",
			eventJSON: `{
				"version": "0",
				"id": "12345678-1234-1234-1234-123456789012",
				"detail-type": "User Registration",
				"source": "com.example.userservice",
				"account": "invalid",
				"time": "2023-01-01T12:00:00Z",
				"region": "us-east-1",
				"detail": {"test": "data"}
			}`,
			wantErr:   true,
			errFields: []string{"account"},
		},
		{
			name: "invalid region format",
			eventJSON: `{
				"version": "0",
				"id": "12345678-1234-1234-1234-123456789012",
				"detail-type": "User Registration",
				"source": "com.example.userservice",
				"account": "123456789012",
				"time": "2023-01-01T12:00:00Z",
				"region": "invalid-region",
				"detail": {"test": "data"}
			}`,
			wantErr:   true,
			errFields: []string{"region"},
		},
		{
			name: "empty detail",
			eventJSON: `{
				"version": "0",
				"id": "12345678-1234-1234-1234-123456789012",
				"detail-type": "User Registration",
				"source": "com.example.userservice",
				"account": "123456789012",
				"time": "2023-01-01T12:00:00Z",
				"region": "us-east-1",
				"detail": {}
			}`,
			wantErr:   true,
			errFields: []string{"detail"},
		},
		{
			name: "malformed JSON",
			eventJSON: `{
				"version": "0",
				"id": "12345678-1234-1234-1234-123456789012",
				"detail-type": "User Registration",
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateEventBridgeEvent([]byte(tt.eventJSON))

			if tt.wantErr && result.Valid {
				t.Errorf("expected validation error, but validation passed")
			}

			if !tt.wantErr && !result.Valid {
				t.Errorf("expected validation to pass, but got errors: %v", result.Errors)
			}

			if tt.wantErr && len(tt.errFields) > 0 {
				for _, expectedField := range tt.errFields {
					found := false
					for _, err := range result.Errors {
						if err.Field == expectedField {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error for field '%s', but not found in: %v", expectedField, result.Errors)
					}
				}
			}
		})
	}
}

func TestEventBridgeValidator_ValidateEventBridgeEventWithSchema(t *testing.T) {
	validator := NewEventBridgeValidator()

	// Create a schema for the detail payload
	detailSchema := &MessageSchema{
		Type: "object",
		Properties: map[string]*Property{
			"userId": {
				Type:        "string",
				Description: "User identifier",
			},
			"email": {
				Type:        "string",
				Description: "User email",
			},
			"age": {
				Type:    "integer",
				Minimum: float64Ptr(0),
			},
		},
		Required: []string{"userId", "email"},
	}

	tests := []struct {
		name      string
		eventJSON string
		wantErr   bool
		errFields []string
	}{
		{
			name: "valid event with valid detail",
			eventJSON: `{
				"version": "0",
				"id": "12345678-1234-1234-1234-123456789012",
				"detail-type": "User Registration",
				"source": "com.example.userservice",
				"account": "123456789012",
				"time": "2023-01-01T12:00:00Z",
				"region": "us-east-1",
				"detail": {
					"userId": "user123",
					"email": "user@example.com",
					"age": 25
				}
			}`,
			wantErr: false,
		},
		{
			name: "valid event with invalid detail",
			eventJSON: `{
				"version": "0",
				"id": "12345678-1234-1234-1234-123456789012",
				"detail-type": "User Registration",
				"source": "com.example.userservice",
				"account": "123456789012",
				"time": "2023-01-01T12:00:00Z",
				"region": "us-east-1",
				"detail": {
					"userId": "user123"
				}
			}`,
			wantErr:   true,
			errFields: []string{"detail.email"},
		},
		{
			name: "valid event with detail constraint violation",
			eventJSON: `{
				"version": "0",
				"id": "12345678-1234-1234-1234-123456789012",
				"detail-type": "User Registration",
				"source": "com.example.userservice",
				"account": "123456789012",
				"time": "2023-01-01T12:00:00Z",
				"region": "us-east-1",
				"detail": {
					"userId": "user123",
					"email": "user@example.com",
					"age": -5
				}
			}`,
			wantErr:   true,
			errFields: []string{"detail.age"},
		},
		{
			name: "invalid event structure",
			eventJSON: `{
				"version": "0",
				"id": "12345678-1234-1234-1234-123456789012"
			}`,
			wantErr:   true,
			errFields: []string{"detail-type", "source", "account", "time", "region", "detail"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateEventBridgeEventWithSchema([]byte(tt.eventJSON), detailSchema)

			if tt.wantErr && result.Valid {
				t.Errorf("expected validation error, but validation passed")
			}

			if !tt.wantErr && !result.Valid {
				t.Errorf("expected validation to pass, but got errors: %v", result.Errors)
			}

			if tt.wantErr && len(tt.errFields) > 0 {
				for _, expectedField := range tt.errFields {
					found := false
					for _, err := range result.Errors {
						if err.Field == expectedField {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error for field '%s', but not found in: %v", expectedField, result.Errors)
					}
				}
			}
		})
	}
}

func TestEventBridgeValidator_ValidateEventPattern(t *testing.T) {
	validator := NewEventBridgeValidator()

	tests := []struct {
		name        string
		patternJSON string
		wantErr     bool
		errFields   []string
	}{
		{
			name: "valid simple pattern",
			patternJSON: `{
				"source": ["com.example.userservice"],
				"detail-type": ["User Registration"]
			}`,
			wantErr: false,
		},
		{
			name: "valid complex pattern with matching rules",
			patternJSON: `{
				"source": ["com.example.userservice"],
				"detail-type": ["User Registration"],
				"detail": {
					"userId": {"exists": true},
					"email": {"prefix": "admin@"}
				}
			}`,
			wantErr: false,
		},
		{
			name: "invalid pattern field",
			patternJSON: `{
				"source": ["com.example.userservice"],
				"invalid-field": ["value"]
			}`,
			wantErr:   true,
			errFields: []string{"invalid-field"},
		},
		{
			name: "invalid matching rule",
			patternJSON: `{
				"source": ["com.example.userservice"],
				"detail": {
					"userId": {"invalid-rule": true}
				}
			}`,
			wantErr:   true,
			errFields: []string{"detail.userId.invalid-rule"},
		},
		{
			name: "empty array pattern",
			patternJSON: `{
				"source": []
			}`,
			wantErr:   true,
			errFields: []string{"source"},
		},
		{
			name: "malformed JSON",
			patternJSON: `{
				"source": ["com.example.userservice",
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateEventPattern([]byte(tt.patternJSON))

			if tt.wantErr && result.Valid {
				t.Errorf("expected validation error, but validation passed")
			}

			if !tt.wantErr && !result.Valid {
				t.Errorf("expected validation to pass, but got errors: %v", result.Errors)
			}

			if tt.wantErr && len(tt.errFields) > 0 {
				for _, expectedField := range tt.errFields {
					found := false
					for _, err := range result.Errors {
						if err.Field == expectedField {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error for field '%s', but not found in: %v", expectedField, result.Errors)
					}
				}
			}
		})
	}
}

func TestEventBridgeValidator_HelperFunctions(t *testing.T) {
	validator := NewEventBridgeValidator()

	// Test isNumeric
	tests := []struct {
		input    string
		expected bool
	}{
		{"123456789012", true},
		{"12345", true},
		{"abc123", false},
		{"123abc", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run("isNumeric_"+tt.input, func(t *testing.T) {
			result := validator.isNumeric(tt.input)
			if result != tt.expected {
				t.Errorf("isNumeric(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}

	// Test isValidAWSRegion
	regionTests := []struct {
		input    string
		expected bool
	}{
		{"us-east-1", true},
		{"eu-west-2", true},
		{"ap-southeast-1", true},
		{"ca-central-1", true},
		{"us-gov-east-1", true},
		{"invalid-region", false},
		{"us", false},
		{"", false},
		{"toolongregionnamethatexceedslimit", false},
	}

	for _, tt := range regionTests {
		t.Run("isValidAWSRegion_"+tt.input, func(t *testing.T) {
			result := validator.isValidAWSRegion(tt.input)
			if result != tt.expected {
				t.Errorf("isValidAWSRegion(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}

	// Test isEmptyValue
	emptyTests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"empty string", "", true},
		{"non-empty string", "test", false},
		{"empty map", map[string]interface{}{}, true},
		{"non-empty map", map[string]interface{}{"key": "value"}, false},
		{"empty slice", []string{}, true},
		{"non-empty slice", []string{"item"}, false},
		{"zero time", time.Time{}, true},
		{"non-zero time", time.Now(), false},
		{"nil", nil, true},
	}

	for _, tt := range emptyTests {
		t.Run("isEmptyValue_"+tt.name, func(t *testing.T) {
			result := validator.isEmptyValue(tt.input)
			if result != tt.expected {
				t.Errorf("isEmptyValue(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCreateEventBridgeEventSchema(t *testing.T) {
	schema := CreateEventBridgeEventSchema()

	// Check that the schema has the correct structure
	if schema.Type != "object" {
		t.Errorf("expected schema type 'object', got '%s'", schema.Type)
	}

	// Check required fields
	expectedRequired := []string{"version", "id", "detail-type", "source", "account", "time", "region", "detail"}
	if len(schema.Required) != len(expectedRequired) {
		t.Errorf("expected %d required fields, got %d", len(expectedRequired), len(schema.Required))
	}

	for _, field := range expectedRequired {
		found := false
		for _, req := range schema.Required {
			if req == field {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected required field '%s' not found", field)
		}
	}

	// Check that all required fields have properties defined
	for _, field := range schema.Required {
		if _, exists := schema.Properties[field]; !exists {
			t.Errorf("required field '%s' does not have a property definition", field)
		}
	}

	// Check specific property types
	if schema.Properties["version"].Type != "string" {
		t.Errorf("expected version property to be string")
	}

	if schema.Properties["detail"].Type != "object" {
		t.Errorf("expected detail property to be object")
	}

	if schema.Properties["resources"].Type != "array" {
		t.Errorf("expected resources property to be array")
	}
}

func TestEventBridgeValidator_Integration(t *testing.T) {
	validator := NewEventBridgeValidator()

	// Create a realistic EventBridge event
	event := EventBridgeEvent{
		Version:    "0",
		ID:         "12345678-1234-1234-1234-123456789012",
		DetailType: "User Registration",
		Source:     "com.example.userservice",
		Account:    "123456789012",
		Time:       time.Now(),
		Region:     "us-east-1",
		Detail: map[string]interface{}{
			"userId": "user123",
			"email":  "user@example.com",
			"age":    25,
		},
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal event: %v", err)
	}

	// Test basic EventBridge validation
	result := validator.ValidateEventBridgeEvent(eventJSON)
	if !result.Valid {
		t.Errorf("valid EventBridge event should pass validation, but got errors: %v", result.Errors)
	}

	// Test with schema validation
	detailSchema := &MessageSchema{
		Type: "object",
		Properties: map[string]*Property{
			"userId": {Type: "string"},
			"email":  {Type: "string"},
			"age":    {Type: "integer"},
		},
		Required: []string{"userId", "email"},
	}

	result = validator.ValidateEventBridgeEventWithSchema(eventJSON, detailSchema)
	if !result.Valid {
		t.Errorf("valid EventBridge event with schema should pass validation, but got errors: %v", result.Errors)
	}

	// Test event pattern validation
	pattern := map[string]interface{}{
		"source":      []string{"com.example.userservice"},
		"detail-type": []string{"User Registration"},
		"detail": map[string]interface{}{
			"userId": map[string]interface{}{"exists": true},
		},
	}

	patternJSON, err := json.Marshal(pattern)
	if err != nil {
		t.Fatalf("failed to marshal pattern: %v", err)
	}

	result = validator.ValidateEventPattern(patternJSON)
	if !result.Valid {
		t.Errorf("valid EventBridge pattern should pass validation, but got errors: %v", result.Errors)
	}
}
