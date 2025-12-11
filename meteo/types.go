package meteo

import "time"

// WeatherSymbol represents weather condition identifiers
type WeatherSymbol string

const (
	ClearSkyDay                              WeatherSymbol = "clearsky_day"
	ClearSkyNight                            WeatherSymbol = "clearsky_night"
	ClearSkyPolarTwilight                    WeatherSymbol = "clearsky_polartwilight"
	FairDay                                  WeatherSymbol = "fair_day"
	FairNight                                WeatherSymbol = "fair_night"
	FairPolarTwilight                        WeatherSymbol = "fair_polartwilight"
	LightSnowShowersAndThunderDay            WeatherSymbol = "lightssnowshowersandthunder_day"
	LightSnowShowersAndThunderNight          WeatherSymbol = "lightssnowshowersandthunder_night"
	LightSnowShowersAndThunderPolarTwilight  WeatherSymbol = "lightssnowshowersandthunder_polartwilight"
	LightSnowShowersDay                      WeatherSymbol = "lightsnowshowers_day"
	LightSnowShowersNight                    WeatherSymbol = "lightsnowshowers_night"
	LightSnowShowersPolarTwilight            WeatherSymbol = "lightsnowshowers_polartwilight"
	HeavyRainAndThunder                      WeatherSymbol = "heavyrainandthunder"
	HeavySnowAndThunder                      WeatherSymbol = "heavysnowandthunder"
	RainAndThunder                           WeatherSymbol = "rainandthunder"
	HeavySleetShowersAndThunderDay           WeatherSymbol = "heavysleetshowersandthunder_day"
	HeavySleetShowersAndThunderNight         WeatherSymbol = "heavysleetshowersandthunder_night"
	HeavySleetShowersAndThunderPolarTwilight WeatherSymbol = "heavysleetshowersandthunder_polartwilight"
	HeavySnow                                WeatherSymbol = "heavysnow"
	HeavyRainShowersDay                      WeatherSymbol = "heavyrainshowers_day"
	HeavyRainShowersNight                    WeatherSymbol = "heavyrainshowers_night"
	HeavyRainShowersPolarTwilight            WeatherSymbol = "heavyrainshowers_polartwilight"
	LightSleet                               WeatherSymbol = "lightsleet"
	HeavyRain                                WeatherSymbol = "heavyrain"
	LightRainShowersDay                      WeatherSymbol = "lightrainshowers_day"
	LightRainShowersNight                    WeatherSymbol = "lightrainshowers_night"
	LightRainShowersPolarTwilight            WeatherSymbol = "lightrainshowers_polartwilight"
	HeavySleetShowersDay                     WeatherSymbol = "heavysleetshowers_day"
	HeavySleetShowersNight                   WeatherSymbol = "heavysleetshowers_night"
	HeavySleetShowersPolarTwilight           WeatherSymbol = "heavysleetshowers_polartwilight"
	LightSleetShowersDay                     WeatherSymbol = "lightsleetshowers_day"
	LightSleetShowersNight                   WeatherSymbol = "lightsleetshowers_night"
	LightSleetShowersPolarTwilight           WeatherSymbol = "lightsleetshowers_polartwilight"
	Snow                                     WeatherSymbol = "snow"
	HeavyRainShowersAndThunderDay            WeatherSymbol = "heavyrainshowersandthunder_day"
	HeavyRainShowersAndThunderNight          WeatherSymbol = "heavyrainshowersandthunder_night"
	HeavyRainShowersAndThunderPolarTwilight  WeatherSymbol = "heavyrainshowersandthunder_polartwilight"
	SnowShowersDay                           WeatherSymbol = "snowshowers_day"
	SnowShowersNight                         WeatherSymbol = "snowshowers_night"
	SnowShowersPolarTwilight                 WeatherSymbol = "snowshowers_polartwilight"
	Fog                                      WeatherSymbol = "fog"
	SnowShowersAndThunderDay                 WeatherSymbol = "snowshowersandthunder_day"
	SnowShowersAndThunderNight               WeatherSymbol = "snowshowersandthunder_night"
	SnowShowersAndThunderPolarTwilight       WeatherSymbol = "snowshowersandthunder_polartwilight"
	LightSnowAndThunder                      WeatherSymbol = "lightsnowandthunder"
	HeavySleetAndThunder                     WeatherSymbol = "heavysleetandthunder"
	LightRain                                WeatherSymbol = "lightrain"
	RainShowersAndThunderDay                 WeatherSymbol = "rainshowersandthunder_day"
	RainShowersAndThunderNight               WeatherSymbol = "rainshowersandthunder_night"
	RainShowersAndThunderPolarTwilight       WeatherSymbol = "rainshowersandthunder_polartwilight"
	Rain                                     WeatherSymbol = "rain"
	LightSnow                                WeatherSymbol = "lightsnow"
	LightRainShowersAndThunderDay            WeatherSymbol = "lightrainshowersandthunder_day"
	LightRainShowersAndThunderNight          WeatherSymbol = "lightrainshowersandthunder_night"
	LightRainShowersAndThunderPolarTwilight  WeatherSymbol = "lightrainshowersandthunder_polartwilight"
	HeavySleet                               WeatherSymbol = "heavysleet"
	SleetAndThunder                          WeatherSymbol = "sleetandthunder"
	LightRainAndThunder                      WeatherSymbol = "lightrainandthunder"
	Sleet                                    WeatherSymbol = "sleet"
	LightSleetShowersAndThunderDay           WeatherSymbol = "lightssleetshowersandthunder_day"
	LightSleetShowersAndThunderNight         WeatherSymbol = "lightssleetshowersandthunder_night"
	LightSleetShowersAndThunderPolarTwilight WeatherSymbol = "lightssleetshowersandthunder_polartwilight"
	LightSleetAndThunder                     WeatherSymbol = "lightsleetandthunder"
	PartlyCloudyDay                          WeatherSymbol = "partlycloudy_day"
	PartlyCloudyNight                        WeatherSymbol = "partlycloudy_night"
	PartlyCloudyPolarTwilight                WeatherSymbol = "partlycloudy_polartwilight"
	SleetShowersAndThunderDay                WeatherSymbol = "sleetshowersandthunder_day"
	SleetShowersAndThunderNight              WeatherSymbol = "sleetshowersandthunder_night"
	SleetShowersAndThunderPolarTwilight      WeatherSymbol = "sleetshowersandthunder_polartwilight"
	RainShowersDay                           WeatherSymbol = "rainshowers_day"
	RainShowersNight                         WeatherSymbol = "rainshowers_night"
	RainShowersPolarTwilight                 WeatherSymbol = "rainshowers_polartwilight"
	SnowAndThunder                           WeatherSymbol = "snowandthunder"
	SleetShowersDay                          WeatherSymbol = "sleetshowers_day"
	SleetShowersNight                        WeatherSymbol = "sleetshowers_night"
	SleetShowersPolarTwilight                WeatherSymbol = "sleetshowers_polartwilight"
	Cloudy                                   WeatherSymbol = "cloudy"
	HeavySnowShowersAndThunderDay            WeatherSymbol = "heavysnowshowersandthunder_day"
	HeavySnowShowersAndThunderNight          WeatherSymbol = "heavysnowshowersandthunder_night"
	HeavySnowShowersAndThunderPolarTwilight  WeatherSymbol = "heavysnowshowersandthunder_polartwilight"
	HeavySnowShowersDay                      WeatherSymbol = "heavysnowshowers_day"
	HeavySnowShowersNight                    WeatherSymbol = "heavysnowshowers_night"
	HeavySnowShowersPolarTwilight            WeatherSymbol = "heavysnowshowers_polartwilight"
)

// PointGeometry represents a GeoJSON point geometry
type PointGeometry struct {
	Type        string    `json:"type"`        // Should be "Point"
	Coordinates []float64 `json:"coordinates"` // [longitude, latitude, altitude]
}

// ForecastUnits contains the units for all forecast values
type ForecastUnits struct {
	AirPressureAtSeaLevel       *string `json:"air_pressure_at_sea_level,omitempty"`
	AirTemperature              *string `json:"air_temperature,omitempty"`
	AirTemperatureMax           *string `json:"air_temperature_max,omitempty"`
	AirTemperatureMin           *string `json:"air_temperature_min,omitempty"`
	CloudAreaFraction           *string `json:"cloud_area_fraction,omitempty"`
	CloudAreaFractionHigh       *string `json:"cloud_area_fraction_high,omitempty"`
	CloudAreaFractionLow        *string `json:"cloud_area_fraction_low,omitempty"`
	CloudAreaFractionMedium     *string `json:"cloud_area_fraction_medium,omitempty"`
	DewPointTemperature         *string `json:"dew_point_temperature,omitempty"`
	FogAreaFraction             *string `json:"fog_area_fraction,omitempty"`
	PrecipitationAmount         *string `json:"precipitation_amount,omitempty"`
	PrecipitationAmountMax      *string `json:"precipitation_amount_max,omitempty"`
	PrecipitationAmountMin      *string `json:"precipitation_amount_min,omitempty"`
	ProbabilityOfPrecipitation  *string `json:"probability_of_precipitation,omitempty"`
	ProbabilityOfThunder        *string `json:"probability_of_thunder,omitempty"`
	RelativeHumidity            *string `json:"relative_humidity,omitempty"`
	UltravioletIndexClearSkyMax *string `json:"ultraviolet_index_clear_sky_max,omitempty"`
	WindFromDirection           *string `json:"wind_from_direction,omitempty"`
	WindSpeed                   *string `json:"wind_speed,omitempty"`
	WindSpeedOfGust             *string `json:"wind_speed_of_gust,omitempty"`
}

// ForecastMeta contains metadata for the forecast
type ForecastMeta struct {
	UpdatedAt time.Time     `json:"updated_at"`
	Units     ForecastUnits `json:"units"`
}

// ForecastTimeInstant contains weather parameters valid for a specific point in time
type ForecastTimeInstant struct {
	AirPressureAtSeaLevel   *float64 `json:"air_pressure_at_sea_level,omitempty"`
	AirTemperature          *float64 `json:"air_temperature,omitempty"`
	CloudAreaFraction       *float64 `json:"cloud_area_fraction,omitempty"`
	CloudAreaFractionHigh   *float64 `json:"cloud_area_fraction_high,omitempty"`
	CloudAreaFractionLow    *float64 `json:"cloud_area_fraction_low,omitempty"`
	CloudAreaFractionMedium *float64 `json:"cloud_area_fraction_medium,omitempty"`
	DewPointTemperature     *float64 `json:"dew_point_temperature,omitempty"`
	FogAreaFraction         *float64 `json:"fog_area_fraction,omitempty"`
	RelativeHumidity        *float64 `json:"relative_humidity,omitempty"`
	WindFromDirection       *float64 `json:"wind_from_direction,omitempty"`
	WindSpeed               *float64 `json:"wind_speed,omitempty"`
	WindSpeedOfGust         *float64 `json:"wind_speed_of_gust,omitempty"`
}

// ForecastTimePeriod contains weather parameters valid for a specified time period
type ForecastTimePeriod struct {
	AirTemperatureMax           *float64 `json:"air_temperature_max,omitempty"`
	AirTemperatureMin           *float64 `json:"air_temperature_min,omitempty"`
	PrecipitationAmount         *float64 `json:"precipitation_amount,omitempty"`
	PrecipitationAmountMax      *float64 `json:"precipitation_amount_max,omitempty"`
	PrecipitationAmountMin      *float64 `json:"precipitation_amount_min,omitempty"`
	ProbabilityOfPrecipitation  *float64 `json:"probability_of_precipitation,omitempty"`
	ProbabilityOfThunder        *float64 `json:"probability_of_thunder,omitempty"`
	UltravioletIndexClearSkyMax *float64 `json:"ultraviolet_index_clear_sky_max,omitempty"`
}

// ForecastSummary contains a summary of weather conditions
type ForecastSummary struct {
	SymbolCode WeatherSymbol `json:"symbol_code"`
}

// ForecastPeriodData contains forecast data for a specific period
type ForecastPeriodData struct {
	Summary *ForecastSummary    `json:"summary,omitempty"`
	Details *ForecastTimePeriod `json:"details,omitempty"`
}

// ForecastInstantData contains instant forecast data
type ForecastInstantData struct {
	Details *ForecastTimeInstant `json:"details,omitempty"`
}

// ForecastTimeStepData contains forecast data for a specific time step
type ForecastTimeStepData struct {
	Instant     *ForecastInstantData `json:"instant,omitempty"`
	Next1Hours  *ForecastPeriodData  `json:"next_1_hours,omitempty"`
	Next6Hours  *ForecastPeriodData  `json:"next_6_hours,omitempty"`
	Next12Hours *ForecastPeriodData  `json:"next_12_hours,omitempty"`
}

// ForecastTimeStep represents a forecast for a specific time step
type ForecastTimeStep struct {
	Time time.Time             `json:"time"`
	Data *ForecastTimeStepData `json:"data,omitempty"`
}

// Forecast contains the main forecast data
type Forecast struct {
	Meta       ForecastMeta       `json:"meta"`
	Timeseries []ForecastTimeStep `json:"timeseries"`
}

// METJSONForecast represents the root forecast response
type METJSONForecast struct {
	Type       string         `json:"type"` // Should be "Feature"
	Geometry   *PointGeometry `json:"geometry,omitempty"`
	Properties *Forecast      `json:"properties,omitempty"`
}

// Location represents coordinates for a forecast request
type Location struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
	Altitude  *int    `json:"altitude,omitempty"`
}

// QueryParams represents query parameters for forecast requests
type QueryParams struct {
	Location Location `json:"location"`
	Format   string   `json:"format,omitempty"` // Default is JSON
}
