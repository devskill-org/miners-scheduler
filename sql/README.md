# SQL Database Setup

This directory contains SQL schema files for the EMS (Energy Management System) application.

## Prerequisites

- PostgreSQL database (version 9.5+)
- The `gen_random_uuid()` function is available in PostgreSQL 13+ by default
- For PostgreSQL versions < 13, you may need to enable the `pgcrypto` extension:
  ```sql
  CREATE EXTENSION IF NOT EXISTS pgcrypto;
  ```

## Database Tables

### metrics.sql

Contains the schema for storing energy metrics and measurements from the system.

**Table:** `metrics`

Stores time-series data including:
- PV power generation
- Grid import/export
- Battery charge/discharge
- Load consumption
- Energy costs
- Battery state of charge
- Weather conditions

### mpc_decisions.sql

Contains the schema for persisting Model Predictive Control (MPC) optimization decisions.

**Table:** `mpc_decisions`

Stores MPC optimization decisions with timestamp as the PRIMARY KEY (only one decision per timestamp):
- Decision timestamps (PRIMARY KEY)
- Horizon hours
- Battery charge/discharge commands
- Grid import/export commands
- Expected profits
- Forecast data (prices, solar, load, weather)

## Setup Instructions

1. Create a PostgreSQL database for the EMS application:
   ```bash
   createdb ems
   ```

2. Set the database connection string in your EMS configuration:
   ```json
   {
     "postgres_conn_string": "postgres://username:password@localhost:5432/ems?sslmode=disable"
   }
   ```

3. Run the SQL schema files to create the tables:
   ```bash
   psql -d ems -f sql/metrics.sql
   psql -d ems -f sql/mpc_decisions.sql
   ```

   Or from within `psql`:
   ```sql
   \i sql/metrics.sql
   \i sql/mpc_decisions.sql
   ```

## MPC Decisions Persistence

The MPC controller automatically:

1. **Saves decisions** to the database after each optimization run
   - Deletes all existing decisions with timestamp >= minimum timestamp of new decisions
   - Inserts/updates new decisions using UPSERT (ON CONFLICT DO UPDATE)
   - Each decision includes forecast data and expected outcomes
   - Timestamp is the PRIMARY KEY, ensuring only one decision per time slot

2. **Loads decisions** on scheduler startup
   - Loads only future decisions (timestamp >= current time)
   - Orders decisions by timestamp
   - This allows the system to continue executing decisions after a restart

## Data Retention

Consider implementing data retention policies:

```sql
-- Example: Delete MPC decisions older than current time (past decisions)
DELETE FROM mpc_decisions WHERE timestamp < EXTRACT(EPOCH FROM NOW());

-- Example: Delete metrics older than 1 year
DELETE FROM metrics WHERE timestamp < NOW() - INTERVAL '1 year';
```

You can set up a periodic cleanup job using `pg_cron` or an external scheduler.

## Indexes

The `mpc_decisions` table uses timestamp as PRIMARY KEY which provides:

- **Fast lookups** by timestamp (primary key constraint creates implicit index)
- **Uniqueness guarantee** - only one decision per timestamp
- **Efficient range queries** for loading future decisions

The `metrics` table includes BRIN indexes for time-series efficiency.

## Monitoring

Query examples for monitoring:

```sql
-- Check future MPC decisions (from now onwards)
SELECT 
    timestamp,
    hour,
    battery_charge,
    battery_discharge,
    grid_import,
    grid_export,
    profit
FROM mpc_decisions
WHERE timestamp >= EXTRACT(EPOCH FROM NOW())
ORDER BY timestamp ASC;

-- Check all MPC decisions
SELECT 
    timestamp,
    hour,
    battery_charge,
    battery_discharge,
    grid_import,
    grid_export,
    profit
FROM mpc_decisions
ORDER BY timestamp ASC;

-- Check recent energy flow metrics
SELECT 
    timestamp,
    pv_total_power,
    grid_import_power,
    grid_export_power,
    battery_soc
FROM metrics
WHERE metric_name = 'energy_flow'
  AND timestamp > NOW() - INTERVAL '24 hours'
ORDER BY timestamp DESC;
```

## Troubleshooting

**Issue:** UUID generation fails with error about `gen_random_uuid()`

**Solution:** Enable the pgcrypto extension:
```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;
```

**Issue:** Connection refused to PostgreSQL

**Solution:** Check that PostgreSQL is running and the connection string is correct:
```bash
psql -d ems  # Test connection
```

**Issue:** Permission denied when creating tables

**Solution:** Ensure your database user has CREATE privileges:
```sql
GRANT CREATE ON DATABASE ems TO your_username;
```
