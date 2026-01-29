package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/devskill-org/ems/mpc"
)

// saveMPCDecisions persists MPC decisions to the database
func (s *MinerScheduler) saveMPCDecisions(ctx context.Context, decisions []mpc.ControlDecision) error {
	if s.db == nil {
		return fmt.Errorf("database connection not available")
	}

	if len(decisions) == 0 {
		return nil
	}

	// Use first decision timestamp as minimum
	// Decisions are ordered by timestamp because:
	// 1. MPC forecast is built from a map and explicitly sorted by hour
	// 2. MPC controller reconstructs path in the same order as forecast
	// 3. Timestamp increases monotonically with hour
	minTimestamp := decisions[0].Timestamp

	// Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing decisions with timestamp >= minTimestamp
	_, err = tx.ExecContext(ctx, `DELETE FROM mpc_decisions WHERE timestamp >= $1`, minTimestamp)
	if err != nil {
		return fmt.Errorf("failed to delete existing decisions: %w", err)
	}

	// Prepare upsert statement
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO mpc_decisions (
			timestamp,
			hour,
			battery_charge,
			battery_discharge,
			grid_import,
			grid_export,
			battery_soc,
			profit,
			import_price,
			export_price,
			solar_forecast,
			load_forecast,
			cloud_coverage,
			weather_symbol
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (timestamp) DO UPDATE SET
			hour = EXCLUDED.hour,
			battery_charge = EXCLUDED.battery_charge,
			battery_discharge = EXCLUDED.battery_discharge,
			grid_import = EXCLUDED.grid_import,
			grid_export = EXCLUDED.grid_export,
			battery_soc = EXCLUDED.battery_soc,
			profit = EXCLUDED.profit,
			import_price = EXCLUDED.import_price,
			export_price = EXCLUDED.export_price,
			solar_forecast = EXCLUDED.solar_forecast,
			load_forecast = EXCLUDED.load_forecast,
			cloud_coverage = EXCLUDED.cloud_coverage,
			weather_symbol = EXCLUDED.weather_symbol
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Insert all decisions
	for _, decision := range decisions {
		_, err := stmt.ExecContext(ctx,
			decision.Timestamp,
			decision.Hour,
			decision.BatteryCharge,
			decision.BatteryDischarge,
			decision.GridImport,
			decision.GridExport,
			decision.BatterySOC,
			decision.Profit,
			decision.ImportPrice,
			decision.ExportPrice,
			decision.SolarForecast,
			decision.LoadForecast,
			decision.CloudCoverage,
			decision.WeatherSymbol,
		)
		if err != nil {
			return fmt.Errorf("failed to insert decision for hour %d: %w", decision.Hour, err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Printf("Saved %d MPC decisions to database", len(decisions))
	return nil
}

// loadLatestMPCDecisions loads MPC decisions from the database with timestamp >= now
func (s *MinerScheduler) loadLatestMPCDecisions(ctx context.Context) ([]mpc.ControlDecision, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	config := s.GetConfig()

	// Get current Unix timestamp
	now := ctx.Value("now")
	var nowTimestamp int64
	if now != nil {
		if t, ok := now.(int64); ok {
			nowTimestamp = t
		}
	}
	if nowTimestamp == 0 {
		nowTimestamp = s.getCurrentTimestamp()
	}

	ts := nowTimestamp - int64(config.CheckPriceInterval.Seconds())

	// Load decisions with timestamp >= now, ordered by timestamp
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			timestamp,
			hour,
			battery_charge,
			battery_discharge,
			grid_import,
			grid_export,
			battery_soc,
			profit,
			import_price,
			export_price,
			solar_forecast,
			load_forecast,
			cloud_coverage,
			weather_symbol
		FROM mpc_decisions
		WHERE timestamp >= $1
		ORDER BY timestamp ASC
	`, ts)
	if err != nil {
		return nil, fmt.Errorf("failed to query decisions: %w", err)
	}
	defer rows.Close()

	var decisions []mpc.ControlDecision
	for rows.Next() {
		var decision mpc.ControlDecision
		var cloudCoverage sql.NullFloat64
		var weatherSymbol sql.NullString

		err := rows.Scan(
			&decision.Timestamp,
			&decision.Hour,
			&decision.BatteryCharge,
			&decision.BatteryDischarge,
			&decision.GridImport,
			&decision.GridExport,
			&decision.BatterySOC,
			&decision.Profit,
			&decision.ImportPrice,
			&decision.ExportPrice,
			&decision.SolarForecast,
			&decision.LoadForecast,
			&cloudCoverage,
			&weatherSymbol,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan decision: %w", err)
		}

		if cloudCoverage.Valid {
			decision.CloudCoverage = cloudCoverage.Float64
		}
		if weatherSymbol.Valid {
			decision.WeatherSymbol = weatherSymbol.String
		}

		decisions = append(decisions, decision)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating decisions: %w", err)
	}

	if len(decisions) == 0 {
		s.logger.Printf("No future MPC decisions found in database")
		return nil, nil
	}

	s.logger.Printf("Loaded %d MPC decisions from database (starting from timestamp %d)", len(decisions), ts)

	return decisions, nil
}

// getCurrentTimestamp returns the current Unix timestamp
func (s *MinerScheduler) getCurrentTimestamp() int64 {
	return time.Now().Unix()
}
