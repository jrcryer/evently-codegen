package generator

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// DefaultSchemaResolver implements the SchemaResolver interface
type DefaultSchemaResolver struct {
	cache           map[string]interface{} // Cache for resolved schemas
	resolutionStack []string               // Stack to detect circular references
	baseURI         string                 // Base URI for relative references
	httpClient      *http.Client           // HTTP client for external references
}

// NewSchemaResolver creates a new schema resolver instance
func NewSchemaResolver(baseURI string) *DefaultSchemaResolver {
	return &DefaultSchemaResolver{
		cache:           make(map[string]interface{}),
		resolutionStack: make([]string, 0),
		baseURI:         baseURI,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ResolveRef resolves a $ref to a MessageSchema
func (r *DefaultSchemaResolver) ResolveRef(ref string) (*MessageSchema, error) {
	if ref == "" {
		return nil, &ResolverError{
			Reference: ref,
			Message:   "empty reference",
		}
	}

	// Check for circular reference
	if r.isCircularReference(ref) {
		return nil, &CircularReferenceError{
			Reference: ref,
			Stack:     append([]string{}, r.resolutionStack...),
		}
	}

	// Check cache first
	if cached, exists := r.cache[ref]; exists {
		if schema, ok := cached.(*MessageSchema); ok {
			return schema, nil
		}
	}

	// Add to resolution stack
	r.resolutionStack = append(r.resolutionStack, ref)
	defer func() {
		// Remove from stack when done
		if len(r.resolutionStack) > 0 {
			r.resolutionStack = r.resolutionStack[:len(r.resolutionStack)-1]
		}
	}()

	// Resolve the reference
	resolved, err := r.resolveReference(ref)
	if err != nil {
		return nil, err
	}

	// Convert to MessageSchema
	schema, err := r.convertToMessageSchema(resolved)
	if err != nil {
		return nil, &ResolverError{
			Reference: ref,
			Message:   fmt.Sprintf("failed to convert resolved reference to MessageSchema: %v", err),
		}
	}

	// Cache the result
	r.cache[ref] = schema

	return schema, nil
}

// ResolveProperty resolves a $ref to a Property
func (r *DefaultSchemaResolver) ResolveProperty(ref string) (*Property, error) {
	if ref == "" {
		return nil, &ResolverError{
			Reference: ref,
			Message:   "empty reference",
		}
	}

	// Check for circular reference
	if r.isCircularReference(ref) {
		return nil, &CircularReferenceError{
			Reference: ref,
			Stack:     append([]string{}, r.resolutionStack...),
		}
	}

	// Check cache first
	if cached, exists := r.cache[ref]; exists {
		if property, ok := cached.(*Property); ok {
			return property, nil
		}
	}

	// Add to resolution stack
	r.resolutionStack = append(r.resolutionStack, ref)
	defer func() {
		// Remove from stack when done
		if len(r.resolutionStack) > 0 {
			r.resolutionStack = r.resolutionStack[:len(r.resolutionStack)-1]
		}
	}()

	// Resolve the reference
	resolved, err := r.resolveReference(ref)
	if err != nil {
		return nil, err
	}

	// Convert to Property
	property, err := r.convertToProperty(resolved)
	if err != nil {
		return nil, &ResolverError{
			Reference: ref,
			Message:   fmt.Sprintf("failed to convert resolved reference to Property: %v", err),
		}
	}

	// Cache the result
	r.cache[ref] = property

	return property, nil
}

// resolveReference resolves a reference to raw data
func (r *DefaultSchemaResolver) resolveReference(ref string) (interface{}, error) {
	// Parse the reference
	uri, fragment, err := r.parseReference(ref)
	if err != nil {
		return nil, &ResolverError{
			Reference: ref,
			Message:   fmt.Sprintf("failed to parse reference: %v", err),
		}
	}

	// Load the document
	var document interface{}
	if uri == "" {
		// Fragment-only reference - should be resolved against current document
		return nil, &ResolverError{
			Reference: ref,
			Message:   "fragment-only references not supported without base document",
		}
	} else {
		// Load external document
		document, err = r.loadDocument(uri)
		if err != nil {
			return nil, &ResolverError{
				Reference: ref,
				Message:   fmt.Sprintf("failed to load document '%s': %v", uri, err),
			}
		}
	}

	// Resolve fragment within document
	if fragment != "" {
		resolved, err := r.resolveFragment(document, fragment)
		if err != nil {
			return nil, &ResolverError{
				Reference: ref,
				Message:   fmt.Sprintf("failed to resolve fragment '%s': %v", fragment, err),
			}
		}
		return resolved, nil
	}

	return document, nil
}

// parseReference parses a reference into URI and fragment parts
func (r *DefaultSchemaResolver) parseReference(ref string) (uri, fragment string, err error) {
	// Split on '#' to separate URI and fragment
	parts := strings.SplitN(ref, "#", 2)

	uri = parts[0]
	if len(parts) > 1 {
		fragment = parts[1]
	}

	// If URI is relative, resolve against base URI
	if uri != "" && !r.isAbsoluteURI(uri) {
		if r.baseURI == "" {
			return "", "", fmt.Errorf("relative reference '%s' requires base URI", ref)
		}

		baseURL, err := url.Parse(r.baseURI)
		if err != nil {
			return "", "", fmt.Errorf("invalid base URI '%s': %v", r.baseURI, err)
		}

		refURL, err := url.Parse(uri)
		if err != nil {
			return "", "", fmt.Errorf("invalid reference URI '%s': %v", uri, err)
		}

		resolvedURL := baseURL.ResolveReference(refURL)
		uri = resolvedURL.String()
	}

	return uri, fragment, nil
}

// isAbsoluteURI checks if a URI is absolute
func (r *DefaultSchemaResolver) isAbsoluteURI(uri string) bool {
	parsed, err := url.Parse(uri)
	if err != nil {
		return false
	}
	return parsed.IsAbs()
}

// loadDocument loads a document from a URI
func (r *DefaultSchemaResolver) loadDocument(uri string) (interface{}, error) {
	// Check cache first
	if cached, exists := r.cache[uri]; exists {
		return cached, nil
	}

	var data []byte
	var err error

	// Determine how to load the document
	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		// Load from HTTP
		data, err = r.loadHTTP(uri)
	} else {
		// Load from file system
		data, err = r.loadFile(uri)
	}

	if err != nil {
		return nil, err
	}

	// Parse the document
	var document interface{}

	// Try JSON first
	if err := json.Unmarshal(data, &document); err != nil {
		// If JSON fails, try YAML
		if yamlErr := yaml.Unmarshal(data, &document); yamlErr != nil {
			return nil, fmt.Errorf("failed to parse as JSON (%v) or YAML (%v)", err, yamlErr)
		}
	}

	// Cache the document
	r.cache[uri] = document

	return document, nil
}

// loadHTTP loads a document from an HTTP URL
func (r *DefaultSchemaResolver) loadHTTP(url string) ([]byte, error) {
	resp, err := r.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP response: %v", err)
	}

	return data, nil
}

// loadFile loads a document from the file system
func (r *DefaultSchemaResolver) loadFile(path string) ([]byte, error) {
	// Convert file:// URLs to file paths
	if newPath, found := strings.CutPrefix(path, "file://"); found {
		path = newPath
	}

	// Clean the path
	path = filepath.Clean(path)

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file '%s': %v", path, err)
	}

	return data, nil
}

// resolveFragment resolves a JSON Pointer fragment within a document
func (r *DefaultSchemaResolver) resolveFragment(document interface{}, fragment string) (interface{}, error) {
	if fragment == "" {
		return document, nil
	}

	// Handle JSON Pointer format
	if strings.HasPrefix(fragment, "/") {
		return r.resolveJSONPointer(document, fragment)
	}

	// Handle simple fragment (assume it's a key in the root object)
	if docMap, ok := document.(map[string]interface{}); ok {
		if value, exists := docMap[fragment]; exists {
			return value, nil
		}
		return nil, fmt.Errorf("fragment '%s' not found in document", fragment)
	}

	return nil, fmt.Errorf("unsupported fragment format: %s", fragment)
}

// resolveJSONPointer resolves a JSON Pointer within a document
func (r *DefaultSchemaResolver) resolveJSONPointer(document interface{}, pointer string) (interface{}, error) {
	if pointer == "/" {
		return document, nil
	}

	// Split pointer into tokens
	tokens := strings.Split(pointer[1:], "/") // Remove leading '/'

	current := document
	for _, token := range tokens {
		// Unescape JSON Pointer tokens
		token = strings.ReplaceAll(token, "~1", "/")
		token = strings.ReplaceAll(token, "~0", "~")

		switch v := current.(type) {
		case map[string]interface{}:
			if value, exists := v[token]; exists {
				current = value
			} else {
				return nil, fmt.Errorf("key '%s' not found in object", token)
			}
		case []interface{}:
			// Handle array index
			if token == "-" {
				return nil, fmt.Errorf("array index '-' not supported for resolution")
			}

			var index int
			if _, err := fmt.Sscanf(token, "%d", &index); err != nil {
				return nil, fmt.Errorf("invalid array index '%s'", token)
			}

			if index < 0 || index >= len(v) {
				return nil, fmt.Errorf("array index %d out of bounds", index)
			}

			current = v[index]
		default:
			return nil, fmt.Errorf("cannot resolve pointer '%s' in non-object/array value", token)
		}
	}

	return current, nil
}

// convertToMessageSchema converts a resolved reference to a MessageSchema
func (r *DefaultSchemaResolver) convertToMessageSchema(data interface{}) (*MessageSchema, error) {
	// Marshal to JSON and unmarshal to MessageSchema for type conversion
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resolved data: %v", err)
	}

	var schema MessageSchema
	if err := json.Unmarshal(jsonData, &schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to MessageSchema: %v", err)
	}

	return &schema, nil
}

// convertToProperty converts a resolved reference to a Property
func (r *DefaultSchemaResolver) convertToProperty(data interface{}) (*Property, error) {
	// Marshal to JSON and unmarshal to Property for type conversion
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resolved data: %v", err)
	}

	var property Property
	if err := json.Unmarshal(jsonData, &property); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to Property: %v", err)
	}

	return &property, nil
}

// isCircularReference checks if a reference would create a circular dependency
func (r *DefaultSchemaResolver) isCircularReference(ref string) bool {
	for _, stackRef := range r.resolutionStack {
		if stackRef == ref {
			return true
		}
	}
	return false
}

// ClearCache clears the resolver's cache
func (r *DefaultSchemaResolver) ClearCache() {
	r.cache = make(map[string]interface{})
}

// SetBaseURI sets the base URI for resolving relative references
func (r *DefaultSchemaResolver) SetBaseURI(baseURI string) {
	r.baseURI = baseURI
}

// GetCacheSize returns the number of cached items
func (r *DefaultSchemaResolver) GetCacheSize() int {
	return len(r.cache)
}
