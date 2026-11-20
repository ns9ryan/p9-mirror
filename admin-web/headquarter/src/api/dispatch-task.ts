import axios from 'axios';

/** 任务状态：1 待执行，2 执行中，3 成功，4 失败 */
export const DISPATCH_TASK_STATUS = {
  PENDING: 1,
  RUNNING: 2,
  SUCCESS: 3,
  FAILED: 4,
} as const;

export const DISPATCH_TASK_TYPE = {
  CREATE_OPERATOR: 'CREATE_OPERATOR',
  UPDATE_OPERATOR: 'UPDATE_OPERATOR',
  DELETE_OPERATOR: 'DELETE_OPERATOR',
} as const;

export interface DispatchTaskListParams {
  /** 页码，从 1 开始 */
  page: number;
  /** 每页数量 */
  page_size: number;
  /** 当前节点编码。节点详情里按节点筛选时传入 */
  node_code?: string;
  /** 任务编号或请求编号 */
  keyword?: string;
  /** 任务状态 */
  status?: number;
  /** 任务类型，如 CREATE_OPERATOR */
  task_type?: string;
}

export interface DispatchTaskRecord {
  task_no?: string;
  request_no?: string;
  target?: string;
  task_type?: string;
  /** JSON 字符串 */
  params?: string | null;
  node_code?: string;
  /** 1 待执行，2 执行中，3 成功，4 失败 */
  status?: number | string;
  /** JSON 字符串 */
  result?: string | null;
  error_message?: string | null;
  started_at?: number | null;
  finished_at?: number | null;
  created_at?: number | null;
  updated_at?: number | null;
}

export interface DispatchTaskListRes {
  total: number;
  list: DispatchTaskRecord[];
}

/** 调度任务分页列表 GET /admin/dispatch/dispatch-task/list（走 28001） */
export function queryDispatchTaskList(params: DispatchTaskListParams) {
  return axios.get<DispatchTaskListRes>('/dispatch/dispatch-task/list', {
    params,
  });
}
