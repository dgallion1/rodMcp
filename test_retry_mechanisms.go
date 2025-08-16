package main

import (
	"context"
	"fmt"
	"rodmcp/internal/browser"
	"rodmcp/internal/logger"
	"rodmcp/internal/retry"
	"rodmcp/internal/webtools"
	"time"
)

func main() {
	// Create logger
	log, err := logger.New(logger.Config{
		LogLevel: "debug",
		LogDir:   "./test-logs",
	})
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		return
	}
	
	// Create browser config
	config := browser.Config{
		Headless:     true,
		Debug:        false,
		WindowWidth:  1920,
		WindowHeight: 1080,
	}
	
	fmt.Println("Testing Retry Mechanisms")
	fmt.Println("=======================")
	
	// Test 1: Test retry strategy creation
	fmt.Print("1. Testing retry strategy creation... ")
	strategyMgr := retry.WithLogger(log.Logger)
	strategies := strategyMgr.ListStrategies()
	if len(strategies) >= 4 { // Should have at least our 4 default strategies
		fmt.Printf("✓ Created %d strategies\n", len(strategies))
	} else {
		fmt.Printf("✗ Expected 4+ strategies, got %d\n", len(strategies))
		return
	}
	
	// Test 2: Test tool operation strategy (base: 500ms, max: 5s, max attempts: 3)
	fmt.Print("2. Testing tool operation strategy configuration... ")
	toolRetrier := retry.ToolOperationRetrier()
	if toolRetrier != nil {
		fmt.Println("✓")
	} else {
		fmt.Println("✗ Failed to create tool operation retrier")
		return
	}
	
	// Test 3: Test retry with context timeout
	fmt.Print("3. Testing retry with timeout... ")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	start := time.Now()
	err = toolRetrier.Do(ctx, func() error {
		return fmt.Errorf("simulated failure")
	})
	duration := time.Since(start)
	
	if err != nil && duration < 3*time.Second {
		fmt.Printf("✓ Retry failed as expected after %v\n", duration)
	} else {
		fmt.Printf("✗ Unexpected behavior: err=%v, duration=%v\n", err, duration)
	}
	
	// Test 4: Test successful retry
	fmt.Print("4. Testing successful retry after failures... ")
	attempt := 0
	start = time.Now()
	err = toolRetrier.Do(context.Background(), func() error {
		attempt++
		if attempt < 3 {
			return fmt.Errorf("context deadline exceeded - simulated failure %d", attempt)
		}
		return nil // Success on 3rd attempt
	})
	duration = time.Since(start)
	
	if err == nil && attempt == 3 {
		fmt.Printf("✓ Succeeded on attempt %d after %v\n", attempt, duration)
	} else {
		fmt.Printf("✗ Unexpected result: err=%v, attempts=%d\n", err, attempt)
	}
	
	// Test 5: Test retry wrapper with enhanced manager
	fmt.Print("5. Testing retry wrapper integration... ")
	em := browser.NewEnhancedManager(log, config)
	retryWrapper := webtools.NewRetryWrapper(em, log)
	
	if retryWrapper != nil {
		fmt.Println("✓ Retry wrapper created successfully")
	} else {
		fmt.Println("✗ Failed to create retry wrapper")
		return
	}
	
	// Test 6: Test strategy information
	fmt.Print("6. Testing strategy information retrieval... ")
	info := retryWrapper.GetStrategyInfo()
	if len(info) >= 4 {
		fmt.Printf("✓ Retrieved info for %d strategies\n", len(info))
		
		// Print details about tool operation strategy
		if toolInfo, exists := info["tool_operation"]; exists {
			fmt.Printf("   Tool Operation Strategy: %+v\n", toolInfo)
		}
	} else {
		fmt.Printf("✗ Expected 4+ strategy infos, got %d\n", len(info))
	}
	
	// Test 7: Test different strategy types
	fmt.Print("7. Testing different strategy performance... ")
	
	strategies_to_test := []string{"tool_operation", "browser_operation", "critical_operation", "network_operation"}
	for _, strategyName := range strategies_to_test {
		start := time.Now()
		testErr := strategyMgr.RetryWithStrategy(context.Background(), strategyName, "test", func() error {
			return fmt.Errorf("test failure")
		})
		duration := time.Since(start)
		
		if testErr != nil {
			fmt.Printf("\n   %s: failed after %v (expected)", strategyName, duration)
		} else {
			fmt.Printf("\n   %s: unexpected success", strategyName)
		}
	}
	fmt.Println("\n   ✓ All strategies tested")
	
	fmt.Println("\nRetry mechanism testing completed!")
	fmt.Println("=================================")
}