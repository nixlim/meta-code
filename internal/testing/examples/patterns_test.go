package examples

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Example 1: Table-Driven Test Pattern
func TestTableDrivenExample(t *testing.T) {
	// Define test cases in a table
	tests := []struct {
		name    string
		input   int
		want    int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "positive_number",
			input:   5,
			want:    25,
			wantErr: false,
		},
		{
			name:    "zero",
			input:   0,
			want:    0,
			wantErr: false,
		},
		{
			name:    "negative_number",
			input:   -5,
			want:    0,
			wantErr: true,
			errMsg:  "negative input not allowed",
		},
		{
			name:    "large_number",
			input:   1000,
			want:    1000000,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call function under test
			got, err := square(tt.input)

			// Check error expectations
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			// Check success case
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Example 2: Mock Usage Pattern
type DataStore interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string) error
}

type MockDataStore struct {
	data   map[string]string
	errors map[string]error
	calls  []string
	mu     sync.Mutex
}

func NewMockDataStore() *MockDataStore {
	return &MockDataStore{
		data:   make(map[string]string),
		errors: make(map[string]error),
		calls:  []string{},
	}
}

func (m *MockDataStore) Get(ctx context.Context, key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls = append(m.calls, fmt.Sprintf("Get:%s", key))

	if err, ok := m.errors[key]; ok {
		return "", err
	}

	value, ok := m.data[key]
	if !ok {
		return "", errors.New("key not found")
	}

	return value, nil
}

func (m *MockDataStore) Set(ctx context.Context, key string, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls = append(m.calls, fmt.Sprintf("Set:%s=%s", key, value))

	if err, ok := m.errors[key]; ok {
		return err
	}

	m.data[key] = value
	return nil
}

func (m *MockDataStore) GetCalls() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string{}, m.calls...)
}

func TestServiceWithMock(t *testing.T) {
	// Create and configure mock
	mock := NewMockDataStore()
	mock.data["existing"] = "value"
	mock.errors["error-key"] = errors.New("simulated error")

	// Create service with mock
	service := NewService(mock)

	t.Run("successful_operation", func(t *testing.T) {
		ctx := context.Background()
		result, err := service.ProcessData(ctx, "existing")

		require.NoError(t, err)
		assert.Equal(t, "processed: value", result)

		// Verify mock was called correctly
		calls := mock.GetCalls()
		assert.Contains(t, calls, "Get:existing")
	})

	t.Run("error_handling", func(t *testing.T) {
		ctx := context.Background()
		_, err := service.ProcessData(ctx, "error-key")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "simulated error")
	})
}

// Example 3: Fixture-Based Testing
type TestFixture struct {
	ID     string          `json:"id"`
	Name   string          `json:"name"`
	Config json.RawMessage `json:"config"`
}

func TestWithFixtures(t *testing.T) {
	// Load fixture data
	fixtures := []TestFixture{
		{
			ID:     "test-1",
			Name:   "Basic Test",
			Config: json.RawMessage(`{"enabled": true}`),
		},
		{
			ID:     "test-2",
			Name:   "Advanced Test",
			Config: json.RawMessage(`{"enabled": true, "level": "high"}`),
		},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			// Parse config
			var config map[string]interface{}
			err := json.Unmarshal(fixture.Config, &config)
			require.NoError(t, err)

			// Validate fixture
			assert.NotEmpty(t, fixture.ID)
			assert.True(t, config["enabled"].(bool))
		})
	}
}

// Example 4: Integration Test Pattern
func TestIntegrationWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup phase
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	server := setupTestServer(t)
	defer server.Close()

	client := setupTestClient(t, server.URL)
	defer client.Close()

	// Test workflow
	t.Run("complete_workflow", func(t *testing.T) {
		// Step 1: Initialize
		t.Run("initialize", func(t *testing.T) {
			err := client.Initialize(ctx)
			require.NoError(t, err)
			assert.Equal(t, "ready", client.State())
		})

		// Step 2: Perform operations
		t.Run("operations", func(t *testing.T) {
			// Create resource
			resourceID, err := client.CreateResource(ctx, "test-resource")
			require.NoError(t, err)
			assert.NotEmpty(t, resourceID)

			// Update resource
			err = client.UpdateResource(ctx, resourceID, map[string]string{
				"status": "active",
			})
			require.NoError(t, err)

			// Verify resource
			resource, err := client.GetResource(ctx, resourceID)
			require.NoError(t, err)
			assert.Equal(t, "active", resource.Status)
		})

		// Step 3: Cleanup
		t.Run("cleanup", func(t *testing.T) {
			err := client.Shutdown(ctx)
			require.NoError(t, err)
		})
	})
}

// Example 5: Benchmark Pattern
func BenchmarkDataProcessing(b *testing.B) {
	// Setup data
	testData := generateTestData(1000)

	// Reset timer to exclude setup
	b.ResetTimer()

	// Run benchmark
	for i := 0; i < b.N; i++ {
		result := processData(testData)
		if len(result) == 0 {
			b.Fatal("Processing returned empty result")
		}
	}
}

func BenchmarkConcurrentProcessing(b *testing.B) {
	testData := generateTestData(100)

	b.RunParallel(func(pb *testing.PB) {
		// Each goroutine gets its own copy
		localData := make([]string, len(testData))
		copy(localData, testData)

		for pb.Next() {
			result := processData(localData)
			if len(result) == 0 {
				b.Fatal("Processing returned empty result")
			}
		}
	})
}

// Example 6: Concurrent Testing Pattern
func TestConcurrentOperations(t *testing.T) {
	manager := NewConcurrentManager()

	const numWorkers = 50
	const opsPerWorker = 100

	var wg sync.WaitGroup
	errors := make(chan error, numWorkers)

	// Launch concurrent workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < opsPerWorker; j++ {
				key := fmt.Sprintf("worker-%d-op-%d", workerID, j)

				// Perform concurrent operations
				if err := manager.Store(key, workerID); err != nil {
					errors <- fmt.Errorf("store failed: %w", err)
					return
				}

				value, err := manager.Load(key)
				if err != nil {
					errors <- fmt.Errorf("load failed: %w", err)
					return
				}

				if value != workerID {
					errors <- fmt.Errorf("value mismatch: got %v, want %v", value, workerID)
					return
				}

				if err := manager.Delete(key); err != nil {
					errors <- fmt.Errorf("delete failed: %w", err)
					return
				}
			}
		}(i)
	}

	// Wait for completion
	wg.Wait()
	close(errors)

	// Check for errors
	var errorCount int
	for err := range errors {
		errorCount++
		t.Errorf("Concurrent operation error: %v", err)
	}

	if errorCount > 0 {
		t.Fatalf("Had %d errors in concurrent operations", errorCount)
	}

	// Verify final state
	assert.Equal(t, 0, manager.Size(), "Manager should be empty after all operations")
}

// Example 7: Error Testing Pattern
func TestErrorScenarios(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() error
		operation   func() error
		wantErr     bool
		errContains string
		errType     error
	}{
		{
			name: "timeout_error",
			setup: func() error {
				// Setup that will cause timeout
				return nil
			},
			operation: func() error {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Microsecond)
				defer cancel()

				// Simulate long operation
				time.Sleep(10 * time.Millisecond)

				return ctx.Err()
			},
			wantErr:     true,
			errContains: "context deadline exceeded",
		},
		{
			name: "validation_error",
			operation: func() error {
				return validateInput("")
			},
			wantErr:     true,
			errContains: "input cannot be empty",
			errType:     ErrValidation,
		},
		{
			name: "success_case",
			operation: func() error {
				return validateInput("valid input")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run setup if provided
			if tt.setup != nil {
				err := tt.setup()
				require.NoError(t, err, "Setup failed")
			}

			// Run operation
			err := tt.operation()

			// Check error expectations
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Helper functions for examples
func square(n int) (int, error) {
	if n < 0 {
		return 0, errors.New("negative input not allowed")
	}
	return n * n, nil
}

type Service struct {
	store DataStore
}

func NewService(store DataStore) *Service {
	return &Service{store: store}
}

func (s *Service) ProcessData(ctx context.Context, key string) (string, error) {
	value, err := s.store.Get(ctx, key)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("processed: %s", value), nil
}

func setupTestServer(t *testing.T) *TestServer {
	t.Helper()
	// In a real implementation, this would start an actual test server
	// For this example, we're returning a mock server with a test URL
	return &TestServer{
		URL: "http://localhost:8080",
	}
}

func setupTestClient(t *testing.T, url string) *TestClient {
	t.Helper()
	return &TestClient{url: url, state: "new"}
}

type TestServer struct {
	URL string
}

func (s *TestServer) Close() {}

type TestClient struct {
	url   string
	state string
}

func (c *TestClient) Close()        {}
func (c *TestClient) State() string { return c.state }
func (c *TestClient) Initialize(ctx context.Context) error {
	c.state = "ready"
	return nil
}
func (c *TestClient) CreateResource(ctx context.Context, name string) (string, error) {
	return "resource-123", nil
}
func (c *TestClient) UpdateResource(ctx context.Context, id string, updates map[string]string) error {
	return nil
}
func (c *TestClient) GetResource(ctx context.Context, id string) (*Resource, error) {
	return &Resource{ID: id, Status: "active"}, nil
}
func (c *TestClient) Shutdown(ctx context.Context) error {
	c.state = "closed"
	return nil
}

type Resource struct {
	ID     string
	Status string
}

func generateTestData(n int) []string {
	data := make([]string, n)
	for i := range data {
		data[i] = fmt.Sprintf("item-%d", i)
	}
	return data
}

func processData(data []string) []string {
	result := make([]string, len(data))
	for i, item := range data {
		result[i] = fmt.Sprintf("processed-%s", item)
	}
	return result
}

type ConcurrentManager struct {
	data sync.Map
}

func NewConcurrentManager() *ConcurrentManager {
	return &ConcurrentManager{}
}

func (m *ConcurrentManager) Store(key string, value interface{}) error {
	m.data.Store(key, value)
	return nil
}

func (m *ConcurrentManager) Load(key string) (interface{}, error) {
	value, ok := m.data.Load(key)
	if !ok {
		return nil, errors.New("key not found")
	}
	return value, nil
}

func (m *ConcurrentManager) Delete(key string) error {
	m.data.Delete(key)
	return nil
}

func (m *ConcurrentManager) Size() int {
	count := 0
	m.data.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

var ErrValidation = errors.New("validation error")

func validateInput(input string) error {
	if input == "" {
		return fmt.Errorf("%w: input cannot be empty", ErrValidation)
	}
	return nil
}
