package conformance

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/meta-mcp/meta-mcp-server/internal/protocol/validator"
)

// TestRunFullConformanceSuite runs the complete MCP protocol conformance test suite
func TestRunFullConformanceSuite(t *testing.T) {
	// Create validator
	v, err := validator.New(validator.Config{
		Enabled:      true,
		CacheSchemas: true,
		StrictMode:   true,
	})
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Create conformance test suite
	suite := NewConformanceTestSuite(v)

	// Run all tests
	startTime := time.Now()
	suite.RunAll(t)
	duration := time.Since(startTime)

	// Generate report
	baseReport := suite.GenerateReport()
	report := &EnhancedConformanceReport{
		ConformanceReport: baseReport,
		TestDuration:      duration.String(),
		Timestamp:         time.Now().UTC().Format(time.RFC3339),
		ValidatorConfig: validator.Config{
			Enabled:      true,
			CacheSchemas: true,
			StrictMode:   true,
		},
	}

	// Print summary
	fmt.Printf("\n=== MCP Protocol Conformance Test Results ===\n")
	fmt.Printf("Total Tests: %d\n", report.TotalTests)
	fmt.Printf("Passed: %d (%.1f%%)\n", report.PassedTests,
		float64(report.PassedTests)/float64(report.TotalTests)*100)
	fmt.Printf("Failed: %d\n", report.FailedTests)
	fmt.Printf("Duration: %s\n", report.TestDuration)
	fmt.Printf("\n")

	// Print category breakdown
	fmt.Printf("Category Breakdown:\n")
	for category, summary := range report.Categories {
		fmt.Printf("  %s: %d/%d passed (%.1f%%)\n",
			category, summary.PassedTests, summary.TotalTests,
			float64(summary.PassedTests)/float64(summary.TotalTests)*100)
	}
	fmt.Printf("\n")

	// Print failed tests
	if report.FailedTests > 0 {
		fmt.Printf("Failed Tests:\n")
		for _, result := range report.TestResults {
			if !result.Passed {
				fmt.Printf("  - [%s] %s: %s\n", result.Category, result.TestName, result.Error)
			}
		}
		fmt.Printf("\n")
	}

	// Save detailed report to file
	if os.Getenv("SAVE_CONFORMANCE_REPORT") == "true" {
		reportPath := fmt.Sprintf("conformance_report_%s.json",
			time.Now().Format("20060102_150405"))
		if err := saveReport(report, reportPath); err != nil {
			t.Errorf("Failed to save conformance report: %v", err)
		} else {
			fmt.Printf("Detailed report saved to: %s\n", reportPath)
		}
	}

	// Fail test if any conformance tests failed
	if report.FailedTests > 0 {
		t.Errorf("Conformance test suite failed: %d tests failed", report.FailedTests)
	}
}

// EnhancedConformanceReport extends ConformanceReport with additional fields
type EnhancedConformanceReport struct {
	*ConformanceReport
	Timestamp       string           `json:"timestamp"`
	TestDuration    string           `json:"testDuration"`
	ValidatorConfig validator.Config `json:"validatorConfig"`
}

// saveReport saves the conformance report to a JSON file
func saveReport(report *EnhancedConformanceReport, path string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// Benchmark validation performance
func BenchmarkValidation(b *testing.B) {
	v, err := validator.New(validator.Config{
		Enabled:      true,
		CacheSchemas: true,
	})
	if err != nil {
		b.Fatalf("Failed to create validator: %v", err)
	}

	ctx := context.Background()

	// Prepare test messages
	initializeMsg := json.RawMessage(`{
		"protocolVersion": "1.0",
		"capabilities": {},
		"clientInfo": {"name": "bench-client", "version": "1.0.0"}
	}`)

	requestMsg := json.RawMessage(`{
		"jsonrpc": "2.0",
		"method": "tools/call",
		"params": {"name": "test", "arguments": {}},
		"id": "123"
	}`)

	b.Run("ValidateInitialize", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = v.ValidateMessage(ctx, "initialize", initializeMsg)
		}
	})

	b.Run("ValidateRequest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = v.ValidateMessage(ctx, "request", requestMsg)
		}
	})

	b.Run("ValidateWithDisabledValidator", func(b *testing.B) {
		disabledValidator, _ := validator.New(validator.Config{Enabled: false})
		for i := 0; i < b.N; i++ {
			_ = disabledValidator.ValidateMessage(ctx, "request", requestMsg)
		}
	})
}
