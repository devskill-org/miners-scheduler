package main

import (
	"fmt"
	"math"
	"time"

	"github.com/devskill-org/miners-scheduler/mpc"
)

// Example usage
func main() {
	// Configure the system
	config := mpc.SystemConfig{
		BatteryCapacity:        10.0, // 10 kWh
		BatteryMaxCharge:       5.0,  // 5 kW
		BatteryMaxDischarge:    5.0,  // 5 kW
		BatteryMinSOC:          0.2,  // 20%
		BatteryMaxSOC:          1.0,  // 100%
		BatteryEfficiency:      0.92, // 92% round-trip
		BatteryDegradationCost: 0.05, // $0.05 per kWh cycled
		MaxGridImport:          10.0, // 10 kW
		MaxGridExport:          10.0, // 10 kW
	}

	// Create MPC controller with 24-hour horizon
	mpcController := mpc.NewMPCController(config, 24, 0.5) // Start at 50% SOC

	// Create 24-hour forecast
	forecast := make([]mpc.TimeSlot, 24)
	for i := range 24 {
		// Example: cheap at night, expensive during day
		importPrice := 0.10
		exportPrice := 0.08
		if i >= 8 && i <= 20 {
			importPrice = 0.25
			exportPrice = 0.20
		}

		// Solar production during day
		solarForecast := 0.0
		if i >= 6 && i <= 18 {
			// Simple sine curve for solar
			solarPeak := 5.0 // 5 kW peak
			angle := float64(i-6) / 12.0 * math.Pi
			solarForecast = solarPeak * math.Sin(angle)
		}

		// Load varies throughout day
		loadForecast := 1.5 // Base load
		if i >= 7 && i <= 22 {
			loadForecast = 2.5 // Higher during active hours
		}

		forecast[i] = mpc.TimeSlot{
			Hour:          i,
			ImportPrice:   importPrice,
			ExportPrice:   exportPrice,
			SolarForecast: solarForecast,
			LoadForecast:  loadForecast,
		}
	}

	// Run optimization
	fmt.Println("Solar Inverter MPC Optimization")
	fmt.Println("================================")
	fmt.Printf("Initial SOC: %.1f%%\n\n", mpcController.CurrentSOC*100)

	startTime := time.Now()
	decisions := mpcController.Optimize(forecast)
	duration := time.Since(startTime)

	fmt.Printf("Optimization completed in %v\n\n", duration)

	// Display results
	totalProfit := 0.0
	fmt.Println("Hour | Solar | Load  | Import$ | Export$ | Batt→ | Grid→ | SOC   | Profit")
	fmt.Println("-----|-------|-------|---------|---------|-------|-------|-------|--------")

	for _, dec := range decisions {
		slot := forecast[dec.Hour]
		totalProfit += dec.Profit

		battAction := "idle"
		if dec.BatteryCharge > 0.1 {
			battAction = fmt.Sprintf("+%.1fkW", dec.BatteryCharge)
		} else if dec.BatteryDischarge > 0.1 {
			battAction = fmt.Sprintf("-%.1fkW", dec.BatteryDischarge)
		}

		gridAction := "idle"
		if dec.GridImport > 0.1 {
			gridAction = fmt.Sprintf("←%.1fkW", dec.GridImport)
		} else if dec.GridExport > 0.1 {
			gridAction = fmt.Sprintf("→%.1fkW", dec.GridExport)
		}

		fmt.Printf("%4d | %5.1f | %5.1f | $%.3f  | $%.3f  | %-6s| %-6s| %4.0f%% | $%.3f\n",
			dec.Hour,
			slot.SolarForecast,
			slot.LoadForecast,
			slot.ImportPrice,
			slot.ExportPrice,
			battAction,
			gridAction,
			dec.BatterySOC*100,
			dec.Profit,
		)
	}

	fmt.Printf("\nTotal Profit (24h): $%.2f\n", totalProfit)
	fmt.Printf("Final SOC: %.1f%%\n", decisions[len(decisions)-1].BatterySOC*100)

	// Show strategy summary
	fmt.Println("\nStrategy Summary:")
	fmt.Println("- Charges battery from grid during cheap night hours")
	fmt.Println("- Charges battery from solar during day")
	fmt.Println("- Discharges battery during expensive peak hours")
	fmt.Println("- Exports excess solar when prices are favorable")
}
