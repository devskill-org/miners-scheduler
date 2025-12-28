package meteo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Client represents a client for the MET Norway Location Forecast API
type Client struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
}

// NewClient creates a new client for the MET Norway Location Forecast API
func NewClient(userAgent string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:   "https://api.met.no/weatherapi/locationforecast/2.0",
		userAgent: userAgent,
	}
}

// NewClientWithHTTPClient creates a new client with a custom HTTP client
func NewClientWithHTTPClient(httpClient *http.Client, userAgent string) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    "https://api.met.no/weatherapi/locationforecast/2.0",
		userAgent:  userAgent,
	}
}

// SetBaseURL sets the base URL for the API (useful for testing)
func (c *Client) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
}

// GetCompact retrieves compact forecast data for the specified location
func (c *Client) GetCompact(params QueryParams) (*METJSONForecast, error) {
	return c.getForecast("compact", params)
}

// GetComplete retrieves complete forecast data for the specified location
func (c *Client) GetComplete(params QueryParams) (*METJSONForecast, error) {
	return c.getForecast("complete", params)
}

// GetClassic retrieves classic forecast data for the specified location
func (c *Client) GetClassic(params QueryParams) (*METJSONForecast, error) {
	return c.getForecast("classic", params)
}

// getForecast is the internal method that performs the actual API request
func (c *Client) getForecast(endpoint string, params QueryParams) (*METJSONForecast, error) {
	reqURL, err := c.buildURL(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var forecast METJSONForecast
	if err := json.Unmarshal(body, &forecast); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &forecast, nil
}

// buildURL constructs the API URL with query parameters
func (c *Client) buildURL(endpoint string, params QueryParams) (string, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return "", err
	}

	u.Path = fmt.Sprintf("%s/%s", u.Path, endpoint)

	query := u.Query()
	query.Set("lat", formatFloat(params.Location.Latitude))
	query.Set("lon", formatFloat(params.Location.Longitude))

	if params.Location.Altitude != nil {
		query.Set("altitude", strconv.Itoa(*params.Location.Altitude))
	}

	u.RawQuery = query.Encode()
	return u.String(), nil
}

// formatFloat formats a float64 to a string with appropriate precision
func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// ValidateLocation validates that the location parameters are within acceptable ranges
func ValidateLocation(loc Location) error {
	if loc.Latitude < -90 || loc.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90, got %f", loc.Latitude)
	}
	if loc.Longitude < -180 || loc.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180, got %f", loc.Longitude)
	}
	if loc.Altitude != nil && *loc.Altitude < 0 {
		return fmt.Errorf("altitude must be non-negative, got %d", *loc.Altitude)
	}
	return nil
}
