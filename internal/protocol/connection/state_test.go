package connection

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestConnectionState_String(t *testing.T) {
	tests := []struct {
		state ConnectionState
		want  string
	}{
		{StateNew, "New"},
		{StateInitializing, "Initializing"},
		{StateReady, "Ready"},
		{StateClosed, "Closed"},
		{ConnectionState(99), "Unknown(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.state.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManager_CreateConnection(t *testing.T) {
	manager := NewManager(10 * time.Second)

	// Test creating new connection
	conn, err := manager.CreateConnection("conn1")
	if err != nil {
		t.Fatalf("CreateConnection() error = %v", err)
	}

	if conn.ID != "conn1" {
		t.Errorf("Connection ID = %v, want conn1", conn.ID)
	}

	if conn.State != StateNew {
		t.Errorf("Initial state = %v, want StateNew", conn.State)
	}

	// Test creating duplicate connection
	_, err = manager.CreateConnection("conn1")
	if err == nil {
		t.Error("Expected error for duplicate connection ID")
	}
}

func TestManager_GetConnection(t *testing.T) {
	manager := NewManager(10 * time.Second)

	// Create a connection
	original, _ := manager.CreateConnection("conn1")

	// Test getting existing connection
	conn, exists := manager.GetConnection("conn1")
	if !exists {
		t.Error("GetConnection() exists = false, want true")
	}

	if conn != original {
		t.Error("GetConnection() returned different connection instance")
	}

	// Test getting non-existent connection
	_, exists = manager.GetConnection("conn2")
	if exists {
		t.Error("GetConnection() exists = true for non-existent connection")
	}
}

func TestManager_RemoveConnection(t *testing.T) {
	manager := NewManager(10 * time.Second)

	// Create and remove connection
	manager.CreateConnection("conn1")
	manager.RemoveConnection("conn1")

	// Verify connection is removed
	_, exists := manager.GetConnection("conn1")
	if exists {
		t.Error("Connection still exists after removal")
	}

	// Test removing non-existent connection (should not panic)
	manager.RemoveConnection("conn2")
}

func TestConnection_StateTransitions(t *testing.T) {
	conn := &Connection{
		ID:         "test",
		State:      StateNew,
		ClientInfo: make(map[string]interface{}),
	}

	tests := []struct {
		name      string
		fromState ConnectionState
		toState   ConnectionState
		wantErr   bool
	}{
		// Valid transitions
		{"New to Initializing", StateNew, StateInitializing, false},
		{"New to Closed", StateNew, StateClosed, false},
		{"Initializing to Ready", StateInitializing, StateReady, false},
		{"Initializing to Closed", StateInitializing, StateClosed, false},
		{"Ready to Closed", StateReady, StateClosed, false},

		// Invalid transitions
		{"New to Ready", StateNew, StateReady, true},
		{"Ready to Initializing", StateReady, StateInitializing, true},
		{"Closed to any", StateClosed, StateNew, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn.State = tt.fromState
			err := conn.SetState(tt.toState)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetState() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && conn.State != tt.toState {
				t.Errorf("State after SetState() = %v, want %v", conn.State, tt.toState)
			}
		})
	}
}

func TestConnection_StartHandshake(t *testing.T) {
	conn := &Connection{
		ID:               "test",
		State:            StateNew,
		HandshakeTimeout: 100 * time.Millisecond,
		ClientInfo:       make(map[string]interface{}),
	}

	timeoutCalled := false
	err := conn.StartHandshake(func() {
		timeoutCalled = true
	})

	if err != nil {
		t.Fatalf("StartHandshake() error = %v", err)
	}

	if conn.State != StateInitializing {
		t.Errorf("State after StartHandshake() = %v, want StateInitializing", conn.State)
	}

	// Test that handshake can only be started once
	err = conn.StartHandshake(nil)
	if err == nil {
		t.Error("Expected error when starting handshake twice")
	}

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	if !timeoutCalled {
		t.Error("Timeout callback was not called")
	}

	if conn.State != StateClosed {
		t.Errorf("State after timeout = %v, want StateClosed", conn.State)
	}
}

func TestConnection_CompleteHandshake(t *testing.T) {
	conn := &Connection{
		ID:               "test",
		State:            StateNew,
		HandshakeTimeout: 1 * time.Second,
		ClientInfo:       make(map[string]interface{}),
	}

	// Start handshake first
	conn.StartHandshake(nil)

	// Complete handshake
	clientInfo := map[string]interface{}{
		"name":    "test-client",
		"version": "1.0.0",
	}

	err := conn.CompleteHandshake("1.0", clientInfo)
	if err != nil {
		t.Fatalf("CompleteHandshake() error = %v", err)
	}

	if conn.State != StateReady {
		t.Errorf("State after CompleteHandshake() = %v, want StateReady", conn.State)
	}

	if conn.ProtocolVersion != "1.0" {
		t.Errorf("ProtocolVersion = %v, want 1.0", conn.ProtocolVersion)
	}

	if conn.ClientInfo["name"] != "test-client" {
		t.Errorf("ClientInfo name = %v, want test-client", conn.ClientInfo["name"])
	}
}

func TestConnection_IsReady(t *testing.T) {
	conn := &Connection{State: StateNew}

	if conn.IsReady() {
		t.Error("IsReady() = true for StateNew")
	}

	conn.State = StateReady
	if !conn.IsReady() {
		t.Error("IsReady() = false for StateReady")
	}
}

func TestConnection_ConcurrentAccess(t *testing.T) {
	manager := NewManager(10 * time.Second)
	conn, _ := manager.CreateConnection("test")

	// Simulate concurrent access
	var wg sync.WaitGroup
	errors := make([]error, 10)

	// Multiple goroutines trying to start handshake
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			errors[idx] = conn.StartHandshake(nil)
		}(i)
	}

	wg.Wait()

	// Only one should succeed
	successCount := 0
	for _, err := range errors {
		if err == nil {
			successCount++
		}
	}

	if successCount != 1 {
		t.Errorf("Expected exactly 1 successful handshake start, got %d", successCount)
	}
}

func TestConnectionFromContext(t *testing.T) {
	manager := NewManager(10 * time.Second)
	conn, _ := manager.CreateConnection("test-id")

	// Test with connection ID in context
	ctx := WithConnectionID(context.Background(), "test-id")
	retrieved, ok := ConnectionFromContext(ctx, manager)

	if !ok {
		t.Error("ConnectionFromContext() ok = false, want true")
	}

	if retrieved != conn {
		t.Error("ConnectionFromContext() returned different connection")
	}

	// Test without connection ID
	emptyCtx := context.Background()
	_, ok = ConnectionFromContext(emptyCtx, manager)

	if ok {
		t.Error("ConnectionFromContext() ok = true for context without ID")
	}
}

// Benchmarks for connection management performance
func BenchmarkManagerCreateConnection(b *testing.B) {
	manager := NewManager(10 * time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connID := fmt.Sprintf("conn-%d", i)
		_, err := manager.CreateConnection(connID)
		if err != nil {
			b.Fatal("CreateConnection failed:", err)
		}
	}
}

func BenchmarkManagerGetConnection(b *testing.B) {
	manager := NewManager(10 * time.Second)

	// Pre-create connections
	for i := 0; i < 1000; i++ {
		connID := fmt.Sprintf("conn-%d", i)
		manager.CreateConnection(connID)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connID := fmt.Sprintf("conn-%d", i%1000)
		_, exists := manager.GetConnection(connID)
		if !exists {
			b.Fatal("Connection not found")
		}
	}
}

func BenchmarkConnectionStateTransition(b *testing.B) {
	conn := &Connection{
		ID:         "test",
		State:      StateNew,
		ClientInfo: make(map[string]interface{}),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset to initial state
		conn.State = StateNew

		// Perform state transitions
		conn.SetState(StateInitializing)
		conn.SetState(StateReady)
		conn.SetState(StateClosed)
	}
}

func BenchmarkConnectionConcurrentAccess(b *testing.B) {
	manager := NewManager(10 * time.Second)

	// Pre-create connections
	for i := 0; i < 100; i++ {
		connID := fmt.Sprintf("conn-%d", i)
		manager.CreateConnection(connID)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			connID := fmt.Sprintf("conn-%d", i%100)
			conn, exists := manager.GetConnection(connID)
			if !exists {
				b.Fatal("Connection not found")
			}

			// Simulate concurrent state access
			_ = conn.GetState()
			_ = conn.IsReady()
			i++
		}
	})
}

func BenchmarkConnectionHandshakeFlow(b *testing.B) {
	manager := NewManager(10 * time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connID := fmt.Sprintf("conn-%d", i)
		conn, err := manager.CreateConnection(connID)
		if err != nil {
			b.Fatal("CreateConnection failed:", err)
		}

		// Start handshake
		err = conn.StartHandshake(nil)
		if err != nil {
			b.Fatal("StartHandshake failed:", err)
		}

		// Complete handshake
		clientInfo := map[string]interface{}{
			"name":    "test-client",
			"version": "1.0.0",
		}
		err = conn.CompleteHandshake("1.0", clientInfo)
		if err != nil {
			b.Fatal("CompleteHandshake failed:", err)
		}

		// Clean up
		manager.RemoveConnection(connID)
	}
}
