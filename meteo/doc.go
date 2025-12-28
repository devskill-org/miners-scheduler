// Package meteo provides a Go client library for the MET Norway Location Forecast API.
//
// This package allows you to retrieve weather forecast data from the Norwegian
// Meteorological Institute's location forecast service. It supports all three
// endpoints: compact, complete, and classic formats.
//
// Basic Usage:
//
//	client := meteo.NewClient("YourApp/1.0 (your-email@example.com)")
//
//	location := meteo.Location{
//		Latitude:  59.9139,  // Oslo
//		Longitude: 10.7522,
//	}
//
//	params := meteo.QueryParams{
//		Location: location,
//	}
//
//	forecast, err := client.GetCompact(params)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Use forecast data...
//	for _, step := range forecast.Properties.Timeseries {
//		fmt.Printf("Time: %v, Temperature: %.1f°C\n",
//			step.Time,
//			*step.Data.Instant.Details.AirTemperature)
//	}
//
// API Endpoints:
//
// - GetCompact(): Returns a compact forecast with essential weather parameters
// - GetComplete(): Returns the complete forecast with all available parameters
// - GetClassic(): Returns forecast data in the classic format
//
// The client automatically handles JSON deserialization according to the MET API
// specification and includes proper error handling for HTTP and validation errors.
//
// For more information about the API, visit: https://api.met.no/weatherapi/locationforecast/2.0/documentation
package meteo
