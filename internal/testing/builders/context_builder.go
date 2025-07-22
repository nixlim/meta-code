package builders

import (
	"time"
)

// ContextBuilder provides a fluent interface for building test contexts
type ContextBuilder struct {
	id       string
	name     string
	ctype    string
	metadata map[string]interface{}
	config   map[string]interface{}
	created  time.Time
}

// NewContext creates a new ContextBuilder with default values
func NewContext() *ContextBuilder {
	return &ContextBuilder{
		id:       "ctx-test-" + generateID(),
		ctype:    "test",
		metadata: make(map[string]interface{}),
		config:   make(map[string]interface{}),
		created:  time.Now(),
	}
}

// WithID sets the context ID
func (b *ContextBuilder) WithID(id string) *ContextBuilder {
	b.id = id
	return b
}

// WithName sets the context name
func (b *ContextBuilder) WithName(name string) *ContextBuilder {
	b.name = name
	return b
}

// WithType sets the context type
func (b *ContextBuilder) WithType(ctype string) *ContextBuilder {
	b.ctype = ctype
	return b
}

// WithMetadata adds metadata to the context
func (b *ContextBuilder) WithMetadata(key string, value interface{}) *ContextBuilder {
	b.metadata[key] = value
	return b
}

// WithConfig adds configuration to the context
func (b *ContextBuilder) WithConfig(key string, value interface{}) *ContextBuilder {
	b.config[key] = value
	return b
}

// WithCreatedTime sets the creation time
func (b *ContextBuilder) WithCreatedTime(t time.Time) *ContextBuilder {
	b.created = t
	return b
}

// Build constructs the final context object
func (b *ContextBuilder) Build() map[string]interface{} {
	return map[string]interface{}{
		"id":       b.id,
		"name":     b.name,
		"type":     b.ctype,
		"metadata": b.metadata,
		"config":   b.config,
		"created":  b.created.Format(time.RFC3339),
	}
}

// generateID generates a simple ID for testing
func generateID() string {
	return time.Now().Format("20060102150405")
}
