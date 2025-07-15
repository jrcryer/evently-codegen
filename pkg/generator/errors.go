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

// ValidationError represents schema validation errors
type ValidationError struct {
	Field   string
	Message string
}

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
