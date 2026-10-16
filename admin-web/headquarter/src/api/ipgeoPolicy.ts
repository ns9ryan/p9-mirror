import axios from 'axios';

/** /p9 接口不走 /admin baseURL */
const p9RequestConfig = { baseURL: '/' };

export interface IpgeoPolicyRecord {
  provider_id: number;
  name: string;
  is_enabled: boolean;
  priority: number;
  tag: string;
}

export interface IpgeoPolicyListRes {
  total: number;
  list: IpgeoPolicyRecord[];
}

/** 当前租户策略列表 POST /p9/integration/ipgeo/policy/list */
export function queryIpgeoPolicyList() {
  return axios.post<IpgeoPolicyListRes>(
    '/p9/integration/ipgeo/policy/list',
    {},
    p9RequestConfig
  );
}

export interface SaveIpgeoPolicyItem {
  provider_id: number;
  is_enabled: boolean;
  priority: number;
}

export interface SaveIpgeoPolicyParams {
  list: SaveIpgeoPolicyItem[];
}

/** 保存全局默认启用与优先级 POST /p9/integration/ipgeo/policy/default/save */
export function saveIpgeoPolicy(data: SaveIpgeoPolicyParams) {
  return axios.post<IpgeoPolicyListRes>(
    '/p9/integration/ipgeo/policy/default/save',
    data,
    p9RequestConfig
  );
}
