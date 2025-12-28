package meteo

import (
	"testing"
	"time"
)

func TestMETJSONForecast_GetCurrentWeather(t *testing.T) {
	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	forecast := &METJSONForecast{
		Properties: &Forecast{
			Timeseries: []ForecastTimeStep{
				{Time: past, Data: &ForecastTimeStepData{}},
				{Time: future, Data: &ForecastTimeStepData{}},
				{Time: now.Add(30 * time.Minute), Data: &ForecastTimeStepData{}}, // Closest to now
			},
		},
	}

	current := forecast.GetCurrentWeather()
	if current == nil {
		t.Fatal("GetCurrentWeather returned nil")
	}

	// Should return the time step closest to now (30 minutes in the future)
	expected := now.Add(30 * time.Minute)
	if !current.Time.Equal(expected) {
		t.Errorf("Expected time %v, got %v", expected, current.Time)
	}
}

func TestMETJSONForecast_GetCurrentWeather_NilForecast(t *testing.T) {
	var forecast *METJSONForecast
	current := forecast.GetCurrentWeather()
	if current != nil {
		t.Error("Expected nil for nil forecast")
	}
}

func TestMETJSONForecast_GetCurrentWeather_EmptyTimeseries(t *testing.T) {
	forecast := &METJSONForecast{
		Properties: &Forecast{
			Timeseries: []ForecastTimeStep{},
		},
	}

	current := forecast.GetCurrentWeather()
	if current != nil {
		t.Error("Expected nil for empty timeseries")
	}
}

func TestMETJSONForecast_GetWeatherAtTime(t *testing.T) {
	target := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	closest := time.Date(2023, 1, 1, 12, 30, 0, 0, time.UTC) // 30 minutes after target
	farther := time.Date(2023, 1, 1, 14, 0, 0, 0, time.UTC)  // 2 hours after target

	forecast := &METJSONForecast{
		Properties: &Forecast{
			Timeseries: []ForecastTimeStep{
				{Time: farther, Data: &ForecastTimeStepData{}},
				{Time: closest, Data: &ForecastTimeStepData{}},
			},
		},
	}

	weather := forecast.GetWeatherAtTime(target)
	if weather == nil {
		t.Fatal("GetWeatherAtTime returned nil")
	}

	if !weather.Time.Equal(closest) {
		t.Errorf("Expected time %v, got %v", closest, weather.Time)
	}
}

func TestMETJSONForecast_GetDayForecast(t *testing.T) {
	date := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	// Times for the test
	beforeDay := date.Add(-1 * time.Hour)               // Previous day
	startOfDay := date                                  // Start of target day
	midDay := date.Add(12 * time.Hour)                  // Middle of target day
	endOfDay := date.Add(23*time.Hour + 59*time.Minute) // End of target day
	afterDay := date.Add(24 * time.Hour)                // Next day

	forecast := &METJSONForecast{
		Properties: &Forecast{
			Timeseries: []ForecastTimeStep{
				{Time: beforeDay, Data: &ForecastTimeStepData{}},
				{Time: startOfDay, Data: &ForecastTimeStepData{}},
				{Time: midDay, Data: &ForecastTimeStepData{}},
				{Time: endOfDay, Data: &ForecastTimeStepData{}},
				{Time: afterDay, Data: &ForecastTimeStepData{}},
			},
		},
	}

	dayForecast := forecast.GetDayForecast(date)
	if len(dayForecast) != 2 { // midDay and endOfDay should be included
		t.Errorf("Expected 2 time steps for the day, got %d", len(dayForecast))
	}

	// Verify the correct times are included
	expectedTimes := []time.Time{midDay, endOfDay}
	for i, step := range dayForecast {
		if !step.Time.Equal(expectedTimes[i]) {
			t.Errorf("Expected time %v at index %d, got %v", expectedTimes[i], i, step.Time)
		}
	}
}

func TestMETJSONForecast_GetForecastForPeriod(t *testing.T) {
	start := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	end := start.Add(6 * time.Hour)

	times := []time.Time{
		start.Add(-1 * time.Hour), // Before period
		start,                     // Start of period
		start.Add(2 * time.Hour),  // Within period
		end,                       // End of period
		end.Add(1 * time.Hour),    // After period
	}

	var timeseries []ForecastTimeStep
	for _, t := range times {
		timeseries = append(timeseries, ForecastTimeStep{
			Time: t,
			Data: &ForecastTimeStepData{},
		})
	}

	forecast := &METJSONForecast{
		Properties: &Forecast{
			Timeseries: timeseries,
		},
	}

	periodForecast := forecast.GetForecastForPeriod(start, end)
	if len(periodForecast) != 3 { // start, within, and end
		t.Errorf("Expected 3 time steps for the period, got %d", len(periodForecast))
	}
}

func TestForecastTimeStep_HasPrecipitation(t *testing.T) {
	tests := []struct {
		name     string
		timeStep *ForecastTimeStep
		expected bool
	}{
		{
			name:     "nil time step",
			timeStep: nil,
			expected: false,
		},
		{
			name: "no precipitation data",
			timeStep: &ForecastTimeStep{
				Data: &ForecastTimeStepData{
					Instant: &ForecastInstantData{},
				},
			},
			expected: false,
		},
		{
			name: "precipitation in next 1 hour",
			timeStep: &ForecastTimeStep{
				Data: &ForecastTimeStepData{
					Next1Hours: &ForecastPeriodData{
						Details: &ForecastTimePeriod{
							PrecipitationAmount: Float64Ptr(2.5),
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "zero precipitation in next 1 hour",
			timeStep: &ForecastTimeStep{
				Data: &ForecastTimeStepData{
					Next1Hours: &ForecastPeriodData{
						Details: &ForecastTimePeriod{
							PrecipitationAmount: Float64Ptr(0.0),
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "precipitation in next 6 hours",
			timeStep: &ForecastTimeStep{
				Data: &ForecastTimeStepData{
					Next6Hours: &ForecastPeriodData{
						Details: &ForecastTimePeriod{
							PrecipitationAmount: Float64Ptr(1.2),
						},
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.timeStep.HasPrecipitation()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestForecastTimeStep_GetTemperature(t *testing.T) {
	tests := []struct {
		name     string
		timeStep *ForecastTimeStep
		expected *float64
	}{
		{
			name:     "nil time step",
			timeStep: nil,
			expected: nil,
		},
		{
			name: "valid temperature",
			timeStep: &ForecastTimeStep{
				Data: &ForecastTimeStepData{
					Instant: &ForecastInstantData{
						Details: &ForecastTimeInstant{
							AirTemperature: Float64Ptr(15.5),
						},
					},
				},
			},
			expected: Float64Ptr(15.5),
		},
		{
			name: "missing instant data",
			timeStep: &ForecastTimeStep{
				Data: &ForecastTimeStepData{},
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.timeStep.GetTemperature()
			if (result == nil) != (tt.expected == nil) {
				t.Errorf("Expected nil status %v, got %v", tt.expected == nil, result == nil)
			}
			if result != nil && tt.expected != nil && *result != *tt.expected {
				t.Errorf("Expected temperature %.1f, got %.1f", *tt.expected, *result)
			}
		})
	}
}

func TestForecastTimeStep_GetSymbolCode(t *testing.T) {
	tests := []struct {
		name     string
		timeStep *ForecastTimeStep
		expected *WeatherSymbol
	}{
		{
			name:     "nil time step",
			timeStep: nil,
			expected: nil,
		},
		{
			name: "symbol from next 1 hour",
			timeStep: &ForecastTimeStep{
				Data: &ForecastTimeStepData{
					Next1Hours: &ForecastPeriodData{
						Summary: &ForecastSummary{
							SymbolCode: ClearSkyDay,
						},
					},
				},
			},
			expected: func() *WeatherSymbol { s := ClearSkyDay; return &s }(),
		},
		{
			name: "symbol from next 6 hours (fallback)",
			timeStep: &ForecastTimeStep{
				Data: &ForecastTimeStepData{
					Next6Hours: &ForecastPeriodData{
						Summary: &ForecastSummary{
							SymbolCode: Rain,
						},
					},
				},
			},
			expected: func() *WeatherSymbol { s := Rain; return &s }(),
		},
		{
			name: "symbol from next 12 hours (fallback)",
			timeStep: &ForecastTimeStep{
				Data: &ForecastTimeStepData{
					Next12Hours: &ForecastPeriodData{
						Summary: &ForecastSummary{
							SymbolCode: PartlyCloudyNight,
						},
					},
				},
			},
			expected: func() *WeatherSymbol { s := PartlyCloudyNight; return &s }(),
		},
		{
			name: "no symbol available",
			timeStep: &ForecastTimeStep{
				Data: &ForecastTimeStepData{
					Instant: &ForecastInstantData{},
				},
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.timeStep.GetSymbolCode()
			if (result == nil) != (tt.expected == nil) {
				t.Errorf("Expected nil status %v, got %v", tt.expected == nil, result == nil)
			}
			if result != nil && tt.expected != nil && *result != *tt.expected {
				t.Errorf("Expected symbol %s, got %s", *tt.expected, *result)
			}
		})
	}
}

func TestWeatherSymbol_IsDay(t *testing.T) {
	tests := []struct {
		symbol   WeatherSymbol
		expected bool
	}{
		{ClearSkyDay, true},
		{ClearSkyNight, false},
		{PartlyCloudyPolarTwilight, false},
		{RainShowersDay, true},
		{Snow, false}, // No day/night suffix
	}

	for _, tt := range tests {
		t.Run(string(tt.symbol), func(t *testing.T) {
			result := tt.symbol.IsDay()
			if result != tt.expected {
				t.Errorf("Expected IsDay() = %v for symbol %s", tt.expected, tt.symbol)
			}
		})
	}
}

func TestWeatherSymbol_IsNight(t *testing.T) {
	tests := []struct {
		symbol   WeatherSymbol
		expected bool
	}{
		{ClearSkyDay, false},
		{ClearSkyNight, true},
		{PartlyCloudyPolarTwilight, false},
		{RainShowersNight, true},
		{Snow, false}, // No day/night suffix
	}

	for _, tt := range tests {
		t.Run(string(tt.symbol), func(t *testing.T) {
			result := tt.symbol.IsNight()
			if result != tt.expected {
				t.Errorf("Expected IsNight() = %v for symbol %s", tt.expected, tt.symbol)
			}
		})
	}
}

func TestWeatherSymbol_IsPolarTwilight(t *testing.T) {
	tests := []struct {
		symbol   WeatherSymbol
		expected bool
	}{
		{ClearSkyDay, false},
		{ClearSkyPolarTwilight, true},
		{PartlyCloudyPolarTwilight, true},
		{RainShowersDay, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.symbol), func(t *testing.T) {
			result := tt.symbol.IsPolarTwilight()
			if result != tt.expected {
				t.Errorf("Expected IsPolarTwilight() = %v for symbol %s", tt.expected, tt.symbol)
			}
		})
	}
}

func TestWeatherSymbol_HasThunder(t *testing.T) {
	tests := []struct {
		symbol   WeatherSymbol
		expected bool
	}{
		{ClearSkyDay, false},
		{RainAndThunder, true},
		{LightRainShowersAndThunderDay, true},
		{HeavySnowAndThunder, true},
		{Rain, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.symbol), func(t *testing.T) {
			result := tt.symbol.HasThunder()
			if result != tt.expected {
				t.Errorf("Expected HasThunder() = %v for symbol %s", tt.expected, tt.symbol)
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test IntPtr
	intVal := 42
	intPtr := IntPtr(intVal)
	if intPtr == nil || *intPtr != intVal {
		t.Errorf("IntPtr failed: expected %d, got %v", intVal, intPtr)
	}

	// Test Float64Ptr
	floatVal := 3.14
	floatPtr := Float64Ptr(floatVal)
	if floatPtr == nil || *floatPtr != floatVal {
		t.Errorf("Float64Ptr failed: expected %f, got %v", floatVal, floatPtr)
	}

	// Test StringPtr
	stringVal := "test"
	stringPtr := StringPtr(stringVal)
	if stringPtr == nil || *stringPtr != stringVal {
		t.Errorf("StringPtr failed: expected %s, got %v", stringVal, stringPtr)
	}
}
