package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the scraper
type Config struct {
	MaxConcurrent           int            `json:"max_concurrent"`
	RequestTimeout          time.Duration  `json:"request_timeout"`
	TotalTimeout            time.Duration  `json:"total_timeout"`
	UserAgent               string         `json:"user_agent"`
	OutputFile              string         `json:"output_file"`
	RetryAttempts           int            `json:"retry_attempts"`
	RetryDelay              time.Duration  `json:"retry_delay"`
	EnableMetrics           bool           `json:"enable_metrics"`
	EnableLogging           bool           `json:"enable_logging"`
	LogLevel                string         `json:"log_level"`
	DomainRateLimit         map[string]int `json:"domain_rate_limit"`
	CircuitBreakerThreshold int            `json:"circuit_breaker_threshold"`
	CircuitBreakerTimeout   time.Duration  `json:"circuit_breaker_timeout"`
	UseHeadless             bool           `json:"use_headless"`
	MaxPages                int            `json:"max_pages"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		MaxConcurrent:           3,
		RequestTimeout:          10 * time.Second,
		TotalTimeout:            30 * time.Second,
		UserAgent:               "Go-Scraper/2.0",
		OutputFile:              "scraping_results.json",
		RetryAttempts:           3,
		RetryDelay:              1 * time.Second,
		EnableMetrics:           true,
		EnableLogging:           true,
		LogLevel:                "info",
		DomainRateLimit:         make(map[string]int),
		CircuitBreakerThreshold: 3,
		CircuitBreakerTimeout:   30 * time.Second,
		UseHeadless:             false,
		MaxPages:                10,
	}
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	config := DefaultConfig()

	// Load from environment variables
	if val := os.Getenv("SCRAPER_MAX_CONCURRENT"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			config.MaxConcurrent = parsed
		}
	}

	if val := os.Getenv("SCRAPER_REQUEST_TIMEOUT"); val != "" {
		if parsed, err := time.ParseDuration(val); err == nil {
			config.RequestTimeout = parsed
		}
	}

	if val := os.Getenv("SCRAPER_TOTAL_TIMEOUT"); val != "" {
		if parsed, err := time.ParseDuration(val); err == nil {
			config.TotalTimeout = parsed
		}
	}

	if val := os.Getenv("SCRAPER_USER_AGENT"); val != "" {
		config.UserAgent = val
	}

	if val := os.Getenv("SCRAPER_OUTPUT_FILE"); val != "" {
		config.OutputFile = val
	}

	if val := os.Getenv("SCRAPER_RETRY_ATTEMPTS"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			config.RetryAttempts = parsed
		}
	}

	if val := os.Getenv("SCRAPER_RETRY_DELAY"); val != "" {
		if parsed, err := time.ParseDuration(val); err == nil {
			config.RetryDelay = parsed
		}
	}

	if val := os.Getenv("SCRAPER_ENABLE_METRICS"); val != "" {
		config.EnableMetrics = val == "true"
	}

	if val := os.Getenv("SCRAPER_ENABLE_LOGGING"); val != "" {
		config.EnableLogging = val == "true"
	}

	if val := os.Getenv("SCRAPER_LOG_LEVEL"); val != "" {
		config.LogLevel = val
	}

	if val := os.Getenv("SCRAPER_CIRCUIT_BREAKER_THRESHOLD"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			config.CircuitBreakerThreshold = parsed
		}
	}

	if val := os.Getenv("SCRAPER_CIRCUIT_BREAKER_TIMEOUT"); val != "" {
		if parsed, err := time.ParseDuration(val); err == nil {
			config.CircuitBreakerTimeout = parsed
		}
	}

	if val := os.Getenv("SCRAPER_USE_HEADLESS"); val != "" {
		config.UseHeadless = val == "true"
	}

	if val := os.Getenv("SCRAPER_MAX_PAGES"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			config.MaxPages = parsed
		}
	}

	return config
}

// Validate ensures configuration is valid
func (c *Config) Validate() error {
	if c.MaxConcurrent <= 0 {
		return fmt.Errorf("max_concurrent must be positive, got %d", c.MaxConcurrent)
	}

	if c.RequestTimeout <= 0 {
		return fmt.Errorf("request_timeout must be positive, got %v", c.RequestTimeout)
	}

	if c.TotalTimeout <= 0 {
		return fmt.Errorf("total_timeout must be positive, got %v", c.TotalTimeout)
	}

	if c.RetryAttempts < 0 {
		return fmt.Errorf("retry_attempts cannot be negative, got %d", c.RetryAttempts)
	}

	if c.RetryDelay < 0 {
		return fmt.Errorf("retry_delay cannot be negative, got %v", c.RetryDelay)
	}

	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[c.LogLevel] {
		return fmt.Errorf("invalid log_level: %s, must be one of: debug, info, warn, error", c.LogLevel)
	}

	return nil
}

// String returns a string representation of the configuration
func (c *Config) String() string {
	return fmt.Sprintf(
		"Config{MaxConcurrent: %d, RequestTimeout: %v, TotalTimeout: %v, UserAgent: %s, OutputFile: %s, RetryAttempts: %d, RetryDelay: %v, EnableMetrics: %t, EnableLogging: %t, LogLevel: %s, CircuitBreakerThreshold: %d, CircuitBreakerTimeout: %v, UseHeadless: %t, MaxPages: %d}",
		c.MaxConcurrent, c.RequestTimeout, c.TotalTimeout, c.UserAgent, c.OutputFile, c.RetryAttempts, c.RetryDelay, c.EnableMetrics, c.EnableLogging, c.LogLevel, c.CircuitBreakerThreshold, c.CircuitBreakerTimeout, c.UseHeadless, c.MaxPages,
	)
}
