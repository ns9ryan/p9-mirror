import {
  AGENT_PAYOUT,
  AGENT_STATUS,
  type AgentListItem,
  type AgentTreeNodeItem,
} from '@/api/agent';
import { formatSiteTime } from '@/utils/site-time';
import type {
  AccountStatus,
  AgentRecord,
  AgentTreeNode,
  WithdrawLimit,
} from './types';

function mapStatus(status: number): AccountStatus {
  return status === AGENT_STATUS.DISABLED ? 'disabled' : 'normal';
}

function mapPayout(value: number): WithdrawLimit {
  if (value === AGENT_PAYOUT.FORBIDDEN) return 'forbidden';
  if (value === AGENT_PAYOUT.NONE) return 'none';
  return 'unset';
}

export function formatUnix(value: number) {
  return formatSiteTime(value, undefined, '');
}

export function createEmptyAgent(
  partial: Partial<AgentRecord> & Pick<AgentRecord, 'id' | 'account'>
): AgentRecord {
  return {
    level: '',
    layer: 0,
    parentId: 0,
    parentAgent: '',
    topAgent: '',
    registerSource: '',
    directMembers: 0,
    directAgents: 0,
    allMembers: 0,
    allAgents: 0,
    shareRatio: 0,
    balance: 0,
    totalCommission: 0,
    unpaidCommission: 0,
    status: 'normal',
    withdrawLimit: 'unset',
    remark: '',
    promoCode: '',
    promoHost: '',
    visitUv: 0,
    registeredAt: '',
    lastLoginAt: '',
    detailReady: false,
    ...partial,
  };
}

export function mapAgentListItem(item: AgentListItem): AgentRecord {
  return createEmptyAgent({
    id: item.id,
    account: item.account,
    level: item.level_alias || String(item.level || ''),
    layer: item.level || 0,
    parentId: item.parent_id || 0,
    parentAgent: item.parent_account || '',
    topAgent: item.top_account || '',
    registerSource: item.register_source ? String(item.register_source) : '',
    directMembers: item.direct_member_count || 0,
    directAgents: item.direct_agent_count || 0,
    allMembers: item.total_member_count || 0,
    allAgents: item.total_agent_count || 0,
    shareRatio: item.share_rate || 0,
    balance: item.wallet_balance || 0,
    totalCommission: item.total_commission || 0,
    unpaidCommission: item.unpaid_commission || 0,
    status: mapStatus(item.status),
    withdrawLimit: mapPayout(item.payout_restricted),
    remark: item.remark || '',
    promoCode: item.promo_link || '',
    visitUv: item.visit_uv || 0,
    registeredAt: formatUnix(item.register_at),
    lastLoginAt: formatUnix(item.last_login_at),
    detailReady: true,
  });
}

export function mapTreeNode(
  item: AgentTreeNodeItem,
  parentId = 0
): AgentTreeNode {
  return {
    ...createEmptyAgent({
      id: item.id,
      account: item.account,
      level: item.level_alias || String(item.level || ''),
      layer: item.level || 0,
      parentId,
      directAgents: item.direct_agent_count || 0,
      lineCode: item.line_code || '',
      detailReady: false,
    }),
    depth: 0,
  };
}
