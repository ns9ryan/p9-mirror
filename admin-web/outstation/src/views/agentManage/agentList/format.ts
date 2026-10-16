export function formatAmount(value: number) {
  return value.toLocaleString('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
}

export function formatRatio(value: number) {
  return `${value.toFixed(2)}%`;
}

export function displayText(value: string) {
  const text = value.trim();
  return text || '--';
}
