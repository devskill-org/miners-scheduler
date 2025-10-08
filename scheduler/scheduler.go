package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/devskill-org/miners-scheduler/entsoe"
	"github.com/devskill-org/miners-scheduler/miners"
)

// MinerScheduler handles the periodic task of managing miners based on electricity prices
type MinerScheduler struct {
	// Configuration
	config *Config

	// State
	discoveredMiners map[string]*miners.AvalonQHost
	latestDocument   *entsoe.PublicationMarketDocument
	isRunning        bool
	stopChan         chan struct{}
	mu               sync.RWMutex

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

func (s *MinerScheduler) getInitialDelay(now time.Time) time.Duration {
	top := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
	delay := now.Sub(top)
	for delay > 0 {
		delay = delay - s.GetConfig().CheckPriceInterval
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
	s.runMinerDiscovery()
	s.runPriceCheck()

	// Get initial delay
	now := time.Now()
	initialDelay := s.getInitialDelay(now) + time.Second
	s.logger.Printf("Schedule next check for %v", now.Add(initialDelay))

	// Start the periodic tickers for price checking, state checking, and miner discovery
	priceCheckTicker := time.NewTicker(s.config.CheckPriceInterval)
	defer priceCheckTicker.Stop()

	stateCheckTicker := time.NewTicker(s.config.MinersStateCheckInterval)
	defer stateCheckTicker.Stop()

	minerDiscoveryTicker := time.NewTicker(s.config.MinerDiscoveryInterval)
	defer minerDiscoveryTicker.Stop()

	initialDelayTick := time.After(initialDelay)

Loop:
	for {
		select {
		case <-ctx.Done():
			s.logger.Printf("Scheduler stopping due to context cancellation")
			s.stop()
			return ctx.Err()
		case <-s.stopChan:
			s.logger.Printf("Scheduler stopping due to stop signal")
			return nil
		case <-initialDelayTick:
			go s.runPriceCheck()
			priceCheckTicker.Reset(s.GetConfig().CheckPriceInterval)
			break Loop
		case <-stateCheckTicker.C:
			go s.runStateCheck()
		case <-minerDiscoveryTicker.C:
			go s.runMinerDiscovery()
		}
	}

	for {
		select {
		case <-ctx.Done():
			s.logger.Printf("Scheduler stopping due to context cancellation")
			s.stop()
			return ctx.Err()
		case <-s.stopChan:
			s.logger.Printf("Scheduler stopping due to stop signal")
			return nil
		case <-priceCheckTicker.C:
			go s.runPriceCheck()
		case <-stateCheckTicker.C:
			go s.runStateCheck()
		case <-minerDiscoveryTicker.C:
			go s.runMinerDiscovery()
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

// runPriceCheck executes the main scheduler task
func (s *MinerScheduler) runPriceCheck() {
	s.logger.Printf("Starting price check task at %s", time.Now().Format(time.RFC3339))

	// Step 1: Get current electricity price
	currentPrice, err := s.getCurrentPrice()
	if err != nil {
		s.logger.Printf("Error getting current price: %v", err)
		return
	}

	s.logger.Printf("Current electricity price: %.2f EUR/MWh", currentPrice)
	s.logger.Printf("Price limit: %.2f EUR/MWh", s.config.PriceLimit)

	// Step 2: Manage miners based on price
	if err := s.manageMiners(currentPrice); err != nil {
		s.logger.Printf("Error managing miners: %v", err)
		return
	}

	s.logger.Printf("Price check task completed successfully")
}

// discoverMiners discovers Avalon miners on the network and stores them
func (s *MinerScheduler) discoverMiners() error {
	s.logger.Printf("Discovering miners on network: %s", s.config.Network)

	newlyDiscoveredMiners := miners.Discover(s.config.Network)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Add only new miners that don't already exist
	newMinersCount := 0
	for _, newMiner := range newlyDiscoveredMiners {
		key := fmt.Sprintf("%s:%d", newMiner.Address, newMiner.Port)
		if _, exists := s.discoveredMiners[key]; !exists {
			s.discoveredMiners[key] = newMiner
			newMinersCount++
			s.logger.Printf("  New miner discovered: %s:%d", newMiner.Address, newMiner.Port)
		}
	}

	totalMiners := len(s.discoveredMiners)
	s.logger.Printf("Discovery complete: %d total miners (%d newly discovered)", totalMiners, newMinersCount)

	return nil
}

// runMinerDiscovery runs the miner discovery process as a scheduled task
func (s *MinerScheduler) runMinerDiscovery() {
	s.logger.Printf("Starting miner discovery task at %s", time.Now().Format(time.RFC3339))

	if err := s.discoverMiners(); err != nil {
		s.logger.Printf("Error discovering miners: %v", err)
		return
	}

	s.logger.Printf("Miner discovery task completed successfully")
}

// getCurrentPrice gets the current electricity price, downloading new data if needed
func (s *MinerScheduler) getCurrentPrice() (float64, error) {
	now := time.Now()

	// Step 2: Try to get price from latest document
	s.mu.RLock()
	latestDoc := s.latestDocument
	s.mu.RUnlock()

	if latestDoc != nil {
		if price, found := latestDoc.LookupPriceByTime(now); found {
			s.logger.Printf("Price found in cached document: %.2f EUR/MWh", price)
			return price, nil
		}
		s.logger.Printf("Price not found in cached document")
	} else {
		s.logger.Printf("No cached document available")
	}

	// Step 3: Download new PublicationMarketDocument
	s.logger.Printf("Downloading new PublicationMarketDocument...")
	newDoc, err := entsoe.DownloadPublicationMarketDocument(s.config.SecurityToken, s.config.UrlFormat, s.config.Location)
	if err != nil {
		return 0, fmt.Errorf("failed to download PublicationMarketDocument: %w", err)
	}

	// Store as latest
	s.mu.Lock()
	s.latestDocument = newDoc
	s.mu.Unlock()

	s.logger.Printf("Successfully downloaded new PublicationMarketDocument")

	// Try to get price from new document
	if price, found := newDoc.LookupPriceByTime(now); found {
		s.logger.Printf("Price found in new document: %.2f EUR/MWh", price)
		return price, nil
	}

	return 0, fmt.Errorf("price not found in new document for time: %s", now.Format(time.RFC3339))
}

// manageMiners manages miner states based on current price vs price limit
func (s *MinerScheduler) manageMiners(currentPrice float64) error {
	priceLimit := s.config.PriceLimit
	minersList := s.GetDiscoveredMiners()

	if len(minersList) == 0 {
		s.logger.Printf("No miners to manage")
		return nil
	}

	isDryRun := s.config.DryRun
	if isDryRun {
		s.logger.Printf("DRY-RUN MODE: Actions will be simulated only")
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(minersList))

	for _, miner := range minersList {
		wg.Add(1)
		go func(m *miners.AvalonQHost) {
			defer wg.Done()

			// Get current stats
			stats, err := m.GetLiteStats()
			if err != nil {
				errChan <- fmt.Errorf("failed to get stats for miner %s:%d: %w", m.Address, m.Port, err)
				return
			}

			if len(stats.Stats) == 0 || stats.Stats[0].MMIDSummary == nil {
				errChan <- fmt.Errorf("invalid stats response for miner %s:%d", m.Address, m.Port)
				return
			}

			currentState := stats.Stats[0].MMIDSummary.State
			s.logger.Printf("Miner %s:%d current state: %s", m.Address, m.Port, currentState.String())

			// Decision logic based on price comparison
			if currentPrice <= priceLimit {
				// Price is low enough - wake up miners
				if currentState == miners.AvalonStateStandBy {
					if isDryRun {
						s.logger.Printf("DRY-RUN: Would wake up miner %s:%d (price %.2f <= limit %.2f)",
							m.Address, m.Port, currentPrice, priceLimit)
					} else {
						s.logger.Printf("Price (%.2f) <= limit (%.2f), waking up miner %s:%d",
							currentPrice, priceLimit, m.Address, m.Port)

						response, err := m.WakeUp()
						if err != nil {
							errChan <- fmt.Errorf("failed to wake up miner %s:%d: %w", m.Address, m.Port, err)
							return
						}
						s.logger.Printf("WakeUp response for miner %s:%d: %s", m.Address, m.Port, response)
					}
				} else {
					s.logger.Printf("Miner %s:%d is already in %s state, no action needed",
						m.Address, m.Port, currentState.String())
				}
			} else {
				// Price is too high - put active miners into standby
				if currentState != miners.AvalonStateStandBy {
					if isDryRun {
						s.logger.Printf("DRY-RUN: Would put miner %s:%d into standby (price %.2f > limit %.2f)",
							m.Address, m.Port, currentPrice, priceLimit)
					} else {
						s.logger.Printf("Price (%.2f) > limit (%.2f), putting miner %s:%d into standby",
							currentPrice, priceLimit, m.Address, m.Port)

						response, err := m.Standby()
						if err != nil {
							errChan <- fmt.Errorf("failed to put miner %s:%d into standby: %w", m.Address, m.Port, err)
							return
						}
						s.logger.Printf("Standby response for miner %s:%d: %s", m.Address, m.Port, response)
					}
				} else {
					s.logger.Printf("Miner %s:%d is already in standby, no action needed",
						m.Address, m.Port)
				}
			}
		}(miner)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	// Collect any errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		s.logger.Printf("Encountered %d errors while managing miners:", len(errors))
		for _, err := range errors {
			s.logger.Printf("  - %v", err)
		}
		return fmt.Errorf("encountered %d errors while managing miners", len(errors))
	}

	if isDryRun {
		s.logger.Printf("DRY-RUN: Successfully simulated management of %d miners", len(minersList))
	} else {
		s.logger.Printf("Successfully managed %d miners", len(minersList))
	}
	return nil
}

// GetLatestDocument returns the latest PublicationMarketDocument
func (s *MinerScheduler) GetLatestDocument() *entsoe.PublicationMarketDocument {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.latestDocument
}

// GetStatus returns the current status of the scheduler
func (s *MinerScheduler) GetStatus() SchedulerStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return SchedulerStatus{
		IsRunning:        s.isRunning,
		MinersCount:      len(s.discoveredMiners),
		HasLatestDoc:     s.latestDocument != nil,
		LastDocumentTime: s.getLastDocumentTime(),
	}
}

// getLastDocumentTime returns the creation time of the latest document
func (s *MinerScheduler) getLastDocumentTime() *time.Time {
	if s.latestDocument == nil {
		return nil
	}

	if t, err := time.Parse(time.RFC3339, s.latestDocument.CreatedDateTime); err == nil {
		return &t
	}

	return nil
}

// runStateCheck executes the state monitoring task for miners
func (s *MinerScheduler) runStateCheck() {
	minersList := s.GetDiscoveredMiners()
	if len(minersList) == 0 {
		return
	}

	isDryRun := s.config.DryRun

	var wg sync.WaitGroup
	errChan := make(chan error, len(minersList))

	for _, miner := range minersList {
		wg.Add(1)
		go func(m *miners.AvalonQHost) {
			defer wg.Done()

			// Get current stats
			stats, err := m.GetLiteStats()
			if err != nil {
				errChan <- fmt.Errorf("failed to get stats for miner %s:%d: %w", m.Address, m.Port, err)
				return
			}

			if len(stats.Stats) == 0 || stats.Stats[0].MMIDSummary == nil {
				errChan <- fmt.Errorf("no stats response for miner %s:%d", m.Address, m.Port)
				return
			}

			fanR := stats.Stats[0].MMIDSummary.FanR
			currentWorkMode := stats.Stats[0].MMIDSummary.WorkMode
			currentState := stats.Stats[0].MMIDSummary.State
			hbiTemp := stats.Stats[0].MMIDSummary.HBITemp
			hboTemp := stats.Stats[0].MMIDSummary.HBOTemp
			iTemp := stats.Stats[0].MMIDSummary.ITemp

			if currentState != miners.AvalonStateMining {
				return
			}

			miner.AddLiteStats(stats.Stats[0].MMIDSummary)

			s.logger.Printf("Miner %s:%d - FanR: %d%%, HBITemp:%d, HBOTemp:%d, ITemp:%d, WorkMode: %d",
				m.Address,
				m.Port,
				fanR,
				hbiTemp,
				hboTemp,
				iTemp,
				currentWorkMode)

			if fanR > s.config.FanRHighThreshold {
				// Decrease work mode
				if currentWorkMode == int(miners.AvalonSuperMode) {
					// Super -> Standard
					if isDryRun {
						s.logger.Printf("DRY-RUN: Would set miner %s:%d to Standard mode (FanR %d%% > %d%%)",
							m.Address, m.Port, fanR, s.config.FanRHighThreshold)
					} else {
						s.logger.Printf("FanR (%d%%) > %d%%, setting miner %s:%d to Standard work mode",
							fanR, s.config.FanRHighThreshold, m.Address, m.Port)
						response, err := m.SetWorkMode(miners.AvalonStandardMode, false)
						if err != nil {
							errChan <- fmt.Errorf("failed to set miner %s:%d to Standard mode: %w", m.Address, m.Port, err)
							return
						}
						s.logger.Printf("SetWorkMode response for miner %s:%d: %s", m.Address, m.Port, response)
					}
				} else if currentWorkMode == int(miners.AvalonStandardMode) {
					// Standard -> Eco
					if isDryRun {
						s.logger.Printf("DRY-RUN: Would set miner %s:%d to Eco mode (FanR %d%% > %d%%)",
							m.Address, m.Port, fanR, s.config.FanRHighThreshold)
					} else {
						s.logger.Printf("FanR (%d%%) > %d%%, setting miner %s:%d to Eco work mode",
							fanR, s.config.FanRHighThreshold, m.Address, m.Port)
						response, err := m.SetWorkMode(miners.AvalonEcoMode, false)
						if err != nil {
							errChan <- fmt.Errorf("failed to set miner %s:%d to Eco mode: %w", m.Address, m.Port, err)
							return
						}
						s.logger.Printf("SetWorkMode response for miner %s:%d: %s", m.Address, m.Port, response)
					}
				}
			} else if fanR < s.config.FanRLowThreshold {
				// Increase work mode only if all LiteStatsHistory fanR values match criteria
				if len(miner.LiteStatsHistory) < 5 {
					return
				}
				for _, stat := range miner.LiteStatsHistory {
					if stat.FanR >= s.config.FanRLowThreshold {
						return
					}
				}
				if currentWorkMode == int(miners.AvalonEcoMode) {
					// Eco -> Standard
					if isDryRun {
						s.logger.Printf("DRY-RUN: Would set miner %s:%d to Standard mode (all FanR < %d%%)",
							m.Address, m.Port, s.config.FanRLowThreshold)
					} else {
						s.logger.Printf("All FanR < %d%%, setting miner %s:%d to Standard work mode",
							s.config.FanRLowThreshold, m.Address, m.Port)
						response, err := m.SetWorkMode(miners.AvalonStandardMode, true)
						if err != nil {
							errChan <- fmt.Errorf("failed to set miner %s:%d to Standard mode: %w", m.Address, m.Port, err)
							return
						}
						s.logger.Printf("SetWorkMode response for miner %s:%d: %s", m.Address, m.Port, response)
					}
				} else if currentWorkMode == int(miners.AvalonStandardMode) {
					// Standard -> Super
					if isDryRun {
						s.logger.Printf("DRY-RUN: Would set miner %s:%d to Super mode (all FanR < %d%%)",
							m.Address, m.Port, s.config.FanRLowThreshold)
					} else {
						s.logger.Printf("All FanR < %d%%, setting miner %s:%d to Super work mode",
							s.config.FanRLowThreshold, m.Address, m.Port)
						response, err := m.SetWorkMode(miners.AvalonSuperMode, true)
						if err != nil {
							errChan <- fmt.Errorf("failed to set miner %s:%d to Super mode: %w", m.Address, m.Port, err)
							return
						}
						s.logger.Printf("SetWorkMode response for miner %s:%d: %s", m.Address, m.Port, response)
					}
				}
				// If already Super, do nothing
			}
		}(miner)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	// Collect any errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		s.logger.Printf("Encountered %d errors during state check:", len(errors))
		for _, err := range errors {
			s.logger.Printf("  - %v", err)
		}
	}
}

// SchedulerStatus represents the current status of the scheduler
type SchedulerStatus struct {
	IsRunning        bool       `json:"is_running"`
	MinersCount      int        `json:"miners_count"`
	HasLatestDoc     bool       `json:"has_latest_document"`
	LastDocumentTime *time.Time `json:"last_document_time,omitempty"`
}
