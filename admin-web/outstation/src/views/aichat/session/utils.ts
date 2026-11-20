import dayjs from 'dayjs';
import type { ChatContentType, ChatRole } from '@/api/aichat';
import { getSiteTimezone, siteNow } from '@/utils/site-time';

export interface ChatUiMessage {
  key: string;
  role: ChatRole;
  content: string;
  contentType?: ChatContentType;
  /** 毫秒时间戳 */
  time: number;
  isNeedConfirm: boolean;
  /** 仅本地发送中的用户消息使用 */
  status?: 'sending' | 'failed';
  /** 失败重试时复用，保证幂等 */
  requestId?: string;
  latencyMs?: number;
  modelName?: string;
  totalTokens?: number;
}

export type TableCell = string | number;

export interface TableData {
  columns: string[];
  rows: Record<string, TableCell>[];
}

export type ContentBlock =
  | { type: 'text'; text: string }
  | { type: 'code'; lang: string; code: string }
  | ({ type: 'table' } & TableData);

const SCENE_ICONS: Record<string, string> = {
  chart: 'icon-bar-chart',
  data: 'icon-bar-chart',
  analysis: 'icon-bar-chart',
  service: 'icon-customer-service',
  customer: 'icon-customer-service',
  knowledge: 'icon-book',
  book: 'icon-book',
  faq: 'icon-question-circle',
  sql: 'icon-storage',
  code: 'icon-code',
};

export function sceneIcon(icon?: string) {
  return (icon && SCENE_ICONS[icon.toLowerCase()]) || 'icon-robot';
}

export function createRequestId(): string {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
    return crypto.randomUUID();
  }
  // 非安全上下文（http + IP 访问）下没有 randomUUID
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = Math.floor(Math.random() * 16);
    return (c === 'x' ? r : (r % 4) + 8).toString(16);
  });
}

function siteDayjs(ms: number) {
  const tz = getSiteTimezone() || 'UTC';
  return dayjs(ms).tz(tz);
}

export function formatListTime(ms: number) {
  if (!ms) return '';
  const d = siteDayjs(ms);
  const now = siteNow();
  if (d.isSame(now, 'day')) return d.format('HH:mm');
  if (d.isSame(now, 'year')) return d.format('MM/DD');
  return d.format('YYYY/MM/DD');
}

export function formatMessageTime(ms: number) {
  if (!ms) return '';
  const d = siteDayjs(ms);
  return d.isSame(siteNow(), 'day')
    ? d.format('HH:mm')
    : d.format('YYYY-MM-DD HH:mm');
}

function toCell(value: unknown): TableCell {
  if (value === null || value === undefined) return '';
  if (typeof value === 'number' || typeof value === 'string') return value;
  return JSON.stringify(value);
}

function tableFromJson(value: unknown): TableData | null {
  if (!Array.isArray(value) || !value.length) return null;
  if (!value.every((v) => v && typeof v === 'object' && !Array.isArray(v))) {
    return null;
  }
  const columns: string[] = [];
  value.forEach((row) => {
    Object.keys(row).forEach((k) => {
      if (!columns.includes(k)) columns.push(k);
    });
  });
  const rows = value.map((row) => {
    const record: Record<string, TableCell> = {};
    columns.forEach((c) => {
      record[c] = toCell((row as Record<string, unknown>)[c]);
    });
    return record;
  });
  return { columns, rows };
}

function tryParseJson(text: string): unknown {
  const s = text.trim();
  if (!/^[[{]/.test(s)) return undefined;
  try {
    return JSON.parse(s);
  } catch {
    return undefined;
  }
}

function splitRow(line: string) {
  return line
    .trim()
    .replace(/^\|/, '')
    .replace(/\|$/, '')
    .split('|')
    .map((c) => c.trim());
}

const TABLE_SEPARATOR = /^\s*\|?\s*:?-{3,}:?\s*(\|\s*:?-{3,}:?\s*)*\|?\s*$/;

/** 把文本中的 Markdown 表格拆出来，其余保持纯文本 */
function parseTextWithTables(text: string): ContentBlock[] {
  const lines = text.split('\n');
  const blocks: ContentBlock[] = [];
  let buffer: string[] = [];
  const flush = () => {
    const joined = buffer.join('\n').trim();
    if (joined) blocks.push({ type: 'text', text: joined });
    buffer = [];
  };

  let i = 0;
  while (i < lines.length) {
    const line = lines[i];
    if (
      line.trim().startsWith('|') &&
      i + 1 < lines.length &&
      TABLE_SEPARATOR.test(lines[i + 1])
    ) {
      flush();
      const columns = splitRow(line);
      const rows: Record<string, TableCell>[] = [];
      i += 2;
      while (i < lines.length && lines[i].trim().startsWith('|')) {
        const cells = splitRow(lines[i]);
        const record: Record<string, TableCell> = {};
        columns.forEach((c, idx) => {
          record[c] = cells[idx] ?? '';
        });
        rows.push(record);
        i += 1;
      }
      blocks.push({ type: 'table', columns, rows });
    } else {
      buffer.push(line);
      i += 1;
    }
  }
  flush();
  return blocks;
}

export function parseContent(
  content: string,
  contentType?: ChatContentType
): ContentBlock[] {
  const json = tryParseJson(content);
  if (json !== undefined) {
    const table = tableFromJson(json);
    if (table) return [{ type: 'table', ...table }];
    return [
      { type: 'code', lang: 'json', code: JSON.stringify(json, null, 2) },
    ];
  }
  if (contentType === 'json') {
    return [{ type: 'code', lang: 'json', code: content }];
  }

  const blocks: ContentBlock[] = [];
  const fence = /```([\w-]*)[^\n]*\n?([\s\S]*?)```/g;
  let last = 0;
  let match = fence.exec(content);
  while (match) {
    blocks.push(...parseTextWithTables(content.slice(last, match.index)));
    const lang = match[1].toLowerCase();
    const code = match[2].replace(/\n$/, '');
    const table = lang === 'json' ? tableFromJson(tryParseJson(code)) : null;
    blocks.push(
      table ? { type: 'table', ...table } : { type: 'code', lang, code }
    );
    last = match.index + match[0].length;
    match = fence.exec(content);
  }
  blocks.push(...parseTextWithTables(content.slice(last)));
  return blocks;
}

export function toNumber(cell: TableCell): number | null {
  if (typeof cell === 'number') return Number.isFinite(cell) ? cell : null;
  const s = cell.replace(/,/g, '').trim();
  if (!s) return null;
  const n = Number(s);
  return Number.isFinite(n) ? n : null;
}

/** 首个非数值列作 X 轴，全数值列作折线；不满足条件返回 null */
export function chartableColumns(table: TableData) {
  if (table.rows.length < 2) return null;
  const numeric = table.columns.filter((c) =>
    table.rows.every((r) => toNumber(r[c]) !== null)
  );
  const xColumn = table.columns.find((c) => !numeric.includes(c));
  if (!xColumn || !numeric.length) return null;
  return { xColumn, series: numeric };
}
