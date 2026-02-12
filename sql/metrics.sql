CREATE TABLE metrics (
    timestamp TIMESTAMPTZ NOT NULL,
    device_id INTEGER,
    metric_name VARCHAR(100),
    pv_total_power NUMERIC,
    cloud_coverage NUMERIC,
    grid_export_power NUMERIC,
    grid_import_power NUMERIC,
    battery_charge_power NUMERIC,
    battery_discharge_power NUMERIC,
    evdc_charge_power NUMERIC,
    load_power NUMERIC,
    grid_export_cost NUMERIC,
    grid_import_cost NUMERIC,
    battery_soc NUMERIC,
    battery_avg_cell_temperature NUMERIC,
    weather_symbol VARCHAR(100),
    PRIMARY KEY (timestamp, device_id, metric_name)
);

-- BRIN index for time series efficiency
CREATE INDEX idx_metrics_time ON metrics USING BRIN (timestamp);

-- Supported metric_name values:
-- - 'pv_total_power': Total PV power in kWh for the integration period
-- - 'energy_flow': Combined energy flow metrics for the integration period

-- Column descriptions:
-- pv_total_power: Total PV power in kWh
-- cloud_coverage: Cloud area fraction percentage (0-100) from weather API
-- grid_export_power: Total power exported to grid in kWh
-- grid_import_power: Total power imported from grid in kWh
-- battery_charge_power: Total power charged to battery in kWh
-- battery_discharge_power: Total power discharged from battery in kWh
-- evdc_charge_power: Total power charged to EV DC charger in kWh
-- load_power: Total power consumed by load in kWh
-- grid_export_cost: Total cost of power exported to grid in EUR
-- grid_import_cost: Total cost of power imported from grid in EUR
-- battery_soc: Battery state of charge in % (snapshot at end of period)
-- battery_avg_cell_temperature: Battery average cell temperature in °C (snapshot at end of period)
-- weather_symbol: Weather condition symbol code (e.g., 'clearsky_day', 'rain', etc.)

-- Migration for existing tables:
-- ALTER TABLE metrics ADD COLUMN grid_export_power NUMERIC;
-- ALTER TABLE metrics ADD COLUMN grid_import_power NUMERIC;
-- ALTER TABLE metrics ADD COLUMN battery_charge_power NUMERIC;
-- ALTER TABLE metrics ADD COLUMN battery_discharge_power NUMERIC;
-- ALTER TABLE metrics ADD COLUMN evdc_charge_power NUMERIC;
-- ALTER TABLE metrics ADD COLUMN load_power NUMERIC;
-- ALTER TABLE metrics ADD COLUMN grid_export_cost NUMERIC;
-- ALTER TABLE metrics ADD COLUMN grid_import_cost NUMERIC;
-- ALTER TABLE metrics ADD COLUMN battery_soc NUMERIC;
-- ALTER TABLE metrics ADD COLUMN battery_avg_cell_temperature NUMERIC;
-- ALTER TABLE metrics ADD COLUMN weather_symbol VARCHAR(100);
