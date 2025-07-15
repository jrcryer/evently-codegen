// Package generator provides functionality for parsing AsyncAPI specifications
// and generating corresponding Go type definitions.
//
// The generator package offers both programmatic API access and CLI functionality
// for converting AsyncAPI event schemas into strongly-typed Go structs. It supports
// AsyncAPI versions 2.x and 3.x and generates idiomatic Go code with proper
// struct tags, comments, and naming conventions.
//
// # Basic Usage
//
// The simplest way to use the generator is through the Generator struct:
//
//	config := &generator.Config{
//		PackageName:     "events",
//		OutputDir:       "./generated",
//		IncludeComments: true,
//		UsePointers:     true,
//	}
//
//	gen := generator.NewGenerator(config)
//
//	// Parse AsyncAPI specification
//	result, err := gen.Parse(asyncAPIData)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Generate Go code
//	generated, err := gen.Generate(result.Messages)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// # Configuration Options
//
// The Config struct provides several options for customizing code generation:
//
//   - PackageName: The Go package name for generated files
//   - OutputDir: Directory where generated files will be written
//   - IncludeComments: Whether to include schema descriptions as Go comments
//   - UsePointers: Whether to use pointer types for optional fields
//
// # File Operations
//
// The generator supports reading AsyncAPI specifications from files and writing
// generated Go code to files:
//
//	// Parse from file and generate to files in one step
//	err := gen.ParseFileAndGenerateToFiles("api-spec.json")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Or work with intermediate results
//	parseResult, err := gen.ParseFile("api-spec.json")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	err = gen.GenerateToFiles(parseResult.Messages)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// # Supported AsyncAPI Features
//
// The generator supports the following AsyncAPI/JSON Schema features:
//
//   - Basic types: string, number, integer, boolean, array, object
//   - String formats: date-time, date, time, email, uri, uuid, etc.
//   - Numeric formats: int32, int64, float, double
//   - Array types with typed elements
//   - Nested object types (generates nested structs)
//   - Required vs optional properties
//   - Schema descriptions and titles
//   - External schema references ($ref)
//
// # Type Mapping
//
// AsyncAPI/JSON Schema types are mapped to Go types as follows:
//
//   - string → string
//   - number → float64
//   - integer → int64
//   - boolean → bool
//   - array → []T (where T is the element type)
//   - object → struct or interface{}
//   - string with date-time format → time.Time
//   - string with email format → string
//   - integer with int32 format → int32
//   - number with float format → float32
//
// # Error Handling
//
// The generator provides structured error types for different failure scenarios:
//
//   - ParseError: Issues parsing AsyncAPI specifications
//   - ValidationError: Schema validation failures
//   - GenerationError: Code generation failures
//   - UnsupportedVersionError: Unsupported AsyncAPI versions
//   - FileError: File I/O operation failures
//   - ResolverError: Schema reference resolution failures
//   - CircularReferenceError: Circular schema reference detection
//
// # Examples
//
// See the examples directory for complete usage examples:
//
//   - examples/basic_usage: Basic programmatic usage
//   - examples/file_operations: File I/O operations and advanced features
//
// # Thread Safety
//
// Generator instances are not thread-safe. Create separate Generator instances
// for concurrent use, or use appropriate synchronization mechanisms.
//
// # Performance Considerations
//
// For large AsyncAPI specifications:
//
//   - The parser uses streaming for efficient memory usage
//   - Schema references are cached to avoid redundant resolution
//   - Generated code is formatted using go/format for optimal output
//
// # Validation
//
// The generator performs validation at multiple levels:
//
//   - AsyncAPI specification syntax and structure validation
//   - Schema reference resolution and circular dependency detection
//   - Generated Go code syntax validation and formatting
//   - Configuration parameter validation
package generator
