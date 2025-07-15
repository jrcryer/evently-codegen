package generator

import (
	"fmt"
	"path/filepath"
)

// Generator provides the main API for AsyncAPI to Go code generation
type Generator struct {
	config   *Config
	parser   Parser
	codegen  CodeGenerator
	resolver SchemaResolver
}

// NewGenerator creates a new Generator instance with the provided configuration
func NewGenerator(config *Config) *Generator {
	if config == nil {
		config = &Config{
			PackageName:     "main",
			OutputDir:       "./generated",
			IncludeComments: true,
			UsePointers:     true,
		}
	}

	return &Generator{
		config:   config,
		parser:   NewAsyncAPIParser(),
		codegen:  NewCodeGenerator(config),
		resolver: NewSchemaResolver(""),
	}
}

// Parse parses an AsyncAPI specification from raw data
func (g *Generator) Parse(data []byte) (*ParseResult, error) {
	if g.parser == nil {
		return nil, &GenerationError{Message: "parser not initialized"}
	}
	return g.parser.Parse(data)
}

// Generate generates Go code from parsed AsyncAPI messages
func (g *Generator) Generate(messages map[string]*MessageSchema) (*GenerateResult, error) {
	if g.codegen == nil {
		return nil, &GenerationError{Message: "code generator not initialized"}
	}
	return g.codegen.GenerateTypes(messages, g.config)
}

// ParseAndGenerate is a convenience method that parses and generates in one call
func (g *Generator) ParseAndGenerate(data []byte) (*GenerateResult, error) {
	parseResult, err := g.Parse(data)
	if err != nil {
		return nil, err
	}

	if len(parseResult.Errors) > 0 {
		return &GenerateResult{Errors: parseResult.Errors}, nil
	}

	return g.Generate(parseResult.Messages)
}

// GetConfig returns the current configuration
func (g *Generator) GetConfig() *Config {
	return g.config
}

// SetConfig updates the generator configuration
func (g *Generator) SetConfig(config *Config) {
	if config != nil {
		g.config = config
		// Reinitialize components with new config
		g.codegen = NewCodeGenerator(config)
	}
}

// ParseFile parses an AsyncAPI specification from a file path
func (g *Generator) ParseFile(filePath string) (*ParseResult, error) {
	fileManager := NewFileManager()
	data, err := fileManager.ReadFile(filePath)
	if err != nil {
		return nil, &GenerationError{
			Schema:  filePath,
			Message: fmt.Sprintf("failed to read file: %v", err),
		}
	}

	// Set base URI for resolver based on file path
	if g.resolver != nil {
		if resolver, ok := g.resolver.(*DefaultSchemaResolver); ok {
			resolver.SetBaseURI(filePath)
		}
	}

	return g.Parse(data)
}

// GenerateToFiles generates Go code and writes it to files in the output directory
func (g *Generator) GenerateToFiles(messages map[string]*MessageSchema) error {
	result, err := g.Generate(messages)
	if err != nil {
		return err
	}

	if len(result.Errors) > 0 {
		return &GenerationError{
			Message: fmt.Sprintf("generation completed with %d errors", len(result.Errors)),
		}
	}

	fileManager := NewFileManager()

	// Create output directory if it doesn't exist
	if err := fileManager.CreateDir(g.config.OutputDir); err != nil {
		return &GenerationError{
			Message: fmt.Sprintf("failed to create output directory: %v", err),
		}
	}

	// Write each generated file
	for filename, content := range result.Files {
		filePath := filepath.Join(g.config.OutputDir, filename)
		if err := fileManager.WriteFile(filePath, []byte(content)); err != nil {
			return &GenerationError{
				Schema:  filename,
				Message: fmt.Sprintf("failed to write file: %v", err),
			}
		}
	}

	return nil
}

// ParseAndGenerateToFiles is a convenience method that parses and generates files in one call
func (g *Generator) ParseAndGenerateToFiles(data []byte) error {
	parseResult, err := g.Parse(data)
	if err != nil {
		return err
	}

	if len(parseResult.Errors) > 0 {
		return &GenerationError{
			Message: fmt.Sprintf("parsing completed with %d errors", len(parseResult.Errors)),
		}
	}

	return g.GenerateToFiles(parseResult.Messages)
}

// ParseFileAndGenerateToFiles parses a file and generates output files
func (g *Generator) ParseFileAndGenerateToFiles(filePath string) error {
	parseResult, err := g.ParseFile(filePath)
	if err != nil {
		return err
	}

	if len(parseResult.Errors) > 0 {
		return &GenerationError{
			Message: fmt.Sprintf("parsing completed with %d errors", len(parseResult.Errors)),
		}
	}

	return g.GenerateToFiles(parseResult.Messages)
}

// ValidateConfig validates the generator configuration
func (g *Generator) ValidateConfig() error {
	if g.config == nil {
		return &ValidationError{
			Field:   "config",
			Message: "configuration is required",
		}
	}

	if g.config.PackageName == "" {
		return &ValidationError{
			Field:   "config.PackageName",
			Message: "package name is required",
		}
	}

	if g.config.OutputDir == "" {
		return &ValidationError{
			Field:   "config.OutputDir",
			Message: "output directory is required",
		}
	}

	// Validate package name is a valid Go identifier
	if !IsValidGoIdentifier(g.config.PackageName) {
		return &ValidationError{
			Field:   "config.PackageName",
			Message: fmt.Sprintf("'%s' is not a valid Go package name", g.config.PackageName),
		}
	}

	return nil
}

// GetSupportedVersions returns the AsyncAPI versions supported by the parser
func (g *Generator) GetSupportedVersions() []string {
	if parser, ok := g.parser.(*AsyncAPIParser); ok {
		return parser.GetSupportedVersions()
	}
	return []string{}
}
