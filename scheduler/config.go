package scheduler

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// Config represents the configuration for the miner scheduler
type Config struct {
	// Scheduler settings
	PriceLimit               float64       `json:"price_limit"`                 // Price limit in EUR/MWh
	Network                  string        `json:"network"`                     // Network to scan for miners (CIDR notation)
	CheckPriceInterval       time.Duration `json:"check_price_interval"`        // How often to run the task
	MinersStateCheckInterval time.Duration `json:"miners_state_check_interval"` // How often to check miners state
	MinerDiscoveryInterval   time.Duration `json:"miner_discovery_interval"`    // How often to discover miners
	DryRun                   bool          `json:"dry_run"`                     // Run in dry-run mode (simulate actions without executing)

	// API settings
	SecurityToken string        `json:"security_token"` // ENTSO-E API token
	APITimeout    time.Duration `json:"api_timeout"`    // Timeout for API calls
	UrlFormat     string        `json:"url_format"`     // ENTSO-E API URL format string

	// Logging settings
	LogLevel  string `json:"log_level"`  // Log level: debug, info, warn, error
	LogFormat string `json:"log_format"` // Log format: text, json

	// Timezone configuration
	Location string `json:"location"` // Timezone location string (e.g., "CET"), when the market data published at 00:00

	// Miner settings
	MinerTimeout time.Duration `json:"miner_timeout"` // Timeout for miner operations

	// Advanced settings
	HealthCheckPort int `json:"health_check_port"` // Port for health check endpoint (0 = disabled)

	// FanR thresholds for work mode switching
	FanRHighThreshold int `json:"fanr_high_threshold"` // FanR threshold to decrease work mode
	FanRLowThreshold  int `json:"fanr_low_threshold"`  // FanR threshold to increase work mode
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		PriceLimit:               60.0,
		Network:                  "192.168.88.0/24",
		CheckPriceInterval:       15 * time.Minute,
		MinersStateCheckInterval: 1 * time.Minute,
		MinerDiscoveryInterval:   10 * time.Minute,
		DryRun:                   false,
		APITimeout:               30 * time.Second,
		LogLevel:                 "info",
		LogFormat:                "text",
		MinerTimeout:             5 * time.Second,
		HealthCheckPort:          0,
		UrlFormat:                "https://web-api.tp.entsoe.eu/api?documentType=A44&out_Domain=10YLV-1001A00074&in_Domain=10YLV-1001A00074&periodStart=%s&periodEnd=%s&securityToken=%s",
	}
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	return LoadConfigFromReader(file)
}

// LoadConfigFromReader loads configuration from an io.Reader
func LoadConfigFromReader(reader io.Reader) (*Config, error) {
	config := DefaultConfig()

	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(config); err != nil {
		return nil, fmt.Errorf("failed to decode config JSON: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// SaveConfig saves the configuration to a JSON file
func (c *Config) SaveConfig(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	return c.SaveConfigToWriter(file)
}

// SaveConfigToWriter saves the configuration to an io.Writer
func (c *Config) SaveConfigToWriter(writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to encode config JSON: %w", err)
	}

	return nil
}

// Validate checks if the configuration values are valid
func (c *Config) Validate() error {
	if c.SecurityToken == "" {
		return fmt.Errorf("security_token cannot be empty")
	}

	if c.Network == "" {
		return fmt.Errorf("network cannot be empty")
	}

	if c.CheckPriceInterval <= 0 {
		return fmt.Errorf("check_price_interval must be greater than 0, got: %s", c.CheckPriceInterval)
	}

	if c.MinersStateCheckInterval <= 0 {
		return fmt.Errorf("miners_state_check_interval must be greater than 0, got: %s", c.MinersStateCheckInterval)
	}

	if c.MinerDiscoveryInterval <= 0 {
		return fmt.Errorf("miner_discovery_interval must be greater than 0, got: %s", c.MinerDiscoveryInterval)
	}

	if c.APITimeout <= 0 {
		return fmt.Errorf("api_timeout must be greater than 0, got: %s", c.APITimeout)
	}

	if c.UrlFormat == "" {
		return fmt.Errorf("url_format cannot be empty")
	}

	if c.MinerTimeout <= 0 {
		return fmt.Errorf("miner_timeout must be greater than 0, got: %s", c.MinerTimeout)
	}

	if c.HealthCheckPort < 0 || c.HealthCheckPort > 65535 {
		return fmt.Errorf("health_check_port must be between 0 and 65535, got: %d", c.HealthCheckPort)
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.LogLevel] {
		return fmt.Errorf("invalid log_level: %s, must be one of: debug, info, warn, error", c.LogLevel)
	}

	// Validate log format
	validLogFormats := map[string]bool{
		"text": true,
		"json": true,
	}
	if !validLogFormats[c.LogFormat] {
		return fmt.Errorf("invalid log_format: %s, must be one of: text, json", c.LogFormat)
	}

	return nil
}

// MarshalJSON implements custom JSON marshaling to handle durations
func (c *Config) MarshalJSON() ([]byte, error) {
	type Alias Config
	return json.Marshal(&struct {
		*Alias
		CheckInterval            string `json:"check_price_interval"`
		MinersStateCheckInterval string `json:"miners_state_check_interval"`
		MinerDiscoveryInterval   string `json:"miner_discovery_interval"`
		APITimeout               string `json:"api_timeout"`
		MinerTimeout             string `json:"miner_timeout"`
	}{
		Alias:                    (*Alias)(c),
		CheckInterval:            c.CheckPriceInterval.String(),
		MinersStateCheckInterval: c.MinersStateCheckInterval.String(),
		MinerDiscoveryInterval:   c.MinerDiscoveryInterval.String(),
		APITimeout:               c.APITimeout.String(),
		MinerTimeout:             c.MinerTimeout.String(),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling to handle durations
func (c *Config) UnmarshalJSON(data []byte) error {
	type Alias Config
	aux := &struct {
		*Alias
		CheckPriceInterval       string `json:"check_price_interval"`
		MinersStateCheckInterval string `json:"miners_state_check_interval"`
		MinerDiscoveryInterval   string `json:"miner_discovery_interval"`
		APITimeout               string `json:"api_timeout"`
		MinerTimeout             string `json:"miner_timeout"`
		UrlFormat                string `json:"url_format"`
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var err error
	if aux.CheckPriceInterval != "" {
		if c.CheckPriceInterval, err = time.ParseDuration(aux.CheckPriceInterval); err != nil {
			return fmt.Errorf("invalid check_price_interval: %w", err)
		}
	}

	if aux.MinersStateCheckInterval != "" {
		if c.MinersStateCheckInterval, err = time.ParseDuration(aux.MinersStateCheckInterval); err != nil {
			return fmt.Errorf("invalid miners_state_check_interval: %w", err)
		}
	}

	if aux.MinerDiscoveryInterval != "" {
		if c.MinerDiscoveryInterval, err = time.ParseDuration(aux.MinerDiscoveryInterval); err != nil {
			return fmt.Errorf("invalid miner_discovery_interval: %w", err)
		}
	}

	if aux.APITimeout != "" {
		if c.APITimeout, err = time.ParseDuration(aux.APITimeout); err != nil {
			return fmt.Errorf("invalid api_timeout: %w", err)
		}
	}

	if aux.MinerTimeout != "" {
		if c.MinerTimeout, err = time.ParseDuration(aux.MinerTimeout); err != nil {
			return fmt.Errorf("invalid miner_timeout: %w", err)
		}
	}

	if aux.UrlFormat != "" {
		c.UrlFormat = aux.UrlFormat
	}

	return nil
}

// String returns a string representation of the config
func (c *Config) String() string {
	data, _ := json.MarshalIndent(c, "", "  ")
	return string(data)
}
