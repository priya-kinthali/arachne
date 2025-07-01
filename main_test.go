package main

import (
	"testing"
	"time"
)

func TestExtractHTMLTitle(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "Valid title",
			html:     "<html><head><title>Test Title</title></head><body>Content</body></html>",
			expected: "Test Title",
		},
		{
			name:     "No title tag",
			html:     "<html><head></head><body>Content</body></html>",
			expected: "No HTML title found",
		},
		{
			name:     "Empty title",
			html:     "<html><head><title></title></head><body>Content</body></html>",
			expected: "Empty HTML title",
		},
		{
			name:     "Malformed title",
			html:     "<html><head><title>Test Title</head><body>Content</body></html>",
			expected: "Malformed HTML title",
		},
		{
			name:     "Case insensitive title",
			html:     "<html><head><TITLE>Test Title</TITLE></head><body>Content</body></html>",
			expected: "Malformed HTML title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractHTMLTitle(tt.html)
			if result != tt.expected {
				t.Errorf("extractHTMLTitle() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractJSONTitle(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected string
	}{
		{
			name:     "Title field",
			json:     `{"title": "Test Title", "content": "test"}`,
			expected: "Test Title",
		},
		{
			name:     "Name field",
			json:     `{"name": "Test Name", "content": "test"}`,
			expected: "Test Name",
		},
		{
			name:     "Login field",
			json:     `{"login": "testuser", "content": "test"}`,
			expected: "testuser",
		},
		{
			name:     "No title fields",
			json:     `{"content": "test", "data": "value"}`,
			expected: "content: test",
		},
		{
			name:     "Invalid JSON",
			json:     `{"title": "test"`,
			expected: "Invalid JSON",
		},
		{
			name:     "Empty JSON",
			json:     `{}`,
			expected: "JSON response (no title field)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJSONTitle(tt.json)
			if result != tt.expected {
				t.Errorf("extractJSONTitle() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		contentType string
		expected    string
	}{
		{
			name:        "JSON content type",
			content:     `{"title": "JSON Title"}`,
			contentType: "application/json",
			expected:    "JSON Title",
		},
		{
			name:        "JSON content without content type",
			content:     `{"title": "JSON Title"}`,
			contentType: "text/plain",
			expected:    "JSON Title",
		},
		{
			name:        "HTML content",
			content:     "<html><head><title>HTML Title</title></head></html>",
			contentType: "text/html",
			expected:    "HTML Title",
		},
		{
			name:        "Array JSON",
			content:     `[{"title": "Array Title"}]`,
			contentType: "application/json",
			expected:    "Invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTitle(tt.content, tt.contentType)
			if result != tt.expected {
				t.Errorf("extractTitle() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "Valid HTTP URL",
			url:     "https://example.com",
			wantErr: false,
		},
		{
			name:    "Valid HTTPS URL",
			url:     "https://api.github.com/users/test",
			wantErr: false,
		},
		{
			name:    "Invalid URL",
			url:     "not-a-url",
			wantErr: false,
		},
		{
			name:    "Empty URL",
			url:     "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "Valid config",
			config: &Config{
				MaxConcurrent:  3,
				RequestTimeout: 10 * time.Second,
				TotalTimeout:   30 * time.Second,
				RetryAttempts:  3,
				RetryDelay:     1 * time.Second,
				LogLevel:       "info",
			},
			wantErr: false,
		},
		{
			name: "Invalid max concurrent",
			config: &Config{
				MaxConcurrent:  0,
				RequestTimeout: 10 * time.Second,
				TotalTimeout:   30 * time.Second,
				RetryAttempts:  3,
				RetryDelay:     1 * time.Second,
				LogLevel:       "info",
			},
			wantErr: true,
		},
		{
			name: "Invalid log level",
			config: &Config{
				MaxConcurrent:  3,
				RequestTimeout: 10 * time.Second,
				TotalTimeout:   30 * time.Second,
				RetryAttempts:  3,
				RetryDelay:     1 * time.Second,
				LogLevel:       "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMetrics(t *testing.T) {
	metrics := NewMetrics()

	// Test initial state
	if metrics.TotalRequests != 0 {
		t.Errorf("Expected 0 total requests, got %d", metrics.TotalRequests)
	}

	// Test recording requests
	metrics.RecordRequest()
	metrics.RecordRequest()
	if metrics.TotalRequests != 2 {
		t.Errorf("Expected 2 total requests, got %d", metrics.TotalRequests)
	}

	// Test recording success
	metrics.RecordSuccess("example.com", 200, 1024, 100*time.Millisecond)
	if metrics.SuccessfulRequests != 1 {
		t.Errorf("Expected 1 successful request, got %d", metrics.SuccessfulRequests)
	}

	// Test recording failure
	metrics.RecordFailure("example.com", 404)
	if metrics.FailedRequests != 1 {
		t.Errorf("Expected 1 failed request, got %d", metrics.FailedRequests)
	}

	// Test success rate calculation
	successRate := metrics.GetSuccessRate()
	expectedRate := 50.0 // 1 success out of 2 total requests
	if successRate != expectedRate {
		t.Errorf("Expected success rate %.1f%%, got %.1f%%", expectedRate, successRate)
	}
}

func TestScraperError(t *testing.T) {
	err := NewScraperError("https://example.com", "Test error", nil)

	if err.URL != "https://example.com" {
		t.Errorf("Expected URL https://example.com, got %s", err.URL)
	}

	if err.Message != "Test error" {
		t.Errorf("Expected message 'Test error', got %s", err.Message)
	}

	// Test HTTP error
	httpErr := NewHTTPError("https://example.com", 500, "Server error")
	if !httpErr.IsRetryable() {
		t.Error("Expected 500 error to be retryable")
	}

	// Test non-retryable error
	nonRetryableErr := NewHTTPError("https://example.com", 404, "Not found")
	if nonRetryableErr.IsRetryable() {
		t.Error("Expected 404 error to not be retryable")
	}
}

func BenchmarkExtractHTMLTitle(b *testing.B) {
	html := `<html><head><title>Benchmark Test Title</title></head><body>Content</body></html>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractHTMLTitle(html)
	}
}

func BenchmarkExtractJSONTitle(b *testing.B) {
	json := `{"title": "Benchmark Test Title", "content": "test content", "data": "more data"}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractJSONTitle(json)
	}
}
