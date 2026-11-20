import axios from 'axios';
import type { AxiosRequestConfig } from 'axios';
import {
  uploadResource,
  type PlatformUploadScene,
} from '@/api/resource';

/** avatar：默认头像；frame：默认头像框 */
export type AvatarMaterialType = 'avatar' | 'frame';

export interface AvatarMaterialRecord {
  /** 素材记录主键，修改时提交；不是会员 ID */
  id: number;
  /** 完整公开地址，"" 表示未配置 */
  image_url: string;
}

const API_PREFIX: Record<AvatarMaterialType, string> = {
  avatar: '/base/default-avatar',
  frame: '/base/default-avatar-frame',
};

const UPLOAD_SCENE: Record<AvatarMaterialType, PlatformUploadScene> = {
  avatar: 'default_avatar',
  frame: 'default_avatar_frame',
};

/** 查询全部默认头像 / 头像框 GET /admin/default-avatar(-frame)/list-all（走 18001） */
export function queryAvatarMaterialList(
  type: AvatarMaterialType,
  config: AxiosRequestConfig & { silentError?: boolean } = {}
) {
  return axios.get<{ list: AvatarMaterialRecord[] }>(
    `${API_PREFIX[type]}/list-all`,
    config
  );
}

/** 上传一张图片 POST /resource/upload；只上传，不修改任何槽位 */
export function uploadAvatarMaterial(type: AvatarMaterialType, file: File) {
  return uploadResource({
    file,
    scene: UPLOAD_SCENE[type],
  });
}

/** 按已有 ID 替换图片 POST /admin/default-avatar(-frame)/update；image_url 传上传返回的 key 或 url */
export function updateAvatarMaterial(
  type: AvatarMaterialType,
  data: { id: number; image_url: string }
) {
  return axios.post<Record<string, never>>(`${API_PREFIX[type]}/update`, data);
}
