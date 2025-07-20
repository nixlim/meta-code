package errors

import (
	"errors"
	"fmt"
)

// WrapError wraps an existing error with MCP error context
// Uses Go 1.13+ error wrapping with %w verb to preserve error chain
func WrapError(err error, code int, message string) *MCPError {
	if err == nil {
		return nil
	}

	category := GetCategory(code)

	return &MCPError{
		Code:     code,
		Message:  message,
		Data:     nil,
		Category: category,
		Cause:    err,
		Context:  make(map[string]interface{}),
	}
}

// WrapErrorf wraps an error with a formatted message
func WrapErrorf(err error, code int, format string, args ...interface{}) *MCPError {
	message := fmt.Sprintf(format, args...)
	return WrapError(err, code, message)
}

// WrapWithContext wraps an error and adds context information
func WrapWithContext(err error, code int, message string, context map[string]interface{}) *MCPError {
	mcpErr := WrapError(err, code, message)
	if mcpErr != nil && context != nil {
		for k, v := range context {
			mcpErr.Context[k] = v
		}
	}
	return mcpErr
}

// ChainError creates a new error that chains multiple errors together
func ChainError(primary error, secondary error, code int, message string) *MCPError {
	if primary == nil && secondary == nil {
		return nil
	}

	var cause error
	if primary != nil {
		cause = primary
	} else {
		cause = secondary
	}

	mcpErr := WrapError(cause, code, message)
	if mcpErr != nil && primary != nil && secondary != nil {
		mcpErr.WithContext("secondary_error", secondary.Error())
	}

	return mcpErr
}

// UnwrapAll returns all errors in the error chain
func UnwrapAll(err error) []error {
	var errs []error
	for err != nil {
		errs = append(errs, err)
		err = errors.Unwrap(err)
	}
	return errs
}

// FindMCPError searches the error chain for an MCPError
func FindMCPError(err error) *MCPError {
	var mcpErr *MCPError
	if errors.As(err, &mcpErr) {
		return mcpErr
	}
	return nil
}

// FindErrorCode searches the error chain for an error with a specific code
func FindErrorCode(err error, code int) bool {
	for err != nil {
		if mcpErr := FindMCPError(err); mcpErr != nil && mcpErr.Code == code {
			return true
		}
		err = errors.Unwrap(err)
	}
	return false
}

// IsTemporary checks if an error represents a temporary condition
func IsTemporary(err error) bool {
	mcpErr := FindMCPError(err)
	if mcpErr == nil {
		return false
	}

	// Consider these error codes as temporary
	temporaryCodes := []int{
		ErrorCodeMCPTransportTimeout,
		ErrorCodeMCPConnectionLost,
		ErrorCodeMCPHandshakeTimeout,
		ErrorCodeMCPRateLimit,
		ErrorCodeMCPResourceLimit,
		ErrorCodeMCPServiceUnavail,
	}

	for _, code := range temporaryCodes {
		if mcpErr.Code == code {
			return true
		}
	}

	return false
}

// IsRetryable checks if an error condition might succeed on retry
func IsRetryable(err error) bool {
	mcpErr := FindMCPError(err)
	if mcpErr == nil {
		return false
	}

	// Consider these error codes as retryable
	retryableCodes := []int{
		ErrorCodeMCPTransportTimeout,
		ErrorCodeMCPConnectionLost,
		ErrorCodeMCPConnectionFailed,
		ErrorCodeMCPHandshakeTimeout,
		ErrorCodeMCPRateLimit,
		ErrorCodeMCPServiceUnavail,
	}

	for _, code := range retryableCodes {
		if mcpErr.Code == code {
			return true
		}
	}

	return false
}

// IsFatal checks if an error represents a fatal condition that should not be retried
func IsFatal(err error) bool {
	mcpErr := FindMCPError(err)
	if mcpErr == nil {
		return false
	}

	// Consider these error codes as fatal
	fatalCodes := []int{
		ErrorCodeMCPVersionMismatch,
		ErrorCodeMCPCapabilityError,
		ErrorCodeMCPUnauthorized,
		ErrorCodeMCPForbidden,
		ErrorCodeMCPToolNotFound,
		ErrorCodeMCPResourceNotFound,
		ErrorCodeMCPPromptNotFound,
	}

	for _, code := range fatalCodes {
		if mcpErr.Code == code {
			return true
		}
	}

	return false
}

// AggregateErrors combines multiple errors into a single error
type AggregateError struct {
	Errors   []error
	Message  string
	Code     int
	Category string
}

// Error implements the error interface for AggregateError
func (ae *AggregateError) Error() string {
	if ae.Message != "" {
		return fmt.Sprintf("%s (%d errors)", ae.Message, len(ae.Errors))
	}
	return fmt.Sprintf("Multiple errors occurred (%d errors)", len(ae.Errors))
}

// Unwrap returns the first error for error chain compatibility
func (ae *AggregateError) Unwrap() error {
	if len(ae.Errors) > 0 {
		return ae.Errors[0]
	}
	return nil
}

// NewAggregateError creates a new aggregate error
func NewAggregateError(errors []error, code int, message string) *AggregateError {
	if len(errors) == 0 {
		return nil
	}

	// Filter out nil errors
	var validErrors []error
	for _, err := range errors {
		if err != nil {
			validErrors = append(validErrors, err)
		}
	}

	if len(validErrors) == 0 {
		return nil
	}

	return &AggregateError{
		Errors:   validErrors,
		Message:  message,
		Code:     code,
		Category: GetCategory(code),
	}
}

// ToMCPError converts an AggregateError to an MCPError
func (ae *AggregateError) ToMCPError() *MCPError {
	if ae == nil || len(ae.Errors) == 0 {
		return nil
	}

	// Use the first error as the primary cause
	mcpErr := WrapError(ae.Errors[0], ae.Code, ae.Message)
	if mcpErr != nil {
		// Add information about additional errors
		if len(ae.Errors) > 1 {
			var additionalErrors []string
			for i := 1; i < len(ae.Errors); i++ {
				additionalErrors = append(additionalErrors, ae.Errors[i].Error())
			}
			mcpErr.WithContext("additional_errors", additionalErrors)
		}
		mcpErr.WithContext("error_count", len(ae.Errors))
	}

	return mcpErr
}
