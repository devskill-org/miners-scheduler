package meteo

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestJSONDeserialization(t *testing.T) {
	// Read the example JSON file
	data, err := os.ReadFile("../test_data/locationforecast/example.json")
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}

	// Deserialize the JSON
	var forecast METJSONForecast
	if err := json.Unmarshal(data, &forecast); err != nil {
		t.Fatalf("Failed to deserialize JSON: %v", err)
	}

	// Validate the basic structure
	if forecast.Type != "Feature" {
		t.Errorf("Expected type 'Feature', got '%s'", forecast.Type)
	}

	if forecast.Geometry == nil {
		t.Fatal("Geometry is nil")
	}

	if forecast.Geometry.Type != "Point" {
		t.Errorf("Expected geometry type 'Point', got '%s'", forecast.Geometry.Type)
	}

	if len(forecast.Geometry.Coordinates) != 3 {
		t.Errorf("Expected 3 coordinates, got %d", len(forecast.Geometry.Coordinates))
	}

	if forecast.Properties == nil {
		t.Fatal("Properties is nil")
	}

	// Validate metadata
	meta := forecast.Properties.Meta
	if meta.UpdatedAt.IsZero() {
		t.Error("UpdatedAt is zero")
	}

	if meta.Units.AirTemperature == nil {
		t.Error("Air temperature unit is nil")
	}

	// Validate timeseries
	if len(forecast.Properties.Timeseries) == 0 {
		t.Fatal("No timeseries data")
	}

	// Check first time step
	firstStep := forecast.Properties.Timeseries[0]
	if firstStep.Time.IsZero() {
		t.Error("First time step has zero time")
	}

	if firstStep.Data == nil {
		t.Fatal("First time step data is nil")
	}

	if firstStep.Data.Instant == nil {
		t.Fatal("First time step instant data is nil")
	}

	if firstStep.Data.Instant.Details == nil {
		t.Fatal("First time step instant details is nil")
	}

	// Validate some expected data fields
	temp := firstStep.GetTemperature()
	if temp == nil {
		t.Error("Temperature is nil in first time step")
	}

	windSpeed := firstStep.GetWindSpeed()
	if windSpeed == nil {
		t.Error("Wind speed is nil in first time step")
	}

	// Test utility functions with real data
	current := forecast.GetCurrentWeather()
	if current == nil {
		t.Error("GetCurrentWeather returned nil")
	}

	// Test day forecast
	now := time.Now()
	dayForecast := forecast.GetDayForecast(now)
	t.Logf("Found %d forecast steps for today", len(dayForecast))

	// Test period forecast
	periodForecast := forecast.GetForecastForPeriod(now, now.Add(24*time.Hour))
	t.Logf("Found %d forecast steps for next 24 hours", len(periodForecast))

	// Count time steps with precipitation
	precipitationCount := 0
	for _, step := range forecast.Properties.Timeseries {
		if step.HasPrecipitation() {
			precipitationCount++
		}
	}
	t.Logf("Found %d time steps with precipitation", precipitationCount)

	// Count different weather symbols
	symbolCount := make(map[WeatherSymbol]int)
	for _, step := range forecast.Properties.Timeseries {
		if symbol := step.GetSymbolCode(); symbol != nil {
			symbolCount[*symbol]++
		}
	}
	t.Logf("Found %d different weather symbols", len(symbolCount))
}

func TestJSONSerialization(t *testing.T) {
	// Create a test forecast
	now := time.Now()
	forecast := METJSONForecast{
		Type: "Feature",
		Geometry: &PointGeometry{
			Type:        "Point",
			Coordinates: []float64{10.7522, 59.9139, 14},
		},
		Properties: &Forecast{
			Meta: ForecastMeta{
				UpdatedAt: now,
				Units: ForecastUnits{
					AirTemperature:      StringPtr("celsius"),
					WindSpeed:           StringPtr("m/s"),
					PrecipitationAmount: StringPtr("mm"),
				},
			},
			Timeseries: []ForecastTimeStep{
				{
					Time: now,
					Data: &ForecastTimeStepData{
						Instant: &ForecastInstantData{
							Details: &ForecastTimeInstant{
								AirTemperature:    Float64Ptr(15.5),
								WindSpeed:         Float64Ptr(3.2),
								RelativeHumidity:  Float64Ptr(85.0),
								CloudAreaFraction: Float64Ptr(50.0),
							},
						},
						Next1Hours: &ForecastPeriodData{
							Summary: &ForecastSummary{
								SymbolCode: PartlyCloudyDay,
							},
							Details: &ForecastTimePeriod{
								PrecipitationAmount: Float64Ptr(0.1),
							},
						},
						Next6Hours: &ForecastPeriodData{
							Summary: &ForecastSummary{
								SymbolCode: LightRainShowersDay,
							},
							Details: &ForecastTimePeriod{
								PrecipitationAmount: Float64Ptr(2.5),
							},
						},
					},
				},
			},
		},
	}

	// Serialize to JSON
	data, err := json.MarshalIndent(forecast, "", "  ")
	if err != nil {
		t.Fatalf("Failed to serialize forecast: %v", err)
	}

	// Deserialize back
	var deserializedForecast METJSONForecast
	if err := json.Unmarshal(data, &deserializedForecast); err != nil {
		t.Fatalf("Failed to deserialize forecast: %v", err)
	}

	// Validate round-trip
	if deserializedForecast.Type != forecast.Type {
		t.Errorf("Type mismatch after round-trip: expected %s, got %s",
			forecast.Type, deserializedForecast.Type)
	}

	if len(deserializedForecast.Properties.Timeseries) != len(forecast.Properties.Timeseries) {
		t.Errorf("Timeseries length mismatch after round-trip: expected %d, got %d",
			len(forecast.Properties.Timeseries), len(deserializedForecast.Properties.Timeseries))
	}

	// Check specific values
	originalTemp := forecast.Properties.Timeseries[0].GetTemperature()
	deserializedTemp := deserializedForecast.Properties.Timeseries[0].GetTemperature()

	if originalTemp == nil || deserializedTemp == nil {
		t.Fatal("Temperature is nil after round-trip")
	}

	if *originalTemp != *deserializedTemp {
		t.Errorf("Temperature mismatch after round-trip: expected %.1f, got %.1f",
			*originalTemp, *deserializedTemp)
	}
}

func TestWeatherSymbolConstants(t *testing.T) {
	// Test that all weather symbol constants are valid strings
	symbols := []WeatherSymbol{
		ClearSkyDay, ClearSkyNight, ClearSkyPolarTwilight,
		FairDay, FairNight, FairPolarTwilight,
		PartlyCloudyDay, PartlyCloudyNight, PartlyCloudyPolarTwilight,
		Cloudy, Fog,
		LightRain, Rain, HeavyRain,
		LightSnow, Snow, HeavySnow,
		LightSleet, Sleet, HeavySleet,
		RainAndThunder, LightRainAndThunder,
		SnowAndThunder, LightSnowAndThunder, HeavySnowAndThunder,
		SleetAndThunder, LightSleetAndThunder, HeavySleetAndThunder,
		RainShowersDay, RainShowersNight, RainShowersPolarTwilight,
		LightRainShowersDay, LightRainShowersNight, LightRainShowersPolarTwilight,
		HeavyRainShowersDay, HeavyRainShowersNight, HeavyRainShowersPolarTwilight,
	}

	for _, symbol := range symbols {
		if string(symbol) == "" {
			t.Errorf("Empty weather symbol: %v", symbol)
		}

		// Test JSON marshaling
		data, err := json.Marshal(symbol)
		if err != nil {
			t.Errorf("Failed to marshal symbol %s: %v", symbol, err)
		}

		var unmarshaled WeatherSymbol
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Errorf("Failed to unmarshal symbol %s: %v", symbol, err)
		}

		if unmarshaled != symbol {
			t.Errorf("Symbol round-trip failed: expected %s, got %s", symbol, unmarshaled)
		}
	}
}

func TestLocationValidationIntegration(t *testing.T) {
	// Test with actual coordinate examples
	validLocations := []Location{
		{Latitude: 59.9139, Longitude: 10.7522},       // Oslo
		{Latitude: 60.472024, Longitude: 8.468946},    // Geilo
		{Latitude: 69.649208, Longitude: 18.955324},   // Tromsø
		{Latitude: -33.868820, Longitude: 151.209290}, // Sydney
		{Latitude: 40.712776, Longitude: -74.005974},  // New York
		{Latitude: 0, Longitude: 0},                   // Null Island
		{Latitude: 90, Longitude: 180},                // Extreme valid
		{Latitude: -90, Longitude: -180},              // Extreme valid
	}

	for _, loc := range validLocations {
		if err := ValidateLocation(loc); err != nil {
			t.Errorf("Valid location failed validation: %+v, error: %v", loc, err)
		}
	}

	invalidLocations := []Location{
		{Latitude: 91, Longitude: 0},                          // Latitude too high
		{Latitude: -91, Longitude: 0},                         // Latitude too low
		{Latitude: 0, Longitude: 181},                         // Longitude too high
		{Latitude: 0, Longitude: -181},                        // Longitude too low
		{Latitude: 60, Longitude: 10, Altitude: IntPtr(-100)}, // Negative altitude
	}

	for _, loc := range invalidLocations {
		if err := ValidateLocation(loc); err == nil {
			t.Errorf("Invalid location passed validation: %+v", loc)
		}
	}
}
