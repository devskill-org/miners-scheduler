interface SolarInfoProps {
  solarAngle?: number;
  sunrise?: string;
  sunset?: string;
  style?: React.CSSProperties;
}

export function SolarInfo({
  solarAngle,
  sunrise,
  sunset,
  style,
}: SolarInfoProps) {
  const formatTime = (dateString?: string) => {
    if (!dateString) return "N/A";
    return new Date(dateString).toLocaleTimeString([], {
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const getSolarAngleColor = (angle?: number) => {
    if (angle === undefined) return "#666";
    if (angle < 0) return "#444"; // Below horizon
    if (angle < 15) return "#ff9800"; // Low angle
    if (angle < 45) return "#ffc107"; // Medium angle
    return "#4caf50"; // High angle
  };

  return (
    <div className="solar-info" style={style}>
      <div className="solar-info-content">
        <div className="solar-info-item">
          <div className="solar-info-label">Solar Angle</div>
          <div
            className="solar-info-value"
            style={{ color: getSolarAngleColor(solarAngle) }}
          >
            {solarAngle !== undefined ? `${solarAngle.toFixed(1)}°` : "N/A"}
          </div>
        </div>
        <div className="solar-info-item">
          <div className="solar-info-label">🌅 Sunrise</div>
          <div className="solar-info-value">{formatTime(sunrise)}</div>
        </div>
        <div className="solar-info-item">
          <div className="solar-info-label">🌇 Sunset</div>
          <div className="solar-info-value">{formatTime(sunset)}</div>
        </div>
      </div>
    </div>
  );
}
