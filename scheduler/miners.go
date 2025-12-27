package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/devskill-org/miners-scheduler/miners"
)

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

// refreshMinersState refreshes the state of all discovered miners and returns miners list
func (s *MinerScheduler) refreshMinersState(ctx context.Context) []*miners.AvalonQHost {
	var wg sync.WaitGroup
	minersList := s.GetDiscoveredMiners()
	for _, miner := range minersList {
		wg.Add(1)
		go func(m *miners.AvalonQHost) {
			defer wg.Done()

			// Get current stats
			m.RefreshLiteStats(ctx)
		}(miner)
	}
	wg.Wait()
	return minersList
}

// manageMiners manages miner states based on current price vs price limit
func (s *MinerScheduler) manageMiners(ctx context.Context, currentPrice float64) error {
	priceLimit := s.config.PriceLimit
	minersList := s.refreshMinersState(ctx)

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
			if m.LastStatsError != nil {
				errChan <- m.LastStatsError
				return
			}

			currentState := m.LastStats.State
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

// runStateCheck executes the state monitoring task for miners
func (s *MinerScheduler) runStateCheck(ctx context.Context) {
	minersList := s.refreshMinersState(ctx)
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
			if m.LastStatsError != nil {
				errChan <- m.LastStatsError
				return
			}

			fanR := m.LastStats.FanR
			currentWorkMode := m.LastStats.WorkMode
			currentState := m.LastStats.State
			hbiTemp := m.LastStats.HBITemp
			hboTemp := m.LastStats.HBOTemp
			iTemp := m.LastStats.ITemp

			if currentState != miners.AvalonStateMining {
				return
			}

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
