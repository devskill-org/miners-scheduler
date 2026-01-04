import { MPCDecisionInfo } from "../types/api";
import "../App.css";

// Helper function to format timestamp
const formatTimestamp = (timestamp: number): string => {
  // Check if timestamp is valid (not 0, undefined, or null)
  if (!timestamp || timestamp === 0) {
    return "N/A";
  }

  const date = new Date(timestamp * 1000);

  // Check if date is valid
  if (isNaN(date.getTime())) {
    return "Invalid Date";
  }

  return date.toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
};

// Helper function to convert weather symbol to emoji
const getWeatherIcon = (symbol: string): string => {
  if (!symbol) return "❓";

  const s = symbol.toLowerCase();

  // Thunder conditions
  if (s.includes("thunder")) {
    if (s.includes("snow")) return "⛈️❄️";
    if (s.includes("sleet")) return "⛈️🌧️";
    return "⛈️";
  }

  // Snow conditions
  if (s.includes("snow")) {
    if (s.includes("heavy")) return "❄️❄️";
    if (s.includes("light")) return "🌨️";
    return "❄️";
  }

  // Sleet conditions
  if (s.includes("sleet")) {
    if (s.includes("heavy")) return "🌧️❄️";
    return "🌨️";
  }

  // Rain conditions
  if (s.includes("rain")) {
    if (s.includes("heavy")) return "🌧️🌧️";
    if (s.includes("light")) return "🌦️";
    return "🌧️";
  }

  // Fog
  if (s.includes("fog")) return "🌫️";

  // Cloud conditions
  if (s.includes("cloudy")) return "☁️";
  if (s.includes("partlycloudy")) {
    if (s.includes("_night")) return "☁️🌙";
    return "⛅";
  }

  // Clear/Fair conditions
  if (s.includes("clearsky") || s.includes("fair")) {
    if (s.includes("_night") || s.includes("polartwilight")) return "🌙";
    return "☀️";
  }

  return "❓";
};

interface MPCDecisionsProps {
  decisions?: MPCDecisionInfo[];
}

export function MPCDecisions({ decisions }: MPCDecisionsProps) {
  if (!decisions || decisions.length === 0) {
    return (
      <section className="card">
        <h2>Model Predictive Control Optimization Results</h2>
        <div className="info-grid">
          <p>No optimization data available yet.</p>
        </div>
      </section>
    );
  }

  // Calculate total profit
  const totalProfit = decisions.reduce((sum, dec) => sum + dec.profit, 0);

  // Find highest and lowest import/export prices
  const importPrices = decisions.map((d) => d.import_price);
  const exportPrices = decisions.map((d) => d.export_price);

  const highestImport = Math.max(...importPrices);
  const lowestImport = Math.min(...importPrices);
  const highestExport = Math.max(...exportPrices);
  const lowestExport = Math.min(...exportPrices);

  // Helper function to determine battery action
  const getBatteryAction = (decision: MPCDecisionInfo) => {
    if (decision.battery_charge > 0.1) {
      return { action: "charge", power: decision.battery_charge };
    } else if (decision.battery_discharge > 0.1) {
      return { action: "discharge", power: decision.battery_discharge };
    }
    return { action: "idle", power: 0 };
  };

  // Helper function to determine grid action
  const getGridAction = (decision: MPCDecisionInfo) => {
    if (decision.grid_import > 0.1) {
      return { action: "import", power: decision.grid_import };
    } else if (decision.grid_export > 0.1) {
      return { action: "export", power: decision.grid_export };
    }
    return { action: "idle", power: 0 };
  };

  // Helper function to get action color class
  const getActionClass = (action: string) => {
    switch (action) {
      case "charge":
      case "import":
        return "action-import";
      case "discharge":
      case "export":
        return "action-export";
      default:
        return "action-idle";
    }
  };

  return (
    <section className="card">
      <h2>Model Predictive Control Optimization Results</h2>
      <div className="mpc-summary">
        <div className="mpc-summary-item">
          <span className="mpc-summary-label">Decisions:</span>
          <span className="mpc-summary-value">{decisions.length} hours</span>
        </div>
        <div className="mpc-summary-item">
          <span className="mpc-summary-label">Total Expected Profit:</span>
          <span
            className={`mpc-summary-value ${totalProfit >= 0 ? "value-success" : "value-error"}`}
          >
            €{totalProfit.toFixed(2)}
          </span>
        </div>
      </div>

      <div className="mpc-table-container">
        <table className="mpc-table">
          <thead>
            <tr>
              <th rowSpan={2}>Time</th>
              <th rowSpan={2}>Battery Action</th>
              <th rowSpan={2}>Grid Action</th>
              <th rowSpan={2}>SOC</th>
              <th
                colSpan={6}
                style={{
                  textAlign: "center",
                  borderBottom: "1px solid var(--color-border)",
                  backgroundColor: "rgba(51, 65, 85, 0.5)",
                }}
              >
                Forecast Data
              </th>
              <th rowSpan={2}>Profit</th>
            </tr>
            <tr>
              <th>Import (€/MWh)</th>
              <th>Export (€/MWh)</th>
              <th>Solar (kW)</th>
              <th>Load (kW)</th>
              <th>Cloud (%)</th>
              <th>Weather</th>
            </tr>
          </thead>
          <tbody>
            {decisions.map((decision) => {
              const batteryAction = getBatteryAction(decision);
              const gridAction = getGridAction(decision);

              return (
                <tr key={decision.hour}>
                  <td>{formatTimestamp(decision.timestamp)}</td>
                  <td>
                    <span className={getActionClass(batteryAction.action)}>
                      {batteryAction.action}
                      {batteryAction.power > 0 &&
                        `: ${batteryAction.power.toFixed(1)} kW`}
                    </span>
                  </td>
                  <td>
                    <span className={getActionClass(gridAction.action)}>
                      {gridAction.action}
                      {gridAction.power > 0 &&
                        `: ${gridAction.power.toFixed(1)} kW`}
                    </span>
                  </td>
                  <td>{(decision.battery_soc * 100).toFixed(1)}%</td>
                  <td>
                    <span
                      className={
                        decision.import_price === highestImport
                          ? "price-highest"
                          : decision.import_price === lowestImport
                            ? "price-lowest"
                            : ""
                      }
                    >
                      {(decision.import_price * 1000).toFixed(2)}
                    </span>
                  </td>
                  <td>
                    <span
                      className={
                        decision.export_price === highestExport
                          ? "price-highest"
                          : decision.export_price === lowestExport
                            ? "price-lowest"
                            : ""
                      }
                    >
                      {(decision.export_price * 1000).toFixed(2)}
                    </span>
                  </td>
                  <td>{decision.solar_forecast.toFixed(1)}</td>
                  <td>{decision.load_forecast.toFixed(1)}</td>
                  <td>{decision.cloud_coverage.toFixed(0)}</td>
                  <td title={decision.weather_symbol || "Unknown"}>
                    {getWeatherIcon(decision.weather_symbol)}
                  </td>
                  <td
                    className={
                      decision.profit >= 0 ? "value-success" : "value-error"
                    }
                  >
                    €{decision.profit.toFixed(3)}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </section>
  );
}
