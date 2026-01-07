CREATE TABLE mpc_decisions (
    timestamp BIGINT PRIMARY KEY,
    hour INTEGER NOT NULL,
    battery_charge NUMERIC NOT NULL,
    battery_discharge NUMERIC NOT NULL,
    grid_import NUMERIC NOT NULL,
    grid_export NUMERIC NOT NULL,
    battery_soc NUMERIC NOT NULL,
    profit NUMERIC NOT NULL,
    import_price NUMERIC NOT NULL,
    export_price NUMERIC NOT NULL,
    solar_forecast NUMERIC NOT NULL,
    load_forecast NUMERIC NOT NULL,
    cloud_coverage NUMERIC,
    weather_symbol VARCHAR(100)
);

-- Column descriptions:
-- timestamp: Unix timestamp when this time slot begins (PRIMARY KEY)
-- hour: Hour index in the optimization horizon (0-based)
-- battery_charge: Battery charging power in kW (positive = charging)
-- battery_discharge: Battery discharging power in kW (positive = discharging)
-- grid_import: Grid import power in kW (positive = importing)
-- grid_export: Grid export power in kW (positive = exporting)
-- battery_soc: Battery state of charge as fraction (0-1)
-- profit: Expected profit for this hour in EUR
-- import_price: Electricity import price in EUR/kWh
-- export_price: Electricity export price in EUR/kWh
-- solar_forecast: Forecasted solar generation in kW (average for the hour)
-- load_forecast: Forecasted load consumption in kW (average for the hour)
-- cloud_coverage: Cloud coverage percentage (0-100)
-- weather_symbol: Weather condition symbol code
