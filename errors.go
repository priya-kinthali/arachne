package main

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// ScraperError represents a scraper-specific error
type ScraperError struct {
	URL         string
	StatusCode  int
	Message     string
	Retryable   bool
	Attempts    int
	LastAttempt time.Time
	Err         error
}

// Error implements the error interface
func (e *ScraperError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("scraper error for %s: %s (status: %d, retryable: %t, attempts: %d): %v",
			e.URL, e.Message, e.StatusCode, e.Retryable, e.Attempts, e.Err)
	}
	return fmt.Sprintf("scraper error for %s: %s (status: %d, retryable: %t, attempts: %d)",
		e.URL, e.Message, e.StatusCode, e.Retryable, e.Attempts)
}

// Unwrap returns the underlying error
func (e *ScraperError) Unwrap() error {
	return e.Err
}

// IsRetryable returns true if the error is retryable
func (e *ScraperError) IsRetryable() bool {
	return e.Retryable
}

// NewScraperError creates a new scraper error
func NewScraperError(url string, message string, err error) *ScraperError {
	return &ScraperError{
		URL:       url,
		Message:   message,
		Retryable: isRetryableError(err),
		Err:       err,
	}
}

// NewHTTPError creates a new HTTP-specific scraper error
func NewHTTPError(url string, statusCode int, message string) *ScraperError {
	return &ScraperError{
		URL:        url,
		StatusCode: statusCode,
		Message:    message,
		Retryable:  isRetryableStatusCode(statusCode),
	}
}

// isRetryableError determines if an error is retryable
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for network-related errors that are typically retryable
	errorString := err.Error()
	retryablePatterns := []string{
		"timeout",
		"connection refused",
		"no route to host",
		"network is unreachable",
		"connection reset",
		"broken pipe",
		"EOF",
	}

	for _, pattern := range retryablePatterns {
		if contains(errorString, pattern) {
			return true
		}
	}

	return false
}

// isRetryableStatusCode determines if an HTTP status code is retryable
func isRetryableStatusCode(statusCode int) bool {
	// Retryable status codes
	retryableCodes := map[int]bool{
		408: true, // Request Timeout
		429: true, // Too Many Requests
		500: true, // Internal Server Error
		502: true, // Bad Gateway
		503: true, // Service Unavailable
		504: true, // Gateway Timeout
	}

	return retryableCodes[statusCode]
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// ValidateURL validates if a URL is properly formatted
func ValidateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %s", urlStr)
	}

	if parsedURL.Scheme == "" {
		return fmt.Errorf("URL must have a scheme (http:// or https://): %s", urlStr)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must have a host: %s", urlStr)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https: %s", urlStr)
	}

	return nil
}

// IsTimeoutError checks if an error is a timeout error
func IsTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	return contains(err.Error(), "timeout")
}

// IsConnectionError checks if an error is a connection error
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}
	errorString := err.Error()
	connectionPatterns := []string{
		"connection refused",
		"no route to host",
		"network is unreachable",
		"connection reset",
		"broken pipe",
	}

	for _, pattern := range connectionPatterns {
		if contains(errorString, pattern) {
			return true
		}
	}
	return false
}

// GetErrorType categorizes the type of error
func GetErrorType(err error) string {
	if err == nil {
		return "none"
	}

	if IsTimeoutError(err) {
		return "timeout"
	}

	if IsConnectionError(err) {
		return "connection"
	}

	if scraperErr, ok := err.(*ScraperError); ok {
		if scraperErr.StatusCode > 0 {
			return fmt.Sprintf("http_%d", scraperErr.StatusCode)
		}
	}

	return "unknown"
}
