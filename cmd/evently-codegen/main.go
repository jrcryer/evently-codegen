package main

import (
	"os"

	"github.com/jrcryer/evently-codegen/internal/cli"
)

func main() {
	// Create CLI instance
	cliApp := cli.NewCLI()

	// Parse command-line arguments
	if err := cliApp.ParseArgs(os.Args[1:]); err != nil {
		// Error already printed by ParseArgs
		os.Exit(cli.ExitCodeGeneralError)
	}

	// Check if help was requested (no error, but don't continue)
	if cliApp.GetConfig().Help {
		os.Exit(cli.ExitCodeSuccess)
	}

	// Run the CLI command and exit with appropriate code
	exitCode := cliApp.Run()
	os.Exit(exitCode)
}
