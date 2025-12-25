package scheduler

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/devskill-org/miners-scheduler/entsoe"
	"github.com/devskill-org/miners-scheduler/miners"
)

func TestNewMinerScheduler(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		logger *log.Logger
	}{
		{
			name:   "valid parameters",
			config: &Config{PriceLimit: 50.0, Network: "192.168.1.0/24", MinerDiscoveryInterval: 10 * time.Minute},
			logger: log.New(os.Stdout, "TEST", log.LstdFlags),
		},
		{
			name:   "nil logger",
			config: &Config{PriceLimit: 75.5, Network: "10.0.0.0/16", MinerDiscoveryInterval: 10 * time.Minute},
			logger: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheduler := NewMinerScheduler(tt.config, tt.logger)

			if scheduler == nil {
				t.Fatal("NewMinerScheduler returned nil")
			}

			status := scheduler.GetStatus()

			if status.IsRunning {
				t.Error("New scheduler should not be running")
			}

			if tt.logger == nil && scheduler.logger == nil {
				t.Error("Expected default logger when nil provided")
			}
		})
	}
}

func TestDryRunConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		dryRun bool
	}{
		{
			name:   "dry-run enabled",
			dryRun: true,
		},
		{
			name:   "dry-run disabled",
			dryRun: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.DryRun = tt.dryRun
			config.SecurityToken = "test-token"

			scheduler := NewMinerScheduler(config, nil)

			if scheduler == nil {
				t.Fatal("NewMinerScheduler returned nil")
			}

			actualConfig := scheduler.GetConfig()
			if actualConfig.DryRun != tt.dryRun {
				t.Errorf("Expected DryRun to be %v, got %v", tt.dryRun, actualConfig.DryRun)
			}
		})
	}
}

func TestSchedulerRunningState(t *testing.T) {
	scheduler := NewMinerScheduler(&Config{
		PriceLimit:               50.0,
		Network:                  "192.168.1.0/24",
		CheckPriceInterval:       time.Minute,
		MinersStateCheckInterval: time.Minute,
		MinerDiscoveryInterval:   10 * time.Minute,
	}, nil)

	// Initially not running
	if scheduler.IsRunning() {
		t.Error("New scheduler should not be running")
	}

	// Test starting and stopping with context cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Start scheduler in goroutine
	done := make(chan error, 1)
	go func() {
		done <- scheduler.Start(ctx)
	}()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	if !scheduler.IsRunning() {
		t.Error("Scheduler should be running after Start()")
	}

	// Cancel context to stop
	cancel()

	// Wait for completion
	select {
	case err := <-done:
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled error, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Scheduler did not stop within timeout")
	}

	if scheduler.IsRunning() {
		t.Error("Scheduler should not be running after context cancellation")
	}
}

func TestSchedulerDoubleStart(t *testing.T) {
	scheduler := NewMinerScheduler(&Config{
		PriceLimit:               50.0,
		Network:                  "192.168.1.0/24",
		CheckPriceInterval:       time.Minute,
		MinersStateCheckInterval: time.Minute,
		MinerDiscoveryInterval:   10 * time.Minute,
	}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start first instance
	done1 := make(chan error, 1)
	go func() {
		done1 <- scheduler.Start(ctx)
	}()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Try to start second instance
	err := scheduler.Start(ctx)
	if err == nil {
		t.Error("Expected error when starting scheduler twice")
	}

	// Clean up
	cancel()
	<-done1
}

func TestSchedulerStop(t *testing.T) {
	scheduler := NewMinerScheduler(&Config{
		PriceLimit:               50.0,
		Network:                  "192.168.1.0/24",
		CheckPriceInterval:       time.Minute,
		MinersStateCheckInterval: time.Minute,
		MinerDiscoveryInterval:   10 * time.Minute,
	}, nil)
	ctx := context.Background()

	// Start scheduler
	done := make(chan error, 1)
	go func() {
		done <- scheduler.Start(ctx)
	}()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	if !scheduler.IsRunning() {
		t.Error("Scheduler should be running")
	}

	// Stop scheduler
	scheduler.Stop()

	// Wait for completion
	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Error("Scheduler did not stop within timeout")
	}

	if scheduler.IsRunning() {
		t.Error("Scheduler should not be running after Stop()")
	}
}

func TestGetDiscoveredMiners(t *testing.T) {
	scheduler := NewMinerScheduler(&Config{
		PriceLimit:               50.0,
		Network:                  "192.168.1.0/24",
		CheckPriceInterval:       time.Minute,
		MinersStateCheckInterval: time.Minute,
		MinerDiscoveryInterval:   10 * time.Minute,
	}, nil)

	// Initially empty
	minersList := scheduler.GetDiscoveredMiners()
	if len(minersList) != 0 {
		t.Errorf("Expected 0 miners initially, got %d", len(minersList))
	}

	// Mock some miners
	mockMiners := []*miners.AvalonQHost{
		{Address: "192.168.1.100", Port: 4028},
		{Address: "192.168.1.101", Port: 4028},
	}

	scheduler.mu.Lock()
	scheduler.discoveredMiners = make(map[string]*miners.AvalonQHost)
	for _, miner := range mockMiners {
		key := fmt.Sprintf("%s:%d", miner.Address, miner.Port)
		scheduler.discoveredMiners[key] = miner
	}
	scheduler.mu.Unlock()

	// Test getting miners
	retrievedMiners := scheduler.GetDiscoveredMiners()
	if len(retrievedMiners) != len(mockMiners) {
		t.Errorf("Expected %d miners, got %d", len(mockMiners), len(retrievedMiners))
	}

	// Test that returned slice is a copy (modifying it shouldn't affect original)
	if len(retrievedMiners) > 0 {
		retrievedMiners[0] = nil
		originalMiners := scheduler.GetDiscoveredMiners()
		if originalMiners[0] == nil {
			t.Error("Modifying returned slice should not affect original")
		}
	}
}

func TestDiscoverMinersPreservesExisting(t *testing.T) {
	scheduler := NewMinerScheduler(&Config{
		PriceLimit:               50.0,
		Network:                  "127.0.0.1/32",
		CheckPriceInterval:       time.Minute,
		MinersStateCheckInterval: time.Minute,
		MinerDiscoveryInterval:   10 * time.Minute,
	}, nil)

	// Mock some initial miners
	initialMiners := []*miners.AvalonQHost{
		{Address: "192.168.1.100", Port: 4028},
		{Address: "192.168.1.101", Port: 4028},
	}

	scheduler.mu.Lock()
	scheduler.discoveredMiners = make(map[string]*miners.AvalonQHost)
	for _, miner := range initialMiners {
		key := fmt.Sprintf("%s:%d", miner.Address, miner.Port)
		scheduler.discoveredMiners[key] = miner
	}
	scheduler.mu.Unlock()

	// Run discovery (should not find anything on 127.0.0.1/32, but existing miners should be preserved)
	err := scheduler.discoverMiners(context.Background())
	if err != nil {
		t.Errorf("Discovery failed: %v", err)
	}

	// Check that existing miners are still there
	finalMiners := scheduler.GetDiscoveredMiners()
	if len(finalMiners) != len(initialMiners) {
		t.Errorf("Expected %d miners after discovery, got %d", len(initialMiners), len(finalMiners))
	}

	// Check that the exact same miners are still there
	for i, expectedMiner := range initialMiners {
		if finalMiners[i].Address != expectedMiner.Address || finalMiners[i].Port != expectedMiner.Port {
			t.Errorf("Expected miner %d to be %s:%d, got %s:%d",
				i, expectedMiner.Address, expectedMiner.Port,
				finalMiners[i].Address, finalMiners[i].Port)
		}
	}
}

func TestGetStatus(t *testing.T) {

	scheduler := NewMinerScheduler(&Config{
		PriceLimit:               50.0,
		Network:                  "192.168.1.0/24",
		CheckPriceInterval:       time.Minute,
		MinersStateCheckInterval: time.Minute,
		MinerDiscoveryInterval:   10 * time.Minute,
	}, nil)

	status := scheduler.GetStatus()

	if status.IsRunning {
		t.Error("Expected running status false")
	}

	if status.MinersCount != 0 {
		t.Errorf("Expected miners count 0, got %d", status.MinersCount)
	}

	if status.HasMarketData {
		t.Error("Expected has no market data")
	}
}

func TestSchedulerStatus_WithData(t *testing.T) {
	scheduler := NewMinerScheduler(&Config{
		PriceLimit:               50.0,
		Network:                  "192.168.1.0/24",
		CheckPriceInterval:       time.Minute,
		MinersStateCheckInterval: time.Minute,
		MinerDiscoveryInterval:   10 * time.Minute,
	}, nil)

	// Add some mock data
	mockMiners := []*miners.AvalonQHost{
		{Address: "192.168.1.100", Port: 4028},
		{Address: "192.168.1.101", Port: 4028},
	}

	mockDoc := &entsoe.PublicationMarketData{
		MRID:            "test-doc",
		CreatedDateTime: "2024-01-15T10:30:00Z",
	}

	scheduler.mu.Lock()
	scheduler.discoveredMiners = make(map[string]*miners.AvalonQHost)
	for _, miner := range mockMiners {
		key := fmt.Sprintf("%s:%d", miner.Address, miner.Port)
		scheduler.discoveredMiners[key] = miner
	}
	scheduler.pricesMarketData = mockDoc
	scheduler.mu.Unlock()

	status := scheduler.GetStatus()

	if status.MinersCount != len(mockMiners) {
		t.Errorf("Expected miners count %d, got %d", len(mockMiners), status.MinersCount)
	}

	if !status.HasMarketData {
		t.Error("Expected has market data")
	}
}

func TestSchedulerConcurrency(t *testing.T) {
	scheduler := NewMinerScheduler(&Config{
		PriceLimit:               50.0,
		Network:                  "192.168.1.0/24",
		CheckPriceInterval:       time.Minute,
		MinersStateCheckInterval: time.Minute,
		MinerDiscoveryInterval:   10 * time.Minute,
	}, nil)

	// Test concurrent access to methods
	done := make(chan bool, 10)

	// Concurrent readers
	for range 5 {
		go func() {
			defer func() { done <- true }()
			for range 100 {
				_ = scheduler.GetDiscoveredMiners()
				_ = scheduler.GetStatus()
				_ = scheduler.IsRunning()
			}
		}()
	}

	// Concurrent writers
	for i := range 5 {
		go func(id int) {
			defer func() { done <- true }()
			for range 100 {
				scheduler.SetConfig(&Config{PriceLimit: 50.0 + float64(id), Network: "192.168." + string(rune('1'+id)) + ".0/24", MinerDiscoveryInterval: 10 * time.Minute})
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for range 10 {
		select {
		case <-done:
			// OK
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent test timed out")
		}
	}
}

// Benchmark tests
func BenchmarkSchedulerGetStatus(b *testing.B) {
	scheduler := NewMinerScheduler(&Config{
		PriceLimit:               50.0,
		Network:                  "192.168.1.0/24",
		CheckPriceInterval:       time.Minute,
		MinersStateCheckInterval: time.Minute,
		MinerDiscoveryInterval:   10 * time.Minute,
	}, nil)

	// Add some mock data
	mockMiners := make([]*miners.AvalonQHost, 100)
	for i := range 100 {
		mockMiners[i] = &miners.AvalonQHost{
			Address: "192.168.1." + string(rune(100+i)),
			Port:    4028,
		}
	}
	scheduler.discoveredMiners = make(map[string]*miners.AvalonQHost)
	for _, miner := range mockMiners {
		key := fmt.Sprintf("%s:%d", miner.Address, miner.Port)
		scheduler.discoveredMiners[key] = miner
	}

	for b.Loop() {
		_ = scheduler.GetStatus()
	}
}

func TestGetInitialDelay(t *testing.T) {
	tests := []struct {
		name               string
		priceCheckInterval time.Duration
		stateCheckInterval time.Duration
		now                time.Time
		expectedDelay      time.Duration
	}{
		{
			name:               "at start of hour with 15min interval",
			priceCheckInterval: 15 * time.Minute,
			stateCheckInterval: 1 * time.Minute,
			now:                time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			expectedDelay:      0 * time.Minute,
		},
		{
			name:               "5 minutes into hour with 15min interval",
			priceCheckInterval: 15 * time.Minute,
			stateCheckInterval: 1 * time.Minute,
			now:                time.Date(2024, 1, 15, 10, 5, 0, 0, time.UTC),
			expectedDelay:      10 * time.Minute,
		},
		{
			name:               "exactly at 15min mark with 15min interval",
			priceCheckInterval: 15 * time.Minute,
			stateCheckInterval: 1 * time.Minute,
			now:                time.Date(2024, 1, 15, 10, 15, 0, 0, time.UTC),
			expectedDelay:      0 * time.Minute,
		},
		{
			name:               "17 minutes into hour with 15min interval",
			priceCheckInterval: 15 * time.Minute,
			now:                time.Date(2024, 1, 15, 10, 17, 0, 0, time.UTC),
			expectedDelay:      13 * time.Minute,
		},
		{
			name:               "at 30min mark with 15min interval",
			priceCheckInterval: 15 * time.Minute,
			stateCheckInterval: 1 * time.Minute,
			now:                time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			expectedDelay:      0 * time.Minute,
		},
		{
			name:               "45 minutes into hour with 15min interval",
			priceCheckInterval: 15 * time.Minute,
			stateCheckInterval: 1 * time.Minute,
			now:                time.Date(2024, 1, 15, 10, 45, 0, 0, time.UTC),
			expectedDelay:      0 * time.Minute,
		},
		{
			name:               "50 minutes into hour with 15min interval",
			priceCheckInterval: 15 * time.Minute,
			stateCheckInterval: 1 * time.Minute,
			now:                time.Date(2024, 1, 15, 10, 50, 0, 0, time.UTC),
			expectedDelay:      10 * time.Minute,
		},
		{
			name:               "at start of hour with 30min interval",
			priceCheckInterval: 30 * time.Minute,
			stateCheckInterval: 1 * time.Minute,
			now:                time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			expectedDelay:      0 * time.Minute,
		},
		{
			name:               "10 minutes into hour with 30min interval",
			priceCheckInterval: 30 * time.Minute,
			stateCheckInterval: 1 * time.Minute,
			now:                time.Date(2024, 1, 15, 10, 10, 0, 0, time.UTC),
			expectedDelay:      20 * time.Minute,
		},
		{
			name:               "exactly at 30min mark with 30min interval",
			priceCheckInterval: 30 * time.Minute,
			now:                time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			expectedDelay:      0 * time.Minute,
		},
		{
			name:               "40 minutes into hour with 30min interval",
			priceCheckInterval: 30 * time.Minute,
			stateCheckInterval: 1 * time.Minute,
			now:                time.Date(2024, 1, 15, 10, 40, 0, 0, time.UTC),
			expectedDelay:      20 * time.Minute,
		},
		{
			name:               "at start of hour with 1hour interval",
			priceCheckInterval: 60 * time.Minute,
			stateCheckInterval: 1 * time.Minute,
			now:                time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			expectedDelay:      0 * time.Minute,
		},
		{
			name:               "30 minutes into hour with 1hour interval",
			priceCheckInterval: 60 * time.Minute,
			stateCheckInterval: 1 * time.Minute,
			now:                time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			expectedDelay:      30 * time.Minute,
		},
		{
			name:               "with seconds precision",
			priceCheckInterval: 15 * time.Minute,
			stateCheckInterval: 1 * time.Minute,
			now:                time.Date(2024, 1, 15, 10, 5, 30, 0, time.UTC),
			expectedDelay:      9*time.Minute + 30*time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				PriceLimit:               50.0,
				Network:                  "192.168.1.0/24",
				CheckPriceInterval:       tt.priceCheckInterval,
				MinersStateCheckInterval: tt.stateCheckInterval,
				MinerDiscoveryInterval:   10 * time.Minute,
			}
			scheduler := NewMinerScheduler(config, nil)

			actualDelay := scheduler.getInitialDelay(tt.now, tt.priceCheckInterval)

			if actualDelay != tt.expectedDelay {
				t.Errorf("Expected delay %v, got %v", tt.expectedDelay, actualDelay)
			}

			// Verify that the delay is always positive
			if actualDelay < 0 {
				t.Errorf("Expected positive delay, got %v", actualDelay)
			}

			// Verify that the delay is less than or equal to the check interval
			if actualDelay > tt.priceCheckInterval {
				t.Errorf("Expected delay (%v) to be less than or equal to check interval (%v)", actualDelay, tt.priceCheckInterval)
			}
		})
	}
}

func BenchmarkSchedulerGetDiscoveredMiners(b *testing.B) {
	scheduler := NewMinerScheduler(&Config{PriceLimit: 50.0, Network: "192.168.1.0/24", CheckPriceInterval: time.Minute, MinerDiscoveryInterval: 10 * time.Minute}, nil)

	// Add mock miners
	mockMiners := make([]*miners.AvalonQHost, 1000)
	for i := range 1000 {
		mockMiners[i] = &miners.AvalonQHost{
			Address: "192.168." + string(rune(i/256)) + "." + string(rune(i%256)),
			Port:    4028,
		}
	}
	scheduler.discoveredMiners = make(map[string]*miners.AvalonQHost)
	for _, miner := range mockMiners {
		key := fmt.Sprintf("%s:%d", miner.Address, miner.Port)
		scheduler.discoveredMiners[key] = miner
	}

	for b.Loop() {
		_ = scheduler.GetDiscoveredMiners()
	}
}

func TestMinersStateCheckInterval(t *testing.T) {
	config := DefaultConfig()
	config.MinersStateCheckInterval = 30 * time.Second

	scheduler := NewMinerScheduler(config, nil)

	if scheduler.config.MinersStateCheckInterval != 30*time.Second {
		t.Errorf("Expected MinersStateCheckInterval 30s, got %v", scheduler.config.MinersStateCheckInterval)
	}

	// Test that default is 1 minute
	defaultConfig := DefaultConfig()
	if defaultConfig.MinersStateCheckInterval != 1*time.Minute {
		t.Errorf("Expected default MinersStateCheckInterval 1m, got %v", defaultConfig.MinersStateCheckInterval)
	}
}

func TestRunStateCheckDryRun(t *testing.T) {
	config := DefaultConfig()
	config.DryRun = true
	config.MinersStateCheckInterval = 10 * time.Second

	scheduler := NewMinerScheduler(config, nil)

	// Test that runStateCheck doesn't panic with no miners
	scheduler.runStateCheck(context.Background())

	// Verify the method exists and can be called
	if scheduler.config.DryRun != true {
		t.Error("Expected DryRun to be true")
	}
}
