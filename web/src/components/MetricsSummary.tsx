import { useEffect, useState, useRef } from "react";
import { MetricsSummary as MetricsSummaryType } from "../types/api";
import "../App.css";

// Check if we're in demo mode
const isDemoMode = typeof __DEMO_MODE__ !== 'undefined' && __DEMO_MODE__;

// Generate mock metrics summary data for demo mode
function generateMockMetricsSummary(date: Date): MetricsSummaryType {
  const dayOfMonth = date.getDate();
  
  // Use day of month to generate consistent but varying data
  const seed = dayOfMonth / 31;
  
  // Generate realistic import/export values
  const totalImportKwh = 50 + seed * 100; // 50-150 kWh
  const totalExportKwh = 30 + seed * 80; // 30-110 kWh
  
  // Typical electricity prices in €/kWh
  const avgImportPrice = 0.12; // 12 cents per kWh
  const avgExportPrice = 0.08; // 8 cents per kWh
  
  const totalImportCost = totalImportKwh * avgImportPrice;
  const totalExportCost = totalExportKwh * avgExportPrice;
  
  const startTime = new Date(date);
  startTime.setHours(0, 0, 0, 0);
  
  const endTime = new Date(date);
  endTime.setHours(23, 59, 59, 999);
  
  return {
    total_import_cost: totalImportCost,
    total_export_cost: totalExportCost,
    total_import_kwh: totalImportKwh,
    total_export_kwh: totalExportKwh,
    start_time: startTime.toISOString(),
    end_time: endTime.toISOString(),
  };
}

export function MetricsSummary() {
  const [metricsSummary, setMetricsSummary] =
    useState<MetricsSummaryType | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedDate, setSelectedDate] = useState<Date>(() => {
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    return today;
  });
  const isFetchingRef = useRef(false);

  useEffect(() => {
    const fetchMetricsSummary = async () => {
      // Prevent duplicate fetches
      if (isFetchingRef.current) {
        console.log("Already fetching metrics, skipping...");
        return;
      }

      isFetchingRef.current = true;

      try {
        setLoading(true);

        // In demo mode, use mock data
        if (isDemoMode) {
          // Simulate network delay
          await new Promise(resolve => setTimeout(resolve, 300));
          const mockData = generateMockMetricsSummary(selectedDate);
          setMetricsSummary(mockData);
          setError(null);
        } else {
          // Calculate time range for selected date (calendar day - midnight to midnight)
          const startTime = selectedDate.toISOString();
          const endDate = new Date(selectedDate);
          endDate.setHours(23, 59, 59, 999);
          const endTime = endDate.toISOString();

          const url = `/api/metrics/summary?start_time=${encodeURIComponent(startTime)}&end_time=${encodeURIComponent(endTime)}`;
          const response = await fetch(url);
          if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
          }
          const data: MetricsSummaryType = await response.json();
          setMetricsSummary(data);
          setError(null);
        }
      } catch (error) {
        console.error("Failed to fetch metrics summary:", error);
        setError("Failed to load actual costs data");
      } finally {
        setLoading(false);
        isFetchingRef.current = false;
      }
    };

    fetchMetricsSummary();
  }, [selectedDate]);

  const handleDateNavigation = (dayShift: number | null) => {
    if (dayShift === null) {
      // Navigate to today
      const today = new Date();
      today.setHours(0, 0, 0, 0);
      setSelectedDate(today);
    } else {
      // Navigate by day shift (positive or negative)
      setSelectedDate((prev) => {
        const newDate = new Date(prev);
        newDate.setDate(newDate.getDate() + dayShift);
        return newDate;
      });
    }
  };

  const formatDateDisplay = (date: Date): string => {
    return date.toLocaleDateString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  };

  const isToday = () => {
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    return selectedDate.getTime() === today.getTime();
  };

  // Calculate net cost from metrics (import cost is negative, export cost is positive revenue)
  const netActualCost = metricsSummary
    ? metricsSummary.total_import_cost - metricsSummary.total_export_cost
    : null;

  return (
    <section className="card">
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          marginBottom: "1rem",
        }}
      >
        <h2 style={{ margin: 0 }}>
          Actual Costs - {formatDateDisplay(selectedDate)}
        </h2>
        <div style={{ display: "flex", gap: "0.5rem", alignItems: "center" }}>
          <button
            onClick={() => handleDateNavigation(-1)}
            style={{
              padding: "0.25rem 0.5rem",
              cursor: "pointer",
              border: "1px solid var(--color-border)",
              borderRadius: "4px",
              backgroundColor: "var(--color-bg-secondary)",
              color: "var(--color-text)",
            }}
            title="Previous day"
          >
            ←
          </button>
          <button
            onClick={() => handleDateNavigation(null)}
            disabled={isToday()}
            style={{
              padding: "0.25rem 0.75rem",
              cursor: isToday() ? "not-allowed" : "pointer",
              border: "1px solid var(--color-border)",
              borderRadius: "4px",
              backgroundColor: isToday()
                ? "var(--color-bg)"
                : "var(--color-bg-secondary)",
              color: isToday()
                ? "var(--color-text-secondary)"
                : "var(--color-text)",
              opacity: isToday() ? 0.5 : 1,
            }}
            title="Today"
          >
            Today
          </button>
          <button
            onClick={() => handleDateNavigation(1)}
            disabled={isToday()}
            style={{
              padding: "0.25rem 0.5rem",
              cursor: isToday() ? "not-allowed" : "pointer",
              border: "1px solid var(--color-border)",
              borderRadius: "4px",
              backgroundColor: isToday()
                ? "var(--color-bg)"
                : "var(--color-bg-secondary)",
              color: isToday()
                ? "var(--color-text-secondary)"
                : "var(--color-text)",
              opacity: isToday() ? 0.5 : 1,
            }}
            title="Next day"
          >
            →
          </button>
        </div>
      </div>

      <div className="mpc-summary">
        {loading && (
          <div className="mpc-summary-item">
            <span className="mpc-summary-label">Loading...</span>
          </div>
        )}
        {error && (
          <div className="mpc-summary-item">
            <span
              className="mpc-summary-label"
              style={{ color: "var(--color-error)" }}
            >
              {error}
            </span>
          </div>
        )}
        {metricsSummary && !loading && (
          <>
            <div className="mpc-summary-item">
              <span className="mpc-summary-label">Import Cost:</span>
              <span className="mpc-summary-value value-error">
                €{metricsSummary.total_import_cost.toFixed(2)} (
                {metricsSummary.total_import_kwh.toFixed(2)} kWh)
              </span>
            </div>
            <div className="mpc-summary-item">
              <span className="mpc-summary-label">Export Revenue:</span>
              <span className="mpc-summary-value value-success">
                €{metricsSummary.total_export_cost.toFixed(2)} (
                {metricsSummary.total_export_kwh.toFixed(2)} kWh)
              </span>
            </div>
            <div className="mpc-summary-item">
              <span className="mpc-summary-label">
                {netActualCost !== null && netActualCost <= 0
                  ? "Net Revenue:"
                  : "Net Cost:"}
              </span>
              <span
                className={`mpc-summary-value ${netActualCost !== null && netActualCost <= 0 ? "value-success" : "value-error"}`}
              >
                {netActualCost !== null
                  ? `€${Math.abs(netActualCost).toFixed(2)}`
                  : "N/A"}
              </span>
            </div>
          </>
        )}
      </div>
    </section>
  );
}
