import axios from 'axios';

/** 昵称去除首尾空白后最多 100 个 Unicode 字符 */
export const NICKNAME_MAX_LENGTH = 100;

/** 总站共用昵称库记录 */
export interface NicknameRecord {
  /** 昵称 ID */
  id: number;
  /** 昵称，库内唯一（区分大小写） */
  name: string;
}

export interface NicknameListParams {
  /** 页码，从 1 开始，默认 1 */
  page?: number;
  /** 每页数量，范围 1-100，默认 20 */
  page_size?: number;
  /** 按名称不区分大小写包含查询，最多 100 字符 */
  keyword?: string;
}

export interface NicknameListRes {
  total: number;
  list: NicknameRecord[];
}

/** 分页昵称列表（按 ID 升序） GET /admin/nickname/list（走 18001） */
export function queryNicknameList(params: NicknameListParams = {}) {
  return axios.get<NicknameListRes>('/base/nickname/list', { params });
}

/** 昵称详情 GET /admin/nickname/get（走 18001） */
export function getNickname(params: { id: number }) {
  return axios.get<{ nickname: NicknameRecord }>('/base/nickname/get', {
    params,
  });
}

/** 新增昵称 POST /admin/nickname/create（走 18001） */
export function createNickname(data: { name: string }) {
  return axios.post<{ nickname: NicknameRecord }>(
    '/base/nickname/create',
    data
  );
}

/** 修改昵称 POST /admin/nickname/update（走 18001） */
export function updateNickname(data: { id: number; name: string }) {
  return axios.post<Record<string, never>>('/base/nickname/update', data);
}

/** 物理删除昵称，不影响已创建会员 POST /admin/nickname/delete（走 18001） */
export function deleteNickname(data: { id: number }) {
  return axios.post<Record<string, never>>('/base/nickname/delete', data);
}
