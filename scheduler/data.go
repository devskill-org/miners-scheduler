package scheduler

import (
	"database/sql"
	"sync"
	"time"

	"github.com/devskill-org/ems/meteo"
	"github.com/devskill-org/ems/sigenergy"
)

type WeatherForecastCache struct {
	mu            sync.RWMutex
	forecast      *meteo.METJSONForecast
	fetchedAt     time.Time
	cacheDuration time.Duration
}

func (w *WeatherForecastCache) Get() (*meteo.METJSONForecast, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.forecast == nil {
		return nil, false
	}

	if time.Since(w.fetchedAt) > w.cacheDuration {
		return nil, false
	}

	return w.forecast, true
}

func (w *WeatherForecastCache) Set(forecast *meteo.METJSONForecast) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.forecast = forecast
	w.fetchedAt = time.Now()
}

type DataSample struct {
	pvPower      float64
	gridPower    float64 // positive = import, negative = export
	batteryPower float64 // positive = charging, negative = discharging
	evdcPower    float64
	batterySoc   float64 // %
	ts           time.Time
}

type DataSamples struct {
	mu      sync.Mutex
	samples []DataSample
}

func (d *DataSamples) AddSample(pvPower, gridPower, batteryPower, evdcPower, batterySoc float64, ts time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.samples = append(d.samples, DataSample{
		pvPower:      pvPower,
		gridPower:    gridPower,
		batteryPower: batteryPower,
		evdcPower:    evdcPower,
		batterySoc:   batterySoc,
		ts:           ts,
	})
}

type IntegratedData struct {
	pvTotalPower          float64
	gridExportPower       float64
	gridImportPower       float64
	batteryChargePower    float64
	batteryDischargePower float64
	evdcChargePower       float64
	loadPower             float64
	batterySoc            float64
}

func (d *DataSamples) IntegrateSamples(pollInterval time.Duration) IntegratedData {
	d.mu.Lock()
	defer d.mu.Unlock()

	var result IntegratedData

	for _, sample := range d.samples {
		energyKWh := pollInterval.Seconds() / 3600.0 // Convert to hours

		result.pvTotalPower += sample.pvPower * energyKWh

		// Grid power: positive = import, negative = export
		if sample.gridPower > 0 {
			result.gridImportPower += sample.gridPower * energyKWh
		} else if sample.gridPower < 0 {
			result.gridExportPower += -sample.gridPower * energyKWh
		}

		// Battery power: positive = charging, negative = discharging
		if sample.batteryPower > 0 {
			result.batteryChargePower += sample.batteryPower * energyKWh
		} else if sample.batteryPower < 0 {
			result.batteryDischargePower += -sample.batteryPower * energyKWh
		}

		// EV DC charging power
		result.evdcChargePower += sample.evdcPower * energyKWh

		// Keep the last battery SOC
		result.batterySoc = sample.batterySoc
	}

	// Calculate load: Load = PV + Battery Discharge + Grid Import - Battery Charge - Grid Export - EV Charge
	result.loadPower = result.pvTotalPower + result.batteryDischargePower + result.gridImportPower -
		result.batteryChargePower - result.gridExportPower - result.evdcChargePower

	d.samples = d.samples[:0]
	return result
}

func (d *DataSamples) IsEmpty() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.samples) == 0
}

// GetLatestPower returns the most recent PV power sample, or 0 if no samples exist
func (d *DataSamples) GetLatestPower() float64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.samples) == 0 {
		return 0
	}
	return d.samples[len(d.samples)-1].pvPower
}

func (s *MinerScheduler) runDataPoll(samples *DataSamples) {
	if s.config.PlantModbusAddress == "" {
		return
	}
	client, err := sigenergy.NewTCPClient(s.config.PlantModbusAddress, sigenergy.PlantAddress)
	if err != nil {
		s.logger.Printf("Data integration: failed to create modbus client: %v", err)
		return
	}
	defer client.Close()
	info, err := client.ReadPlantRunningInfo()
	if err != nil {
		s.logger.Printf("Data integration: failed to read PlantRunningInfo: %v", err)
		return
	}
	samples.AddSample(
		info.PhotovoltaicPower,
		info.GridSensorActivePower,
		info.ESSPower,
		info.DCChargerOutputPower,
		info.ESSSOC,
		time.Now(),
	)
}

func (s *MinerScheduler) runDataIntegration(samples *DataSamples, pollInterval time.Duration, dataDB *sql.DB, deviceID int, dryRun bool) {
	if samples.IsEmpty() {
		s.logger.Printf("Data integration: no samples collected in period")
		return
	}

	data := samples.IntegrateSamples(pollInterval)
	timestamp := time.Now()

	if dataDB == nil {
		return
	}

	// Fetch weather data from meteo API
	cloudCoverage, err := s.fetchCloudCoverage()
	if err != nil {
		s.logger.Printf("Data integration: failed to fetch cloud coverage: %v", err)
	}

	weatherSymbol, err := s.fetchWeatherSymbol()
	if err != nil {
		s.logger.Printf("Data integration: failed to fetch weather symbol: %v", err)
	}

	// Calculate costs using current energy prices
	config := s.GetConfig()

	// Get current spot price for cost calculations
	var gridImportCost, gridExportCost float64
	s.mu.RLock()
	marketData := s.pricesMarketData
	s.mu.RUnlock()

	if marketData != nil {
		spotPrice, found := marketData.LookupAveragePriceInHourByTime(timestamp)
		if found && spotPrice > 0 {
			// Import cost: (spot price + operator fee + delivery fee) * energy in MWh
			importPricePerMWh := spotPrice + config.ImportPriceOperatorFee + config.ImportPriceDeliveryFee
			gridImportCost = (importPricePerMWh / 1000.0) * data.gridImportPower // Convert to EUR

			// Export revenue (negative cost): (spot price - operator fee) * energy in MWh
			exportPricePerMWh := spotPrice - config.ExportPriceOperatorFee
			gridExportCost = (exportPricePerMWh / 1000.0) * data.gridExportPower // Convert to EUR
		}
	}

	if dryRun {
		// DRY-RUN MODE: Log the action without saving to database
		s.logger.Printf("Data integration [DRY-RUN]: would save metrics for device_id=%d at %s", deviceID, timestamp.Format(time.RFC3339))
		s.logger.Printf("  PV: %.3f kWh, Grid Import: %.3f kWh (€%.3f), Grid Export: %.3f kWh (€%.3f)",
			data.pvTotalPower, data.gridImportPower, gridImportCost, data.gridExportPower, gridExportCost)
		s.logger.Printf("  Battery Charge: %.3f kWh, Battery Discharge: %.3f kWh, SOC: %.1f%%",
			data.batteryChargePower, data.batteryDischargePower, data.batterySoc)
		s.logger.Printf("  EV Charge: %.3f kWh, Load: %.3f kWh", data.evdcChargePower, data.loadPower)
		if cloudCoverage != nil {
			s.logger.Printf("  Cloud Coverage: %.1f%%", *cloudCoverage)
		}
		if weatherSymbol != nil {
			s.logger.Printf("  Weather: %s", *weatherSymbol)
		}
	} else {
		// Insert comprehensive energy flow data
		_, err = dataDB.Exec(
			`INSERT INTO metrics (
				timestamp, device_id, metric_name,
				pv_total_power, cloud_coverage, weather_symbol,
				grid_export_power, grid_import_power,
				battery_charge_power, battery_discharge_power, battery_soc,
				evdc_charge_power, load_power,
				grid_export_cost, grid_import_cost
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
			timestamp, deviceID, "energy_flow",
			data.pvTotalPower, cloudCoverage, weatherSymbol,
			data.gridExportPower, data.gridImportPower,
			data.batteryChargePower, data.batteryDischargePower, data.batterySoc,
			data.evdcChargePower, data.loadPower,
			gridExportCost, gridImportCost,
		)
		if err != nil {
			s.logger.Printf("Data integration: failed to insert metrics: %v", err)
			return
		}

		s.logger.Printf("Data integration: saved metrics for device_id=%d at %s", deviceID, timestamp.Format(time.RFC3339))
		s.logger.Printf("  PV: %.3f kWh, Grid Import: %.3f kWh (€%.3f), Grid Export: %.3f kWh (€%.3f)",
			data.pvTotalPower, data.gridImportPower, gridImportCost, data.gridExportPower, gridExportCost)
		s.logger.Printf("  Battery Charge: %.3f kWh, Battery Discharge: %.3f kWh, SOC: %.1f%%",
			data.batteryChargePower, data.batteryDischargePower, data.batterySoc)
		s.logger.Printf("  EV Charge: %.3f kWh, Load: %.3f kWh", data.evdcChargePower, data.loadPower)
	}
}

func (s *MinerScheduler) fetchCloudCoverage() (*float64, error) {
	// Check cache first
	if cachedForecast, ok := s.weatherCache.Get(); ok {
		current := cachedForecast.GetCurrentWeather()
		if current == nil {
			return nil, nil
		}
		return current.GetCloudCoverage(), nil
	}

	// Cache miss, fetch from API
	s.logger.Printf("Data integration: fetching weather forecast from API")
	config := s.GetConfig()
	client := meteo.NewClient(config.UserAgent)

	location := meteo.Location{
		Latitude:  config.Latitude,
		Longitude: config.Longitude,
	}

	params := meteo.QueryParams{Location: location}
	forecast, err := client.GetCompact(params)
	if err != nil {
		return nil, err
	}

	// Store in cache
	s.weatherCache.Set(forecast)

	current := forecast.GetCurrentWeather()
	if current == nil {
		return nil, nil
	}

	return current.GetCloudCoverage(), nil
}

func (s *MinerScheduler) fetchWeatherSymbol() (*string, error) {
	// Check cache first
	if cachedForecast, ok := s.weatherCache.Get(); ok {
		current := cachedForecast.GetCurrentWeather()
		if current == nil {
			return nil, nil
		}
		symbol := current.GetSymbolCode()
		if symbol == nil {
			return nil, nil
		}
		symbolStr := string(*symbol)
		return &symbolStr, nil
	}

	// Cache miss, fetch from API
	config := s.GetConfig()
	client := meteo.NewClient(config.UserAgent)

	location := meteo.Location{
		Latitude:  config.Latitude,
		Longitude: config.Longitude,
	}

	params := meteo.QueryParams{Location: location}
	forecast, err := client.GetCompact(params)
	if err != nil {
		return nil, err
	}

	// Store in cache
	s.weatherCache.Set(forecast)

	current := forecast.GetCurrentWeather()
	if current == nil {
		return nil, nil
	}

	symbol := current.GetSymbolCode()
	if symbol == nil {
		return nil, nil
	}
	symbolStr := string(*symbol)
	return &symbolStr, nil
}

// GetPlantRunningInfo returns the current plant running information
// If PlantModbusAddress is not configured, returns nil
func (s *MinerScheduler) GetPlantRunningInfo() *sigenergy.PlantRunningInfo {
	if s.config.PlantModbusAddress == "" {
		return nil
	}

	client, err := sigenergy.NewTCPClient(s.config.PlantModbusAddress, sigenergy.PlantAddress)
	if err != nil {
		s.logger.Printf("Failed to create modbus client for plant info: %v", err)
		return nil
	}
	defer client.Close()

	info, err := client.ReadPlantRunningInfo()
	if err != nil {
		s.logger.Printf("Failed to read plant running info: %v", err)
		return nil
	}

	return info
}
