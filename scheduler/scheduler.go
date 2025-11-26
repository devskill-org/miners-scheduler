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
	"github.com/devskill-org/miners-scheduler/sigenergy"
	_ "github.com/lib/pq"
)

// MinerScheduler handles the periodic task of managing miners based on electricity prices

type PVSample struct {
	value float64
	ts    time.Time
}

type PVSamples struct {
	mu      sync.Mutex
	samples []PVSample
}

func (p *PVSamples) AddSample(value float64, ts time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.samples = append(p.samples, PVSample{value: value, ts: ts})
}

func (p *PVSamples) IntegrateSamples(pollInterval time.Duration) float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	var total float64
	for _, sample := range p.samples {
		total += sample.value * pollInterval.Seconds() / 3600.0 // kW * sec / 3600 = kWh
	}
	p.samples = p.samples[:0]
	return total
}

func (p *PVSamples) IsEmpty() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.samples) == 0
}

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
				go s.runPVIntegration(pvSamples, config.PVPollInterval, pvDB, config.DeviceID)
			}
		}
	}
}

func (s *MinerScheduler) runPVPoll(samples *PVSamples) {
	if s.config.PlantModbusIP == "" {
		return
	}
	client, err := sigenergy.NewTCPClient(s.config.PlantModbusIP, sigenergy.PlantAddress)
	if err != nil {
		s.logger.Printf("PV integration: failed to create modbus client: %v", err)
		return
	}
	defer client.Close()
	info, err := client.ReadPlantRunningInfo()
	if err != nil {
		s.logger.Printf("PV integration: failed to read PlantRunningInfo: %v", err)
		return
	}
	samples.AddSample(info.PhotovoltaicPower, time.Now())
}

func (s *MinerScheduler) runPVIntegration(samples *PVSamples, pollInterval time.Duration, pvDB *sql.DB, deviceID int) {
	if samples.IsEmpty() {
		s.logger.Printf("PV integration: no samples collected in period")
		return
	}
	total := samples.IntegrateSamples(pollInterval)
	timestamp := time.Now()
	if pvDB != nil {
		_, err := pvDB.Exec(
			`INSERT INTO metrics (timestamp, device_id, metric_name, value) VALUES ($1, $2, $3, $4)`,
			timestamp, deviceID, "pv_total_power", total,
		)
		if err != nil {
			s.logger.Printf("PV integration: failed to insert metric: %v", err)
		} else {
			s.logger.Printf("PV integration: saved %.3f kWh for device_id=%d at %s", total, deviceID, timestamp.Format(time.RFC3339))
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
func (s *MinerScheduler) runPriceCheck(ctx context.Context) {
	s.logger.Printf("Starting price check task at %s", time.Now().Format(time.RFC3339))

	// Step 1: Get current electricity price
	currentPrice, err := s.getCurrentAvgPrice(ctx)
	if err != nil {
		s.logger.Printf("Error getting current price: %v", err)
		return
	}

	s.logger.Printf("Current hourly average electricity price: %.2f EUR/MWh", currentPrice)
	s.logger.Printf("Price limit: %.2f EUR/MWh", s.config.PriceLimit)

	// Step 2: Manage miners based on price
	if err := s.manageMiners(ctx, currentPrice); err != nil {
		s.logger.Printf("Error managing miners: %v", err)
		return
	}

	s.logger.Printf("Price check task completed successfully")
}

// discoverMiners discovers Avalon miners on the network and stores them
func (s *MinerScheduler) discoverMiners(ctx context.Context) error {
	s.logger.Printf("Discovering miners on network: %s", s.config.Network)

	newlyDiscoveredMiners := miners.Discover(ctx, s.config.Network)

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
func (s *MinerScheduler) runMinerDiscovery(ctx context.Context) {
	s.logger.Printf("Starting miner discovery task at %s", time.Now().Format(time.RFC3339))

	if err := s.discoverMiners(ctx); err != nil {
		s.logger.Printf("Error discovering miners: %v", err)
		return
	}

	s.logger.Printf("Miner discovery task completed successfully")
}

// getCurrentAvgPrice gets the current hourly average electricity price, downloading new data if needed
func (s *MinerScheduler) getCurrentAvgPrice(ctx context.Context) (float64, error) {
	now := time.Now()

	// Step 2: Try to get price from latest document
	s.mu.RLock()
	latestDoc := s.latestDocument
	s.mu.RUnlock()

	if latestDoc != nil {
		if price, found := latestDoc.LookupAveragePriceInHourByTime(now); found {
			s.logger.Printf("Price found in cached document: %.2f EUR/MWh", price)
			return price, nil
		}
		s.logger.Printf("Price not found in cached document")
	} else {
		s.logger.Printf("No cached document available")
	}

	// Step 3: Download new PublicationMarketDocument
	s.logger.Printf("Downloading new PublicationMarketDocument...")
	newDoc, err := entsoe.DownloadPublicationMarketDocument(ctx, s.config.SecurityToken, s.config.UrlFormat, s.config.Location)
	if err != nil {
		return 0, fmt.Errorf("failed to download PublicationMarketDocument: %w", err)
	}

	// Store as latest
	s.mu.Lock()
	s.latestDocument = newDoc
	s.mu.Unlock()

	s.logger.Printf("Successfully downloaded new PublicationMarketDocument")

	// Try to get price from new document
	if price, found := newDoc.LookupAveragePriceInHourByTime(now); found {
		s.logger.Printf("Price found in new document: %.2f EUR/MWh", price)
		return price, nil
	}

	return 0, fmt.Errorf("price not found in new document for time: %s", now.Format(time.RFC3339))
}

// manageMiners manages miner states based on current price vs price limit
func (s *MinerScheduler) manageMiners(ctx context.Context, currentPrice float64) error {
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
			stats, err := m.GetLiteStats(ctx)
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

						response, err := m.WakeUp(ctx)
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

						response, err := m.Standby(ctx)
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
func (s *MinerScheduler) runStateCheck(ctx context.Context) {
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
			stats, err := m.GetLiteStats(ctx)
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
						response, err := m.SetWorkMode(ctx, miners.AvalonStandardMode, false)
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
						response, err := m.SetWorkMode(ctx, miners.AvalonEcoMode, false)
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
						response, err := m.SetWorkMode(ctx, miners.AvalonStandardMode, true)
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
						response, err := m.SetWorkMode(ctx, miners.AvalonSuperMode, true)
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
