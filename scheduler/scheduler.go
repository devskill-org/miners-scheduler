package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/devskill-org/miners-scheduler/entsoe"
	"github.com/devskill-org/miners-scheduler/miners"
	_ "github.com/lib/pq"
)

type MinerScheduler struct {
	// Configuration
	config *Config

	// State
	discoveredMiners       map[string]*miners.AvalonQHost
	pricesMarketData       *entsoe.PublicationMarketData
	pricesMarketDataExpiry time.Time
	isRunning              bool
	stopChan               chan struct{}
	mu                     sync.RWMutex

	// Weather forecast cache
	weatherCache WeatherForecastCache

	// Health server
	healthServer *HealthServer

	// Logging
	logger *log.Logger
}

// NewMinerScheduler creates a new scheduler instance
func NewMinerScheduler(config *Config, logger *log.Logger) *MinerScheduler {
	if logger == nil {
		logger = log.Default()
	}

	scheduler := &MinerScheduler{
		config:           config,
		discoveredMiners: make(map[string]*miners.AvalonQHost),
		stopChan:         make(chan struct{}),
		logger:           logger,
		weatherCache: WeatherForecastCache{
			cacheDuration: 2 * time.Hour,
		},
	}

	return scheduler
}

// NewMinerSchedulerWithHealthCheck creates a new scheduler instance with health check server
func NewMinerSchedulerWithHealthCheck(config *Config, logger *log.Logger) *MinerScheduler {
	scheduler := NewMinerScheduler(config, logger)
	scheduler.healthServer = NewHealthServer(scheduler, config.HealthCheckPort)
	return scheduler
}

// SetConfig updates the configuration for miner management
func (s *MinerScheduler) SetConfig(config *Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
	s.logger.Printf("Configuration updated")
}

// GetConfig returns the current configuration
func (s *MinerScheduler) GetConfig() *Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// GetDiscoveredMiners returns a copy of the currently discovered miners
func (s *MinerScheduler) GetDiscoveredMiners() []*miners.AvalonQHost {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Convert map to slice and return a copy
	minersCopy := make([]*miners.AvalonQHost, 0, len(s.discoveredMiners))
	for _, miner := range s.discoveredMiners {
		minersCopy = append(minersCopy, miner)
	}
	return minersCopy
}

func (s *MinerScheduler) getInitialDelay(now time.Time, delayInterval time.Duration) time.Duration {
	top := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
	delay := now.Sub(top)
	for delay > 0 {
		delay = delay - delayInterval
	}
	return -delay
}

// Start begins the scheduler's periodic task
func (s *MinerScheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is already running")
	}
	s.isRunning = true
	s.mu.Unlock()

	if s.config.DryRun {
		s.logger.Printf("DRY-RUN MODE ENABLED: Actions will be simulated only")
	}

	// Start health server if configured
	if s.healthServer != nil {
		if err := s.healthServer.Start(); err != nil {
			s.logger.Printf("Failed to start health server: %v", err)
		} else {
			s.logger.Printf("Health server started on port %d", s.healthServer.port)
		}
	}

	// Run the first checks immediately
	go func() {
		s.runMinerDiscovery(ctx)
		s.runPriceCheck(ctx)
		s.runMPCOptimize(ctx)
	}()

	config := s.GetConfig()

	// PV integration state
	pvSamples := &PVSamples{}
	var pvDB *sql.DB
	var pvDBErr error
	if s.config.PostgresConnString != "" {
		pvDB, pvDBErr = sql.Open("postgres", s.config.PostgresConnString)
		if pvDBErr != nil {
			s.logger.Printf("PV integration: failed to connect to DB: %v", pvDBErr)
			pvDB = nil
		}
	}

	// Get initial delays
	now := time.Now()
	minersControlInitialDelay := s.getInitialDelay(now, config.CheckPriceInterval) + time.Second
	pvDataInitialDelay := s.getInitialDelay(now, config.PVIntegrationPeriod)
	s.logger.Printf("Schedule next miners check for %v", now.Add(minersControlInitialDelay))
	s.logger.Printf("Schedule next PV data collection for %v", now.Add(pvDataInitialDelay))

	pvTicker := time.NewTicker(config.PVPollInterval)
	defer pvTicker.Stop()

	pvResetTicker := time.NewTicker(config.PVIntegrationPeriod)
	defer pvResetTicker.Stop()

	priceCheckTicker := time.NewTicker(config.CheckPriceInterval)
	defer priceCheckTicker.Stop()

	stateCheckTicker := time.NewTicker(config.MinersStateCheckInterval)
	defer stateCheckTicker.Stop()

	minerDiscoveryTicker := time.NewTicker(config.MinerDiscoveryInterval)
	defer minerDiscoveryTicker.Stop()

	minersControlInitialDelayTick := time.After(minersControlInitialDelay)
	minersControlInitialDelayPassed := false

	pvDataInitialDelayTick := time.After(pvDataInitialDelay)
	pvDataInitialDelayPassed := false

	for {
		select {
		case <-ctx.Done():
			s.logger.Printf("Scheduler stopping due to context cancellation")
			s.stop()
			return ctx.Err()
		case <-s.stopChan:
			s.logger.Printf("Scheduler stopping due to stop signal")
			return nil
		case <-minersControlInitialDelayTick:
			go s.runPriceCheck(ctx)
			go s.runMPCOptimize(ctx)
			priceCheckTicker.Reset(config.CheckPriceInterval)
			minersControlInitialDelayPassed = true
		case <-pvDataInitialDelayTick:
			go s.runPVPoll(pvSamples)
			pvTicker.Reset(config.PVPollInterval)
			pvResetTicker.Reset(config.PVIntegrationPeriod)
			pvDataInitialDelayPassed = true
		case <-priceCheckTicker.C:
			if minersControlInitialDelayPassed {
				go s.runPriceCheck(ctx)
				go s.runMPCOptimize(ctx)
			}
		case <-stateCheckTicker.C:
			go s.runStateCheck(ctx)
		case <-minerDiscoveryTicker.C:
			go s.runMinerDiscovery(ctx)
		case <-pvTicker.C:
			if pvDataInitialDelayPassed {
				go s.runPVPoll(pvSamples)
			}
		case <-pvResetTicker.C:
			if pvDataInitialDelayPassed {
				go s.runPVIntegration(pvSamples, config.PVPollInterval, pvDB, config.DeviceID, config.DryRun)
			}
		}
	}
}

// Stop gracefully stops the scheduler
func (s *MinerScheduler) Stop() {
	s.stop()
}

func (s *MinerScheduler) stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return
	}

	s.isRunning = false
	close(s.stopChan)

	// Stop health server if running
	if s.healthServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.healthServer.Stop(ctx); err != nil {
			s.logger.Printf("Error stopping health server: %v", err)
		}
	}
}

// IsRunning returns whether the scheduler is currently running
func (s *MinerScheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

// GetStatus returns the current status of the scheduler
func (s *MinerScheduler) GetStatus() SchedulerStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return SchedulerStatus{
		IsRunning:     s.isRunning,
		MinersCount:   len(s.discoveredMiners),
		HasMarketData: s.pricesMarketData != nil,
	}
}

// SchedulerStatus represents the current status of the scheduler
type SchedulerStatus struct {
	IsRunning     bool `json:"is_running"`
	MinersCount   int  `json:"miners_count"`
	HasMarketData bool `json:"has_latest_document"`
}
