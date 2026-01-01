interface StatusBadgeProps {
  isActive: boolean;
  activeLabel: string;
  inactiveLabel: string;
}

export function StatusBadge({
  isActive,
  activeLabel,
  inactiveLabel,
}: StatusBadgeProps) {
  return (
    <div className={`status-badge ${isActive ? "healthy" : "unhealthy"}`}>
      {isActive ? activeLabel : inactiveLabel}
    </div>
  );
}
