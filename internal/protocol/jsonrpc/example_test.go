package jsonrpc_test

import (
	"fmt"
	"log"

	"github.com/meta-mcp/meta-mcp-server/internal/protocol/jsonrpc"
)

func ExampleRequest() {
	// Create a new JSON-RPC request
	req := jsonrpc.NewRequest("get_user", map[string]any{
		"id":   123,
		"name": "john_doe",
	}, "req-1")

	// Validate the request
	if err := req.Validate(); err != nil {
		log.Fatal(err)
	}

	// Marshal to JSON
	data, err := jsonrpc.Marshal(req)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Request JSON: %s\n", string(data))

	// Parse it back
	msg, err := jsonrpc.ParseMessage(data)
	if err != nil {
		log.Fatal(err)
	}

	if parsedReq, ok := msg.(*jsonrpc.Request); ok {
		fmt.Printf("Parsed method: %s\n", parsedReq.Method)
		fmt.Printf("Is request: %t\n", parsedReq.IsRequest())
	}

	// Output:
	// Request JSON: {"jsonrpc":"2.0","method":"get_user","params":{"id":123,"name":"john_doe"},"id":"req-1"}
	// Parsed method: get_user
	// Is request: true
}

func ExampleNotification() {
	// Create a notification (no ID)
	notif := jsonrpc.NewNotification("user_updated", map[string]any{
		"user_id": 123,
		"status":  "active",
	})

	// Validate the notification
	if err := notif.Validate(); err != nil {
		log.Fatal(err)
	}

	// Marshal to JSON
	data, err := jsonrpc.Marshal(notif)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Notification JSON: %s\n", string(data))

	// Output:
	// Notification JSON: {"jsonrpc":"2.0","method":"user_updated","params":{"status":"active","user_id":123}}
}

func ExampleResponse() {
	// Create a successful response
	resp := jsonrpc.NewResponse(map[string]any{
		"user_id": 123,
		"name":    "John Doe",
		"email":   "john@example.com",
	}, "req-1")

	// Validate the response
	if err := resp.Validate(); err != nil {
		log.Fatal(err)
	}

	// Marshal to JSON
	data, err := jsonrpc.Marshal(resp)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Response JSON: %s\n", string(data))
	fmt.Printf("Has result: %t\n", resp.HasResult())

	// Output:
	// Response JSON: {"jsonrpc":"2.0","result":{"email":"john@example.com","name":"John Doe","user_id":123},"id":"req-1"}
	// Has result: true
}

func ExampleNewErrorResponse() {
	// Create an error response
	err := jsonrpc.NewMethodNotFoundError("unknown_method")
	resp := jsonrpc.NewErrorResponse(err, "req-1")

	// Validate the response
	if validationErr := resp.Validate(); validationErr != nil {
		log.Fatal(validationErr)
	}

	// Marshal to JSON
	data, marshalErr := jsonrpc.Marshal(resp)
	if marshalErr != nil {
		log.Fatal(marshalErr)
	}

	fmt.Printf("Error response JSON: %s\n", string(data))
	fmt.Printf("Has error: %t\n", resp.HasError())

	// Output:
	// Error response JSON: {"jsonrpc":"2.0","error":{"code":-32601,"message":"Method not found","data":"unknown_method"},"id":"req-1"}
	// Has error: true
}

func ExampleParse_batch() {
	// Parse a batch of JSON-RPC messages
	batchJSON := `[
		{"jsonrpc":"2.0","method":"get_user","params":{"id":1},"id":"req-1"},
		{"jsonrpc":"2.0","method":"user_updated","params":{"id":1,"status":"active"}},
		{"jsonrpc":"2.0","result":{"id":1,"name":"John"},"id":"req-1"}
	]`

	messages, err := jsonrpc.Parse([]byte(batchJSON))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Parsed %d messages:\n", len(messages))
	for i, msg := range messages {
		switch m := msg.(type) {
		case *jsonrpc.Request:
			fmt.Printf("  %d: Request - %s\n", i+1, m.Method)
		case *jsonrpc.Notification:
			fmt.Printf("  %d: Notification - %s\n", i+1, m.Method)
		case *jsonrpc.Response:
			if m.HasResult() {
				fmt.Printf("  %d: Response - success\n", i+1)
			} else {
				fmt.Printf("  %d: Response - error\n", i+1)
			}
		}
	}

	// Output:
	// Parsed 3 messages:
	//   1: Request - get_user
	//   2: Notification - user_updated
	//   3: Response - success
}

func ExampleRequest_BindParams() {
	// Define a struct for parameters
	type UserParams struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	// Create a request with parameters
	req := jsonrpc.NewRequest("get_user", map[string]any{
		"id":   123,
		"name": "john_doe",
	}, 1)

	// Bind parameters to struct
	var params UserParams
	if err := req.BindParams(&params); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("User ID: %d\n", params.ID)
	fmt.Printf("User Name: %s\n", params.Name)

	// Output:
	// User ID: 123
	// User Name: john_doe
}
