import axios from 'axios';

/** 1 启用 / 2 停用 */
export const NODE_STATUS = {
  ENABLED: 1,
  DISABLED: 2,
} as const;

/** 1 在线 / 2 离线 */
export const NODE_ONLINE_STATUS = {
  ONLINE: true,
  OFFLINE: false,
} as const;

export interface NodeRecord {
  /** 节点 ID */
  id: number;
  /** 节点编码 */
  code: string;
  /** 节点名称 */
  name: string;
  /** 1 启用 / 2 停用 */
  status: number;
  /** 是否在线 */
  online: boolean;
  /** 已分配分站数 */
  assigned_operator_count: number;
  /** 最近活动时间，unix 秒；没有则为 null */
  last_active_at: number | null;
  /** 备注 */
  remark: string;
}

export interface NodeListParams {
  /** 页码，从 1 开始 */
  page: number;
  /** 每页数量，范围 1-100 */
  page_size: number;
  /** 搜索关键字，匹配节点名称或节点编码 */
  keyword?: string;
}

export interface NodeListRes {
  total: number;
  list: NodeRecord[];
}

/** 部署节点分页列表 GET /admin/dispatch/node/list（走 28001） */
export function queryNodeList(params: NodeListParams) {
  return axios.get<NodeListRes>('/dispatch/node/list', { params });
}

export interface CreateNodeParams {
  /** 节点名称，非空，最大 100 个字符 */
  name: string;
  /** 1 启用 / 2 停用；不传默认 1 */
  status?: number;
  /** 运维备注，最大 1000 个字符 */
  remark?: string;
}

export interface CreateNodeRes {
  /** 节点 ID */
  id: number;
  /** 后端生成的节点业务编码 */
  code: string;
  /**
   * 节点认证密钥明文，仅本次创建成功时返回。
   * 后端只保存哈希，关闭弹窗后前端不得继续保存或展示。
   */
  auth_secret: string;
}

/** 创建节点 POST /admin/dispatch/node/create（走 28001） */
export function createNode(data: CreateNodeParams) {
  return axios.post<CreateNodeRes>('/dispatch/node/create', data);
}

export interface UpdateNodeParams {
  /** 节点 ID，大于 0 */
  id: number;
  /** 节点名称，非空，最大 100 个字符 */
  name?: string;
  /** 1 启用 / 2 停用 */
  status?: number;
  /** 运维备注，最大 1000 个字符 */
  remark?: string;
}

/** 修改节点 POST /admin/dispatch/node/update（走 28001）。除 id 外至少传一个字段 */
export function updateNode(data: UpdateNodeParams) {
  return axios.post<Record<string, never>>('/dispatch/node/update', data);
}

export interface ResetNodeAuthSecretParams {
  /** 节点 ID，大于 0 */
  id: number;
}

export interface ResetNodeAuthSecretRes {
  /**
   * 新认证密钥明文，仅本次重置成功时返回。
   * 原密钥立即失效，关闭弹窗后前端不得继续保存或展示。
   */
  auth_secret: string;
}

/** 重置节点认证密钥 POST /admin/dispatch/node/reset-auth-secret（走 28001） */
export function resetNodeAuthSecret(data: ResetNodeAuthSecretParams) {
  return axios.post<ResetNodeAuthSecretRes>('/dispatch/node/reset-auth-secret', data);
}

export interface NodeDetail {
  /** 节点 ID */
  id: number;
  /** 节点编码 */
  code: string;
  /** 节点名称 */
  name: string;
  /** 1 启用 / 2 停用 */
  status: number;
  /** 是否在线 */
  online: boolean;
  /** 最近在线时间，unix 秒；没有则为 null */
  last_seen_at: number | null;
  /** 运维备注 */
  remark: string;
  /** 创建时间，unix 秒 */
  created_at: number;
  /** 更新时间，unix 秒 */
  updated_at: number;
}

/** 获取节点详情 GET /admin/dispatch/node/get（走 28001）。data 直接为节点信息 */
export function queryNodeDetail(id: number) {
  return axios.get<NodeDetail>('/dispatch/node/get', { params: { id } });
}
