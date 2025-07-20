package logging

import (
	"io"
	"os"
	"strings"
)

// Environment constants
const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
	EnvStaging     = "staging"
)

// ConfigFromEnv creates a Config based on environment variables
func ConfigFromEnv() Config {
	cfg := Config{
		Output:    os.Stderr,
		Level:     LogLevelInfo,
		DebugMode: false,
		Sanitize:  true,
		Pretty:    false,
	}

	// Check environment
	env := strings.ToLower(os.Getenv("ENVIRONMENT"))
	if env == "" {
		env = strings.ToLower(os.Getenv("ENV"))
	}
	if env == "" {
		env = strings.ToLower(os.Getenv("GO_ENV"))
	}

	// Set defaults based on environment
	switch env {
	case EnvDevelopment, "dev", "local":
		cfg.Pretty = true
		cfg.DebugMode = true
		cfg.Level = LogLevelDebug
		cfg.Sanitize = false
	case EnvStaging, "stage":
		cfg.Level = LogLevelDebug
		cfg.DebugMode = true
	case EnvProduction, "prod":
		// Keep defaults
	default:
		// If no environment is set, check if we're in a TTY (development)
		if fileInfo, _ := os.Stderr.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
			cfg.Pretty = true
			cfg.DebugMode = true
			cfg.Level = LogLevelDebug
		}
	}

	// Override with specific environment variables
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.Level = ParseLogLevel(logLevel)
	}

	if debugMode := os.Getenv("DEBUG"); debugMode != "" {
		cfg.DebugMode = strings.ToLower(debugMode) == "true" || debugMode == "1"
	}

	if pretty := os.Getenv("LOG_PRETTY"); pretty != "" {
		cfg.Pretty = strings.ToLower(pretty) == "true" || pretty == "1"
	}

	if sanitize := os.Getenv("LOG_SANITIZE"); sanitize != "" {
		cfg.Sanitize = strings.ToLower(sanitize) == "true" || sanitize == "1"
	}

	return cfg
}

// ParseLogLevel parses a string log level into a LogLevel
func ParseLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug", "trace":
		return LogLevelDebug
	case "info":
		return LogLevelInfo
	case "warn", "warning":
		return LogLevelWarn
	case "error":
		return LogLevelError
	case "fatal", "panic":
		return LogLevelFatal
	default:
		return LogLevelInfo
	}
}

// DevelopmentConfig returns a configuration suitable for development
func DevelopmentConfig() Config {
	return Config{
		Output:    os.Stderr,
		Level:     LogLevelDebug,
		DebugMode: true,
		Sanitize:  false,
		Pretty:    true,
	}
}

// ProductionConfig returns a configuration suitable for production
func ProductionConfig() Config {
	return Config{
		Output:    os.Stderr,
		Level:     LogLevelInfo,
		DebugMode: false,
		Sanitize:  true,
		Pretty:    false,
	}
}

// TestConfig returns a configuration suitable for testing
func TestConfig(output io.Writer) Config {
	if output == nil {
		output = os.Stderr
	}
	return Config{
		Output:    output,
		Level:     LogLevelDebug,
		DebugMode: true,
		Sanitize:  false,
		Pretty:    false, // JSON output for easier parsing in tests
	}
}
