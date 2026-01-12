package meteo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	userAgent := "TestApp/1.0 (test@example.com)"
	client := NewClient(userAgent)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.userAgent != userAgent {
		t.Errorf("Expected user agent %q, got %q", userAgent, client.userAgent)
	}

	if client.baseURL != "https://api.met.no/weatherapi/locationforecast/2.0" {
		t.Errorf("Expected default base URL, got %q", client.baseURL)
	}

	if client.httpClient == nil {
		t.Error("HTTP client is nil")
	}
}

func TestNewClientWithHTTPClient(t *testing.T) {
	httpClient := &http.Client{Timeout: 5 * time.Second}
	userAgent := "TestApp/1.0"
	client := NewClientWithHTTPClient(httpClient, userAgent)

	if client.httpClient != httpClient {
		t.Error("Custom HTTP client was not set")
	}
}

func TestSetBaseURL(t *testing.T) {
	client := NewClient("TestApp/1.0")
	newURL := "https://custom.example.com/api"

	client.SetBaseURL(newURL)

	if client.baseURL != newURL {
		t.Errorf("Expected base URL %q, got %q", newURL, client.baseURL)
	}
}

func TestBuildURL(t *testing.T) {
	client := NewClient("TestApp/1.0")
	client.SetBaseURL("https://api.example.com")

	tests := []struct {
		name     string
		endpoint string
		params   QueryParams
		expected string
	}{
		{
			name:     "compact endpoint basic",
			endpoint: "compact",
			params: QueryParams{
				Location: Location{
					Latitude:  59.9139,
					Longitude: 10.7522,
				},
			},
			expected: "https://api.example.com/compact?lat=59.9139&lon=10.7522",
		},
		{
			name:     "with altitude",
			endpoint: "complete",
			params: QueryParams{
				Location: Location{
					Latitude:  60.5,
					Longitude: 11.59,
					Altitude:  IntPtr(1001),
				},
			},
			expected: "https://api.example.com/complete?altitude=1001&lat=60.5&lon=11.59",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := client.buildURL(tt.endpoint, tt.params)
			if err != nil {
				t.Fatalf("buildURL returned error: %v", err)
			}
			if url != tt.expected {
				t.Errorf("Expected URL %q, got %q", tt.expected, url)
			}
		})
	}
}

func TestValidateLocation(t *testing.T) {
	tests := []struct {
		name        string
		location    Location
		expectError bool
	}{
		{
			name: "valid location",
			location: Location{
				Latitude:  59.9139,
				Longitude: 10.7522,
			},
			expectError: false,
		},
		{
			name: "valid with altitude",
			location: Location{
				Latitude:  60.0,
				Longitude: 11.0,
				Altitude:  IntPtr(500),
			},
			expectError: false,
		},
		{
			name: "latitude too high",
			location: Location{
				Latitude:  91.0,
				Longitude: 10.0,
			},
			expectError: true,
		},
		{
			name: "latitude too low",
			location: Location{
				Latitude:  -91.0,
				Longitude: 10.0,
			},
			expectError: true,
		},
		{
			name: "longitude too high",
			location: Location{
				Latitude:  60.0,
				Longitude: 181.0,
			},
			expectError: true,
		},
		{
			name: "longitude too low",
			location: Location{
				Latitude:  60.0,
				Longitude: -181.0,
			},
			expectError: true,
		},
		{
			name: "negative altitude",
			location: Location{
				Latitude:  60.0,
				Longitude: 11.0,
				Altitude:  IntPtr(-100),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLocation(tt.location)
			if tt.expectError && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

func TestGetCompact(t *testing.T) {
	// Create test forecast data
	testForecast := METJSONForecast{
		Type: "Feature",
		Geometry: &PointGeometry{
			Type:        "Point",
			Coordinates: []float64{10.7522, 59.9139, 14},
		},
		Properties: &Forecast{
			Meta: ForecastMeta{
				UpdatedAt: time.Now(),
				Units: ForecastUnits{
					AirTemperature: StringPtr("celsius"),
					WindSpeed:      StringPtr("m/s"),
				},
			},
			Timeseries: []ForecastTimeStep{
				{
					Time: time.Now(),
					Data: &ForecastTimeStepData{
						Instant: &ForecastInstantData{
							Details: &ForecastTimeInstant{
								AirTemperature: Float64Ptr(15.5),
								WindSpeed:      Float64Ptr(3.2),
							},
						},
						Next1Hours: &ForecastPeriodData{
							Summary: &ForecastSummary{
								SymbolCode: ClearSkyDay,
							},
							Details: &ForecastTimePeriod{
								PrecipitationAmount: Float64Ptr(0.0),
							},
						},
					},
				},
			},
		},
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("User-Agent") != "TestApp/1.0" {
			t.Errorf("Expected User-Agent 'TestApp/1.0', got '%s'", r.Header.Get("User-Agent"))
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Expected Accept 'application/json', got '%s'", r.Header.Get("Accept"))
		}

		// Verify URL parameters
		if r.URL.Query().Get("lat") != "59.9139" {
			t.Errorf("Expected lat parameter '59.9139', got '%s'", r.URL.Query().Get("lat"))
		}
		if r.URL.Query().Get("lon") != "10.7522" {
			t.Errorf("Expected lon parameter '10.7522', got '%s'", r.URL.Query().Get("lon"))
		}

		// Return test data
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(testForecast)
	}))
	defer server.Close()

	// Create client and set test server URL
	client := NewClient("TestApp/1.0")
	client.SetBaseURL(server.URL)

	// Make request
	params := QueryParams{
		Location: Location{
			Latitude:  59.9139,
			Longitude: 10.7522,
		},
	}

	forecast, err := client.GetCompact(params)
	if err != nil {
		t.Fatalf("GetCompact returned error: %v", err)
	}

	// Verify response
	if forecast.Type != "Feature" {
		t.Errorf("Expected type 'Feature', got '%s'", forecast.Type)
	}

	if forecast.Properties == nil {
		t.Fatal("Properties is nil")
	}

	if len(forecast.Properties.Timeseries) != 1 {
		t.Errorf("Expected 1 time step, got %d", len(forecast.Properties.Timeseries))
	}

	step := forecast.Properties.Timeseries[0]
	if temp := step.GetTemperature(); temp == nil || *temp != 15.5 {
		t.Errorf("Expected temperature 15.5, got %v", temp)
	}
}

func TestAPIError(t *testing.T) {
	// Create test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request: Invalid parameters"))
	}))
	defer server.Close()

	client := NewClient("TestApp/1.0")
	client.SetBaseURL(server.URL)

	params := QueryParams{
		Location: Location{
			Latitude:  59.9139,
			Longitude: 10.7522,
		},
	}

	_, err := client.GetCompact(params)
	if err == nil {
		t.Fatal("Expected API error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError, got %T", err)
	}

	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, apiErr.StatusCode)
	}

	expectedMessage := "Bad Request: Invalid parameters"
	if apiErr.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, apiErr.Message)
	}
}

func TestFormatFloat(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{59.9139, "59.9139"},
		{10.0, "10"},
		{-123.456789, "-123.456789"},
		{0.0, "0"},
	}

	for _, tt := range tests {
		result := formatFloat(tt.input)
		if result != tt.expected {
			t.Errorf("formatFloat(%.6f) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}
