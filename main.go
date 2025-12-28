package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/devskill-org/miners-scheduler/scheduler"
	"github.com/devskill-org/miners-scheduler/sigenergy"
)

func main() {
	// Command line flags
	var (
		configFile = flag.String("config", "config.json", "Configuration file path")
		info       = flag.Bool("info", false, "Show Plant Information")
		help       = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	config, err := scheduler.LoadConfig(*configFile)
	if err != nil {
		fmt.Println("Error loading configuration:", err)
		return
	}

	if *info {
		if err := sigenergy.ShowPlantInfo(config.PlantModbusAddress); err != nil {
			fmt.Println("Error:", err)
			return
		}
		return
	}
	fmt.Printf("Starting miners-scheduler with the following configuration:\n")
	fmt.Printf("  Price Limit: %.2f EUR/MWh\n", config.PriceLimit)
	fmt.Printf("  Network: %s\n", config.Network)
	fmt.Printf("  Check Price Interval: %s\n", config.CheckPriceInterval)
	fmt.Printf("  FanR High Threshold: %d\n", config.FanRHighThreshold)
	fmt.Printf("  FanR Low Threshold: %d\n", config.FanRLowThreshold)

	if config.DryRun {
		fmt.Printf("  Mode: DRY-RUN (actions will be simulated only)\n")
	}
	fmt.Println()

	// Create logger
	logger := log.New(os.Stdout, "[SCHEDULER] ", log.LstdFlags)

	// Create scheduler
	minerScheduler := scheduler.NewMinerSchedulerWithHealthCheck(config, logger)

	// Set up context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start scheduler in a goroutine
	go func() {
		if err := minerScheduler.Start(ctx); err != nil {
			if err != context.Canceled {
				logger.Printf("Scheduler error: %v", err)
			}
		}
	}()

	logger.Printf("Scheduler started. Press Ctrl+C to stop...")

	// Wait for shutdown signal
	<-sigChan
	logger.Printf("Shutdown signal received, stopping scheduler...")

	// Cancel context to stop scheduler
	cancel()

	// Give the scheduler a moment to clean up
	minerScheduler.Stop()

	logger.Printf("Scheduler stopped successfully")
}

func showHelp() {
	fmt.Println("Miners Scheduler - Manages Avalon miners based on electricity prices")
	fmt.Println()
	fmt.Println("DESCRIPTION:")
	fmt.Println("  This application periodically (every 15 minutes) performs the following tasks:")
	fmt.Println("  1. Discovers Avalon miners on the specified network")
	fmt.Println("  2. Gets current electricity price from ENTSO-E API")
	fmt.Println("  3. Compares price with the configured limit")
	fmt.Println("  4. Manages miner states:")
	fmt.Println("     - If price ≤ limit: Wake up miners")
	fmt.Println("     - If price > limit: Put active miners into standby")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  miners-scheduler [OPTIONS]")
	fmt.Println()
	fmt.Println("OPTIONS:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Basic usage with default settings")
	fmt.Println("  miners-scheduler")
	fmt.Println()
	fmt.Println("  # Custom settings")
	fmt.Println("  miners-scheduler --config=config.json")
	fmt.Println()
	fmt.Println("  # Show this help")
	fmt.Println("  miners-scheduler -help")
}
