// Package connection provides connection state management for MCP protocol handshakes.
package connection

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ConnectionState represents the current state of an MCP connection.
type ConnectionState int

const (
	// StateNew indicates a new connection that hasn't started handshake.
	StateNew ConnectionState = iota
	// StateInitializing indicates the connection is in the handshake process.
	StateInitializing
	// StateReady indicates the handshake is complete and connection is ready.
	StateReady
	// StateClosed indicates the connection has been closed.
	StateClosed
)

// String returns a string representation of the connection state.
func (s ConnectionState) String() string {
	switch s {
	case StateNew:
		return "New"
	case StateInitializing:
		return "Initializing"
	case StateReady:
		return "Ready"
	case StateClosed:
		return "Closed"
	default:
		return fmt.Sprintf("Unknown(%d)", s)
	}
}

// contextKey is a type for context keys to avoid collisions.
type contextKey string

const (
	// ConnectionIDKey is the context key for storing connection ID.
	ConnectionIDKey contextKey = "mcp:connection:id"
	// ConnectionStateKey is the context key for storing connection state.
	ConnectionStateKey contextKey = "mcp:connection:state"
)

// Connection represents a single MCP connection with its state and metadata.
type Connection struct {
	ID               string
	State            ConnectionState
	HandshakeStarted time.Time
	HandshakeTimeout time.Duration
	ProtocolVersion  string
	ClientInfo       map[string]interface{}

	mu            sync.RWMutex
	handshakeOnce sync.Once
	timeoutTimer  *time.Timer
}

// Manager manages connection states for multiple concurrent connections.
type Manager struct {
	connections map[string]*Connection
	mu          sync.RWMutex

	defaultTimeout time.Duration
}

// NewManager creates a new connection manager with the specified default timeout.
func NewManager(defaultTimeout time.Duration) *Manager {
	if defaultTimeout <= 0 {
		defaultTimeout = 30 * time.Second
	}

	return &Manager{
		connections:    make(map[string]*Connection),
		defaultTimeout: defaultTimeout,
	}
}

// CreateConnection creates a new connection with the given ID.
func (m *Manager) CreateConnection(id string) (*Connection, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.connections[id]; exists {
		return nil, fmt.Errorf("connection %s already exists", id)
	}

	conn := &Connection{
		ID:               id,
		State:            StateNew,
		HandshakeTimeout: m.defaultTimeout,
		ClientInfo:       make(map[string]interface{}),
	}

	m.connections[id] = conn
	return conn, nil
}

// GetConnection retrieves a connection by ID.
func (m *Manager) GetConnection(id string) (*Connection, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conn, exists := m.connections[id]
	return conn, exists
}

// RemoveConnection removes a connection from the manager.
func (m *Manager) RemoveConnection(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if conn, exists := m.connections[id]; exists {
		conn.Close()
		delete(m.connections, id)
	}
}

// GetState returns the current state of the connection.
func (c *Connection) GetState() ConnectionState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.State
}

// SetState updates the connection state with validation.
func (c *Connection) SetState(newState ConnectionState) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Validate state transitions
	if !c.isValidTransition(c.State, newState) {
		return fmt.Errorf("invalid state transition from %s to %s", c.State, newState)
	}

	c.State = newState

	// Handle state-specific logic
	switch newState {
	case StateInitializing:
		c.HandshakeStarted = time.Now()
	case StateReady, StateClosed:
		// Cancel timeout timer if it exists
		if c.timeoutTimer != nil {
			c.timeoutTimer.Stop()
			c.timeoutTimer = nil
		}
	}

	return nil
}

// IsReady returns true if the connection has completed handshake.
func (c *Connection) IsReady() bool {
	return c.GetState() == StateReady
}

// StartHandshake initiates the handshake process with timeout.
func (c *Connection) StartHandshake(timeoutCallback func()) error {
	var err error
	handshakeStarted := false

	c.handshakeOnce.Do(func() {
		handshakeStarted = true
		if e := c.SetState(StateInitializing); e != nil {
			err = fmt.Errorf("failed to start handshake: %w", e)
			return
		}

		// Start timeout timer
		c.mu.Lock()
		c.timeoutTimer = time.AfterFunc(c.HandshakeTimeout, func() {
			c.mu.Lock()
			if c.State == StateInitializing {
				c.State = StateClosed
			}
			c.mu.Unlock()

			if timeoutCallback != nil {
				timeoutCallback()
			}
		})
		c.mu.Unlock()
	})

	if err != nil {
		return err
	}

	// If handshakeOnce.Do didn't run, handshake was already started
	if !handshakeStarted {
		return fmt.Errorf("handshake already started")
	}

	return nil
}

// CompleteHandshake marks the handshake as complete.
func (c *Connection) CompleteHandshake(protocolVersion string, clientInfo map[string]interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.State != StateInitializing {
		return fmt.Errorf("cannot complete handshake in state %s", c.State)
	}

	c.State = StateReady
	c.ProtocolVersion = protocolVersion

	// Store client info
	for k, v := range clientInfo {
		c.ClientInfo[k] = v
	}

	// Cancel timeout timer
	if c.timeoutTimer != nil {
		c.timeoutTimer.Stop()
		c.timeoutTimer = nil
	}

	return nil
}

// Close closes the connection and cleans up resources.
func (c *Connection) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.State = StateClosed

	if c.timeoutTimer != nil {
		c.timeoutTimer.Stop()
		c.timeoutTimer = nil
	}
}

// isValidTransition checks if a state transition is allowed.
func (c *Connection) isValidTransition(from, to ConnectionState) bool {
	switch from {
	case StateNew:
		return to == StateInitializing || to == StateClosed
	case StateInitializing:
		return to == StateReady || to == StateClosed
	case StateReady:
		return to == StateClosed
	case StateClosed:
		return false // No transitions from closed
	default:
		return false
	}
}

// ConnectionFromContext retrieves the connection from context.
func ConnectionFromContext(ctx context.Context, manager *Manager) (*Connection, bool) {
	id, ok := ctx.Value(ConnectionIDKey).(string)
	if !ok {
		return nil, false
	}

	return manager.GetConnection(id)
}

// WithConnectionID adds a connection ID to the context.
func WithConnectionID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ConnectionIDKey, id)
}

// GetConnectionID retrieves the connection ID from the context.
func GetConnectionID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(ConnectionIDKey).(string)
	return id, ok
}
