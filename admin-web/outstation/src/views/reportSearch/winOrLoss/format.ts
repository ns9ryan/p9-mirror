export function formatAmount(value: number) {
  return value.toLocaleString('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
}

export function formatCount(value: number) {
  return value.toLocaleString('en-US');
}

/** 杀数 = 会员输赢 / 有效投注；没有有效投注时显示横线 */
export function formatKillRate(memberWinLoss: number, validBet: number) {
  if (!validBet) return '-';
  return `${((memberWinLoss / validBet) * 100).toFixed(2)}%`;
}
