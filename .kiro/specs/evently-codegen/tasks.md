# Implementation Plan

- [x] 1. Set up project structure and core interfaces

  - Create Go module with proper directory structure (cmd/, pkg/, internal/)
  - Define core interfaces for Parser, CodeGenerator, and TypeMapper
  - Set up basic error types and configuration structs
  - _Requirements: 5.1, 5.2_

- [x] 2. Implement AsyncAPI data models and parsing foundation

  - [x] 2.1 Create AsyncAPI schema data structures

    - Implement MessageSchema, Property, and AsyncAPISpec structs
    - Add JSON/YAML struct tags for unmarshaling
    - Write unit tests for data model validation
    - _Requirements: 1.1, 1.4_

  - [x] 2.2 Implement basic AsyncAPI parser
    - Create parser that can read JSON and YAML AsyncAPI files
    - Add version detection and validation logic
    - Implement error handling for invalid syntax and unsupported versions
    - Write unit tests with valid and invalid AsyncAPI samples
    - _Requirements: 1.1, 1.2, 1.3_

- [x] 3. Implement schema resolution and reference handling

  - [x] 3.1 Create schema resolver for external references
    - Implement $ref resolution for local and external schema references
    - Add caching mechanism for resolved schemas
    - Handle circular reference detection and prevention
    - Write unit tests for various reference scenarios
    - _Requirements: 1.4_

- [x] 4. Implement Go type mapping and code generation

  - [x] 4.1 Create type mapper for AsyncAPI to Go type conversion

    - Implement mapping logic for primitive types (string, number, boolean, array, object)
    - Add support for format-specific mappings (date-time, email, etc.)
    - Handle optional vs required field type generation (pointers vs values)
    - Write unit tests for all type mapping scenarios
    - _Requirements: 2.1, 2.4, 4.2_

  - [x] 4.2 Implement Go struct code generator
    - Create template-based Go struct generation
    - Add PascalCase naming conversion for structs and fields
    - Generate appropriate JSON tags and field comments
    - Handle nested object and array type generation
    - Write unit tests that verify generated code compiles
    - _Requirements: 2.1, 2.2, 2.5, 4.1, 4.2, 4.4_

- [x] 5. Implement file I/O and output management

  - [x] 5.1 Create file I/O manager

    - Implement functions for reading AsyncAPI files from disk
    - Add directory creation and Go file writing capabilities
    - Handle file path validation and error reporting
    - Write unit tests for file operations
    - _Requirements: 3.1, 3.2_

  - [x] 5.2 Implement Go code formatting and package generation
    - Add Go code formatting using go/format package
    - Generate proper package declarations and imports
    - Ensure generated code follows Go conventions
    - Write integration tests that compile generated code
    - _Requirements: 4.3, 4.4, 4.5_

- [x] 6. Implement core library public API

  - [x] 6.1 Create main Generator struct and public methods

    - Implement Generator with Parse and Generate methods
    - Add configuration options through Config struct
    - Provide structured error handling and return types
    - Write unit tests for public API methods
    - _Requirements: 5.1, 5.2, 5.3, 5.4_

  - [x] 6.2 Add programmatic usage examples and documentation
    - Create example Go programs using the library
    - Add comprehensive package documentation
    - Write integration tests demonstrating library usage
    - _Requirements: 5.1, 5.2_

- [x] 7. Implement CLI interface and command handling

  - [x] 7.1 Create CLI argument parsing and validation

    - Implement command-line flag parsing for input file, output directory, and package name
    - Add help text and usage information display
    - Validate required parameters and provide clear error messages
    - Write unit tests for CLI argument handling
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

  - [x] 7.2 Implement CLI command execution and error handling
    - Wire CLI arguments to core library functionality
    - Add proper exit codes for different error conditions
    - Implement verbose output and progress reporting
    - Write integration tests for complete CLI workflows
    - _Requirements: 3.6_

- [x] 8. Add comprehensive testing and validation

  - [x] 8.1 Create end-to-end integration tests

    - Test complete workflow from AsyncAPI file to compiled Go code
    - Add tests with various AsyncAPI specification patterns
    - Verify generated code compiles and runs correctly
    - Test error scenarios and edge cases
    - _Requirements: All requirements validation_

  - [x] 8.2 Add performance and compatibility testing
    - Test with large AsyncAPI specifications
    - Verify memory usage and processing speed
    - Test compatibility with different AsyncAPI versions
    - Add benchmark tests for performance regression detection
    - _Requirements: 1.3, 2.1_

- [x] 9. Finalize project structure and documentation

  - [x] 9.1 Add project documentation and examples

    - Create comprehensive README with usage examples
    - Add API documentation and code comments
    - Create sample AsyncAPI files and generated output examples
    - _Requirements: 3.5, 5.1_

  - [x] 9.2 Set up build and release configuration
    - Add Makefile or build scripts for compilation
    - Configure Go modules and dependency management
    - Add version information and build metadata
    - Create release artifacts and installation instructions
    - _Requirements: 3.1, 3.6_

- [x] 10. Code quality improvements and optimizations

  - [x] 10.1 Address linting issues and code quality improvements

    - Replace interface{} with any for Go 1.18+ compatibility
    - Fix unused parameters and simplify loops using slices.Contains
    - Optimize string operations using fmt.Fprintf instead of WriteString(fmt.Sprintf)
    - Use strings.CutPrefix instead of HasPrefix + TrimPrefix combinations
    - _Requirements: 4.4, 4.5_

  - [x] 10.2 Enhance nested object handling

    - Improve nested struct generation to create separate type definitions
    - Add proper handling of deeply nested object structures
    - Ensure generated nested structs have proper naming and avoid conflicts
    - Write tests for complex nested object scenarios
    - _Requirements: 2.1, 2.5, 4.2_

  - [x] 10.3 Add missing CLI tests
    - Create comprehensive unit tests for CLI configuration and argument parsing
    - Add integration tests for CLI error scenarios and edge cases
    - Test CLI help output and usage information display
    - Verify CLI exit codes for different error conditions
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6_

- [x] 11. Implement JSON Schema validation capabilities

  - [x] 11.1 Create validation interface and core validation logic

    - Define Validator interface for JSON schema validation
    - Implement core validation engine that can validate JSON against AsyncAPI schemas
    - Add support for basic type validation (string, number, boolean, array, object)
    - Write unit tests for basic validation scenarios
    - _Requirements: 6.1, 6.2, 6.3_

  - [x] 11.2 Implement constraint validation

    - Add validation for string constraints (minLength, maxLength, pattern)
    - Add validation for numeric constraints (minimum, maximum, multipleOf)
    - Add validation for array constraints (minItems, maxItems, uniqueItems)
    - Add validation for enum value constraints
    - Write unit tests for all constraint validation scenarios
    - _Requirements: 6.4_

  - [x] 11.3 Add required field and additional properties validation

    - Implement validation for required fields in JSON objects
    - Add configurable handling of additional properties (strict/permissive modes)
    - Generate descriptive error messages for missing required fields
    - Write unit tests for required field validation scenarios
    - _Requirements: 6.5, 6.6_

  - [x] 11.4 Generate validation methods for Go structs

    - Modify code generator to include Validate() methods on generated structs
    - Add ValidateJSON() methods that accept raw JSON input
    - Ensure validation methods return structured error types with field paths
    - Write integration tests that validate generated validation methods work correctly
    - _Requirements: 6.1, 6.2, 6.3_

  - [x] 11.5 Add EventBridge-specific validation features
    - Implement validation for AWS EventBridge event structure requirements
    - Add support for validating event detail payload against AsyncAPI schemas
    - Create helper functions for validating EventBridge event patterns
    - Write tests with sample EventBridge events and AsyncAPI schemas
    - _Requirements: 6.1, 6.2, 6.3, 6.4_

- [x] 12. Implement Go enum type generation for AsyncAPI enum properties

  - [x] 12.1 Enhance type mapper to detect and handle enum properties

    - Modify MapProperty method to detect when a property has enum values
    - Create enum type generation logic for string and numeric enums
    - Generate custom type aliases with const declarations for enum values
    - Write unit tests for enum type mapping scenarios
    - _Requirements: 2.8, 2.9, 2.10_

  - [x] 12.2 Update code generator to produce enum type definitions

    - Modify GenerateTypes to create separate enum type definitions
    - Generate const blocks with proper Go naming conventions for enum values
    - Ensure enum types are used in struct field declarations instead of primitive types
    - Add validation methods for enum types to check valid values
    - Write integration tests that verify generated enum code compiles and works correctly
    - _Requirements: 2.8, 2.9, 2.10, 4.1, 4.2_

  - [x] 12.3 Add enum validation integration
    - Update validation methods to use generated enum types for validation
    - Ensure enum validation works with both struct validation and JSON validation
    - Add proper error messages for invalid enum values that reference the allowed values
    - Write tests that verify enum validation works correctly with generated types
    - _Requirements: 2.8, 2.9, 2.10, 6.4_

- [x] 13. Add comprehensive documentation for JSON validation functionality

  - [x] 13.1 Update README with validation documentation section

    - Add a new "JSON Validation" section to the main README
    - Document the validation capabilities including schema validation, constraint validation, and EventBridge validation
    - Include code examples showing how to use validation methods on generated structs
    - Add examples of ValidateJSON methods and error handling
    - _Requirements: 6.1, 6.2, 6.3, 6.4_

  - [x] 13.2 Create validation usage example

    - Create a new example in examples/validation_usage/ directory
    - Demonstrate validation of JSON data against AsyncAPI schemas
    - Show both successful validation and error scenarios
    - Include examples of enum validation and constraint validation
    - Show EventBridge-specific validation usage
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6_

  - [x] 13.3 Add validation API documentation

    - Document the Validator interface and SchemaValidator implementation
    - Add comprehensive godoc comments for all validation methods
    - Document ValidationResult and ValidationError types
    - Include usage examples in godoc comments
    - _Requirements: 6.1, 6.2, 6.3_

  - [x] 13.4 Update examples README with validation information

    - Add validation example to the examples overview
    - Include validation in the "Running the Examples" section
    - Add validation to the tips and customization guidance
    - _Requirements: 6.1, 6.2_
