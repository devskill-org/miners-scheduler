package meteo

import (
	"time"
)

// GetCurrentWeather returns the current weather data from the forecast
func (f *METJSONForecast) GetCurrentWeather() *ForecastTimeStep {
	if f == nil || f.Properties == nil || len(f.Properties.Timeseries) == 0 {
		return nil
	}

	now := time.Now()
	var closest *ForecastTimeStep
	minDiff := time.Duration(1<<63 - 1) // Max duration

	for i := range f.Properties.Timeseries {
		step := &f.Properties.Timeseries[i]
		diff := step.Time.Sub(now)
		if diff < 0 {
			diff = -diff
		}
		if diff < minDiff {
			minDiff = diff
			closest = step
		}
	}

	return closest
}

// GetWeatherAtTime returns the weather data closest to the specified time
func (f *METJSONForecast) GetWeatherAtTime(targetTime time.Time) *ForecastTimeStep {
	if f == nil || f.Properties == nil || len(f.Properties.Timeseries) == 0 {
		return nil
	}

	var closest *ForecastTimeStep
	minDiff := time.Duration(1<<63 - 1) // Max duration

	for i := range f.Properties.Timeseries {
		step := &f.Properties.Timeseries[i]
		diff := step.Time.Sub(targetTime)
		if diff < 0 {
			diff = -diff
		}
		if diff < minDiff {
			minDiff = diff
			closest = step
		}
	}

	return closest
}

// GetDayForecast returns all weather data for a specific day
func (f *METJSONForecast) GetDayForecast(date time.Time) []ForecastTimeStep {
	if f == nil || f.Properties == nil {
		return nil
	}

	var dayForecast []ForecastTimeStep
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	for _, step := range f.Properties.Timeseries {
		if step.Time.After(startOfDay) && step.Time.Before(endOfDay) {
			dayForecast = append(dayForecast, step)
		}
	}

	return dayForecast
}

// GetForecastForPeriod returns all weather data within the specified time period
func (f *METJSONForecast) GetForecastForPeriod(start, end time.Time) []ForecastTimeStep {
	if f == nil || f.Properties == nil {
		return nil
	}

	var periodForecast []ForecastTimeStep
	for _, step := range f.Properties.Timeseries {
		if (step.Time.Equal(start) || step.Time.After(start)) &&
			(step.Time.Equal(end) || step.Time.Before(end)) {
			periodForecast = append(periodForecast, step)
		}
	}

	return periodForecast
}

// HasPrecipitation checks if there's any precipitation in the given time step
func (ts *ForecastTimeStep) HasPrecipitation() bool {
	if ts == nil || ts.Data == nil {
		return false
	}

	// Check 1-hour precipitation
	if ts.Data.Next1Hours != nil &&
		ts.Data.Next1Hours.Details != nil &&
		ts.Data.Next1Hours.Details.PrecipitationAmount != nil &&
		*ts.Data.Next1Hours.Details.PrecipitationAmount > 0 {
		return true
	}

	// Check 6-hour precipitation
	if ts.Data.Next6Hours != nil &&
		ts.Data.Next6Hours.Details != nil &&
		ts.Data.Next6Hours.Details.PrecipitationAmount != nil &&
		*ts.Data.Next6Hours.Details.PrecipitationAmount > 0 {
		return true
	}

	return false
}

// GetTemperature returns the air temperature if available
func (ts *ForecastTimeStep) GetTemperature() *float64 {
	if ts == nil || ts.Data == nil || ts.Data.Instant == nil || ts.Data.Instant.Details == nil {
		return nil
	}
	return ts.Data.Instant.Details.AirTemperature
}

// GetWindSpeed returns the wind speed if available
func (ts *ForecastTimeStep) GetWindSpeed() *float64 {
	if ts == nil || ts.Data == nil || ts.Data.Instant == nil || ts.Data.Instant.Details == nil {
		return nil
	}
	return ts.Data.Instant.Details.WindSpeed
}

// GetWindDirection returns the wind direction if available
func (ts *ForecastTimeStep) GetWindDirection() *float64 {
	if ts == nil || ts.Data == nil || ts.Data.Instant == nil || ts.Data.Instant.Details == nil {
		return nil
	}
	return ts.Data.Instant.Details.WindFromDirection
}

// GetHumidity returns the relative humidity if available
func (ts *ForecastTimeStep) GetHumidity() *float64 {
	if ts == nil || ts.Data == nil || ts.Data.Instant == nil || ts.Data.Instant.Details == nil {
		return nil
	}
	return ts.Data.Instant.Details.RelativeHumidity
}

// GetCloudCoverage returns the cloud area fraction if available
func (ts *ForecastTimeStep) GetCloudCoverage() *float64 {
	if ts == nil || ts.Data == nil || ts.Data.Instant == nil || ts.Data.Instant.Details == nil {
		return nil
	}
	return ts.Data.Instant.Details.CloudAreaFraction
}

// GetSymbolCode returns the weather symbol code for the next hour if available
func (ts *ForecastTimeStep) GetSymbolCode() *WeatherSymbol {
	if ts == nil || ts.Data == nil {
		return nil
	}

	// Try next 1 hour first
	if ts.Data.Next1Hours != nil && ts.Data.Next1Hours.Summary != nil {
		return &ts.Data.Next1Hours.Summary.SymbolCode
	}

	// Fallback to next 6 hours
	if ts.Data.Next6Hours != nil && ts.Data.Next6Hours.Summary != nil {
		return &ts.Data.Next6Hours.Summary.SymbolCode
	}

	// Fallback to next 12 hours
	if ts.Data.Next12Hours != nil && ts.Data.Next12Hours.Summary != nil {
		return &ts.Data.Next12Hours.Summary.SymbolCode
	}

	return nil
}

// IsDay checks if the weather symbol indicates daytime conditions
func (ws WeatherSymbol) IsDay() bool {
	str := string(ws)
	return len(str) >= 4 && str[len(str)-4:] == "_day"
}

// IsNight checks if the weather symbol indicates nighttime conditions
func (ws WeatherSymbol) IsNight() bool {
	str := string(ws)
	return len(str) >= 6 && str[len(str)-6:] == "_night"
}

// IsPolarTwilight checks if the weather symbol indicates polar twilight conditions
func (ws WeatherSymbol) IsPolarTwilight() bool {
	str := string(ws)
	return len(str) >= 14 && str[len(str)-14:] == "_polartwilight"
}

// HasThunder checks if the weather symbol indicates thunder
func (ws WeatherSymbol) HasThunder() bool {
	str := string(ws)
	// Check if contains "thunder" anywhere in the symbol
	for i := 0; i <= len(str)-7; i++ {
		if str[i:i+7] == "thunder" {
			return true
		}
	}
	return false
}

// IntPtr is a helper function to get a pointer to an int value
func IntPtr(i int) *int {
	return &i
}

// Float64Ptr is a helper function to get a pointer to a float64 value
func Float64Ptr(f float64) *float64 {
	return &f
}

// StringPtr is a helper function to get a pointer to a string value
func StringPtr(s string) *string {
	return &s
}
