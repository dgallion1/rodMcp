package retry

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestRetry_DefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	if config.MaxAttempts != 3 {
		t.Errorf("Expected max attempts 3, got %d", config.MaxAttempts)
	}
	if config.InitialDelay != 1*time.Second {
		t.Errorf("Expected initial delay 1s, got %v", config.InitialDelay)
	}
	if config.MaxDelay != 30*time.Second {
		t.Errorf("Expected max delay 30s, got %v", config.MaxDelay)
	}
	if config.Multiplier != 2.0 {
		t.Errorf("Expected multiplier 2.0, got %f", config.Multiplier)
	}
}

func TestRetry_New(t *testing.T) {
	config := DefaultConfig()
	retrier := New(config)
	
	if retrier == nil {
		t.Fatal("New returned nil retrier")
	}
	
	if retrier.config.MaxAttempts != config.MaxAttempts {
		t.Errorf("Expected max attempts %d, got %d", config.MaxAttempts, retrier.config.MaxAttempts)
	}
}

func TestRetry_NewWithDefaults(t *testing.T) {
	retrier := NewWithDefaults()
	
	if retrier == nil {
		t.Fatal("NewWithDefaults returned nil retrier")
	}
	
	if retrier.config.MaxAttempts != 3 {
		t.Errorf("Expected default max attempts 3, got %d", retrier.config.MaxAttempts)
	}
}

func TestRetry_Do_Success(t *testing.T) {
	retrier := NewWithDefaults()
	
	attempts := 0
	ctx := context.Background()
	err := retrier.Do(ctx, func() error {
		attempts++
		return nil
	})
	
	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", attempts)
	}
}

func TestRetry_Do_EventualSuccess(t *testing.T) {
	retrier := NewWithDefaults().WithDelay(10 * time.Millisecond) // Fast for testing
	
	attempts := 0
	ctx := context.Background()
	err := retrier.Do(ctx, func() error {
		attempts++
		if attempts < 3 {
			return errors.New("timeout") // Use a retryable error
		}
		return nil
	})
	
	if err != nil {
		t.Errorf("Expected eventual success, got error: %v", err)
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetry_Do_MaxAttemptsExceeded(t *testing.T) {
	retrier := NewWithDefaults().WithMaxAttempts(2).WithDelay(10 * time.Millisecond)
	
	attempts := 0
	testErr := errors.New("timeout") // Use retryable error
	ctx := context.Background()
	err := retrier.Do(ctx, func() error {
		attempts++
		return testErr
	})
	
	if err == nil {
		t.Error("Expected error after max attempts, got nil")
	}
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestRetry_DoWithResult_Success(t *testing.T) {
	retrier := NewWithDefaults()
	
	ctx := context.Background()
	result, err := retrier.DoWithResult(ctx, func() (interface{}, error) {
		return "success_result", nil
	})
	
	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if result != "success_result" {
		t.Errorf("Expected result 'success_result', got %v", result)
	}
}

func TestRetry_DoWithResult_EventualSuccess(t *testing.T) {
	retrier := NewWithDefaults().WithDelay(10 * time.Millisecond)
	
	attempts := 0
	ctx := context.Background()
	result, err := retrier.DoWithResult(ctx, func() (interface{}, error) {
		attempts++
		if attempts < 3 {
			return nil, errors.New("timeout") // Use retryable error
		}
		return "eventual_success", nil
	})
	
	if err != nil {
		t.Errorf("Expected eventual success, got error: %v", err)
	}
	if result != "eventual_success" {
		t.Errorf("Expected result 'eventual_success', got %v", result)
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetry_CalculateDelay(t *testing.T) {
	config := Config{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
	}
	retrier := New(config)
	
	testCases := []struct {
		attempt     int
		expectMin   time.Duration
		expectMax   time.Duration
	}{
		{1, 100 * time.Millisecond, 200 * time.Millisecond},
		{2, 200 * time.Millisecond, 400 * time.Millisecond},
		{3, 400 * time.Millisecond, 800 * time.Millisecond},
		{4, 800 * time.Millisecond, 1 * time.Second}, // Capped at MaxDelay
	}
	
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Attempt%d", tc.attempt), func(t *testing.T) {
			delay := retrier.calculateDelay(tc.attempt)
			if delay < tc.expectMin || delay > tc.expectMax {
				t.Errorf("Attempt %d: delay %v not in range [%v, %v]", 
					tc.attempt, delay, tc.expectMin, tc.expectMax)
			}
		})
	}
}

func TestRetry_IsRetryable(t *testing.T) {
	config := DefaultConfig()
	retrier := New(config)
	
	testCases := []struct {
		name      string
		err       error
		retryable bool
	}{
		{
			name:      "NilError",
			err:       nil,
			retryable: false,
		},
		{
			name:      "ContextCanceled",
			err:       context.Canceled,
			retryable: true, // context canceled is in the retryable list
		},
		{
			name:      "ContextDeadlineExceeded",
			err:       context.DeadlineExceeded,
			retryable: true, // context deadline exceeded is in the retryable list
		},
		{
			name:      "GenericError",
			err:       errors.New("generic error"),
			retryable: false, // generic error is not in retryable list
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := retrier.isRetryable(tc.err)
			if result != tc.retryable {
				t.Errorf("isRetryable(%v) = %v, want %v", tc.err, result, tc.retryable)
			}
		})
	}
}

func TestRetry_WithMethods(t *testing.T) {
	retrier := NewWithDefaults()
	
	// Test chaining
	modified := retrier.
		WithMaxAttempts(5).
		WithDelay(50 * time.Millisecond).
		WithMaxDelay(2 * time.Second).
		WithMultiplier(1.5).
		WithJitter(true)
	
	if modified.config.MaxAttempts != 5 {
		t.Errorf("Expected max attempts 5, got %d", modified.config.MaxAttempts)
	}
	if modified.config.InitialDelay != 50*time.Millisecond {
		t.Errorf("Expected delay 50ms, got %v", modified.config.InitialDelay)
	}
	if modified.config.MaxDelay != 2*time.Second {
		t.Errorf("Expected max delay 2s, got %v", modified.config.MaxDelay)
	}
	if modified.config.Multiplier != 1.5 {
		t.Errorf("Expected multiplier 1.5, got %f", modified.config.Multiplier)
	}
	if !modified.config.Jitter {
		t.Error("Expected jitter to be enabled")
	}
}

func TestRetry_AddRetryableError(t *testing.T) {
	retrier := NewWithDefaults()
	
	customErr := errors.New("custom retryable error")
	retrier.AddRetryableError(customErr.Error())
	
	// Test that custom error is now retryable
	if !retrier.isRetryable(customErr) {
		t.Error("Custom error should be retryable after adding")
	}
	
	// Test that it actually retries
	attempts := 0
	ctx := context.Background()
	err := retrier.WithMaxAttempts(2).WithDelay(10 * time.Millisecond).Do(ctx, func() error {
		attempts++
		if attempts == 1 {
			return customErr
		}
		return nil
	})
	
	if err != nil {
		t.Errorf("Expected success after retry, got error: %v", err)
	}
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestRetry_WithJitter(t *testing.T) {
	config := Config{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     10 * time.Second, // Set MaxDelay to prevent capping at 0
		Multiplier:   2.0,
		Jitter:       true,
	}
	retrier := New(config)
	
	// Test that jitter produces different delays - use more samples for better reliability
	delays := make([]time.Duration, 50)
	for i := 0; i < 50; i++ {
		delays[i] = retrier.calculateDelay(2)
	}
	
	// Check that delays are within expected range (base delay 400ms + up to 25% jitter)
	baseDelay := 400 * time.Millisecond // 100ms * 2^2
	expectedMin := baseDelay
	expectedMax := time.Duration(float64(baseDelay) * 1.25) // base + 25% jitter
	
	for i, delay := range delays {
		if delay < expectedMin || delay > expectedMax {
			t.Errorf("Delay %d: %v not in expected jitter range [%v, %v]", i, delay, expectedMin, expectedMax)
		}
	}
	
	// Check that at least some delays are different (statistical approach)
	uniqueDelays := make(map[time.Duration]bool)
	for _, delay := range delays {
		uniqueDelays[delay] = true
	}
	
	// With 50 samples and random jitter, we should have multiple unique values
	if len(uniqueDelays) < 3 {
		t.Errorf("Expected at least 3 unique delays with jitter, got %d: %v", len(uniqueDelays), uniqueDelays)
	}
}