import axios from 'axios';

/**
 * 统一图片上传场景（POST /resource/upload）
 * 分站入口仅支持 vipicon、logo；总站场景在此预留类型
 */
export type ResourceUploadScene =
  | 'vipicon'
  | 'default_avatar'
  | 'default_avatar_frame'
  | 'logo';

/** 上传成功返回；只代表 OSS 已创建，不会自动写入业务配置 */
export interface ResourceUploadResult {
  /** 完整永久公开 URL，用于预览 */
  url: string;
  /** 桶内对象 Key；VIP / 头像类保存用这个 */
  key: string;
  scene: string;
  /** 固定 public */
  visibility: string;
  /** 固定 0，表示永久公开地址 */
  url_expires_in: number;
}

export interface UploadResourceParams {
  file: File;
  scene: ResourceUploadScene;
  /**
   * 仅 vipicon 必填：0 / 10 / … / 90
   * 分站保存时用 group_no，换算为 min_level = group_no × 10
   */
  min_level?: number;
}

/**
 * 统一图片上传 POST /resource/upload
 * form-data：VIP 恰好 file + scene + min_level；其余恰好 file + scene
 */
export function uploadResource(params: UploadResourceParams) {
  const data = new FormData();
  data.append('file', params.file);
  data.append('scene', params.scene);
  if (params.scene === 'vipicon') {
    data.append('min_level', String(params.min_level ?? 0));
  }
  return axios.post<ResourceUploadResult>('/resource/upload', data);
}
