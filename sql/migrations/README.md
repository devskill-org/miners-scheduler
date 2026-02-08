# SQL Migrations

This directory contains SQL migration scripts for the EMS database schema.

## Migration Files

### 001_add_battery_charge_split.sql

This migration adds support for splitting battery charge into two components:

- `battery_charge_from_pv`: Battery charging power based on current solar forecast (kW)
- `battery_charge_from_grid`: Battery charging power based on zero solar forecast/grid-only scenario (kW)

This split allows the MPC optimizer to:
1. Run optimization with full solar forecast to determine PV-based charging
2. Run optimization without solar (grid-only) to determine grid-based charging
3. Use the results to intelligently decide between:
   - **Mode 2** (Self-use): Charge from PV surplus only when `BatteryChargeFromGrid == 0`
   - **Mode 4** (Command charging): Charge from PV and grid when `BatteryChargeFromGrid > 0`

The `battery_charge` column is kept for backward compatibility but is now deprecated.

## How to Apply Migrations

### PostgreSQL

To apply migrations to your PostgreSQL database, run:

```bash
psql -d your_database_name -f migrations/001_add_battery_charge_split.sql
```

Or connect to your database and run:

```sql
\i migrations/001_add_battery_charge_split.sql
```

### Verification

After applying the migration, verify the new columns exist:

```sql
\d mpc_decisions
```

You should see:
- `battery_charge_from_pv` (NUMERIC NOT NULL DEFAULT 0)
- `battery_charge_from_grid` (NUMERIC NOT NULL DEFAULT 0)

## Migration Safety

- All migrations use `ADD COLUMN IF NOT EXISTS` to be idempotent
- Default values are provided for existing rows
- The migration automatically updates existing rows to preserve data integrity
- The old `battery_charge` column is retained for backward compatibility

## Rollback

If you need to rollback the migration (not recommended in production):

```sql
ALTER TABLE mpc_decisions DROP COLUMN IF EXISTS battery_charge_from_pv;
ALTER TABLE mpc_decisions DROP COLUMN IF EXISTS battery_charge_from_grid;
```

Note: This will lose the split charge information, but the original `battery_charge` column will still contain the data.