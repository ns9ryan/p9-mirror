<template>
  <div class="member-list-row" :class="{ 'is-selected': selected }">
    <div class="member-list-header" @click="toggleExpanded">
      <a-checkbox :model-value="selected" @click.stop @change="onCheck" />
      <button
        type="button"
        class="expand-btn"
        :aria-expanded="expanded"
        :aria-label="
          expanded
            ? t('memberManage.card.collapse')
            : t('memberManage.card.expand')
        "
        @click.stop="toggleExpanded"
      >
        <icon-down class="expand-icon" :class="{ 'is-open': expanded }" />
      </button>
      <span>{{ t('memberManage.card.memberId') }}：{{ record.id }}</span>
      <span>{{ t('memberManage.card.vip') }}：VIP {{ record.vipLevel }}</span>
      <span>
        {{ t('memberManage.card.status') }}：
        <a-tag
          size="small"
          :color="record.status === 'normal' ? 'green' : 'red'"
        >
          {{ t(`memberManage.status.${record.status}`) }}
        </a-tag>
      </span>
      <span>
        {{ t('memberManage.card.registeredAt') }}：{{ record.registeredAt }}
      </span>
      <a-space
        v-if="!expanded"
        class="row-actions header-actions"
        :size="8"
        @click.stop
      >
        <a-button type="primary" size="small" @click="emit('edit')">
          {{ t('memberManage.card.edit') }}
        </a-button>
        <a-button size="small" @click="emit('detail')">
          {{ t('memberManage.card.detail') }}
        </a-button>
        <a-button
          size="small"
          :status="record.status === 'normal' ? 'danger' : 'success'"
          @click="emit('toggleStatus')"
        >
          {{
            record.status === 'normal'
              ? t('memberManage.card.disable')
              : t('memberManage.card.enable')
          }}
        </a-button>
      </a-space>
    </div>

    <div class="member-list-panel" :class="{ 'is-open': expanded }">
      <div class="member-list-panel__inner">
        <div class="member-list-body">
          <div class="info-block">
            <div class="block-title">
              <span class="block-mark"></span>{{ t('memberManage.card.basic') }}
            </div>
            <div class="block-content">
              <div class="primary-name">{{ displayText(record.account) }}</div>
              <div class="compact-list">
                <div>
                  <strong>{{ t('memberManage.card.phone') }}：</strong>
                  {{ displayText(record.phone) }}
                </div>
                <div class="ellipsis" :title="record.email">
                  <strong>{{ t('memberManage.card.email') }}：</strong>
                  {{ displayText(record.email) }}
                </div>
                <div>
                  <strong>{{ t('memberManage.card.registerType') }}：</strong>
                  {{ t(`memberManage.registerType.${record.registerType}`) }}
                </div>
              </div>
            </div>
          </div>

          <div class="info-block">
            <div class="block-title">
              <span class="block-mark orange"></span>
              {{ t('memberManage.card.agent') }}
            </div>
            <div class="block-content compact-list">
              <div>
                <strong>{{ t('memberManage.card.level') }}：</strong>
                {{ record.level }}
              </div>
              <div>
                <strong>{{ t('memberManage.card.source') }}：</strong>
                {{ record.registerSource }}
              </div>
              <div>
                <strong>{{ t('memberManage.card.channel') }}：</strong>
                {{ displayText(record.channelSource) }}
              </div>
              <div>
                <strong>{{ t('memberManage.card.parent') }}：</strong>
                {{ displayText(record.parentAgent) }}
              </div>
              <div>
                <strong>{{ t('memberManage.card.parentId') }}：</strong>
                {{ displayText(record.parentAgentId) }}
              </div>
            </div>
          </div>

          <div class="info-block">
            <div class="block-title">
              <span class="block-mark blue"></span>
              {{ t('memberManage.card.funds') }}
            </div>
            <div class="block-content compact-list fund-list">
              <div v-for="item in fundItems" :key="item.label">
                <strong>{{ item.label }}：</strong>
                <span class="num" :class="`is-${item.tone}`">{{
                  item.value
                }}</span>
              </div>
            </div>
          </div>

          <div class="info-block">
            <div class="block-title">
              <span class="block-mark green"></span>
              {{ t('memberManage.card.risk') }}
            </div>
            <div class="block-content compact-list">
              <div>
                <strong>{{ t('memberManage.card.lastLogin') }}：</strong>
                {{ displayText(record.lastLoginAt) }}
              </div>
              <div class="ellipsis" :title="record.lastLoginIp">
                <strong>{{ t('memberManage.card.loginIp') }}：</strong>
                {{ displayText(record.lastLoginIp) }}
              </div>
              <div>
                <strong>{{ t('memberManage.card.referral') }}：</strong>
                <a-tag
                  v-if="record.referralStatus === 'normal'"
                  size="small"
                  color="green"
                >
                  {{ t('memberManage.referral.normal') }}
                </a-tag>
                <template v-else>{{ t('memberManage.empty') }}</template>
              </div>
              <div class="tag-line">
                <strong>{{ t('memberManage.card.tags') }}：</strong>
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
              </div>
            </div>
          </div>

          <div class="info-block status-block">
            <div class="block-title">
              <span class="block-mark purple"></span>
              {{ t('memberManage.card.status') }}
            </div>
            <div class="block-content status-wrap">
              <a-tag
                size="medium"
                :color="record.status === 'normal' ? 'green' : 'red'"
              >
                {{ t(`memberManage.status.${record.status}`) }}
              </a-tag>
              <span class="status-description">
                {{ t(`memberManage.levelKind.${record.levelKind}`) }}
              </span>
            </div>
          </div>
        </div>

        <div class="member-list-footer">
          <span>
            {{ t('memberManage.card.lastLogin') }}：{{
              displayText(record.lastLoginAt)
            }}
          </span>
          <span>
            {{ t('memberManage.card.loginIp') }}：{{
              displayText(record.lastLoginIp)
            }}
          </span>
          <a-space class="row-actions" :size="8">
            <a-button type="primary" size="small" @click="emit('edit')">
              {{ t('memberManage.card.edit') }}
            </a-button>
            <a-button size="small" @click="emit('detail')">
              {{ t('memberManage.card.detail') }}
            </a-button>
            <a-button
              size="small"
              :status="record.status === 'normal' ? 'danger' : 'success'"
              @click="emit('toggleStatus')"
            >
              {{
                record.status === 'normal'
                  ? t('memberManage.card.disable')
                  : t('memberManage.card.enable')
              }}
            </a-button>
          </a-space>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import {
    displayText,
    formatAmount,
    formatAmountCount,
    memberTagKeys,
    type MemberTagKey,
  } from '../format';
  import type { MemberRecord } from '../types';

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
    (e: 'toggleStatus'): void;
  }>();

  const { t } = useI18n();

  type FundTone = 'balance' | 'deposit' | 'up' | 'down' | 'zero';

  function signedTone(value: number): FundTone {
    if (value > 0) return 'up';
    if (value < 0) return 'down';
    return 'zero';
  }

  function positiveTone(value: number): FundTone {
    return value > 0 ? 'up' : 'zero';
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
        value: formatAmount(record.todayDeposit),
        tone: positiveTone(record.todayDeposit),
      },
      {
        label: t('memberManage.card.todayWithdraw'),
        value: formatAmount(record.todayWithdraw),
        tone: record.todayWithdraw > 0 ? 'down' : 'zero',
      },
      {
        label: t('memberManage.card.todayBet'),
        value: formatAmount(record.todayValidBet),
        tone: positiveTone(record.todayValidBet),
      },
      {
        label: t('memberManage.card.todayWin'),
        value: formatAmount(record.todayWinLoss),
        tone: signedTone(record.todayWinLoss),
      },
      {
        label: t('memberManage.card.todayBonus'),
        value: formatAmount(record.todayBonus),
        tone: positiveTone(record.todayBonus),
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

  function onCheck(value: boolean | (string | number | boolean)[]) {
    emit('update:selected', value === true);
  }

  function toggleExpanded() {
    emit('update:expanded', !props.expanded);
  }
</script>

<style scoped lang="less">
  .member-list-row {
    overflow: hidden;
    background-color: var(--color-bg-2);
  }

  .member-list-row.is-selected .member-list-header {
    background-color: rgb(var(--primary-1));
  }

  .member-list-header {
    display: flex;
    flex-wrap: wrap;
    gap: 16px;
    align-items: center;
    min-height: 38px;
    padding: 8px 16px;
    color: var(--color-text-2);
    font-size: 13px;
    line-height: 1.6;
    background-color: var(--color-fill-1);
    border-bottom: 1px solid var(--color-border-2);
    cursor: pointer;
  }

  .expand-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    padding: 0;
    color: var(--color-text-2);
    background: transparent;
    border: none;
    border-radius: 4px;
    cursor: pointer;
  }

  .expand-btn:hover {
    background: var(--color-fill-3);
  }

  .expand-icon {
    transform: rotate(-90deg);
    transition: transform 0.2s ease;
  }

  .expand-icon.is-open {
    transform: rotate(0deg);
  }

  .header-actions {
    margin-left: auto;
  }

  .member-list-panel {
    display: grid;
    grid-template-rows: 0fr;
    transition: grid-template-rows 0.2s ease;
  }

  .member-list-panel.is-open {
    grid-template-rows: 1fr;
  }

  .member-list-panel__inner {
    min-height: 0;
    overflow: hidden;
  }

  .member-list-header span {
    display: inline-flex;
    gap: 6px;
    align-items: center;
    white-space: nowrap;
  }

  .member-list-body {
    display: grid;
    grid-template-columns:
      minmax(200px, 1fr) minmax(200px, 1fr) minmax(320px, 1.45fr)
      minmax(220px, 1.05fr) minmax(140px, 0.65fr);
    min-height: 168px;
    background-color: var(--color-bg-2);
  }

  .info-block {
    position: relative;
    min-height: 168px;
    padding: 14px 16px;
    background-color: var(--color-bg-2);
  }

  .info-block:not(:last-child) {
    border-right: 1px solid var(--color-border-2);
  }

  .block-title {
    display: flex;
    gap: 8px;
    align-items: center;
    margin-bottom: 10px;
    color: var(--color-text-2);
    font-weight: 600;
    font-size: 14px;
    letter-spacing: 0.04em;
  }

  .block-mark {
    display: inline-block;
    width: 10px;
    height: 10px;
    background: linear-gradient(135deg, #34d399, #10b981);
    border-radius: 3px;
    box-shadow: 0 0 12px rgba(52, 211, 153, 0.45);
  }

  .block-mark.orange {
    background: linear-gradient(135deg, #fbbf24, #f59e0b);
    box-shadow: 0 0 12px rgba(251, 191, 36, 0.45);
  }

  .block-mark.blue {
    background: linear-gradient(135deg, #60a5fa, #2563eb);
    box-shadow: 0 0 12px rgba(96, 165, 250, 0.45);
  }

  .block-mark.green {
    background: linear-gradient(135deg, #4ade80, #16a34a);
    box-shadow: 0 0 12px rgba(74, 222, 128, 0.45);
  }

  .block-mark.purple {
    background: linear-gradient(135deg, #c084fc, #8b5cf6);
    box-shadow: 0 0 12px rgba(192, 132, 252, 0.45);
  }

  .block-content {
    display: flex;
    flex-direction: column;
    gap: 4px;
    color: var(--color-text-2);
    line-height: 1.6;
  }

  .primary-name {
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 16px;
    line-height: 1.4;
  }

  .compact-list {
    color: var(--color-text-2);
    font-size: 14px;
  }

  .compact-list strong {
    color: var(--color-text-1);
    font-weight: 600;
  }

  .fund-list {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 4px 16px;
  }

  .ellipsis {
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .tag-line {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    align-items: center;
  }

  .num {
    font-variant-numeric: tabular-nums;
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

  .status-block {
    display: flex;
    flex-direction: column;
  }

  .status-wrap {
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    height: 100%;
  }

  .status-wrap :deep(.arco-tag) {
    align-self: flex-start;
  }

  .status-description {
    color: var(--color-text-3);
    font-size: 13px;
  }

  .member-list-footer {
    display: flex;
    gap: 20px;
    align-items: center;
    min-height: 46px;
    padding: 8px 16px;
    color: var(--color-text-2);
    font-size: 13px;
    border-top: 1px solid var(--color-border-2);
  }

  .row-actions {
    display: flex;
    flex-wrap: wrap;
    margin-left: auto;
  }

  .row-actions :deep(.arco-btn) {
    min-width: 52px;
  }
</style>
