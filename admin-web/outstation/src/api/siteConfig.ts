import axios from 'axios';
import {
  uploadResource,
  type ResourceUploadResult,
} from './resource';

/** 本站勾选 / 最终生效状态：1 开启（生效） / 2 关闭（未生效） */
export const SITE_OPTION_STATUS = {
  ON: 1,
  OFF: 2,
} as const;

/** 登录 / 注册验证码模式：0 不验证 / 1 拖拽 / 2 图形 */
export const SITE_CAPTCHA_MODE = {
  NONE: 0,
  DRAG: 1,
  IMAGE: 2,
} as const;

export interface SiteBaseConfig {
  operator_code: string;
  /** 本次查询对应的后台域名 */
  domain_name: string;
  /** 本站勾选的语言，不保证都实际可用 */
  language_codes: string[];
  /** 本站勾选的地区，不保证都实际可用 */
  region_codes: string[];
  settlement_currency_code: string;
  timezone_code: string;
}

export interface SiteLanguageOption {
  /** Core 中的语言资源 ID；基础资料缺失的占位对象为 0 */
  id: number;
  lang: string;
  /** Core 语言开关：0 未禁用 / 1 已禁用；不表示本站是否勾选 */
  disabled: number;
  /** 本站是否勾选：1 开启 / 2 关闭 */
  status: number;
  /** 最终能否使用：1 生效 / 2 未生效 */
  effective_status: number;
  created_at: number;
  updated_at: number;
  /** 基础资料缺失时可能为空 */
  name: string;
  sort_no: number;
  i18n_key: string;
}

export interface SiteRegionBase {
  /** 基础资料缺失的占位对象为 0 */
  id: number;
  code: string;
  calling_code: string;
  name_key: string;
  /** 基础数据是否开启：1 开启 / 2 关闭；不是本站选择 */
  status: number;
  sort_no: number;
}

export interface SiteRegionOption {
  region: SiteRegionBase;
  /** 本站是否勾选：1 开启 / 2 关闭 */
  status: number;
  /** 最终能否使用：1 生效 / 2 未生效 */
  effective_status: number;
}

export interface SiteCurrency {
  id: number;
  code: string;
  name_key: string;
  /** 1 法定货币 / 2 虚拟货币 */
  currency_type: number;
  symbol: string;
  /** 金额换算倍率，不是汇率 */
  amount_factor: number;
  /** 1 开启 / 2 关闭；0 为缺少资料的占位值 */
  status: number;
  sort_no: number;
}

export interface SiteTimezone {
  id: number;
  code: string;
  name_key: string;
  /** 1 开启 / 2 关闭；0 为缺少资料的占位值 */
  status: number;
  sort_no: number;
}

export interface SiteSettings {
  operator_code: string;
  site_title: string;
  /** 可能是站内绝对路径或空字符串 */
  site_logo_url: string;
  /** 以下四项阈值：0 不限制，允许范围 0-100 */
  email_login_limit: number;
  sms_login_limit: number;
  member_password_error_limit: number;
  ip_member_register_limit: number;
  /** 0 不验证 / 1 拖拽 / 2 图形 */
  register_captcha_mode: number;
  /** 0 不验证 / 1 拖拽 / 2 图形 */
  login_captcha_mode: number;
  /** 最后修改人用户 ID；0 为初始化或尚未保存 */
  updated_by: number;
  created_at: number;
  updated_at: number;
  /** false 为尚未保存、使用默认值 */
  configured: boolean;
}

export interface SiteInfoRes {
  site_config: SiteBaseConfig;
  /** 分配给本站的语言，包含本站关闭的选项 */
  languages: SiteLanguageOption[];
  /** 分配给本站的地区，包含本站关闭的选项 */
  region_options: SiteRegionOption[];
  effective_region_codes: string[];
  effective_language_codes: string[];
  currency: SiteCurrency;
  timezone: SiteTimezone;
  site_settings: SiteSettings;
}

/**
 * 查询本站配置 GET /operator/info
 * 不传参数，后端通过 Host 与登录身份确定当前站点
 */
export function getSiteInfo() {
  return axios.get<SiteInfoRes>('/site/info');
}

/**
 * 保存本站设置 PUT /operator/settings
 * 必须一次提交以下 8 个字段；站点由 Host 确定，禁止传 domain_name / operator_code 等
 */
export interface UpdateSiteSettingsParams {
  site_title: string;
  site_logo_url: string;
  email_login_limit: number;
  sms_login_limit: number;
  member_password_error_limit: number;
  ip_member_register_limit: number;
  register_captcha_mode: number;
  login_captcha_mode: number;
}

export function updateSiteSettings(data: UpdateSiteSettingsParams) {
  return axios.put<SiteSettings>('/site/settings', data);
}

/**
 * 保存本站开放语言 PUT /operator/languages
 * 数组内为开启，已分配但未放入的关闭；传 [] 表示全部关闭；字段不可省略或传 null
 */
export function updateSiteLanguages(data: { language_codes: string[] }) {
  return axios.put<{ language_codes: string[] }>('/site/languages', data);
}

/**
 * 保存本站经营地区 PUT /operator/regions
 * 数组内为开启，已分配但未放入的关闭；传 [] 表示全部关闭；字段不可省略或传 null
 */
export function updateSiteRegions(data: { region_codes: string[] }) {
  return axios.put<{ region_codes: string[] }>('/site/regions', data);
}

/** Logo 上传结果（统一 /resource/upload 响应）；site_logo_url 写入完整 url */
export type SiteLogoUploadRes = ResourceUploadResult;

/**
 * 上传本站 Logo POST /resource/upload
 * scene=logo，恰好 file + scene；上传成功后仍需 PUT /operator/settings 写入 site_logo_url
 */
export function uploadSiteLogo(file: File) {
  return uploadResource({ file, scene: 'logo' });
}
