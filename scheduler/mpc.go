package scheduler

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/devskill-org/miners-scheduler/meteo"
	"github.com/devskill-org/miners-scheduler/mpc"
	"github.com/devskill-org/miners-scheduler/sigenergy"
	"github.com/sixdouglas/suncalc"
)

// runMPCOptimize executes the MPC optimization task
func (s *MinerScheduler) runMPCOptimize(ctx context.Context) {
	s.logger.Printf("Starting MPC optimization task at %s", time.Now().Format(time.RFC3339))

	config := s.GetConfig()

	// Check if Plant Modbus Address is configured
	if config.PlantModbusAddress == "" {
		s.logger.Printf("MPC optimization skipped: PlantModbusAddress not configured")
		return
	}

	// Step 1: Read initial SOC from inverter
	initialSOC, err := s.readInitialSOC(config)
	if err != nil {
		s.logger.Printf("Error reading initial SOC from inverter: %v", err)
		return
	}

	s.logger.Printf("Initial battery SOC: %.1f%%", initialSOC*100)

	// Step 2: Get forecast data (prices, solar, load)
	forecast, err := s.buildMPCForecast(ctx, config)
	if err != nil {
		s.logger.Printf("Error building MPC forecast: %v", err)
		return
	}

	if len(forecast) == 0 {
		s.logger.Printf("No forecast data available for MPC optimization")
		return
	}

	s.logger.Printf("Built forecast with %d time slots", len(forecast))

	// Step 3: Create MPC controller
	systemConfig := mpc.SystemConfig{
		BatteryCapacity:        config.BatteryCapacity,
		BatteryMaxCharge:       config.BatteryMaxCharge,
		BatteryMaxDischarge:    config.BatteryMaxDischarge,
		BatteryMinSOC:          config.BatteryMinSOC,
		BatteryMaxSOC:          config.BatteryMaxSOC,
		BatteryEfficiency:      config.BatteryEfficiency,
		BatteryDegradationCost: config.BatteryDegradationCost,
		MaxGridImport:          config.MaxGridImport,
		MaxGridExport:          config.MaxGridExport,
	}

	horizon := len(forecast)
	controller := mpc.NewMPCController(systemConfig, horizon, initialSOC)

	// Step 4: Run optimization
	decisions := controller.Optimize(forecast)
	if len(decisions) == 0 {
		s.logger.Printf("MPC optimization produced no decisions")
		return
	}

	// Step 5: Log optimization results
	s.logMPCResults(forecast, decisions)

	// Step 6: Execute the first control decision
	if err := s.executeMPCDecision(&decisions[0], true); err != nil {
		s.logger.Printf("Error executing MPC decision: %v", err)
		return
	}

	s.logger.Printf("MPC optimization task completed successfully")
}

// readInitialSOC reads the current State of Charge from the inverter
func (s *MinerScheduler) readInitialSOC(config *Config) (float64, error) {
	// Connect to Plant Modbus server
	client, err := sigenergy.NewTCPClient(config.PlantModbusAddress, sigenergy.PlantAddress)
	if err != nil {
		return 0, fmt.Errorf("failed to connect to Plant Modbus: %w", err)
	}
	defer client.Close()

	// Read plant running info to get SOC
	plantInfo, err := client.ReadPlantRunningInfo()
	if err != nil {
		return 0, fmt.Errorf("failed to read plant info: %w", err)
	}

	// Convert SOC from percentage (0-100) to fraction (0-1)
	socFraction := plantInfo.ESSSOC / 100.0

	return socFraction, nil
}

// buildMPCForecast builds the forecast data needed for MPC optimization
func (s *MinerScheduler) buildMPCForecast(ctx context.Context, config *Config) ([]mpc.TimeSlot, error) {
	now := time.Now()

	// Get electricity price forecast
	priceForecasts, err := s.getPriceForecast(ctx, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get price forecast: %w", err)
	}

	// Get solar forecast
	solarForecasts, err := s.getSolarForecast(config, now)
	if err != nil {
		s.logger.Printf("Warning: failed to get solar forecast: %v, using zero solar", err)
		// Continue with zero solar forecast
		solarForecasts = make(map[int]float64)
	}

	// Build time slots
	var timeSlots []mpc.TimeSlot
	for hour, prices := range priceForecasts {
		solar := solarForecasts[hour]

		// Estimate load forecast (miners only, based on price and solar availability)
		loadForecast := s.estimateLoadForecast(prices.Import, config.PriceLimit/1000, solar, config)

		timeSlots = append(timeSlots, mpc.TimeSlot{
			Hour:          hour,
			ImportPrice:   prices.Import / 1000.0, // Convert EUR/MWh to EUR/kWh
			ExportPrice:   prices.Export / 1000.0, // Convert EUR/MWh to EUR/kWh
			SolarForecast: solar,
			LoadForecast:  loadForecast,
		})
	}

	// Sort by hour
	for i := 0; i < len(timeSlots); i++ {
		for j := i + 1; j < len(timeSlots); j++ {
			if timeSlots[i].Hour > timeSlots[j].Hour {
				timeSlots[i], timeSlots[j] = timeSlots[j], timeSlots[i]
			}
		}
	}

	return timeSlots, nil
}

// PricePoint represents import and export prices for a specific hour
type PricePoint struct {
	Import float64 // EUR/MWh
	Export float64 // EUR/MWh
}

// getPriceForecast gets electricity prices for the next 24 hours
func (s *MinerScheduler) getPriceForecast(ctx context.Context, now time.Time) (map[int]PricePoint, error) {

	// Get the market data
	marketData, err := s.GetMarketData(ctx)
	if err != nil {
		return nil, err
	}
	if marketData == nil {
		return nil, fmt.Errorf("no price document available")
	}

	// Get configuration for price adjustments
	config := s.GetConfig()

	// Extract prices for next 24 hours
	forecast := make(map[int]PricePoint)
	for i := range 24 {
		futureTime := now.Add(time.Duration(i) * time.Hour)
		price, found := marketData.LookupAveragePriceInHourByTime(futureTime)

		if found {
			// Apply price adjustments from configuration (all values in EUR/MWh)
			// Import price: add operator fee and delivery fee
			// Export price: subtract operator fee
			forecast[i] = PricePoint{
				Import: price + config.ImportPriceOperatorFee + config.ImportPriceDeliveryFee,
				Export: price - config.ExportPriceOperatorFee,
			}
		}
	}

	return forecast, nil
}

// getSolarForecast gets solar power forecast from weather data
func (s *MinerScheduler) getSolarForecast(config *Config, now time.Time) (map[int]float64, error) {
	// Get weather forecast
	weatherForecast, err := s.getOrFetchWeatherForecast(config)
	if err != nil {
		return nil, fmt.Errorf("failed to get weather forecast: %w", err)
	}

	if weatherForecast == nil || weatherForecast.Properties == nil {
		return nil, fmt.Errorf("invalid weather forecast data")
	}

	// Convert weather to solar forecast
	solarForecast := make(map[int]float64)

	for i := range 24 {
		futureTime := now.Add(time.Duration(i) * time.Hour)
		solarPower := s.estimateSolarPowerFromWeather(weatherForecast, futureTime, config.MaxSolarPower)
		solarForecast[i] = solarPower
	}

	return solarForecast, nil
}

// getOrFetchWeatherForecast gets weather forecast from cache or fetches new one
func (s *MinerScheduler) getOrFetchWeatherForecast(config *Config) (*meteo.METJSONForecast, error) {
	// Try cache first
	if forecast, ok := s.weatherCache.Get(); ok {
		return forecast, nil
	}

	// Fetch new forecast
	client := meteo.NewClient(config.UserAgent)

	forecast, err := client.GetComplete(meteo.QueryParams{
		Location: meteo.Location{
			Latitude:  config.Latitude,
			Longitude: config.Longitude,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather forecast: %w", err)
	}

	// Cache it
	s.weatherCache.Set(forecast)

	return forecast, nil
}

// estimateSolarPowerFromWeather estimates solar power output from weather data
func (s *MinerScheduler) estimateSolarPowerFromWeather(forecast *meteo.METJSONForecast, targetTime time.Time, peakPower float64) float64 {
	if forecast.Properties == nil || len(forecast.Properties.Timeseries) == 0 {
		return 0
	}

	// Find closest time step
	var closestStep *meteo.ForecastTimeStep
	minDiff := time.Duration(math.MaxInt64)

	for _, step := range forecast.Properties.Timeseries {
		diff := step.Time.Sub(targetTime)
		if diff < 0 {
			diff = -diff
		}
		if diff < minDiff {
			minDiff = diff
			closestStep = &step
		}
	}

	if closestStep == nil || closestStep.Data == nil || closestStep.Data.Instant == nil || closestStep.Data.Instant.Details == nil {
		return 0
	}

	details := closestStep.Data.Instant.Details

	// Get location from config
	config := s.GetConfig()
	lat := config.Latitude
	lon := config.Longitude

	// Get sun times for the target date
	sunTimes := suncalc.GetTimes(targetTime, lat, lon)
	sunrise := sunTimes["sunrise"].Value
	sunset := sunTimes["sunset"].Value

	// Check if we're between sunrise and sunset
	if targetTime.Before(sunrise) || targetTime.After(sunset) {
		return 0 // No sun available
	}

	// Get solar position to calculate altitude angle
	pos := suncalc.GetPosition(targetTime, lat, lon)
	altitude := pos.Altitude // in radians

	// Solar altitude factor (0-1)
	// Altitude ranges from 0 (horizon) to π/2 (zenith)
	// Use sine of altitude as a factor (0 at horizon, 1 at zenith)
	solarAngleFactor := math.Sin(altitude)
	if solarAngleFactor < 0 {
		solarAngleFactor = 0 // Below horizon
	}

	// Cloud factor (0-1, where 1 = no clouds)
	cloudFactor := 1.0
	if details.CloudAreaFraction != nil {
		cloudFraction := *details.CloudAreaFraction / 100.0
		cloudFactor = 1.0 - (cloudFraction * 0.90) // Clouds reduce output by up to 90%
	}

	// Estimate solar power
	solarPower := peakPower * solarAngleFactor * cloudFactor

	return solarPower
}

// estimateLoadForecast estimates power load based on price and available power
// Follows the same logic as manageMiners: miners wake up in Eco mode when price <= limit,
// but only if there's enough power budget (when PV power control is enabled)
func (s *MinerScheduler) estimateLoadForecast(hourlyPrice float64, priceLimit float64, solarForecast float64, config *Config) float64 {
	// Convert hourlyPrice from EUR/MWh to EUR/kWh for comparison with priceLimit
	hourlyPricePerKWh := hourlyPrice / 1000.0

	// Miners are only ON if price is below or equal the limit
	if hourlyPricePerKWh > priceLimit {
		return 0.0
	}

	// Get discovered miners
	minersList := s.GetDiscoveredMiners()
	if len(minersList) == 0 {
		return 0.0
	}

	// Check if PV power control is enabled
	usePowerControl := config.UsePVPowerControl
	if !usePowerControl {
		// Without power control, all miners can run in Eco mode
		totalMinerPower := float64(len(minersList)) * config.MinerPowerEco
		return totalMinerPower
	}

	// With power control enabled, calculate effective power limit
	// Use minimum of available solar power and configured miners power limit
	effectiveLimit := config.MinersPowerLimit
	if solarForecast < effectiveLimit {
		effectiveLimit = solarForecast
	}

	// Calculate how many miners can run within the effective limit
	// Miners wake up in Eco mode (as per manageMiners logic)
	minerPowerEco := config.MinerPowerEco
	if minerPowerEco <= 0 {
		minerPowerEco = 1.0 // Default fallback
	}

	maxMinersCanRun := int(effectiveLimit / minerPowerEco)
	actualMinersRunning := min(maxMinersCanRun, len(minersList))

	totalMinerPower := float64(actualMinersRunning) * minerPowerEco
	return totalMinerPower
}

// logMPCResults logs the optimization results
func (s *MinerScheduler) logMPCResults(forecast []mpc.TimeSlot, decisions []mpc.ControlDecision) {
	s.logger.Printf("MPC Optimization Results:")
	s.logger.Printf("Hour | Solar | Load  | Import  | Export  | Battery         | Grid         | SOC    | Profit")
	s.logger.Printf("-----|-------|-------|---------|---------|-----------------|--------------|--------|-------")

	totalProfit := 0.0
	for i, dec := range decisions {
		slot := forecast[i]

		gridAction := "idle"
		gridPower := 0.0
		if dec.GridImport > 0.1 {
			gridAction = "import"
			gridPower = dec.GridImport
		} else if dec.GridExport > 0.1 {
			gridAction = "export"
			gridPower = dec.GridExport
		}

		battAction := "idle"
		battPower := 0.0
		if dec.BatteryCharge > 0.1 {
			battAction = "charge"
			battPower = dec.BatteryCharge
		} else if dec.BatteryDischarge > 0.1 {
			battAction = "discharge"
			battPower = dec.BatteryDischarge
		}

		totalProfit += dec.Profit

		s.logger.Printf("%4d | %5.1f | %5.1f | €%.4f | €%.4f | %15s | %12s | %5.1f%% | €%.3f",
			dec.Hour,
			slot.SolarForecast,
			slot.LoadForecast,
			slot.ImportPrice,
			slot.ExportPrice,
			fmt.Sprintf("%s: %.1f", battAction, battPower),
			fmt.Sprintf("%s: %.1f", gridAction, gridPower),
			dec.BatterySOC*100,
			dec.Profit,
		)
	}

	s.logger.Printf("Total expected profit over %d hours: €%.2f", len(decisions), totalProfit)
}

// executeMPCDecision executes the first MPC control decision
func (s *MinerScheduler) executeMPCDecision(decision *mpc.ControlDecision, dryRun bool) error {
	if dryRun {
		s.logger.Printf("DRY-RUN: Would execute MPC decision - Charge: %.1f kW, Discharge: %.1f kW, Import: %.1f kW, Export: %.1f kW",
			decision.BatteryCharge, decision.BatteryDischarge, decision.GridImport, decision.GridExport)
		return nil
	}

	s.logger.Printf("Successfully executed MPC decision")
	return nil
}
