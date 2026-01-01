interface InfoItemProps {
  label: string;
  value: React.ReactNode;
  valueClassName?: string;
}

export function InfoItem({ label, value, valueClassName }: InfoItemProps) {
  return (
    <div className="info-item">
      <span className="label">{label}</span>
      <span className={valueClassName || "value"}>{value}</span>
    </div>
  );
}
