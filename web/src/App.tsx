import "./App.css";
import emsIcon from "../assets/icon-32.png";
import { InfoItem } from "./components/InfoItem";
import { StatusBadge } from "./components/StatusBadge";
import { PowerDisplay } from "./components/PowerDisplay";
import { SolarInfo } from "./components/SolarInfo";
import { MPCDecisions } from "./components/MPCDecisions";
import { MetricsSummary } from "./components/MetricsSummary";
import { useWebSocket } from "./hooks/useWebSocket";

function App() {
  const { health, status, loading, error, wsConnected } = useWebSocket();

  if (loading) {
    return (
      <div className="app">
        <div className="loading">Connecting to server...</div>
      </div>
    );
  }

  if (error && !wsConnected) {
    return (
      <div className="app">
        <div className="error">
          <p>Error: {error}</p>
          <p>Attempting to reconnect...</p>
        </div>
      </div>
    );
  }

  const isHealthy = health?.status === "healthy";
  const currentPrice = status?.price_data?.current_avg_price;
  const priceLimit = status?.price_data?.limit;

  return (
    <div className="app">
      <header className="header">
        <h1>
          <img
            src={emsIcon}
            alt="EMS"
            style={{
              height: "32px",
              marginRight: "12px",
              verticalAlign: "middle",
            }}
          />
          Energy Management System
        </h1>
        <div className="status-badges">
          <StatusBadge
            isActive={isHealthy}
            activeLabel="✓ Healthy"
            inactiveLabel="✗ Unhealthy"
          />
          <StatusBadge
            isActive={wsConnected}
            activeLabel="🔗 Connected"
            inactiveLabel="⚠️ Disconnected"
          />
        </div>
      </header>

      <main className="main">
        <section className="card">
          <h2>Scheduler Status</h2>
          <div className="info-grid">
            <InfoItem
              label="Running:"
              value={health?.scheduler.is_running ? "Yes" : "No"}
              valueClassName={
                health?.scheduler.is_running ? "value-success" : "value-error"
              }
            />
            <InfoItem label="Network:" value={health?.scheduler.network} />
            <InfoItem label="Miners Count:" value={status?.miners.count || 0} />
            <InfoItem
              label="Market Data:"
              value={
                health?.scheduler.has_market_data
                  ? "Available"
                  : "Not Available"
              }
              valueClassName={
                health?.scheduler.has_market_data
                  ? "value-success"
                  : "value-warning"
              }
            />
          </div>
        </section>

        {currentPrice !== undefined && priceLimit !== undefined && (
          <section className="card">
            <h2>Price Information</h2>
            <div className="info-grid">
              <InfoItem
                label="Current Avg Price:"
                value={`${currentPrice.toFixed(2)} €/MWh`}
              />
              <InfoItem
                label="Price Limit:"
                value={`${priceLimit.toFixed(2)} €/MWh`}
              />
            </div>
          </section>
        )}

        {status?.miners.list && status.miners.list.length > 0 && (
          <section className="card">
            <h2>Discovered Miners</h2>
            <div className="miners-list">
              {status.miners.list.map((miner, index) => (
                <div key={index} className="miner-item">
                  <div className="miner-ip">{miner.ip}</div>
                  <div
                    className={`miner-status status-${miner.status?.toLowerCase()}`}
                  >
                    {miner.status || "Unknown"}
                  </div>
                </div>
              ))}
            </div>
          </section>
        )}

        <section className="card devices-section">
          <h2>Devices</h2>
          <div className="devices-content" style={{ position: "relative" }}>
            <PowerDisplay
              value={health?.ems?.current_pv_power}
              style={{ position: "absolute", top: "202px", right: "170px" }}
            />

            <PowerDisplay
              value={health?.ems?.ess_power}
              invertColors={true}
              label={
                health?.ems?.ess_soc !== undefined
                  ? `${health.ems.ess_soc.toFixed(1)}%`
                  : "N/A"
              }
              style={{ position: "absolute", top: "643px", right: "622px" }}
            />

            <PowerDisplay
              value={
                health?.ems?.grid_sensor_status === 1
                  ? health?.ems?.grid_sensor_active_power
                  : undefined
              }
              style={{ position: "absolute", top: "556px", left: "220px" }}
            />

            <PowerDisplay
              value={health?.ems?.plant_active_power}
              label="Active Power"
              invertColors={true}
              showLabel={true}
              style={{ position: "absolute", top: "194px", left: "223px" }}
            />

            <PowerDisplay
              value={health?.ems?.dc_charger_output_power}
              label={
                health?.ems?.dc_charger_vehicle_soc !== undefined
                  ? `${health.ems.dc_charger_vehicle_soc.toFixed(1)}%`
                  : "N/A"
              }
              style={{ position: "absolute", top: "428px", right: "162px" }}
            />

            <SolarInfo
              solarAngle={health?.sun?.solar_angle}
              sunrise={health?.sun?.sunrise}
              sunset={health?.sun?.sunset}
              style={{
                position: "absolute",
                top: "-75px",
                left: "100px",
                width: "175px",
              }}
            />
          </div>
        </section>

        <MPCDecisions decisions={health?.scheduler.mpc_decisions} />

        <MetricsSummary />

        <section className="card system-info">
          <h2>System Information</h2>
          <div className="info-grid">
            <InfoItem label="Version:" value={health?.version} />
            <InfoItem label="Uptime:" value={health?.system.uptime} />
            <InfoItem
              label="Last Updated:"
              value={
                status?.timestamp
                  ? new Date(status.timestamp).toLocaleString()
                  : "N/A"
              }
            />
          </div>
        </section>
      </main>

      <footer className="footer">
        <p>
          Avalon miners scheduler based on electricity prices and plant
          available power
        </p>
      </footer>
    </div>
  );
}

export default App;
