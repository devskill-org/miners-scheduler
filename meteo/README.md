# Meteo - Go Client for MET Norway Location Forecast API

A Go client library for the [MET Norway Location Forecast API](https://api.met.no/weatherapi/locationforecast/2.0/documentation). This package provides easy access to weather forecast data from the Norwegian Meteorological Institute.

## Features

- Support for all three API endpoints: `compact`, `complete`, and `classic`
- Full JSON deserialization according to the MET API specification
- Type-safe weather symbol constants
- Utility functions for common operations
- Comprehensive error handling
- Location validation
- Time-based weather data filtering

## Installation

```bash
go get github.com/devskill-org/energy-management-system/meteo
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/devskill-org/energy-management-system/meteo"
)

func main() {
    // Create client with proper User-Agent (required by MET API)
    client := meteo.NewClient("MyApp/1.0 (contact@example.com)")
    
    // Define location (Oslo, Norway)
    location := meteo.Location{
        Latitude:  59.9139,
        Longitude: 10.7522,
        Altitude:  meteo.IntPtr(14), // Optional
    }
    
    // Validate location
    if err := meteo.ValidateLocation(location); err != nil {
        log.Fatal(err)
    }
    
    // Get forecast
    params := meteo.QueryParams{Location: location}
    forecast, err := client.GetCompact(params)
    if err != nil {
        log.Fatal(err)
    }
    
    // Get current weather
    current := forecast.GetCurrentWeather()
    if current != nil {
        if temp := current.GetTemperature(); temp != nil {
            fmt.Printf("Current temperature: %.1f°C\n", *temp)
        }
        
        if symbol := current.GetSymbolCode(); symbol != nil {
            fmt.Printf("Weather: %s\n", *symbol)
        }
    }
}
```

## API Endpoints

### GetCompact()
Returns essential weather parameters in a compact format. This is the most commonly used endpoint.

```go
forecast, err := client.GetCompact(params)
```

### GetComplete()
Returns all available weather parameters. Use this when you need detailed meteorological data.

```go
forecast, err := client.GetComplete(params)
```

### GetClassic()
Returns weather data in the classic format for backward compatibility.

```go
forecast, err := client.GetClassic(params)
```

## Data Types

### Core Types

- **`METJSONForecast`** - Root forecast response
- **`Forecast`** - Main forecast data with metadata and timeseries
- **`ForecastTimeStep`** - Individual forecast for a specific time
- **`ForecastTimeInstant`** - Instant weather parameters
- **`ForecastTimePeriod`** - Period-based forecast data
- **`WeatherSymbol`** - Type-safe weather condition identifiers

### Location

```go
type Location struct {
    Latitude  float64 `json:"lat"`
    Longitude float64 `json:"lon"`
    Altitude  *int    `json:"altitude,omitempty"`
}
```

## Utility Functions

### Time-based Filtering

```go
// Get current weather conditions
current := forecast.GetCurrentWeather()

// Get weather for a specific time
weather := forecast.GetWeatherAtTime(time.Now().Add(6 * time.Hour))

// Get all forecasts for today
today := forecast.GetDayForecast(time.Now())

// Get forecasts for a time period
period := forecast.GetForecastForPeriod(start, end)
```

### Weather Data Access

```go
// Check for precipitation
hasPrecipitation := timeStep.HasPrecipitation()

// Get weather parameters
temperature := timeStep.GetTemperature()
windSpeed := timeStep.GetWindSpeed()
humidity := timeStep.GetHumidity()
symbolCode := timeStep.GetSymbolCode()
```

### Weather Symbol Methods

```go
symbol := meteo.RainShowersDay

isDay := symbol.IsDay()                    // true
isNight := symbol.IsNight()               // false
isPolarTwilight := symbol.IsPolarTwilight() // false
hasThunder := symbol.HasThunder()         // false
```

## Weather Symbols

The package includes all weather symbols as type-safe constants:

```go
meteo.ClearSkyDay
meteo.PartlyCloudyNight
meteo.RainAndThunder
meteo.HeavySnow
meteo.Fog
// ... and many more
```

## Error Handling

The package provides specific error types for different failure scenarios:

```go
forecast, err := client.GetCompact(params)
if err != nil {
    switch e := err.(type) {
    case *meteo.APIError:
        fmt.Printf("API error %d: %s\n", e.StatusCode, e.Message)
    case *meteo.ValidationError:
        fmt.Printf("Validation error: %s\n", e.Message)
    case *meteo.NetworkError:
        fmt.Printf("Network error: %v\n", e.Err)
    default:
        fmt.Printf("Unknown error: %v\n", err)
    }
}
```

## Configuration

### Custom HTTP Client

```go
httpClient := &http.Client{
    Timeout: 10 * time.Second,
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    },
}

client := meteo.NewClientWithHTTPClient(httpClient, "MyApp/1.0")
```

### Custom Base URL

```go
client := meteo.NewClient("MyApp/1.0")
client.SetBaseURL("https://custom-api.example.com/weatherapi/locationforecast/2.0")
```

## Examples

### Check for Rain in the Next 24 Hours

```go
forecast, err := client.GetCompact(params)
if err != nil {
    return err
}

now := time.Now()
next24h := forecast.GetForecastForPeriod(now, now.Add(24*time.Hour))

for _, step := range next24h {
    if step.HasPrecipitation() {
        fmt.Printf("Rain expected at %s\n", step.Time.Format("15:04"))
    }
}
```

### Get Daily Temperature Summary

```go
dayForecast := forecast.GetDayForecast(time.Now())
var temps []float64

for _, step := range dayForecast {
    if temp := step.GetTemperature(); temp != nil {
        temps = append(temps, *temp)
    }
}

if len(temps) > 0 {
    min := temps[0]
    max := temps[0]
    for _, t := range temps[1:] {
        if t < min { min = t }
        if t > max { max = t }
    }
    fmt.Printf("Temperature range: %.1f°C to %.1f°C\n", min, max)
}
```

## Requirements

- Go 1.16 or later
- Valid User-Agent string (required by MET API terms of service)

## User-Agent Requirements

The MET Norway API requires a proper User-Agent header that identifies your application. The format should be:

```
ApplicationName/Version (contact-email@example.com)
```

Example:
```go
client := meteo.NewClient("WeatherBot/2.1 (weather-support@mycompany.com)")
```

## Rate Limiting

The MET API has rate limiting in place. Be respectful and cache responses when possible. The API returns standard HTTP rate limiting headers.

## License

This package is released under the MIT License. See the LICENSE file for details.

## API Documentation

For complete API documentation, visit:
https://api.met.no/weatherapi/locationforecast/2.0/documentation

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.