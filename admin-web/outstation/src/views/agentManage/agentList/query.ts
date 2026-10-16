import dayjs from 'dayjs';
import { DEFAULT_CLIENT_HOST } from './options';
import type {
  AgentQuery,
  AgentRecord,
  AgentTreeNode,
  QuickRange,
} from './types';

export function createAgentQuery(): AgentQuery {
  return {
    timeType: 'register',
    dateRange: [],
    keywordType: 'id',
    keyword: '',
    registerSource: '',
    directMemberMin: undefined,
    directMemberMax: undefined,
    withdrawLimit: '',
    level: '',
    status: '',
    clientHost: DEFAULT_CLIENT_HOST,
  };
}

export function buildQuickRange(
  key: QuickRange,
  now = dayjs()
): [string, string] {
  const end = now.format('YYYY-MM-DD');
  const offsetByKey: Record<QuickRange, number> = {
    today: 0,
    d7: 6,
    d30: 29,
  };
  return [now.subtract(offsetByKey[key], 'day').format('YYYY-MM-DD'), end];
}

export function detectQuickRange(
  range: string[],
  now = dayjs()
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

function inDateRange(value: string, range: string[]) {
  if (!range[0] || !range[1]) return true;
  if (!value) return false;
  const time = dayjs(value);
  const start = dayjs(range[0]).startOf('day');
  const end = dayjs(range[1]).endOf('day');
  return !time.isBefore(start) && !time.isAfter(end);
}

export function matchAgent(record: AgentRecord, query: AgentQuery) {
  const timeValue =
    query.timeType === 'register' ? record.registeredAt : record.lastLoginAt;
  if (!inDateRange(timeValue, query.dateRange)) return false;

  const keyword = query.keyword.trim().toLowerCase();
  if (keyword) {
    const source =
      query.keywordType === 'id' ? String(record.id) : record.account;
    if (!source.toLowerCase().includes(keyword)) return false;
  }

  if (query.registerSource && record.registerSource !== query.registerSource) {
    return false;
  }
  if (
    query.directMemberMin != null &&
    record.directMembers < query.directMemberMin
  ) {
    return false;
  }
  if (
    query.directMemberMax != null &&
    record.directMembers > query.directMemberMax
  ) {
    return false;
  }
  if (query.withdrawLimit && record.withdrawLimit !== query.withdrawLimit) {
    return false;
  }
  if (query.level && record.level !== query.level) return false;
  if (query.status && record.status !== query.status) return false;
  if (query.clientHost && record.promoHost !== query.clientHost) return false;
  return true;
}

function parentKey(record: AgentRecord) {
  return record.parentAgent.trim().toLowerCase();
}

function dropCyclicParents(parentOf: Map<number, number>) {
  const state = new Map<number, 0 | 1 | 2>();
  const visit = (id: number) => {
    const mark = state.get(id) ?? 0;
    if (mark === 2) return;
    if (mark === 1) return;
    state.set(id, 1);
    const parentId = parentOf.get(id);
    if (parentId != null) {
      if (state.get(parentId) === 1) parentOf.delete(id);
      else visit(parentId);
    }
    state.set(id, 2);
  };
  [...parentOf.keys()].forEach(visit);
}

export function buildAgentTree(list: AgentRecord[]): AgentTreeNode[] {
  const byAccount = new Map<string, AgentRecord>();
  list.forEach((item) => {
    const key = item.account.trim().toLowerCase();
    if (key && !byAccount.has(key)) byAccount.set(key, item);
  });

  const parentOf = new Map<number, number>();
  list.forEach((item) => {
    const parent = byAccount.get(parentKey(item));
    if (!parent || parent.id === item.id) return;
    parentOf.set(item.id, parent.id);
  });
  dropCyclicParents(parentOf);

  const childrenOf = new Map<number, AgentRecord[]>();
  const roots: AgentRecord[] = [];
  list.forEach((item) => {
    const parentId = parentOf.get(item.id);
    if (parentId == null) {
      roots.push(item);
      return;
    }
    const bucket = childrenOf.get(parentId) ?? [];
    bucket.push(item);
    childrenOf.set(parentId, bucket);
  });

  const toNode = (item: AgentRecord, depth: number): AgentTreeNode => {
    const children = childrenOf.get(item.id);
    if (!children?.length) return { ...item, depth };
    return {
      ...item,
      depth,
      children: children.map((child) => toNode(child, depth + 1)),
    };
  };

  return roots.map((item) => toNode(item, 0));
}

/** 展示到指定层级（从 1 开始）：保留该层和上面的上级，去掉更深层的下级。 */
export function pickAgentLayer(
  nodes: AgentTreeNode[],
  layer: number
): AgentTreeNode[] {
  if (layer < 1) return nodes;
  const trim = (list: AgentTreeNode[]): AgentTreeNode[] =>
    list.flatMap((node) => {
      const current = node.depth + 1;
      if (current > layer) return [];
      if (current === layer || !node.children?.length) {
        return [{ ...node, children: undefined }];
      }
      return [{ ...node, children: trim(node.children) }];
    });
  return trim(nodes);
}

export function flattenAgentTree(nodes: AgentTreeNode[]): AgentTreeNode[] {
  return nodes.flatMap((node) => [
    node,
    ...(node.children ? flattenAgentTree(node.children) : []),
  ]);
}

function findAgentNode(
  nodes: AgentTreeNode[],
  id: number
): AgentTreeNode | null {
  return nodes.reduce<AgentTreeNode | null>((found, node) => {
    if (found) return found;
    if (node.id === id) return node;
    if (!node.children?.length) return null;
    return findAgentNode(node.children, id);
  }, null);
}

/** 节点自身加上全部下级，供单选时联动勾选。 */
export function agentSubtreeIds(nodes: AgentTreeNode[], id: number): number[] {
  const target = findAgentNode(nodes, id);
  if (!target) return [id];
  return flattenAgentTree([target]).map((item) => item.id);
}

export function agentParentKeys(nodes: AgentTreeNode[]): number[] {
  return nodes.flatMap((node) =>
    node.children?.length ? [node.id, ...agentParentKeys(node.children)] : []
  );
}
