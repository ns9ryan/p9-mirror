import axios from 'axios';

/** /p9 接口不走 /admin baseURL */
const p9RequestConfig = { baseURL: '/' };

/** 上传 / 换链共用返回 */
export interface OssFileResult {
  url: string;
  key: string;
  scene: string;
  visibility: string;
  url_expires_in?: number;
}

export interface OssPresignResult extends OssFileResult {
  put_url: string;
  method: string;
  headers: Record<string, string>;
  expires_in: number;
}

export interface OssFileRecord {
  id: number;
  key: string;
  scene: string;
  bucket: string;
  name: string;
  size: number;
  content_type: string;
  url: string;
  visibility: string;
  url_expires_in?: number;
  created_at: number;
}

export interface OssFileListRes {
  list: OssFileRecord[];
  total: number;
}

export interface OssProvider {
  id?: number;
  name: string;
  is_default: boolean;
  driver: string;
  endpoint: string;
  file_url: string;
  access_key: string;
  secret_key?: string;
  path_style: boolean;
  region: string;
  buckets?: Record<string, { mode: string }>;
  source?: string;
}

/** 小文件中转上传 POST /p9/oss/file/upload */
export function uploadOssFile(data: FormData) {
  return axios.post<OssFileResult>('/p9/oss/file/upload', data, p9RequestConfig);
}

/** 用 key 换可打开 url GET /p9/oss/url/detail */
export function getOssUrlDetail(params: { key?: string; url?: string }) {
  return axios.get<OssFileResult>('/p9/oss/url/detail', {
    ...p9RequestConfig,
    params,
  });
}

/** 大文件第 1 步：签发 PUT POST /p9/oss/upload/presign */
export function presignOssUpload(data: {
  filename: string;
  size: number;
  scene?: string;
}) {
  return axios.post<OssPresignResult>(
    '/p9/oss/upload/presign',
    data,
    p9RequestConfig
  );
}

/** 大文件第 3 步：确认完成 POST /p9/oss/upload/complete */
export function completeOssUpload(data: { key: string; scene?: string }) {
  return axios.post<OssFileResult>(
    '/p9/oss/upload/complete',
    data,
    p9RequestConfig
  );
}

/**
 * 大文件第 2 步：直传存储（PUT put_url）。
 * 不能走业务 axios（会带 Authorization / X-Lang，破坏预签名）。
 */
export async function putOssObject(
  putUrl: string,
  file: Blob,
  headers: Record<string, string> = {}
) {
  const res = await fetch(putUrl, {
    method: 'PUT',
    headers,
    body: file,
  });
  if (!res.ok) {
    const text = await res.text().catch(() => '');
    throw new Error(text || `PUT failed: HTTP ${res.status}`);
  }
  return res;
}

/** 档案列表 POST /p9/oss/file/list */
export function queryOssFileList(data: {
  scene?: string;
  page?: number;
  page_size?: number;
} = {}) {
  return axios.post<OssFileListRes>('/p9/oss/file/list', data, p9RequestConfig);
}

/** 删对象 + 档案 POST /p9/oss/file/delete */
export function deleteOssFile(data: { key: string }) {
  return axios.post<OssFileResult>('/p9/oss/file/delete', data, p9RequestConfig);
}

/** 当前生效服务商 GET /p9/oss/provider/active */
export function getOssProviderActive() {
  return axios.get<OssProvider>('/p9/oss/provider/active', p9RequestConfig);
}
