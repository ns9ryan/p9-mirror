import axios from 'axios';
import type { AxiosRequestConfig } from 'axios';

/** 统计粒度 */
export type TrendGranularity = 'day' | 'hour';

/** 留存天数，对应 retention{N}d / retention{N}dRate */
export const RETENTION_DAYS = [1, 2, 3, 7, 15, 30, 60] as const;
export type RetentionDay = (typeof RETENTION_DAYS)[number];

/**
 * register 注册数，login 登录数；
 * retention{N}d 留存人数，retention{N}dRate 留存率（百分比整数，0~100）
 */
export type MemberTrendType =
  | 'register'
  | 'login'
  | `retention${RetentionDay}d`
  | `retention${RetentionDay}dRate`;

export interface MemberTrendParams {
  /** YYYY-MM-DD */
  start_date: string;
  /** YYYY-MM-DD */
  end_date: string;
  granularity: TrendGranularity;
  type: MemberTrendType;
  /** 代理ID，后端必填，默认空字符串（查全部的取值待后端确认） */
  agent_id?: string;
  /** 来源，默认空字符串 */
  source?: string;
  /** 查询所用的 IANA 时区（如 Asia/Tokyo）；不传则由后端使用默认时区 */
  timezone?: string;
}

export interface MemberTrendPoint {
  /** YYYY-MM-DD 或 YYYY-MM-DD HH:mm */
  date: string;
  count: number;
}

/** 会员增长趋势 POST /report/dashboard/get-member-growth-trend */
export async function getMemberGrowthTrend(
  params: MemberTrendParams,
  silentError = true
) {
  const { timezone, ...rest } = params;
  const res = await axios.post<{ list?: MemberTrendPoint[] | null }>(
    '/report/dashboard/get-member-growth-trend',
    { agent_id: '', source: '', ...rest, ...(timezone ? { timezone } : {}) },
    { silentError } as AxiosRequestConfig
  );
  return res.data?.list || [];
}

/** 会员概览，增长率均为百分比数值（如 12.8 表示 12.8%） */
export interface MemberSummary {
  /** 会员总数（存量） */
  total_members: number;
  /** 较昨日会员增长比例 */
  members_growth_rate: number;
  /** 今日新增（今天 00:00 起） */
  today_new_members: number;
  /** 过去 24 小时新增 */
  last_24h_new_members: number;
  /** 过去 24 小时新增较前一个 24 小时的增长比例 */
  last_24h_new_growth_rate: number;
  /** 今日充值金额（元） */
  today_deposit_amount: number;
  /** 充值较昨日增长比例 */
  deposit_growth_rate: number;
  /** 活跃会员总数 */
  active_members: number;
  /** 今日活跃会员数 */
  today_active_members: number;
  /** 过去 24 小时活跃会员数 */
  last_24h_active_members: number;
  /** 过去 24 小时活跃较前一个 24 小时的增长比例 */
  last_24h_active_growth_rate: number;
}

/** 会员概览 GET /report/dashboard/get-member-summary */
export async function getMemberSummary(
  options: { agentId?: string; timezone?: string } = {},
  silentError = true
) {
  const { agentId = '', timezone } = options;
  const res = await axios.get<{ data?: MemberSummary | null }>(
    '/report/dashboard/get-member-summary',
    {
      params: { agent_id: agentId, ...(timezone ? { timezone } : {}) },
      silentError,
    } as AxiosRequestConfig
  );
  return res.data?.data || null;
}
