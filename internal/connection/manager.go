package connection

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"rodmcp/internal/logger"
	"strings"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
)

// CircularBuffer implements a thread-safe circular buffer for connection management
type CircularBuffer struct {
	data     []byte
	head     int
	tail     int
	size     int
	capacity int
	mutex    sync.RWMutex
	full     bool
}

// NewCircularBuffer creates a new circular buffer with the specified capacity
func NewCircularBuffer(capacity int) *CircularBuffer {
	return &CircularBuffer{
		data:     make([]byte, capacity),
		capacity: capacity,
	}
}

// Write writes data to the buffer, overwriting old data if full
func (cb *CircularBuffer) Write(data []byte) int {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	written := 0
	for _, b := range data {
		cb.data[cb.head] = b
		cb.head = (cb.head + 1) % cb.capacity
		written++

		if cb.full {
			cb.tail = (cb.tail + 1) % cb.capacity
		} else if cb.head == cb.tail {
			cb.full = true
		}

		if !cb.full {
			cb.size++
		}
	}

	return written
}

// Read reads up to len(p) bytes from the buffer
func (cb *CircularBuffer) Read(p []byte) (int, error) {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	if cb.size == 0 {
		return 0, io.EOF
	}

	read := 0
	for read < len(p) && cb.size > 0 {
		p[read] = cb.data[cb.tail]
		cb.tail = (cb.tail + 1) % cb.capacity
		read++
		cb.size--
		if cb.full {
			cb.full = false
		}
	}

	return read, nil
}

// Size returns the current size of the buffer
func (cb *CircularBuffer) Size() int {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.size
}

// ConnectionManager handles robust stdio connections with automatic recovery
type ConnectionManager struct {
	logger        *logger.Logger
	inputBuffer   *CircularBuffer
	outputBuffer  *CircularBuffer
	mutex         sync.RWMutex
	reconnectCh   chan struct{}
	healthCheck   *time.Ticker
	ctx           context.Context
	cancel        context.CancelFunc
	connected     bool
	lastActivity  time.Time
	activityMutex sync.RWMutex
	
	// Connection stats
	connectionAttempts int64
	reconnectCount     int64
	lastReconnect      time.Time
	
	// Configuration
	config Config
}

// Config defines configuration options for the ConnectionManager
type Config struct {
	// Buffer sizes
	InputBufferSize  int
	OutputBufferSize int
	
	// Timeouts
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	HeartbeatInterval time.Duration
	
	// Reconnection settings
	MaxReconnectAttempts int
	ReconnectBaseDelay   time.Duration
	ReconnectMaxDelay    time.Duration
	
	// Health check settings
	HealthCheckInterval time.Duration
	MaxIdleTime         time.Duration
}

// DefaultConfig returns a default configuration for the ConnectionManager
func DefaultConfig() Config {
	return Config{
		InputBufferSize:      1024 * 1024, // 1MB
		OutputBufferSize:     1024 * 1024, // 1MB
		ReadTimeout:          30 * time.Second,
		WriteTimeout:         30 * time.Second,
		HeartbeatInterval:    30 * time.Second,
		MaxReconnectAttempts: 5,
		ReconnectBaseDelay:   1 * time.Second,
		ReconnectMaxDelay:    30 * time.Second,
		HealthCheckInterval:  10 * time.Second,
		MaxIdleTime:          5 * time.Minute,
	}
}

// NewConnectionManager creates a new ConnectionManager
func NewConnectionManager(log *logger.Logger, config Config) *ConnectionManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &ConnectionManager{
		logger:       log,
		inputBuffer:  NewCircularBuffer(config.InputBufferSize),
		outputBuffer: NewCircularBuffer(config.OutputBufferSize),
		reconnectCh:  make(chan struct{}, 1),
		ctx:          ctx,
		cancel:       cancel,
		lastActivity: time.Now(),
		config:       config,
	}
}

// Start initializes the connection manager
func (cm *ConnectionManager) Start() error {
	cm.logger.WithComponent("connection").Info("Starting connection manager",
		zap.Int("input_buffer_size", cm.config.InputBufferSize),
		zap.Int("output_buffer_size", cm.config.OutputBufferSize))

	// Start health checking
	cm.healthCheck = time.NewTicker(cm.config.HealthCheckInterval)
	
	// Start monitoring goroutines
	go cm.healthCheckLoop()
	go cm.reconnectLoop()
	
	cm.connected = true
	cm.updateActivity()
	
	return nil
}

// Stop gracefully shuts down the connection manager
func (cm *ConnectionManager) Stop() error {
	cm.logger.WithComponent("connection").Info("Stopping connection manager")
	
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	cm.connected = false
	cm.cancel()
	
	if cm.healthCheck != nil {
		cm.healthCheck.Stop()
	}
	
	return nil
}

// ReadMessage reads a message from stdin with timeout and error recovery
func (cm *ConnectionManager) ReadMessage() (string, error) {
	if !cm.isConnected() {
		// Attempt reconnection if not connected
		select {
		case cm.reconnectCh <- struct{}{}:
			// Wait a bit for reconnection attempt
			time.Sleep(100 * time.Millisecond)
		default:
			// Reconnection already in progress
		}
		return "", fmt.Errorf("not connected")
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(cm.ctx, cm.config.ReadTimeout)
	defer cancel()

	// Channel to receive the result
	resultCh := make(chan string, 1)
	errorCh := make(chan error, 1)

	// Read in goroutine with proper signal handling
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errorCh <- fmt.Errorf("read panic: %v", r)
			}
		}()

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Buffer(make([]byte, cm.config.InputBufferSize), cm.config.InputBufferSize)
		
		if scanner.Scan() {
			line := scanner.Text()
			cm.inputBuffer.Write([]byte(line + "\n"))
			cm.updateActivity()
			resultCh <- line
		} else {
			if err := scanner.Err(); err != nil {
				// Check for specific error types
				if isConnectionError(err) {
					cm.handleConnectionLoss(err)
					errorCh <- fmt.Errorf("connection lost: %w", err)
				} else {
					cm.logger.WithComponent("connection").Warn("Scanner error, continuing", zap.Error(err))
					// For non-critical scanner errors, signal to continue instead of failing
					errorCh <- fmt.Errorf("scanner error (recoverable): %w", err)
				}
			} else {
				// EOF - handle gracefully without terminating
				cm.logger.WithComponent("connection").Debug("EOF received, checking connection health")
				if cm.testConnection() {
					// Connection is still healthy, this might be temporary
					cm.logger.WithComponent("connection").Debug("Connection still healthy after EOF, continuing")
					errorCh <- fmt.Errorf("EOF (recoverable)")
				} else {
					// Connection actually lost
					cm.handleConnectionLoss(io.EOF)
					errorCh <- io.EOF
				}
			}
		}
	}()

	select {
	case result := <-resultCh:
		return result, nil
	case err := <-errorCh:
		return "", err
	case <-ctx.Done():
		return "", fmt.Errorf("read timeout after %v", cm.config.ReadTimeout)
	case <-cm.ctx.Done():
		return "", fmt.Errorf("connection manager stopped")
	}
}

// WriteMessage writes a message to stdout with timeout and error recovery
func (cm *ConnectionManager) WriteMessage(message string) error {
	if !cm.isConnected() {
		return fmt.Errorf("not connected")
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(cm.ctx, cm.config.WriteTimeout)
	defer cancel()

	// Channel to receive the result
	errorCh := make(chan error, 1)

	// Write in goroutine with proper signal handling
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errorCh <- fmt.Errorf("write panic: %v", r)
			}
		}()

		data := []byte(message + "\n")
		cm.outputBuffer.Write(data)
		
		// Write to stdout with signal handling
		_, err := os.Stdout.Write(data)
		if err != nil {
			if isConnectionError(err) {
				cm.handleConnectionLoss(err)
				errorCh <- fmt.Errorf("connection lost during write: %w", err)
			} else {
				errorCh <- fmt.Errorf("write error: %w", err)
			}
		} else {
			cm.updateActivity()
			errorCh <- nil
		}
	}()

	select {
	case err := <-errorCh:
		return err
	case <-ctx.Done():
		return fmt.Errorf("write timeout after %v", cm.config.WriteTimeout)
	case <-cm.ctx.Done():
		return fmt.Errorf("connection manager stopped")
	}
}

// healthCheckLoop performs periodic health checks
func (cm *ConnectionManager) healthCheckLoop() {
	for {
		select {
		case <-cm.ctx.Done():
			return
		case <-cm.healthCheck.C:
			cm.performHealthCheck()
		}
	}
}

// performHealthCheck checks connection health
func (cm *ConnectionManager) performHealthCheck() {
	cm.activityMutex.RLock()
	lastActivity := cm.lastActivity
	cm.activityMutex.RUnlock()

	idleTime := time.Since(lastActivity)
	
	if idleTime > cm.config.MaxIdleTime {
		cm.logger.WithComponent("connection").Warn("Connection idle for too long",
			zap.Duration("idle_time", idleTime))
	}

	// Log connection stats periodically
	cm.logger.WithComponent("connection").Debug("Connection health check",
		zap.Duration("idle_time", idleTime),
		zap.Int64("reconnect_count", cm.reconnectCount),
		zap.Int("input_buffer_size", cm.inputBuffer.Size()),
		zap.Int("output_buffer_size", cm.outputBuffer.Size()))
}

// reconnectLoop handles connection reconnection
func (cm *ConnectionManager) reconnectLoop() {
	for {
		select {
		case <-cm.ctx.Done():
			return
		case <-cm.reconnectCh:
			cm.attemptReconnect()
		}
	}
}

// handleConnectionLoss handles when a connection is lost
func (cm *ConnectionManager) handleConnectionLoss(err error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	if !cm.connected {
		return // Already handling
	}
	
	cm.connected = false
	cm.logger.WithComponent("connection").Error("Connection lost",
		zap.Error(err))
	
	// Signal for reconnection
	select {
	case cm.reconnectCh <- struct{}{}:
	default:
		// Channel full, reconnection already queued
	}
}

// attemptReconnect attempts to reconnect with exponential backoff
func (cm *ConnectionManager) attemptReconnect() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	if cm.connected {
		return // Already connected
	}
	
	attempts := int64(0)
	baseDelay := cm.config.ReconnectBaseDelay
	
	for attempts < int64(cm.config.MaxReconnectAttempts) {
		// Calculate exponential backoff delay
		delay := baseDelay * time.Duration(1<<attempts)
		if delay > cm.config.ReconnectMaxDelay {
			delay = cm.config.ReconnectMaxDelay
		}
		
		cm.logger.WithComponent("connection").Info("Attempting reconnection",
			zap.Int64("attempt", attempts+1),
			zap.Duration("delay", delay))
		
		// Wait before retry
		if attempts > 0 {
			select {
			case <-cm.ctx.Done():
				return
			case <-time.After(delay):
			}
		}
		
		// Try to reconnect by testing stdin/stdout
		if cm.testConnection() {
			cm.connected = true
			cm.reconnectCount++
			cm.lastReconnect = time.Now()
			cm.updateActivity()
			
			cm.logger.WithComponent("connection").Info("Reconnection successful",
				zap.Int64("attempt", attempts+1),
				zap.Int64("total_reconnects", cm.reconnectCount))
			return
		}
		
		attempts++
	}
	
	cm.logger.WithComponent("connection").Error("Reconnection failed after all attempts",
		zap.Int64("max_attempts", int64(cm.config.MaxReconnectAttempts)))
}

// testConnection tests if the connection is working
func (cm *ConnectionManager) testConnection() bool {
	// Test by checking if stdin/stdout are still valid
	// This is a simple test - in a more complex scenario we might send a ping
	
	// Check if we can stat stdin
	if stat, err := os.Stdin.Stat(); err != nil {
		cm.logger.WithComponent("connection").Debug("Failed to stat stdin", zap.Error(err))
		return false
	} else {
		// Check if it's a pipe/character device (expected for MCP)
		mode := stat.Mode()
		if mode&os.ModeNamedPipe == 0 && mode&os.ModeCharDevice == 0 {
			cm.logger.WithComponent("connection").Debug("Stdin is not a pipe or character device", zap.String("mode", mode.String()))
		}
	}
	
	// Check if we can stat stdout
	if stat, err := os.Stdout.Stat(); err != nil {
		cm.logger.WithComponent("connection").Debug("Failed to stat stdout", zap.Error(err))
		return false
	} else {
		// Check if it's a pipe/character device (expected for MCP)
		mode := stat.Mode()
		if mode&os.ModeNamedPipe == 0 && mode&os.ModeCharDevice == 0 {
			cm.logger.WithComponent("connection").Debug("Stdout is not a pipe or character device", zap.String("mode", mode.String()))
		}
	}
	
	return true
}

// isConnected checks if the connection is active
func (cm *ConnectionManager) isConnected() bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.connected
}

// updateActivity updates the last activity timestamp
func (cm *ConnectionManager) updateActivity() {
	cm.activityMutex.Lock()
	defer cm.activityMutex.Unlock()
	cm.lastActivity = time.Now()
}

// GetStats returns connection statistics
func (cm *ConnectionManager) GetStats() map[string]interface{} {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	cm.activityMutex.RLock()
	lastActivity := cm.lastActivity
	cm.activityMutex.RUnlock()
	
	return map[string]interface{}{
		"connected":            cm.connected,
		"connection_attempts":  cm.connectionAttempts,
		"reconnect_count":      cm.reconnectCount,
		"last_reconnect":       cm.lastReconnect,
		"last_activity":        lastActivity,
		"idle_time":           time.Since(lastActivity),
		"input_buffer_size":   cm.inputBuffer.Size(),
		"output_buffer_size":  cm.outputBuffer.Size(),
	}
}

// isConnectionError checks if an error is connection-related
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	
	// Check for common connection errors
	errStr := err.Error()
	
	// SIGPIPE and broken pipe errors
	if err == syscall.EPIPE || err == syscall.ECONNRESET {
		return true
	}
	
	// String-based checks for common connection issues
	connectionErrors := []string{
		"broken pipe",
		"connection reset",
		"connection refused",
		"no such device",
		"input/output error",
		"bad file descriptor",
		"EOF (recoverable)",
		"scanner error (recoverable)",
	}
	
	for _, connErr := range connectionErrors {
		if len(errStr) > 0 && len(connErr) > 0 && (errStr == connErr || strings.Contains(errStr, connErr)) {
			return true
		}
	}
	
	return false
}