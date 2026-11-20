import { exportMembers, type MemberExportRes } from '@/api/member';

/** 可选的导出列：key 为后端字段名，label 为多语言文案 key */
export const MEMBER_EXPORT_COLUMNS = [
  { key: 'member_code', label: 'memberManage.export.col.memberCode' },
  { key: 'username', label: 'memberManage.export.col.username' },
  { key: 'nickname', label: 'memberManage.export.col.nickname' },
  { key: 'status', label: 'memberManage.export.col.status' },
  { key: 'agent_account', label: 'memberManage.export.col.agentAccount' },
  { key: 'created_at', label: 'memberManage.export.col.createdAt' },
];

/** 其余导出参数固定由前端传入，不随页面筛选变化 */
const MEMBER_EXPORT_FIXED = {
  status: 1,
  agent_type: 1,
  include_demo: true,
};

/**
 * 按所选列调用服务端导出并触发下载。返回的 url 为预签名地址，有效期很短，需立即使用。
 * 接口失败时由 axios 拦截器统一提示，这里直接抛出。
 */
export default async function exportMembersFile(
  columns: string[]
): Promise<MemberExportRes> {
  const { data } = await exportMembers({ ...MEMBER_EXPORT_FIXED, columns });
  const link = document.createElement('a');
  link.href = data.url;
  // 跨域下载时 download 属性不生效，靠 Content-Type 触发下载；新开标签页避免覆盖当前页面
  link.target = '_blank';
  link.rel = 'noopener';
  link.download = data.key.split('/').pop() || 'members.csv';
  document.body.appendChild(link);
  link.click();
  link.remove();
  return data;
}
