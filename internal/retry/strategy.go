package retry

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Strategy defines different retry strategies for various operations
type Strategy struct {
	Name        string
	Config      Config
	Description string
}

// Pre-defined retry strategies as specified in enhancement plan
var (
	// ToolOperationStrategy - For tool operations (base: 500ms, max: 5s, max attempts: 3)
	ToolOperationStrategy = Strategy{
		Name:        "tool_operation",
		Description: "Strategy for general tool operations with moderate retry",
		Config: Config{
			MaxAttempts:  3,
			InitialDelay: 500 * time.Millisecond,
			MaxDelay:     5 * time.Second,
			Multiplier:   2.0,
			Jitter:       true,
			RetryableErrors: []string{
				"context canceled",
				"context cancelled",
				"context deadline exceeded",
				"context timeout",
				"timeout",
				"connection reset",
				"broken pipe",
				"target closed",
				"browser not started",
				"browser connection unhealthy",
				"page not found",
				"websocket: close",
				"connection refused",
				"network unreachable",
				"no such host",
				"operation was canceled",
				"operation was cancelled",
				"stale element reference",
				"element not found",
				"element not interactable",
			},
		},
	}

	// BrowserOperationStrategy - For browser-level operations (faster retry)
	BrowserOperationStrategy = Strategy{
		Name:        "browser_operation",
		Description: "Strategy for browser operations requiring quick recovery",
		Config: Config{
			MaxAttempts:  5,
			InitialDelay: 250 * time.Millisecond,
			MaxDelay:     3 * time.Second,
			Multiplier:   1.5,
			Jitter:       true,
			RetryableErrors: []string{
				"context canceled",
				"context cancelled",
				"context deadline exceeded",
				"browser not started",
				"browser connection unhealthy",
				"connection reset",
				"broken pipe",
				"websocket: close",
				"connection refused",
			},
		},
	}

	// CriticalOperationStrategy - For critical operations (minimal retry)
	CriticalOperationStrategy = Strategy{
		Name:        "critical_operation",
		Description: "Strategy for critical operations that should fail fast",
		Config: Config{
			MaxAttempts:  2,
			InitialDelay: 100 * time.Millisecond,
			MaxDelay:     1 * time.Second,
			Multiplier:   2.0,
			Jitter:       false,
			RetryableErrors: []string{
				"context canceled",
				"context cancelled",
				"context deadline exceeded",
			},
		},
	}

	// NetworkOperationStrategy - For network-related operations (longer retry)
	NetworkOperationStrategy = Strategy{
		Name:        "network_operation",
		Description: "Strategy for network operations that may have intermittent failures",
		Config: Config{
			MaxAttempts:  4,
			InitialDelay: 1 * time.Second,
			MaxDelay:     10 * time.Second,
			Multiplier:   2.5,
			Jitter:       true,
			RetryableErrors: []string{
				"timeout",
				"connection reset",
				"broken pipe",
				"connection refused",
				"network unreachable",
				"no such host",
				"temporary failure in name resolution",
				"no route to host",
				"connection timed out",
			},
		},
	}
)

// StrategyManager manages different retry strategies
type StrategyManager struct {
	strategies map[string]Strategy
	logger     *zap.Logger
}

// NewStrategyManager creates a new strategy manager
func NewStrategyManager(logger *zap.Logger) *StrategyManager {
	sm := &StrategyManager{
		strategies: make(map[string]Strategy),
		logger:     logger,
	}

	// Register default strategies
	sm.RegisterStrategy(ToolOperationStrategy)
	sm.RegisterStrategy(BrowserOperationStrategy)
	sm.RegisterStrategy(CriticalOperationStrategy)
	sm.RegisterStrategy(NetworkOperationStrategy)

	return sm
}

// RegisterStrategy registers a new retry strategy
func (sm *StrategyManager) RegisterStrategy(strategy Strategy) {
	sm.strategies[strategy.Name] = strategy
}

// GetStrategy retrieves a strategy by name
func (sm *StrategyManager) GetStrategy(name string) (Strategy, error) {
	strategy, exists := sm.strategies[name]
	if !exists {
		return Strategy{}, fmt.Errorf("strategy '%s' not found", name)
	}
	return strategy, nil
}

// ListStrategies returns all available strategies
func (sm *StrategyManager) ListStrategies() []Strategy {
	strategies := make([]Strategy, 0, len(sm.strategies))
	for _, strategy := range sm.strategies {
		strategies = append(strategies, strategy)
	}
	return strategies
}

// CreateRetrier creates a retrier for the specified strategy
func (sm *StrategyManager) CreateRetrier(strategyName string) (*Retrier, error) {
	strategy, err := sm.GetStrategy(strategyName)
	if err != nil {
		return nil, err
	}
	return New(strategy.Config), nil
}

// RetryWithStrategy executes a function with the specified retry strategy
func (sm *StrategyManager) RetryWithStrategy(ctx context.Context, strategyName string, operation string, fn RetryableFunc) error {
	retrier, err := sm.CreateRetrier(strategyName)
	if err != nil {
		return fmt.Errorf("failed to create retrier for strategy '%s': %w", strategyName, err)
	}

	if sm.logger != nil {
		sm.logger.Debug("Starting retry operation",
			zap.String("strategy", strategyName),
			zap.String("operation", operation))
	}

	err = retrier.Do(ctx, fn)
	
	if sm.logger != nil {
		if err != nil {
			sm.logger.Warn("Retry operation failed",
				zap.String("strategy", strategyName),
				zap.String("operation", operation),
				zap.Error(err))
		} else {
			sm.logger.Debug("Retry operation succeeded",
				zap.String("strategy", strategyName),
				zap.String("operation", operation))
		}
	}

	return err
}

// RetryWithStrategyAndResult executes a function with the specified retry strategy and returns a result
func (sm *StrategyManager) RetryWithStrategyAndResult(ctx context.Context, strategyName string, operation string, fn RetryableWithResultFunc) (interface{}, error) {
	retrier, err := sm.CreateRetrier(strategyName)
	if err != nil {
		return nil, fmt.Errorf("failed to create retrier for strategy '%s': %w", strategyName, err)
	}

	if sm.logger != nil {
		sm.logger.Debug("Starting retry operation with result",
			zap.String("strategy", strategyName),
			zap.String("operation", operation))
	}

	result, err := retrier.DoWithResult(ctx, fn)
	
	if sm.logger != nil {
		if err != nil {
			sm.logger.Warn("Retry operation with result failed",
				zap.String("strategy", strategyName),
				zap.String("operation", operation),
				zap.Error(err))
		} else {
			sm.logger.Debug("Retry operation with result succeeded",
				zap.String("strategy", strategyName),
				zap.String("operation", operation))
		}
	}

	return result, err
}

// ToolOperationRetrier creates a retrier specifically for tool operations
// This follows the enhancement plan specification: base: 500ms, max: 5s, max attempts: 3
func ToolOperationRetrier() *Retrier {
	return New(ToolOperationStrategy.Config)
}

// BrowserOperationRetrier creates a retrier for browser operations
func BrowserOperationRetrier() *Retrier {
	return New(BrowserOperationStrategy.Config)
}

// CriticalOperationRetrier creates a retrier for critical operations
func CriticalOperationRetrier() *Retrier {
	return New(CriticalOperationStrategy.Config)
}

// NetworkOperationRetrier creates a retrier for network operations
func NetworkOperationRetrier() *Retrier {
	return New(NetworkOperationStrategy.Config)
}

// WithLogger creates a new strategy manager with logging
func WithLogger(logger *zap.Logger) *StrategyManager {
	return NewStrategyManager(logger)
}

// WithoutLogger creates a new strategy manager without logging
func WithoutLogger() *StrategyManager {
	return NewStrategyManager(nil)
}

// IsRetryableError checks if an error is retryable using any of the registered strategies
func (sm *StrategyManager) IsRetryableError(err error, strategyName string) bool {
	if err == nil {
		return false
	}

	strategy, exists := sm.strategies[strategyName]
	if !exists {
		// Default to tool operation strategy
		strategy = ToolOperationStrategy
	}

	errStr := strings.ToLower(err.Error())
	for _, retryableErr := range strategy.Config.RetryableErrors {
		if strings.Contains(errStr, strings.ToLower(retryableErr)) {
			return true
		}
	}

	return false
}

// GetStrategyInfo returns information about a strategy
func (sm *StrategyManager) GetStrategyInfo(strategyName string) (map[string]interface{}, error) {
	strategy, err := sm.GetStrategy(strategyName)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":           strategy.Name,
		"description":    strategy.Description,
		"max_attempts":   strategy.Config.MaxAttempts,
		"initial_delay":  strategy.Config.InitialDelay.String(),
		"max_delay":      strategy.Config.MaxDelay.String(),
		"multiplier":     strategy.Config.Multiplier,
		"jitter":         strategy.Config.Jitter,
		"retryable_errors": strategy.Config.RetryableErrors,
	}, nil
}