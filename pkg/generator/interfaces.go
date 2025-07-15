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
