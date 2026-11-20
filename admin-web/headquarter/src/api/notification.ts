import axios from 'axios';

/** 1 官方通知 */
export const NOTIFICATION_TYPE = {
  OFFICIAL: 1,
} as const;

/** 1 待发送 / 2 已发送 / 3 发送失败 */
export const NOTIFICATION_STATUS = {
  PENDING: 1,
  SENT: 2,
  FAILED: 3,
} as const;

/** 1 永久 */
export const NOTIFICATION_VALIDITY = {
  PERMANENT: 1,
} as const;

/** 关键字匹配字段 */
export type NotificationKeywordField = 'title' | 'content';

export interface NotificationTargetOperator {
  /** 分站 ID */
  id: number;
  /** 分站业务编码 */
  code: string;
  /** 分站名称 */
  name: string;
}

/** 多语言文本，key 为语言码（如 zh-CN / en-US），value 为该语言下的文本 */
export type NotificationI18nText = Record<string, string>;

export interface NotificationRecord {
  id: number;
  /** 默认语言标题 */
  title: string;
  /** 默认语言内容 */
  content: string;
  /** 多语言标题 */
  title_i18n?: NotificationI18nText;
  /** 多语言内容 */
  content_i18n?: NotificationI18nText;
  /** 1 官方通知 */
  type: number;
  /** 发送对象分站 */
  target_operators: NotificationTargetOperator[];
  /** 1 永久 */
  validity: number;
  /** 1 待发送 / 2 已发送 / 3 发送失败 */
  status: number;
  /** 最后操作人 */
  updated_by: string;
  created_at: number;
  updated_at: number;
}

export interface NotificationListParams {
  /** 页码，从 1 开始 */
  page: number;
  /** 每页数量，范围 1-100 */
  page_size: number;
  /** 关键字匹配字段，keyword 非空时生效 */
  keyword_field?: NotificationKeywordField;
  /** 搜索关键字，最大 100 个字符 */
  keyword?: string;
  /** 1 官方通知 */
  type?: number;
  /** 1 待发送 / 2 已发送 / 3 发送失败 */
  status?: number;
}

export interface NotificationListRes {
  total: number;
  list: NotificationRecord[];
}

/** 通知推送分页列表 GET /admin/notification/list，按创建时间倒序 */
export function queryNotificationList(params: NotificationListParams) {
  return axios.get<NotificationListRes>('/notification/list', { params });
}

export interface CreateNotificationParams {
  /** 1 官方通知 */
  type: number;
  /** 发送对象分站 ID，至少 1 个 */
  operator_ids: number[];
  /** 1 永久 */
  validity: number;
  /** 多语言标题，至少 1 种语言，单条最大 200 个字符 */
  title_i18n: NotificationI18nText;
  /** 多语言内容，至少 1 种语言，单条最大 5000 个字符 */
  content_i18n: NotificationI18nText;
}

export interface CreateNotificationRes {
  id: number;
}

/** 新增通知推送 POST /admin/notification/create */
export function createNotification(data: CreateNotificationParams) {
  return axios.post<CreateNotificationRes>('/notification/create', data);
}
