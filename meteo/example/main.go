// Package main provides an example of using the meteo client to fetch weather forecasts.
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/devskill-org/ems/meteo"
)

func main() {
	// Create a client with proper User-Agent (required by MET API)
	client := meteo.NewClient("MyApp/1.0 (username@example.com)")

	// Define location (Oslo, Norway)
	location := meteo.Location{
		Latitude:  56.9496,
		Longitude: 24.1052,
		Altitude:  meteo.IntPtr(14), // Optional altitude in meters
	}

	// Validate location before making request
	if err := meteo.ValidateLocation(location); err != nil {
		log.Fatalf("Invalid location: %v", err)
	}

	fmt.Printf("Getting weather forecast for Oslo (%.4f, %.4f)\n\n",
		location.Latitude, location.Longitude)

	// Create query parameters
	params := meteo.QueryParams{
		Location: location,
	}

	// Get compact forecast (most commonly used)
	forecast, err := client.GetCompact(params)
	if err != nil {
		// Handle different error types
		switch e := err.(type) {
		case *meteo.APIError:
			log.Fatalf("API error %d: %s", e.StatusCode, e.Message)
		case *meteo.ValidationError:
			log.Fatalf("Validation error: %s", e.Message)
		case *meteo.NetworkError:
			log.Fatalf("Network error: %v", e.Err)
		default:
			log.Fatalf("Unknown error: %v", err)
		}
	}

	// Display basic forecast information
	fmt.Printf("Forecast updated: %s\n",
		forecast.Properties.Meta.UpdatedAt.Format("2006-01-02 15:04:05 UTC"))

	if forecast.Properties.Meta.Units.AirTemperature != nil {
		fmt.Printf("Temperature unit: %s\n", *forecast.Properties.Meta.Units.AirTemperature)
	}

	fmt.Printf("Total forecast steps: %d\n\n", len(forecast.Properties.Timeseries))

	// Get current weather
	current := forecast.GetCurrentWeather()
	if current != nil {
		fmt.Println("=== CURRENT WEATHER ===")
		fmt.Printf("Time: %s\n", current.Time.Format("2006-01-02 15:04:05"))

		if temp := current.GetTemperature(); temp != nil {
			fmt.Printf("Temperature: %.1f°C\n", *temp)
		}

		if humidity := current.GetHumidity(); humidity != nil {
			fmt.Printf("Humidity: %.1f%%\n", *humidity)
		}

		if windSpeed := current.GetWindSpeed(); windSpeed != nil {
			fmt.Printf("Wind speed: %.1f m/s\n", *windSpeed)
		}

		if windDir := current.GetWindDirection(); windDir != nil {
			fmt.Printf("Wind direction: %.0f°\n", *windDir)
		}

		if clouds := current.GetCloudCoverage(); clouds != nil {
			fmt.Printf("Cloud coverage: %.1f%%\n", *clouds)
		}

		if symbol := current.GetSymbolCode(); symbol != nil {
			fmt.Printf("Weather condition: %s", *symbol)
			if symbol.IsDay() {
				fmt.Print(" (day)")
			} else if symbol.IsNight() {
				fmt.Print(" (night)")
			} else if symbol.IsPolarTwilight() {
				fmt.Print(" (polar twilight)")
			}

			if symbol.HasThunder() {
				fmt.Print(" [THUNDER WARNING]")
			}
			fmt.Println()
		}

		if current.HasPrecipitation() {
			fmt.Println("☔ Precipitation expected")
		} else {
			fmt.Println("☀️ No precipitation expected")
		}
	}

	fmt.Println()

	// Get today's forecast
	today := time.Now()
	dayForecast := forecast.GetDayForecast(today)

	if len(dayForecast) > 0 {
		fmt.Println("=== TODAY'S FORECAST ===")
		for _, step := range dayForecast {
			fmt.Printf("%s", step.Time.Format("15:04"))

			if temp := step.GetTemperature(); temp != nil {
				fmt.Printf(" | %.1f°C", *temp)
			}

			if step.HasPrecipitation() {
				fmt.Printf(" | ☔")
			} else {
				fmt.Printf(" | ☀️")
			}

			if symbol := step.GetSymbolCode(); symbol != nil {
				fmt.Printf(" | %s", *symbol)
			}

			fmt.Println()
		}
	}

	fmt.Println()

	// Check for precipitation in next 24 hours
	now := time.Now()
	next24h := forecast.GetForecastForPeriod(now, now.Add(24*time.Hour))

	precipitationTimes := []string{}
	for _, step := range next24h {
		if step.HasPrecipitation() {
			precipitationTimes = append(precipitationTimes,
				step.Time.Format("Mon 15:04"))
		}
	}

	if len(precipitationTimes) > 0 {
		fmt.Println("=== PRECIPITATION ALERT ===")
		fmt.Println("Rain expected at:")
		for _, timeStr := range precipitationTimes {
			fmt.Printf("  - %s\n", timeStr)
		}
	} else {
		fmt.Println("=== NO PRECIPITATION ===")
		fmt.Println("No rain expected in the next 24 hours")
	}

	fmt.Println()

	// Temperature summary for next 24 hours
	var temps []float64
	for _, step := range next24h {
		if temp := step.GetTemperature(); temp != nil {
			temps = append(temps, *temp)
		}
	}

	if len(temps) > 0 {
		minTemp := temps[0]
		maxTemp := temps[0]
		sumTemp := 0.0

		for _, temp := range temps {
			if temp < minTemp {
				minTemp = temp
			}
			if temp > maxTemp {
				maxTemp = temp
			}
			sumTemp += temp
		}

		avgTemp := sumTemp / float64(len(temps))

		fmt.Println("=== 24-HOUR TEMPERATURE SUMMARY ===")
		fmt.Printf("Minimum: %.1f°C\n", minTemp)
		fmt.Printf("Maximum: %.1f°C\n", maxTemp)
		fmt.Printf("Average: %.1f°C\n", avgTemp)
	}

	// Get weather for tomorrow at noon
	tomorrow := now.Add(24 * time.Hour)
	tomorrowNoon := time.Date(
		tomorrow.Year(), tomorrow.Month(), tomorrow.Day(),
		12, 0, 0, 0, tomorrow.Location(),
	)

	weatherAtNoon := forecast.GetWeatherAtTime(tomorrowNoon)
	if weatherAtNoon != nil {
		fmt.Println()
		fmt.Println("=== TOMORROW AT NOON ===")

		if temp := weatherAtNoon.GetTemperature(); temp != nil {
			fmt.Printf("Temperature: %.1f°C\n", *temp)
		}

		if symbol := weatherAtNoon.GetSymbolCode(); symbol != nil {
			fmt.Printf("Condition: %s\n", *symbol)
		}

		if weatherAtNoon.HasPrecipitation() {
			fmt.Println("Precipitation: Yes")
		} else {
			fmt.Println("Precipitation: No")
		}
	}
}
