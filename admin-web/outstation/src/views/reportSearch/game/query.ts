import { siteNow } from '@/utils/site-time';
import type { GameReportRecord, GameReportRow, QuickRange } from './types';

export function buildQuickRange(
  key: QuickRange,
  now = siteNow()
): [string, string] {
  const end = now.format('YYYY-MM-DD');
  if (key === 'd30') {
    return [now.subtract(1, 'month').format('YYYY-MM-DD'), end];
  }
  const offset = key === 'd7' ? 6 : 0;
  return [now.subtract(offset, 'day').format('YYYY-MM-DD'), end];
}

export function detectQuickRange(
  range: string[],
  now = siteNow()
): QuickRange | '' {
  if (!range[0] || !range[1]) return '';
  const keys: QuickRange[] = ['today', 'd7', 'd30'];
  return (
    keys.find((key) => {
      const [start, end] = buildQuickRange(key, now);
      return range[0] === start && range[1] === end;
    }) || ''
  );
}

export function toGameRow(
  record: GameReportRecord,
  excludeTest: boolean
): GameReportRow | null {
  const betCount = record.betCount - (excludeTest ? record.testBetCount : 0);
  const bettorCount =
    record.bettorCount - (excludeTest ? record.testBettorCount : 0);
  const validBet = record.validBet - (excludeTest ? record.testValidBet : 0);
  const memberWinLoss =
    record.memberWinLoss - (excludeTest ? record.testMemberWinLoss : 0);
  if (betCount <= 0 && validBet === 0 && memberWinLoss === 0) return null;
  return {
    date: record.date,
    currency: record.currency,
    betCount: Math.max(0, betCount),
    bettorCount: Math.max(0, bettorCount),
    validBet,
    memberWinLoss,
  };
}

export function matchGameDate(record: GameReportRecord, range: string[]) {
  if (!range[0] || !range[1]) return true;
  return record.date >= range[0] && record.date <= range[1];
}
