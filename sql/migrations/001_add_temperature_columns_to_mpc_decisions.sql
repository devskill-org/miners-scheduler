-- Migration: Add temperature and battery preheat columns to mpc_decisions table
-- Date: 2024-01-XX
-- Description: Adds battery_avg_cell_temp, air_temperature, and battery_preheat_active columns
--              to support battery temperature forecasting and preheating cost optimization

-- Add battery_avg_cell_temp column (forecasted battery average cell temperature)
ALTER TABLE mpc_decisions 
ADD COLUMN IF NOT EXISTS battery_avg_cell_temp NUMERIC;

-- Add air_temperature column (forecasted air temperature)
ALTER TABLE mpc_decisions 
ADD COLUMN IF NOT EXISTS air_temperature NUMERIC;

-- Add battery_preheat_active column (whether battery preheating is active)
ALTER TABLE mpc_decisions 
ADD COLUMN IF NOT EXISTS battery_preheat_active BOOLEAN;

-- Add comments to document the new columns
COMMENT ON COLUMN mpc_decisions.battery_avg_cell_temp IS 'Forecasted battery average cell temperature in °C for this time slot';
COMMENT ON COLUMN mpc_decisions.air_temperature IS 'Forecasted air temperature in °C for this time slot';
COMMENT ON COLUMN mpc_decisions.battery_preheat_active IS 'Whether battery preheating is active during this time slot (true when charging below threshold temperature)';