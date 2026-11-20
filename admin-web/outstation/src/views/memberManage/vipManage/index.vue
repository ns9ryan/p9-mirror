<template>
  <div class="container list-page">
    <a-card class="general-card search-card" :bordered="false">
      <a-tabs
        v-model:active-key="activeTab"
        class="vip-tabs"
        hide-content
        :header-padding="false"
      >
        <a-tab-pane key="level" :title="t('vipManage.tab.level')" />
        <a-tab-pane key="setting" :title="t('vipManage.tab.setting')" />
        <a-tab-pane key="icon" :title="t('vipManage.tab.icon')" />
        <a-tab-pane key="record" :title="t('vipManage.tab.record')" />
        <a-tab-pane key="change" :title="t('vipManage.tab.change')" />
      </a-tabs>

      <a-alert
        v-if="activeTab === 'level' && tipVisible"
        class="vip-tip"
        closable
        @close="tipVisible = false"
      >
        <div>{{ t('vipManage.tip.line1') }}</div>
        <div>{{ t('vipManage.tip.line2') }}</div>
      </a-alert>

      <a-alert v-if="activeTab === 'icon'" class="vip-tip">
        {{ t('vipManage.icon.tip') }}
      </a-alert>

      <VipRewardSearch
        v-if="activeTab === 'record'"
        :query="recordQuery"
        :levels="levelNumbers"
        @search="onRecordSearch"
      />

      <VipChangeSearch
        v-if="activeTab === 'change'"
        :query="changeQuery"
        :levels="levelNumbers"
        @search="onChangeSearch"
      />
    </a-card>

    <a-card class="general-card table-card" :bordered="false">
      <template v-if="activeTab === 'level'">
        <div class="list-header">
          <div class="list-title">{{ t('vipManage.list.title') }}</div>
          <a-space :size="8">
            <a-button :loading="loading" @click="refresh">
              <template #icon><icon-refresh /></template>
              {{ t('front.common.toolbar.refresh') }}
            </a-button>
            <a-button type="primary" @click="openCreate">
              <template #icon><icon-plus /></template>
              {{ t('vipManage.toolbar.create') }}
            </a-button>
          </a-space>
        </div>
        <a-table
          row-key="id"
          :loading="loading"
          :columns="levelColumns"
          :data="levels"
          :pagination="false"
          :bordered="false"
          :scroll="{ x: 1080 }"
        >
          <template #level="{ record }">
            <div class="level-cell">
              <span class="level-name">VIP{{ record.level }}</span>
              <a-tag
                v-if="record.level === 0"
                size="small"
                color="arcoblue"
                bordered
              >
                {{ t('vipManage.default') }}
              </a-tag>
            </div>
          </template>
          <template #promote="{ record }">
            <div class="condition-cell">
              <div>
                {{ t('vipManage.form.deposit') }}:
                {{ formatNumber(record.promoteDeposit) }}
              </div>
              <div>
                {{ t('vipManage.form.wash') }}:
                {{ formatNumber(record.promoteWash) }}
              </div>
            </div>
          </template>
          <template #keep="{ record }">
            <div class="condition-cell">
              <div>
                {{ t('vipManage.form.deposit') }}:
                {{ formatNumber(record.keepDeposit) }}
              </div>
              <div>
                {{ t('vipManage.form.wash') }}:
                {{ formatNumber(record.keepWash) }}
              </div>
            </div>
          </template>
          <template #reward="{ record }">
            {{ formatNumber(record.reward) }}
          </template>
          <template #count="{ record }">
            <a-link
              v-if="record.memberCount > 0"
              class="count-link"
              @click="openMembers(record)"
            >
              {{ formatNumber(record.memberCount) }}
            </a-link>
            <span v-else class="count-zero">0</span>
          </template>
          <template #operations="{ record }">
            <a-space :size="20" class="operation-links">
              <a-link
                :loading="detailLoadingId === record.id"
                @click="openEdit(record)"
              >
                {{ t('front.common.action.edit') }}
              </a-link>
              <span v-if="record.level === 0" class="count-zero">-</span>
              <a-tooltip
                v-else
                :content="deleteBlockReason(record)"
                :disabled="!deleteBlockReason(record)"
              >
                <span>
                  <a-link
                    status="danger"
                    :disabled="Boolean(deleteBlockReason(record))"
                    @click="confirmDelete(record)"
                  >
                    {{ t('front.common.action.delete') }}
                  </a-link>
                </span>
              </a-tooltip>
            </a-space>
          </template>
          <template #empty>
            <a-empty :description="t('vipManage.empty')" />
          </template>
        </a-table>
        <div class="table-pagination">
          <a-pagination
            v-model:current="levelPage.current"
            v-model:page-size="levelPage.size"
            :total="levelTotal"
            :page-size-options="pageSizeOptions"
            :page-size-props="{ popupContainer: 'body' }"
            show-total
            show-page-size
          />
        </div>
      </template>

      <a-spin
        v-else-if="activeTab === 'setting'"
        class="setting-spin"
        :loading="settingLoading"
      >
        <VipSettingPanel
          :settings="settings"
          :saving="settingSaving"
          @save="onSettingSave"
        />
      </a-spin>

      <VipIconPanel
        v-else-if="activeTab === 'icon'"
        :groups="iconGroups"
        :loading="iconLoading"
        :drafts="iconDrafts"
        :busy="iconBusy"
        @upload="onIconUpload"
        @save="onIconSave"
        @discard="(group) => clearIconDraft(group.no)"
        @restore="onIconRestore"
        @refresh="fetchVipIcons"
      />

      <VipRewardTable
        v-else-if="activeTab === 'record'"
        ref="rewardTableRef"
        :records="displayRecords"
        :loading="loading"
        @dispatch="onDispatch"
        @view-changes="onViewChanges"
      />

      <VipChangeTable
        v-else
        ref="changeTableRef"
        :records="displayChanges"
      />
    </a-card>

    <VipFormModal
      v-model:visible="formVisible"
      :mode="formMode"
      :record="editing"
      :levels="takenLevels"
      :next-level="levelTotal"
      @submit="onSubmit"
    />

    <a-drawer
      :visible="drawerRecord !== null"
      :width="640"
      :footer="false"
      unmount-on-close
      :title="drawerTitle"
      @cancel="closeMembers"
    >
      <template v-if="drawerRecord">
        <div class="drawer-summary">
          {{
            t('vipManage.drawer.count', {
              count: formatNumber(drawerRecord.memberCount),
            })
          }}
        </div>
        <div v-if="drawerMembers.length" class="drawer-note">
          {{ t('vipManage.drawer.sample') }}
        </div>
        <a-table
          v-if="drawerMembers.length"
          row-key="id"
          :columns="memberColumns"
          :data="drawerMembers"
          :pagination="false"
          :bordered="true"
        >
          <template #account="{ record }">
            {{ memberAccount(record) }}
          </template>
        </a-table>
        <a-empty v-else :description="t('vipManage.drawer.empty')" />
      </template>
    </a-drawer>
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import axios from 'axios';
  import { formatTime, siteNow } from '@/utils/site-time';
  import { useI18n } from 'vue-i18n';
  import { Message, Modal } from '@arco-design/web-vue';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import useLoading from '@/hooks/loading';
  import {
    createVipLevel,
    deleteVipLevel,
    getVipConfig,
    getVipLevelDetail,
    queryVipIconList,
    queryVipLevelList,
    resetVipIcon,
    updateVipConfig,
    updateVipIcon,
    updateVipLevel,
    uploadVipIcon,
    VIP_WATER_MULTIPLIER_SCALE,
    type CreateVipLevelParams,
    type UpdateVipConfigParams,
    type UpdateVipLevelParams,
    type VipConfig,
    type VipIconGroupRecord,
    type VipLevelApiRecord,
  } from '@/api/member';
  import createMemberSeed from '../memberList/mock';
  import type { MemberRecord } from '../memberList/types';
  import VipFormModal from './components/vip-form-modal.vue';
  import VipIconPanel from './components/vip-icon-panel.vue';
  import VipSettingPanel from './components/vip-setting-panel.vue';
  import VipRewardSearch from './components/vip-reward-search.vue';
  import VipRewardTable from './components/vip-reward-table.vue';
  import VipChangeSearch from './components/vip-change-search.vue';
  import VipChangeTable from './components/vip-change-table.vue';
  import {
    createChangeQuery,
    createLevelChangeSeed,
    createRewardQuery,
    createVipLevelSeed,
    createVipRewardSeed,
    filterLevelChanges,
    filterVipRewards,
  } from './mock';
  import type {
    VipIconAction,
    VipIconBusy,
    VipIconDraft,
    VipIconGroup,
    VipLevelChangeQuery,
    VipLevelFormValue,
    VipLevelRecord,
    VipRewardQuery,
    VipRewardRecord,
    VipSettings,
    VipTab,
  } from './types';

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(false);
  const activeTab = ref<VipTab>('level');
  const tipVisible = ref(true);
  const levels = ref<VipLevelRecord[]>([]);
  const levelTotal = ref(0);
  let levelLoadTicket = 0;
  const detailLoadingId = ref<number | null>(null);
  const memberSeed = createMemberSeed();
  const settings = ref<VipSettings>({
    levelEnabled: true,
    rewardEnabled: true,
    keepEnabled: true,
    dispatchMode: 'manual',
    washMultiple: 3,
  });
  const { loading: settingLoading, setLoading: setSettingLoading } =
    useLoading(false);
  const { loading: settingSaving, setLoading: setSettingSaving } =
    useLoading(false);
  const records = ref(
    createVipRewardSeed(createVipLevelSeed(), settings.value.washMultiple)
  );

  const pageSizeOptions = [15, 30, 50, 100];
  const levelPage = reactive({ current: 1, size: 15 });

  function formatLevelTime(ms: number) {
    return formatTime(ms);
  }

  function toVipLevelRecord(item: VipLevelApiRecord): VipLevelRecord {
    return {
      id: item.id,
      level: item.level,
      reward: item.upgrade_reward,
      promoteDeposit: item.upgrade_deposit,
      promoteWash: item.upgrade_water,
      keepDeposit: item.retention_deposit,
      keepWash: item.retention_water,
      memberCount: item.member_count,
      updatedAt: formatLevelTime(item.updated_at),
    };
  }

  async function fetchVipLevels(options?: { silent?: boolean }) {
    levelLoadTicket += 1;
    const ticket = levelLoadTicket;
    if (!options?.silent) setLoading(true);
    try {
      const { data } = await queryVipLevelList({
        page: levelPage.current,
        page_size: levelPage.size,
      });
      if (ticket !== levelLoadTicket) return false;
      levels.value = (data?.list || []).map(toVipLevelRecord);
      levelTotal.value = data?.total || 0;
      return true;
    } catch {
      if (ticket !== levelLoadTicket) return false;
      levels.value = [];
      levelTotal.value = 0;
      return false;
    } finally {
      if (ticket === levelLoadTicket && !options?.silent) setLoading(false);
    }
  }

  function toVipSettings(config: VipConfig): VipSettings {
    return {
      levelEnabled: config.vip_enabled,
      rewardEnabled: config.upgrade_reward_enabled,
      keepEnabled: config.retention_check_enabled,
      dispatchMode: config.reward_grant_mode === 2 ? 'auto' : 'manual',
      washMultiple:
        config.reward_water_multiplier / VIP_WATER_MULTIPLIER_SCALE,
    };
  }

  function toUpdateVipConfigParams(value: VipSettings): UpdateVipConfigParams {
    return {
      vip_enabled: value.levelEnabled,
      upgrade_reward_enabled: value.rewardEnabled,
      retention_check_enabled: value.keepEnabled,
      reward_grant_mode: value.dispatchMode === 'auto' ? 2 : 1,
      reward_water_multiplier: Math.round(
        value.washMultiple * VIP_WATER_MULTIPLIER_SCALE
      ),
    };
  }

  async function fetchVipConfig(options?: { withLoading?: boolean }) {
    const withLoading = options?.withLoading !== false;
    if (withLoading) setSettingLoading(true);
    try {
      const { data } = await getVipConfig();
      settings.value = toVipSettings(data);
    } catch {
      // 错误提示已由全局拦截器处理
    } finally {
      if (withLoading) setSettingLoading(false);
    }
  }

  const iconGroups = ref<VipIconGroup[]>([]);
  const { loading: iconLoading, setLoading: setIconLoading } =
    useLoading(false);
  const iconBusy = ref<VipIconBusy | null>(null);
  const iconDrafts = ref<Record<number, VipIconDraft>>({});

  function toVipIconGroup(item: VipIconGroupRecord): VipIconGroup {
    return {
      id: item.id,
      no: item.group_no,
      startLevel: item.start_level,
      endLevel: item.end_level,
      defaultIcon: item.default_icon_url,
      icon: item.icon_url,
      updatedAt: formatLevelTime(item.updated_at),
    };
  }

  /** 查询失败时保留现有列表，不当作已清空 */
  async function fetchVipIcons() {
    setIconLoading(true);
    try {
      const { data } = await queryVipIconList();
      iconGroups.value = (data?.list || []).map(toVipIconGroup);
      return true;
    } catch {
      return false;
    } finally {
      setIconLoading(false);
    }
  }

  watch(
    activeTab,
    (tab) => {
      if (tab === 'setting') fetchVipConfig();
      if (tab === 'level') fetchVipLevels();
      if (tab === 'icon') fetchVipIcons();
    },
    { immediate: true }
  );

  watch(
    () => [levelPage.current, levelPage.size] as const,
    () => {
      if (activeTab.value === 'level') fetchVipLevels();
    }
  );

  const recordQuery = ref<VipRewardQuery>(createRewardQuery());
  const rewardTableRef = ref<InstanceType<typeof VipRewardTable>>();

  const changes = ref(createLevelChangeSeed(records.value));
  const changeQuery = ref<VipLevelChangeQuery>(createChangeQuery());
  const changeTableRef = ref<InstanceType<typeof VipChangeTable>>();
  const displayChanges = computed(() =>
    filterLevelChanges(changes.value, changeQuery.value)
  );

  const formVisible = ref(false);
  const formMode = ref<'create' | 'edit'>('create');
  const editing = ref<VipLevelRecord | null>(null);
  const drawerLevel = ref<number | null>(null);

  const takenLevels = computed(() =>
    levels.value
      .filter((item) => item.level !== editing.value?.level)
      .map((item) => item.level)
  );

  const levelColumns = computed<TableColumnData[]>(() => [
    {
      title: t('vipManage.column.level'),
      slotName: 'level',
      width: 120,
    },
    {
      title: t('vipManage.column.promote'),
      slotName: 'promote',
      width: 200,
    },
    {
      title: t('vipManage.column.keep'),
      slotName: 'keep',
      width: 200,
    },
    {
      title: t('vipManage.column.reward'),
      slotName: 'reward',
      width: 120,
    },
    {
      title: t('vipManage.column.count'),
      slotName: 'count',
      width: 120,
    },
    {
      title: t('vipManage.column.updatedAt'),
      dataIndex: 'updatedAt',
      width: 180,
    },
    {
      title: t('vipManage.column.operations'),
      slotName: 'operations',
      width: 120,
    },
  ]);

  const levelNumbers = computed(() => levels.value.map((item) => item.level));

  const displayRecords = computed(() =>
    filterVipRewards(records.value, recordQuery.value)
  );

  const memberColumns = computed<TableColumnData[]>(() => [
    {
      title: t('memberManage.card.memberId'),
      dataIndex: 'id',
      width: 140,
    },
    {
      title: t('memberManage.card.account'),
      slotName: 'account',
    },
    {
      title: t('memberManage.card.vip'),
      dataIndex: 'vipLevel',
      width: 120,
    },
  ]);

  const drawerRecord = computed(
    () => levels.value.find((item) => item.level === drawerLevel.value) ?? null
  );

  const drawerTitle = computed(() => {
    if (!drawerRecord.value) return t('vipManage.drawer.title');
    return `${t('vipManage.drawer.title')} · VIP${drawerRecord.value.level}`;
  });

  const drawerMembers = computed(() => {
    const level = drawerRecord.value?.level;
    if (level == null) return [];
    return memberSeed.filter((item) => item.vipLevel === level);
  });

  function formatNumber(value: number) {
    return value.toLocaleString('en-US');
  }

  function now() {
    return siteNow().format('YYYY-MM-DD HH:mm:ss');
  }

  function onRecordSearch(query: VipRewardQuery) {
    recordQuery.value = query;
    rewardTableRef.value?.resetPage();
  }

  function onChangeSearch(query: VipLevelChangeQuery) {
    changeQuery.value = query;
    changeTableRef.value?.resetPage();
  }

  function onViewChanges(account: string) {
    changeQuery.value = { ...createChangeQuery(), account };
    activeTab.value = 'change';
  }

  function onDispatch(record: VipRewardRecord) {
    records.value = records.value.map((item): VipRewardRecord =>
      item.orderNo === record.orderNo
        ? { ...item, status: 'dispatched', claimedAt: now() }
        : item
    );
    Message.success(t('vipManage.record.dispatched'));
  }

  function memberAccount(record: MemberRecord) {
    return record.account || record.email || record.phone || '--';
  }

  function openCreate() {
    formMode.value = 'create';
    editing.value = null;
    formVisible.value = true;
  }

  async function openEdit(record: VipLevelRecord) {
    if (detailLoadingId.value != null) return;
    detailLoadingId.value = record.id;
    try {
      const { data } = await getVipLevelDetail(record.id);
      const detail = toVipLevelRecord(data);
      levels.value = levels.value.map((item) =>
        item.id === detail.id ? detail : item
      );
      formMode.value = 'edit';
      editing.value = detail;
      formVisible.value = true;
    } catch {
      // 错误提示已由全局拦截器处理
    } finally {
      detailLoadingId.value = null;
    }
  }

  async function onSubmit(value: VipLevelFormValue) {
    if (formMode.value === 'create') {
      try {
        await createVipLevel({
          upgrade_deposit: value.promoteDeposit,
          upgrade_water: value.promoteWash,
          retention_deposit: value.keepDeposit,
          retention_water: value.keepWash,
          upgrade_reward: value.reward,
        });
      } catch {
        return false;
      }
      Message.success(t('vipManage.message.created'));
      // 新等级编号最大，排在最后一页
      const lastPage = Math.max(
        1,
        Math.ceil((levelTotal.value + 1) / levelPage.size)
      );
      if (levelPage.current !== lastPage) levelPage.current = lastPage;
      else await fetchVipLevels();
      return true;
    }
    const current = editing.value;
    if (!current) return false;
    const fields: [keyof CreateVipLevelParams, number, number][] = [
      ['upgrade_deposit', value.promoteDeposit, current.promoteDeposit],
      ['upgrade_water', value.promoteWash, current.promoteWash],
      ['retention_deposit', value.keepDeposit, current.keepDeposit],
      ['retention_water', value.keepWash, current.keepWash],
      ['upgrade_reward', value.reward, current.reward],
    ];
    const changed = fields.filter(([, next, prev]) => next !== prev);
    if (!changed.length) {
      Message.info(t('vipManage.message.noChange'));
      return true;
    }
    const params: UpdateVipLevelParams = { id: current.id };
    changed.forEach(([key, next]) => {
      params[key] = next;
    });
    try {
      await updateVipLevel(params);
    } catch {
      return false;
    }
    Message.success(t('vipManage.message.updated'));
    await fetchVipLevels();
    return true;
  }

  /** 等级编号从 0 连续递增，最高等级 = 总数 - 1 */
  function deleteBlockReason(record: VipLevelRecord) {
    if (record.level !== levelTotal.value - 1) {
      return t('vipManage.delete.notHighest');
    }
    if (record.memberCount > 0) return t('vipManage.delete.hasMembers');
    return '';
  }

  function confirmDelete(record: VipLevelRecord) {
    if (deleteBlockReason(record)) return;
    Modal.confirm({
      title: t('vipManage.confirm.deleteTitle'),
      content: t('vipManage.confirm.deleteContent', { level: record.level }),
      okText: t('front.common.delete.okText'),
      cancelText: t('memberManage.confirm.cancel'),
      okButtonProps: { status: 'danger' },
      async onBeforeOk() {
        try {
          await deleteVipLevel(record.id);
        } catch {
          return false;
        }
        if (drawerLevel.value === record.level) drawerLevel.value = null;
        Message.success(t('vipManage.message.deleted'));
        const lastPage = Math.max(
          1,
          Math.ceil((levelTotal.value - 1) / levelPage.size)
        );
        if (levelPage.current > lastPage) levelPage.current = lastPage;
        else fetchVipLevels();
        return true;
      },
    });
  }

  function openMembers(record: VipLevelRecord) {
    drawerLevel.value = record.level;
  }

  function closeMembers() {
    drawerLevel.value = null;
  }

  async function onSettingSave(value: VipSettings) {
    setSettingSaving(true);
    try {
      await updateVipConfig(toUpdateVipConfigParams(value));
      // 接口不回写配置，按文档再拉一次同步；按钮已有 loading，不再整页转圈
      await fetchVipConfig({ withLoading: false });
      Message.success(t('vipManage.message.settingSaved'));
    } catch {
      // 错误提示已由全局拦截器处理
    } finally {
      setSettingSaving(false);
    }
  }

  function clearIconDraft(no: number) {
    if (!(no in iconDrafts.value)) return;
    const next = { ...iconDrafts.value };
    delete next[no];
    iconDrafts.value = next;
  }

  /** 图标操作全局串行，避免旧的待保存图片覆盖刚完成的恢复 */
  async function runIconAction(
    no: number,
    action: VipIconAction,
    task: () => Promise<void>
  ) {
    if (iconBusy.value) return;
    iconBusy.value = { no, action };
    try {
      await task();
    } finally {
      iconBusy.value = null;
    }
  }

  /** 无响应或网关超时：写请求可能已生效，需要重新查询确认 */
  function isUncertainResult(error: unknown) {
    return (
      axios.isAxiosError(error) &&
      (!error.response || error.response.status === 504)
    );
  }

  function onIconUpload(group: VipIconGroup, file: File) {
    return runIconAction(group.no, 'upload', async () => {
      try {
        const { data } = await uploadVipIcon(group.no, file);
        iconDrafts.value = {
          ...iconDrafts.value,
          [group.no]: { url: data.url, key: data.key },
        };
        Message.success(t('vipManage.message.iconUploaded'));
      } catch {
        // 错误提示已由全局拦截器处理，保留当前已保存配置
      }
    });
  }

  function onIconSave(group: VipIconGroup) {
    const draft = iconDrafts.value[group.no];
    if (!draft) return Promise.resolve();
    return runIconAction(group.no, 'save', async () => {
      try {
        await updateVipIcon({ group_no: group.no, icon_url: draft.key });
      } catch (error) {
        if (isUncertainResult(error) && (await fetchVipIcons())) {
          const saved = iconGroups.value.find((item) => item.no === group.no);
          if (saved?.icon === draft.url) clearIconDraft(group.no);
        }
        return;
      }
      clearIconDraft(group.no);
      if (await fetchVipIcons()) {
        Message.success(t('vipManage.message.iconSaved'));
      } else {
        Message.warning(t('vipManage.message.iconSavedRefreshFailed'));
      }
    });
  }

  function onIconRestore(group: VipIconGroup) {
    return runIconAction(group.no, 'restore', async () => {
      try {
        await resetVipIcon(group.no);
      } catch (error) {
        if (isUncertainResult(error)) await fetchVipIcons();
        return;
      }
      clearIconDraft(group.no);
      if (await fetchVipIcons()) {
        Message.success(t('vipManage.message.iconRestored'));
      } else {
        Message.warning(t('vipManage.message.iconRestoredRefreshFailed'));
      }
    });
  }

  async function refresh() {
    const ok = await fetchVipLevels();
    if (ok) Message.success(t('vipManage.message.refreshed'));
  }
</script>

<style scoped lang="less">
  .general-card {
    margin-top: 16px;
  }

  .search-card :deep(.arco-card-body) {
    padding: 12px 20px 20px;
  }

  .table-card :deep(.arco-card-body) {
    padding: 20px;
  }

  .setting-spin {
    display: block;
  }

  .vip-tabs {
    :deep(.arco-tabs-nav-type-line .arco-tabs-tab) {
      margin: 0 28px 0 0;
      padding: 0 0 8px;
      line-height: 22px;
      white-space: nowrap;
    }

    :deep(.arco-tabs-nav::before) {
      right: 0;
      left: 0;
    }
  }

  .vip-tip {
    margin-top: 16px;
    border-radius: 6px;

    :deep(.arco-alert-content) {
      color: var(--color-text-2);
      font-size: 13px;
      line-height: 22px;
    }
  }

  .list-header {
    display: flex;
    gap: 12px;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;

    .list-title {
      margin-bottom: 0;
    }
  }

  .list-title {
    margin-bottom: 16px;
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 16px;
  }

  :deep(.arco-table-td .arco-table-cell) {
    font-variant-numeric: tabular-nums;
  }

  .level-cell {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .level-name {
    color: var(--color-text-1);
    font-weight: 600;
  }

  .condition-cell {
    color: var(--color-text-2);
    font-size: 13px;
    line-height: 20px;
  }

  .count-link {
    padding: 0;
  }

  .count-zero {
    color: var(--color-text-3);
  }

  .operation-links :deep(.arco-link) {
    padding: 0;
  }

  .table-pagination {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }

  :deep(.arco-pagination-options .arco-select) {
    min-width: 100px;
  }

  .drawer-summary {
    margin-bottom: 8px;
    color: var(--color-text-1);
    font-weight: 500;
  }

  .drawer-note {
    margin-bottom: 12px;
    color: var(--color-text-3);
    font-size: 12px;
  }
</style>
