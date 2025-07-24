package transport

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/meta-mcp/meta-mcp-server/internal/protocol/jsonrpc"
)

// ConnectionType represents the type of transport connection
type ConnectionType string

const (
	// ConnectionTypeSTDIO represents STDIO transport (subprocess)
	ConnectionTypeSTDIO ConnectionType = "stdio"
	// ConnectionTypeHTTP represents HTTP/SSE transport (network)
	ConnectionTypeHTTP ConnectionType = "http"
)

// ConnectionConfig holds configuration for creating a transport connection
type ConnectionConfig struct {
	// Type specifies the transport type
	Type ConnectionType

	// For STDIO transport
	Command string   // Command to execute
	Args    []string // Command arguments
	Env     []string // Environment variables

	// For HTTP transport (future implementation)
	URL     string            // Server URL
	Headers map[string]string // HTTP headers
	TLS     *TLSConfig        // TLS configuration
}

// TLSConfig holds TLS configuration for secure connections
type TLSConfig struct {
	InsecureSkipVerify bool
	CertFile           string
	KeyFile            string
	CAFile             string
}

// Manager manages multiple transport connections
type Manager struct {
	connections map[string]jsonrpc.Transport
	configs     map[string]*ConnectionConfig
	mu          sync.RWMutex
}

// NewManager creates a new transport manager
func NewManager() *Manager {
	return &Manager{
		connections: make(map[string]jsonrpc.Transport),
		configs:     make(map[string]*ConnectionConfig),
	}
}

// AddConnection creates and adds a new transport connection
func (m *Manager) AddConnection(id string, config *ConnectionConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if connection already exists
	if _, exists := m.connections[id]; exists {
		return fmt.Errorf("connection %s already exists", id)
	}

	// Create transport based on type
	var transport jsonrpc.Transport
	var err error

	switch config.Type {
	case ConnectionTypeSTDIO:
		transport, err = m.createSTDIOTransport(config)
	case ConnectionTypeHTTP:
		// TODO: Implement HTTP transport
		return fmt.Errorf("HTTP transport not yet implemented")
	default:
		return fmt.Errorf("unknown connection type: %s", config.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to create transport: %w", err)
	}

	// Store connection and config
	m.connections[id] = transport
	m.configs[id] = config

	return nil
}

// RemoveConnection removes and closes a connection
func (m *Manager) RemoveConnection(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	transport, exists := m.connections[id]
	if !exists {
		return fmt.Errorf("connection %s not found", id)
	}

	// Close the transport
	if err := transport.Close(); err != nil {
		return fmt.Errorf("failed to close transport: %w", err)
	}

	// Remove from maps
	delete(m.connections, id)
	delete(m.configs, id)

	return nil
}

// GetConnection retrieves a connection by ID
func (m *Manager) GetConnection(id string) (jsonrpc.Transport, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	transport, exists := m.connections[id]
	return transport, exists
}

// ListConnections returns all connection IDs
func (m *Manager) ListConnections() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.connections))
	for id := range m.connections {
		ids = append(ids, id)
	}
	return ids
}

// GetConnectionInfo returns information about a connection
func (m *Manager) GetConnectionInfo(id string) (info ConnectionInfo, exists bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	transport, exists := m.connections[id]
	if !exists {
		return ConnectionInfo{}, false
	}

	config, _ := m.configs[id]

	info = ConnectionInfo{
		ID:        id,
		Type:      config.Type,
		Connected: transport.IsConnected(),
		Config:    config,
	}

	// Add transport-specific info
	if stdioTransport, ok := transport.(*STDIOTransport); ok {
		pid, running := stdioTransport.GetProcessInfo()
		info.ProcessID = pid
		info.Running = running
	}

	return info, true
}

// Broadcast sends a message to all connected transports
func (m *Manager) Broadcast(ctx context.Context, message jsonrpc.Message) error {
	m.mu.RLock()
	transports := make(map[string]jsonrpc.Transport)
	for id, transport := range m.connections {
		transports[id] = transport
	}
	m.mu.RUnlock()

	var errors []error
	for id, transport := range transports {
		if !transport.IsConnected() {
			continue
		}

		if err := transport.Send(ctx, message); err != nil {
			// Only record errors that aren't due to disconnection
			if !strings.Contains(err.Error(), "broken pipe") && !strings.Contains(err.Error(), "transport is not connected") {
				errors = append(errors, fmt.Errorf("failed to send to %s: %w", id, err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("broadcast errors: %v", errors)
	}

	return nil
}

// Close closes all connections
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errors []error
	for id, transport := range m.connections {
		if err := transport.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close %s: %w", id, err))
		}
	}

	// Clear maps
	m.connections = make(map[string]jsonrpc.Transport)
	m.configs = make(map[string]*ConnectionConfig)

	if len(errors) > 0 {
		return fmt.Errorf("close errors: %v", errors)
	}

	return nil
}

// ConnectionInfo holds information about a connection
type ConnectionInfo struct {
	ID        string
	Type      ConnectionType
	Connected bool
	Config    *ConnectionConfig

	// STDIO-specific
	ProcessID int
	Running   bool
}

// createSTDIOTransport creates a new STDIO transport from config
func (m *Manager) createSTDIOTransport(config *ConnectionConfig) (jsonrpc.Transport, error) {
	if config.Command == "" {
		return nil, fmt.Errorf("command is required for STDIO transport")
	}

	cmd := exec.Command(config.Command, config.Args...)
	if len(config.Env) > 0 {
		cmd.Env = config.Env
	}

	return NewSTDIOTransport(cmd)
}

// HealthCheck checks the health of all connections
func (m *Manager) HealthCheck() map[string]HealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]HealthStatus)
	for id, transport := range m.connections {
		status[id] = HealthStatus{
			ID:        id,
			Connected: transport.IsConnected(),
		}

		// Add transport-specific health info
		if stdioTransport, ok := transport.(*STDIOTransport); ok {
			pid, running := stdioTransport.GetProcessInfo()
			healthStatus := status[id]
			healthStatus.ProcessID = pid
			healthStatus.Running = running
			healthStatus.LastError = stdioTransport.GetLastError()
			status[id] = healthStatus
		}
	}

	return status
}

// HealthStatus represents the health status of a connection
type HealthStatus struct {
	ID        string
	Connected bool
	ProcessID int
	Running   bool
	LastError error
}

// RestartConnection restarts a connection with the same configuration
func (m *Manager) RestartConnection(id string) error {
	m.mu.Lock()
	config, exists := m.configs[id]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("connection %s not found", id)
	}

	// Make a copy of the config
	configCopy := *config
	m.mu.Unlock()

	// Remove the old connection
	if err := m.RemoveConnection(id); err != nil {
		return fmt.Errorf("failed to remove old connection: %w", err)
	}

	// Add a new connection with the same config
	if err := m.AddConnection(id, &configCopy); err != nil {
		return fmt.Errorf("failed to create new connection: %w", err)
	}

	return nil
}