# Requirements Document

## Introduction

This feature involves creating a Go library with a CLI interface that can parse AsyncAPI specification files and generate corresponding Go type definitions. The tool will enable developers to automatically generate strongly-typed Go structs from AsyncAPI event schemas, reducing manual coding effort and ensuring consistency between API specifications and Go code.

## Requirements

### Requirement 1

**User Story:** As a Go developer, I want to parse AsyncAPI specification files, so that I can understand the event schemas defined in my asynchronous API documentation.

#### Acceptance Criteria

1. WHEN the CLI tool is provided with a valid AsyncAPI specification file THEN the system SHALL successfully parse the file and extract event schema definitions
2. WHEN the AsyncAPI file contains invalid JSON or YAML syntax THEN the system SHALL return a clear error message indicating the parsing failure
3. WHEN the AsyncAPI file uses an unsupported version THEN the system SHALL return an error message specifying supported versions
4. IF the AsyncAPI file contains references to external schemas THEN the system SHALL resolve and include those references in the parsing process

### Requirement 2

**User Story:** As a Go developer, I want to generate Go type definitions from AsyncAPI event schemas, so that I can use strongly-typed structs in my Go applications.

#### Acceptance Criteria

1. WHEN event schemas are successfully parsed THEN the system SHALL generate corresponding Go struct definitions with appropriate field types
2. WHEN schema properties have descriptions THEN the system SHALL include those descriptions as Go struct field comments
3. WHEN schema properties are marked as required THEN the system SHALL generate Go struct fields without pointer types
4. WHEN schema properties are optional THEN the system SHALL generate Go struct fields as pointer types or with appropriate zero values
5. WHEN nested objects are defined in schemas THEN the system SHALL generate nested Go struct definitions
6. WHEN array types are defined in schemas THEN the system SHALL generate Go slice types with appropriate element types
7. WHEN oneOf types are defined in schemas THEN the system SHALL generate Go structs that support polymorphism using interface types or union structs
8. WHEN schema properties are defined with enum property THEN the system SHALL generate Go type aliases with const declarations for each enum value, providing type safety and code completion
9. WHEN enum properties contain string values THEN the system SHALL generate a custom string type with const declarations for each allowed value
10. WHEN enum properties contain numeric values THEN the system SHALL generate a custom numeric type with const declarations for each allowed value
11. WHEN schema properties have validation constraints (min/max length, patterns) THEN the system SHALL include validation tags in the generated Go struct fields

### Requirement 3

**User Story:** As a Go developer, I want to use a command-line interface to control the code generation process, so that I can integrate the tool into my development workflow and build scripts.

#### Acceptance Criteria

1. WHEN the CLI tool is invoked with an input file parameter THEN the system SHALL read the specified AsyncAPI file
2. WHEN the CLI tool is invoked with an output directory parameter THEN the system SHALL write generated Go files to the specified directory
3. WHEN the CLI tool is invoked with a package name parameter THEN the system SHALL generate Go code with the specified package declaration
4. WHEN the CLI tool is invoked without required parameters THEN the system SHALL display usage instructions and exit with an error code
5. WHEN the CLI tool is invoked with a help flag THEN the system SHALL display detailed usage information and available options
6. WHEN the CLI tool encounters errors during execution THEN the system SHALL exit with appropriate error codes and descriptive error messages

### Requirement 4

**User Story:** As a Go developer, I want the generated Go types to follow Go naming conventions and best practices, so that the code integrates seamlessly with my existing Go codebase.

#### Acceptance Criteria

1. WHEN generating Go struct names THEN the system SHALL convert schema names to PascalCase following Go conventions
2. WHEN generating Go field names THEN the system SHALL convert property names to PascalCase and include appropriate JSON tags
3. WHEN generating Go package names THEN the system SHALL ensure package names are valid Go identifiers
4. WHEN generating Go code THEN the system SHALL include proper imports for required standard library packages
5. WHEN generating Go code THEN the system SHALL format the output using standard Go formatting conventions

### Requirement 5

**User Story:** As a Go developer, I want to use the tool as both a CLI application and a Go library, so that I can integrate the functionality directly into my Go applications when needed.

#### Acceptance Criteria

1. WHEN importing the library in Go code THEN the system SHALL provide public functions for parsing AsyncAPI specifications
2. WHEN using the library programmatically THEN the system SHALL provide public functions for generating Go type definitions
3. WHEN using the library programmatically THEN the system SHALL return structured error types that can be handled appropriately
4. WHEN using the library programmatically THEN the system SHALL allow configuration of generation options through function parameters or configuration structs

### Requirement 6

**User Story:** As a Go developer, I want the generated Go types to include JSON validation capabilities, so that I can validate incoming JSON data against the AsyncAPI specification at runtime.

#### Acceptance Criteria

1. WHEN generating Go struct types THEN the system SHALL include validation methods for each struct
2. WHEN a validation method is called with valid JSON data THEN the system SHALL return no errors
3. WHEN a validation method is called with invalid JSON data THEN the system SHALL return descriptive validation errors
4. WHEN schema properties have validation constraints (min/max length, patterns, enums) THEN the system SHALL enforce these constraints during validation
5. WHEN required fields are missing from JSON data THEN the system SHALL return validation errors indicating the missing fields
6. WHEN JSON data contains extra fields not defined in the schema THEN the system SHALL handle them according to configuration (ignore or error)
