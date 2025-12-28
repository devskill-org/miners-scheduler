package scheduler

import (
	"database/sql"
	"sync"
	"time"

	"github.com/devskill-org/miners-scheduler/meteo"
	"github.com/devskill-org/miners-scheduler/sigenergy"
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

// GetLatestPower returns the most recent PV power sample, or 0 if no samples exist
func (p *PVSamples) GetLatestPower() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.samples) == 0 {
		return 0
	}
	return p.samples[len(p.samples)-1].value
}

func (s *MinerScheduler) runPVPoll(samples *PVSamples) {
	if s.config.PlantModbusAddress == "" {
		return
	}
	client, err := sigenergy.NewTCPClient(s.config.PlantModbusAddress, sigenergy.PlantAddress)
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

func (s *MinerScheduler) runPVIntegration(samples *PVSamples, pollInterval time.Duration, pvDB *sql.DB, deviceID int, dryRun bool) {
	if samples.IsEmpty() {
		s.logger.Printf("PV integration: no samples collected in period")
		return
	}
	total := samples.IntegrateSamples(pollInterval)
	timestamp := time.Now()
	if pvDB != nil {
		// Fetch cloud coverage from meteo API
		cloudCoverage, err := s.fetchCloudCoverage()
		if err != nil {
			s.logger.Printf("PV integration: failed to fetch cloud coverage: %v", err)
		}

		if dryRun {
			// DRY-RUN MODE: Log the action without saving to database
			if cloudCoverage != nil {
				s.logger.Printf("PV integration [DRY-RUN]: would save %.3f kWh with cloud coverage %.1f%% for device_id=%d at %s", total, *cloudCoverage, deviceID, timestamp.Format(time.RFC3339))
			} else {
				s.logger.Printf("PV integration [DRY-RUN]: would save %.3f kWh without cloud coverage for device_id=%d at %s", total, deviceID, timestamp.Format(time.RFC3339))
			}
		} else {
			// Insert with cloud coverage
			_, err = pvDB.Exec(
				`INSERT INTO metrics (timestamp, device_id, metric_name, pv_total_power, cloud_coverage) VALUES ($1, $2, $3, $4, $5)`,
				timestamp, deviceID, "pv_total_power", total, cloudCoverage,
			)
			if err != nil {
				s.logger.Printf("PV integration: failed to insert metric: %v", err)
				return
			}
			if cloudCoverage != nil {
				s.logger.Printf("PV integration: saved %.3f kWh with cloud coverage %.1f%% for device_id=%d at %s", total, *cloudCoverage, deviceID, timestamp.Format(time.RFC3339))
			} else {
				s.logger.Printf("PV integration: saved %.3f kWh without cloud coverage for device_id=%d at %s", total, deviceID, timestamp.Format(time.RFC3339))
			}
		}
	}
}

func (s *MinerScheduler) fetchCloudCoverage() (*float64, error) {
	// Check cache first
	if cachedForecast, ok := s.weatherCache.Get(); ok {
		s.logger.Printf("PV integration: using cached weather forecast")
		current := cachedForecast.GetCurrentWeather()
		if current == nil {
			return nil, nil
		}
		return current.GetCloudCoverage(), nil
	}

	// Cache miss, fetch from API
	s.logger.Printf("PV integration: fetching weather forecast from API")
	client := meteo.NewClient(s.config.UserAgent)

	location := meteo.Location{
		Latitude:  s.config.Latitude,
		Longitude: s.config.Longitude,
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

// GetCurrentPVPower returns the current PV power in kW
// If PlantModbusAddress is not configured, returns 0
func (s *MinerScheduler) GetCurrentPVPower() float64 {
	if s.config.PlantModbusAddress == "" {
		return 0
	}

	client, err := sigenergy.NewTCPClient(s.config.PlantModbusAddress, sigenergy.PlantAddress)
	if err != nil {
		s.logger.Printf("Failed to create modbus client for PV power: %v", err)
		return 0
	}
	defer client.Close()

	info, err := client.ReadPlantRunningInfo()
	if err != nil {
		s.logger.Printf("Failed to read PV power: %v", err)
		return 0
	}

	// Return power in kW
	return info.PhotovoltaicPower
}
