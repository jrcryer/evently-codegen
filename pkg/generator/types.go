package generator

// Config holds generation configuration options
type Config struct {
	PackageName     string
	OutputDir       string
	IncludeComments bool
	UsePointers     bool
}

// ParseResult contains the parsed AsyncAPI specification
type ParseResult struct {
	Spec     *AsyncAPISpec
	Messages map[string]*MessageSchema
	Errors   []error
}

// GenerateResult contains the generated Go code
type GenerateResult struct {
	Files  map[string]string // filename -> content
	Errors []error
}

// AsyncAPISpec represents a parsed AsyncAPI specification
type AsyncAPISpec struct {
	AsyncAPI     string                 `json:"asyncapi" yaml:"asyncapi"`
	ID           string                 `json:"id,omitempty" yaml:"id,omitempty"`
	Info         Info                   `json:"info" yaml:"info"`
	Servers      map[string]*Server     `json:"servers,omitempty" yaml:"servers,omitempty"`
	Channels     map[string]*Channel    `json:"channels,omitempty" yaml:"channels,omitempty"`
	Components   *Components            `json:"components,omitempty" yaml:"components,omitempty"`
	Tags         []*Tag                 `json:"tags,omitempty" yaml:"tags,omitempty"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
}

// Info represents AsyncAPI info section
type Info struct {
	Title          string   `json:"title" yaml:"title"`
	Version        string   `json:"version" yaml:"version"`
	Description    string   `json:"description,omitempty" yaml:"description,omitempty"`
	TermsOfService string   `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty" yaml:"contact,omitempty"`
	License        *License `json:"license,omitempty" yaml:"license,omitempty"`
}

// Contact represents contact information
type Contact struct {
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
	URL   string `json:"url,omitempty" yaml:"url,omitempty"`
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
}

// License represents license information
type License struct {
	Name string `json:"name" yaml:"name"`
	URL  string `json:"url,omitempty" yaml:"url,omitempty"`
}

// Server represents an AsyncAPI server
type Server struct {
	URL         string                `json:"url" yaml:"url"`
	Protocol    string                `json:"protocol" yaml:"protocol"`
	Description string                `json:"description,omitempty" yaml:"description,omitempty"`
	Variables   map[string]*Variable  `json:"variables,omitempty" yaml:"variables,omitempty"`
	Security    []map[string][]string `json:"security,omitempty" yaml:"security,omitempty"`
	Tags        []*Tag                `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// Variable represents a server variable
type Variable struct {
	Enum        []string `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default     string   `json:"default,omitempty" yaml:"default,omitempty"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Examples    []string `json:"examples,omitempty" yaml:"examples,omitempty"`
}

// Components represents AsyncAPI components section
type Components struct {
	Schemas           map[string]*MessageSchema  `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	Messages          map[string]*Message        `json:"messages,omitempty" yaml:"messages,omitempty"`
	SecuritySchemes   map[string]*SecurityScheme `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
	Parameters        map[string]*Parameter      `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	CorrelationIDs    map[string]*CorrelationID  `json:"correlationIds,omitempty" yaml:"correlationIds,omitempty"`
	OperationTraits   map[string]*OperationTrait `json:"operationTraits,omitempty" yaml:"operationTraits,omitempty"`
	MessageTraits     map[string]*MessageTrait   `json:"messageTraits,omitempty" yaml:"messageTraits,omitempty"`
	ServerBindings    map[string]interface{}     `json:"serverBindings,omitempty" yaml:"serverBindings,omitempty"`
	ChannelBindings   map[string]interface{}     `json:"channelBindings,omitempty" yaml:"channelBindings,omitempty"`
	OperationBindings map[string]interface{}     `json:"operationBindings,omitempty" yaml:"operationBindings,omitempty"`
	MessageBindings   map[string]interface{}     `json:"messageBindings,omitempty" yaml:"messageBindings,omitempty"`
}

// SecurityScheme represents a security scheme
type SecurityScheme struct {
	Type             string      `json:"type" yaml:"type"`
	Description      string      `json:"description,omitempty" yaml:"description,omitempty"`
	Name             string      `json:"name,omitempty" yaml:"name,omitempty"`
	In               string      `json:"in,omitempty" yaml:"in,omitempty"`
	Scheme           string      `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	BearerFormat     string      `json:"bearerFormat,omitempty" yaml:"bearerFormat,omitempty"`
	Flows            *OAuthFlows `json:"flows,omitempty" yaml:"flows,omitempty"`
	OpenIDConnectURL string      `json:"openIdConnectUrl,omitempty" yaml:"openIdConnectUrl,omitempty"`
}

// OAuthFlows represents OAuth flows
type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty" yaml:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty" yaml:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty" yaml:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty" yaml:"authorizationCode,omitempty"`
}

// OAuthFlow represents an OAuth flow
type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty" yaml:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty" yaml:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty" yaml:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes" yaml:"scopes"`
}

// CorrelationID represents a correlation ID
type CorrelationID struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Location    string `json:"location" yaml:"location"`
}

// OperationTrait represents an operation trait
type OperationTrait struct {
	OperationID  string                 `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Summary      string                 `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description  string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Tags         []*Tag                 `json:"tags,omitempty" yaml:"tags,omitempty"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Bindings     interface{}            `json:"bindings,omitempty" yaml:"bindings,omitempty"`
}

// MessageTrait represents a message trait
type MessageTrait struct {
	Headers       *MessageSchema           `json:"headers,omitempty" yaml:"headers,omitempty"`
	CorrelationID *CorrelationID           `json:"correlationId,omitempty" yaml:"correlationId,omitempty"`
	SchemaFormat  string                   `json:"schemaFormat,omitempty" yaml:"schemaFormat,omitempty"`
	ContentType   string                   `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	Name          string                   `json:"name,omitempty" yaml:"name,omitempty"`
	Title         string                   `json:"title,omitempty" yaml:"title,omitempty"`
	Summary       string                   `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description   string                   `json:"description,omitempty" yaml:"description,omitempty"`
	Tags          []*Tag                   `json:"tags,omitempty" yaml:"tags,omitempty"`
	ExternalDocs  *ExternalDocumentation   `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Bindings      interface{}              `json:"bindings,omitempty" yaml:"bindings,omitempty"`
	Examples      []map[string]interface{} `json:"examples,omitempty" yaml:"examples,omitempty"`
}

// Tag represents a tag
type Tag struct {
	Name         string                 `json:"name" yaml:"name"`
	Description  string                 `json:"description,omitempty" yaml:"description,omitempty"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
}

// ExternalDocumentation represents external documentation
type ExternalDocumentation struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	URL         string `json:"url" yaml:"url"`
}

// Channel represents an AsyncAPI channel
type Channel struct {
	Ref         string                `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Description string                `json:"description,omitempty" yaml:"description,omitempty"`
	Subscribe   *Operation            `json:"subscribe,omitempty" yaml:"subscribe,omitempty"`
	Publish     *Operation            `json:"publish,omitempty" yaml:"publish,omitempty"`
	Parameters  map[string]*Parameter `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Bindings    interface{}           `json:"bindings,omitempty" yaml:"bindings,omitempty"`
}

// Operation represents a channel operation
type Operation struct {
	OperationID  string                 `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Summary      string                 `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description  string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Security     []map[string][]string  `json:"security,omitempty" yaml:"security,omitempty"`
	Tags         []*Tag                 `json:"tags,omitempty" yaml:"tags,omitempty"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Bindings     interface{}            `json:"bindings,omitempty" yaml:"bindings,omitempty"`
	Traits       []*OperationTrait      `json:"traits,omitempty" yaml:"traits,omitempty"`
	Message      *Message               `json:"message,omitempty" yaml:"message,omitempty"`
}

// Message represents an AsyncAPI message
type Message struct {
	Ref           string                   `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Headers       *MessageSchema           `json:"headers,omitempty" yaml:"headers,omitempty"`
	Payload       *MessageSchema           `json:"payload,omitempty" yaml:"payload,omitempty"`
	CorrelationID *CorrelationID           `json:"correlationId,omitempty" yaml:"correlationId,omitempty"`
	SchemaFormat  string                   `json:"schemaFormat,omitempty" yaml:"schemaFormat,omitempty"`
	ContentType   string                   `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	Name          string                   `json:"name,omitempty" yaml:"name,omitempty"`
	Title         string                   `json:"title,omitempty" yaml:"title,omitempty"`
	Summary       string                   `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description   string                   `json:"description,omitempty" yaml:"description,omitempty"`
	Tags          []*Tag                   `json:"tags,omitempty" yaml:"tags,omitempty"`
	ExternalDocs  *ExternalDocumentation   `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Bindings      interface{}              `json:"bindings,omitempty" yaml:"bindings,omitempty"`
	Examples      []map[string]interface{} `json:"examples,omitempty" yaml:"examples,omitempty"`
	Traits        []*MessageTrait          `json:"traits,omitempty" yaml:"traits,omitempty"`
}

// Parameter represents a channel parameter
type Parameter struct {
	Description string    `json:"description,omitempty" yaml:"description,omitempty"`
	Schema      *Property `json:"schema,omitempty" yaml:"schema,omitempty"`
}

// MessageSchema represents an AsyncAPI message schema
type MessageSchema struct {
	// Schema identification
	Ref    string `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	ID     string `json:"$id,omitempty" yaml:"$id,omitempty"`
	Schema string `json:"$schema,omitempty" yaml:"$schema,omitempty"`

	// Basic schema properties
	Title       string        `json:"title,omitempty" yaml:"title,omitempty"`
	Description string        `json:"description,omitempty" yaml:"description,omitempty"`
	Default     interface{}   `json:"default,omitempty" yaml:"default,omitempty"`
	Examples    []interface{} `json:"examples,omitempty" yaml:"examples,omitempty"`

	// Type and validation
	Type  string        `json:"type,omitempty" yaml:"type,omitempty"`
	Enum  []interface{} `json:"enum,omitempty" yaml:"enum,omitempty"`
	Const interface{}   `json:"const,omitempty" yaml:"const,omitempty"`

	// Numeric validation
	MultipleOf       *float64 `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
	Maximum          *float64 `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	ExclusiveMaximum *float64 `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
	Minimum          *float64 `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	ExclusiveMinimum *float64 `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`

	// String validation
	MaxLength *int   `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	MinLength *int   `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	Pattern   string `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	Format    string `json:"format,omitempty" yaml:"format,omitempty"`

	// Array validation
	Items           *Property `json:"items,omitempty" yaml:"items,omitempty"`
	AdditionalItems *Property `json:"additionalItems,omitempty" yaml:"additionalItems,omitempty"`
	MaxItems        *int      `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	MinItems        *int      `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	UniqueItems     *bool     `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`

	// Object validation
	Properties           map[string]*Property `json:"properties,omitempty" yaml:"properties,omitempty"`
	PatternProperties    map[string]*Property `json:"patternProperties,omitempty" yaml:"patternProperties,omitempty"`
	AdditionalProperties interface{}          `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Required             []string             `json:"required,omitempty" yaml:"required,omitempty"`
	PropertyNames        *Property            `json:"propertyNames,omitempty" yaml:"propertyNames,omitempty"`
	MaxProperties        *int                 `json:"maxProperties,omitempty" yaml:"maxProperties,omitempty"`
	MinProperties        *int                 `json:"minProperties,omitempty" yaml:"minProperties,omitempty"`

	// Composition
	AllOf []*Property `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	AnyOf []*Property `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	OneOf []*Property `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	Not   *Property   `json:"not,omitempty" yaml:"not,omitempty"`

	// Conditional
	If   *Property `json:"if,omitempty" yaml:"if,omitempty"`
	Then *Property `json:"then,omitempty" yaml:"then,omitempty"`
	Else *Property `json:"else,omitempty" yaml:"else,omitempty"`

	// Annotations
	ReadOnly  *bool `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	WriteOnly *bool `json:"writeOnly,omitempty" yaml:"writeOnly,omitempty"`

	// Non-standard fields for internal use
	Name string `json:"-" yaml:"-"` // Used internally for struct naming
}

// Property represents a schema property (same structure as MessageSchema for JSON Schema compatibility)
type Property struct {
	// Schema identification
	Ref    string `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	ID     string `json:"$id,omitempty" yaml:"$id,omitempty"`
	Schema string `json:"$schema,omitempty" yaml:"$schema,omitempty"`

	// Basic schema properties
	Title       string        `json:"title,omitempty" yaml:"title,omitempty"`
	Description string        `json:"description,omitempty" yaml:"description,omitempty"`
	Default     interface{}   `json:"default,omitempty" yaml:"default,omitempty"`
	Examples    []interface{} `json:"examples,omitempty" yaml:"examples,omitempty"`

	// Type and validation
	Type  string        `json:"type,omitempty" yaml:"type,omitempty"`
	Enum  []interface{} `json:"enum,omitempty" yaml:"enum,omitempty"`
	Const interface{}   `json:"const,omitempty" yaml:"const,omitempty"`

	// Numeric validation
	MultipleOf       *float64 `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
	Maximum          *float64 `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	ExclusiveMaximum *float64 `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
	Minimum          *float64 `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	ExclusiveMinimum *float64 `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`

	// String validation
	MaxLength *int   `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	MinLength *int   `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	Pattern   string `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	Format    string `json:"format,omitempty" yaml:"format,omitempty"`

	// Array validation
	Items           *Property `json:"items,omitempty" yaml:"items,omitempty"`
	AdditionalItems *Property `json:"additionalItems,omitempty" yaml:"additionalItems,omitempty"`
	MaxItems        *int      `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	MinItems        *int      `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	UniqueItems     *bool     `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`

	// Object validation
	Properties           map[string]*Property `json:"properties,omitempty" yaml:"properties,omitempty"`
	PatternProperties    map[string]*Property `json:"patternProperties,omitempty" yaml:"patternProperties,omitempty"`
	AdditionalProperties interface{}          `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Required             []string             `json:"required,omitempty" yaml:"required,omitempty"`
	PropertyNames        *Property            `json:"propertyNames,omitempty" yaml:"propertyNames,omitempty"`
	MaxProperties        *int                 `json:"maxProperties,omitempty" yaml:"maxProperties,omitempty"`
	MinProperties        *int                 `json:"minProperties,omitempty" yaml:"minProperties,omitempty"`

	// Composition
	AllOf []*Property `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	AnyOf []*Property `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	OneOf []*Property `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	Not   *Property   `json:"not,omitempty" yaml:"not,omitempty"`

	// Conditional
	If   *Property `json:"if,omitempty" yaml:"if,omitempty"`
	Then *Property `json:"then,omitempty" yaml:"then,omitempty"`
	Else *Property `json:"else,omitempty" yaml:"else,omitempty"`

	// Annotations
	ReadOnly  *bool `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	WriteOnly *bool `json:"writeOnly,omitempty" yaml:"writeOnly,omitempty"`
}

// GoStruct represents a generated Go struct
type GoStruct struct {
	Name        string
	PackageName string
	Fields      []*GoField
	Comments    []string
}

// GoField represents a Go struct field
type GoField struct {
	Name     string
	Type     string
	JSONTag  string
	Comment  string
	Optional bool
}
