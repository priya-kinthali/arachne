package main

import (
	"sync"
	"time"
)

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

// String returns the string representation of the circuit breaker state
func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreaker implements the circuit breaker pattern for fault tolerance
type CircuitBreaker struct {
	mu sync.RWMutex

	// Configuration
	failureThreshold int           // Number of failures before opening
	resetTimeout     time.Duration // Time to wait before attempting to close
	halfOpenLimit    int           // Number of requests to allow in half-open state

	// State
	state CircuitBreakerState

	// Counters
	failureCount    int
	successCount    int
	lastFailureTime time.Time

	// Statistics
	totalRequests   int64
	totalFailures   int64
	totalSuccesses  int64
	lastStateChange time.Time
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(failureThreshold int, resetTimeout time.Duration) *CircuitBreaker {
	return NewCircuitBreakerWithConfig(failureThreshold, resetTimeout, 1)
}

// NewCircuitBreakerWithConfig creates a new circuit breaker with custom configuration
func NewCircuitBreakerWithConfig(failureThreshold int, resetTimeout time.Duration, halfOpenLimit int) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
		halfOpenLimit:    halfOpenLimit,
		state:            StateClosed,
		lastStateChange:  time.Now(),
	}
}

// Execute runs a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if !cb.canExecute() {
		return &CircuitBreakerError{
			State: cb.getState(),
			Msg:   "circuit breaker is open",
		}
	}

	cb.recordRequest()
	err := fn()

	if err != nil {
		cb.recordFailure()
	} else {
		cb.recordSuccess()
	}

	return err
}

// canExecute determines if the circuit breaker allows execution
func (cb *CircuitBreaker) canExecute() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if we should transition from open to half-open
	if cb.state == StateOpen && time.Since(cb.lastFailureTime) >= cb.resetTimeout {
		cb.transitionToHalfOpen()
	}

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		return false
	case StateHalfOpen:
		return cb.successCount < cb.halfOpenLimit
	default:
		return false
	}
}

// recordRequest records a request attempt
func (cb *CircuitBreaker) recordRequest() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.totalRequests++
}

// recordFailure records a failure and potentially opens the circuit
func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.totalFailures++
	cb.failureCount++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failureCount >= cb.failureThreshold {
			cb.transitionToOpen()
		}
	case StateHalfOpen:
		cb.transitionToOpen()
	}
}

// recordSuccess records a success and potentially closes the circuit
func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.totalSuccesses++
	cb.successCount++

	if cb.state == StateHalfOpen && cb.successCount >= cb.halfOpenLimit {
		cb.transitionToClosed()
	}
}

// transitionToOpen transitions the circuit breaker to open state
func (cb *CircuitBreaker) transitionToOpen() {
	if cb.state != StateOpen {
		cb.state = StateOpen
		cb.lastStateChange = time.Now()
		cb.failureCount = 0
		cb.successCount = 0
	}
}

// transitionToHalfOpen transitions the circuit breaker to half-open state
func (cb *CircuitBreaker) transitionToHalfOpen() {
	if cb.state != StateHalfOpen {
		cb.state = StateHalfOpen
		cb.lastStateChange = time.Now()
		cb.failureCount = 0
		cb.successCount = 0
	}
}

// transitionToClosed transitions the circuit breaker to closed state
func (cb *CircuitBreaker) transitionToClosed() {
	if cb.state != StateClosed {
		cb.state = StateClosed
		cb.lastStateChange = time.Now()
		cb.failureCount = 0
		cb.successCount = 0
	}
}

// getState returns the current state of the circuit breaker
func (cb *CircuitBreaker) getState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats returns statistics about the circuit breaker
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":             cb.state.String(),
		"total_requests":    cb.totalRequests,
		"total_failures":    cb.totalFailures,
		"total_successes":   cb.totalSuccesses,
		"failure_count":     cb.failureCount,
		"success_count":     cb.successCount,
		"last_failure":      cb.lastFailureTime,
		"last_state_change": cb.lastStateChange,
		"failure_rate":      cb.getFailureRate(),
	}
}

// getFailureRate calculates the failure rate as a percentage
func (cb *CircuitBreaker) getFailureRate() float64 {
	if cb.totalRequests == 0 {
		return 0.0
	}
	return float64(cb.totalFailures) / float64(cb.totalRequests) * 100.0
}

// CircuitBreakerError represents a circuit breaker error
type CircuitBreakerError struct {
	State CircuitBreakerState
	Msg   string
}

// Error implements the error interface
func (e *CircuitBreakerError) Error() string {
	return e.Msg
}

// IsCircuitBreakerError checks if an error is a circuit breaker error
func IsCircuitBreakerError(err error) bool {
	_, ok := err.(*CircuitBreakerError)
	return ok
}
