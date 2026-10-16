import axios from 'axios';

/** 操作结果：后端未提供明确枚举，按常见语义处理 / 0 全部 / 1 成功 / 2 失败 */
export const LOG_ACTION_RESULT = {
  ALL: 0,
  SUCCESS: 1,
  FAILED: 2,
} as const;

export interface OperationLogRecord {
  id: number;
  user_id: number;
  username: string;
  client_ip: string;
  request_method: string;
  request_path: string;
  request_query: string;
  request_body: string;
  response_status: number;
  response_body: string;
  /** 操作结果，后端以字符串返回，如 success / fail */
  action_result: string;
  duration_ms: number;
  user_agent: string;
  created_at: number;
}

export interface OperationLogListParams {
  /** 页码 */
  page?: number;
  /** 每页条数 */
  page_size?: number;
  /** 用户 ID */
  user_id?: number;
  /** 用户名模糊搜索 */
  username?: string;
  /** GET / POST / PUT / PATCH / DELETE */
  request_method?: string;
  /** 请求路径模糊搜索 */
  request_path?: string;
  /** 操作结果：0 全部 / 1 成功 / 2 失败 */
  action_result?: number;
  /** 创建时间起始（秒级时间戳） */
  created_at_from?: number;
  /** 创建时间结束（秒级时间戳） */
  created_at_to?: number;
}

export interface OperationLogListRes {
  list: OperationLogRecord[];
  total: number;
}

/** 操作日志列表 POST /admin/core/log/action/list */
export function queryOperationLogList(data: OperationLogListParams = {}) {
  return axios.post<OperationLogListRes>('/core/log/action/list', data);
}

export interface LoginLogRecord {
  id: number;
  user_id: number;
  username: string;
  login_ip: string;
  device_id: number;
  /** 登录结果，后端以字符串返回，如 success / fail */
  login_result: string;
  /** 失败原因，登录成功时通常为空 */
  failure_reason: string;
  user_agent: string;
  login_at: number;
}

export interface LoginLogListParams {
  /** 页码 */
  page?: number;
  /** 每页条数 */
  page_size?: number;
  /** 用户 ID */
  user_id?: number;
  /** 用户名模糊搜索 */
  username?: string;
  /** 登录结果：0 全部 / 1 成功 / 2 失败 */
  login_result?: number;
  /** 登录时间起始（秒级时间戳） */
  login_at_from?: number;
  /** 登录时间结束（秒级时间戳） */
  login_at_to?: number;
}

export interface LoginLogListRes {
  list: LoginLogRecord[];
  total: number;
}

/** 登录日志列表 POST /admin/core/log/login/list */
export function queryLoginLogList(data: LoginLogListParams = {}) {
  return axios.post<LoginLogListRes>('/core/log/login/list', data);
}

export interface ErrorLogRecord {
  id: number;
  service_name: string;
  request_method: string;
  request_path: string;
  request_query: string;
  request_body: string;
  response_status: number;
  response_body: string;
  duration_ms: number;
  client_ip: string;
  user_id: number;
  username: string;
  operator_id: number;
  subject: string;
  detail: string;
  user_agent: string;
  created_at: number;
}

export interface ErrorLogListParams {
  /** 页码 */
  page?: number;
  /** 每页条数 */
  page_size?: number;
  /** 用户 ID */
  user_id?: number;
  /** 服务名 */
  service_name?: string;
  /** 请求路径模糊搜索 */
  request_path?: string;
  /** 响应状态码 */
  response_status?: number;
  /** 创建时间起始（秒级时间戳） */
  created_at_from?: number;
  /** 创建时间结束（秒级时间戳） */
  created_at_to?: number;
}

export interface ErrorLogListRes {
  list: ErrorLogRecord[];
  total: number;
}

/** 错误日志列表 POST /admin/core/log/error/list */
export function queryErrorLogList(data: ErrorLogListParams = {}) {
  return axios.post<ErrorLogListRes>('/core/log/error/list', data);
}
