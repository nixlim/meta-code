package builders

import (
	"encoding/json"
)

// RequestBuilder provides a fluent interface for building test requests
type RequestBuilder struct {
	jsonrpc string
	method  string
	params  map[string]interface{}
	id      interface{}
}

// NewRequest creates a new RequestBuilder with default values
func NewRequest() *RequestBuilder {
	return &RequestBuilder{
		jsonrpc: "2.0",
		params:  make(map[string]interface{}),
	}
}

// WithMethod sets the method for the request
func (b *RequestBuilder) WithMethod(method string) *RequestBuilder {
	b.method = method
	return b
}

// WithParam adds a parameter to the request
func (b *RequestBuilder) WithParam(key string, value interface{}) *RequestBuilder {
	b.params[key] = value
	return b
}

// WithParams sets all parameters at once
func (b *RequestBuilder) WithParams(params map[string]interface{}) *RequestBuilder {
	b.params = params
	return b
}

// WithID sets the request ID
func (b *RequestBuilder) WithID(id interface{}) *RequestBuilder {
	b.id = id
	return b
}

// Build constructs the final request object
func (b *RequestBuilder) Build() map[string]interface{} {
	request := map[string]interface{}{
		"jsonrpc": b.jsonrpc,
		"method":  b.method,
		"params":  b.params,
	}

	if b.id != nil {
		request["id"] = b.id
	}

	return request
}

// BuildJSON returns the request as a JSON string
func (b *RequestBuilder) BuildJSON() (string, error) {
	request := b.Build()
	bytes, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// BuildBytes returns the request as JSON bytes
func (b *RequestBuilder) BuildBytes() ([]byte, error) {
	request := b.Build()
	return json.Marshal(request)
}
