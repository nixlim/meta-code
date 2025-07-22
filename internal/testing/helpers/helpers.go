// Package helpers provides common testing utilities and helper functions
// for the MCP protocol implementation test suite.
package helpers

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHelper provides common testing utilities
type TestHelper struct {
	t *testing.T
}

// New creates a new TestHelper instance
func New(t *testing.T) *TestHelper {
	return &TestHelper{t: t}
}

// LoadFixture loads a JSON fixture file and unmarshals it into the provided interface
func (h *TestHelper) LoadFixture(filename string, v interface{}) {
	h.t.Helper()

	// Look for fixture in multiple possible locations
	possiblePaths := []string{
		filepath.Join("internal", "testing", "fixtures", filename),
		filepath.Join("fixtures", filename),
		filename,
	}

	var data []byte
	var err error
	var foundPath string

	for _, path := range possiblePaths {
		data, err = os.ReadFile(path)
		if err == nil {
			foundPath = path
			break
		}
	}

	require.NoError(h.t, err, "Failed to load fixture file: %s", filename)
	require.NotEmpty(h.t, foundPath, "Fixture file not found: %s", filename)

	err = json.Unmarshal(data, v)
	require.NoError(h.t, err, "Failed to unmarshal fixture: %s", foundPath)
}

// LoadFixtureString loads a fixture file as a string
func (h *TestHelper) LoadFixtureString(filename string) string {
	h.t.Helper()

	possiblePaths := []string{
		filepath.Join("internal", "testing", "fixtures", filename),
		filepath.Join("fixtures", filename),
		filename,
	}

	var data []byte
	var err error

	for _, path := range possiblePaths {
		data, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}

	require.NoError(h.t, err, "Failed to load fixture file: %s", filename)
	return string(data)
}

// AssertJSONEqual compares two JSON strings for equality, ignoring formatting
func (h *TestHelper) AssertJSONEqual(expected, actual string, msgAndArgs ...interface{}) {
	h.t.Helper()

	var expectedObj, actualObj interface{}

	err := json.Unmarshal([]byte(expected), &expectedObj)
	require.NoError(h.t, err, "Failed to unmarshal expected JSON")

	err = json.Unmarshal([]byte(actual), &actualObj)
	require.NoError(h.t, err, "Failed to unmarshal actual JSON")

	assert.Equal(h.t, expectedObj, actualObj, msgAndArgs...)
}

// AssertValidJSON checks if a string is valid JSON
func (h *TestHelper) AssertValidJSON(jsonStr string, msgAndArgs ...interface{}) {
	h.t.Helper()

	var obj interface{}
	err := json.Unmarshal([]byte(jsonStr), &obj)
	assert.NoError(h.t, err, msgAndArgs...)
}

// CaptureOutput captures stdout/stderr during test execution
func (h *TestHelper) CaptureOutput(fn func()) (stdout, stderr string) {
	h.t.Helper()

	// This is a simplified version - in a real implementation,
	// you might want to use more sophisticated output capture
	fn()
	return "", ""
}

// TableTest represents a table-driven test case
type TableTest struct {
	Name     string
	Input    interface{}
	Expected interface{}
	Error    string
	Setup    func(*testing.T)
	Cleanup  func(*testing.T)
}

// RunTableTests executes a slice of table-driven tests
func (h *TestHelper) RunTableTests(tests []TableTest, testFunc func(*testing.T, TableTest)) {
	h.t.Helper()

	for _, tt := range tests {
		h.t.Run(tt.Name, func(t *testing.T) {
			if tt.Setup != nil {
				tt.Setup(t)
			}

			if tt.Cleanup != nil {
				defer tt.Cleanup(t)
			}

			testFunc(t, tt)
		})
	}
}

// Global helper functions that don't require a TestHelper instance

// AssertJSONEqualGlobal compares JSON strings globally
func AssertJSONEqualGlobal(t *testing.T, expected, actual string, msgAndArgs ...interface{}) {
	t.Helper()
	helper := New(t)
	helper.AssertJSONEqual(expected, actual, msgAndArgs...)
}

// AssertValidJSONGlobal checks JSON validity globally
func AssertValidJSONGlobal(t *testing.T, jsonStr string, msgAndArgs ...interface{}) {
	t.Helper()
	helper := New(t)
	helper.AssertValidJSON(jsonStr, msgAndArgs...)
}

// MustMarshalJSON marshals an object to JSON or fails the test
func MustMarshalJSON(t *testing.T, v interface{}) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	require.NoError(t, err, "Failed to marshal JSON")
	return data
}

// MustUnmarshalJSON unmarshals JSON or fails the test
func MustUnmarshalJSON(t *testing.T, data []byte, v interface{}) {
	t.Helper()
	err := json.Unmarshal(data, v)
	require.NoError(t, err, "Failed to unmarshal JSON")
}

// CreateTempFile creates a temporary file for testing
func CreateTempFile(t *testing.T, content string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "test-*.json")
	require.NoError(t, err, "Failed to create temp file")

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err, "Failed to write to temp file")

	err = tmpFile.Close()
	require.NoError(t, err, "Failed to close temp file")

	// Clean up the file when test completes
	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})

	return tmpFile.Name()
}

// ReadAllString reads all content from a reader as string
func ReadAllString(t *testing.T, r io.Reader) string {
	t.Helper()

	data, err := io.ReadAll(r)
	require.NoError(t, err, "Failed to read from reader")

	return string(data)
}

// SkipIfShort skips the test if running in short mode
func SkipIfShort(t *testing.T, reason string) {
	if testing.Short() {
		t.Skipf("Skipping test in short mode: %s", reason)
	}
}

// RequireNoError is a convenience wrapper for require.NoError with better error messages
func RequireNoError(t *testing.T, err error, operation string, args ...interface{}) {
	t.Helper()
	if len(args) > 0 {
		operation = fmt.Sprintf(operation, args...)
	}
	require.NoError(t, err, "Operation failed: %s", operation)
}

// AssertNoError is a convenience wrapper for assert.NoError with better error messages
func AssertNoError(t *testing.T, err error, operation string, args ...interface{}) bool {
	t.Helper()
	if len(args) > 0 {
		operation = fmt.Sprintf(operation, args...)
	}
	return assert.NoError(t, err, "Operation failed: %s", operation)
}
