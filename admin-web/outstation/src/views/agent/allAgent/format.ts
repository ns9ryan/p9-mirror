export function formatAmount(value: number) {
  return value.toFixed(2);
}

export function displayText(value: string) {
  const text = value.trim();
  return text || '--';
}
