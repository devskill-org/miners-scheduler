import "./DemoInfo.css";

interface DemoInfoProps {
  onClose: () => void;
}

export function DemoInfo({ onClose }: DemoInfoProps) {
  return (
    <div className="demo-info-overlay" onClick={onClose}>
      <div className="demo-info-card" onClick={(e) => e.stopPropagation()}>
        <button className="demo-info-close" onClick={onClose}>
          ✕
        </button>
        <h2>🎭 Demo Mode</h2>
        <p className="demo-info-description">
          You are viewing a demonstration version of the Energy Management
          System with simulated data.
        </p>
        <div className="demo-info-section">
          <h3>What is Demo Mode?</h3>
          <p>
            This application is running without a backend server. All data you
            see is generated in real-time to simulate realistic energy system
            behavior.
          </p>
        </div>
        <div className="demo-info-section">
          <h3>Data Updates:</h3>
          <p>
            Mock data refreshes every 10 seconds to simulate real-time updates.
            Values change based on time of day to reflect realistic energy
            patterns.
          </p>
        </div>
        <button className="demo-info-button" onClick={onClose}>
          Got it!
        </button>
      </div>
    </div>
  );
}
