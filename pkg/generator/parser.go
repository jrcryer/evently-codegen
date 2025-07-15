package generator

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// AsyncAPIParser implements the Parser interface
type AsyncAPIParser struct {
	supportedVersions []string
}

// NewAsyncAPIParser creates a new AsyncAPI parser instance
func NewAsyncAPIParser() *AsyncAPIParser {
	return &AsyncAPIParser{
		supportedVersions: []string{"2.0.0", "2.1.0", "2.2.0", "2.3.0", "2.4.0", "2.5.0", "2.6.0", "3.0.0"},
	}
}

// Parse parses AsyncAPI specification data and returns a ParseResult
func (p *AsyncAPIParser) Parse(data []byte) (*ParseResult, error) {
	if len(data) == 0 {
		return nil, &ParseError{Message: "empty input data"}
	}

	// Try to determine if it's JSON or YAML
	var spec AsyncAPISpec
	var parseErr error

	// First try JSON
	if err := json.Unmarshal(data, &spec); err != nil {
		// If JSON fails, try YAML
		if yamlErr := yaml.Unmarshal(data, &spec); yamlErr != nil {
			// Both failed, return more descriptive error
			return nil, &ParseError{
				Message: fmt.Sprintf("failed to parse as JSON (%v) or YAML (%v)", err, yamlErr),
			}
		}
	}

	// Validate the parsed specification
	if err := p.validateSpec(&spec); err != nil {
		return nil, err
	}

	// Validate AsyncAPI version
	if err := p.ValidateVersion(spec.AsyncAPI); err != nil {
		return nil, err
	}

	// Extract messages from the specification
	messages, errors := p.extractMessages(&spec)

	result := &ParseResult{
		Spec:     &spec,
		Messages: messages,
		Errors:   errors,
	}

	return result, parseErr
}

// ValidateVersion validates if the AsyncAPI version is supported
func (p *AsyncAPIParser) ValidateVersion(version string) error {
	if version == "" {
		return &ValidationError{
			Field:   "asyncapi",
			Message: "AsyncAPI version is required",
		}
	}

	// Normalize version (remove any 'v' prefix)
	normalizedVersion := strings.TrimPrefix(version, "v")

	if slices.Contains(p.supportedVersions, normalizedVersion) {
		return nil
	}

	return &UnsupportedVersionError{
		Version:           version,
		SupportedVersions: p.supportedVersions,
	}
}

// validateSpec performs basic validation on the parsed specification
func (p *AsyncAPIParser) validateSpec(spec *AsyncAPISpec) error {
	if spec == nil {
		return &ValidationError{Message: "specification is nil"}
	}

	// Validate required fields
	if spec.AsyncAPI == "" {
		return &ValidationError{
			Field:   "asyncapi",
			Message: "AsyncAPI version is required",
		}
	}

	if spec.Info.Title == "" {
		return &ValidationError{
			Field:   "info.title",
			Message: "info.title is required",
		}
	}

	if spec.Info.Version == "" {
		return &ValidationError{
			Field:   "info.version",
			Message: "info.version is required",
		}
	}

	return nil
}

// extractMessages extracts message schemas from various parts of the AsyncAPI spec
func (p *AsyncAPIParser) extractMessages(spec *AsyncAPISpec) (map[string]*MessageSchema, []error) {
	messages := make(map[string]*MessageSchema)
	var errors []error

	// Extract messages from components
	if spec.Components != nil {
		// Extract from components.messages
		if spec.Components.Messages != nil {
			for name, message := range spec.Components.Messages {
				if message.Payload != nil {
					// Set the name for internal use
					message.Payload.Name = name
					messages[name] = message.Payload
				}
			}
		}

		// Extract schemas from components.schemas
		if spec.Components.Schemas != nil {
			for name, schema := range spec.Components.Schemas {
				schema.Name = name
				messages[name] = schema
			}
		}
	}

	// Extract messages from channel operations
	if spec.Channels != nil {
		for channelName, channel := range spec.Channels {
			// Extract from subscribe operations
			if channel.Subscribe != nil && channel.Subscribe.Message != nil {
				messageName := p.getMessageName(channel.Subscribe.Message, channelName, "subscribe")
				if channel.Subscribe.Message.Payload != nil {
					channel.Subscribe.Message.Payload.Name = messageName
					messages[messageName] = channel.Subscribe.Message.Payload
				}
			}

			// Extract from publish operations
			if channel.Publish != nil && channel.Publish.Message != nil {
				messageName := p.getMessageName(channel.Publish.Message, channelName, "publish")
				if channel.Publish.Message.Payload != nil {
					channel.Publish.Message.Payload.Name = messageName
					messages[messageName] = channel.Publish.Message.Payload
				}
			}
		}
	}

	return messages, errors
}

// getMessageName generates a name for a message based on context
func (p *AsyncAPIParser) getMessageName(message *Message, channelName, operation string) string {
	if message.Name != "" {
		return message.Name
	}
	if message.Title != "" {
		return p.sanitizeName(message.Title)
	}

	// Generate name from channel and operation
	channelParts := strings.Split(channelName, "/")
	var nameParts []string
	for _, part := range channelParts {
		// Skip parameter placeholders
		if !strings.HasPrefix(part, "{") {
			nameParts = append(nameParts, p.sanitizeName(part))
		}
	}

	if len(nameParts) > 0 {
		return strings.Join(nameParts, "") + p.titleCase(operation)
	}

	return "Message" + p.titleCase(operation)
}

// sanitizeName converts a string to a valid Go identifier
func (p *AsyncAPIParser) sanitizeName(name string) string {
	// Remove non-alphanumeric characters and convert to PascalCase
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	parts := reg.Split(name, -1)

	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			// Split on digit/letter boundaries to handle cases like "123invalid"
			subParts := regexp.MustCompile(`(\d+|\D+)`).FindAllString(part, -1)
			for _, subPart := range subParts {
				if len(subPart) > 0 {
					// Capitalize first letter and make rest lowercase
					if len(subPart) == 1 {
						result.WriteString(strings.ToUpper(subPart))
					} else {
						result.WriteString(strings.ToUpper(string(subPart[0])) + strings.ToLower(subPart[1:]))
					}
				}
			}
		}
	}

	sanitized := result.String()
	if sanitized == "" {
		return "Message"
	}

	// Ensure it starts with a letter
	if len(sanitized) > 0 && sanitized[0] >= '0' && sanitized[0] <= '9' {
		sanitized = "Message" + sanitized
	}

	return sanitized
}

// titleCase converts the first character of a string to uppercase
func (p *AsyncAPIParser) titleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(string(s[0])) + strings.ToLower(s[1:])
}

// GetSupportedVersions returns the list of supported AsyncAPI versions
func (p *AsyncAPIParser) GetSupportedVersions() []string {
	return append([]string{}, p.supportedVersions...)
}
