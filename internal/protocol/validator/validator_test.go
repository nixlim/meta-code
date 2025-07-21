package validator

import (
	"context"
	"encoding/json"
	"testing"
	
	"github.com/xeipuuv/gojsonschema"
)

func TestSchemaValidator_New(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "disabled validator",
			config: Config{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "enabled validator with schema directory",
			config: Config{
				Enabled:      true,
				SchemaDir:    "testdata/schemas",
				CacheSchemas: true,
			},
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, err := New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && validator == nil {
				t.Error("New() returned nil validator without error")
			}
			if err == nil && validator.IsEnabled() != tt.config.Enabled {
				t.Errorf("IsEnabled() = %v, want %v", validator.IsEnabled(), tt.config.Enabled)
			}
		})
	}
}

func TestSchemaValidator_ValidateMessage_Disabled(t *testing.T) {
	validator := &SchemaValidator{
		enabled: false,
		schemas: make(map[string]*gojsonschema.Schema),
	}
	
	ctx := context.Background()
	message := json.RawMessage(`{"jsonrpc": "2.0", "method": "test"}`)
	
	err := validator.ValidateMessage(ctx, "request", message)
	if err != nil {
		t.Errorf("ValidateMessage() with disabled validator should return nil, got %v", err)
	}
}

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  ValidationError
		want string
	}{
		{
			name: "with field",
			err: ValidationError{
				Field:   "params.name",
				Message: "required field missing",
			},
			want: "validation error at field 'params.name': required field missing",
		},
		{
			name: "without field",
			err: ValidationError{
				Message: "invalid message format",
			},
			want: "invalid message format",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("ValidationError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultiValidationError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  MultiValidationError
		want string
	}{
		{
			name: "single error",
			err: MultiValidationError{
				Errors: []ValidationError{
					{Message: "invalid format"},
				},
			},
			want: "invalid format",
		},
		{
			name: "multiple errors",
			err: MultiValidationError{
				Errors: []ValidationError{
					{Message: "error 1"},
					{Message: "error 2"},
				},
			},
			want: "multiple validation errors (2 errors)",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("MultiValidationError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}