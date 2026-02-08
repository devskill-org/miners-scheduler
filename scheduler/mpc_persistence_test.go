package scheduler

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"github.com/devskill-org/ems/mpc"
	_ "github.com/lib/pq"
)

// TestMPCPersistence_SaveAndLoad tests the save and load cycle
func TestMPCPersistence_SaveAndLoad(t *testing.T) {
	// Skip if no database connection available
	connString := os.Getenv("TEST_POSTGRES_CONN")
	if connString == "" {
		t.Skip("Skipping test: TEST_POSTGRES_CONN not set")
	}

	db, err := sql.Open("postgres", connString)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Clean up table before test
	_, err = db.Exec("DELETE FROM mpc_decisions")
	if err != nil {
		t.Fatalf("Failed to clean up table: %v", err)
	}

	// Create scheduler with database
	config := &Config{}
	scheduler := &MinerScheduler{
		config: config,
		db:     db,
		logger: log.New(os.Stdout, "TEST: ", log.LstdFlags),
	}

	// Create test decisions with timestamps in the future
	now := time.Now().Unix()
	decisions := []mpc.ControlDecision{
		{
			Hour:                 0,
			Timestamp:            now + 3600,
			BatteryCharge:        10.5,
			BatteryChargeFromPV:  10.5,
			BatteryChargeFromGrid: 5.0,
			BatteryDischarge:     0,
			GridImport:           5.0,
			GridExport:           0,
			BatterySOC:           0.6,
			Profit:               2.5,
			ImportPrice:          0.1,
			ExportPrice:          0.05,
			SolarForecast:        15.0,
			LoadForecast:         10.0,
			CloudCoverage:        30.0,
			WeatherSymbol:        "clearsky_day",
		},
		{
			Hour:                 1,
			Timestamp:            now + 7200,
			BatteryCharge:        0,
			BatteryChargeFromPV:  0,
			BatteryChargeFromGrid: 0,
			BatteryDischarge:     8.0,
			GridImport:           0,
			GridExport:           3.0,
			BatterySOC:           0.5,
			Profit:               3.2,
			ImportPrice:          0.12,
			ExportPrice:          0.06,
			SolarForecast:        20.0,
			LoadForecast:         12.0,
			CloudCoverage:        10.0,
			WeatherSymbol:        "fair_day",
		},
	}

	ctx := context.Background()

	// Save decisions
	err = scheduler.saveMPCDecisions(ctx, decisions)
	if err != nil {
		t.Fatalf("Failed to save decisions: %v", err)
	}

	// Load decisions
	loaded, err := scheduler.loadLatestMPCDecisions(ctx)
	if err != nil {
		t.Fatalf("Failed to load decisions: %v", err)
	}

	// Verify loaded decisions
	if len(loaded) != len(decisions) {
		t.Errorf("Expected %d decisions, got %d", len(decisions), len(loaded))
	}

	for i, decision := range loaded {
		if decision.Timestamp != decisions[i].Timestamp {
			t.Errorf("Decision %d: expected timestamp %d, got %d", i, decisions[i].Timestamp, decision.Timestamp)
		}
		if decision.BatteryCharge != decisions[i].BatteryCharge {
			t.Errorf("Decision %d: expected battery_charge %.2f, got %.2f", i, decisions[i].BatteryCharge, decision.BatteryCharge)
		}
		if decision.BatteryChargeFromPV != decisions[i].BatteryChargeFromPV {
			t.Errorf("Decision %d: expected battery_charge_from_pv %.2f, got %.2f", i, decisions[i].BatteryChargeFromPV, decision.BatteryChargeFromPV)
		}
		if decision.BatteryChargeFromGrid != decisions[i].BatteryChargeFromGrid {
			t.Errorf("Decision %d: expected battery_charge_from_grid %.2f, got %.2f", i, decisions[i].BatteryChargeFromGrid, decision.BatteryChargeFromGrid)
		}
		if decision.Profit != decisions[i].Profit {
			t.Errorf("Decision %d: expected profit %.2f, got %.2f", i, decisions[i].Profit, decision.Profit)
		}
	}
}

// TestMPCPersistence_DeleteOldDecisions tests that old decisions are replaced
func TestMPCPersistence_DeleteOldDecisions(t *testing.T) {
	// Skip if no database connection available
	connString := os.Getenv("TEST_POSTGRES_CONN")
	if connString == "" {
		t.Skip("Skipping test: TEST_POSTGRES_CONN not set")
	}

	db, err := sql.Open("postgres", connString)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Clean up table before test
	_, err = db.Exec("DELETE FROM mpc_decisions")
	if err != nil {
		t.Fatalf("Failed to clean up table: %v", err)
	}

	// Create scheduler with database
	config := &Config{}
	scheduler := &MinerScheduler{
		config: config,
		db:     db,
		logger: log.New(os.Stdout, "TEST: ", log.LstdFlags),
	}

	now := time.Now().Unix()
	ctx := context.Background()

	// First, save decisions for hours 0-2
	firstDecisions := []mpc.ControlDecision{
		{Hour: 0, Timestamp: now + 3600, Profit: 1.0},
		{Hour: 1, Timestamp: now + 7200, Profit: 2.0},
		{Hour: 2, Timestamp: now + 10800, Profit: 3.0},
	}
	err = scheduler.saveMPCDecisions(ctx, firstDecisions)
	if err != nil {
		t.Fatalf("Failed to save first decisions: %v", err)
	}

	// Then, save new decisions starting from hour 1 (should replace hours 1-2)
	secondDecisions := []mpc.ControlDecision{
		{Hour: 1, Timestamp: now + 7200, Profit: 20.0},  // Updated
		{Hour: 2, Timestamp: now + 10800, Profit: 30.0}, // Updated
		{Hour: 3, Timestamp: now + 14400, Profit: 40.0}, // New
	}
	err = scheduler.saveMPCDecisions(ctx, secondDecisions)
	if err != nil {
		t.Fatalf("Failed to save second decisions: %v", err)
	}

	// Load all decisions (including past)
	var allDecisions []mpc.ControlDecision
	rows, err := db.Query("SELECT timestamp, profit FROM mpc_decisions ORDER BY timestamp")
	if err != nil {
		t.Fatalf("Failed to query decisions: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var d mpc.ControlDecision
		err := rows.Scan(&d.Timestamp, &d.Profit)
		if err != nil {
			t.Fatalf("Failed to scan decision: %v", err)
		}
		allDecisions = append(allDecisions, d)
	}

	// Should have 4 decisions: hour 0 (unchanged), hours 1-3 (new/updated)
	if len(allDecisions) != 4 {
		t.Errorf("Expected 4 decisions, got %d", len(allDecisions))
	}

	// Verify hour 0 is unchanged
	if allDecisions[0].Timestamp == now+3600 && allDecisions[0].Profit != 1.0 {
		t.Errorf("Hour 0 should be unchanged with profit 1.0, got %.2f", allDecisions[0].Profit)
	}

	// Verify hour 1 is updated
	if allDecisions[1].Timestamp == now+7200 && allDecisions[1].Profit != 20.0 {
		t.Errorf("Hour 1 should be updated with profit 20.0, got %.2f", allDecisions[1].Profit)
	}
}

// TestMPCPersistence_LoadOnlyFutureDecisions tests that only future decisions are loaded
func TestMPCPersistence_LoadOnlyFutureDecisions(t *testing.T) {
	// Skip if no database connection available
	connString := os.Getenv("TEST_POSTGRES_CONN")
	if connString == "" {
		t.Skip("Skipping test: TEST_POSTGRES_CONN not set")
	}

	db, err := sql.Open("postgres", connString)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Clean up table before test
	_, err = db.Exec("DELETE FROM mpc_decisions")
	if err != nil {
		t.Fatalf("Failed to clean up table: %v", err)
	}

	// Create scheduler with database
	config := &Config{}
	scheduler := &MinerScheduler{
		config: config,
		db:     db,
		logger: log.New(os.Stdout, "TEST: ", log.LstdFlags),
	}

	now := time.Now().Unix()
	ctx := context.Background()

	// Save decisions: some in the past, some in the future
	decisions := []mpc.ControlDecision{
		{Hour: 0, Timestamp: now - 3600, Profit: 1.0}, // Past
		{Hour: 1, Timestamp: now - 1800, Profit: 2.0}, // Past
		{Hour: 2, Timestamp: now + 1800, Profit: 3.0}, // Future
		{Hour: 3, Timestamp: now + 3600, Profit: 4.0}, // Future
		{Hour: 4, Timestamp: now + 7200, Profit: 5.0}, // Future
	}

	// Insert directly to test load filtering
	for _, d := range decisions {
		_, err := db.Exec(`
			INSERT INTO mpc_decisions (timestamp, hour, battery_charge, battery_charge_from_pv,
				battery_charge_from_grid, battery_discharge, grid_import, grid_export, battery_soc,
				profit, import_price, export_price, solar_forecast, load_forecast)
			VALUES ($1, $2, 0, 0, 0, 0, 0, 0, 0.5, $3, 0.1, 0.05, 10, 5)
		`, d.Timestamp, d.Hour, d.Profit)
		if err != nil {
			t.Fatalf("Failed to insert decision: %v", err)
		}
	}

	// Load decisions (should only get future ones)
	loaded, err := scheduler.loadLatestMPCDecisions(ctx)
	if err != nil {
		t.Fatalf("Failed to load decisions: %v", err)
	}

	// Should only load 3 future decisions
	if len(loaded) != 3 {
		t.Errorf("Expected 3 future decisions, got %d", len(loaded))
	}

	// Verify all loaded decisions are in the future
	for i, decision := range loaded {
		if decision.Timestamp < now {
			t.Errorf("Decision %d has past timestamp %d (now: %d)", i, decision.Timestamp, now)
		}
	}

	// Verify they are ordered by timestamp
	for i := 1; i < len(loaded); i++ {
		if loaded[i].Timestamp <= loaded[i-1].Timestamp {
			t.Errorf("Decisions not properly ordered by timestamp")
		}
	}
}

// TestMPCPersistence_UniqueTimestamp tests that timestamp PRIMARY KEY prevents duplicates
func TestMPCPersistence_UniqueTimestamp(t *testing.T) {
	// Skip if no database connection available
	connString := os.Getenv("TEST_POSTGRES_CONN")
	if connString == "" {
		t.Skip("Skipping test: TEST_POSTGRES_CONN not set")
	}

	db, err := sql.Open("postgres", connString)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Clean up table before test
	_, err = db.Exec("DELETE FROM mpc_decisions")
	if err != nil {
		t.Fatalf("Failed to clean up table: %v", err)
	}

	now := time.Now().Unix()
	timestamp := now + 3600

	// Insert first decision
	_, err = db.Exec(`
		INSERT INTO mpc_decisions (timestamp, hour, battery_charge, battery_charge_from_pv,
			battery_charge_from_grid, battery_discharge, grid_import, grid_export, battery_soc,
			profit, import_price, export_price, solar_forecast, load_forecast)
		VALUES ($1, 0, 10, 10, 0, 0, 5, 0, 0.6, 2.5, 0.1, 0.05, 15, 10)
	`, timestamp)
	if err != nil {
		t.Fatalf("Failed to insert first decision: %v", err)
	}

	// Try to insert duplicate timestamp (should be handled by UPSERT in saveMPCDecisions)
	_, err = db.Exec(`
		INSERT INTO mpc_decisions (timestamp, hour, battery_charge, battery_charge_from_pv,
			battery_charge_from_grid, battery_discharge, grid_import, grid_export, battery_soc,
			profit, import_price, export_price, solar_forecast, load_forecast)
		VALUES ($1, 1, 20, 20, 0, 0, 10, 0, 0.7, 5.0, 0.12, 0.06, 20, 12)
		ON CONFLICT (timestamp) DO UPDATE SET
			hour = EXCLUDED.hour,
			profit = EXCLUDED.profit
	`, timestamp)
	if err != nil {
		t.Fatalf("UPSERT failed: %v", err)
	}

	// Verify only one row exists and it's updated
	var count int
	var profit float64
	var hour int
	err = db.QueryRow("SELECT COUNT(*), MAX(hour), MAX(profit) FROM mpc_decisions WHERE timestamp = $1", timestamp).Scan(&count, &hour, &profit)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 row, got %d", count)
	}
	if hour != 1 {
		t.Errorf("Expected hour to be updated to 1, got %d", hour)
	}
	if profit != 5.0 {
		t.Errorf("Expected profit to be updated to 5.0, got %.2f", profit)
	}
}
