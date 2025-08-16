package retry

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

// Config defines retry configuration
type Config struct {
	MaxAttempts     int           // Maximum number of attempts
	InitialDelay    time.Duration // Initial delay between attempts
	MaxDelay        time.Duration // Maximum delay between attempts
	Multiplier      float64       // Multiplier for exponential backoff
	Jitter          bool          // Add jitter to delays
	RetryableErrors []string      // List of error strings that are retryable
}

// DefaultConfig returns a default retry configuration
func DefaultConfig() Config {
	return Config{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
		RetryableErrors: []string{
			"context canceled",
			"context deadline exceeded",
			"timeout",
			"connection reset",
			"broken pipe",
			"target closed",
			"browser not started",
		},
	}
}

// RetryableFunc is a function that can be retried
type RetryableFunc func() error

// RetryableWithResultFunc is a function that returns a result and can be retried
type RetryableWithResultFunc func() (interface{}, error)

// Retrier handles retry logic with exponential backoff
type Retrier struct {
	config Config
}

// New creates a new Retrier with the given configuration
func New(config Config) *Retrier {
	return &Retrier{config: config}
}

// NewWithDefaults creates a new Retrier with default configuration
func NewWithDefaults() *Retrier {
	return New(DefaultConfig())
}

// Do executes the function with retry logic
func (r *Retrier) Do(ctx context.Context, fn RetryableFunc) error {
	var lastErr error
	
	for attempt := 0; attempt < r.config.MaxAttempts; attempt++ {
		// Check context before attempting
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		// Execute the function
		err := fn()
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		// Check if error is retryable
		if !r.isRetryable(err) {
			return err
		}
		
		// Don't delay after the last attempt
		if attempt < r.config.MaxAttempts-1 {
			delay := r.calculateDelay(attempt)
			
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// Continue to next attempt
			}
		}
	}
	
	return fmt.Errorf("failed after %d attempts: %w", r.config.MaxAttempts, lastErr)
}

// DoWithResult executes the function with retry logic and returns a result
func (r *Retrier) DoWithResult(ctx context.Context, fn RetryableWithResultFunc) (interface{}, error) {
	var lastErr error
	
	for attempt := 0; attempt < r.config.MaxAttempts; attempt++ {
		// Check context before attempting
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		
		// Execute the function
		result, err := fn()
		if err == nil {
			return result, nil
		}
		
		lastErr = err
		
		// Check if error is retryable
		if !r.isRetryable(err) {
			return nil, err
		}
		
		// Don't delay after the last attempt
		if attempt < r.config.MaxAttempts-1 {
			delay := r.calculateDelay(attempt)
			
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				// Continue to next attempt
			}
		}
	}
	
	return nil, fmt.Errorf("failed after %d attempts: %w", r.config.MaxAttempts, lastErr)
}

// calculateDelay calculates the delay for the given attempt number
func (r *Retrier) calculateDelay(attempt int) time.Duration {
	// Calculate exponential backoff
	delay := float64(r.config.InitialDelay) * math.Pow(r.config.Multiplier, float64(attempt))
	
	// Apply maximum delay cap
	if delay > float64(r.config.MaxDelay) {
		delay = float64(r.config.MaxDelay)
	}
	
	// Add jitter if configured
	if r.config.Jitter {
		// Add random jitter between 0% and 25% of the delay
		jitter := rand.Float64() * 0.25 * delay
		delay += jitter
	}
	
	return time.Duration(delay)
}

// isRetryable checks if an error is retryable
func (r *Retrier) isRetryable(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := strings.ToLower(err.Error())
	
	// Check against configured retryable errors
	for _, retryableErr := range r.config.RetryableErrors {
		if strings.Contains(errStr, strings.ToLower(retryableErr)) {
			return true
		}
	}
	
	return false
}

// WithMaxAttempts sets the maximum number of attempts
func (r *Retrier) WithMaxAttempts(attempts int) *Retrier {
	r.config.MaxAttempts = attempts
	return r
}

// WithDelay sets the initial delay
func (r *Retrier) WithDelay(delay time.Duration) *Retrier {
	r.config.InitialDelay = delay
	return r
}

// WithMaxDelay sets the maximum delay
func (r *Retrier) WithMaxDelay(maxDelay time.Duration) *Retrier {
	r.config.MaxDelay = maxDelay
	return r
}

// WithMultiplier sets the backoff multiplier
func (r *Retrier) WithMultiplier(multiplier float64) *Retrier {
	r.config.Multiplier = multiplier
	return r
}

// WithJitter enables or disables jitter
func (r *Retrier) WithJitter(jitter bool) *Retrier {
	r.config.Jitter = jitter
	return r
}

// AddRetryableError adds an error string to the list of retryable errors
func (r *Retrier) AddRetryableError(errStr string) *Retrier {
	r.config.RetryableErrors = append(r.config.RetryableErrors, errStr)
	return r
}