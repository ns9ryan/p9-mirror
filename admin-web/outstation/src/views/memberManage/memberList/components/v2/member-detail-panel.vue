<template>
  <aside class="detail-panel">
    <div class="panel-head">
      <div class="panel-title">
        <span>
          {{ t('memberManage.v2.panel.title') }}
          <icon-loading v-if="loading" class="title-loading" />
        </span>
        <a-button
          type="text"
          size="mini"
          :aria-label="t('memberManage.v2.panel.close')"
          @click="emit('close')"
        >
          <template #icon><icon-close /></template>
        </a-button>
      </div>

      <div class="profile">
        <a-avatar :size="52" class="profile-avatar">
          <img
            v-if="avatarUrl"
            :src="avatarUrl"
            alt=""
            @error="avatarFailed = true"
          />
          <template v-else>{{ initial }}</template>
        </a-avatar>
        <div class="profile-main">
          <div class="profile-name">
            <strong class="ellipsis" :title="record.account">
              {{ displayText(record.account) }}
            </strong>
            <a-tag size="small" :color="vipColor(record.vipLevel)">
              VIP{{ record.vipLevel }}
            </a-tag>
            <a-tag size="small" :color="STATUS_COLORS[record.status]">
              {{ t(`memberManage.status.${record.status}`) }}
            </a-tag>
          </div>
          <div class="profile-sub ellipsis">
            {{ displayText(record.nickname) }}
          </div>
          <div class="profile-sub num">
            {{ t('memberManage.column.memberId') }}:
            {{ displayText(record.code) }}
          </div>
        </div>
      </div>

      <div class="profile-actions">
        <a-button type="primary" size="small" @click="emit('edit')">
          {{ t('memberManage.v2.action.edit') }}
        </a-button>
        <a-dropdown
          position="br"
          :popup-max-height="false"
          @select="onCommand"
        >
          <a-button size="small">
            {{ t('memberManage.v2.action.more') }}
            <icon-down />
          </a-button>
          <template #content>
            <a-doption
              v-for="item in OPERATION_COMMANDS"
              :key="item.key"
              :value="item.key"
            >
              {{ t(item.label) }}
            </a-doption>
            <a-doption
              v-for="command in statusCommandsOf(record.status)"
              :key="command"
              :value="command"
            >
              {{ t(STATUS_COMMANDS[command].label) }}
            </a-doption>
          </template>
        </a-dropdown>
      </div>

      <a-tabs
        class="panel-tabs"
        size="small"
        :active-key="activeSection"
        @change="scrollToSection"
      >
        <a-tab-pane
          v-for="item in SECTIONS"
          :key="item.key"
          :title="t(item.tab)"
        />
      </a-tabs>
    </div>

    <div ref="bodyRef" class="panel-body" @scroll="onBodyScroll">
      <section data-section="basic" class="panel-card">
        <h3>{{ t('memberManage.v2.section.basic') }}</h3>
        <dl class="info-list">
          <template v-for="item in basicItems" :key="item.label">
            <dt>{{ item.label }}</dt>
            <dd :title="item.value">{{ item.value }}</dd>
          </template>
        </dl>
      </section>

      <section data-section="funds" class="panel-card">
        <h3>{{ t('memberManage.v2.section.funds') }}</h3>
        <div class="stat-grid">
          <div v-for="item in fundItems" :key="item.label" class="stat-item">
            <span class="stat-icon" :class="`is-${item.tone}`">
              <component :is="item.icon" />
            </span>
            <div class="stat-main">
              <span class="stat-label">{{ item.label }}</span>
              <strong class="num">{{ item.value }}</strong>
              <small class="num">{{ item.sub }}</small>
            </div>
          </div>
        </div>
      </section>

      <section data-section="game" class="panel-card">
        <div class="game-grid">
          <div>
            <h3>{{ t('memberManage.v2.section.game') }}</h3>
            <p v-for="item in gameItems" :key="item.label" class="kv">
              <span>{{ item.label }}</span>
              <strong class="num">{{ item.value }}</strong>
            </p>
          </div>
          <div>
            <h3>{{ t('memberManage.v2.section.winLoss') }}</h3>
            <p v-for="item in winLossItems" :key="item.label" class="kv">
              <span>{{ item.label }}</span>
              <strong class="num" :class="winLossClass(item.amount)">
                {{ formatAmount(item.amount) }}
              </strong>
            </p>
          </div>
        </div>
      </section>

      <section data-section="records" class="panel-card">
        <h3>{{ t('memberManage.v2.section.records') }}</h3>
        <a-tabs v-model:active-key="recordType" class="record-tabs" size="small">
          <a-tab-pane
            v-for="item in RECORD_TYPES"
            :key="item"
            :title="t(`memberManage.v2.records.tab.${item}`)"
          />
        </a-tabs>
        <!-- 会员维度的记录接口尚未提供 -->
        <a-table
          :columns="recordColumns"
          :data="[]"
          :pagination="false"
          :bordered="false"
          size="small"
        >
          <template #empty>
            <a-empty :description="t('memberManage.v2.records.empty')" />
          </template>
        </a-table>
      </section>
    </div>
  </aside>
</template>

<script lang="ts" setup>
  import { computed, nextTick, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import {
    STATUS_COLORS,
    displayText,
    fixedLevelsText,
    formatAmount,
    formatAmountCount,
    registerChannelText,
    registerMethodText,
    restrictionText,
    vipColor,
  } from '../../format';
  import type { MemberDetailRecord, MemberRecord } from '../../types';
  import { OPERATION_COMMANDS } from '../operations/commands';
  import {
    STATUS_COMMANDS,
    statusCommandsOf,
    type MemberCommand,
  } from './actions';

  const props = defineProps<{
    record: MemberRecord;
    /** 详情接口数据，加载完成前为 null，此时补充字段显示 -- */
    detail: MemberDetailRecord | null;
    /** 按 detail.avatarId 从头像资源解析，未设置或资源未加载时为空 */
    avatarUrl: string;
    loading: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'close'): void;
    (e: 'edit'): void;
    (e: 'command', command: MemberCommand): void;
  }>();

  const { t } = useI18n();

  type SectionKey = 'basic' | 'funds' | 'game' | 'records';

  const SECTIONS: { key: SectionKey; tab: string }[] = [
    { key: 'basic', tab: 'memberManage.v2.tab.basic' },
    { key: 'funds', tab: 'memberManage.v2.tab.funds' },
    { key: 'game', tab: 'memberManage.v2.tab.game' },
    { key: 'records', tab: 'memberManage.v2.tab.records' },
  ];

  type RecordType = 'deposit' | 'withdraw' | 'bet' | 'operation';

  const RECORD_TYPES: RecordType[] = ['deposit', 'withdraw', 'bet', 'operation'];

  const RECORD_COLUMNS: Record<RecordType, string[]> = {
    deposit: ['time', 'orderNo', 'amount', 'status', 'channel'],
    withdraw: ['time', 'orderNo', 'amount', 'status', 'channel'],
    bet: ['time', 'betNo', 'game', 'betAmount', 'winLoss'],
    operation: ['time', 'operator', 'content'],
  };

  const bodyRef = ref<HTMLElement>();
  const activeSection = ref<SectionKey>('basic');
  const recordType = ref<RecordType>('deposit');

  const recordColumns = computed(() =>
    RECORD_COLUMNS[recordType.value].map((key) => ({
      title: t(`memberManage.v2.records.column.${key}`),
      dataIndex: key,
      ellipsis: true,
      tooltip: true,
    }))
  );

  const initial = computed(() =>
    (props.record.account || props.record.email || 'U').charAt(0).toUpperCase()
  );

  const avatarFailed = ref(false);
  const avatarUrl = computed(() => (avatarFailed.value ? '' : props.avatarUrl));

  watch(
    () => props.avatarUrl,
    () => {
      avatarFailed.value = false;
    }
  );

  const basicItems = computed(() => {
    const { record, detail } = props;
    const agent = record.parentAgent
      ? `${record.parentAgent}${
          record.parentAgentId ? ` (ID: ${record.parentAgentId})` : ''
        }`
      : '';
    return [
      { label: t('memberManage.v2.field.account'), value: record.account },
      { label: t('memberManage.v2.field.nickname'), value: record.nickname },
      {
        label: t('memberManage.detail.gender'),
        value: detail ? t(`memberManage.gender.${detail.gender}`) : '',
      },
      { label: t('memberManage.detail.birthday'), value: detail?.birthday },
      { label: t('memberManage.detail.region'), value: detail?.regionCode },
      { label: t('memberManage.v2.field.phone'), value: record.phone },
      { label: t('memberManage.v2.field.email'), value: record.email },
      { label: t('memberManage.v2.field.agent'), value: agent },
      {
        label: t('memberManage.detail.agentLine'),
        value: detail?.agentLineCode,
      },
      {
        label: t('memberManage.levelKind.auto'),
        value: record.autoLevelName,
      },
      {
        label: t('memberManage.levelKind.fixed'),
        value: fixedLevelsText(record),
      },
      {
        label: t('memberManage.detail.restriction'),
        value: detail ? restrictionText(detail, t) : '',
      },
      {
        label: t('memberManage.v2.field.registerSource'),
        value: registerChannelText(record.registerSource, t),
      },
      {
        label: t('memberManage.v2.field.registerType'),
        value: registerMethodText(record.registerType, t),
      },
      {
        label: t('memberManage.v2.field.registeredAt'),
        value: record.registeredAt,
      },
      { label: t('memberManage.detail.registerIp'), value: detail?.registerIp },
      {
        label: t('memberManage.detail.registerDevice'),
        value: detail?.registerDeviceCode,
      },
      {
        label: t('memberManage.detail.registerDomain'),
        value: detail?.registerDomain,
      },
      {
        label: t('memberManage.v2.field.lastLogin'),
        value: record.lastLoginAt,
      },
      { label: t('memberManage.v2.field.loginIp'), value: record.lastLoginIp },
      {
        label: t('memberManage.detail.inviter'),
        value: detail?.inviterMemberId,
      },
      { label: t('memberManage.v2.field.remark'), value: record.remark },
    ].map((item) => ({ ...item, value: displayText(item.value || '') }));
  });

  const fundItems = computed(() => {
    const { record } = props;
    return [
      {
        label: t('memberManage.v2.funds.balance'),
        value: formatAmount(record.balance),
        sub: `${t('memberManage.v2.funds.unsettled')} ${formatAmount(
          record.unsettled
        )}`,
        icon: 'icon-storage',
        tone: 'blue',
      },
      {
        label: t('memberManage.v2.funds.todayDeposit'),
        value: formatAmount(record.periodDeposit),
        sub: `${t('memberManage.v2.funds.totalDeposit')} ${formatAmountCount(
          record.totalDepositAmount,
          record.totalDepositCount
        )}`,
        icon: 'icon-arrow-rise',
        tone: 'green',
      },
      {
        label: t('memberManage.v2.funds.todayWithdraw'),
        value: formatAmount(record.periodWithdraw),
        sub: `${t('memberManage.v2.funds.totalWithdraw')} ${formatAmountCount(
          record.totalWithdrawAmount,
          record.totalWithdrawCount
        )}`,
        icon: 'icon-arrow-fall',
        tone: 'red',
      },
      {
        label: t('memberManage.v2.funds.firstDeposit'),
        value: formatAmount(record.firstDepositAmount),
        sub: `${t('memberManage.v2.funds.diff')} ${formatAmount(
          record.depositWithdrawDiff
        )}`,
        icon: 'icon-gift',
        tone: 'orange',
      },
    ];
  });

  const gameItems = computed(() => [
    {
      label: t('memberManage.v2.game.todayBet'),
      value: formatAmount(props.record.periodBet),
    },
    {
      label: t('memberManage.v2.game.todayValidBet'),
      value: formatAmount(props.record.periodValidBet),
    },
    {
      label: t('memberManage.v2.game.todayBonus'),
      value: formatAmount(props.record.periodBonus),
    },
  ]);

  const winLossItems = computed(() => [
    {
      label: t('memberManage.v2.game.todayWinLoss'),
      amount: props.record.periodWinLoss,
    },
    {
      label: t('memberManage.v2.game.totalWinLoss'),
      amount: props.record.totalWinLoss,
    },
  ]);

  function winLossClass(value: number) {
    if (value > 0) return 'is-win';
    if (value < 0) return 'is-loss';
    return '';
  }

  function sectionEl(key: SectionKey) {
    return bodyRef.value?.querySelector<HTMLElement>(`[data-section="${key}"]`);
  }

  function scrollToSection(key: string | number) {
    const el = sectionEl(key as SectionKey);
    if (!bodyRef.value || !el) return;
    activeSection.value = key as SectionKey;
    bodyRef.value.scrollTo({ top: el.offsetTop - 12, behavior: 'smooth' });
  }

  function onBodyScroll() {
    const body = bodyRef.value;
    if (!body) return;
    const atBottom =
      body.scrollTop + body.clientHeight >= body.scrollHeight - 4;
    if (atBottom) {
      activeSection.value = SECTIONS[SECTIONS.length - 1].key;
      return;
    }
    const current = SECTIONS.filter((item) => {
      const el = sectionEl(item.key);
      return el && el.offsetTop - 24 <= body.scrollTop;
    }).pop();
    activeSection.value = current?.key || 'basic';
  }

  function onCommand(
    value: string | number | Record<string, unknown> | undefined
  ) {
    emit('command', value as MemberCommand);
  }

  watch(
    () => props.record.id,
    async () => {
      activeSection.value = 'basic';
      recordType.value = 'deposit';
      await nextTick();
      bodyRef.value?.scrollTo({ top: 0 });
    }
  );
</script>

<style scoped lang="less">
  .detail-panel {
    display: flex;
    flex-direction: column;
    min-height: 0;
    overflow: hidden;
    background: var(--color-bg-2);
    border-radius: 4px;
  }

  .panel-head {
    flex: none;
    padding: 16px 20px 0;
    border-bottom: 1px solid var(--color-neutral-3);
  }

  .panel-title {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 16px;
  }

  .profile {
    display: flex;
    gap: 14px;
    align-items: center;
  }

  .title-loading {
    margin-left: 6px;
    color: var(--color-text-3);
    font-size: 14px;
  }

  .profile-avatar {
    flex: none;
    color: rgb(var(--primary-6));
    font-weight: 600;
    font-size: 22px;
    background: rgb(var(--primary-1));

    img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }
  }

  .profile-main {
    flex: 1;
    min-width: 0;
  }

  .profile-name {
    display: flex;
    gap: 8px;
    align-items: center;
    min-width: 0;

    strong {
      min-width: 0;
      color: var(--color-text-1);
      font-size: 17px;
    }
  }

  .profile-sub {
    margin-top: 2px;
    color: var(--color-text-3);
    font-size: 12px;
  }

  .profile-actions {
    display: flex;
    gap: 8px;
    margin-top: 14px;
  }

  .panel-tabs {
    margin-top: 8px;

    :deep(.arco-tabs-content) {
      display: none;
    }

    :deep(.arco-tabs-nav::before) {
      display: none;
    }
  }

  .panel-body {
    position: relative;
    flex: 1;
    min-height: 0;
    padding: 12px;
    overflow-y: auto;
    background: var(--color-fill-1);
  }

  .panel-card {
    padding: 16px;
    background: var(--color-bg-2);
    border-radius: 6px;

    & + & {
      margin-top: 12px;
    }

    h3 {
      margin: 0 0 12px;
      color: var(--color-text-1);
      font-weight: 600;
      font-size: 14px;
    }
  }

  .info-list {
    display: grid;
    grid-template-columns: 76px minmax(0, 1fr);
    row-gap: 10px;
    column-gap: 12px;
    margin: 0;
    font-size: 13px;

    dt {
      color: var(--color-text-3);
    }

    dd {
      margin: 0;
      overflow: hidden;
      color: var(--color-text-1);
      white-space: nowrap;
      text-overflow: ellipsis;
    }
  }

  .stat-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;
  }

  .stat-item {
    display: flex;
    gap: 10px;
    min-width: 0;
    padding: 12px;
    background: var(--color-fill-1);
    border-radius: 6px;
  }

  .stat-icon {
    display: inline-flex;
    flex: none;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    font-size: 16px;
    border-radius: 8px;

    &.is-blue {
      color: rgb(var(--arcoblue-6));
      background: rgb(var(--arcoblue-1));
    }

    &.is-green {
      color: rgb(var(--green-6));
      background: rgb(var(--green-1));
    }

    &.is-red {
      color: rgb(var(--red-6));
      background: rgb(var(--red-1));
    }

    &.is-orange {
      color: rgb(var(--orange-6));
      background: rgb(var(--orange-1));
    }
  }

  .stat-main {
    display: flex;
    flex-direction: column;
    min-width: 0;
    line-height: 1.5;

    strong {
      color: var(--color-text-1);
      font-size: 16px;
    }

    small {
      overflow: hidden;
      color: var(--color-text-3);
      font-size: 12px;
      white-space: nowrap;
      text-overflow: ellipsis;
    }
  }

  .stat-label {
    color: var(--color-text-2);
    font-size: 12px;
  }

  .game-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 20px;
  }

  .record-tabs {
    margin-bottom: 8px;

    :deep(.arco-tabs-content) {
      display: none;
    }
  }

  .kv {
    display: flex;
    justify-content: space-between;
    margin: 0 0 8px;
    font-size: 13px;

    span {
      color: var(--color-text-3);
    }
  }

  .num {
    font-variant-numeric: tabular-nums;
  }

  .ellipsis {
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .is-win {
    color: rgb(var(--red-6));
  }

  .is-loss {
    color: rgb(var(--green-6));
  }
</style>
