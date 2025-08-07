package generator

// Parser handles AsyncAPI specification parsing
type Parser interface {
	Parse(data []byte) (*ParseResult, error)
	ValidateVersion(version string) error
}

// CodeGenerator transforms schemas into Go code
type CodeGenerator interface {
	GenerateTypes(messages map[string]*MessageSchema, config *Config) (*GenerateResult, error)
	GenerateStruct(schema *MessageSchema, name string) (string, error)
}

// TypeMapper maps AsyncAPI types to Go types
type TypeMapper interface {
	MapType(schemaType string, format string) string
	MapProperty(prop *Property) *GoField
	MapPropertyWithContext(prop *Property, fieldName string, required bool) *GoField
}

// SchemaResolver handles external schema references
type SchemaResolver interface {
	ResolveRef(ref string) (*MessageSchema, error)
	ResolveProperty(ref string) (*Property, error)
}

// FileManager handles file I/O operations
type FileManager interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, content []byte) error
	CreateDir(path string) error
}

// Validator provides JSON schema validation capabilities for AsyncAPI schemas.
// It validates data against AsyncAPI/JSON Schema constraints including type validation,
// constraint validation (min/max length, numeric ranges, patterns), enum validation,
// and required field validation.
//
// Example usage:
//
//	validator := NewValidator(false) // permissive mode
//	result := validator.ValidateJSON(jsonData, schema)
//	if !result.Valid {
//	    for _, err := range result.Errors {
//	        fmt.Printf("Field '%s': %s\n", err.Field, err.Message)
//	    }
//	}
type Validator interface {
	// ValidateValue validates a Go value against a Property schema.
	// The value can be any Go type (string, int, map[string]any, []any, etc.).
	// The schema defines the expected structure and constraints.
	// The fieldPath is used for error reporting and should represent the path
	// to the field being validated (e.g., "user.profile.age").
	//
	// Returns a ValidationResult containing validation status and any errors.
	ValidateValue(value any, schema *Property, fieldPath string) *ValidationResult

	// ValidateJSON validates raw JSON data against a MessageSchema.
	// The JSON data is first parsed and then validated against the schema.
	// This method handles JSON parsing errors as well as schema validation errors.
	//
	// Returns a ValidationResult containing validation status and any errors.
	// If the JSON is malformed, the result will contain a parsing error.
	ValidateJSON(jsonData []byte, schema *MessageSchema) *ValidationResult

	// ValidateMessage validates a message payload against its MessageSchema.
	// The data should be a Go value (typically map[string]any from JSON unmarshaling)
	// that represents the message payload.
	//
	// Returns a ValidationResult containing validation status and any errors.
	ValidateMessage(data any, message *MessageSchema) *ValidationResult
}
