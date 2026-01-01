// Utility function to determine power value color
export function getPowerColor(
  value: number | undefined,
  invertColors = false,
): string {
  if (value === undefined) return "#f1f5f9";
  if (value === 0) return "#f1f5f9";

  if (invertColors) {
    return value < 0 ? "#ea580c" : "#16a34a";
  }
  return value > 0 ? "#ea580c" : "#16a34a";
}
