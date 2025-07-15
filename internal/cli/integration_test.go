package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_Run_Success(t *testing.T) {
	// Create a temporary AsyncAPI file
	tmpDir, err := os.MkdirTemp("", "cli_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a simple AsyncAPI spec
	asyncAPIContent := `{
  "asyncapi": "2.6.0",
  "info": {
    "title": "Test API",
    "version": "1.0.0"
  },
  "channels": {
    "user/signup": {
      "publish": {
        "message": {
          "payload": {
            "type": "object",
            "properties": {
              "userId": {
                "type": "string"
              },
              "email": {
                "type": "string",
                "format": "email"
              }
            },
            "required": ["userId", "email"]
          }
        }
      }
    }
  }
}`

	inputFile := filepath.Join(tmpDir, "asyncapi.json")
	if err := os.WriteFile(inputFile, []byte(asyncAPIContent), 0644); err != nil {
		t.Fatalf("Failed to write AsyncAPI file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "output")

	// Create CLI and parse arguments
	cli := NewCLI()
	args := []string{
		"-i", inputFile,
		"-o", outputDir,
		"-p", "testpkg",
	}

	err = cli.ParseArgs(args)
	if err != nil {
		t.Fatalf("ParseArgs failed: %v", err)
	}

	// Run CLI command
	exitCode := cli.Run()
	if exitCode != ExitCodeSuccess {
		t.Errorf("Expected exit code %d, got %d", ExitCodeSuccess, exitCode)
	}

	// Verify output directory was created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Errorf("Output directory was not created: %s", outputDir)
	}
}

func TestCLI_Run_VerboseMode(t *testing.T) {
	// Create a temporary AsyncAPI file
	tmpDir, err := os.MkdirTemp("", "cli_test_verbose_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a simple AsyncAPI spec
	asyncAPIContent := `{
  "asyncapi": "2.6.0",
  "info": {
    "title": "Test API",
    "version": "1.0.0"
  },
  "channels": {}
}`

	inputFile := filepath.Join(tmpDir, "asyncapi.json")
	if err := os.WriteFile(inputFile, []byte(asyncAPIContent), 0644); err != nil {
		t.Fatalf("Failed to write AsyncAPI file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "output")

	// Create CLI and parse arguments with verbose flag
	cli := NewCLI()
	args := []string{
		"-i", inputFile,
		"-o", outputDir,
		"-p", "testpkg",
		"-v",
	}

	err = cli.ParseArgs(args)
	if err != nil {
		t.Fatalf("ParseArgs failed: %v", err)
	}

	// Verify verbose mode is enabled
	if !cli.GetConfig().Verbose {
		t.Error("Expected verbose mode to be enabled")
	}

	// Run CLI command
	exitCode := cli.Run()
	if exitCode != ExitCodeSuccess {
		t.Errorf("Expected exit code %d, got %d", ExitCodeSuccess, exitCode)
	}
}

func TestCLI_Run_InvalidAsyncAPIFile(t *testing.T) {
	// Create a temporary invalid AsyncAPI file
	tmpDir, err := os.MkdirTemp("", "cli_test_invalid_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create an invalid AsyncAPI spec (invalid JSON)
	invalidContent := `{
  "asyncapi": "2.6.0",
  "info": {
    "title": "Test API"
    "version": "1.0.0"  // Missing comma - invalid JSON
  }
}`

	inputFile := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(inputFile, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to write invalid AsyncAPI file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "output")

	// Create CLI and parse arguments
	cli := NewCLI()
	args := []string{
		"-i", inputFile,
		"-o", outputDir,
		"-p", "testpkg",
	}

	err = cli.ParseArgs(args)
	if err != nil {
		t.Fatalf("ParseArgs failed: %v", err)
	}

	// Run CLI command - should fail
	exitCode := cli.Run()
	if exitCode == ExitCodeSuccess {
		t.Error("Expected CLI to fail with invalid AsyncAPI file, but it succeeded")
	}
}

func TestCLI_Run_NonExistentInputFile(t *testing.T) {
	// Create CLI with non-existent input file
	cli := NewCLI()
	args := []string{
		"-i", "/non/existent/file.yaml",
		"-o", "./output",
		"-p", "testpkg",
	}

	// This should fail during argument validation
	err := cli.ParseArgs(args)
	if err == nil {
		t.Error("Expected ParseArgs to fail with non-existent file, but it succeeded")
	}

	if !strings.Contains(err.Error(), "input file does not exist") {
		t.Errorf("Expected error about non-existent file, got: %v", err)
	}
}

func TestCLI_Run_InvalidPackageName(t *testing.T) {
	// Create a temporary AsyncAPI file
	tmpDir, err := os.MkdirTemp("", "cli_test_pkg_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	asyncAPIContent := `{"asyncapi": "2.6.0", "info": {"title": "Test", "version": "1.0.0"}}`
	inputFile := filepath.Join(tmpDir, "asyncapi.json")
	if err := os.WriteFile(inputFile, []byte(asyncAPIContent), 0644); err != nil {
		t.Fatalf("Failed to write AsyncAPI file: %v", err)
	}

	// Create CLI with invalid package name
	cli := NewCLI()
	args := []string{
		"-i", inputFile,
		"-o", "./output",
		"-p", "Invalid-Package-Name",
	}

	// This should fail during argument validation
	err = cli.ParseArgs(args)
	if err == nil {
		t.Error("Expected ParseArgs to fail with invalid package name, but it succeeded")
	}

	if !strings.Contains(err.Error(), "invalid Go package name") {
		t.Errorf("Expected error about invalid package name, got: %v", err)
	}
}

func TestCLI_ExitCodes(t *testing.T) {
	tests := []struct {
		name     string
		expected int
		actual   int
	}{
		{"Success", 0, ExitCodeSuccess},
		{"General Error", 1, ExitCodeGeneralError},
		{"Config Error", 2, ExitCodeConfigError},
		{"Parse Error", 3, ExitCodeParseError},
		{"Generation Error", 4, ExitCodeGenerationError},
		{"Validation Error", 5, ExitCodeValidationError},
		{"File Error", 6, ExitCodeFileError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.actual != tt.expected {
				t.Errorf("Expected exit code %d, got %d", tt.expected, tt.actual)
			}
		})
	}
}

func TestCLI_PrintError(t *testing.T) {
	cli := NewCLI()

	// Test printError method doesn't panic
	cli.printError("Test error", nil)
	cli.printError("Test error with details", os.ErrNotExist)
}

func TestCLI_PrintVerbose(t *testing.T) {
	cli := NewCLI()

	// Test verbose output when verbose is disabled
	cli.printVerbose("This should not print")

	// Enable verbose mode
	cli.config.Verbose = true
	cli.printVerbose("This should print: %s", "test")
}

func TestCLI_HelpFlag_DoesNotExecute(t *testing.T) {
	cli := NewCLI()
	args := []string{"-h"}

	err := cli.ParseArgs(args)
	if err != nil {
		t.Fatalf("ParseArgs with help flag failed: %v", err)
	}

	if !cli.GetConfig().Help {
		t.Error("Expected Help flag to be true")
	}

	// Help flag should not cause Run() to be called in main
	// This test verifies the flag is properly set
}

func TestCLI_AllFlagVariations(t *testing.T) {
	// Create a temporary AsyncAPI file
	tmpDir, err := os.MkdirTemp("", "cli_test_flags_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	asyncAPIContent := `{"asyncapi": "2.6.0", "info": {"title": "Test", "version": "1.0.0"}}`
	inputFile := filepath.Join(tmpDir, "asyncapi.json")
	if err := os.WriteFile(inputFile, []byte(asyncAPIContent), 0644); err != nil {
		t.Fatalf("Failed to write AsyncAPI file: %v", err)
	}

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "short flags",
			args: []string{"-i", inputFile, "-o", "./output", "-p", "test", "-v"},
		},
		{
			name: "long flags",
			args: []string{"--input", inputFile, "--output", "./output", "--package", "test", "--verbose"},
		},
		{
			name: "mixed flags",
			args: []string{"-i", inputFile, "--output", "./output", "-p", "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := NewCLI()
			err := cli.ParseArgs(tt.args)
			if err != nil {
				t.Errorf("ParseArgs failed for %s: %v", tt.name, err)
			}

			config := cli.GetConfig()
			if config.InputFile != inputFile {
				t.Errorf("Expected InputFile %s, got %s", inputFile, config.InputFile)
			}
			if config.OutputDir != "./output" {
				t.Errorf("Expected OutputDir './output', got %s", config.OutputDir)
			}
			if config.PackageName != "test" {
				t.Errorf("Expected PackageName 'test', got %s", config.PackageName)
			}
		})
	}
}
