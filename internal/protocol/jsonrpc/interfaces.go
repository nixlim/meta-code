package jsonrpc

import (
	"context"
	"io"
)

// Transport defines the interface for JSON-RPC transport mechanisms
type Transport interface {
	// Send sends a message over the transport
	Send(ctx context.Context, message Message) error

	// Receive receives a message from the transport
	Receive(ctx context.Context) (Message, error)

	// SendBatch sends multiple messages as a batch
	SendBatch(ctx context.Context, messages []Message) error

	// ReceiveBatch receives multiple messages as a batch
	ReceiveBatch(ctx context.Context) ([]Message, error)

	// Close closes the transport connection
	Close() error

	// IsConnected returns true if the transport is connected
	IsConnected() bool
}

// Handler defines the interface for handling JSON-RPC requests
type Handler interface {
	// Handle processes a JSON-RPC request and returns a response
	Handle(ctx context.Context, request *Request) *Response

	// HandleNotification processes a JSON-RPC notification
	HandleNotification(ctx context.Context, notification *Notification)

	// GetMethods returns a list of supported methods
	GetMethods() []string

	// HasMethod returns true if the handler supports the given method
	HasMethod(method string) bool
}

// Server defines the interface for a JSON-RPC server
type Server interface {
	// Start starts the server
	Start(ctx context.Context) error

	// Stop stops the server
	Stop(ctx context.Context) error

	// RegisterHandler registers a handler for JSON-RPC requests
	RegisterHandler(handler Handler)

	// RegisterMethod registers a single method handler
	RegisterMethod(method string, handler func(ctx context.Context, params any) (any, error))

	// IsRunning returns true if the server is running
	IsRunning() bool
}

// Client defines the interface for a JSON-RPC client
type Client interface {
	// Call makes a JSON-RPC request and waits for a response
	Call(ctx context.Context, method string, params any, result any) error

	// CallWithID makes a JSON-RPC request with a specific ID
	CallWithID(ctx context.Context, id any, method string, params any, result any) error

	// Notify sends a JSON-RPC notification (no response expected)
	Notify(ctx context.Context, method string, params any) error

	// BatchCall makes multiple JSON-RPC requests in a single batch
	BatchCall(ctx context.Context, requests []*Request) ([]*Response, error)

	// Connect establishes a connection to the server
	Connect(ctx context.Context) error

	// Disconnect closes the connection to the server
	Disconnect() error

	// IsConnected returns true if the client is connected
	IsConnected() bool
}

// Codec defines the interface for encoding/decoding JSON-RPC messages
type Codec interface {
	// Encode encodes a message to the writer
	Encode(w io.Writer, message Message) error

	// Decode decodes a message from the reader
	Decode(r io.Reader) (Message, error)

	// EncodeBatch encodes multiple messages to the writer
	EncodeBatch(w io.Writer, messages []Message) error

	// DecodeBatch decodes multiple messages from the reader
	DecodeBatch(r io.Reader) ([]Message, error)
}

// MethodHandler is a function type for handling individual methods
type MethodHandler func(ctx context.Context, params any) (any, error)

// NotificationHandler is a function type for handling notifications
type NotificationHandler func(ctx context.Context, params any)

// Middleware defines the interface for request/response middleware
type Middleware interface {
	// ProcessRequest processes a request before it reaches the handler
	ProcessRequest(ctx context.Context, request *Request) (*Request, error)

	// ProcessResponse processes a response before it's sent back
	ProcessResponse(ctx context.Context, response *Response) (*Response, error)

	// ProcessNotification processes a notification before it reaches the handler
	ProcessNotification(ctx context.Context, notification *Notification) (*Notification, error)
}

// Logger defines the interface for logging JSON-RPC operations
type Logger interface {
	// LogRequest logs an incoming request
	LogRequest(ctx context.Context, request *Request)

	// LogResponse logs an outgoing response
	LogResponse(ctx context.Context, response *Response)

	// LogNotification logs a notification
	LogNotification(ctx context.Context, notification *Notification)

	// LogError logs an error
	LogError(ctx context.Context, err error)
}

// ConnectionManager defines the interface for managing multiple connections
type ConnectionManager interface {
	// AddConnection adds a new connection
	AddConnection(id string, transport Transport) error

	// RemoveConnection removes a connection
	RemoveConnection(id string) error

	// GetConnection gets a connection by ID
	GetConnection(id string) (Transport, bool)

	// ListConnections returns all connection IDs
	ListConnections() []string

	// Broadcast sends a message to all connections
	Broadcast(ctx context.Context, message Message) error

	// Close closes all connections
	Close() error
}
