import axios from 'axios';
import { uploadResource } from '@/api/resource';

export interface VipDefaultIconGroup {
  /** 等级下限（含），只能为 0、10、…、90 */
  min_level: number;
  /** 等级上限（含），等于 min_level + 9 */
  max_level: number;
  /** 完整公开地址，"" 表示未配置或已清空 */
  icon_url: string;
}

/** 查询固定十组 VIP 默认图标 GET /admin/vip-default-icon/list-all（走 18001） */
export function queryVipDefaultIconList() {
  return axios.get<{ list: VipDefaultIconGroup[] }>(
    '/base/vip-default-icon/list-all'
  );
}

/** 为指定等级组上传图标 POST /resource/upload；只上传，尚未保存 */
export function uploadVipDefaultIcon(minLevel: number, file: File) {
  return uploadResource({
    file,
    scene: 'vipicon',
    minLevel,
  });
}

/**
 * 保存等级组 POST /admin/vip-default-icon/save
 * 只提交要修改的组且三个字段都要带上；漏传 icon_url 会被当成清空
 */
export function saveVipDefaultIcons(list: VipDefaultIconGroup[]) {
  return axios.post<Record<string, never>>('/base/vip-default-icon/save', {
    list,
  });
}
