package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/devskill-org/miners-scheduler/entsoe"
)

// GetMarketData returns the latest PublicationMarketData, downloading new data if needed
func (s *MinerScheduler) GetMarketData(ctx context.Context) (*entsoe.PublicationMarketData, error) {
	now := time.Now()

	s.mu.RLock()
	marketData := s.pricesMarketData
	expiry := s.pricesMarketDataExpiry
	s.mu.RUnlock()

	// Check if we have cached data and it hasn't expired
	if marketData != nil && now.Before(expiry) {
		return marketData, nil
	}

	// Cache expired or no cached document, download new data
	if marketData != nil {
		s.logger.Printf("Cached pricing data expired at %s, downloading new PublicationMarketData...", expiry.Format(time.RFC3339))
	} else {
		s.logger.Printf("No cached pricing data available, downloading new PublicationMarketData...")
	}

	location, err := time.LoadLocation(s.config.Location)
	if err != nil {
		return nil, err
	}

	newDoc, err := entsoe.DownloadPublicationMarketData(ctx, s.config.SecurityToken, s.config.UrlFormat, location)
	if err != nil {
		return nil, fmt.Errorf("failed to download PublicationMarketData: %w", err)
	}

	// Calculate next expiry time at 13:00
	nextExpiry := time.Date(now.Year(), now.Month(), now.Day(), 13, 0, 0, 0, location)

	// If it's already past 13:00 today, set expiry to 13:00 tomorrow
	if now.Hour() >= 13 {
		nextExpiry = nextExpiry.Add(24 * time.Hour)
	}

	// Store as latest with expiry time
	s.mu.Lock()
	s.pricesMarketData = newDoc
	s.pricesMarketDataExpiry = nextExpiry
	s.mu.Unlock()

	s.logger.Printf("Successfully downloaded new PublicationMarketData, cache expires at %s", nextExpiry.Format(time.RFC3339))
	return newDoc, nil
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

// getCurrentAvgPrice gets the current hourly average electricity price, downloading new data if needed
func (s *MinerScheduler) getCurrentAvgPrice(ctx context.Context) (float64, error) {
	now := time.Now()

	marketData, err := s.GetMarketData(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get market prices: %w", err)
	}

	if price, found := marketData.LookupAveragePriceInHourByTime(now); found {
		s.logger.Printf("Price found: %.2f EUR/MWh", price)
		return price, nil
	}

	return 0, fmt.Errorf("price not found for time: %s", now.Format(time.RFC3339))
}
