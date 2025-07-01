package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// ScrapedData represents the data we extract from websites
type ScrapedData struct {
	URL     string    `json:"url"`
	Title   string    `json:"title"`
	Status  int       `json:"status"`
	Size    int       `json:"size"`
	Error   string    `json:"error,omitempty"`
	Scraped time.Time `json:"scraped"`
}

// Scraper handles concurrent web scraping with rate limiting
type Scraper struct {
	config          *Config
	logger          *Logger
	metrics         *Metrics
	client          *http.Client
	rateLimiter     chan struct{}
	domainLimiters  map[string]chan struct{}
	circuitBreakers map[string]*CircuitBreaker
	results         chan ScrapedData
	wg              sync.WaitGroup
	mu              sync.RWMutex
}

// NewScraper creates a new scraper with configurable concurrency
func NewScraper(config *Config) *Scraper {
	// Create transport with connection pooling and HTTP/2 support
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableCompression:  false, // Enable compression
		ForceAttemptHTTP2:   true,  // Force HTTP/2 when possible
	}

	scraper := &Scraper{
		config:  config,
		logger:  NewLogger(config.LogLevel),
		metrics: NewMetrics(),
		client: &http.Client{
			Timeout:   config.RequestTimeout,
			Transport: transport,
		},
		rateLimiter:     make(chan struct{}, config.MaxConcurrent),
		domainLimiters:  make(map[string]chan struct{}),
		circuitBreakers: make(map[string]*CircuitBreaker),
		results:         make(chan ScrapedData, 100),
	}

	// Initialize domain-specific rate limiters
	for domain, limit := range config.DomainRateLimit {
		scraper.domainLimiters[domain] = make(chan struct{}, limit)
	}

	return scraper
}

// scrapeURL fetches a single URL and extracts basic information with retry logic
func (s *Scraper) scrapeURL(ctx context.Context, urlStr string) {
	defer s.wg.Done()

	// Validate URL
	if err := ValidateURL(urlStr); err != nil {
		s.logger.Error("Invalid URL: %s", urlStr)
		data := ScrapedData{
			URL:     urlStr,
			Error:   err.Error(),
			Scraped: time.Now(),
		}
		s.results <- data
		return
	}

	// Extract domain for rate limiting and circuit breaker
	parsedURL, _ := url.Parse(urlStr)
	domain := parsedURL.Host

	// Get or create circuit breaker for this domain
	s.mu.Lock()
	cb, exists := s.circuitBreakers[domain]
	if !exists {
		cb = NewCircuitBreaker(s.config.CircuitBreakerThreshold, s.config.CircuitBreakerTimeout)
		s.circuitBreakers[domain] = cb
	}
	s.mu.Unlock()

	// Acquire global rate limiter slot
	s.rateLimiter <- struct{}{}
	defer func() { <-s.rateLimiter }()

	// Acquire domain-specific rate limiter if configured
	s.mu.RLock()
	domainLimiter, hasDomainLimit := s.domainLimiters[domain]
	s.mu.RUnlock()

	if hasDomainLimit {
		domainLimiter <- struct{}{}
		defer func() { <-domainLimiter }()
	}

	data := ScrapedData{
		URL:     urlStr,
		Scraped: time.Now(),
	}

	// Record request in metrics
	s.metrics.RecordRequest()

	// Attempt scraping with retry logic and circuit breaker
	var lastErr error
	for attempt := 1; attempt <= s.config.RetryAttempts; attempt++ {
		start := time.Now()

		// Execute request with circuit breaker protection
		err := cb.Execute(func() error {
			// Create request with context for cancellation
			req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
			if err != nil {
				return NewScraperError(urlStr, "Failed to create request", err)
			}

			// Set user agent to be respectful
			req.Header.Set("User-Agent", s.config.UserAgent)

			// Make the request
			resp, err := s.client.Do(req)
			if err != nil {
				return NewScraperError(urlStr, "Request failed", err)
			}
			defer resp.Body.Close()

			// Check for HTTP errors
			if resp.StatusCode >= 400 {
				return NewHTTPError(urlStr, resp.StatusCode, fmt.Sprintf("HTTP %d", resp.StatusCode))
			}

			// Read response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return NewScraperError(urlStr, "Failed to read body", err)
			}

			// Extract title from response
			title := extractTitle(string(body), resp.Header.Get("Content-Type"))

			// Record success in metrics
			duration := time.Since(start)
			s.metrics.RecordSuccess(domain, resp.StatusCode, int64(len(body)), duration)

			// Log success
			s.logger.LogSuccess(urlStr, resp.StatusCode, len(body), duration)

			// Set data
			data.Status = resp.StatusCode
			data.Size = len(body)
			data.Title = title

			return nil
		})

		if err != nil {
			lastErr = err

			// Check if it's a circuit breaker error
			if IsCircuitBreakerError(err) {
				s.logger.Warn("Circuit breaker open for %s: %v", domain, err)
				break
			}

			// Log retry attempt if retryable
			if scraperErr, ok := err.(*ScraperError); ok && scraperErr.IsRetryable() && attempt < s.config.RetryAttempts {
				s.metrics.RecordRetry()
				s.logger.LogRetry(urlStr, attempt, err)
				time.Sleep(s.config.RetryDelay * time.Duration(attempt)) // Exponential backoff
				continue
			}
			break
		}

		// Success - break out of retry loop
		lastErr = nil
		break
	}

	// Handle final error if all retries failed
	if lastErr != nil {
		data.Error = lastErr.Error()
		s.metrics.RecordFailure(domain, 0)
		s.logger.LogFailure(urlStr, lastErr)
	}

	s.results <- data
}

// extractTitle extracts title from HTML or JSON responses
func extractTitle(content, contentType string) string {
	// Check if it's JSON based on content type or content
	if strings.Contains(contentType, "application/json") ||
		(strings.HasPrefix(content, "{") || strings.HasPrefix(content, "[")) {
		return extractJSONTitle(content)
	}

	// Otherwise treat as HTML
	return extractHTMLTitle(content)
}

// extractHTMLTitle extracts title from HTML
func extractHTMLTitle(html string) string {
	// Look for <title> tag
	titleStart := strings.Index(strings.ToLower(html), "<title>")
	if titleStart == -1 {
		return "No HTML title found"
	}

	titleStart += 7 // length of "<title>"
	titleEnd := strings.Index(html[titleStart:], "</title>")
	if titleEnd == -1 {
		return "Malformed HTML title"
	}

	title := html[titleStart : titleStart+titleEnd]
	title = strings.TrimSpace(title)

	if title == "" {
		return "Empty HTML title"
	}

	return title
}

// extractJSONTitle extracts meaningful title from JSON responses
func extractJSONTitle(jsonStr string) string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "Invalid JSON"
	}

	// Look for common title fields in JSON
	titleFields := []string{"title", "name", "login", "message", "description"}
	for _, field := range titleFields {
		if value, exists := data[field]; exists {
			if str, ok := value.(string); ok && str != "" {
				return str
			}
		}
	}

	// If no title field, return first meaningful string value in sorted order
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := data[key]
		if str, ok := value.(string); ok && len(str) < 100 && str != "" {
			return fmt.Sprintf("%s: %s", key, str)
		}
	}

	return "JSON response (no title field)"
}

// ScrapeURLs concurrently scrapes multiple URLs
func (s *Scraper) ScrapeURLs(urls []string) []ScrapedData {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.TotalTimeout)
	defer cancel()

	s.logger.Info("Starting to scrape %d URLs with %d max concurrent requests", len(urls), s.config.MaxConcurrent)

	// Start scraping goroutines
	for _, url := range urls {
		s.wg.Add(1)
		go s.scrapeURL(ctx, url)
	}

	// Close results channel when all goroutines complete
	go func() {
		s.wg.Wait()
		close(s.results)
	}()

	// Collect results
	var results []ScrapedData
	for data := range s.results {
		results = append(results, data)
	}

	// Finish metrics collection
	s.metrics.Finish()

	return results
}

// ResultProcessor processes and formats results
type ResultProcessor struct{}

// ProcessResults formats and prints results
func (rp *ResultProcessor) ProcessResults(results []ScrapedData) {
	fmt.Printf("\n=== Scraping Results (%d URLs) ===\n", len(results))

	successCount := 0
	totalSize := 0

	for _, data := range results {
		if data.Error != "" {
			fmt.Printf("❌ %s: %s\n", data.URL, data.Error)
		} else {
			fmt.Printf("✅ %s (Status: %d, Size: %d bytes)\n", data.URL, data.Status, data.Size)
			fmt.Printf("   Title: %s\n", data.Title)
			successCount++
			totalSize += data.Size
		}
		fmt.Println()
	}

	fmt.Printf("Summary: %d/%d successful, %d total bytes\n",
		successCount, len(results), totalSize)
}

// JSONExporter exports results to JSON
func (rp *ResultProcessor) ExportToJSON(results []ScrapedData, filename string) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	// Actually write to file
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	fmt.Printf("✅ JSON saved to %s\n", filename)
	return nil
}

func main() {
	// Parse command-line flags
	var (
		maxConcurrent  = flag.Int("concurrent", 3, "Maximum concurrent requests")
		requestTimeout = flag.Duration("timeout", 10*time.Second, "Request timeout")
		totalTimeout   = flag.Duration("total-timeout", 30*time.Second, "Total timeout for all requests")
		outputFile     = flag.String("output", "scraping_results.json", "Output file for results")
		retryAttempts  = flag.Int("retries", 3, "Number of retry attempts")
		retryDelay     = flag.Duration("retry-delay", 1*time.Second, "Delay between retries")
		logLevel       = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
		enableMetrics  = flag.Bool("metrics", true, "Enable metrics collection")
		enableLogging  = flag.Bool("logging", true, "Enable logging")
		userAgent      = flag.String("user-agent", "Go-Scraper/2.0", "User-Agent string")
	)
	flag.Parse()

	// Load configuration
	config := LoadConfig()

	// Override with command-line flags
	config.MaxConcurrent = *maxConcurrent
	config.RequestTimeout = *requestTimeout
	config.TotalTimeout = *totalTimeout
	config.OutputFile = *outputFile
	config.RetryAttempts = *retryAttempts
	config.RetryDelay = *retryDelay
	config.LogLevel = *logLevel
	config.EnableMetrics = *enableMetrics
	config.EnableLogging = *enableLogging
	config.UserAgent = *userAgent

	// Validate configuration
	if err := config.Validate(); err != nil {
		fmt.Printf("❌ Configuration error: %v\n", err)
		os.Exit(1)
	}

	// URLs to scrape - mix of HTML and JSON APIs
	urls := []string{
		"https://golang.org",                           // HTML with title
		"https://httpbin.org/get",                      // JSON API
		"https://jsonplaceholder.typicode.com/posts/1", // JSON API
		"https://api.github.com/users/golang",          // JSON API
		"https://httpbin.org/status/404",               // Error response
		"https://httpbin.org/delay/2",                  // Delayed response
		"https://httpbin.org/status/500",               // Server error (retryable)
		"https://httpbin.org/status/429",               // Rate limit (retryable)
	}

	fmt.Println("🚀 Starting Enhanced Concurrent Web Scraper in Go!")
	fmt.Printf("Configuration: %s\n", config.String())
	fmt.Printf("Scraping %d URLs with rate limiting...\n", len(urls))

	// Create scraper with configuration
	scraper := NewScraper(config)

	// Start timing
	start := time.Now()

	// Scrape URLs concurrently
	results := scraper.ScrapeURLs(urls)

	// Calculate duration
	duration := time.Since(start)

	// Process and display results
	processor := &ResultProcessor{}
	processor.ProcessResults(results)

	// Print metrics if enabled
	if config.EnableMetrics {
		scraper.metrics.PrintSummary()

		// Print circuit breaker statistics
		fmt.Printf("\n🔌 Circuit Breaker Statistics:\n")
		fmt.Printf("===========================\n")
		for domain, cb := range scraper.circuitBreakers {
			stats := cb.GetStats()
			fmt.Printf("🌐 %s: %s (%.1f%% failure rate)\n",
				domain,
				stats["state"],
				stats["failure_rate"])
		}
	}

	fmt.Printf("\n⏱️  Total time: %v\n", duration)

	// Export to JSON
	fmt.Println("\n📄 Exporting results to JSON...")
	if err := processor.ExportToJSON(results, config.OutputFile); err != nil {
		fmt.Printf("❌ Failed to export results: %v\n", err)
	}

	// Export metrics if enabled
	if config.EnableMetrics {
		metricsFile := "scraping_metrics.json"
		metricsData, err := json.MarshalIndent(scraper.metrics.GetMetrics(), "", "  ")
		if err == nil {
			if err := os.WriteFile(metricsFile, metricsData, 0644); err == nil {
				fmt.Printf("✅ Metrics saved to %s\n", metricsFile)
			}
		}
	}

	fmt.Println("\n✨ Enhanced scraping complete! This demonstrates:")
	fmt.Println("   • Advanced configuration management with environment variables")
	fmt.Println("   • Structured logging with multiple levels")
	fmt.Println("   • Comprehensive metrics collection and analysis")
	fmt.Println("   • Retry logic with exponential backoff")
	fmt.Println("   • Custom error types and error categorization")
	fmt.Println("   • Domain-specific rate limiting")
	fmt.Println("   • Command-line flag parsing")
	fmt.Println("   • Goroutines for concurrent execution")
	fmt.Println("   • Channels for communication and rate limiting")
	fmt.Println("   • Context for timeout and cancellation")
	fmt.Println("   • Error handling with custom error types")
	fmt.Println("   • Structs, methods, and interfaces")
	fmt.Println("   • JSON marshaling and parsing")
	fmt.Println("   • HTTP client usage with proper headers")
	fmt.Println("   • Content-type detection and parsing")
}
