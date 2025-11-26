CREATE TABLE metrics (
    timestamp TIMESTAMPTZ NOT NULL,
    device_id INTEGER,
    metric_name VARCHAR(100),
    value NUMERIC,
    PRIMARY KEY (timestamp, device_id, metric_name)
);

-- BRIN index for time series efficiency
CREATE INDEX idx_metrics_time ON metrics USING BRIN (timestamp);
