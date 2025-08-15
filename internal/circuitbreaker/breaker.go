package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// State represents the current state of the circuit breaker
type State int

const (
	// StateClosed means the circuit breaker is closed and operations are allowed
	StateClosed State = iota
	// StateOpen means the circuit breaker is open and operations are rejected
	StateOpen
	// StateHalfOpen means the circuit breaker is testing if the service has recovered
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// Config defines configuration for the circuit breaker
type Config struct {
	// MaxFailures is the maximum number of failures before opening the circuit
	MaxFailures int64
	// Timeout is how long to keep the circuit open before attempting recovery
	Timeout time.Duration
	// MaxRequests is the maximum number of requests allowed in half-open state
	MaxRequests int64
	// Interval is the time window for failure counting
	Interval time.Duration
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		MaxFailures: 5,
		Timeout:     30 * time.Second,
		MaxRequests: 3,
		Interval:    60 * time.Second,
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	config    Config
	mutex     sync.RWMutex
	state     State
	failures  int64
	requests  int64
	lastFailTime time.Time
	stateChanged time.Time
	
	// Callback functions
	onStateChange func(from, to State)
}

// New creates a new circuit breaker
func New(config Config) *CircuitBreaker {
	return &CircuitBreaker{
		config:       config,
		state:        StateClosed,
		stateChanged: time.Now(),
	}
}

// Execute executes the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	allowed, err := cb.beforeExecution()
	if !allowed {
		return err
	}
	
	err = fn()
	cb.afterExecution(err == nil)
	
	return err
}

// beforeExecution checks if execution is allowed and updates state
func (cb *CircuitBreaker) beforeExecution() (bool, error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	switch cb.state {
	case StateClosed:
		return true, nil
		
	case StateOpen:
		// Check if timeout has passed
		if time.Since(cb.stateChanged) >= cb.config.Timeout {
			cb.changeState(StateHalfOpen)
			return true, nil
		}
		return false, errors.New("circuit breaker is open")
		
	case StateHalfOpen:
		// Allow limited number of requests
		if cb.requests >= cb.config.MaxRequests {
			return false, errors.New("circuit breaker is half-open and at request limit")
		}
		cb.requests++
		return true, nil
		
	default:
		return false, errors.New("unknown circuit breaker state")
	}
}

// afterExecution updates the circuit breaker state after execution
func (cb *CircuitBreaker) afterExecution(success bool) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	if success {
		// Reset failures on success
		switch cb.state {
		case StateHalfOpen:
			// If we've had enough successful requests, close the circuit
			if cb.requests >= cb.config.MaxRequests {
				cb.changeState(StateClosed)
			}
		case StateClosed:
			// Reset failure count on success
			cb.failures = 0
		}
	} else {
		// Handle failure
		cb.failures++
		cb.lastFailTime = time.Now()
		
		switch cb.state {
		case StateClosed:
			// Open circuit if we've exceeded failure threshold
			if cb.failures >= cb.config.MaxFailures {
				cb.changeState(StateOpen)
			}
		case StateHalfOpen:
			// Go back to open state on any failure
			cb.changeState(StateOpen)
		}
	}
}

// changeState changes the circuit breaker state
func (cb *CircuitBreaker) changeState(newState State) {
	oldState := cb.state
	cb.state = newState
	cb.stateChanged = time.Now()
	
	// Reset counters based on new state
	switch newState {
	case StateClosed:
		cb.failures = 0
		cb.requests = 0
	case StateOpen:
		cb.requests = 0
	case StateHalfOpen:
		cb.requests = 0
	}
	
	// Call callback if set
	if cb.onStateChange != nil {
		cb.onStateChange(oldState, newState)
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() State {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// GetStats returns statistics about the circuit breaker
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	
	return map[string]interface{}{
		"state":          cb.state.String(),
		"failures":       cb.failures,
		"requests":       cb.requests,
		"last_fail_time": cb.lastFailTime,
		"state_changed":  cb.stateChanged,
		"config": map[string]interface{}{
			"max_failures": cb.config.MaxFailures,
			"timeout":      cb.config.Timeout,
			"max_requests": cb.config.MaxRequests,
			"interval":     cb.config.Interval,
		},
	}
}

// OnStateChange sets a callback function to be called when state changes
func (cb *CircuitBreaker) OnStateChange(fn func(from, to State)) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.onStateChange = fn
}

// BrowserCircuitBreaker wraps browser operations with circuit breaker protection
type BrowserCircuitBreaker struct {
	CircuitBreaker *CircuitBreaker
}

// NewBrowserCircuitBreaker creates a new circuit breaker for browser operations
func NewBrowserCircuitBreaker() *BrowserCircuitBreaker {
	config := Config{
		MaxFailures: 3,           // Open after 3 browser failures
		Timeout:     60 * time.Second, // Wait 1 minute before retry
		MaxRequests: 2,           // Allow 2 test requests when half-open
		Interval:    30 * time.Second, // 30-second failure window
	}
	
	return &BrowserCircuitBreaker{
		CircuitBreaker: New(config),
	}
}

// ExecuteBrowserOperation executes a browser operation with circuit breaker protection
func (bcb *BrowserCircuitBreaker) ExecuteBrowserOperation(operation func() error) error {
	return bcb.CircuitBreaker.Execute(operation)
}

// GetState returns the current circuit breaker state
func (bcb *BrowserCircuitBreaker) GetState() State {
	return bcb.CircuitBreaker.GetState()
}

// GetStats returns circuit breaker statistics
func (bcb *BrowserCircuitBreaker) GetStats() map[string]interface{} {
	return bcb.CircuitBreaker.GetStats()
}

// IsOperationAllowed checks if an operation would be allowed without executing it
func (bcb *BrowserCircuitBreaker) IsOperationAllowed() bool {
	allowed, _ := bcb.CircuitBreaker.beforeExecution()
	return allowed
}

// NetworkCircuitBreaker wraps network operations with circuit breaker protection
type NetworkCircuitBreaker struct {
	CircuitBreaker *CircuitBreaker
}

// NewNetworkCircuitBreaker creates a new circuit breaker for network operations
func NewNetworkCircuitBreaker() *NetworkCircuitBreaker {
	config := Config{
		MaxFailures: 5,           // Open after 5 network failures
		Timeout:     30 * time.Second, // Wait 30 seconds before retry
		MaxRequests: 3,           // Allow 3 test requests when half-open
		Interval:    60 * time.Second, // 1-minute failure window
	}
	
	return &NetworkCircuitBreaker{
		CircuitBreaker: New(config),
	}
}

// ExecuteNetworkOperation executes a network operation with circuit breaker protection
func (ncb *NetworkCircuitBreaker) ExecuteNetworkOperation(operation func() error) error {
	return ncb.CircuitBreaker.Execute(operation)
}

// GetState returns the current circuit breaker state
func (ncb *NetworkCircuitBreaker) GetState() State {
	return ncb.CircuitBreaker.GetState()
}

// GetStats returns circuit breaker statistics
func (ncb *NetworkCircuitBreaker) GetStats() map[string]interface{} {
	return ncb.CircuitBreaker.GetStats()
}

// Multi-level circuit breaker for different operation types
type MultiLevelCircuitBreaker struct {
	BrowserCircuitBreaker *BrowserCircuitBreaker
	NetworkCircuitBreaker *NetworkCircuitBreaker
	mutex                 sync.RWMutex
}

// NewMultiLevelCircuitBreaker creates a multi-level circuit breaker
func NewMultiLevelCircuitBreaker() *MultiLevelCircuitBreaker {
	return &MultiLevelCircuitBreaker{
		BrowserCircuitBreaker: NewBrowserCircuitBreaker(),
		NetworkCircuitBreaker: NewNetworkCircuitBreaker(),
	}
}

// ExecuteBrowserOperation executes a browser operation with protection
func (mlcb *MultiLevelCircuitBreaker) ExecuteBrowserOperation(operation func() error) error {
	return mlcb.BrowserCircuitBreaker.ExecuteBrowserOperation(operation)
}

// ExecuteNetworkOperation executes a network operation with protection
func (mlcb *MultiLevelCircuitBreaker) ExecuteNetworkOperation(operation func() error) error {
	return mlcb.NetworkCircuitBreaker.ExecuteNetworkOperation(operation)
}

// GetOverallStats returns statistics for all circuit breakers
func (mlcb *MultiLevelCircuitBreaker) GetOverallStats() map[string]interface{} {
	mlcb.mutex.RLock()
	defer mlcb.mutex.RUnlock()
	
	return map[string]interface{}{
		"browser": mlcb.BrowserCircuitBreaker.GetStats(),
		"network": mlcb.NetworkCircuitBreaker.GetStats(),
		"overall_healthy": mlcb.BrowserCircuitBreaker.GetState() != StateOpen &&
			mlcb.NetworkCircuitBreaker.GetState() != StateOpen,
	}
}