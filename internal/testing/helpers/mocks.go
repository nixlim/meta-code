// Package helpers provides mock implementations for testing
package helpers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

// MockTransport provides a mock implementation of MCP transport
type MockTransport struct {
	mu             sync.RWMutex
	sentMessages   [][]byte
	receiveQueue   [][]byte
	onSend         func([]byte) error
	onReceive      func() ([]byte, error)
	closed         bool
	closeErr       error
	readDelay      time.Duration
	writeDelay     time.Duration
	failAfterCount int
	sendCount      int
	receiveCount   int
	t              *testing.T
}

// NewMockTransport creates a new mock transport
func NewMockTransport(t *testing.T) *MockTransport {
	return &MockTransport{
		t:            t,
		sentMessages: make([][]byte, 0),
		receiveQueue: make([][]byte, 0),
	}
}

// Send sends a message through the transport
func (mt *MockTransport) Send(data []byte) error {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	if mt.closed {
		return errors.New("transport closed")
	}

	mt.sendCount++

	// Check if we should fail
	if mt.failAfterCount > 0 && mt.sendCount > mt.failAfterCount {
		return errors.New("send failure triggered")
	}

	// Apply write delay
	if mt.writeDelay > 0 {
		time.Sleep(mt.writeDelay)
	}

	// Store sent message
	msgCopy := make([]byte, len(data))
	copy(msgCopy, data)
	mt.sentMessages = append(mt.sentMessages, msgCopy)

	// Call custom handler if set
	if mt.onSend != nil {
		return mt.onSend(data)
	}

	return nil
}

// Receive receives a message from the transport
func (mt *MockTransport) Receive() ([]byte, error) {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	if mt.closed {
		return nil, io.EOF
	}

	mt.receiveCount++

	// Apply read delay
	if mt.readDelay > 0 {
		time.Sleep(mt.readDelay)
	}

	// Call custom handler if set
	if mt.onReceive != nil {
		return mt.onReceive()
	}

	// Return from queue
	if len(mt.receiveQueue) > 0 {
		msg := mt.receiveQueue[0]
		mt.receiveQueue = mt.receiveQueue[1:]
		return msg, nil
	}

	return nil, io.EOF
}

// Close closes the transport
func (mt *MockTransport) Close() error {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	if mt.closed {
		return errors.New("already closed")
	}

	mt.closed = true
	return mt.closeErr
}

// QueueReceive adds a message to the receive queue
func (mt *MockTransport) QueueReceive(data []byte) {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	msgCopy := make([]byte, len(data))
	copy(msgCopy, data)
	mt.receiveQueue = append(mt.receiveQueue, msgCopy)
}

// QueueJSON adds a JSON message to the receive queue
func (mt *MockTransport) QueueJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	mt.QueueReceive(data)
	return nil
}

// GetSentMessages returns all sent messages
func (mt *MockTransport) GetSentMessages() [][]byte {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	result := make([][]byte, len(mt.sentMessages))
	for i, msg := range mt.sentMessages {
		msgCopy := make([]byte, len(msg))
		copy(msgCopy, msg)
		result[i] = msgCopy
	}
	return result
}

// GetLastSent returns the last sent message
func (mt *MockTransport) GetLastSent() []byte {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	if len(mt.sentMessages) == 0 {
		return nil
	}

	last := mt.sentMessages[len(mt.sentMessages)-1]
	result := make([]byte, len(last))
	copy(result, last)
	return result
}

// ClearSent clears all sent messages
func (mt *MockTransport) ClearSent() {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	mt.sentMessages = make([][]byte, 0)
}

// SetOnSend sets a custom send handler
func (mt *MockTransport) SetOnSend(handler func([]byte) error) {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	mt.onSend = handler
}

// SetOnReceive sets a custom receive handler
func (mt *MockTransport) SetOnReceive(handler func() ([]byte, error)) {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	mt.onReceive = handler
}

// SetReadDelay sets a delay for read operations
func (mt *MockTransport) SetReadDelay(delay time.Duration) {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	mt.readDelay = delay
}

// SetWriteDelay sets a delay for write operations
func (mt *MockTransport) SetWriteDelay(delay time.Duration) {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	mt.writeDelay = delay
}

// SetFailAfterCount sets the transport to fail after n operations
func (mt *MockTransport) SetFailAfterCount(count int) {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	mt.failAfterCount = count
}

// SetCloseError sets the error to return on close
func (mt *MockTransport) SetCloseError(err error) {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	mt.closeErr = err
}

// IsClosed returns whether the transport is closed
func (mt *MockTransport) IsClosed() bool {
	mt.mu.RLock()
	defer mt.mu.RUnlock()
	return mt.closed
}

// MockHandler provides a mock implementation of request handlers
type MockHandler struct {
	mu             sync.RWMutex
	handlers       map[string]HandlerFunc
	defaultHandler HandlerFunc
	callCount      map[string]int
	lastRequest    map[string]interface{}
	responseDelay  time.Duration
	t              *testing.T
}

// HandlerFunc is a function that handles a request
type HandlerFunc func(method string, params interface{}) (interface{}, error)

// NewMockHandler creates a new mock handler
func NewMockHandler(t *testing.T) *MockHandler {
	return &MockHandler{
		t:           t,
		handlers:    make(map[string]HandlerFunc),
		callCount:   make(map[string]int),
		lastRequest: make(map[string]interface{}),
	}
}

// HandleRequest handles an incoming request
func (mh *MockHandler) HandleRequest(ctx context.Context, method string, params interface{}) (interface{}, error) {
	mh.mu.Lock()
	defer mh.mu.Unlock()

	// Apply response delay
	if mh.responseDelay > 0 {
		time.Sleep(mh.responseDelay)
	}

	// Update call tracking
	mh.callCount[method]++
	mh.lastRequest[method] = params

	// Check for specific handler
	if handler, ok := mh.handlers[method]; ok {
		return handler(method, params)
	}

	// Use default handler if set
	if mh.defaultHandler != nil {
		return mh.defaultHandler(method, params)
	}

	// Return error if no handler found
	return nil, errors.New("method not found: " + method)
}

// RegisterHandler registers a handler for a specific method
func (mh *MockHandler) RegisterHandler(method string, handler HandlerFunc) {
	mh.mu.Lock()
	defer mh.mu.Unlock()
	mh.handlers[method] = handler
}

// RegisterHandlers registers multiple handlers at once
func (mh *MockHandler) RegisterHandlers(handlers map[string]HandlerFunc) {
	mh.mu.Lock()
	defer mh.mu.Unlock()

	for method, handler := range handlers {
		mh.handlers[method] = handler
	}
}

// SetDefaultHandler sets the default handler for unregistered methods
func (mh *MockHandler) SetDefaultHandler(handler HandlerFunc) {
	mh.mu.Lock()
	defer mh.mu.Unlock()
	mh.defaultHandler = handler
}

// GetCallCount returns the number of times a method was called
func (mh *MockHandler) GetCallCount(method string) int {
	mh.mu.RLock()
	defer mh.mu.RUnlock()
	return mh.callCount[method]
}

// GetLastRequest returns the last request params for a method
func (mh *MockHandler) GetLastRequest(method string) interface{} {
	mh.mu.RLock()
	defer mh.mu.RUnlock()
	return mh.lastRequest[method]
}

// WasCalled returns whether a method was called
func (mh *MockHandler) WasCalled(method string) bool {
	return mh.GetCallCount(method) > 0
}

// Reset resets all call tracking
func (mh *MockHandler) Reset() {
	mh.mu.Lock()
	defer mh.mu.Unlock()

	mh.callCount = make(map[string]int)
	mh.lastRequest = make(map[string]interface{})
}

// SetResponseDelay sets a delay for all responses
func (mh *MockHandler) SetResponseDelay(delay time.Duration) {
	mh.mu.Lock()
	defer mh.mu.Unlock()
	mh.responseDelay = delay
}

// MockConnection provides a mock network connection
type MockConnection struct {
	mu            sync.RWMutex
	readBuffer    []byte
	writeBuffer   []byte
	readErr       error
	writeErr      error
	closeErr      error
	closed        bool
	localAddr     net.Addr
	remoteAddr    net.Addr
	readDeadline  time.Time
	writeDeadline time.Time
	deadline      time.Time
	t             *testing.T
}

// NewMockConnection creates a new mock connection
func NewMockConnection(t *testing.T) *MockConnection {
	return &MockConnection{
		t:          t,
		localAddr:  &mockAddr{network: "tcp", address: "127.0.0.1:8080"},
		remoteAddr: &mockAddr{network: "tcp", address: "127.0.0.1:9090"},
	}
}

// Read reads data from the connection
func (mc *MockConnection) Read(b []byte) (n int, err error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.closed {
		return 0, io.EOF
	}

	if mc.readErr != nil {
		return 0, mc.readErr
	}

	// Check deadline
	if !mc.readDeadline.IsZero() && time.Now().After(mc.readDeadline) {
		return 0, errors.New("read deadline exceeded")
	}

	// Copy data from buffer
	n = copy(b, mc.readBuffer)
	mc.readBuffer = mc.readBuffer[n:]

	if n == 0 {
		return 0, io.EOF
	}

	return n, nil
}

// Write writes data to the connection
func (mc *MockConnection) Write(b []byte) (n int, err error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.closed {
		return 0, errors.New("connection closed")
	}

	if mc.writeErr != nil {
		return 0, mc.writeErr
	}

	// Check deadline
	if !mc.writeDeadline.IsZero() && time.Now().After(mc.writeDeadline) {
		return 0, errors.New("write deadline exceeded")
	}

	// Append to write buffer
	mc.writeBuffer = append(mc.writeBuffer, b...)
	return len(b), nil
}

// Close closes the connection
func (mc *MockConnection) Close() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.closed {
		return errors.New("already closed")
	}

	mc.closed = true
	return mc.closeErr
}

// LocalAddr returns the local network address
func (mc *MockConnection) LocalAddr() net.Addr {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.localAddr
}

// RemoteAddr returns the remote network address
func (mc *MockConnection) RemoteAddr() net.Addr {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.remoteAddr
}

// SetDeadline sets the read and write deadlines
func (mc *MockConnection) SetDeadline(t time.Time) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.deadline = t
	mc.readDeadline = t
	mc.writeDeadline = t
	return nil
}

// SetReadDeadline sets the deadline for future Read calls
func (mc *MockConnection) SetReadDeadline(t time.Time) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.readDeadline = t
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls
func (mc *MockConnection) SetWriteDeadline(t time.Time) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.writeDeadline = t
	return nil
}

// AddReadData adds data to be read from the connection
func (mc *MockConnection) AddReadData(data []byte) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.readBuffer = append(mc.readBuffer, data...)
}

// GetWrittenData returns all data written to the connection
func (mc *MockConnection) GetWrittenData() []byte {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := make([]byte, len(mc.writeBuffer))
	copy(result, mc.writeBuffer)
	return result
}

// ClearWrittenData clears the write buffer
func (mc *MockConnection) ClearWrittenData() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.writeBuffer = nil
}

// SetReadError sets an error to be returned on Read
func (mc *MockConnection) SetReadError(err error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.readErr = err
}

// SetWriteError sets an error to be returned on Write
func (mc *MockConnection) SetWriteError(err error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.writeErr = err
}

// SetCloseError sets an error to be returned on Close
func (mc *MockConnection) SetCloseError(err error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.closeErr = err
}

// IsClosed returns whether the connection is closed
func (mc *MockConnection) IsClosed() bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.closed
}

// mockAddr implements net.Addr
type mockAddr struct {
	network string
	address string
}

func (ma *mockAddr) Network() string {
	return ma.network
}

func (ma *mockAddr) String() string {
	return ma.address
}

// MockServer provides a mock MCP server implementation
type MockServer struct {
	mu          sync.RWMutex
	handler     *MockHandler
	transport   *MockTransport
	started     bool
	stopChan    chan struct{}
	connections []net.Conn
	t           *testing.T
}

// NewMockServer creates a new mock server
func NewMockServer(t *testing.T) *MockServer {
	return &MockServer{
		t:           t,
		handler:     NewMockHandler(t),
		transport:   NewMockTransport(t),
		stopChan:    make(chan struct{}),
		connections: make([]net.Conn, 0),
	}
}

// Start starts the mock server
func (ms *MockServer) Start() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.started {
		return errors.New("server already started")
	}

	ms.started = true
	return nil
}

// Stop stops the mock server
func (ms *MockServer) Stop() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if !ms.started {
		return errors.New("server not started")
	}

	close(ms.stopChan)
	ms.started = false

	// Close all connections
	for _, conn := range ms.connections {
		conn.Close()
	}

	return nil
}

// GetHandler returns the server's handler
func (ms *MockServer) GetHandler() *MockHandler {
	return ms.handler
}

// GetTransport returns the server's transport
func (ms *MockServer) GetTransport() *MockTransport {
	return ms.transport
}

// AddConnection adds a connection to track
func (ms *MockServer) AddConnection(conn net.Conn) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.connections = append(ms.connections, conn)
}

// IsStarted returns whether the server is started
func (ms *MockServer) IsStarted() bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.started
}

// MockClient provides a mock MCP client implementation
type MockClient struct {
	mu        sync.RWMutex
	transport *MockTransport
	requests  []interface{}
	responses map[interface{}]interface{}
	errors    map[interface{}]error
	t         *testing.T
}

// NewMockClient creates a new mock client
func NewMockClient(t *testing.T) *MockClient {
	return &MockClient{
		t:         t,
		transport: NewMockTransport(t),
		requests:  make([]interface{}, 0),
		responses: make(map[interface{}]interface{}),
		errors:    make(map[interface{}]error),
	}
}

// SendRequest sends a request and returns the configured response
func (mc *MockClient) SendRequest(ctx context.Context, method string, params interface{}) (interface{}, error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Record the request
	req := map[string]interface{}{
		"method": method,
		"params": params,
	}
	mc.requests = append(mc.requests, req)

	// Check for configured error
	if err, ok := mc.errors[method]; ok {
		return nil, err
	}

	// Return configured response
	if resp, ok := mc.responses[method]; ok {
		return resp, nil
	}

	// Default response
	return map[string]interface{}{"success": true}, nil
}

// SetResponse sets the response for a method
func (mc *MockClient) SetResponse(method string, response interface{}) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.responses[method] = response
}

// SetError sets an error for a method
func (mc *MockClient) SetError(method string, err error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.errors[method] = err
}

// GetRequests returns all sent requests
func (mc *MockClient) GetRequests() []interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := make([]interface{}, len(mc.requests))
	copy(result, mc.requests)
	return result
}

// GetTransport returns the client's transport
func (mc *MockClient) GetTransport() *MockTransport {
	return mc.transport
}

// Reset clears all requests and configured responses
func (mc *MockClient) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.requests = make([]interface{}, 0)
	mc.responses = make(map[interface{}]interface{})
	mc.errors = make(map[interface{}]error)
}
