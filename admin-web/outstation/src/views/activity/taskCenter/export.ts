import dayjs from 'dayjs';
import type { TaskConfigItem, TaskRewardItem } from '@/api/task';
import { displayText, formatAmount, formatTime } from './format';

type Translate = (key: string) => string;

const EXPORT_PAGE_SIZE = 100;

/** 按接口最大分页逐页拉取当前筛选下的全部数据 */
export async function fetchAllPages<T>(
  fetchPage: (
    page: number,
    pageSize: number
  ) => Promise<{ total: number; list: T[] }>
) {
  const all: T[] = [];
  let page = 1;
  // eslint-disable-next-line no-constant-condition
  while (true) {
    // eslint-disable-next-line no-await-in-loop
    const { total, list } = await fetchPage(page, EXPORT_PAGE_SIZE);
    all.push(...list);
    if (!list.length || all.length >= total) return all;
    page += 1;
  }
}

function csvCell(value: string | number) {
  const text = String(value ?? '');
  if (/[",\n]/.test(text)) return `"${text.replace(/"/g, '""')}"`;
  return text;
}

function downloadCsv(
  prefix: string,
  headers: string[],
  rows: Array<Array<string | number>>
) {
  const csv = [headers, ...rows]
    .map((cols) => cols.map((col) => csvCell(col)).join(','))
    .join('\n');
  const blob = new Blob([`\uFEFF${csv}`], {
    type: 'text/csv;charset=utf-8;',
  });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = `${prefix}-${dayjs().format('YYYYMMDD-HHmmss')}.csv`;
  link.click();
  URL.revokeObjectURL(url);
}

export default function exportTasksCsv(list: TaskConfigItem[], t: Translate) {
  const headers = [
    t('discountCenter.taskCenter.column.taskId'),
    t('discountCenter.taskCenter.column.taskCode'),
    t('discountCenter.taskCenter.column.name'),
    t('discountCenter.taskCenter.column.reward'),
    t('discountCenter.taskCenter.column.completionAmount'),
    t('discountCenter.taskCenter.column.category'),
    t('discountCenter.taskCenter.column.description'),
    t('discountCenter.taskCenter.column.payoutMode'),
    t('discountCenter.taskCenter.column.timezone'),
    t('discountCenter.taskCenter.column.enabled'),
    t('discountCenter.taskCenter.column.sortValue'),
    t('discountCenter.taskCenter.column.updatedAt'),
  ];
  const rows = list.map((item) => [
    item.id,
    item.task_code,
    item.name,
    formatAmount(item.reward_amount),
    formatAmount(item.completion_amount),
    t(`discountCenter.taskCenter.category.${item.category}`),
    displayText(item.description),
    t(`discountCenter.taskCenter.payoutMode.${item.payout_mode}`),
    displayText(item.timezone),
    t(`discountCenter.taskCenter.status.${item.status}`),
    item.sort,
    formatTime(item.updated_at),
  ]);
  downloadCsv('tasks', headers, rows);
}

export function exportClaimsCsv(
  list: TaskRewardItem[],
  taskNames: Map<string, string>,
  t: Translate
) {
  const headers = [
    t('discountCenter.taskCenter.claims.column.orderNo'),
    t('discountCenter.taskCenter.claims.column.memberId'),
    t('discountCenter.taskCenter.column.name'),
    t('discountCenter.taskCenter.column.category'),
    t('discountCenter.taskCenter.claims.column.periodKey'),
    t('discountCenter.taskCenter.claims.column.amount'),
    t('discountCenter.taskCenter.column.payoutMode'),
    t('discountCenter.taskCenter.claims.column.status'),
    t('discountCenter.taskCenter.claims.column.claimIp'),
    t('discountCenter.taskCenter.claims.column.deviceId'),
    t('discountCenter.taskCenter.claims.column.walletId'),
    t('discountCenter.taskCenter.claims.column.transactionId'),
    t('discountCenter.taskCenter.claims.column.createdAt'),
    t('discountCenter.taskCenter.claims.column.payoutAt'),
  ];
  // 前置制表符防止 Excel 把长数字串显示成科学计数法
  const rows = list.map((item) => [
    `\t${item.order_no}`,
    `\t${item.member_id}`,
    taskNames.get(item.task_code) || item.task_code,
    t(`discountCenter.taskCenter.category.${item.category}`),
    displayText(item.period_key),
    formatAmount(item.amount),
    t(`discountCenter.taskCenter.payoutMode.${item.payout_mode}`),
    t(`discountCenter.taskCenter.claims.status.${item.status}`),
    displayText(item.claim_ip),
    displayText(item.device_id),
    displayText(item.wallet_id),
    `\t${displayText(item.wallet_transaction_id)}`,
    formatTime(item.created_at),
    formatTime(item.payout_at),
  ]);
  downloadCsv('task-claims', headers, rows);
}
