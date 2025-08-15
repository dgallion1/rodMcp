package health

import (
	"context"
	"fmt"
	"rodmcp/internal/logger"
	"sync"
	"time"

	"go.uber.org/zap"
)

// CheckType represents the type of health check
type CheckType string

const (
	CheckTypeBrowser    CheckType = "browser"
	CheckTypeConnection CheckType = "connection"
	CheckTypeMemory     CheckType = "memory"
	CheckTypeCustom     CheckType = "custom"
)

// Status represents the health status
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
	StatusUnknown   Status = "unknown"
)

// Check represents a health check
type Check struct {
	Name        string
	Type        CheckType
	CheckFunc   func() error
	Interval    time.Duration
	Timeout     time.Duration
	Critical    bool // If true, failure affects overall health
	LastCheck   time.Time
	LastStatus  Status
	LastError   error
	FailureCount int
	SuccessCount int
}

// Monitor manages health checks
type Monitor struct {
	logger     *logger.Logger
	checks     map[string]*Check
	checkMutex sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	
	// Overall health tracking
	overallStatus     Status
	lastHealthy       time.Time
	statusListeners   []func(Status)
	listenerMutex     sync.RWMutex
	
	// Configuration
	maxFailures       int
	degradedThreshold int
}

// NewMonitor creates a new health monitor
func NewMonitor(log *logger.Logger) *Monitor {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Monitor{
		logger:            log,
		checks:            make(map[string]*Check),
		ctx:               ctx,
		cancel:            cancel,
		overallStatus:     StatusUnknown,
		lastHealthy:       time.Now(),
		maxFailures:       3,
		degradedThreshold: 2,
	}
}

// RegisterCheck registers a new health check
func (m *Monitor) RegisterCheck(check *Check) {
	m.checkMutex.Lock()
	defer m.checkMutex.Unlock()
	
	if check.Interval == 0 {
		check.Interval = 30 * time.Second
	}
	if check.Timeout == 0 {
		check.Timeout = 5 * time.Second
	}
	
	m.checks[check.Name] = check
	
	m.logger.WithComponent("health").Info("Health check registered",
		zap.String("name", check.Name),
		zap.String("type", string(check.Type)),
		zap.Duration("interval", check.Interval))
}

// Start begins health monitoring
func (m *Monitor) Start() {
	m.logger.WithComponent("health").Info("Starting health monitor")
	
	// Start individual check goroutines
	m.checkMutex.RLock()
	for name, check := range m.checks {
		go m.runCheck(name, check)
	}
	m.checkMutex.RUnlock()
	
	// Start overall health evaluator
	go m.evaluateOverallHealth()
}

// Stop stops health monitoring
func (m *Monitor) Stop() {
	m.logger.WithComponent("health").Info("Stopping health monitor")
	m.cancel()
}

// runCheck runs a single health check periodically
func (m *Monitor) runCheck(name string, check *Check) {
	ticker := time.NewTicker(check.Interval)
	defer ticker.Stop()
	
	// Run initial check
	m.performCheck(check)
	
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.performCheck(check)
		}
	}
}

// performCheck performs a single health check
func (m *Monitor) performCheck(check *Check) {
	ctx, cancel := context.WithTimeout(m.ctx, check.Timeout)
	defer cancel()
	
	// Run check in goroutine to respect timeout
	errCh := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errCh <- fmt.Errorf("check panicked: %v", r)
			}
		}()
		errCh <- check.CheckFunc()
	}()
	
	var err error
	select {
	case err = <-errCh:
		// Check completed
	case <-ctx.Done():
		err = fmt.Errorf("check timed out after %v", check.Timeout)
	}
	
	// Update check status
	m.checkMutex.Lock()
	check.LastCheck = time.Now()
	check.LastError = err
	
	if err == nil {
		check.LastStatus = StatusHealthy
		check.SuccessCount++
		check.FailureCount = 0
		
		m.logger.WithComponent("health").Debug("Health check passed",
			zap.String("name", check.Name))
	} else {
		check.FailureCount++
		check.SuccessCount = 0
		
		if check.FailureCount >= m.maxFailures {
			check.LastStatus = StatusUnhealthy
		} else if check.FailureCount >= m.degradedThreshold {
			check.LastStatus = StatusDegraded
		} else {
			check.LastStatus = StatusHealthy // Still healthy with minor failures
		}
		
		logLevel := zap.DebugLevel
		if check.Critical && check.LastStatus == StatusUnhealthy {
			logLevel = zap.WarnLevel
		}
		
		m.logger.WithComponent("health").Log(logLevel, "Health check failed",
			zap.String("name", check.Name),
			zap.Error(err),
			zap.Int("failures", check.FailureCount))
	}
	m.checkMutex.Unlock()
}

// evaluateOverallHealth evaluates the overall system health
func (m *Monitor) evaluateOverallHealth() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.updateOverallStatus()
		}
	}
}

// updateOverallStatus updates the overall health status
func (m *Monitor) updateOverallStatus() {
	m.checkMutex.RLock()
	defer m.checkMutex.RUnlock()
	
	if len(m.checks) == 0 {
		m.setOverallStatus(StatusUnknown)
		return
	}
	
	criticalUnhealthy := false
	anyDegraded := false
	anyUnhealthy := false
	
	for _, check := range m.checks {
		switch check.LastStatus {
		case StatusUnhealthy:
			anyUnhealthy = true
			if check.Critical {
				criticalUnhealthy = true
			}
		case StatusDegraded:
			anyDegraded = true
		}
	}
	
	var newStatus Status
	if criticalUnhealthy {
		newStatus = StatusUnhealthy
	} else if anyUnhealthy || anyDegraded {
		newStatus = StatusDegraded
	} else {
		newStatus = StatusHealthy
	}
	
	m.setOverallStatus(newStatus)
}

// setOverallStatus sets the overall status and notifies listeners
func (m *Monitor) setOverallStatus(status Status) {
	oldStatus := m.overallStatus
	m.overallStatus = status
	
	if status == StatusHealthy {
		m.lastHealthy = time.Now()
	}
	
	if oldStatus != status {
		m.logger.WithComponent("health").Info("Overall health status changed",
			zap.String("from", string(oldStatus)),
			zap.String("to", string(status)))
		
		// Notify listeners
		m.listenerMutex.RLock()
		listeners := make([]func(Status), len(m.statusListeners))
		copy(listeners, m.statusListeners)
		m.listenerMutex.RUnlock()
		
		for _, listener := range listeners {
			go listener(status)
		}
	}
}

// OnStatusChange registers a callback for status changes
func (m *Monitor) OnStatusChange(callback func(Status)) {
	m.listenerMutex.Lock()
	defer m.listenerMutex.Unlock()
	m.statusListeners = append(m.statusListeners, callback)
}

// GetStatus returns the current overall status
func (m *Monitor) GetStatus() Status {
	m.checkMutex.RLock()
	defer m.checkMutex.RUnlock()
	return m.overallStatus
}

// GetCheckStatus returns the status of a specific check
func (m *Monitor) GetCheckStatus(name string) (*CheckStatus, error) {
	m.checkMutex.RLock()
	defer m.checkMutex.RUnlock()
	
	check, exists := m.checks[name]
	if !exists {
		return nil, fmt.Errorf("check %s not found", name)
	}
	
	return &CheckStatus{
		Name:         check.Name,
		Type:         check.Type,
		Status:       check.LastStatus,
		LastCheck:    check.LastCheck,
		LastError:    check.LastError,
		FailureCount: check.FailureCount,
		SuccessCount: check.SuccessCount,
		Critical:     check.Critical,
	}, nil
}

// GetAllStatuses returns the status of all checks
func (m *Monitor) GetAllStatuses() map[string]*CheckStatus {
	m.checkMutex.RLock()
	defer m.checkMutex.RUnlock()
	
	statuses := make(map[string]*CheckStatus)
	for name, check := range m.checks {
		statuses[name] = &CheckStatus{
			Name:         check.Name,
			Type:         check.Type,
			Status:       check.LastStatus,
			LastCheck:    check.LastCheck,
			LastError:    check.LastError,
			FailureCount: check.FailureCount,
			SuccessCount: check.SuccessCount,
			Critical:     check.Critical,
		}
	}
	
	return statuses
}

// CheckStatus represents the status of a health check
type CheckStatus struct {
	Name         string
	Type         CheckType
	Status       Status
	LastCheck    time.Time
	LastError    error
	FailureCount int
	SuccessCount int
	Critical     bool
}

// GetReport generates a health report
func (m *Monitor) GetReport() *HealthReport {
	m.checkMutex.RLock()
	defer m.checkMutex.RUnlock()
	
	checkStatuses := make(map[string]*CheckStatus)
	for name, check := range m.checks {
		checkStatuses[name] = &CheckStatus{
			Name:         check.Name,
			Type:         check.Type,
			Status:       check.LastStatus,
			LastCheck:    check.LastCheck,
			LastError:    check.LastError,
			FailureCount: check.FailureCount,
			SuccessCount: check.SuccessCount,
			Critical:     check.Critical,
		}
	}
	
	return &HealthReport{
		OverallStatus: m.overallStatus,
		LastHealthy:   m.lastHealthy,
		Checks:        checkStatuses,
		Timestamp:     time.Now(),
	}
}

// HealthReport represents a complete health report
type HealthReport struct {
	OverallStatus Status
	LastHealthy   time.Time
	Checks        map[string]*CheckStatus
	Timestamp     time.Time
}

// IsHealthy returns true if the system is healthy
func (m *Monitor) IsHealthy() bool {
	return m.GetStatus() == StatusHealthy
}

// IsDegraded returns true if the system is degraded
func (m *Monitor) IsDegraded() bool {
	return m.GetStatus() == StatusDegraded
}

// IsUnhealthy returns true if the system is unhealthy
func (m *Monitor) IsUnhealthy() bool {
	return m.GetStatus() == StatusUnhealthy
}