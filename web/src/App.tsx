import { useEffect, useState, useCallback, useRef } from "react";
import "./App.css";
import emsIcon from "../assets/icon-32.png";

interface SchedulerStatus {
  is_running: boolean;
  miners_count: number;
  has_market_data: boolean;
  price_limit: number;
  network: string;
}

interface HealthResponse {
  status: string;
  timestamp: string;
  version: string;
  scheduler: SchedulerStatus;
  system: {
    uptime: string;
    goroutines: number;
  };
  ems: {
    current_pv_power: number;
    ess_power: number;
    ess_soc: number;
    grid_sensor_status: number;
    grid_sensor_active_power: number;
    plant_active_power: number;
  };
}

interface StatusResponse {
  scheduler_status: {
    is_running: boolean;
    miners_count: number;
    has_market_data: boolean;
  };
  miners: {
    count: number;
    list: Array<{
      ip: string;
      status: string;
    }>;
  };
  price_data: {
    has_document: boolean;
    current_avg_price?: number;
    current?: number;
    limit?: number;
  };

  timestamp: string;
}

interface WebSocketMessage {
  type: string;
  health: HealthResponse;
  status: StatusResponse;
}

function App() {
  const [health, setHealth] = useState<HealthResponse | null>(null);
  const [status, setStatus] = useState<StatusResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [wsConnected, setWsConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<number | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const isConnectingRef = useRef(false);

  const connectWebSocket = useCallback(() => {
    // Prevent duplicate connections
    if (
      isConnectingRef.current ||
      (wsRef.current && wsRef.current.readyState === WebSocket.OPEN)
    ) {
      console.log("Already connecting or connected, skipping...");
      return;
    }

    isConnectingRef.current = true;

    // Clear any existing reconnect timeout
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    // Close existing connection if any
    if (wsRef.current) {
      wsRef.current.close();
    }

    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsUrl = `${protocol}//${window.location.host}/api/ws`;

    console.log("Connecting to WebSocket:", wsUrl);

    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log("WebSocket connected");
      setWsConnected(true);
      setError(null);
      setLoading(false);
      reconnectAttemptsRef.current = 0;
      isConnectingRef.current = false;
    };

    ws.onmessage = (event) => {
      try {
        const data: WebSocketMessage = JSON.parse(event.data);

        if (data.type === "status_update") {
          setHealth(data.health);
          setStatus(data.status);
          setError(null);
        }
      } catch (err) {
        console.error("Failed to parse WebSocket message:", err);
        setError("Failed to parse server data");
      }
    };

    ws.onerror = (event) => {
      console.error("WebSocket error:", event);
      setError("WebSocket connection error");
      setWsConnected(false);
      isConnectingRef.current = false;
    };

    ws.onclose = (event) => {
      console.log("WebSocket closed:", event.code, event.reason);
      setWsConnected(false);
      wsRef.current = null;
      isConnectingRef.current = false;

      // Attempt to reconnect with exponential backoff
      reconnectAttemptsRef.current += 1;
      const delay = Math.min(
        1000 * Math.pow(2, reconnectAttemptsRef.current),
        30000,
      );

      console.log(
        `Reconnecting in ${delay}ms (attempt ${reconnectAttemptsRef.current})`,
      );

      reconnectTimeoutRef.current = setTimeout(() => {
        connectWebSocket();
      }, delay);
    };
  }, []);

  useEffect(() => {
    connectWebSocket();

    // Cleanup on unmount
    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

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
          <div
            className={`status-badge ${isHealthy ? "healthy" : "unhealthy"}`}
          >
            {isHealthy ? "✓ Healthy" : "✗ Unhealthy"}
          </div>
          <div
            className={`status-badge ${wsConnected ? "connected" : "disconnected"}`}
          >
            {wsConnected ? "🔗 Connected" : "⚠️ Disconnected"}
          </div>
        </div>
      </header>

      <main className="main">
        <section className="card">
          <h2>Scheduler Status</h2>
          <div className="info-grid">
            <div className="info-item">
              <span className="label">Running:</span>
              <span
                className={
                  health?.scheduler.is_running ? "value-success" : "value-error"
                }
              >
                {health?.scheduler.is_running ? "Yes" : "No"}
              </span>
            </div>
            <div className="info-item">
              <span className="label">Network:</span>
              <span className="value">{health?.scheduler.network}</span>
            </div>
            <div className="info-item">
              <span className="label">Miners Count:</span>
              <span className="value">{status?.miners.count || 0}</span>
            </div>
            <div className="info-item">
              <span className="label">Market Data:</span>
              <span
                className={
                  health?.scheduler.has_market_data
                    ? "value-success"
                    : "value-warning"
                }
              >
                {health?.scheduler.has_market_data
                  ? "Available"
                  : "Not Available"}
              </span>
            </div>
          </div>
        </section>

        {currentPrice !== undefined && priceLimit !== undefined && (
          <section className="card">
            <h2>Price Information</h2>
            <div className="info-grid">
              <div className="info-item">
                <span className="label">Current Avg Price:</span>
                <span className="value">{currentPrice.toFixed(2)} €/MWh</span>
              </div>
              <div className="info-item">
                <span className="label">Price Limit:</span>
                <span className="value">{priceLimit.toFixed(2)} €/MWh</span>
              </div>
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
            <div
              className="pv-power-display"
              style={{ position: "absolute", top: "202px", right: "170px" }}
            >
              <div className="pv-power-value">
                {health?.ems?.current_pv_power !== undefined
                  ? `${health.ems.current_pv_power.toFixed(2)} kW`
                  : "N/A"}
              </div>
            </div>
            <div
              className="pv-power-display"
              style={{ position: "absolute", top: "643px", right: "622px" }}
            >
              <div
                className="pv-power-value"
                style={{
                  color:
                    health?.ems?.ess_power !== undefined
                      ? health.ems.ess_power > 0
                        ? "#ea580c"
                        : health.ems.ess_power < 0
                          ? "#16a34a"
                          : "#f1f5f9"
                      : "#f1f5f9",
                }}
              >
                {health?.ems?.ess_power !== undefined
                  ? `${health.ems.ess_power.toFixed(2)} kW`
                  : "N/A"}
              </div>
              <div className="pv-power-label" style={{ marginTop: "8px" }}>
                {health?.ems?.ess_soc !== undefined
                  ? `${health.ems.ess_soc.toFixed(1)}%`
                  : "N/A"}
              </div>
            </div>
            <div
              className="pv-power-display"
              style={{ position: "absolute", top: "556px", left: "220px" }}
            >
              <div
                className="pv-power-value"
                style={{
                  color:
                    health?.ems?.grid_sensor_active_power !== undefined
                      ? health.ems.grid_sensor_active_power > 0
                        ? "#ea580c"
                        : health.ems.grid_sensor_active_power < 0
                          ? "#16a34a"
                          : "#f1f5f9"
                      : "#f1f5f9",
                }}
              >
                {health?.ems?.grid_sensor_active_power !== undefined &&
                health?.ems?.grid_sensor_status === 1
                  ? `${health.ems.grid_sensor_active_power.toFixed(2)} kW`
                  : "N/A"}
              </div>
            </div>
            <div
              className="pv-power-display"
              style={{ position: "absolute", top: "194px", left: "223px" }}
            >
              <div className="pv-power-label">Active Power</div>
              <div
                className="pv-power-value"
                style={{
                  color:
                    health?.ems?.plant_active_power !== undefined
                      ? health.ems.plant_active_power < 0
                        ? "#ea580c"
                        : health.ems.plant_active_power > 0
                          ? "#16a34a"
                          : "#f1f5f9"
                      : "#f1f5f9",
                }}
              >
                {health?.ems?.plant_active_power !== undefined
                  ? `${health.ems.plant_active_power.toFixed(2)} kW`
                  : "N/A"}
              </div>
            </div>
          </div>
        </section>

        <section className="card system-info">
          <h2>System Information</h2>
          <div className="info-grid">
            <div className="info-item">
              <span className="label">Version:</span>
              <span className="value">{health?.version}</span>
            </div>
            <div className="info-item">
              <span className="label">Uptime:</span>
              <span className="value">{health?.system.uptime}</span>
            </div>
            <div className="info-item">
              <span className="label">Last Updated:</span>
              <span className="value">
                {status?.timestamp
                  ? new Date(status.timestamp).toLocaleString()
                  : "N/A"}
              </span>
            </div>
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
