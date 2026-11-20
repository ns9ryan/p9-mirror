import { createEmptyAgent } from './mapper';
import type { AgentTreeNode } from './types';

/**
 * 开发环境压测：URL 加 ?treeStress=200
 * 会把某一层树节点补到指定条数，用于测「加载更多」截断渲染。
 * 假节点 id 为负数，点进详情可能失败，只用于测左侧树展示。
 */
export default function applyTreeStressPad(
  nodes: AgentTreeNode[],
  parentId: number
): AgentTreeNode[] {
  if (!import.meta.env.DEV) return nodes;
  const raw = new URLSearchParams(window.location.search).get('treeStress');
  const target = Number(raw);
  if (!Number.isFinite(target) || target <= 0) return nodes;
  if (nodes.length >= target) return nodes;

  const sample = nodes[0];
  const layer = sample?.layer || 1;
  const level = sample?.level || `${layer}层`;
  const padded = [...nodes];
  for (let i = nodes.length; i < target; i += 1) {
    padded.push({
      ...createEmptyAgent({
        id: -(parentId * 100000 + i + 1),
        account: `stress_${parentId || 'root'}_${String(i + 1).padStart(
          3,
          '0'
        )}`,
        layer,
        level,
        parentId,
        directAgents: 0,
        detailReady: false,
      }),
      depth: 0,
    });
  }
  return padded;
}
