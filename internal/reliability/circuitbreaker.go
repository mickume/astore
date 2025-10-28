package reliability

import (
	"context"
	"fmt"
	"sync"
	"time"

	"zotregistry.io/zot/pkg/log"
)

// CircuitState represents the state of a circuit breaker
type CircuitState string

const (
	StateClosed    CircuitState = "closed"     // Normal operation
	StateOpen      CircuitState = "open"       // Failing, reject requests
	StateHalfOpen  CircuitState = "half_open"  // Testing recovery
)

// CircuitBreakerConfig configures circuit breaker behavior
type CircuitBreakerConfig struct {
	MaxFailures     int           // Number of failures before opening
	Timeout         time.Duration // Time to wait before attempting recovery
	HalfOpenSuccess int           // Successful calls needed in half-open to close
	HalfOpenMax     int           // Max calls allowed in half-open state
}

// DefaultCircuitBreakerConfig returns sensible defaults
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		MaxFailures:     5,
		Timeout:         30 * time.Second,
		HalfOpenSuccess: 2,
		HalfOpenMax:     5,
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	config          *CircuitBreakerConfig
	state           CircuitState
	failures        int
	successes       int
	halfOpenAttempts int
	lastFailTime    time.Time
	mu              sync.RWMutex
	logger          log.Logger
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config *CircuitBreakerConfig, logger log.Logger) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}

	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
		logger: logger,
	}
}

// Execute runs the function through the circuit breaker
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	// Check if we can execute
	if !cb.canExecute() {
		return fmt.Errorf("circuit breaker is open")
	}

	// Execute the function
	err := fn(ctx)

	// Record result
	cb.recordResult(err)

	return err
}

// canExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) canExecute() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		// Always allow in closed state
		return true

	case StateOpen:
		// Check if timeout has passed
		if time.Since(cb.lastFailTime) > cb.config.Timeout {
			// Transition to half-open
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = StateHalfOpen
			cb.halfOpenAttempts = 0
			cb.successes = 0
			cb.logger.Info().Msg("circuit breaker transitioning to half-open")
			cb.mu.Unlock()
			cb.mu.RLock()
			return true
		}
		return false

	case StateHalfOpen:
		// Allow limited requests in half-open
		return cb.halfOpenAttempts < cb.config.HalfOpenMax

	default:
		return false
	}
}

// recordResult records the result of an execution
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.recordFailure()
	} else {
		cb.recordSuccess()
	}
}

// recordFailure records a failed execution
func (cb *CircuitBreaker) recordFailure() {
	cb.failures++
	cb.lastFailTime = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.config.MaxFailures {
			cb.state = StateOpen
			cb.logger.Warn().
				Int("failures", cb.failures).
				Msg("circuit breaker opened")
		}

	case StateHalfOpen:
		// Any failure in half-open reopens the circuit
		cb.state = StateOpen
		cb.halfOpenAttempts = 0
		cb.successes = 0
		cb.logger.Warn().Msg("circuit breaker reopened after half-open failure")
	}
}

// recordSuccess records a successful execution
func (cb *CircuitBreaker) recordSuccess() {
	switch cb.state {
	case StateClosed:
		// Reset failure count on success
		cb.failures = 0

	case StateHalfOpen:
		cb.successes++
		cb.halfOpenAttempts++

		if cb.successes >= cb.config.HalfOpenSuccess {
			// Enough successes, close the circuit
			cb.state = StateClosed
			cb.failures = 0
			cb.successes = 0
			cb.halfOpenAttempts = 0
			cb.logger.Info().Msg("circuit breaker closed after successful recovery")
		}
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetMetrics returns current circuit breaker metrics
func (cb *CircuitBreaker) GetMetrics() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":            cb.state,
		"failures":         cb.failures,
		"successes":        cb.successes,
		"halfOpenAttempts": cb.halfOpenAttempts,
		"lastFailTime":     cb.lastFailTime,
	}
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
	cb.halfOpenAttempts = 0
	cb.logger.Info().Msg("circuit breaker manually reset")
}

// CircuitBreakerManager manages multiple circuit breakers by name
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	config   *CircuitBreakerConfig
	logger   log.Logger
	mu       sync.RWMutex
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager(config *CircuitBreakerConfig, logger log.Logger) *CircuitBreakerManager {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}

	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
		config:   config,
		logger:   logger,
	}
}

// GetBreaker gets or creates a circuit breaker for the given name
func (m *CircuitBreakerManager) GetBreaker(name string) *CircuitBreaker {
	m.mu.RLock()
	breaker, exists := m.breakers[name]
	m.mu.RUnlock()

	if exists {
		return breaker
	}

	// Create new breaker
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if breaker, exists := m.breakers[name]; exists {
		return breaker
	}

	breaker = NewCircuitBreaker(m.config, m.logger)
	m.breakers[name] = breaker

	return breaker
}

// Execute executes a function through a named circuit breaker
func (m *CircuitBreakerManager) Execute(ctx context.Context, name string, fn func(ctx context.Context) error) error {
	breaker := m.GetBreaker(name)
	return breaker.Execute(ctx, fn)
}

// GetAllMetrics returns metrics for all circuit breakers
func (m *CircuitBreakerManager) GetAllMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := make(map[string]interface{})
	for name, breaker := range m.breakers {
		metrics[name] = breaker.GetMetrics()
	}

	return metrics
}
