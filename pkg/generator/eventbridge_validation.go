package generator

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// EventBridgeEvent represents the structure of an AWS EventBridge event
type EventBridgeEvent struct {
	Version    string                 `json:"version"`
	ID         string                 `json:"id"`
	DetailType string                 `json:"detail-type"`
	Source     string                 `json:"source"`
	Account    string                 `json:"account"`
	Time       time.Time              `json:"time"`
	Region     string                 `json:"region"`
	Detail     map[string]interface{} `json:"detail"`
	Resources  []string               `json:"resources,omitempty"`
}

// EventBridgeValidator provides specialized validation functionality for AWS EventBridge events.
// It validates both the EventBridge event structure (version, id, detail-type, source, etc.)
// and the event detail payload against AsyncAPI schemas.
//
// EventBridge events have a specific structure that must be followed:
//   - version: Must be "0"
//   - id: Unique event identifier
//   - detail-type: Descriptive event type
//   - source: Event source in reverse DNS notation
//   - account: 12-digit AWS account ID
//   - time: ISO 8601 timestamp
//   - region: Valid AWS region
//   - detail: Event payload object
//
// Example usage:
//
//	validator := NewEventBridgeValidator()
//
//	// Validate EventBridge event structure
//	result := validator.ValidateEventBridgeEvent(eventJSON)
//
//	// Validate with detail schema
//	result = validator.ValidateEventBridgeEventWithSchema(eventJSON, detailSchema)
type EventBridgeValidator struct {
	// baseValidator is used for validating the detail payload against AsyncAPI schemas
	baseValidator Validator
}

// NewEventBridgeValidator creates a new EventBridgeValidator instance.
// The validator uses a permissive base validator for detail payload validation,
// allowing additional properties in the detail object.
//
// Returns:
//   - A new EventBridgeValidator instance ready for validating EventBridge events
//
// Example:
//
//	validator := NewEventBridgeValidator()
//	result := validator.ValidateEventBridgeEvent(eventJSON)
func NewEventBridgeValidator() *EventBridgeValidator {
	return &EventBridgeValidator{
		baseValidator: NewValidator(false),
	}
}

// ValidateEventBridgeEvent validates the structure and format of an AWS EventBridge event.
// This method validates the required EventBridge fields and their formats, but does not
// validate the detail payload against a specific schema.
//
// Parameters:
//   - eventData: Raw JSON data representing the EventBridge event
//
// Returns:
//   - ValidationResult indicating whether the event structure is valid
//
// The method validates:
//   - Required fields: version, id, detail-type, source, account, time, region, detail
//   - Field formats: version should be "0", account should be 12-digit number, etc.
//   - Source format: should follow reverse DNS notation
//   - Region format: should be a valid AWS region format
//
// Example:
//
//	eventJSON := []byte(`{
//	    "version": "0",
//	    "id": "12345678-1234-1234-1234-123456789012",
//	    "detail-type": "User Signup",
//	    "source": "com.example.userservice",
//	    "account": "123456789012",
//	    "time": "2023-01-01T12:00:00Z",
//	    "region": "us-east-1",
//	    "detail": {"userId": "user123"}
//	}`)
//	result := validator.ValidateEventBridgeEvent(eventJSON)
func (v *EventBridgeValidator) ValidateEventBridgeEvent(eventData []byte) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Parse the event data
	var event EventBridgeEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		result.AddError("", fmt.Sprintf("invalid EventBridge event JSON: %v", err), nil)
		return result
	}

	// Validate required EventBridge fields
	v.validateRequiredEventBridgeFields(&event, result)

	// Validate field formats
	v.validateEventBridgeFieldFormats(&event, result)

	return result
}

// ValidateEventBridgeEventWithSchema validates both the EventBridge event structure
// and the detail payload against a provided AsyncAPI schema.
//
// Parameters:
//   - eventData: Raw JSON data representing the EventBridge event
//   - detailSchema: AsyncAPI MessageSchema to validate the detail payload against
//
// Returns:
//   - ValidationResult indicating whether both the event structure and detail payload are valid
//
// This method first validates the EventBridge event structure using ValidateEventBridgeEvent,
// then validates the detail payload against the provided schema. All validation errors
// from both validations are included in the result.
//
// Example:
//
//	detailSchema := &MessageSchema{
//	    Type: "object",
//	    Properties: map[string]*Property{
//	        "userId": {Type: "string"},
//	        "email":  {Type: "string"},
//	    },
//	    Required: []string{"userId", "email"},
//	}
//	result := validator.ValidateEventBridgeEventWithSchema(eventJSON, detailSchema)
func (v *EventBridgeValidator) ValidateEventBridgeEventWithSchema(eventData []byte, detailSchema *MessageSchema) *ValidationResult {
	result := v.ValidateEventBridgeEvent(eventData)
	if !result.Valid {
		return result
	}

	// Parse the event to get the detail payload
	var event EventBridgeEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		result.AddError("", fmt.Sprintf("failed to parse event for detail validation: %v", err), nil)
		return result
	}

	// Validate the detail payload against the provided schema
	if detailSchema != nil {
		detailResult := v.baseValidator.ValidateMessage(event.Detail, detailSchema)
		if !detailResult.Valid {
			// Prefix all detail validation errors with "detail."
			for _, err := range detailResult.Errors {
				prefixedField := "detail"
				if err.Field != "" {
					prefixedField = "detail." + err.Field
				}
				result.AddError(prefixedField, err.Message, nil)
			}
			result.Valid = false
		}
	}

	return result
}

// ValidateEventPattern validates an EventBridge event pattern
func (v *EventBridgeValidator) ValidateEventPattern(patternData []byte) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Parse the pattern data
	var pattern map[string]interface{}
	if err := json.Unmarshal(patternData, &pattern); err != nil {
		result.AddError("", fmt.Sprintf("invalid event pattern JSON: %v", err), nil)
		return result
	}

	// Validate pattern structure
	v.validateEventPatternStructure(pattern, result)

	return result
}

// validateRequiredEventBridgeFields validates that all required EventBridge fields are present
func (v *EventBridgeValidator) validateRequiredEventBridgeFields(event *EventBridgeEvent, result *ValidationResult) {
	requiredFields := map[string]interface{}{
		"version":     event.Version,
		"id":          event.ID,
		"detail-type": event.DetailType,
		"source":      event.Source,
		"account":     event.Account,
		"time":        event.Time,
		"region":      event.Region,
		"detail":      event.Detail,
	}

	for fieldName, fieldValue := range requiredFields {
		if v.isEmptyValue(fieldValue) {
			result.AddError(fieldName, "required EventBridge field is missing or empty", fieldValue)
		}
	}
}

// validateEventBridgeFieldFormats validates the format of EventBridge fields
func (v *EventBridgeValidator) validateEventBridgeFieldFormats(event *EventBridgeEvent, result *ValidationResult) {
	// Validate version format
	if event.Version != "" && event.Version != "0" {
		result.AddError("version", "EventBridge version should be '0'", event.Version)
	}

	// Validate ID format (should be a UUID-like string)
	if event.ID != "" && len(event.ID) < 8 {
		result.AddError("id", "EventBridge ID should be a valid identifier", event.ID)
	}

	// Validate source format (should not contain spaces and follow reverse DNS notation)
	if event.Source != "" {
		if strings.Contains(event.Source, " ") {
			result.AddError("source", "EventBridge source should not contain spaces", event.Source)
		}
		if !strings.Contains(event.Source, ".") {
			result.AddError("source", "EventBridge source should follow reverse DNS notation (e.g., com.example.myapp)", event.Source)
		}
	}

	// Validate detail-type format (should be descriptive)
	if event.DetailType != "" && len(event.DetailType) < 3 {
		result.AddError("detail-type", "EventBridge detail-type should be descriptive", event.DetailType)
	}

	// Validate account format (should be 12-digit AWS account ID)
	if event.Account != "" && (len(event.Account) != 12 || !v.isNumeric(event.Account)) {
		result.AddError("account", "EventBridge account should be a 12-digit AWS account ID", event.Account)
	}

	// Validate region format (should be valid AWS region)
	if event.Region != "" && !v.isValidAWSRegion(event.Region) {
		result.AddError("region", "EventBridge region should be a valid AWS region", event.Region)
	}

	// Validate detail is not empty
	if event.Detail != nil && len(event.Detail) == 0 {
		result.AddError("detail", "EventBridge detail should not be empty", event.Detail)
	}
}

// validateEventPatternStructure validates the structure of an EventBridge event pattern
func (v *EventBridgeValidator) validateEventPatternStructure(pattern map[string]interface{}, result *ValidationResult) {
	// EventBridge patterns can contain various matching rules
	validPatternFields := map[string]bool{
		"source":         true,
		"detail-type":    true,
		"detail":         true,
		"account":        true,
		"region":         true,
		"time":           true,
		"id":             true,
		"version":        true,
		"resources":      true,
		"replay-name":    true,
		"ingestion-time": true,
	}

	for fieldName := range pattern {
		if !validPatternFields[fieldName] {
			result.AddError(fieldName, fmt.Sprintf("'%s' is not a valid EventBridge pattern field", fieldName), pattern[fieldName])
		}
	}

	// Validate pattern matching rules
	for fieldName, fieldValue := range pattern {
		v.validatePatternMatchingRules(fieldName, fieldValue, result)
	}
}

// validatePatternMatchingRules validates EventBridge pattern matching rules
func (v *EventBridgeValidator) validatePatternMatchingRules(fieldName string, fieldValue interface{}, result *ValidationResult) {
	switch value := fieldValue.(type) {
	case []interface{}:
		// Array values are valid (OR matching)
		if len(value) == 0 {
			result.AddError(fieldName, "pattern array should not be empty", fieldValue)
		}
	case map[string]interface{}:
		// Object values can contain either:
		// 1. Nested field patterns (for detail object)
		// 2. Matching rules (exists, prefix, etc.)

		// Check if this looks like a matching rules object
		isMatchingRules := v.isMatchingRulesObject(value)

		if isMatchingRules {
			// Validate matching rules
			for ruleName := range value {
				validRules := map[string]bool{
					"exists":       true,
					"prefix":       true,
					"anything-but": true,
					"numeric":      true,
				}
				if !validRules[ruleName] {
					result.AddError(fmt.Sprintf("%s.%s", fieldName, ruleName),
						fmt.Sprintf("'%s' is not a valid EventBridge pattern matching rule", ruleName), value[ruleName])
				}
			}
		} else {
			// This is a nested field pattern, recursively validate
			for nestedFieldName, nestedFieldValue := range value {
				nestedPath := fmt.Sprintf("%s.%s", fieldName, nestedFieldName)
				v.validatePatternMatchingRules(nestedPath, nestedFieldValue, result)
			}
		}
	case string, float64, bool:
		// Primitive values are valid (exact matching)
	default:
		result.AddError(fieldName, "invalid pattern value type", fieldValue)
	}
}

// isMatchingRulesObject determines if a map contains matching rules vs nested field patterns
func (v *EventBridgeValidator) isMatchingRulesObject(obj map[string]interface{}) bool {
	validMatchingRules := map[string]bool{
		"exists":       true,
		"prefix":       true,
		"anything-but": true,
		"numeric":      true,
	}

	// If any key is a known matching rule, treat the whole object as matching rules
	for key := range obj {
		if validMatchingRules[key] {
			return true
		}
	}

	// Also check for patterns that look like matching rules (kebab-case or camelCase rule names)
	// This helps catch invalid matching rules
	for key := range obj {
		// If the key contains hyphens or looks like a rule name, treat as matching rules
		if strings.Contains(key, "-") && len(key) > 3 {
			// Likely a matching rule (valid or invalid)
			return true
		}
	}

	return false
}

// Helper functions

// isEmptyValue checks if a value is considered empty for EventBridge validation
func (v *EventBridgeValidator) isEmptyValue(value interface{}) bool {
	switch v := value.(type) {
	case string:
		return v == ""
	case map[string]interface{}:
		return len(v) == 0
	case []string:
		return len(v) == 0
	case time.Time:
		return v.IsZero()
	default:
		return value == nil
	}
}

// isNumeric checks if a string contains only digits
func (v *EventBridgeValidator) isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// isValidAWSRegion checks if a string is a valid AWS region format
func (v *EventBridgeValidator) isValidAWSRegion(region string) bool {
	// Basic validation for AWS region format (e.g., us-east-1, eu-west-2)
	if len(region) < 8 || len(region) > 15 {
		return false
	}

	// Should contain at least one hyphen
	if !strings.Contains(region, "-") {
		return false
	}

	// Common AWS region patterns
	validPrefixes := []string{
		"us-", "eu-", "ap-", "ca-", "sa-", "af-", "me-", "cn-", "us-gov-",
	}

	for _, prefix := range validPrefixes {
		if strings.HasPrefix(region, prefix) {
			return true
		}
	}

	return false
}

// CreateEventBridgeEventSchema creates a schema for validating EventBridge events
func CreateEventBridgeEventSchema() *MessageSchema {
	return &MessageSchema{
		Type: "object",
		Properties: map[string]*Property{
			"version": {
				Type:        "string",
				Description: "EventBridge event version",
			},
			"id": {
				Type:        "string",
				Description: "Unique event identifier",
			},
			"detail-type": {
				Type:        "string",
				Description: "Event detail type",
			},
			"source": {
				Type:        "string",
				Description: "Event source",
			},
			"account": {
				Type:        "string",
				Description: "AWS account ID",
			},
			"time": {
				Type:        "string",
				Format:      "date-time",
				Description: "Event timestamp",
			},
			"region": {
				Type:        "string",
				Description: "AWS region",
			},
			"detail": {
				Type:        "object",
				Description: "Event detail payload",
			},
			"resources": {
				Type: "array",
				Items: &Property{
					Type: "string",
				},
				Description: "Event resources",
			},
		},
		Required: []string{"version", "id", "detail-type", "source", "account", "time", "region", "detail"},
	}
}
