export type LayerKind = 'auto' | 'fixed';

export interface LayerRecord {
  id: number;
  kind: LayerKind;
  name: string;
  isDefault: boolean;
  /** 系统内置的固定层级，不允许删除 */
  isSystem: boolean;
  description: string;
  deposit: number | null;
  wash: number | null;
  /** 单笔提款上限，null 表示不限制 */
  withdrawLimit: number | null;
  memberCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface LayerFormValue {
  name: string;
  description?: string;
  deposit?: number;
  wash?: number;
  withdrawLimit?: number | null;
}
