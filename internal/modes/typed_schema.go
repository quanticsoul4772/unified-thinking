// Package modes - Type-safe JSON schema builder for tool definitions
package modes

// JSONSchema represents a type-safe JSON schema structure.
// This replaces map[string]interface{} usage for schema definitions
// while maintaining Claude API compatibility.
type JSONSchema struct {
	Type                 string                    `json:"type"`
	Description          string                    `json:"description,omitempty"`
	Properties           map[string]PropertySchema `json:"properties,omitempty"`
	Required             []string                  `json:"required,omitempty"`
	AdditionalProperties *bool                     `json:"additionalProperties,omitempty"`
}

// PropertySchema defines a property within a JSON schema
type PropertySchema struct {
	Type        string          `json:"type"`
	Description string          `json:"description,omitempty"`
	Default     any             `json:"default,omitempty"`
	Enum        []string        `json:"enum,omitempty"`
	Minimum     *float64        `json:"minimum,omitempty"`
	Maximum     *float64        `json:"maximum,omitempty"`
	MinItems    *int            `json:"minItems,omitempty"`
	MaxItems    *int            `json:"maxItems,omitempty"`
	Items       *PropertySchema `json:"items,omitempty"`
}

// SchemaBuilder provides a fluent API for constructing JSON schemas
type SchemaBuilder struct {
	schema JSONSchema
}

// NewSchemaBuilder creates a new schema builder
func NewSchemaBuilder(description string) *SchemaBuilder {
	return &SchemaBuilder{
		schema: JSONSchema{
			Type:        "object",
			Description: description,
			Properties:  make(map[string]PropertySchema),
			Required:    []string{},
		},
	}
}

// AddString adds a string property
func (b *SchemaBuilder) AddString(name, description string, required bool) *SchemaBuilder {
	b.schema.Properties[name] = PropertySchema{
		Type:        "string",
		Description: description,
	}
	if required {
		b.schema.Required = append(b.schema.Required, name)
	}
	return b
}

// AddStringWithDefault adds a string property with a default value
func (b *SchemaBuilder) AddStringWithDefault(name, description, defaultValue string) *SchemaBuilder {
	b.schema.Properties[name] = PropertySchema{
		Type:        "string",
		Description: description,
		Default:     defaultValue,
	}
	return b
}

// AddStringEnum adds a string property with enumerated values
func (b *SchemaBuilder) AddStringEnum(name, description string, enum []string, required bool) *SchemaBuilder {
	b.schema.Properties[name] = PropertySchema{
		Type:        "string",
		Description: description,
		Enum:        enum,
	}
	if required {
		b.schema.Required = append(b.schema.Required, name)
	}
	return b
}

// AddNumber adds a number property
func (b *SchemaBuilder) AddNumber(name, description string, required bool) *SchemaBuilder {
	b.schema.Properties[name] = PropertySchema{
		Type:        "number",
		Description: description,
	}
	if required {
		b.schema.Required = append(b.schema.Required, name)
	}
	return b
}

// AddNumberWithRange adds a number property with min/max bounds
func (b *SchemaBuilder) AddNumberWithRange(name, description string, min, max float64, required bool) *SchemaBuilder {
	b.schema.Properties[name] = PropertySchema{
		Type:        "number",
		Description: description,
		Minimum:     &min,
		Maximum:     &max,
	}
	if required {
		b.schema.Required = append(b.schema.Required, name)
	}
	return b
}

// AddNumberWithDefault adds a number property with a default value
func (b *SchemaBuilder) AddNumberWithDefault(name, description string, defaultValue float64) *SchemaBuilder {
	b.schema.Properties[name] = PropertySchema{
		Type:        "number",
		Description: description,
		Default:     defaultValue,
	}
	return b
}

// AddInteger adds an integer property
func (b *SchemaBuilder) AddInteger(name, description string, required bool) *SchemaBuilder {
	b.schema.Properties[name] = PropertySchema{
		Type:        "integer",
		Description: description,
	}
	if required {
		b.schema.Required = append(b.schema.Required, name)
	}
	return b
}

// AddIntegerWithDefault adds an integer property with a default value
func (b *SchemaBuilder) AddIntegerWithDefault(name, description string, defaultValue int) *SchemaBuilder {
	b.schema.Properties[name] = PropertySchema{
		Type:        "integer",
		Description: description,
		Default:     defaultValue,
	}
	return b
}

// AddBoolean adds a boolean property
func (b *SchemaBuilder) AddBoolean(name, description string, required bool) *SchemaBuilder {
	b.schema.Properties[name] = PropertySchema{
		Type:        "boolean",
		Description: description,
	}
	if required {
		b.schema.Required = append(b.schema.Required, name)
	}
	return b
}

// AddBooleanWithDefault adds a boolean property with a default value
func (b *SchemaBuilder) AddBooleanWithDefault(name, description string, defaultValue bool) *SchemaBuilder {
	b.schema.Properties[name] = PropertySchema{
		Type:        "boolean",
		Description: description,
		Default:     defaultValue,
	}
	return b
}

// AddStringArray adds a string array property
func (b *SchemaBuilder) AddStringArray(name, description string, required bool) *SchemaBuilder {
	b.schema.Properties[name] = PropertySchema{
		Type:        "array",
		Description: description,
		Items: &PropertySchema{
			Type: "string",
		},
	}
	if required {
		b.schema.Required = append(b.schema.Required, name)
	}
	return b
}

// AddStringArrayWithBounds adds a string array property with size bounds
func (b *SchemaBuilder) AddStringArrayWithBounds(name, description string, minItems, maxItems int, required bool) *SchemaBuilder {
	b.schema.Properties[name] = PropertySchema{
		Type:        "array",
		Description: description,
		MinItems:    &minItems,
		MaxItems:    &maxItems,
		Items: &PropertySchema{
			Type: "string",
		},
	}
	if required {
		b.schema.Required = append(b.schema.Required, name)
	}
	return b
}

// NoAdditionalProperties sets additionalProperties to false (strict mode)
func (b *SchemaBuilder) NoAdditionalProperties() *SchemaBuilder {
	f := false
	b.schema.AdditionalProperties = &f
	return b
}

// Build returns the constructed schema
func (b *SchemaBuilder) Build() JSONSchema {
	return b.schema
}

// ToMap converts the schema to map[string]interface{} for API compatibility
// This is the bridge to the existing code that uses map[string]interface{}
func (b *SchemaBuilder) ToMap() map[string]interface{} {
	props := make(map[string]interface{})
	for name, prop := range b.schema.Properties {
		propMap := map[string]interface{}{
			"type":        prop.Type,
			"description": prop.Description,
		}
		if prop.Default != nil {
			propMap["default"] = prop.Default
		}
		if prop.Enum != nil {
			propMap["enum"] = prop.Enum
		}
		if prop.Minimum != nil {
			propMap["minimum"] = *prop.Minimum
		}
		if prop.Maximum != nil {
			propMap["maximum"] = *prop.Maximum
		}
		if prop.MinItems != nil {
			propMap["minItems"] = *prop.MinItems
		}
		if prop.MaxItems != nil {
			propMap["maxItems"] = *prop.MaxItems
		}
		if prop.Items != nil {
			propMap["items"] = map[string]string{"type": prop.Items.Type}
		}
		props[name] = propMap
	}

	result := map[string]interface{}{
		"type":       b.schema.Type,
		"properties": props,
		"required":   b.schema.Required,
	}
	if b.schema.Description != "" {
		result["description"] = b.schema.Description
	}
	if b.schema.AdditionalProperties != nil {
		result["additionalProperties"] = *b.schema.AdditionalProperties
	}
	return result
}

// TypedToolSchemas provides type-safe schema definitions using the builder
var TypedToolSchemas = struct {
	Think                  map[string]interface{}
	SearchSimilarThoughts  map[string]interface{}
	BuildCausalGraph       map[string]interface{}
	GenerateHypotheses     map[string]interface{}
	AnalyzePerspectives    map[string]interface{}
	DecomposeProblem       map[string]interface{}
	MakeDecision           map[string]interface{}
	DetectBiases           map[string]interface{}
	DetectFallacies        map[string]interface{}
	AssessEvidence         map[string]interface{}
	ProbabilisticReasoning map[string]interface{}
	SynthesizeInsights     map[string]interface{}
}{
	Think: NewSchemaBuilder("Process a thought with structured reasoning").
		AddString("content", "The thought content to process", true).
		AddStringWithDefault("mode", "Thinking mode: linear, tree, divergent, auto", "auto").
		ToMap(),

	SearchSimilarThoughts: NewSchemaBuilder("Find semantically similar thoughts").
		AddString("query", "Search query text", true).
		AddIntegerWithDefault("limit", "Maximum results", 10).
		ToMap(),

	BuildCausalGraph: NewSchemaBuilder("Build a causal graph from observations").
		AddStringArray("observations", "List of observations", true).
		AddString("context", "Context for causal analysis", false).
		ToMap(),

	GenerateHypotheses: NewSchemaBuilder("Generate explanatory hypotheses for observations").
		AddStringArray("observations", "Observations to explain", true).
		AddStringArray("constraints", "Constraints on hypotheses", false).
		ToMap(),

	AnalyzePerspectives: NewSchemaBuilder("Analyze a situation from multiple perspectives").
		AddString("situation", "The situation to analyze", true).
		AddStringArray("stakeholders", "Stakeholder perspectives to consider", false).
		ToMap(),

	DecomposeProblem: NewSchemaBuilder("Decompose a complex problem into subproblems").
		AddString("problem", "The problem to decompose", true).
		AddIntegerWithDefault("max_depth", "Maximum decomposition depth", 3).
		ToMap(),

	MakeDecision: NewSchemaBuilder("Analyze a decision with options and criteria").
		AddString("decision", "The decision to make", true).
		AddStringArray("options", "Available options", true).
		AddStringArray("criteria", "Evaluation criteria", false).
		AddStringArray("constraints", "Decision constraints", false).
		ToMap(),

	DetectBiases: NewSchemaBuilder("Detect cognitive biases in reasoning").
		AddString("content", "Content to analyze for biases", true).
		AddString("context", "Context of the reasoning", false).
		ToMap(),

	DetectFallacies: NewSchemaBuilder("Detect logical fallacies in arguments").
		AddString("argument", "The argument to analyze", true).
		AddString("context", "Context of the argument", false).
		ToMap(),

	AssessEvidence: NewSchemaBuilder("Assess the quality and reliability of evidence").
		AddString("claim", "The claim being supported", true).
		AddStringArray("evidence", "Evidence items to assess", true).
		ToMap(),

	ProbabilisticReasoning: NewSchemaBuilder("Perform probabilistic reasoning with beliefs").
		AddString("hypothesis", "The hypothesis to evaluate", true).
		AddStringArray("evidence", "Evidence items", true).
		AddNumberWithDefault("prior", "Prior probability", 0.5).
		ToMap(),

	SynthesizeInsights: NewSchemaBuilder("Synthesize insights across multiple analyses").
		AddStringArray("inputs", "Analysis inputs to synthesize", true).
		AddString("context", "Context for synthesis", false).
		ToMap(),
}
