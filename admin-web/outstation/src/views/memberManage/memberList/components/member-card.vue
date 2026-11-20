<template>
  <div class="member-item" :class="{ 'is-selected': selected }">
    <div class="member-row" @click="toggleExpanded">
      <span class="check-cell" @click.stop>
        <a-checkbox :model-value="selected" @change="onCheck" />
      </span>
      <button
        type="button"
        class="fold"
        :aria-expanded="expanded"
        :aria-label="
          expanded
            ? t('memberManage.card.collapse')
            : t('memberManage.card.expand')
        "
        @click.stop="toggleExpanded"
      >
        <AgentFoldIcon :open="expanded" />
      </button>
      <span class="num member-id">{{ displayText(record.code) }}</span>
      <span class="stack">
        <span class="account-line">
          <strong class="ellipsis" :title="record.account">
            {{ displayText(record.account) }}
          </strong>
          <a-tooltip v-if="record.remark" :content="record.remark">
            <icon-message class="remark-icon" @click.stop />
          </a-tooltip>
        </span>
        <small class="ellipsis" :title="record.nickname">
          {{ displayText(record.nickname) }}
        </small>
      </span>
      <span class="stack">
        <span class="ellipsis" :title="record.parentAgent">
          {{ displayText(record.parentAgent) }}
        </span>
        <small class="ellipsis">
          ID: {{ displayText(record.parentAgentId) }}
        </small>
      </span>
      <span>
        <a-tag size="small" :color="vipColor(record.vipLevel)">
          VIP{{ record.vipLevel }}
        </a-tag>
      </span>
      <span>
        <a-tag size="small" :color="STATUS_COLORS[record.status]">
          {{ t(`memberManage.status.${record.status}`) }}
        </a-tag>
      </span>
      <span class="stack is-num">
        <strong class="num">{{ formatAmount(record.balance) }}</strong>
        <small class="num">{{ formatAmount(record.unsettled) }}</small>
      </span>
      <span class="stack is-num">
        <span class="num">{{ formatAmount(record.periodDeposit) }}</span>
        <small class="num">{{ formatAmount(record.totalDepositAmount) }}</small>
      </span>
      <span class="stack is-num">
        <span class="num">{{ formatAmount(record.periodWithdraw) }}</span>
        <small class="num">
          {{ formatAmount(record.totalWithdrawAmount) }}
        </small>
      </span>
      <span class="stack is-num">
        <span class="num">{{ formatAmount(record.periodBet) }}</span>
        <small class="num">{{ formatAmount(record.periodValidBet) }}</small>
      </span>
      <span class="stack is-num">
        <span :class="['num', winLossClass(record.periodWinLoss)]">
          {{ formatAmount(record.periodWinLoss) }}
        </span>
        <small :class="['num', winLossClass(record.totalWinLoss)]">
          {{ formatAmount(record.totalWinLoss) }}
        </small>
      </span>
      <span class="stack">
        <span class="ellipsis" :title="record.registerSource">
          {{ registerChannelText(record.registerSource, t) }}
        </span>
        <small>{{ registerMethodText(record.registerType, t) }}</small>
      </span>
      <span class="stack">
        <span class="num">{{ record.registeredAt }}</span>
        <small class="num">{{ displayText(record.lastLoginAt) }}</small>
      </span>
      <span class="ops" @click.stop>
        <a-dropdown :popup-max-height="false" @select="onMore">
          <a-button size="small">
            {{ t('memberManage.card.more') }}
            <icon-down />
          </a-button>
          <template #content>
            <a-doption value="detail">
              {{ t('memberManage.card.detail') }}
            </a-doption>
            <a-doption value="edit">
              {{ t('memberManage.card.edit') }}
            </a-doption>
            <!-- 业务操作：所有状态都显示 -->
            <a-doption value="level">
              {{ t('memberManage.ops.level.menu') }}
            </a-doption>
            <a-doption value="remark">
              {{ t('memberManage.ops.remark.menu') }}
            </a-doption>
            <a-doption value="restriction">
              {{ t('memberManage.ops.restriction.menu') }}
            </a-doption>

            <!-- 状态操作：按当前账号状态显示 -->
            <a-doption v-if="record.status === 'normal'" value="freeze">
              {{ t('memberManage.v2.action.freeze') }}
            </a-doption>
            <a-doption v-if="record.status === 'frozen'" value="unfreeze">
              {{ t('memberManage.v2.action.unfreeze') }}
            </a-doption>
            <a-doption v-if="record.status !== 'disabled'" value="disable">
              {{ t('memberManage.v2.action.disable') }}
            </a-doption>
            <a-doption v-if="record.status === 'disabled'" value="enable">
              {{ t('memberManage.v2.action.enable') }}
            </a-doption>
          </template>
        </a-dropdown>
      </span>
    </div>

    <div class="member-panel" :class="{ 'is-open': expanded }">
      <div class="member-panel__inner">
        <div class="detail-grid">
          <section>
            <h4
              ><i class="dot is-green"></i
              >{{ t('memberManage.card.basic') }}</h4
            >
            <p>
              <span>{{ t('memberManage.card.account') }}</span>
              <strong>{{ displayText(record.account) }}</strong>
            </p>
            <p>
              <span>{{ t('memberManage.card.phone') }}</span>
              {{ displayText(record.phone) }}
            </p>
            <p class="ellipsis-line" :title="record.email">
              <span>{{ t('memberManage.card.email') }}</span>
              <b>{{ displayText(record.email) }}</b>
            </p>
            <p>
              <span>{{ t('memberManage.card.registerType') }}</span>
              {{ registerMethodText(record.registerType, t) }}
            </p>
          </section>

          <section>
            <h4
              ><i class="dot is-orange"></i
              >{{ t('memberManage.card.agent') }}</h4
            >
            <p>
              <span>{{ t('memberManage.levelKind.auto') }}</span>
              {{ displayText(record.autoLevelName) }}
            </p>
            <p class="ellipsis-line" :title="fixedLevelsText(record)">
              <span>{{ t('memberManage.levelKind.fixed') }}</span>
              <b>{{ displayText(fixedLevelsText(record)) }}</b>
            </p>
            <p>
              <span>{{ t('memberManage.card.source') }}</span>
              {{ registerChannelText(record.registerSource, t) }}
            </p>
            <p>
              <span>{{ t('memberManage.card.parent') }}</span>
              {{ displayText(record.parentAgent) }}
            </p>
            <p>
              <span>{{ t('memberManage.card.parentId') }}</span>
              {{ displayText(record.parentAgentId) }}
            </p>
          </section>

          <section>
            <h4
              ><i class="dot is-blue"></i>{{ t('memberManage.card.funds') }}</h4
            >
            <div class="fund-list">
              <p v-for="item in fundItems" :key="item.label">
                <span>{{ item.label }}</span>
                <em class="num" :class="`is-${item.tone}`">{{ item.value }}</em>
              </p>
            </div>
          </section>

          <section>
            <h4
              ><i class="dot is-lime"></i>{{ t('memberManage.card.risk') }}</h4
            >
            <p>
              <span>{{ t('memberManage.card.lastLogin') }}</span>
              {{ displayText(record.lastLoginAt) }}
            </p>
            <p class="ellipsis-line" :title="record.lastLoginIp">
              <span>{{ t('memberManage.card.loginIp') }}</span>
              <b>{{ displayText(record.lastLoginIp) }}</b>
            </p>
            <p>
              <span>{{ t('memberManage.card.referral') }}</span>
              <a-tag
                v-if="record.referralStatus === 'normal'"
                size="small"
                color="green"
              >
                {{ t('memberManage.referral.normal') }}
              </a-tag>
              <template v-else>{{ t('memberManage.empty') }}</template>
            </p>
            <p class="tag-line">
              <span>{{ t('memberManage.card.tags') }}</span>
              <a-space wrap :size="4">
                <a-tag
                  v-for="tag in tags"
                  :key="tag.key"
                  size="small"
                  :color="tagColor(tag.key)"
                >
                  {{ tag.label }}
                </a-tag>
              </a-space>
            </p>
            <p class="ellipsis-line" :title="record.remark">
              <span>{{ t('memberManage.v2.field.remark') }}</span>
              <b>{{ displayText(record.remark) }}</b>
            </p>
          </section>

          <section>
            <h4>
              <i class="dot is-purple"></i>{{ t('memberManage.card.status') }}
            </h4>
            <p>
              <span>{{ t('memberManage.card.status') }}</span>
              <em :class="['state', record.status]">
                {{ t(`memberManage.status.${record.status}`) }}
              </em>
            </p>
            <p>
              <span>{{ t('memberManage.card.vip') }}</span>
              VIP {{ record.vipLevel }}
            </p>
            <p>
              <span>{{ t('memberManage.card.registeredAt') }}</span>
              {{ record.registeredAt }}
            </p>
          </section>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import AgentFoldIcon from '@/views/agent/list/components/agent-fold-icon.vue';
  import {
    STATUS_COLORS,
    displayText,
    formatAmount,
    fixedLevelsText,
    formatAmountCount,
    memberTagKeys,
    registerChannelText,
    registerMethodText,
    vipColor,
    type MemberTagKey,
  } from '../format';
  import type { MemberRecord } from '../types';
  import type { OperationCommand } from './operations/commands';
  import type { StatusCommand } from './v2/actions';

  const props = defineProps<{
    record: MemberRecord;
    selected: boolean;
    expanded: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'update:selected', value: boolean): void;
    (e: 'update:expanded', value: boolean): void;
    (e: 'detail'): void;
    (e: 'edit'): void;
    (e: 'status', command: StatusCommand): void;
    (e: 'operation', command: OperationCommand): void;
  }>();

  const { t } = useI18n();

  type FundTone = 'balance' | 'deposit' | 'up' | 'down' | 'zero';

  function signedTone(value: number): FundTone {
    if (value > 0) return 'up';
    if (value < 0) return 'down';
    return 'zero';
  }

  function positiveTone(value: number | null): FundTone {
    return value != null && value > 0 ? 'up' : 'zero';
  }

  const fundItems = computed(() => {
    const { record } = props;
    return [
      {
        label: t('memberManage.card.balance'),
        value: formatAmount(record.balance),
        tone: 'balance' as FundTone,
      },
      {
        label: t('memberManage.card.unsettled'),
        value: formatAmount(record.unsettled),
        tone: positiveTone(record.unsettled),
      },
      {
        label: t('memberManage.card.totalDeposit'),
        value: formatAmountCount(
          record.totalDepositAmount,
          record.totalDepositCount
        ),
        tone: 'deposit' as FundTone,
      },
      {
        label: t('memberManage.card.totalWithdraw'),
        value: formatAmountCount(
          record.totalWithdrawAmount,
          record.totalWithdrawCount
        ),
        tone: record.totalWithdrawAmount > 0 ? 'down' : 'zero',
      },
      {
        label: t('memberManage.card.diff'),
        value: formatAmount(record.depositWithdrawDiff),
        tone: signedTone(record.depositWithdrawDiff),
      },
      {
        label: t('memberManage.card.firstDeposit'),
        value: formatAmount(record.firstDepositAmount),
        tone: positiveTone(record.firstDepositAmount),
      },
      {
        label: t('memberManage.card.todayDeposit'),
        value: formatAmount(record.periodDeposit),
        tone: positiveTone(record.periodDeposit),
      },
      {
        label: t('memberManage.card.todayWithdraw'),
        value: formatAmount(record.periodWithdraw),
        tone: record.periodWithdraw > 0 ? 'down' : 'zero',
      },
      {
        label: t('memberManage.card.todayBet'),
        value: formatAmount(record.periodValidBet),
        tone: positiveTone(record.periodValidBet),
      },
      {
        label: t('memberManage.card.todayWin'),
        value: formatAmount(record.periodWinLoss),
        tone: signedTone(record.periodWinLoss),
      },
      {
        label: t('memberManage.card.todayBonus'),
        value: formatAmount(record.periodBonus),
        tone: positiveTone(record.periodBonus),
      },
    ];
  });

  const tags = computed(() =>
    memberTagKeys(props.record).map((key) => ({
      key,
      label: t(`memberManage.tag.${key}`),
    }))
  );

  function tagColor(key: MemberTagKey) {
    if (key === 'deposited') return 'green';
    if (key === 'betBanned') return 'red';
    return 'orangered';
  }

  function winLossClass(value: number) {
    if (value > 0) return 'is-win';
    if (value < 0) return 'is-loss';
    return '';
  }

  function onMore(
    value: string | number | Record<string, unknown> | undefined
  ) {
    switch (value) {
      case 'detail':
        emit('detail');
        break;
      case 'edit':
        emit('edit');
        break;
      case 'level':
      case 'remark':
      case 'restriction':
        emit('operation', value);
        break;
      case 'freeze':
      case 'unfreeze':
      case 'disable':
      case 'enable':
        emit('status', value);
        break;
      default:
        break;
    }
  }

  function onCheck(value: boolean | (string | number | boolean)[]) {
    emit('update:selected', value === true);
  }

  function toggleExpanded() {
    emit('update:expanded', !props.expanded);
  }
</script>

<style scoped lang="less">
  .member-item.is-selected {
    background: rgb(var(--primary-1));
  }

  .member-row {
    display: grid;
    grid-template-columns: var(--member-cols);
    gap: 8px;
    align-items: center;
    min-height: 64px;
    padding: 0 16px;
    color: var(--color-text-1);
    font-size: 13px;
    cursor: pointer;
  }

  .member-row > span {
    min-width: 0;
  }

  .check-cell {
    display: inline-flex;
    align-items: center;
  }

  .fold {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    padding: 0;
    color: var(--color-text-3);
    background: transparent;
    border: none;
    cursor: pointer;
  }

  .member-id {
    font-weight: 600;
  }

  .stack {
    display: flex;
    flex-direction: column;
    line-height: 1.5;
  }

  .stack > * {
    min-width: 0;
  }

  .stack strong {
    color: var(--color-text-1);
    font-weight: 600;
  }

  .stack small {
    margin-top: 2px;
    color: var(--color-text-3);
    font-size: 12px;
  }

  .account-line {
    display: flex;
    gap: 4px;
    align-items: center;
  }

  .remark-icon {
    flex: none;
    color: rgb(var(--primary-6));
    font-size: 13px;
    cursor: help;
  }

  .stack.is-num {
    align-items: flex-end;
    text-align: right;
  }

  .stack .is-win {
    color: rgb(var(--red-6));
  }

  .stack .is-loss {
    color: rgb(var(--green-6));
  }

  .ellipsis {
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .num {
    font-variant-numeric: tabular-nums;
  }

  .state {
    font-style: normal;
  }

  .state.normal {
    color: rgb(var(--green-6));
  }

  .state.frozen {
    color: rgb(var(--orange-6));
  }

  .state.disabled {
    color: rgb(var(--red-6));
  }

  .ops {
    display: flex;
    gap: 8px;
    justify-content: center;
  }

  .ops :deep(.arco-btn) {
    padding: 0 10px;
  }

  .member-panel {
    display: grid;
    grid-template-rows: 0fr;
    transition: grid-template-rows 0.2s ease;
  }

  .member-panel.is-open {
    grid-template-rows: 1fr;
  }

  .member-panel__inner {
    min-height: 0;
    overflow: hidden;
  }

  .detail-grid {
    display: grid;
    grid-template-columns:
      minmax(0, 1fr) minmax(0, 1fr) minmax(0, 1.6fr) minmax(0, 1.1fr)
      minmax(0, 0.9fr);
    margin: 0 16px 16px;
    padding: 4px 0;
    background: var(--color-fill-2);
    border-radius: 8px;
  }

  section {
    min-width: 0;
    padding: 14px 18px;
  }

  section + section {
    border-left: 1px solid var(--color-border-2);
  }

  h4 {
    display: flex;
    gap: 8px;
    align-items: center;
    margin: 0 0 10px;
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 14px;
  }

  p {
    display: flex;
    gap: 8px;
    align-items: baseline;
    margin: 0 0 8px;
    color: var(--color-text-1);
    font-size: 13px;
    line-height: 1.6;
  }

  p > span {
    flex: none;
    color: var(--color-text-3);
  }

  p strong {
    font-weight: 600;
  }

  p em {
    font-style: normal;
  }

  .ellipsis-line b {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    font-weight: normal;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .tag-line {
    flex-wrap: wrap;
    align-items: center;
  }

  .fund-list {
    display: grid;
    grid-template-columns: 1fr 1fr;
    column-gap: 16px;
  }

  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }

  .is-green {
    background: #14c9c9;
  }

  .is-orange {
    background: #f59a23;
  }

  .is-blue {
    background: #3b6cff;
  }

  .is-lime {
    background: #00b42a;
  }

  .is-purple {
    background: #7c6bf2;
  }

  .is-balance {
    color: rgb(var(--orange-6));
    font-weight: 600;
  }

  .is-deposit,
  .is-up {
    color: rgb(var(--green-6));
    font-weight: 600;
  }

  .is-down {
    color: rgb(var(--red-6));
    font-weight: 600;
  }

  @media (max-width: 1400px) {
    .detail-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    section + section {
      border-left: none;
    }
  }
</style>
