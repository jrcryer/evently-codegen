package generator

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestAsyncAPISpecUnmarshal(t *testing.T) {
	jsonData := `{
		"asyncapi": "2.6.0",
		"id": "urn:example:com:smartylighting:streetlights:server",
		"info": {
			"title": "Streetlights API",
			"version": "1.0.0",
			"description": "The Smartylighting Streetlights API allows you to remotely manage the city lights."
		},
		"servers": {
			"production": {
				"url": "api.streetlights.smartylighting.com:{port}",
				"protocol": "mqtt",
				"description": "Test broker"
			}
		},
		"channels": {
			"smartylighting/streetlights/1/0/event/{streetlightId}/lighting/measured": {
				"description": "The topic on which measured values may be produced and consumed.",
				"parameters": {
					"streetlightId": {
						"description": "The ID of the streetlight."
					}
				}
			}
		}
	}`

	var spec AsyncAPISpec
	err := json.Unmarshal([]byte(jsonData), &spec)
	if err != nil {
		t.Fatalf("Failed to unmarshal AsyncAPI spec: %v", err)
	}

	// Validate basic fields
	if spec.AsyncAPI != "2.6.0" {
		t.Errorf("Expected AsyncAPI version '2.6.0', got '%s'", spec.AsyncAPI)
	}

	if spec.Info.Title != "Streetlights API" {
		t.Errorf("Expected title 'Streetlights API', got '%s'", spec.Info.Title)
	}

	if spec.Info.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", spec.Info.Version)
	}

	// Validate servers
	if len(spec.Servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(spec.Servers))
	}

	production, exists := spec.Servers["production"]
	if !exists {
		t.Error("Expected 'production' server to exist")
	} else {
		if production.Protocol != "mqtt" {
			t.Errorf("Expected protocol 'mqtt', got '%s'", production.Protocol)
		}
	}

	// Validate channels
	if len(spec.Channels) != 1 {
		t.Errorf("Expected 1 channel, got %d", len(spec.Channels))
	}
}

func TestAsyncAPISpecYAMLUnmarshal(t *testing.T) {
	yamlData := `
asyncapi: '2.6.0'
info:
  title: Account Service
  version: 1.0.0
  description: This service is in charge of processing user signups
channels:
  user/signedup:
    subscribe:
      message:
        $ref: '#/components/messages/UserSignedUp'
components:
  messages:
    UserSignedUp:
      payload:
        type: object
        properties:
          displayName:
            type: string
            description: Name of the user
          email:
            type: string
            format: email
            description: Email of the user
`

	var spec AsyncAPISpec
	err := yaml.Unmarshal([]byte(yamlData), &spec)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML AsyncAPI spec: %v", err)
	}

	if spec.AsyncAPI != "2.6.0" {
		t.Errorf("Expected AsyncAPI version '2.6.0', got '%s'", spec.AsyncAPI)
	}

	if spec.Info.Title != "Account Service" {
		t.Errorf("Expected title 'Account Service', got '%s'", spec.Info.Title)
	}

	// Validate components
	if spec.Components == nil {
		t.Fatal("Expected components to be present")
	}

	if len(spec.Components.Messages) != 1 {
		t.Errorf("Expected 1 message in components, got %d", len(spec.Components.Messages))
	}

	userSignedUp, exists := spec.Components.Messages["UserSignedUp"]
	if !exists {
		t.Error("Expected 'UserSignedUp' message to exist")
	} else {
		if userSignedUp.Payload == nil {
			t.Error("Expected payload to be present")
		} else {
			if userSignedUp.Payload.Type != "object" {
				t.Errorf("Expected payload type 'object', got '%s'", userSignedUp.Payload.Type)
			}
		}
	}
}

func TestMessageSchemaUnmarshal(t *testing.T) {
	jsonData := `{
		"type": "object",
		"title": "User",
		"description": "A user object",
		"required": ["id", "email"],
		"properties": {
			"id": {
				"type": "integer",
				"format": "int64",
				"description": "Unique identifier"
			},
			"email": {
				"type": "string",
				"format": "email",
				"description": "User email address"
			},
			"name": {
				"type": "string",
				"description": "User display name"
			},
			"age": {
				"type": "integer",
				"minimum": 0,
				"maximum": 150
			}
		}
	}`

	var schema MessageSchema
	err := json.Unmarshal([]byte(jsonData), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal message schema: %v", err)
	}

	// Validate basic properties
	if schema.Type != "object" {
		t.Errorf("Expected type 'object', got '%s'", schema.Type)
	}

	if schema.Title != "User" {
		t.Errorf("Expected title 'User', got '%s'", schema.Title)
	}

	if len(schema.Required) != 2 {
		t.Errorf("Expected 2 required fields, got %d", len(schema.Required))
	}

	// Validate properties
	if len(schema.Properties) != 4 {
		t.Errorf("Expected 4 properties, got %d", len(schema.Properties))
	}

	// Test ID property
	idProp, exists := schema.Properties["id"]
	if !exists {
		t.Error("Expected 'id' property to exist")
	} else {
		if idProp.Type != "integer" {
			t.Errorf("Expected id type 'integer', got '%s'", idProp.Type)
		}
		if idProp.Format != "int64" {
			t.Errorf("Expected id format 'int64', got '%s'", idProp.Format)
		}
	}

	// Test age property with validation
	ageProp, exists := schema.Properties["age"]
	if !exists {
		t.Error("Expected 'age' property to exist")
	} else {
		if ageProp.Minimum == nil || *ageProp.Minimum != 0 {
			t.Error("Expected age minimum to be 0")
		}
		if ageProp.Maximum == nil || *ageProp.Maximum != 150 {
			t.Error("Expected age maximum to be 150")
		}
	}
}

func TestPropertyValidation(t *testing.T) {
	jsonData := `{
		"type": "string",
		"minLength": 1,
		"maxLength": 100,
		"pattern": "^[a-zA-Z0-9]+$",
		"enum": ["active", "inactive", "pending"]
	}`

	var prop Property
	err := json.Unmarshal([]byte(jsonData), &prop)
	if err != nil {
		t.Fatalf("Failed to unmarshal property: %v", err)
	}

	if prop.Type != "string" {
		t.Errorf("Expected type 'string', got '%s'", prop.Type)
	}

	if prop.MinLength == nil || *prop.MinLength != 1 {
		t.Error("Expected minLength to be 1")
	}

	if prop.MaxLength == nil || *prop.MaxLength != 100 {
		t.Error("Expected maxLength to be 100")
	}

	if prop.Pattern != "^[a-zA-Z0-9]+$" {
		t.Errorf("Expected pattern '^[a-zA-Z0-9]+$', got '%s'", prop.Pattern)
	}

	if len(prop.Enum) != 3 {
		t.Errorf("Expected 3 enum values, got %d", len(prop.Enum))
	}
}

func TestArrayProperty(t *testing.T) {
	jsonData := `{
		"type": "array",
		"items": {
			"type": "string"
		},
		"minItems": 1,
		"maxItems": 10,
		"uniqueItems": true
	}`

	var prop Property
	err := json.Unmarshal([]byte(jsonData), &prop)
	if err != nil {
		t.Fatalf("Failed to unmarshal array property: %v", err)
	}

	if prop.Type != "array" {
		t.Errorf("Expected type 'array', got '%s'", prop.Type)
	}

	if prop.Items == nil {
		t.Fatal("Expected items to be present")
	}

	if prop.Items.Type != "string" {
		t.Errorf("Expected items type 'string', got '%s'", prop.Items.Type)
	}

	if prop.MinItems == nil || *prop.MinItems != 1 {
		t.Error("Expected minItems to be 1")
	}

	if prop.MaxItems == nil || *prop.MaxItems != 10 {
		t.Error("Expected maxItems to be 10")
	}

	if prop.UniqueItems == nil || !*prop.UniqueItems {
		t.Error("Expected uniqueItems to be true")
	}
}

func TestNestedObjectProperty(t *testing.T) {
	jsonData := `{
		"type": "object",
		"properties": {
			"address": {
				"type": "object",
				"properties": {
					"street": {
						"type": "string"
					},
					"city": {
						"type": "string"
					}
				},
				"required": ["street"]
			}
		}
	}`

	var prop Property
	err := json.Unmarshal([]byte(jsonData), &prop)
	if err != nil {
		t.Fatalf("Failed to unmarshal nested object property: %v", err)
	}

	if prop.Type != "object" {
		t.Errorf("Expected type 'object', got '%s'", prop.Type)
	}

	addressProp, exists := prop.Properties["address"]
	if !exists {
		t.Fatal("Expected 'address' property to exist")
	}

	if addressProp.Type != "object" {
		t.Errorf("Expected address type 'object', got '%s'", addressProp.Type)
	}

	if len(addressProp.Properties) != 2 {
		t.Errorf("Expected 2 address properties, got %d", len(addressProp.Properties))
	}

	if len(addressProp.Required) != 1 {
		t.Errorf("Expected 1 required field in address, got %d", len(addressProp.Required))
	}
}

func TestSchemaComposition(t *testing.T) {
	jsonData := `{
		"allOf": [
			{
				"type": "object",
				"properties": {
					"name": {
						"type": "string"
					}
				}
			},
			{
				"type": "object",
				"properties": {
					"age": {
						"type": "integer"
					}
				}
			}
		]
	}`

	var schema MessageSchema
	err := json.Unmarshal([]byte(jsonData), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal composition schema: %v", err)
	}

	if len(schema.AllOf) != 2 {
		t.Errorf("Expected 2 allOf schemas, got %d", len(schema.AllOf))
	}

	// Validate first schema in allOf
	if schema.AllOf[0].Type != "object" {
		t.Errorf("Expected first allOf type 'object', got '%s'", schema.AllOf[0].Type)
	}

	if len(schema.AllOf[0].Properties) != 1 {
		t.Errorf("Expected 1 property in first allOf, got %d", len(schema.AllOf[0].Properties))
	}
}

func TestReferenceHandling(t *testing.T) {
	jsonData := `{
		"$ref": "#/components/schemas/User"
	}`

	var schema MessageSchema
	err := json.Unmarshal([]byte(jsonData), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal reference schema: %v", err)
	}

	if schema.Ref != "#/components/schemas/User" {
		t.Errorf("Expected ref '#/components/schemas/User', got '%s'", schema.Ref)
	}
}

func TestConfigDefaults(t *testing.T) {
	config := &Config{}

	// Test that zero values are properly handled
	if config.PackageName != "" {
		t.Errorf("Expected empty package name, got '%s'", config.PackageName)
	}

	if config.OutputDir != "" {
		t.Errorf("Expected empty output dir, got '%s'", config.OutputDir)
	}

	// Test with values
	config.PackageName = "mypackage"
	config.OutputDir = "./output"
	config.IncludeComments = true
	config.UsePointers = false

	if config.PackageName != "mypackage" {
		t.Errorf("Expected package name 'mypackage', got '%s'", config.PackageName)
	}

	if !config.IncludeComments {
		t.Error("Expected IncludeComments to be true")
	}

	if config.UsePointers {
		t.Error("Expected UsePointers to be false")
	}
}
