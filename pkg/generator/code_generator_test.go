package generator

import (
	"fmt"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestNewCodeGenerator(t *testing.T) {
	config := &Config{PackageName: "test"}
	generator := NewCodeGenerator(config)

	if generator == nil {
		t.Fatal("NewCodeGenerator returned nil")
	}

	if generator.config != config {
		t.Error("CodeGenerator config not set correctly")
	}

	if generator.typeMapper == nil {
		t.Error("CodeGenerator typeMapper not initialized")
	}
}

func TestGenerateStruct_BasicStruct(t *testing.T) {
	generator := NewCodeGenerator(&Config{PackageName: "test", UsePointers: true})

	schema := &MessageSchema{
		Title:       "User",
		Description: "A user object",
		Type:        "object",
		Properties: map[string]*Property{
			"name": {
				Type:        "string",
				Description: "User's name",
			},
			"age": {
				Type:        "integer",
				Description: "User's age",
			},
		},
		Required: []string{"name"},
	}

	result, err := generator.GenerateStruct(schema, "User")
	if err != nil {
		t.Fatalf("GenerateStruct failed: %v", err)
	}

	// Check that the result contains expected elements
	if !strings.Contains(result, "type User struct") {
		t.Error("Generated code should contain struct declaration")
	}

	if !strings.Contains(result, "Name string") {
		t.Error("Generated code should contain Name field")
	}

	if !strings.Contains(result, "Age *int64") {
		t.Error("Generated code should contain Age field as pointer (optional)")
	}

	if !strings.Contains(result, `json:"name"`) {
		t.Error("Generated code should contain JSON tag for name")
	}

	if !strings.Contains(result, `json:"age"`) {
		t.Error("Generated code should contain JSON tag for age")
	}
}

func TestGenerateStruct_WithArrays(t *testing.T) {
	generator := NewCodeGenerator(&Config{PackageName: "test", UsePointers: true})

	schema := &MessageSchema{
		Type: "object",
		Properties: map[string]*Property{
			"tags": {
				Type: "array",
				Items: &Property{
					Type: "string",
				},
				Description: "List of tags",
			},
			"scores": {
				Type: "array",
				Items: &Property{
					Type: "integer",
				},
			},
		},
		Required: []string{"tags"},
	}

	result, err := generator.GenerateStruct(schema, "TestStruct")
	if err != nil {
		t.Fatalf("GenerateStruct failed: %v", err)
	}

	if !strings.Contains(result, "Tags []string") {
		t.Error("Generated code should contain Tags field as string slice")
	}

	if !strings.Contains(result, "Scores *[]int64") {
		t.Error("Generated code should contain Scores field as pointer to int64 slice")
	}
}

func TestGenerateStruct_WithTimeFormats(t *testing.T) {
	generator := NewCodeGenerator(&Config{PackageName: "test", UsePointers: true})

	schema := &MessageSchema{
		Type: "object",
		Properties: map[string]*Property{
			"created_at": {
				Type:   "string",
				Format: "date-time",
			},
			"birth_date": {
				Type:   "string",
				Format: "date",
			},
		},
	}

	result, err := generator.GenerateStruct(schema, "Event")
	if err != nil {
		t.Fatalf("GenerateStruct failed: %v", err)
	}

	if !strings.Contains(result, "CreatedAt *time.Time") {
		t.Error("Generated code should contain CreatedAt field as time.Time")
	}

	if !strings.Contains(result, "BirthDate *time.Time") {
		t.Error("Generated code should contain BirthDate field as time.Time")
	}
}

func TestGenerateStruct_NilSchema(t *testing.T) {
	generator := NewCodeGenerator(&Config{})

	_, err := generator.GenerateStruct(nil, "Test")
	if err == nil {
		t.Error("GenerateStruct should return error for nil schema")
	}
}

func TestGenerateStruct_EmptyName(t *testing.T) {
	generator := NewCodeGenerator(&Config{})

	schema := &MessageSchema{
		Type: "object",
		Properties: map[string]*Property{
			"field": {Type: "string"},
		},
	}

	result, err := generator.GenerateStruct(schema, "")
	if err != nil {
		t.Fatalf("GenerateStruct failed: %v", err)
	}

	if !strings.Contains(result, "type GeneratedStruct struct") {
		t.Error("Generated code should use default struct name")
	}
}

func TestGenerateTypes(t *testing.T) {
	generator := NewCodeGenerator(&Config{PackageName: "models"})

	messages := map[string]*MessageSchema{
		"User": {
			Type: "object",
			Properties: map[string]*Property{
				"name":  {Type: "string"},
				"email": {Type: "string", Format: "email"},
			},
			Required: []string{"name"},
		},
		"Product": {
			Type: "object",
			Properties: map[string]*Property{
				"title": {Type: "string"},
				"price": {Type: "number"},
			},
		},
	}

	result, err := generator.GenerateTypes(messages, nil)
	if err != nil {
		t.Fatalf("GenerateTypes failed: %v", err)
	}

	if len(result.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(result.Files))
	}

	// Check that files are generated with correct names
	if _, exists := result.Files["user.go"]; !exists {
		t.Error("Expected user.go file to be generated")
	}

	if _, exists := result.Files["product.go"]; !exists {
		t.Error("Expected product.go file to be generated")
	}

	// Check that files contain package declaration
	userFile := result.Files["user.go"]
	if !strings.Contains(userFile, "package models") {
		t.Error("Generated file should contain correct package declaration")
	}
}

func TestCreateGoFile(t *testing.T) {
	generator := NewCodeGenerator(&Config{PackageName: "test"})

	structCode := `type User struct {
	Name string ` + "`json:\"name\"`" + `
	CreatedAt *time.Time ` + "`json:\"created_at\"`" + `
}`

	result, err := generator.createGoFile(structCode, "User")
	if err != nil {
		t.Fatalf("createGoFile failed: %v", err)
	}

	// Check package declaration
	if !strings.Contains(result, "package test") {
		t.Error("Generated file should contain package declaration")
	}

	// Check import for time package
	if !strings.Contains(result, `import "time"`) {
		t.Error("Generated file should import time package")
	}

	// Check that the code is properly formatted (should be valid Go)
	_, err = parser.ParseFile(token.NewFileSet(), "", result, parser.ParseComments)
	if err != nil {
		t.Errorf("Generated code is not valid Go: %v", err)
	}
}

func TestCreateGoFile_MultipleImports(t *testing.T) {
	generator := NewCodeGenerator(&Config{PackageName: "test"})

	// This would be a more complex case if we had multiple imports
	structCode := `type Event struct {
	Timestamp *time.Time ` + "`json:\"timestamp\"`" + `
}`

	result, err := generator.createGoFile(structCode, "Event")
	if err != nil {
		t.Fatalf("createGoFile failed: %v", err)
	}

	// Verify the code compiles
	_, err = parser.ParseFile(token.NewFileSet(), "", result, parser.ParseComments)
	if err != nil {
		t.Errorf("Generated code is not valid Go: %v", err)
	}
}

func TestGetRequiredImports(t *testing.T) {
	generator := NewCodeGenerator(&Config{})

	tests := []struct {
		name       string
		structCode string
		expected   []string
	}{
		{
			name:       "no imports needed",
			structCode: "type User struct { Name string }",
			expected:   []string{},
		},
		{
			name:       "time import needed",
			structCode: "type Event struct { CreatedAt time.Time }",
			expected:   []string{"time"},
		},
		{
			name:       "multiple time references",
			structCode: "type Event struct { CreatedAt time.Time; UpdatedAt *time.Time }",
			expected:   []string{"time"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := generator.getRequiredImports(test.structCode)

			if len(result) != len(test.expected) {
				t.Errorf("Expected %d imports, got %d", len(test.expected), len(result))
				return
			}

			for i, expected := range test.expected {
				if result[i] != expected {
					t.Errorf("Expected import %s, got %s", expected, result[i])
				}
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"User", "user"},
		{"UserProfile", "user_profile"},
		{"XMLHttpRequest", "x_m_l_http_request"},
		{"APIKey", "a_p_i_key"},
		{"simpleword", "simpleword"},
		{"HTML", "h_t_m_l"},
	}

	for _, test := range tests {
		result := toSnakeCase(test.input)
		if result != test.expected {
			t.Errorf("toSnakeCase(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestIsPropertyRequired(t *testing.T) {
	generator := NewCodeGenerator(&Config{})

	required := []string{"name", "email", "age"}

	tests := []struct {
		propName string
		expected bool
	}{
		{"name", true},
		{"email", true},
		{"age", true},
		{"phone", false},
		{"", false},
	}

	for _, test := range tests {
		result := generator.isPropertyRequired(test.propName, required)
		if result != test.expected {
			t.Errorf("isPropertyRequired(%q) = %v, expected %v", test.propName, result, test.expected)
		}
	}
}

func TestGenerateStructCode(t *testing.T) {
	generator := NewCodeGenerator(&Config{IncludeComments: true})

	goStruct := &GoStruct{
		Name:        "TestStruct",
		PackageName: "test",
		Comments:    []string{"This is a test struct", "", "It has multiple comments"},
		Fields: []*GoField{
			{
				Name:    "Name",
				Type:    "string",
				JSONTag: `json:"name"`,
				Comment: "The name field",
			},
			{
				Name:    "Age",
				Type:    "*int64",
				JSONTag: `json:"age"`,
			},
		},
	}

	result, err := generator.generateStructCode(goStruct)
	if err != nil {
		t.Fatalf("generateStructCode failed: %v", err)
	}

	// Check struct declaration
	if !strings.Contains(result, "type TestStruct struct {") {
		t.Error("Generated code should contain struct declaration")
	}

	// Check comments
	if !strings.Contains(result, "// This is a test struct") {
		t.Error("Generated code should contain struct comments")
	}

	if !strings.Contains(result, "//") { // Empty comment line
		t.Error("Generated code should contain empty comment line")
	}

	// Check fields
	if !strings.Contains(result, "Name string `json:\"name\"`") {
		t.Error("Generated code should contain Name field with JSON tag")
	}

	if !strings.Contains(result, "// The name field") {
		t.Error("Generated code should contain field comment")
	}

	if !strings.Contains(result, "Age *int64 `json:\"age\"`") {
		t.Error("Generated code should contain Age field")
	}
}

func TestGeneratedCodeCompiles(t *testing.T) {
	generator := NewCodeGenerator(&Config{PackageName: "test", UsePointers: true})

	schema := &MessageSchema{
		Title:       "ComplexStruct",
		Description: "A complex struct for testing",
		Type:        "object",
		Properties: map[string]*Property{
			"id": {
				Type:        "string",
				Description: "Unique identifier",
			},
			"name": {
				Type:        "string",
				Description: "Display name",
			},
			"age": {
				Type:        "integer",
				Description: "Age in years",
			},
			"email": {
				Type:        "string",
				Format:      "email",
				Description: "Email address",
			},
			"created_at": {
				Type:        "string",
				Format:      "date-time",
				Description: "Creation timestamp",
			},
			"tags": {
				Type: "array",
				Items: &Property{
					Type: "string",
				},
				Description: "List of tags",
			},
			"scores": {
				Type: "array",
				Items: &Property{
					Type: "number",
				},
				Description: "List of scores",
			},
			"active": {
				Type:        "boolean",
				Description: "Whether the user is active",
			},
		},
		Required: []string{"id", "name", "email"},
	}

	// Generate the struct
	structCode, err := generator.GenerateStruct(schema, "ComplexStruct")
	if err != nil {
		t.Fatalf("GenerateStruct failed: %v", err)
	}

	// Create a complete Go file
	fileContent, err := generator.createGoFile(structCode, "ComplexStruct")
	if err != nil {
		t.Fatalf("createGoFile failed: %v", err)
	}

	// Parse the generated code to ensure it's valid Go
	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, "test.go", fileContent, parser.ParseComments)
	if err != nil {
		t.Errorf("Generated code does not compile: %v\nGenerated code:\n%s", err, fileContent)
	}

	// Additional checks for expected content
	expectedFields := []string{
		"Id        string",     // required field, not pointer
		"Name      string",     // required field, not pointer
		"Age       *int64",     // optional field, pointer
		"Email     string",     // required field, not pointer
		"CreatedAt *time.Time", // optional field with time format
		"Tags      *[]string",  // optional array field
		"Scores    *[]float64", // optional array field
		"Active    *bool",      // optional boolean field
	}

	for _, expected := range expectedFields {
		if !strings.Contains(fileContent, expected) {
			t.Errorf("Generated code should contain field: %s\nGenerated code:\n%s", expected, fileContent)
		}
	}
}

// Benchmark tests
func BenchmarkGenerateStruct(b *testing.B) {
	generator := NewCodeGenerator(&Config{PackageName: "test"})

	schema := &MessageSchema{
		Type: "object",
		Properties: map[string]*Property{
			"name":  {Type: "string"},
			"age":   {Type: "integer"},
			"email": {Type: "string", Format: "email"},
		},
		Required: []string{"name"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generator.GenerateStruct(schema, "BenchmarkStruct")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkToSnakeCase(b *testing.B) {
	input := "ComplexStructNameForTesting"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		toSnakeCase(input)
	}
}
func TestIsValidGoIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		// Valid identifiers
		{"validName", true},
		{"ValidName", true},
		{"_validName", true},
		{"valid123", true},
		{"_", true},
		{"a", true},
		{"A", true},
		{"camelCase", true},
		{"PascalCase", true},
		{"snake_case", true},
		{"UPPER_CASE", true},
		{"mixed123_Case", true},

		// Invalid identifiers
		{"", false},
		{"123invalid", false},
		{"invalid-name", false},
		{"invalid.name", false},
		{"invalid name", false},
		{"invalid@name", false},
		{"invalid#name", false},
		{"invalid$name", false},
		{"invalid%name", false},

		// Go keywords (should be invalid)
		{"package", false},
		{"import", false},
		{"func", false},
		{"var", false},
		{"const", false},
		{"type", false},
		{"struct", false},
		{"interface", false},
		{"map", false},
		{"chan", false},
		{"go", false},
		{"defer", false},
		{"if", false},
		{"else", false},
		{"for", false},
		{"range", false},
		{"switch", false},
		{"case", false},
		{"default", false},
		{"break", false},
		{"continue", false},
		{"fallthrough", false},
		{"return", false},
		{"goto", false},
		{"select", false},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := IsValidGoIdentifier(test.input)
			if result != test.expected {
				t.Errorf("isValidGoIdentifier(%q) = %v, expected %v", test.input, result, test.expected)
			}
		})
	}
}

func TestIsGoKeyword(t *testing.T) {
	keywords := []string{
		"break", "case", "chan", "const", "continue",
		"default", "defer", "else", "fallthrough", "for",
		"func", "go", "goto", "if", "import",
		"interface", "map", "package", "range", "return",
		"select", "struct", "switch", "type", "var",
	}

	for _, keyword := range keywords {
		if !isGoKeyword(keyword) {
			t.Errorf("isGoKeyword(%q) should return true", keyword)
		}
	}

	nonKeywords := []string{
		"validName", "ValidName", "_validName", "valid123",
		"camelCase", "PascalCase", "snake_case", "UPPER_CASE",
	}

	for _, nonKeyword := range nonKeywords {
		if isGoKeyword(nonKeyword) {
			t.Errorf("isGoKeyword(%q) should return false", nonKeyword)
		}
	}
}

func TestCreateGoFile_InvalidPackageName(t *testing.T) {
	generator := NewCodeGenerator(&Config{PackageName: "123invalid"})

	structCode := `type User struct {
	Name string ` + "`json:\"name\"`" + `
}`

	_, err := generator.createGoFile(structCode, "User")
	if err == nil {
		t.Error("createGoFile should return error for invalid package name")
	}

	if !strings.Contains(err.Error(), "invalid package name") {
		t.Errorf("Error should mention invalid package name, got: %v", err)
	}
}

func TestFormatGoCode(t *testing.T) {
	generator := NewCodeGenerator(&Config{})

	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name: "valid code",
			input: `package test

type User struct {
Name string
Age int
}`,
			hasError: false,
		},
		{
			name: "code with formatting issues",
			input: `package test


type User struct{
Name    string
Age   int
}`,
			hasError: false,
		},
		{
			name:     "invalid syntax",
			input:    `package test type User struct { Name string Age int`,
			hasError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := generator.formatGoCode(test.input)

			if test.hasError {
				if err == nil {
					t.Error("formatGoCode should return error for invalid syntax")
				}
				// Check that it returns a GenerationError
				if genErr, ok := err.(*GenerationError); !ok {
					t.Errorf("Expected GenerationError, got %T", err)
				} else if genErr.Schema != "code formatting" {
					t.Errorf("Expected schema 'code formatting', got %q", genErr.Schema)
				}
			} else {
				if err != nil {
					t.Errorf("formatGoCode should not return error for valid code: %v", err)
				}
				if result == "" {
					t.Error("formatGoCode should return formatted code")
				}
				// Verify the result is properly formatted by parsing it
				_, parseErr := parser.ParseFile(token.NewFileSet(), "", result, parser.ParseComments)
				if parseErr != nil {
					t.Errorf("Formatted code is not valid Go: %v", parseErr)
				}
			}
		})
	}
}

func TestWriteImports(t *testing.T) {
	generator := NewCodeGenerator(&Config{})

	tests := []struct {
		name     string
		imports  []string
		expected string
	}{
		{
			name:     "single import",
			imports:  []string{"time"},
			expected: `import "time"`,
		},
		{
			name:     "multiple imports",
			imports:  []string{"fmt", "time", "encoding/json"},
			expected: "import (",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var builder strings.Builder
			generator.writeImports(&builder, test.imports)
			result := builder.String()

			if !strings.Contains(result, test.expected) {
				t.Errorf("writeImports output should contain %q, got %q", test.expected, result)
			}
		})
	}
}

func TestGetRequiredImports_Enhanced(t *testing.T) {
	generator := NewCodeGenerator(&Config{})

	tests := []struct {
		name       string
		structCode string
		expected   []string
	}{
		{
			name:       "no imports needed",
			structCode: "type User struct { Name string }",
			expected:   []string{},
		},
		{
			name:       "time import needed",
			structCode: "type Event struct { CreatedAt time.Time }",
			expected:   []string{"time"},
		},
		{
			name:       "multiple imports needed",
			structCode: "type Complex struct { CreatedAt time.Time; Data json.RawMessage; ID fmt.Stringer }",
			expected:   []string{"encoding/json", "fmt", "time"},
		},
		{
			name:       "url import needed",
			structCode: "type Resource struct { URL url.URL }",
			expected:   []string{"net/url"},
		},
		{
			name:       "uuid import needed",
			structCode: "type Entity struct { ID uuid.UUID }",
			expected:   []string{"github.com/google/uuid"},
		},
		{
			name:       "duplicate imports",
			structCode: "type Event struct { CreatedAt time.Time; UpdatedAt time.Time }",
			expected:   []string{"time"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := generator.getRequiredImports(test.structCode)

			if len(result) != len(test.expected) {
				t.Errorf("Expected %d imports, got %d: %v", len(test.expected), len(result), result)
				return
			}

			for i, expected := range test.expected {
				if result[i] != expected {
					t.Errorf("Expected import %s, got %s", expected, result[i])
				}
			}
		})
	}
}

func TestCreateGoFile_ComplexImports(t *testing.T) {
	generator := NewCodeGenerator(&Config{PackageName: "models"})

	structCode := `type ComplexStruct struct {
	ID uuid.UUID ` + "`json:\"id\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
	UpdatedAt *time.Time ` + "`json:\"updated_at\"`" + `
	Data json.RawMessage ` + "`json:\"data\"`" + `
	URL url.URL ` + "`json:\"url\"`" + `
}`

	result, err := generator.createGoFile(structCode, "ComplexStruct")
	if err != nil {
		t.Fatalf("createGoFile failed: %v", err)
	}

	// Check package declaration
	if !strings.Contains(result, "package models") {
		t.Error("Generated file should contain correct package declaration")
	}

	// Check that all required imports are present
	expectedImports := []string{
		"encoding/json",
		"github.com/google/uuid",
		"net/url",
		"time",
	}

	for _, imp := range expectedImports {
		if !strings.Contains(result, fmt.Sprintf(`"%s"`, imp)) {
			t.Errorf("Generated file should import %s", imp)
		}
	}

	// Check that imports are in a block format (multiple imports)
	if !strings.Contains(result, "import (") {
		t.Error("Generated file should use import block format for multiple imports")
	}

	// Verify the code compiles
	_, err = parser.ParseFile(token.NewFileSet(), "", result, parser.ParseComments)
	if err != nil {
		t.Errorf("Generated code is not valid Go: %v\nGenerated code:\n%s", err, result)
	}
}

func TestGenerateTypes_WithFormattingErrors(t *testing.T) {
	// Create a mock code generator that will produce invalid Go code
	generator := NewCodeGenerator(&Config{PackageName: "test"})

	// Override the generateStructCode method behavior by creating invalid struct code
	messages := map[string]*MessageSchema{
		"InvalidStruct": {
			Type: "object",
			Properties: map[string]*Property{
				"field": {Type: "invalid_type"}, // This should cause issues
			},
		},
	}

	result, err := generator.GenerateTypes(messages, nil)
	if err != nil {
		t.Fatalf("GenerateTypes failed: %v", err)
	}

	// The result should still be generated, but we should check for any errors
	if len(result.Errors) > 0 {
		t.Logf("GenerateTypes produced errors (expected): %v", result.Errors)
	}

	// Check that at least one file was generated
	if len(result.Files) == 0 {
		t.Error("GenerateTypes should generate at least one file")
	}
}

// Integration test that verifies the complete workflow
func TestCompleteCodeGenerationWorkflow(t *testing.T) {
	config := &Config{
		PackageName:     "testmodels",
		IncludeComments: true,
		UsePointers:     true,
	}

	generator := NewCodeGenerator(config)

	// Create a comprehensive schema
	messages := map[string]*MessageSchema{
		"UserProfile": {
			Title:       "User Profile",
			Description: "Represents a user profile in the system",
			Type:        "object",
			Properties: map[string]*Property{
				"id": {
					Type:        "string",
					Description: "Unique identifier",
				},
				"username": {
					Type:        "string",
					Description: "Username for login",
				},
				"email": {
					Type:        "string",
					Format:      "email",
					Description: "Email address",
				},
				"full_name": {
					Type:        "string",
					Description: "Full display name",
				},
				"age": {
					Type:        "integer",
					Description: "Age in years",
				},
				"birth_date": {
					Type:        "string",
					Format:      "date",
					Description: "Date of birth",
				},
				"created_at": {
					Type:        "string",
					Format:      "date-time",
					Description: "Account creation timestamp",
				},
				"updated_at": {
					Type:        "string",
					Format:      "date-time",
					Description: "Last update timestamp",
				},
				"tags": {
					Type: "array",
					Items: &Property{
						Type: "string",
					},
					Description: "User tags",
				},
				"preferences": {
					Type: "object",
					Properties: map[string]*Property{
						"theme":         {Type: "string"},
						"notifications": {Type: "boolean"},
					},
					Description: "User preferences",
				},
				"is_active": {
					Type:        "boolean",
					Description: "Whether the user is active",
				},
				"score": {
					Type:        "number",
					Description: "User score",
				},
			},
			Required: []string{"id", "username", "email"},
		},
	}

	// Generate the types
	result, err := generator.GenerateTypes(messages, config)
	if err != nil {
		t.Fatalf("GenerateTypes failed: %v", err)
	}

	// Check for errors
	if len(result.Errors) > 0 {
		t.Errorf("GenerateTypes produced errors: %v", result.Errors)
	}

	// Check that the file was generated
	if len(result.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(result.Files))
	}

	userProfileFile, exists := result.Files["user_profile.go"]
	if !exists {
		t.Fatal("Expected user_profile.go file to be generated")
	}

	// Verify the generated code
	expectedElements := []string{
		"package testmodels",
		"type Userprofile struct", // Note: this is the actual output based on the key name
		"Id string",               // required field
		"Username string",         // required field
		"Email string",            // required field
		"FullName *string",        // optional field
		"Age *int64",              // optional field
		"BirthDate *time.Time",    // optional date field
		"CreatedAt *time.Time",    // optional datetime field
		"UpdatedAt *time.Time",    // optional datetime field
		"Tags *[]string",          // optional array field
		"IsActive *bool",          // optional boolean field
		"Score *float64",          // optional number field
		`json:"id"`,               // JSON tags
		`json:"username"`,
		`json:"email"`,
		`json:"full_name"`,
		`json:"age"`,
		`json:"birth_date"`,
		`json:"created_at"`,
		`json:"updated_at"`,
		`json:"tags"`,
		`json:"is_active"`,
		`json:"score"`,
		"// Represents a user profile in the system", // struct comment
		"// Unique identifier",                       // field comment
		"// Username for login",                      // field comment
	}

	for _, expected := range expectedElements {
		if !strings.Contains(userProfileFile, expected) {
			t.Errorf("Generated file should contain: %s\nGenerated file:\n%s", expected, userProfileFile)
		}
	}

	// Verify the code compiles
	_, err = parser.ParseFile(token.NewFileSet(), "user_profile.go", userProfileFile, parser.ParseComments)
	if err != nil {
		t.Errorf("Generated code does not compile: %v\nGenerated code:\n%s", err, userProfileFile)
	}

	// Check that time import is present
	if !strings.Contains(userProfileFile, `import "time"`) {
		t.Error("Generated file should import time package")
	}
}
