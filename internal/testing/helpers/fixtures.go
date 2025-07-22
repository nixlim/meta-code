// Package helpers provides fixture management utilities for testing
package helpers

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// FixtureManager manages test fixtures with caching and templating
type FixtureManager struct {
	basePaths  []string
	cache      map[string][]byte
	cacheMu    sync.RWMutex
	templates  map[string]*template.Template
	templateMu sync.RWMutex
	t          *testing.T
}

// NewFixtureManager creates a new fixture manager
func NewFixtureManager(t *testing.T, basePaths ...string) *FixtureManager {
	// Default base paths if none provided
	if len(basePaths) == 0 {
		basePaths = []string{
			"internal/testing/fixtures",
			"fixtures",
			"testdata",
		}
	}

	return &FixtureManager{
		basePaths: basePaths,
		cache:     make(map[string][]byte),
		templates: make(map[string]*template.Template),
		t:         t,
	}
}

// Load loads a fixture file by name
func (fm *FixtureManager) Load(name string) []byte {
	fm.t.Helper()

	// Check cache first
	if data := fm.getFromCache(name); data != nil {
		return data
	}

	// Try to load from disk
	data, err := fm.loadFromDisk(name)
	require.NoError(fm.t, err, "Failed to load fixture: %s", name)

	// Cache the data
	fm.putInCache(name, data)

	return data
}

// LoadString loads a fixture as a string
func (fm *FixtureManager) LoadString(name string) string {
	return string(fm.Load(name))
}

// LoadJSON loads and unmarshals a JSON fixture
func (fm *FixtureManager) LoadJSON(name string, v interface{}) {
	fm.t.Helper()

	data := fm.Load(name)
	err := json.Unmarshal(data, v)
	require.NoError(fm.t, err, "Failed to unmarshal JSON fixture: %s", name)
}

// LoadYAML loads and unmarshals a YAML fixture
func (fm *FixtureManager) LoadYAML(name string, v interface{}) {
	fm.t.Helper()

	data := fm.Load(name)
	err := yaml.Unmarshal(data, v)
	require.NoError(fm.t, err, "Failed to unmarshal YAML fixture: %s", name)
}

// LoadTemplate loads a fixture as a template and executes it
func (fm *FixtureManager) LoadTemplate(name string, data interface{}) string {
	fm.t.Helper()

	// Check template cache
	tmpl := fm.getTemplate(name)
	if tmpl == nil {
		// Load and parse template
		content := fm.LoadString(name)
		var err error
		tmpl, err = template.New(name).Parse(content)
		require.NoError(fm.t, err, "Failed to parse template: %s", name)

		// Cache the template
		fm.putTemplate(name, tmpl)
	}

	// Execute template
	var buf strings.Builder
	err := tmpl.Execute(&buf, data)
	require.NoError(fm.t, err, "Failed to execute template: %s", name)

	return buf.String()
}

// LoadMultiple loads multiple fixtures at once
func (fm *FixtureManager) LoadMultiple(names ...string) map[string][]byte {
	result := make(map[string][]byte)
	for _, name := range names {
		result[name] = fm.Load(name)
	}
	return result
}

// Exists checks if a fixture exists
func (fm *FixtureManager) Exists(name string) bool {
	_, err := fm.findFixturePath(name)
	return err == nil
}

// AddBasePath adds a new base path to search for fixtures
func (fm *FixtureManager) AddBasePath(path string) {
	fm.basePaths = append(fm.basePaths, path)
}

// ClearCache clears the fixture cache
func (fm *FixtureManager) ClearCache() {
	fm.cacheMu.Lock()
	defer fm.cacheMu.Unlock()
	fm.cache = make(map[string][]byte)

	fm.templateMu.Lock()
	defer fm.templateMu.Unlock()
	fm.templates = make(map[string]*template.Template)
}

// loadFromDisk loads a fixture from disk
func (fm *FixtureManager) loadFromDisk(name string) ([]byte, error) {
	path, err := fm.findFixturePath(name)
	if err != nil {
		return nil, err
	}

	return os.ReadFile(path)
}

// findFixturePath finds the path to a fixture file
func (fm *FixtureManager) findFixturePath(name string) (string, error) {
	// Try exact name first
	for _, basePath := range fm.basePaths {
		path := filepath.Join(basePath, name)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// Try with common extensions
	extensions := []string{".json", ".yaml", ".yml", ".txt", ".xml", ".tmpl"}
	for _, ext := range extensions {
		if strings.HasSuffix(name, ext) {
			continue // Already has extension
		}

		for _, basePath := range fm.basePaths {
			path := filepath.Join(basePath, name+ext)
			if _, err := os.Stat(path); err == nil {
				return path, nil
			}
		}
	}

	return "", fmt.Errorf("fixture not found: %s", name)
}

// getFromCache retrieves data from cache
func (fm *FixtureManager) getFromCache(name string) []byte {
	fm.cacheMu.RLock()
	defer fm.cacheMu.RUnlock()

	if data, ok := fm.cache[name]; ok {
		// Return a copy to prevent mutations
		result := make([]byte, len(data))
		copy(result, data)
		return result
	}

	return nil
}

// putInCache stores data in cache
func (fm *FixtureManager) putInCache(name string, data []byte) {
	fm.cacheMu.Lock()
	defer fm.cacheMu.Unlock()

	// Store a copy to prevent mutations
	cacheCopy := make([]byte, len(data))
	copy(cacheCopy, data)
	fm.cache[name] = cacheCopy
}

// getTemplate retrieves a template from cache
func (fm *FixtureManager) getTemplate(name string) *template.Template {
	fm.templateMu.RLock()
	defer fm.templateMu.RUnlock()
	return fm.templates[name]
}

// putTemplate stores a template in cache
func (fm *FixtureManager) putTemplate(name string, tmpl *template.Template) {
	fm.templateMu.Lock()
	defer fm.templateMu.Unlock()
	fm.templates[name] = tmpl
}

// FixtureLoader provides simplified fixture loading functions
type FixtureLoader struct {
	manager *FixtureManager
}

// NewFixtureLoader creates a new fixture loader
func NewFixtureLoader(t *testing.T, basePaths ...string) *FixtureLoader {
	return &FixtureLoader{
		manager: NewFixtureManager(t, basePaths...),
	}
}

// JSON loads a JSON fixture into the provided interface
func (fl *FixtureLoader) JSON(name string, v interface{}) {
	fl.manager.LoadJSON(name, v)
}

// YAML loads a YAML fixture into the provided interface
func (fl *FixtureLoader) YAML(name string, v interface{}) {
	fl.manager.LoadYAML(name, v)
}

// String loads a fixture as a string
func (fl *FixtureLoader) String(name string) string {
	return fl.manager.LoadString(name)
}

// Bytes loads a fixture as bytes
func (fl *FixtureLoader) Bytes(name string) []byte {
	return fl.manager.Load(name)
}

// Template loads and executes a template fixture
func (fl *FixtureLoader) Template(name string, data interface{}) string {
	return fl.manager.LoadTemplate(name, data)
}

// FixtureSet represents a collection of related fixtures
type FixtureSet struct {
	name     string
	fixtures map[string]interface{}
	manager  *FixtureManager
	t        *testing.T
}

// NewFixtureSet creates a new fixture set
func NewFixtureSet(t *testing.T, name string) *FixtureSet {
	return &FixtureSet{
		name:     name,
		fixtures: make(map[string]interface{}),
		manager:  NewFixtureManager(t),
		t:        t,
	}
}

// Load loads all fixtures in the set
func (fs *FixtureSet) Load(patterns ...string) {
	fs.t.Helper()

	// Default pattern if none provided
	if len(patterns) == 0 {
		patterns = []string{fmt.Sprintf("%s/*.json", fs.name)}
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		require.NoError(fs.t, err, "Failed to glob pattern: %s", pattern)

		for _, match := range matches {
			base := filepath.Base(match)
			ext := filepath.Ext(base)
			name := strings.TrimSuffix(base, ext)

			data, err := os.ReadFile(match)
			require.NoError(fs.t, err, "Failed to read fixture: %s", match)

			// Parse based on extension
			var value interface{}
			switch ext {
			case ".json":
				err = json.Unmarshal(data, &value)
			case ".yaml", ".yml":
				err = yaml.Unmarshal(data, &value)
			default:
				value = string(data)
			}

			if err != nil {
				fs.t.Errorf("Failed to parse fixture %s: %v", match, err)
				continue
			}

			fs.fixtures[name] = value
		}
	}
}

// Get retrieves a fixture from the set
func (fs *FixtureSet) Get(name string) interface{} {
	return fs.fixtures[name]
}

// GetString retrieves a fixture as a string
func (fs *FixtureSet) GetString(name string) string {
	if v, ok := fs.fixtures[name].(string); ok {
		return v
	}
	fs.t.Fatalf("Fixture %s is not a string", name)
	return ""
}

// GetJSON retrieves a fixture and marshals it to JSON
func (fs *FixtureSet) GetJSON(name string) string {
	fs.t.Helper()

	v := fs.Get(name)
	data, err := json.Marshal(v)
	require.NoError(fs.t, err, "Failed to marshal fixture to JSON: %s", name)

	return string(data)
}

// All returns all fixtures in the set
func (fs *FixtureSet) All() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range fs.fixtures {
		result[k] = v
	}
	return result
}

// FixtureWriter helps create fixtures from test data
type FixtureWriter struct {
	basePath string
	t        *testing.T
}

// NewFixtureWriter creates a new fixture writer
func NewFixtureWriter(t *testing.T, basePath string) *FixtureWriter {
	return &FixtureWriter{
		basePath: basePath,
		t:        t,
	}
}

// WriteJSON writes data as a JSON fixture
func (fw *FixtureWriter) WriteJSON(name string, data interface{}) {
	fw.t.Helper()

	content, err := json.MarshalIndent(data, "", "  ")
	require.NoError(fw.t, err, "Failed to marshal data to JSON")

	fw.writeFile(name+".json", content)
}

// WriteYAML writes data as a YAML fixture
func (fw *FixtureWriter) WriteYAML(name string, data interface{}) {
	fw.t.Helper()

	content, err := yaml.Marshal(data)
	require.NoError(fw.t, err, "Failed to marshal data to YAML")

	fw.writeFile(name+".yaml", content)
}

// WriteString writes a string as a fixture
func (fw *FixtureWriter) WriteString(name string, content string) {
	fw.writeFile(name, []byte(content))
}

// writeFile writes content to a file
func (fw *FixtureWriter) writeFile(name string, content []byte) {
	fw.t.Helper()

	path := filepath.Join(fw.basePath, name)
	dir := filepath.Dir(path)

	// Ensure directory exists
	err := os.MkdirAll(dir, 0755)
	require.NoError(fw.t, err, "Failed to create directory: %s", dir)

	// Write file
	err = os.WriteFile(path, content, 0644)
	require.NoError(fw.t, err, "Failed to write fixture: %s", path)

	fw.t.Logf("Wrote fixture: %s", path)
}

// FixtureGenerator generates fixtures from various sources
type FixtureGenerator struct {
	t *testing.T
}

// NewFixtureGenerator creates a new fixture generator
func NewFixtureGenerator(t *testing.T) *FixtureGenerator {
	return &FixtureGenerator{t: t}
}

// FromReader generates a fixture from a reader
func (fg *FixtureGenerator) FromReader(r io.Reader) []byte {
	fg.t.Helper()

	data, err := io.ReadAll(r)
	require.NoError(fg.t, err, "Failed to read from reader")

	return data
}

// FromHTTPResponse generates a fixture from an HTTP response (mock)
func (fg *FixtureGenerator) FromHTTPResponse(statusCode int, headers map[string]string, body string) map[string]interface{} {
	return map[string]interface{}{
		"statusCode": statusCode,
		"headers":    headers,
		"body":       body,
	}
}

// FromError generates a fixture from an error
func (fg *FixtureGenerator) FromError(err error) map[string]interface{} {
	if err == nil {
		return nil
	}

	return map[string]interface{}{
		"error": err.Error(),
		"type":  fmt.Sprintf("%T", err),
	}
}

// FromStruct generates a fixture from a struct with field filtering
func (fg *FixtureGenerator) FromStruct(v interface{}, includeFields ...string) map[string]interface{} {
	fg.t.Helper()

	// Convert to JSON and back to get a map
	data, err := json.Marshal(v)
	require.NoError(fg.t, err, "Failed to marshal struct")

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(fg.t, err, "Failed to unmarshal to map")

	// Filter fields if specified
	if len(includeFields) > 0 {
		filtered := make(map[string]interface{})
		for _, field := range includeFields {
			if value, ok := result[field]; ok {
				filtered[field] = value
			}
		}
		return filtered
	}

	return result
}

// Global fixture functions for convenience

var globalFixtureManager *FixtureManager
var globalFixtureMu sync.Mutex

// getGlobalFixtureManager gets or creates the global fixture manager
func getGlobalFixtureManager(t *testing.T) *FixtureManager {
	globalFixtureMu.Lock()
	defer globalFixtureMu.Unlock()

	if globalFixtureManager == nil {
		globalFixtureManager = NewFixtureManager(t)
	}

	return globalFixtureManager
}

// LoadFixtureGlobal loads a fixture using the global manager
func LoadFixtureGlobal(t *testing.T, name string) []byte {
	t.Helper()
	return getGlobalFixtureManager(t).Load(name)
}

// LoadFixtureStringGlobal loads a fixture as string using the global manager
func LoadFixtureStringGlobal(t *testing.T, name string) string {
	t.Helper()
	return getGlobalFixtureManager(t).LoadString(name)
}

// LoadFixtureJSONGlobal loads a JSON fixture using the global manager
func LoadFixtureJSONGlobal(t *testing.T, name string, v interface{}) {
	t.Helper()
	getGlobalFixtureManager(t).LoadJSON(name, v)
}

// LoadFixtureTemplateGlobal loads and executes a template using the global manager
func LoadFixtureTemplateGlobal(t *testing.T, name string, data interface{}) string {
	t.Helper()
	return getGlobalFixtureManager(t).LoadTemplate(name, data)
}
