import { ref } from 'vue';
import {
  AGENT_FILTER,
  AGENT_STATUS,
  getPhoneCallingCodes,
  listAgents,
  type AgentListItem,
  type AgentListParams,
  type PhoneCallingCode,
} from '@/api/agent';
import {
  queryAutoLevelList,
  queryFixedLevelList,
  queryVipLevelList,
  type AutoLevelRecord,
  type FixedLevelRecord,
  type VipLevelApiRecord,
} from '@/api/member';

const AGENT_SEARCH_SIZE = 50;
const LEVEL_PAGE_SIZE = 100;

/**
 * 代理远程搜索：纯数字按 ID，否则按账号模糊。
 * onlyNormal 用于创建会员，服务端要求所属代理状态正常
 */
export function useAgentSearch(onlyNormal = false) {
  const agents = ref<AgentListItem[]>([]);
  const loading = ref(false);
  let ticket = 0;

  async function search(keyword = '') {
    ticket += 1;
    const current = ticket;
    const kw = keyword.trim();
    const params: AgentListParams = {
      page: 1,
      page_size: AGENT_SEARCH_SIZE,
      status: onlyNormal ? AGENT_STATUS.NORMAL : undefined,
    };
    // 服务端 agent_filter_type 为 0 时会忽略 agent_keyword
    if (kw) {
      params.agent_filter_type = AGENT_FILTER.SELF;
      params.agent_keyword = kw;
    }
    loading.value = true;
    try {
      const { data } = await listAgents(params, true);
      if (current === ticket) agents.value = data?.list || [];
    } catch {
      if (current === ticket) agents.value = [];
    } finally {
      if (current === ticket) loading.value = false;
    }
  }

  return { agents, loading, search };
}

/** 层级筛选项，只取第一页（上限 100 个） */
export function useLevelOptions() {
  const autoLevels = ref<AutoLevelRecord[]>([]);
  const fixedLevels = ref<FixedLevelRecord[]>([]);

  async function load() {
    const params = { page: 1, page_size: LEVEL_PAGE_SIZE };
    const [auto, fixed] = await Promise.allSettled([
      queryAutoLevelList(params),
      queryFixedLevelList(params),
    ]);
    if (auto.status === 'fulfilled') {
      autoLevels.value = auto.value.data?.list || [];
    }
    if (fixed.status === 'fulfilled') {
      fixedLevels.value = fixed.value.data?.list || [];
    }
  }

  return { autoLevels, fixedLevels, load };
}

/** VIP 等级选项（按 level 升序，上限 100 个） */
export function useVipLevelOptions() {
  const vipLevels = ref<VipLevelApiRecord[]>([]);

  async function load() {
    try {
      const { data } = await queryVipLevelList({
        page: 1,
        page_size: LEVEL_PAGE_SIZE,
      });
      vipLevels.value = data?.list || [];
    } catch {
      vipLevels.value = [];
    }
  }

  return { vipLevels, load };
}

/** 国际手机区号 */
export function useCallingCodes() {
  const callingCodes = ref<PhoneCallingCode[]>([]);
  const loading = ref(false);

  async function load() {
    if (callingCodes.value.length) return;
    loading.value = true;
    try {
      const { data } = await getPhoneCallingCodes();
      callingCodes.value = data?.list || [];
    } catch {
      callingCodes.value = [];
    } finally {
      loading.value = false;
    }
  }

  return { callingCodes, loading, load };
}
