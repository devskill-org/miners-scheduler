package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/devskill-org/ems/scheduler"
	"github.com/devskill-org/ems/sigenergy"
)

func main() {
	// Command line flags
	var (
		configFile = flag.String("config", "config.json", "Configuration file path")
		info       = flag.Bool("info", false, "Show Plant Information")
		help       = flag.Bool("help", false, "Show help message")
		serverOnly = flag.Bool("serverOnly", false, "Run only web server without periodic checks")
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
	fmt.Printf("Starting Energy Management System with the following configuration:\n")
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
		if err := minerScheduler.Start(ctx, *serverOnly); err != nil {
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
	fmt.Println("Energy Management System (EMS) - Optimize energy consumption, production, and storage")
	fmt.Println()
	fmt.Println("DESCRIPTION:")
	fmt.Println("  A comprehensive energy management system that integrates solar (PV), battery storage,")
	fmt.Println("  grid connection, and controllable loads. The system uses real-time electricity prices,")
	fmt.Println("  weather forecasts, and Model Predictive Control (MPC) to minimize energy costs.")
	fmt.Println()
	fmt.Println("  Key Features:")
	fmt.Println("  - Solar power monitoring via Modbus")
	fmt.Println("  - Battery charge/discharge optimization")
	fmt.Println("  - Price-based load management")
	fmt.Println("  - Weather-integrated solar forecasting")
	fmt.Println("  - Real-time web dashboard")
	fmt.Println("  - Thermal protection for devices")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  ems [OPTIONS]")
	fmt.Println()
	fmt.Println("OPTIONS:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Basic usage with default settings")
	fmt.Println("  ems")
	fmt.Println()
	fmt.Println("  # Custom configuration")
	fmt.Println("  ems --config=config.json")
	fmt.Println()
	fmt.Println("  # Show plant/system information")
	fmt.Println("  ems -info")
	fmt.Println()
	fmt.Println("  # Run only web server without periodic checks")
	fmt.Println("  ems -serverOnly")
	fmt.Println()
	fmt.Println("  # Show this help")
	fmt.Println("  ems -help")
}
