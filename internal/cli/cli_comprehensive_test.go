package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCLI_HelpOutput tests that help output is displayed correctly
func TestCLI_HelpOutput(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"short help flag", []string{"-h"}},
		{"long help flag", []string{"--help"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr output
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			cli := NewCLI()
			err := cli.ParseArgs(tt.args)

			// Close writer and restore stderr
			w.Close()
			os.Stderr = oldStderr

			// Read captured output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Should not return error for help
			if err != nil {
				t.Errorf("ParseArgs with help flag should not return error, got: %v", err)
				return
			}

			// Verify help content
			expectedContent := []string{
				"AsyncAPI Go Code Generator",
				"USAGE:",
				"OPTIONS:",
				"EXAMPLES:",
				"-i, --input",
				"-o, --output",
				"-p, --package",
				"-v, --verbose",
				"-h, --help",
			}

			for _, expected := range expectedContent {
				if !strings.Contains(output, expected) {
					t.Errorf("Help output should contain '%s'\nActual output:\n%s", expected, output)
				}
			}
		})
	}
}

// TestCLI_UsageOutput tests usage output when no arguments provided
func TestCLI_UsageOutput(t *testing.T) {
	// Capture stderr output
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	cli := NewCLI()
	err := cli.ParseArgs([]string{})

	// Close writer and restore stderr
	w.Close()
	os.Stderr = oldStderr

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	_ = buf.String() // We don't need to check stderr output for this test

	// Should return error for missing required args
	if err == nil {
		t.Error("ParseArgs with no arguments should return error")
	}

	// Error should mention input file requirement
	if !strings.Contains(err.Error(), "input file is required") {
		t.Errorf("Error should mention input file requirement, got: %v", err)
	}
}

// TestCLI_ExitCodeScenarios tests various exit code scenarios
func TestCLI_ExitCodeScenarios(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func() (string, func()) // Returns input file path and cleanup function
		args         []string
		expectedCode int
		description  string
	}{
		{
			name: "config validation error",
			setupFunc: func() (string, func()) {
				tmpFile, _ := os.CreateTemp("", "test*.yaml")
				tmpFile.WriteString(`{"asyncapi": "2.6.0", "info": {"title": "Test", "version": "1.0.0"}}`)
				tmpFile.Close()
				return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
			},
			args:         []string{"-i", "", "-o", "", "-p", ""},
			expectedCode: ExitCodeConfigError,
			description:  "Empty output directory should cause config error",
		},
		{
			name: "invalid AsyncAPI file",
			setupFunc: func() (string, func()) {
				tmpFile, _ := os.CreateTemp("", "test*.yaml")
				tmpFile.WriteString(`invalid json content`)
				tmpFile.Close()
				return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
			},
			args:         []string{"-i", "", "-o", "./output", "-p", "test"},
			expectedCode: ExitCodeGenerationError,
			description:  "Invalid AsyncAPI content should cause generation error",
		},
		{
			name: "successful generation",
			setupFunc: func() (string, func()) {
				tmpDir, _ := os.MkdirTemp("", "cli_test_*")
				tmpFile := filepath.Join(tmpDir, "test.yaml")
				os.WriteFile(tmpFile, []byte(`{
					"asyncapi": "2.6.0",
					"info": {"title": "Test", "version": "1.0.0"},
					"channels": {
						"test": {
							"publish": {
								"message": {
									"payload": {
										"type": "object",
										"properties": {
											"id": {"type": "string"}
										}
									}
								}
							}
						}
					}
				}`), 0644)
				return tmpFile, func() { os.RemoveAll(tmpDir) }
			},
			args:         []string{"-i", "", "-o", "", "-p", "test"},
			expectedCode: ExitCodeSuccess,
			description:  "Valid AsyncAPI file should succeed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputFile, cleanup := tt.setupFunc()
			defer cleanup()

			// Replace placeholders in args
			args := make([]string, len(tt.args))
			copy(args, tt.args)
			for i, arg := range args {
				if arg == "" && i > 0 && args[i-1] == "-i" {
					args[i] = inputFile
				} else if arg == "" && i > 0 && args[i-1] == "-o" {
					tmpDir, _ := os.MkdirTemp("", "output_*")
					defer os.RemoveAll(tmpDir)
					args[i] = tmpDir
				}
			}

			cli := NewCLI()

			// For config error test, we expect ParseArgs to fail
			if tt.expectedCode == ExitCodeConfigError {
				err := cli.ParseArgs(args)
				if err == nil {
					t.Errorf("Expected ParseArgs to fail for config error scenario")
					return
				}
				return // ParseArgs failed as expected
			}

			err := cli.ParseArgs(args)
			if err != nil {
				t.Errorf("ParseArgs failed: %v", err)
				return
			}

			exitCode := cli.Run()
			if exitCode != tt.expectedCode {
				t.Errorf("Expected exit code %d (%s), got %d", tt.expectedCode, tt.description, exitCode)
			}
		})
	}
}

// TestCLI_EdgeCases tests various edge cases
func TestCLI_EdgeCases(t *testing.T) {
	t.Run("empty flag values", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test*.yaml")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		// Test empty package name
		cli1 := NewCLI()
		err = cli1.ParseArgs([]string{"-i", tmpFile.Name(), "-p", ""})
		if err == nil {
			t.Error("Expected error for empty package name")
		}
		if !strings.Contains(err.Error(), "package name cannot be empty") {
			t.Errorf("Expected error about empty package name, got: %v", err)
		}

		// Test empty output directory - need new CLI instance
		cli2 := NewCLI()
		err = cli2.ParseArgs([]string{"-i", tmpFile.Name(), "-o", ""})
		if err == nil {
			t.Error("Expected error for empty output directory")
		}
		if !strings.Contains(err.Error(), "output directory cannot be empty") {
			t.Errorf("Expected error about empty output directory, got: %v", err)
		}
	})

	t.Run("special characters in paths", func(t *testing.T) {
		// Create temp file with special characters in path
		tmpDir, err := os.MkdirTemp("", "test with spaces*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		tmpFile := filepath.Join(tmpDir, "test file.yaml")
		err = os.WriteFile(tmpFile, []byte(`{"asyncapi": "2.6.0", "info": {"title": "Test", "version": "1.0.0"}}`), 0644)
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		cli := NewCLI()
		err = cli.ParseArgs([]string{"-i", tmpFile, "-o", tmpDir, "-p", "test"})
		if err != nil {
			t.Errorf("Should handle paths with spaces, got error: %v", err)
		}
	})

	t.Run("very long arguments", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test*.yaml")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		// Very long package name (but valid)
		longPackageName := strings.Repeat("a", 100)

		cli := NewCLI()
		err = cli.ParseArgs([]string{"-i", tmpFile.Name(), "-p", longPackageName})
		if err != nil {
			t.Errorf("Should handle long package names, got error: %v", err)
		}

		config := cli.GetConfig()
		if config.PackageName != longPackageName {
			t.Errorf("Package name not set correctly")
		}
	})

	t.Run("duplicate flags", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test*.yaml")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		cli := NewCLI()
		// Last flag value should win
		err = cli.ParseArgs([]string{"-i", tmpFile.Name(), "-p", "first", "-p", "second"})
		if err != nil {
			t.Errorf("Should handle duplicate flags, got error: %v", err)
		}

		config := cli.GetConfig()
		if config.PackageName != "second" {
			t.Errorf("Expected package name 'second', got '%s'", config.PackageName)
		}
	})
}

// TestCLI_ArgumentParsing tests comprehensive argument parsing scenarios
func TestCLI_ArgumentParsing(t *testing.T) {
	t.Run("flag parsing order independence", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test*.yaml")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		testCases := [][]string{
			{"-i", tmpFile.Name(), "-o", "./output", "-p", "test", "-v"},
			{"-v", "-p", "test", "-o", "./output", "-i", tmpFile.Name()},
			{"-p", "test", "-v", "-i", tmpFile.Name(), "-o", "./output"},
		}

		for i, args := range testCases {
			t.Run(fmt.Sprintf("order_%d", i), func(t *testing.T) {
				cli := NewCLI()
				err := cli.ParseArgs(args)
				if err != nil {
					t.Errorf("ParseArgs failed for order %d: %v", i, err)
				}

				config := cli.GetConfig()
				if config.InputFile != tmpFile.Name() {
					t.Errorf("InputFile not parsed correctly")
				}
				if config.OutputDir != "./output" {
					t.Errorf("OutputDir not parsed correctly")
				}
				if config.PackageName != "test" {
					t.Errorf("PackageName not parsed correctly")
				}
				if !config.Verbose {
					t.Errorf("Verbose flag not parsed correctly")
				}
			})
		}
	})

	t.Run("mixed short and long flags", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test*.yaml")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		cli := NewCLI()
		err = cli.ParseArgs([]string{
			"--input", tmpFile.Name(),
			"-o", "./output",
			"--package", "test",
			"-v",
		})
		if err != nil {
			t.Errorf("ParseArgs failed with mixed flags: %v", err)
		}

		config := cli.GetConfig()
		if config.InputFile != tmpFile.Name() {
			t.Errorf("InputFile not parsed correctly with mixed flags")
		}
		if config.OutputDir != "./output" {
			t.Errorf("OutputDir not parsed correctly with mixed flags")
		}
		if config.PackageName != "test" {
			t.Errorf("PackageName not parsed correctly with mixed flags")
		}
		if !config.Verbose {
			t.Errorf("Verbose flag not parsed correctly with mixed flags")
		}
	})
}

// TestCLI_PackageNameValidation tests comprehensive package name validation
func TestCLI_PackageNameValidation(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		expectValid bool
		description string
	}{
		{"simple lowercase", "main", true, "basic valid package name"},
		{"with underscore", "my_package", true, "underscore is allowed"},
		{"with numbers", "pkg123", true, "numbers are allowed after letters"},
		{"starts with underscore", "_private", true, "can start with underscore"},
		{"empty", "", false, "empty string not allowed"},
		{"starts with number", "1pkg", false, "cannot start with number"},
		{"uppercase letters", "MyPackage", false, "uppercase not allowed"},
		{"hyphen", "my-package", false, "hyphen not allowed"},
		{"space", "my package", false, "space not allowed"},
		{"special chars", "pkg@name", false, "special characters not allowed"},
		{"dot", "my.package", false, "dot not allowed"},
		{"mixed case", "myPackage", false, "camelCase not allowed"},
		{"all caps", "PACKAGE", false, "all caps not allowed"},
		{"number only", "123", false, "number only not allowed"},
		{"underscore only", "_", true, "single underscore is valid"},
		{"long name", strings.Repeat("a", 50), true, "long names are allowed"},
		{"unicode", "пакет", false, "unicode not allowed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidGoPackageName(tt.packageName)
			if result != tt.expectValid {
				t.Errorf("isValidGoPackageName(%q) = %t, expected %t (%s)",
					tt.packageName, result, tt.expectValid, tt.description)
			}
		})
	}
}

// TestCLI_FileExtensionValidation tests file extension validation
func TestCLI_FileExtensionValidation(t *testing.T) {
	validExtensions := []string{".json", ".yaml", ".yml"}
	invalidExtensions := []string{".txt", ".xml", ".toml", ".ini", "", ".JSON", ".YAML"}

	// Test valid extensions
	for _, ext := range validExtensions {
		t.Run("valid_"+ext, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "test*"+ext)
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())
			tmpFile.Close()

			cli := NewCLI()
			err = cli.ParseArgs([]string{"-i", tmpFile.Name()})
			if err != nil {
				t.Errorf("Valid extension %s should be accepted, got error: %v", ext, err)
			}
		})
	}

	// Test invalid extensions
	for _, ext := range invalidExtensions {
		t.Run("invalid_"+ext, func(t *testing.T) {
			filename := "test" + ext
			if ext == "" {
				filename = "test"
			}

			cli := NewCLI()
			err := cli.ParseArgs([]string{"-i", filename})
			if err == nil {
				t.Errorf("Invalid extension %s should be rejected", ext)
			}
			// For uppercase extensions, the error might be about file existence first
			// since the validation order checks extension before existence
			if !strings.Contains(err.Error(), "input file must have .json, .yaml, or .yml extension") &&
				!strings.Contains(err.Error(), "input file does not exist") {
				t.Errorf("Expected extension or file existence validation error, got: %v", err)
			}
		})
	}
}

// TestCLI_VerboseOutput tests verbose output functionality
func TestCLI_VerboseOutput(t *testing.T) {
	t.Run("verbose disabled by default", func(t *testing.T) {
		cli := NewCLI()
		if cli.GetConfig().Verbose {
			t.Error("Verbose should be disabled by default")
		}
	})

	t.Run("verbose enabled with short flag", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test*.yaml")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		cli := NewCLI()
		err = cli.ParseArgs([]string{"-i", tmpFile.Name(), "-v"})
		if err != nil {
			t.Errorf("ParseArgs failed: %v", err)
		}

		if !cli.GetConfig().Verbose {
			t.Error("Verbose should be enabled with -v flag")
		}
	})

	t.Run("verbose enabled with long flag", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test*.yaml")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		cli := NewCLI()
		err = cli.ParseArgs([]string{"-i", tmpFile.Name(), "--verbose"})
		if err != nil {
			t.Errorf("ParseArgs failed: %v", err)
		}

		if !cli.GetConfig().Verbose {
			t.Error("Verbose should be enabled with --verbose flag")
		}
	})
}
