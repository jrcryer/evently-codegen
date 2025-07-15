package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jrcryer/evently-codegen/pkg/generator"
)

// CLIConfig holds command-line configuration
type CLIConfig struct {
	InputFile   string
	OutputDir   string
	PackageName string
	Verbose     bool
	Help        bool
}

// CLI handles command-line operations
type CLI struct {
	config  *CLIConfig
	flagSet *flag.FlagSet
	appName string
	version string
}

// NewCLI creates a new CLI instance
func NewCLI() *CLI {
	return &CLI{
		config: &CLIConfig{
			OutputDir:   "./generated",
			PackageName: "main",
			Verbose:     false,
			Help:        false,
		},
		flagSet: flag.NewFlagSet("evently-codegen", flag.ContinueOnError),
		appName: "evently-codegen",
		version: "1.0.0",
	}
}

// GetConfig returns the CLI configuration
func (c *CLI) GetConfig() *CLIConfig {
	return c.config
}

// ParseArgs parses command-line arguments and validates them
func (c *CLI) ParseArgs(args []string) error {
	// Define flags
	c.flagSet.StringVar(&c.config.InputFile, "input", "", "Path to AsyncAPI specification file (required)")
	c.flagSet.StringVar(&c.config.InputFile, "i", "", "Path to AsyncAPI specification file (required) (shorthand)")
	c.flagSet.StringVar(&c.config.OutputDir, "output", "./generated", "Output directory for generated Go files")
	c.flagSet.StringVar(&c.config.OutputDir, "o", "./generated", "Output directory for generated Go files (shorthand)")
	c.flagSet.StringVar(&c.config.PackageName, "package", "main", "Package name for generated Go code")
	c.flagSet.StringVar(&c.config.PackageName, "p", "main", "Package name for generated Go code (shorthand)")
	c.flagSet.BoolVar(&c.config.Verbose, "verbose", false, "Enable verbose output")
	c.flagSet.BoolVar(&c.config.Verbose, "v", false, "Enable verbose output (shorthand)")
	c.flagSet.BoolVar(&c.config.Help, "help", false, "Show help information")
	c.flagSet.BoolVar(&c.config.Help, "h", false, "Show help information (shorthand)")

	// Set custom usage function
	c.flagSet.Usage = c.printUsage

	// Parse arguments
	if err := c.flagSet.Parse(args); err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Show help if requested
	if c.config.Help {
		c.printUsage()
		return nil
	}

	// Validate required parameters
	if err := c.validateArgs(); err != nil {
		return err
	}

	return nil
}

// validateArgs validates the parsed command-line arguments
func (c *CLI) validateArgs() error {
	var errors []string

	// Validate input file
	if c.config.InputFile == "" {
		errors = append(errors, "input file is required")
	} else {
		// Check file extension first
		ext := strings.ToLower(filepath.Ext(c.config.InputFile))
		if ext != ".json" && ext != ".yaml" && ext != ".yml" {
			errors = append(errors, "input file must have .json, .yaml, or .yml extension")
		}

		// Then check if input file exists
		if _, err := os.Stat(c.config.InputFile); os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("input file does not exist: %s", c.config.InputFile))
		}
	}

	// Validate package name
	if c.config.PackageName == "" {
		errors = append(errors, "package name cannot be empty")
	} else if !isValidGoPackageName(c.config.PackageName) {
		errors = append(errors, fmt.Sprintf("invalid Go package name: %s", c.config.PackageName))
	}

	// Validate output directory
	if c.config.OutputDir == "" {
		errors = append(errors, "output directory cannot be empty")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}

// printUsage prints detailed usage information
func (c *CLI) printUsage() {
	fmt.Fprintf(os.Stderr, "%s v%s - AsyncAPI Go Code Generator\n\n", c.appName, c.version)
	fmt.Fprintf(os.Stderr, "USAGE:\n")
	fmt.Fprintf(os.Stderr, "  %s [OPTIONS]\n\n", c.appName)
	fmt.Fprintf(os.Stderr, "DESCRIPTION:\n")
	fmt.Fprintf(os.Stderr, "  Generates Go type definitions from AsyncAPI specification files.\n")
	fmt.Fprintf(os.Stderr, "  Supports AsyncAPI 2.x and 3.x specifications in JSON or YAML format.\n\n")
	fmt.Fprintf(os.Stderr, "OPTIONS:\n")
	fmt.Fprintf(os.Stderr, "  -i, --input <file>      Path to AsyncAPI specification file (required)\n")
	fmt.Fprintf(os.Stderr, "  -o, --output <dir>      Output directory for generated Go files (default: ./generated)\n")
	fmt.Fprintf(os.Stderr, "  -p, --package <name>    Package name for generated Go code (default: main)\n")
	fmt.Fprintf(os.Stderr, "  -v, --verbose           Enable verbose output\n")
	fmt.Fprintf(os.Stderr, "  -h, --help              Show this help information\n\n")
	fmt.Fprintf(os.Stderr, "EXAMPLES:\n")
	fmt.Fprintf(os.Stderr, "  %s -i api.yaml -o ./types -p events\n", c.appName)
	fmt.Fprintf(os.Stderr, "  %s --input asyncapi.json --output ./generated --package myapi\n", c.appName)
	fmt.Fprintf(os.Stderr, "  %s -i spec.yml -v\n\n", c.appName)
}

// isValidGoPackageName checks if a string is a valid Go package name
func isValidGoPackageName(name string) bool {
	if name == "" {
		return false
	}

	// Go package names must start with a letter or underscore
	if !isLetter(rune(name[0])) && name[0] != '_' {
		return false
	}

	// Rest of the characters must be letters, digits, or underscores
	for _, r := range name[1:] {
		if !isLetter(r) && !isDigit(r) && r != '_' {
			return false
		}
	}

	// Package names should be lowercase (convention)
	if strings.ToLower(name) != name {
		return false
	}

	return true
}

// isLetter checks if a rune is a letter
func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

// isDigit checks if a rune is a digit
func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// Run executes the CLI command with the parsed configuration
func (c *CLI) Run() int {
	// Create generator configuration from CLI config
	genConfig := &generator.Config{
		PackageName:     c.config.PackageName,
		OutputDir:       c.config.OutputDir,
		IncludeComments: true,
		UsePointers:     true,
	}

	// Create generator instance
	gen := generator.NewGenerator(genConfig)

	// Validate configuration
	if err := gen.ValidateConfig(); err != nil {
		c.printError("Configuration validation failed", err)
		return ExitCodeConfigError
	}

	if c.config.Verbose {
		c.printVerbose("Starting AsyncAPI code generation...")
		c.printVerbose("Input file: %s", c.config.InputFile)
		c.printVerbose("Output directory: %s", c.config.OutputDir)
		c.printVerbose("Package name: %s", c.config.PackageName)
	}

	// Parse and generate files
	if err := gen.ParseFileAndGenerateToFiles(c.config.InputFile); err != nil {
		c.printError("Code generation failed", err)
		return ExitCodeGenerationError
	}

	if c.config.Verbose {
		c.printVerbose("Code generation completed successfully!")
	} else {
		fmt.Printf("Generated Go types in %s\n", c.config.OutputDir)
	}

	return ExitCodeSuccess
}

// printError prints an error message to stderr
func (c *CLI) printError(message string, err error) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", message)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Details: %v\n", err)
	}
}

// printVerbose prints a verbose message if verbose mode is enabled
func (c *CLI) printVerbose(format string, args ...interface{}) {
	if c.config.Verbose {
		fmt.Printf("[VERBOSE] "+format+"\n", args...)
	}
}

// Exit codes for different error conditions
const (
	ExitCodeSuccess         = 0
	ExitCodeGeneralError    = 1
	ExitCodeConfigError     = 2
	ExitCodeParseError      = 3
	ExitCodeGenerationError = 4
	ExitCodeValidationError = 5
	ExitCodeFileError       = 6
)
