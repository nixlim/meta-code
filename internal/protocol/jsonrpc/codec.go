package jsonrpc

import (
	"bytes"
	"encoding/json"
)

// Message represents any JSON-RPC message (Request, Response, or Notification)
type Message interface {
	Validate() error
}

// ParseMessage parses a single JSON-RPC message from raw bytes
func ParseMessage(raw []byte) (Message, error) {
	// First, parse into a generic map to determine the message type
	var generic map[string]json.RawMessage
	if err := json.Unmarshal(raw, &generic); err != nil {
		return nil, NewParseError("Invalid JSON")
	}

	// Check for required jsonrpc field
	versionRaw, hasVersion := generic["jsonrpc"]
	if !hasVersion {
		return nil, NewInvalidRequestError("Missing jsonrpc field")
	}

	var version string
	if err := json.Unmarshal(versionRaw, &version); err != nil {
		return nil, NewInvalidRequestError("Invalid jsonrpc field")
	}

	if version != Version {
		return nil, NewInvalidRequestError("jsonrpc field must be \"2.0\"")
	}

	// Determine message type based on presence of fields
	_, hasMethod := generic["method"]
	_, hasResult := generic["result"]
	_, hasError := generic["error"]
	_, hasID := generic["id"]

	if hasMethod {
		// This is either a Request or Notification
		if hasID {
			// Request
			var req Request
			if err := json.Unmarshal(raw, &req); err != nil {
				return nil, NewParseError("Invalid request format")
			}
			if err := req.Validate(); err != nil {
				return nil, err
			}
			return &req, nil
		} else {
			// Notification
			var notif Notification
			if err := json.Unmarshal(raw, &notif); err != nil {
				return nil, NewParseError("Invalid notification format")
			}
			if err := notif.Validate(); err != nil {
				return nil, err
			}
			return &notif, nil
		}
	} else if hasResult || hasError {
		// This is a Response
		var resp Response
		if err := json.Unmarshal(raw, &resp); err != nil {
			return nil, NewParseError("Invalid response format")
		}
		if err := resp.Validate(); err != nil {
			return nil, err
		}
		return &resp, nil
	}

	return nil, NewInvalidRequestError("Invalid message format")
}

// Parse handles both single messages and batch requests
func Parse(raw []byte) ([]Message, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return nil, NewParseError("Empty request body")
	}

	// Check if this is a batch request (starts with '[')
	if trimmed[0] == '[' {
		return parseBatch(trimmed)
	}

	// Single message
	if trimmed[0] == '{' {
		msg, err := ParseMessage(trimmed)
		if err != nil {
			return nil, err
		}
		return []Message{msg}, nil
	}

	return nil, NewParseError("Message must be a JSON object or array")
}

// parseBatch parses a batch of JSON-RPC messages
func parseBatch(raw []byte) ([]Message, error) {
	var rawMessages []json.RawMessage
	if err := json.Unmarshal(raw, &rawMessages); err != nil {
		return nil, NewParseError("Invalid batch format")
	}

	if len(rawMessages) == 0 {
		return nil, NewInvalidRequestError("Batch array must not be empty")
	}

	results := make([]Message, 0, len(rawMessages))
	for _, rawMsg := range rawMessages {
		msg, err := ParseMessage(rawMsg)
		if err != nil {
			// For batch requests, we continue parsing other messages
			// but include the error in the results
			results = append(results, &Response{
				Version: Version,
				Error:   err.(*Error),
				ID:      nil, // Unknown ID for parse errors
			})
		} else {
			results = append(results, msg)
		}
	}

	return results, nil
}

// Marshal serializes a message to JSON bytes
func Marshal(msg Message) ([]byte, error) {
	return json.Marshal(msg)
}

// MarshalBatch serializes multiple messages as a JSON array
func MarshalBatch(messages []Message) ([]byte, error) {
	if len(messages) == 0 {
		return nil, NewInvalidRequestError("Cannot marshal empty batch")
	}

	if len(messages) == 1 {
		// Single message, don't wrap in array
		return Marshal(messages[0])
	}

	// Multiple messages, wrap in array
	return json.Marshal(messages)
}
