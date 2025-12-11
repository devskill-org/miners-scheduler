package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/devskill-org/miners-scheduler/entsoe"
)

// GetLatestDocument returns the latest PublicationMarketDocument
func (s *MinerScheduler) GetLatestDocument() *entsoe.PublicationMarketDocument {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.latestDocument
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
