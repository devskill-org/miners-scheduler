-- Migration: Add battery_charge_from_pv and battery_charge_from_grid columns
-- This migration splits battery_charge into two components:
-- 1. battery_charge_from_pv: charge power based on current solar forecast
-- 2. battery_charge_from_grid: charge power based on zero solar forecast (grid-only scenario)

-- Add new columns with default values
ALTER TABLE mpc_decisions 
ADD COLUMN IF NOT EXISTS battery_charge_from_pv NUMERIC NOT NULL DEFAULT 0;

ALTER TABLE mpc_decisions 
ADD COLUMN IF NOT EXISTS battery_charge_from_grid NUMERIC NOT NULL DEFAULT 0;

-- For existing rows, copy battery_charge to battery_charge_from_pv
-- (assuming existing data represents PV-based charging)
UPDATE mpc_decisions 
SET battery_charge_from_pv = battery_charge, 
    battery_charge_from_grid = 0
WHERE battery_charge_from_pv = 0 AND battery_charge_from_grid = 0;