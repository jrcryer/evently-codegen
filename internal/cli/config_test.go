package cli

import (
	"os"
	"strings"
	"testing"
)

func TestNewCLI(t *testing.T) {
	cli := NewCLI()

	if cli == nil {
		t.Fatal("NewCLI() returned nil")
	}

	config := cli.GetConfig()
	if config.OutputDir != "./generated" {
		t.Errorf("Expected default OutputDir to be './generated', got %s", config.OutputDir)
	}

	if config.PackageName != "main" {
		t.Errorf("Expected default PackageName to be 'main', got %s", config.PackageName)
	}

	if config.Verbose != false {
		t.Errorf("Expected default Verbose to be false, got %t", config.Verbose)
	}

	if config.Help != false {
		t.Errorf("Expected default Help to be false, got %t", config.Help)
	}
}

func TestParseArgs_ValidArguments(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	tests := []struct {
		name     string
		args     []string
		expected CLIConfig
	}{
		{
			name: "short flags",
			args: []string{"-i", tmpFile.Name(), "-o", "./output", "-p", "testpkg", "-v"},
			expected: CLIConfig{
				InputFile:   tmpFile.Name(),
				OutputDir:   "./output",
				PackageName: "testpkg",
				Verbose:     true,
				Help:        false,
			},
		},
		{
			name: "long flags",
			args: []string{"--input", tmpFile.Name(), "--output", "./output", "--package", "testpkg", "--verbose"},
			expected: CLIConfig{
				InputFile:   tmpFile.Name(),
				OutputDir:   "./output",
				PackageName: "testpkg",
				Verbose:     true,
				Help:        false,
			},
		},
		{
			name: "minimal required args",
			args: []string{"-i", tmpFile.Name()},
			expected: CLIConfig{
				InputFile:   tmpFile.Name(),
				OutputDir:   "./generated",
				PackageName: "main",
				Verbose:     false,
				Help:        false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := NewCLI()
			err := cli.ParseArgs(tt.args)

			if err != nil {
				t.Fatalf("ParseArgs() returned error: %v", err)
			}

			config := cli.GetConfig()
			if config.InputFile != tt.expected.InputFile {
				t.Errorf("Expected InputFile %s, got %s", tt.expected.InputFile, config.InputFile)
			}
			if config.OutputDir != tt.expected.OutputDir {
				t.Errorf("Expected OutputDir %s, got %s", tt.expected.OutputDir, config.OutputDir)
			}
			if config.PackageName != tt.expected.PackageName {
				t.Errorf("Expected PackageName %s, got %s", tt.expected.PackageName, config.PackageName)
			}
			if config.Verbose != tt.expected.Verbose {
				t.Errorf("Expected Verbose %t, got %t", tt.expected.Verbose, config.Verbose)
			}
			if config.Help != tt.expected.Help {
				t.Errorf("Expected Help %t, got %t", tt.expected.Help, config.Help)
			}
		})
	}
}

func TestParseArgs_HelpFlag(t *testing.T) {
	tests := []string{"-h", "--help"}

	for _, flag := range tests {
		t.Run(flag, func(t *testing.T) {
			cli := NewCLI()
			err := cli.ParseArgs([]string{flag})

			if err != nil {
				t.Fatalf("ParseArgs() with %s returned error: %v", flag, err)
			}

			config := cli.GetConfig()
			if !config.Help {
				t.Errorf("Expected Help to be true when %s flag is used", flag)
			}
		})
	}
}

func TestParseArgs_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError string
	}{
		{
			name:        "missing input file",
			args:        []string{},
			expectError: "input file is required",
		},
		{
			name:        "non-existent input file",
			args:        []string{"-i", "/non/existent/file.yaml"},
			expectError: "input file does not exist",
		},
		{
			name:        "invalid file extension",
			args:        []string{"-i", "test.txt"},
			expectError: "input file must have .json, .yaml, or .yml extension",
		},
		{
			name:        "empty package name",
			args:        []string{"-i", "test.yaml", "-p", ""},
			expectError: "package name cannot be empty",
		},
		{
			name:        "invalid package name with uppercase",
			args:        []string{"-i", "test.yaml", "-p", "TestPkg"},
			expectError: "invalid Go package name",
		},
		{
			name:        "invalid package name starting with digit",
			args:        []string{"-i", "test.yaml", "-p", "1pkg"},
			expectError: "invalid Go package name",
		},
		{
			name:        "empty output directory",
			args:        []string{"-i", "test.yaml", "-o", ""},
			expectError: "output directory cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file for tests that need existing file
			if !strings.Contains(tt.expectError, "input file does not exist") &&
				!strings.Contains(tt.expectError, "input file must have") &&
				len(tt.args) > 0 && tt.args[0] != "" {
				tmpFile, err := os.CreateTemp("", "test*.yaml")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				defer os.Remove(tmpFile.Name())
				tmpFile.Close()

				// Replace the input file argument with temp file
				for i, arg := range tt.args {
					if arg == "test.yaml" {
						tt.args[i] = tmpFile.Name()
						break
					}
				}
			}

			cli := NewCLI()
			err := cli.ParseArgs(tt.args)

			if err == nil {
				t.Fatalf("Expected error containing '%s', but got no error", tt.expectError)
			}

			if !strings.Contains(err.Error(), tt.expectError) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.expectError, err.Error())
			}
		})
	}
}

func TestIsValidGoPackageName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid lowercase", "main", true},
		{"valid with underscore", "my_package", true},
		{"valid with numbers", "pkg123", true},
		{"empty string", "", false},
		{"starts with number", "1pkg", false},
		{"contains uppercase", "MyPackage", false},
		{"contains special chars", "pkg-name", false},
		{"contains spaces", "my package", false},
		{"starts with underscore", "_private", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidGoPackageName(tt.input)
			if result != tt.expected {
				t.Errorf("isValidGoPackageName(%q) = %t, expected %t", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseArgs_FileExtensions(t *testing.T) {
	extensions := []string{".json", ".yaml", ".yml"}

	for _, ext := range extensions {
		t.Run(ext, func(t *testing.T) {
			// Create temp file with specific extension
			tmpFile, err := os.CreateTemp("", "test*"+ext)
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())
			tmpFile.Close()

			cli := NewCLI()
			err = cli.ParseArgs([]string{"-i", tmpFile.Name()})

			if err != nil {
				t.Errorf("ParseArgs() with %s file returned error: %v", ext, err)
			}
		})
	}
}

func TestParseArgs_InvalidFlags(t *testing.T) {
	cli := NewCLI()
	err := cli.ParseArgs([]string{"--invalid-flag"})

	if err == nil {
		t.Fatal("Expected error for invalid flag, but got none")
	}

	if !strings.Contains(err.Error(), "failed to parse arguments") {
		t.Errorf("Expected error about parsing arguments, got: %v", err)
	}
}
