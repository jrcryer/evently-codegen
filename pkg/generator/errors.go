package generator

import "fmt"

// ParseError represents AsyncAPI parsing errors
type ParseError struct {
	Message string
	Line    int
	Column  int
}

func (e *ParseError) Error() string {
	if e.Line > 0 && e.Column > 0 {
		return fmt.Sprintf("parse error at line %d, column %d: %s", e.Line, e.Column, e.Message)
	}
	return fmt.Sprintf("parse error: %s", e.Message)
}

// ValidationError represents a single schema validation error with field path and message.
// It provides detailed information about what validation rule was violated
// and where in the data structure the error occurred.
//
// ValidationError implements the error interface and can be used as a standard Go error.
//
// Example:
//
//	err := &ValidationError{
//	    Field:   "user.profile.age",
//	    Message: "value 200 exceeds maximum 150",
//	}
//	fmt.Println(err.Error()) // "validation error in field 'user.profile.age': value 200 exceeds maximum 150"
type ValidationError struct {
	// Field is the path to the field that failed validation (e.g., "user.profile.age").
	// An empty field indicates a root-level validation error.
	Field string

	// Message is a human-readable description of the validation error.
	Message string
}

// Error implements the error interface for ValidationError.
// It returns a formatted error message that includes the field path (if available) and the validation message.
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// GenerationError represents code generation errors
type GenerationError struct {
	Schema  string
	Message string
}

func (e *GenerationError) Error() string {
	if e.Schema != "" {
		return fmt.Sprintf("generation error in schema '%s': %s", e.Schema, e.Message)
	}
	return fmt.Sprintf("generation error: %s", e.Message)
}

// UnsupportedVersionError represents unsupported AsyncAPI version errors
type UnsupportedVersionError struct {
	Version           string
	SupportedVersions []string
}

func (e *UnsupportedVersionError) Error() string {
	return fmt.Sprintf("unsupported AsyncAPI version '%s', supported versions: %v", e.Version, e.SupportedVersions)
}

// ResolverError represents schema resolution errors
type ResolverError struct {
	Reference string
	Message   string
}

func (e *ResolverError) Error() string {
	return fmt.Sprintf("resolver error for reference '%s': %s", e.Reference, e.Message)
}

// CircularReferenceError represents circular reference errors
type CircularReferenceError struct {
	Reference string
	Stack     []string
}

func (e *CircularReferenceError) Error() string {
	return fmt.Sprintf("circular reference detected for '%s', resolution stack: %v", e.Reference, e.Stack)
}
