CREATE TABLE metrics (
    timestamp TIMESTAMPTZ NOT NULL,
    device_id INTEGER,
    metric_name VARCHAR(100),
    pv_total_power NUMERIC,
    cloud_coverage NUMERIC,
    PRIMARY KEY (timestamp, device_id, metric_name)
);

-- BRIN index for time series efficiency
CREATE INDEX idx_metrics_time ON metrics USING BRIN (timestamp);

-- Supported metric_name values:
-- - 'pv_total_power': Total PV power in kWh for the integration period

-- cloud_coverage column stores the cloud area fraction percentage (0-100) from weather API

-- Migration for existing tables:
-- ALTER TABLE metrics ADD COLUMN cloud_coverage NUMERIC;
-- ALTER TABLE metrics RENAME COLUMN value TO pv_total_power;
