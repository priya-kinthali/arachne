package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics tracks various statistics during scraping
type Metrics struct {
	mu sync.RWMutex

	// Atomic counters for thread-safe access
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	RetryAttempts      int64
	TotalBytes         int64

	// Timing information
	StartTime     time.Time
	EndTime       time.Time
	TotalDuration time.Duration

	// Per-domain statistics
	DomainStats map[string]*DomainMetrics

	// Response time statistics
	MinResponseTime time.Duration
	MaxResponseTime time.Duration
	AvgResponseTime time.Duration
	ResponseTimes   []time.Duration

	// Status code distribution
	StatusCodeCounts map[int]int64
}

// DomainMetrics tracks statistics for a specific domain
type DomainMetrics struct {
	Requests        int64
	Successes       int64
	Failures        int64
	TotalBytes      int64
	AvgResponseTime time.Duration
	ResponseTimes   []time.Duration
}

// NewMetrics creates a new metrics tracker
func NewMetrics() *Metrics {
	return &Metrics{
		StartTime:        time.Now(),
		DomainStats:      make(map[string]*DomainMetrics),
		StatusCodeCounts: make(map[int]int64),
		ResponseTimes:    make([]time.Duration, 0),
	}
}

// RecordRequest records a new request
func (m *Metrics) RecordRequest() {
	atomic.AddInt64(&m.TotalRequests, 1)
}

// RecordSuccess records a successful request
func (m *Metrics) RecordSuccess(domain string, statusCode int, bytes int64, responseTime time.Duration) {
	atomic.AddInt64(&m.SuccessfulRequests, 1)
	atomic.AddInt64(&m.TotalBytes, bytes)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Update status code counts
	m.StatusCodeCounts[statusCode]++

	// Update response time statistics
	m.ResponseTimes = append(m.ResponseTimes, responseTime)
	if m.MinResponseTime == 0 || responseTime < m.MinResponseTime {
		m.MinResponseTime = responseTime
	}
	if responseTime > m.MaxResponseTime {
		m.MaxResponseTime = responseTime
	}

	// Update domain statistics
	if m.DomainStats[domain] == nil {
		m.DomainStats[domain] = &DomainMetrics{
			ResponseTimes: make([]time.Duration, 0),
		}
	}
	dm := m.DomainStats[domain]
	dm.Requests++
	dm.Successes++
	dm.TotalBytes += bytes
	dm.ResponseTimes = append(dm.ResponseTimes, responseTime)
}

// RecordFailure records a failed request
func (m *Metrics) RecordFailure(domain string, statusCode int) {
	atomic.AddInt64(&m.FailedRequests, 1)

	m.mu.Lock()
	defer m.mu.Unlock()

	if statusCode > 0 {
		m.StatusCodeCounts[statusCode]++
	}

	if m.DomainStats[domain] == nil {
		m.DomainStats[domain] = &DomainMetrics{
			ResponseTimes: make([]time.Duration, 0),
		}
	}
	m.DomainStats[domain].Requests++
	m.DomainStats[domain].Failures++
}

// RecordRetry records a retry attempt
func (m *Metrics) RecordRetry() {
	atomic.AddInt64(&m.RetryAttempts, 1)
}

// Finish marks the end of scraping and calculates final statistics
func (m *Metrics) Finish() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.EndTime = time.Now()
	m.TotalDuration = m.EndTime.Sub(m.StartTime)

	// Calculate average response time
	if len(m.ResponseTimes) > 0 {
		total := time.Duration(0)
		for _, rt := range m.ResponseTimes {
			total += rt
		}
		m.AvgResponseTime = total / time.Duration(len(m.ResponseTimes))
	}

	// Calculate domain averages
	for _, dm := range m.DomainStats {
		if len(dm.ResponseTimes) > 0 {
			total := time.Duration(0)
			for _, rt := range dm.ResponseTimes {
				total += rt
			}
			dm.AvgResponseTime = total / time.Duration(len(dm.ResponseTimes))
		}
	}
}

// GetSuccessRate returns the success rate as a percentage
func (m *Metrics) GetSuccessRate() float64 {
	total := atomic.LoadInt64(&m.TotalRequests)
	if total == 0 {
		return 0.0
	}
	successful := atomic.LoadInt64(&m.SuccessfulRequests)
	return float64(successful) / float64(total) * 100.0
}

// GetRequestsPerSecond returns the requests per second rate
func (m *Metrics) GetRequestsPerSecond() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.TotalDuration == 0 {
		return 0.0
	}
	return float64(m.TotalRequests) / m.TotalDuration.Seconds()
}

// PrintSummary prints a formatted summary of metrics
func (m *Metrics) PrintSummary() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	fmt.Printf("\n📊 Scraping Metrics Summary\n")
	fmt.Printf("========================\n")
	fmt.Printf("⏱️  Total Duration: %v\n", m.TotalDuration)
	fmt.Printf("📈 Total Requests: %d\n", m.TotalRequests)
	fmt.Printf("✅ Successful: %d (%.1f%%)\n", m.SuccessfulRequests, m.GetSuccessRate())
	fmt.Printf("❌ Failed: %d\n", m.FailedRequests)
	fmt.Printf("🔄 Retry Attempts: %d\n", m.RetryAttempts)
	fmt.Printf("📦 Total Bytes: %d (%.2f MB)\n", m.TotalBytes, float64(m.TotalBytes)/1024/1024)
	fmt.Printf("⚡ Requests/Second: %.2f\n", m.GetRequestsPerSecond())

	if len(m.ResponseTimes) > 0 {
		fmt.Printf("\n⏱️  Response Time Statistics:\n")
		fmt.Printf("   Min: %v\n", m.MinResponseTime)
		fmt.Printf("   Max: %v\n", m.MaxResponseTime)
		fmt.Printf("   Avg: %v\n", m.AvgResponseTime)
	}

	if len(m.StatusCodeCounts) > 0 {
		fmt.Printf("\n📋 Status Code Distribution:\n")
		for code, count := range m.StatusCodeCounts {
			fmt.Printf("   %d: %d\n", code, count)
		}
	}

	if len(m.DomainStats) > 0 {
		fmt.Printf("\n🌐 Per-Domain Statistics:\n")
		for domain, stats := range m.DomainStats {
			successRate := float64(0)
			if stats.Requests > 0 {
				successRate = float64(stats.Successes) / float64(stats.Requests) * 100
			}
			fmt.Printf("   %s: %d/%d (%.1f%%) - %v avg\n",
				domain, stats.Successes, stats.Requests, successRate, stats.AvgResponseTime)
		}
	}
}

// GetMetrics returns a copy of current metrics for JSON serialization
func (m *Metrics) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"total_requests":      m.TotalRequests,
		"successful_requests": m.SuccessfulRequests,
		"failed_requests":     m.FailedRequests,
		"retry_attempts":      m.RetryAttempts,
		"total_bytes":         m.TotalBytes,
		"success_rate":        m.GetSuccessRate(),
		"requests_per_second": m.GetRequestsPerSecond(),
		"total_duration":      m.TotalDuration.String(),
		"start_time":          m.StartTime,
		"end_time":            m.EndTime,
		"response_times": map[string]interface{}{
			"min": m.MinResponseTime.String(),
			"max": m.MaxResponseTime.String(),
			"avg": m.AvgResponseTime.String(),
		},
		"status_codes": m.StatusCodeCounts,
		"domains":      m.DomainStats,
	}
}
