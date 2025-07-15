package generator

import (
	"strings"
	"testing"
)

func TestNewAsyncAPIParser(t *testing.T) {
	parser := NewAsyncAPIParser()
	if parser == nil {
		t.Fatal("NewAsyncAPIParser returned nil")
	}

	versions := parser.GetSupportedVersions()
	if len(versions) == 0 {
		t.Error("Expected supported versions to be non-empty")
	}

	// Check that common versions are supported
	expectedVersions := []string{"2.6.0", "3.0.0"}
	for _, expected := range expectedVersions {
		found := false
		for _, version := range versions {
			if version == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected version %s to be supported", expected)
		}
	}
}

func TestParseValidJSON(t *testing.T) {
	validJSON := `{
		"asyncapi": "2.6.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0",
			"description": "A test AsyncAPI specification"
		},
		"channels": {
			"user/signup": {
				"subscribe": {
					"message": {
						"payload": {
							"type": "object",
							"properties": {
								"userId": {
									"type": "string",
									"description": "User identifier"
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

	parser := NewAsyncAPIParser()
	result, err := parser.Parse([]byte(validJSON))

	if err != nil {
		t.Fatalf("Failed to parse valid JSON: %v", err)
	}

	if result == nil {
		t.Fatal("Parse result is nil")
	}

	if result.Spec == nil {
		t.Fatal("Parsed spec is nil")
	}

	// Validate basic fields
	if result.Spec.AsyncAPI != "2.6.0" {
		t.Errorf("Expected AsyncAPI version '2.6.0', got '%s'", result.Spec.AsyncAPI)
	}

	if result.Spec.Info.Title != "Test API" {
		t.Errorf("Expected title 'Test API', got '%s'", result.Spec.Info.Title)
	}

	// Validate extracted messages
	if len(result.Messages) == 0 {
		t.Error("Expected messages to be extracted")
	}

	// Check for the extracted message
	found := false
	for name, schema := range result.Messages {
		if strings.Contains(strings.ToLower(name), "signup") || strings.Contains(strings.ToLower(name), "user") {
			found = true
			if schema.Type != "object" {
				t.Errorf("Expected message type 'object', got '%s'", schema.Type)
			}
			if len(schema.Properties) != 2 {
				t.Errorf("Expected 2 properties, got %d", len(schema.Properties))
			}
		}
	}

	if !found {
		t.Error("Expected to find extracted message from channel")
	}
}

func TestParseValidYAML(t *testing.T) {
	validYAML := `
asyncapi: '2.6.0'
info:
  title: Account Service
  version: 1.0.0
  description: This service handles user accounts
channels:
  user/created:
    publish:
      message:
        payload:
          type: object
          properties:
            id:
              type: integer
              format: int64
            username:
              type: string
              minLength: 3
              maxLength: 50
            createdAt:
              type: string
              format: date-time
          required:
            - id
            - username
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: integer
        name:
          type: string
      required:
        - id
        - name
`

	parser := NewAsyncAPIParser()
	result, err := parser.Parse([]byte(validYAML))

	if err != nil {
		t.Fatalf("Failed to parse valid YAML: %v", err)
	}

	if result.Spec.AsyncAPI != "2.6.0" {
		t.Errorf("Expected AsyncAPI version '2.6.0', got '%s'", result.Spec.AsyncAPI)
	}

	if result.Spec.Info.Title != "Account Service" {
		t.Errorf("Expected title 'Account Service', got '%s'", result.Spec.Info.Title)
	}

	// Should extract messages from both channels and components
	if len(result.Messages) < 2 {
		t.Errorf("Expected at least 2 messages, got %d", len(result.Messages))
	}

	// Check for User schema from components
	userSchema, exists := result.Messages["User"]
	if !exists {
		t.Error("Expected 'User' schema to be extracted from components")
	} else {
		if userSchema.Type != "object" {
			t.Errorf("Expected User type 'object', got '%s'", userSchema.Type)
		}
	}
}

func TestParseInvalidJSON(t *testing.T) {
	invalidJSON := `{
		"asyncapi": "2.6.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0"
		},
		"channels": {
			"invalid": {
				"subscribe": {
					"message": {
						"payload": "unclosed_string
					}
				}
			}
		}
	}`

	parser := NewAsyncAPIParser()
	_, err := parser.Parse([]byte(invalidJSON))

	if err == nil {
		t.Error("Expected error when parsing invalid JSON")
	}

	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Errorf("Expected ParseError, got %T", err)
	} else {
		if !strings.Contains(parseErr.Message, "JSON") && !strings.Contains(parseErr.Message, "YAML") {
			t.Errorf("Expected error message to mention JSON or YAML parsing, got: %s", parseErr.Message)
		}
	}
}

func TestParseInvalidYAML(t *testing.T) {
	invalidYAML := `
asyncapi: '2.6.0'
info:
  title: Test API
  version: 1.0.0
channels:
  test:
    subscribe:
      message:
        payload:
          type: object
          properties:
            - invalid: yaml structure
`

	parser := NewAsyncAPIParser()
	_, err := parser.Parse([]byte(invalidYAML))

	if err == nil {
		t.Error("Expected error when parsing invalid YAML")
	}

	_, ok := err.(*ParseError)
	if !ok {
		t.Errorf("Expected ParseError, got %T", err)
	}
}

func TestParseEmptyInput(t *testing.T) {
	parser := NewAsyncAPIParser()
	_, err := parser.Parse([]byte{})

	if err == nil {
		t.Error("Expected error when parsing empty input")
	}

	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Errorf("Expected ParseError, got %T", err)
	} else {
		if parseErr.Message != "empty input data" {
			t.Errorf("Expected 'empty input data' error, got: %s", parseErr.Message)
		}
	}
}

func TestValidateVersion(t *testing.T) {
	parser := NewAsyncAPIParser()

	// Test valid versions
	validVersions := []string{"2.6.0", "3.0.0", "2.0.0"}
	for _, version := range validVersions {
		err := parser.ValidateVersion(version)
		if err != nil {
			t.Errorf("Expected version %s to be valid, got error: %v", version, err)
		}
	}

	// Test version with 'v' prefix
	err := parser.ValidateVersion("v2.6.0")
	if err != nil {
		t.Errorf("Expected version v2.6.0 to be valid, got error: %v", err)
	}

	// Test invalid versions
	invalidVersions := []string{"1.0.0", "4.0.0", "2.7.0"}
	for _, version := range invalidVersions {
		err := parser.ValidateVersion(version)
		if err == nil {
			t.Errorf("Expected version %s to be invalid", version)
		}

		unsupportedErr, ok := err.(*UnsupportedVersionError)
		if !ok {
			t.Errorf("Expected UnsupportedVersionError for version %s, got %T", version, err)
		} else {
			if unsupportedErr.Version != version {
				t.Errorf("Expected error version %s, got %s", version, unsupportedErr.Version)
			}
		}
	}

	// Test empty version
	err = parser.ValidateVersion("")
	if err == nil {
		t.Error("Expected error for empty version")
	}

	validationErr, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("Expected ValidationError for empty version, got %T", err)
	} else {
		if validationErr.Field != "asyncapi" {
			t.Errorf("Expected field 'asyncapi', got '%s'", validationErr.Field)
		}
	}
}

func TestValidateSpec(t *testing.T) {
	parser := NewAsyncAPIParser()

	// Test valid spec
	validSpec := &AsyncAPISpec{
		AsyncAPI: "2.6.0",
		Info: Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
	}

	err := parser.validateSpec(validSpec)
	if err != nil {
		t.Errorf("Expected valid spec to pass validation, got: %v", err)
	}

	// Test nil spec
	err = parser.validateSpec(nil)
	if err == nil {
		t.Error("Expected error for nil spec")
	}

	// Test missing AsyncAPI version
	invalidSpec := &AsyncAPISpec{
		Info: Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
	}

	err = parser.validateSpec(invalidSpec)
	if err == nil {
		t.Error("Expected error for missing AsyncAPI version")
	}

	// Test missing title
	invalidSpec2 := &AsyncAPISpec{
		AsyncAPI: "2.6.0",
		Info: Info{
			Version: "1.0.0",
		},
	}

	err = parser.validateSpec(invalidSpec2)
	if err == nil {
		t.Error("Expected error for missing title")
	}

	validationErr, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("Expected ValidationError, got %T", err)
	} else {
		if validationErr.Field != "info.title" {
			t.Errorf("Expected field 'info.title', got '%s'", validationErr.Field)
		}
	}
}

func TestExtractMessagesFromComponents(t *testing.T) {
	spec := &AsyncAPISpec{
		Components: &Components{
			Schemas: map[string]*MessageSchema{
				"User": {
					Type: "object",
					Properties: map[string]*Property{
						"id":   {Type: "integer"},
						"name": {Type: "string"},
					},
				},
			},
			Messages: map[string]*Message{
				"UserCreated": {
					Payload: &MessageSchema{
						Type: "object",
						Properties: map[string]*Property{
							"userId": {Type: "string"},
						},
					},
				},
			},
		},
	}

	parser := NewAsyncAPIParser()
	messages, errors := parser.extractMessages(spec)

	if len(errors) > 0 {
		t.Errorf("Unexpected errors: %v", errors)
	}

	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}

	// Check User schema
	userSchema, exists := messages["User"]
	if !exists {
		t.Error("Expected 'User' schema to be extracted")
	} else {
		if userSchema.Name != "User" {
			t.Errorf("Expected schema name 'User', got '%s'", userSchema.Name)
		}
	}

	// Check UserCreated message
	userCreatedSchema, exists := messages["UserCreated"]
	if !exists {
		t.Error("Expected 'UserCreated' message to be extracted")
	} else {
		if userCreatedSchema.Name != "UserCreated" {
			t.Errorf("Expected message name 'UserCreated', got '%s'", userCreatedSchema.Name)
		}
	}
}

func TestSanitizeName(t *testing.T) {
	parser := NewAsyncAPIParser()

	testCases := []struct {
		input    string
		expected string
	}{
		{"user-signup", "UserSignup"},
		{"user_created", "UserCreated"},
		{"user.deleted", "UserDeleted"},
		{"user/updated", "UserUpdated"},
		{"user@domain", "UserDomain"},
		{"123invalid", "Message123Invalid"},
		{"", "Message"},
		{"simple", "Simple"},
		{"UPPER", "Upper"},
		{"mixed-Case_test", "MixedCaseTest"},
	}

	for _, tc := range testCases {
		result := parser.sanitizeName(tc.input)
		if result != tc.expected {
			t.Errorf("sanitizeName(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

func TestGetMessageName(t *testing.T) {
	parser := NewAsyncAPIParser()

	// Test with message name
	message := &Message{Name: "UserSignup"}
	result := parser.getMessageName(message, "user/signup", "subscribe")
	if result != "UserSignup" {
		t.Errorf("Expected 'UserSignup', got '%s'", result)
	}

	// Test with message title
	message2 := &Message{Title: "User Registration"}
	result2 := parser.getMessageName(message2, "user/signup", "subscribe")
	if result2 != "UserRegistration" {
		t.Errorf("Expected 'UserRegistration', got '%s'", result2)
	}

	// Test with channel name
	message3 := &Message{}
	result3 := parser.getMessageName(message3, "user/signup", "subscribe")
	if result3 != "UserSignupSubscribe" {
		t.Errorf("Expected 'UserSignupSubscribe', got '%s'", result3)
	}

	// Test with complex channel name
	message4 := &Message{}
	result4 := parser.getMessageName(message4, "api/v1/user/{id}/profile", "publish")
	if result4 != "ApiV1UserProfilePublish" {
		t.Errorf("Expected 'ApiV1UserProfilePublish', got '%s'", result4)
	}
}

func TestParseWithAsyncAPI3(t *testing.T) {
	asyncAPI3JSON := `{
		"asyncapi": "3.0.0",
		"info": {
			"title": "AsyncAPI 3.0 Test",
			"version": "1.0.0"
		},
		"channels": {
			"userEvents": {
				"messages": {
					"userCreated": {
						"payload": {
							"type": "object",
							"properties": {
								"id": {"type": "string"},
								"email": {"type": "string"}
							}
						}
					}
				}
			}
		}
	}`

	parser := NewAsyncAPIParser()
	result, err := parser.Parse([]byte(asyncAPI3JSON))

	if err != nil {
		t.Fatalf("Failed to parse AsyncAPI 3.0: %v", err)
	}

	if result.Spec.AsyncAPI != "3.0.0" {
		t.Errorf("Expected AsyncAPI version '3.0.0', got '%s'", result.Spec.AsyncAPI)
	}
}

func TestParseUnsupportedVersion(t *testing.T) {
	unsupportedVersionJSON := `{
		"asyncapi": "4.0.0",
		"info": {
			"title": "Unsupported Version Test",
			"version": "1.0.0"
		}
	}`

	parser := NewAsyncAPIParser()
	_, err := parser.Parse([]byte(unsupportedVersionJSON))

	if err == nil {
		t.Error("Expected error for unsupported AsyncAPI version")
	}

	unsupportedErr, ok := err.(*UnsupportedVersionError)
	if !ok {
		t.Errorf("Expected UnsupportedVersionError, got %T", err)
	} else {
		if unsupportedErr.Version != "4.0.0" {
			t.Errorf("Expected version '4.0.0', got '%s'", unsupportedErr.Version)
		}
		if len(unsupportedErr.SupportedVersions) == 0 {
			t.Error("Expected supported versions to be listed")
		}
	}
}

func TestParseMissingRequiredFields(t *testing.T) {
	testCases := []struct {
		name     string
		json     string
		expected string
	}{
		{
			name: "missing info.title",
			json: `{
				"asyncapi": "2.6.0",
				"info": {
					"version": "1.0.0"
				}
			}`,
			expected: "info.title",
		},
		{
			name: "missing info.version",
			json: `{
				"asyncapi": "2.6.0",
				"info": {
					"title": "Test API"
				}
			}`,
			expected: "info.version",
		},
		{
			name: "missing asyncapi version",
			json: `{
				"info": {
					"title": "Test API",
					"version": "1.0.0"
				}
			}`,
			expected: "asyncapi",
		},
	}

	parser := NewAsyncAPIParser()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parser.Parse([]byte(tc.json))

			if err == nil {
				t.Errorf("Expected error for %s", tc.name)
				return
			}

			validationErr, ok := err.(*ValidationError)
			if !ok {
				t.Errorf("Expected ValidationError, got %T", err)
				return
			}

			if validationErr.Field != tc.expected {
				t.Errorf("Expected field '%s', got '%s'", tc.expected, validationErr.Field)
			}
		})
	}
}

func TestParseComplexChannelStructure(t *testing.T) {
	complexJSON := `{
		"asyncapi": "2.6.0",
		"info": {
			"title": "Complex Channel Test",
			"version": "1.0.0"
		},
		"channels": {
			"user/{userId}/notifications": {
				"subscribe": {
					"message": {
						"name": "NotificationReceived",
						"payload": {
							"type": "object",
							"properties": {
								"id": {"type": "string"},
								"message": {"type": "string"},
								"timestamp": {"type": "string", "format": "date-time"}
							},
							"required": ["id", "message"]
						}
					}
				},
				"publish": {
					"message": {
						"title": "Notification Sent",
						"payload": {
							"type": "object",
							"properties": {
								"recipientId": {"type": "string"},
								"content": {"type": "string"}
							}
						}
					}
				}
			}
		}
	}`

	parser := NewAsyncAPIParser()
	result, err := parser.Parse([]byte(complexJSON))

	if err != nil {
		t.Fatalf("Failed to parse complex channel structure: %v", err)
	}

	if len(result.Messages) < 2 {
		t.Errorf("Expected at least 2 messages, got %d", len(result.Messages))
	}

	// Check that messages were extracted with proper names
	foundNotificationReceived := false
	foundNotificationSent := false

	for name, schema := range result.Messages {
		if name == "NotificationReceived" {
			foundNotificationReceived = true
			if len(schema.Properties) != 3 {
				t.Errorf("Expected 3 properties for NotificationReceived, got %d", len(schema.Properties))
			}
		}
		if strings.Contains(name, "NotificationSent") || strings.Contains(name, "Notification") {
			foundNotificationSent = true
		}
	}

	if !foundNotificationReceived {
		t.Error("Expected to find NotificationReceived message")
	}
	if !foundNotificationSent {
		t.Error("Expected to find notification sent message")
	}
}

func TestParseWithReferences(t *testing.T) {
	jsonWithRefs := `{
		"asyncapi": "2.6.0",
		"info": {
			"title": "Reference Test",
			"version": "1.0.0"
		},
		"channels": {
			"user/events": {
				"subscribe": {
					"message": {
						"$ref": "#/components/messages/UserEvent"
					}
				}
			}
		},
		"components": {
			"messages": {
				"UserEvent": {
					"payload": {
						"$ref": "#/components/schemas/User"
					}
				}
			},
			"schemas": {
				"User": {
					"type": "object",
					"properties": {
						"id": {"type": "string"},
						"name": {"type": "string"}
					},
					"required": ["id"]
				}
			}
		}
	}`

	parser := NewAsyncAPIParser()
	result, err := parser.Parse([]byte(jsonWithRefs))

	if err != nil {
		t.Fatalf("Failed to parse JSON with references: %v", err)
	}

	// Should extract User schema from components
	userSchema, exists := result.Messages["User"]
	if !exists {
		t.Error("Expected 'User' schema to be extracted from components")
	} else {
		if userSchema.Type != "object" {
			t.Errorf("Expected User type 'object', got '%s'", userSchema.Type)
		}
		if len(userSchema.Properties) != 2 {
			t.Errorf("Expected 2 properties, got %d", len(userSchema.Properties))
		}
	}
}

func TestTitleCaseHelper(t *testing.T) {
	parser := NewAsyncAPIParser()

	testCases := []struct {
		input    string
		expected string
	}{
		{"subscribe", "Subscribe"},
		{"publish", "Publish"},
		{"UPPERCASE", "Uppercase"},
		{"lowercase", "Lowercase"},
		{"", ""},
		{"a", "A"},
		{"mixedCase", "Mixedcase"},
	}

	for _, tc := range testCases {
		result := parser.titleCase(tc.input)
		if result != tc.expected {
			t.Errorf("titleCase(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}
