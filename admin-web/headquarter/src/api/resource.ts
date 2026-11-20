import axios from 'axios';
import type { OssFileResult } from '@/api/oss';

/** 总站统一上传场景 */
export type PlatformUploadScene =
  | 'vipicon'
  | 'default_avatar'
  | 'default_avatar_frame';

export interface UploadResourceParams {
  file: File;
  scene: PlatformUploadScene;
  /** 仅 vipicon 必填：0、10、…、90；其他场景禁止传 */
  minLevel?: number;
}

/**
 * 总站统一图片上传 POST /resource/upload
 * 走默认 /admin baseURL，开发代理将 /admin/resource 改写为 /resource
 * 只创建 OSS 对象，不写入业务配置；保存仍走原接口
 */
export function uploadResource(params: UploadResourceParams) {
  const form = new FormData();
  form.append('file', params.file);
  form.append('scene', params.scene);
  if (params.scene === 'vipicon') {
    if (params.minLevel == null) {
      throw new Error('vipicon 上传必须提交 min_level');
    }
    form.append('min_level', String(params.minLevel));
  }
  return axios.post<OssFileResult>('/resource/upload', form);
}
