package browser

import (
	"sync"
	"testing"
	"time"

	"rodmcp/internal/logger"
)

// TestBrowserPanicDetection tests that the browser manager properly handles panics
// This is an integration test that uses real browser instances
func TestBrowserPanicDetection(t *testing.T) {
	log, err := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	
	config := Config{
		Headless:     true,
		WindowWidth:  800,
		WindowHeight: 600,
	}
	
	// Create manager
	manager := NewManager(log, config)
	
	// Track panics that occur
	panicCount := 0
	panicMutex := &sync.Mutex{}
	
	// Start the browser
	err = manager.Start(config)
	if err != nil {
		t.Fatalf("Failed to start browser: %v", err)
	}
	defer manager.Stop()
	
	// Test 1: Normal health check should work
	t.Run("NormalHealthCheck", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				panicMutex.Lock()
				panicCount++
				panicMutex.Unlock()
				t.Errorf("Normal health check panicked: %v", r)
			}
		}()
		
		err := manager.CheckHealth()
		if err != nil {
			t.Errorf("Health check failed on healthy browser: %v", err)
		}
	})
	
	// Test 2: Force browser to close and test health check
	t.Run("ClosedBrowserHealthCheck", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				panicMutex.Lock()
				panicCount++
				panicMutex.Unlock()
				t.Errorf("Closed browser health check panicked: %v", r)
			}
		}()
		
		// Force close the browser connection
		manager.mutex.Lock()
		if manager.browser != nil {
			// Use MustClose to force close the browser
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Browser close can panic if already closed
						t.Logf("Browser close panicked (expected): %v", r)
					}
				}()
				manager.browser.MustClose()
			}()
			// Don't set to nil - let the health check detect the closed state
		}
		manager.mutex.Unlock()
		
		// Give browser time to close
		time.Sleep(100 * time.Millisecond)
		
		// Health check should fail but not panic
		err := manager.CheckHealth()
		if err == nil {
			t.Error("Health check should have failed on closed browser")
		}
		
		// Restart for next test
		err = manager.Start(config)
		if err != nil {
			t.Logf("Failed to restart browser: %v", err)
		}
	})
	
	// Test 3: Concurrent health checks during restart
	t.Run("ConcurrentHealthChecksDuringRestart", func(t *testing.T) {
		var wg sync.WaitGroup
		stopChan := make(chan struct{})
		healthCheckPanics := 0
		
		// Start multiple goroutines doing health checks
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				ticker := time.NewTicker(50 * time.Millisecond)
				defer ticker.Stop()
				
				for {
					select {
					case <-stopChan:
						return
					case <-ticker.C:
						func() {
							defer func() {
								if r := recover(); r != nil {
									panicMutex.Lock()
									healthCheckPanics++
									panicMutex.Unlock()
									t.Logf("Health check goroutine %d panicked: %v", id, r)
								}
							}()
							_ = manager.CheckHealth()
						}()
					}
				}
			}(i)
		}
		
		// Let health checks run
		time.Sleep(200 * time.Millisecond)
		
		// Trigger a browser failure
		t.Log("Triggering browser failure...")
		manager.mutex.Lock()
		if manager.browser != nil {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Logf("Browser close panicked during restart test: %v", r)
					}
				}()
				manager.browser.MustClose()
			}()
			manager.browser = nil
		}
		manager.mutex.Unlock()
		
		// Let health checks detect the issue and trigger restart
		time.Sleep(300 * time.Millisecond)
		
		// Try to restart browser
		err := manager.Start(config)
		if err != nil {
			t.Logf("Failed to restart browser: %v", err)
		}
		
		// Let health checks run on new browser
		time.Sleep(200 * time.Millisecond)
		
		// Stop health check goroutines
		close(stopChan)
		wg.Wait()
		
		if healthCheckPanics > 0 {
			panicMutex.Lock()
			panicCount += healthCheckPanics
			panicMutex.Unlock()
			t.Errorf("Concurrent health checks caused %d panics", healthCheckPanics)
		}
	})
	
	// Test 4: Rapid restart attempts
	t.Run("RapidRestartAttempts", func(t *testing.T) {
		var wg sync.WaitGroup
		restartPanics := 0
		
		// Try to trigger concurrent restarts
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						panicMutex.Lock()
						restartPanics++
						panicMutex.Unlock()
						t.Logf("Restart goroutine %d panicked: %v", id, r)
					}
				}()
				
				// Force unhealthy state
				manager.mutex.Lock()
				if manager.browser != nil {
					func() {
						defer func() {
							if r := recover(); r != nil {
								// Expected if browser already closed
							}
						}()
						manager.browser.MustClose()
					}()
				}
				manager.mutex.Unlock()
				
				// Try to ensure healthy (which should trigger restart)
				manager.EnsureHealthy()
			}(i)
			time.Sleep(10 * time.Millisecond) // Small delay between attempts
		}
		
		wg.Wait()
		
		if restartPanics > 0 {
			panicMutex.Lock()
			panicCount += restartPanics
			panicMutex.Unlock()
			t.Errorf("Rapid restart attempts caused %d panics", restartPanics)
		}
		
		// Browser should be healthy after all restart attempts
		time.Sleep(500 * time.Millisecond)
		err := manager.CheckHealth()
		if err != nil {
			// Try to restart if unhealthy
			manager.Start(config)
		}
	})
	
	// Test 5: Nil browser reference handling
	t.Run("NilBrowserReference", func(t *testing.T) {
		nilRefPanics := 0
		
		// Set browser to nil directly
		manager.mutex.Lock()
		manager.browser = nil
		manager.mutex.Unlock()
		
		// Health check should fail gracefully
		func() {
			defer func() {
				if r := recover(); r != nil {
					nilRefPanics++
					t.Errorf("Health check on nil browser panicked: %v", r)
				}
			}()
			
			err := manager.CheckHealth()
			if err == nil {
				t.Error("Health check should fail with nil browser")
			}
		}()
		
		// Multiple health checks on nil browser
		for i := 0; i < 5; i++ {
			func() {
				defer func() {
					if r := recover(); r != nil {
						nilRefPanics++
						t.Logf("Health check %d on nil browser panicked: %v", i, r)
					}
				}()
				_ = manager.CheckHealth()
			}()
		}
		
		if nilRefPanics > 0 {
			panicMutex.Lock()
			panicCount += nilRefPanics
			panicMutex.Unlock()
			t.Errorf("Nil browser checks caused %d panics", nilRefPanics)
		}
	})
	
	// Test 6: Check internal panic recovery in CheckHealth
	t.Run("InternalPanicRecovery", func(t *testing.T) {
		// This test verifies that CheckHealth has proper panic recovery
		internalPanics := 0
		
		// Create a situation that might cause internal panics
		manager.mutex.Lock()
		oldBrowser := manager.browser
		manager.browser = nil
		manager.mutex.Unlock()
		
		// Run CheckHealth multiple times
		for i := 0; i < 10; i++ {
			func() {
				defer func() {
					if r := recover(); r != nil {
						internalPanics++
						t.Errorf("CheckHealth leaked panic on iteration %d: %v", i, r)
					}
				}()
				
				// This should NOT panic externally
				_ = manager.CheckHealth()
			}()
		}
		
		// Restore browser
		manager.mutex.Lock()
		manager.browser = oldBrowser
		manager.mutex.Unlock()
		
		if internalPanics > 0 {
			panicMutex.Lock()
			panicCount += internalPanics
			panicMutex.Unlock()
		}
	})
	
	// Final check: Report any panics
	panicMutex.Lock()
	finalPanicCount := panicCount
	panicMutex.Unlock()
	
	if finalPanicCount > 0 {
		t.Fatalf("TEST FAILED: Detected %d panics during execution. The browser manager is NOT properly handling panics!", finalPanicCount)
	} else {
		t.Log("SUCCESS: No panics detected. Browser manager properly handles all error conditions.")
	}
}

// TestBrowserStressWithPanicDetection performs a stress test while monitoring for panics
func TestBrowserStressWithPanicDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}
	
	log, err := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	
	config := Config{
		Headless:     true,
		WindowWidth:  800,
		WindowHeight: 600,
	}
	
	manager := NewManager(log, config)
	
	// Track panics
	panicCount := 0
	panicMutex := &sync.Mutex{}
	
	// Start browser
	err = manager.Start(config)
	if err != nil {
		t.Fatalf("Failed to start browser: %v", err)
	}
	defer manager.Stop()
	
	// Run stress test
	var wg sync.WaitGroup
	stopChan := make(chan struct{})
	
	// Health check goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		
		for {
			select {
			case <-stopChan:
				return
			case <-ticker.C:
				func() {
					defer func() {
						if r := recover(); r != nil {
							panicMutex.Lock()
							panicCount++
							panicMutex.Unlock()
							t.Logf("Health check panicked during stress: %v", r)
						}
					}()
					_ = manager.CheckHealth()
				}()
			}
		}
	}()
	
	// Browser killer goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		
		for {
			select {
			case <-stopChan:
				return
			case <-ticker.C:
				func() {
					defer func() {
						if r := recover(); r != nil {
							// Browser operations can panic
							t.Logf("Browser operation panicked (expected): %v", r)
						}
					}()
					
					// Randomly kill browser
					manager.mutex.Lock()
					if manager.browser != nil {
						manager.browser.MustClose()
						manager.browser = nil
					}
					manager.mutex.Unlock()
					
					// Try to restart
					time.Sleep(100 * time.Millisecond)
					_ = manager.Start(config)
				}()
			}
		}
	}()
	
	// Run for 3 seconds
	time.Sleep(3 * time.Second)
	
	// Stop all goroutines
	close(stopChan)
	wg.Wait()
	
	// Check results
	panicMutex.Lock()
	finalPanicCount := panicCount
	panicMutex.Unlock()
	
	if finalPanicCount > 0 {
		t.Errorf("Stress test detected %d panics", finalPanicCount)
	} else {
		t.Log("Stress test completed without panics")
	}
}