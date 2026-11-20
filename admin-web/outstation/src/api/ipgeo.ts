import axios from 'axios';

/** /p9 接口不走 /admin baseURL */
const p9RequestConfig = { baseURL: '/' };

export interface IpgeoLocalRecord {
  id: number;
  ip: string;
  isp: string;
  as: string;
  org: string;
  country: string;
  country_code: string;
  continent: string;
  continent_code: string;
  province: string;
  province_iso_code: string;
  city: string;
  zip: string;
  postal: string;
  lat: number;
  lon: number;
  timezone: string;
}

export interface IpgeoLocalListParams {
  /** IP 筛选 */
  ip?: string;
  /** 页码，从 1 开始 */
  page?: number;
  /** 每页条数 */
  page_size?: number;
}

export interface IpgeoLocalListRes {
  total: number;
  list: IpgeoLocalRecord[];
}

/** 本地 IP 表分页列表 POST /p9/integration/ipgeo/local/list */
export function queryIpgeoLocalList(data: IpgeoLocalListParams = {}) {
  return axios.post<IpgeoLocalListRes>(
    '/p9/integration/ipgeo/local/list',
    data,
    p9RequestConfig
  );
}

export type CreateIpgeoLocalParams = Omit<IpgeoLocalRecord, 'id'> & {
  id?: number;
};

/** 新增本地 IP POST /p9/integration/ipgeo/local/create */
export function createIpgeoLocal(data: CreateIpgeoLocalParams) {
  return axios.post<IpgeoLocalRecord>(
    '/p9/integration/ipgeo/local/create',
    data,
    p9RequestConfig
  );
}

export type UpdateIpgeoLocalParams = IpgeoLocalRecord;

/** 更新本地 IP POST /p9/integration/ipgeo/local/update（必填 id） */
export function updateIpgeoLocal(data: UpdateIpgeoLocalParams) {
  return axios.post<IpgeoLocalRecord>(
    '/p9/integration/ipgeo/local/update',
    data,
    p9RequestConfig
  );
}

export interface DeleteIpgeoLocalParams {
  id: number;
}

/** 删除本地 IP POST /p9/integration/ipgeo/local/delete */
export function deleteIpgeoLocal(data: DeleteIpgeoLocalParams) {
  return axios.post<{ result: string }>(
    '/p9/integration/ipgeo/local/delete',
    data,
    p9RequestConfig
  );
}

export interface IpgeoLocalDetailParams {
  id: number;
}

/** 本地 IP 详情 GET /p9/integration/ipgeo/local/detail */
export function getIpgeoLocalDetail(params: IpgeoLocalDetailParams) {
  return axios.get<IpgeoLocalRecord>(
    '/p9/integration/ipgeo/local/detail',
    { ...p9RequestConfig, params }
  );
}

export interface IpgeoIpSearchParams {
  /** 待查询 IP 列表 */
  ips: string[];
  /** 语言，如 zh-CN / en-US */
  lang: string;
}

export interface IpgeoIpSearchRecord {
  query: string;
  country: string;
  country_code: string;
  province: string;
  city: string;
  isp: string;
  org: string;
  as: string;
  lat: number;
  lon: number;
  continent: string;
  continent_code: string;
  timezone: string;
  zip: string;
  postal: string;
  source: string;
}

export interface IpgeoIpSearchRes {
  list: IpgeoIpSearchRecord[];
}

/** 批量查询 IP 地理信息 POST /p9/integration/ipgeo/ip/search */
export function searchIpgeoIp(data: IpgeoIpSearchParams) {
  return axios.post<IpgeoIpSearchRes>(
    '/p9/integration/ipgeo/ip/search',
    data,
    p9RequestConfig
  );
}
