#!/bin/bash

# AsyncAPI Go Code Generator CLI Examples
# This script demonstrates various CLI usage patterns

set -e

echo "=== AsyncAPI Go Code Generator CLI Examples ==="
echo

# Ensure the CLI tool is built
if ! command -v evently-codegen &> /dev/null; then
    echo "Building evently-codegen CLI tool..."
    cd ../../
    go build -o evently-codegen ./cmd/evently-codegen
    export PATH="$PWD:$PATH"
    cd examples/cli_examples
    echo "✓ CLI tool built successfully"
    echo
fi

# Create temporary directory for examples
TEMP_DIR=$(mktemp -d)
echo "Working directory: $TEMP_DIR"
echo

# Example 1: Basic usage with YAML file
echo "=== Example 1: Basic Usage with YAML ===
evently-codegen -i ../../testdata/user-service.yaml -o $TEMP_DIR/example1 -p userevents"

evently-codegen -i ../../testdata/user-service.yaml -o "$TEMP_DIR/example1" -p userevents

echo "Generated files:"
ls -la "$TEMP_DIR/example1/"
echo

# Example 2: JSON file with verbose output
echo "=== Example 2: JSON File with Verbose Output ==="
echo "evently-codegen -i ../../testdata/ecommerce-api.json -o $TEMP_DIR/example2 -p ecommerce -v"

evently-codegen -i ../../testdata/ecommerce-api.json -o "$TEMP_DIR/example2" -p ecommerce -v

echo "Generated files:"
ls -la "$TEMP_DIR/example2/"
echo

# Example 3: Show help
echo "=== Example 3: Help Information ==="
echo "evently-codegen -h"
evently-codegen -h
echo

# Example 4: Error handling - invalid file
echo "=== Example 4: Error Handling - Invalid File ==="
echo "evently-codegen -i nonexistent.yaml -o $TEMP_DIR/example4 -p test"
if evently-codegen -i nonexistent.yaml -o "$TEMP_DIR/example4" -p test 2>&1; then
    echo "ERROR: Should have failed!"
    exit 1
else
    echo "✓ Correctly handled missing file error"
fi
echo

# Example 5: Error handling - invalid package name
echo "=== Example 5: Error Handling - Invalid Package Name ==="
echo "evently-codegen -i ../../testdata/user-service.yaml -o $TEMP_DIR/example5 -p 123invalid"
if evently-codegen -i ../../testdata/user-service.yaml -o "$TEMP_DIR/example5" -p 123invalid 2>&1; then
    echo "ERROR: Should have failed!"
    exit 1
else
    echo "✓ Correctly handled invalid package name error"
fi
echo

# Example 6: Batch processing multiple files
echo "=== Example 6: Batch Processing Multiple Files ==="
echo "Processing multiple AsyncAPI files..."

# Process user service
echo "Processing user-service.yaml..."
evently-codegen -i ../../testdata/user-service.yaml -o "$TEMP_DIR/batch/user" -p userevents

# Process ecommerce API
echo "Processing ecommerce-api.json..."
evently-codegen -i ../../testdata/ecommerce-api.json -o "$TEMP_DIR/batch/ecommerce" -p ecommerce

echo "Batch processing completed. Generated files:"
find "$TEMP_DIR/batch" -name "*.go" -exec echo "  {}" \;
echo

# Example 7: Show generated code preview
echo "=== Example 7: Generated Code Preview ==="
echo "Sample generated Go code from user-service.yaml:"
echo "--- UserSignupPayload struct ---"
if [ -f "$TEMP_DIR/example1/user_signup_payload.go" ]; then
    head -20 "$TEMP_DIR/example1/user_signup_payload.go"
else
    echo "Generated file not found, showing directory contents:"
    ls -la "$TEMP_DIR/example1/"
fi
echo

# Example 8: Integration with build script
echo "=== Example 8: Integration with Build Script ==="
cat > "$TEMP_DIR/build_script.sh" << 'EOF'
#!/bin/bash
# Example build script that generates Go types and builds the project

set -e

echo "Generating Go types from AsyncAPI specifications..."

# Generate types for different services
evently-codegen -i api/user-service.yaml -o pkg/events/user -p userevents
evently-codegen -i api/order-service.yaml -o pkg/events/order -p orderevents
evently-codegen -i api/payment-service.yaml -o pkg/events/payment -p paymentevents

echo "Building Go project..."
go build ./...

echo "Running tests..."
go test ./...

echo "Build completed successfully!"
EOF

chmod +x "$TEMP_DIR/build_script.sh"
echo "Created example build script at: $TEMP_DIR/build_script.sh"
echo "Content:"
cat "$TEMP_DIR/build_script.sh"
echo

# Example 9: Makefile integration
echo "=== Example 9: Makefile Integration ==="
cat > "$TEMP_DIR/Makefile" << 'EOF'
# Example Makefile with AsyncAPI code generation

.PHONY: generate build test clean

# Generate Go types from AsyncAPI specifications
generate:
	@echo "Generating Go types from AsyncAPI specifications..."
	evently-codegen -i api/user-service.yaml -o pkg/events/user -p userevents
	evently-codegen -i api/ecommerce-api.json -o pkg/events/ecommerce -p ecommerce
	@echo "Code generation completed"

# Build the project
build: generate
	@echo "Building project..."
	go build ./...

# Run tests
test: generate
	@echo "Running tests..."
	go test ./...

# Clean generated files
clean:
	@echo "Cleaning generated files..."
	rm -rf pkg/events/*/
	@echo "Clean completed"

# Full build pipeline
all: clean generate build test
	@echo "Full build pipeline completed"
EOF

echo "Created example Makefile at: $TEMP_DIR/Makefile"
echo "Content:"
cat "$TEMP_DIR/Makefile"
echo

# Summary
echo "=== Summary ==="
echo "CLI examples completed successfully!"
echo "Generated files are located in: $TEMP_DIR"
echo
echo "Key CLI usage patterns demonstrated:"
echo "1. Basic usage with YAML and JSON files"
echo "2. Verbose output for debugging"
echo "3. Error handling for invalid inputs"
echo "4. Batch processing multiple files"
echo "5. Integration with build scripts and Makefiles"
echo
echo "To explore the generated files:"
echo "  cd $TEMP_DIR"
echo "  find . -name '*.go' -exec head -10 {} +"
echo
echo "To clean up:"
echo "  rm -rf $TEMP_DIR"