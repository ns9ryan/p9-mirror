<template>
  <div
    class="agent-list-row"
    :class="{ 'is-selected': selected, 'is-open': expanded }"
  >
    <div class="agent-list-header" @click="toggleExpanded">
      <a-checkbox :model-value="selected" @click.stop @change="onCheck" />
      <span class="head-id">
        {{ t('agentList.card.agentId') }}：{{ record.id }}
      </span>
      <span class="head-account">
        <span class="level-badge" :style="{ background: levelColor }">
          {{ record.level }}
        </span>
        {{ record.account }}
      </span>
      <span class="head-status">
        <i class="status-dot" :class="`is-${record.status}`"></i>
        {{ t('agentList.card.status') }}：{{
          t(`agentList.status.${record.status}`)
        }}
      </span>
      <span class="head-time">
        {{ t('agentList.card.registeredAt') }}：{{ record.registeredAt }}
      </span>
      <a-space
        v-if="!expanded"
        class="row-actions header-actions"
        :size="8"
        @click.stop
      >
        <a-button type="primary" size="small" @click="emit('detail')">
          {{ t('agentList.card.detail') }}
        </a-button>
        <a-dropdown trigger="click" @select="onOperate">
          <a-button size="small">{{ t('agentList.card.operate') }}</a-button>
          <template #content>
            <a-doption value="edit">{{ t('agentList.card.edit') }}</a-doption>
            <a-doption value="toggle">{{ statusActionLabel }}</a-doption>
          </template>
        </a-dropdown>
      </a-space>
    </div>

    <div class="agent-list-panel" :class="{ 'is-open': expanded }">
      <div class="agent-list-panel__inner">
        <div class="agent-list-body">
          <div class="info-block">
            <div class="block-title">
              <span class="block-mark"></span>{{ t('agentList.card.basic') }}
            </div>
            <div class="block-content">
              <div class="profile">
                <a-avatar :size="46" class="profile-avatar">
                  <icon-user />
                </a-avatar>
                <div class="primary-name">{{ record.account }}</div>
              </div>
              <div class="compact-list">
                <div>
                  <strong>{{ t('agentList.card.address') }}：</strong>
                  {{ record.level }} ·
                  {{ t('agentList.card.layer', { layer: layer }) }}
                </div>
                <div>
                  <strong>{{ t('agentList.card.parent') }}：</strong>
                  {{ displayText(record.parentAgent) }}
                </div>
                <div>
                  <strong>{{ t('agentList.card.origin') }}：</strong>
                  {{ displayText(record.topAgent) }}
                </div>
                <div>
                  <strong>{{ t('agentList.card.source') }}：</strong>
                  {{ record.registerSource }}
                </div>
              </div>
            </div>
          </div>

          <div class="info-block">
            <div class="block-title">
              <span class="block-mark orange"></span>
              {{ t('agentList.card.team') }}
            </div>
            <div class="block-content compact-list">
              <div>
                <strong>{{ t('agentList.card.pushMembers') }}：</strong>
                {{ record.directMembers }}
              </div>
              <div>
                <strong>{{ t('agentList.card.pushAgents') }}：</strong>
                {{ record.directAgents }}
              </div>
              <div>
                <strong>{{ t('agentList.card.allAgents') }}：</strong>
                {{ record.allAgents }}
              </div>
              <div>
                <strong>{{ t('agentList.card.allMembers') }}：</strong>
                {{ record.allMembers }}
              </div>
              <div>
                <strong>{{ t('agentList.card.shareHold') }}：</strong>
                {{ formatRatio(record.shareRatio) }}
              </div>
            </div>
          </div>

          <div class="info-block">
            <div class="block-title">
              <span class="block-mark blue"></span>
              {{ t('agentList.card.funds') }}
            </div>
            <div class="block-content compact-list">
              <div>
                <strong>{{ t('agentList.card.available') }}：</strong>
                <span class="num is-balance">
                  {{ formatAmount(record.balance) }}
                </span>
              </div>
              <div>
                <strong>{{ t('agentList.card.commission') }}：</strong>
                <span class="num" :class="commissionTone">
                  {{ formatAmount(record.totalCommission) }}
                </span>
              </div>
              <div>
                <strong>{{ t('agentList.card.unpaid') }}：</strong>
                <span class="num" :class="unpaidTone">
                  {{ formatAmount(record.unpaidCommission) }}
                </span>
              </div>
            </div>
          </div>

          <div class="info-block">
            <div class="block-title">
              <span class="block-mark green"></span>
              {{ t('agentList.card.promo') }}
            </div>
            <div class="block-content compact-list">
              <div>
                <strong>{{ t('agentList.card.payoutRule') }}：</strong>
                {{ t(`agentList.withdraw.${record.withdrawLimit}`) }}
              </div>
              <div class="ellipsis" :title="record.remark">
                <strong>{{ t('agentList.card.remark') }}：</strong>
                {{ displayText(record.remark) }}
              </div>
              <div class="link-line" @click.stop>
                <strong>{{ t('agentList.card.link') }}：</strong>
                <a class="promo-link" :href="promoLink" target="_blank">
                  {{ promoLink }}
                </a>
                <a-button type="text" size="mini" @click="copyLink">
                  {{ t('agentList.card.copy') }}
                </a-button>
              </div>
              <div>
                <strong>{{ t('agentList.card.uv') }}：</strong>
                {{ record.visitUv }}
              </div>
              <div>
                <strong>{{ t('agentList.card.lastLogin') }}：</strong>
                {{ displayText(record.lastLoginAt) }}
              </div>
            </div>
          </div>

          <div class="info-block status-block">
            <div class="block-title">
              <span class="block-mark purple"></span>
              {{ t('agentList.card.status') }}
            </div>
            <div class="block-content status-wrap">
              <a-tag
                size="medium"
                :color="record.status === 'normal' ? 'green' : 'red'"
              >
                {{ t(`agentList.status.${record.status}`) }}
              </a-tag>
              <div class="channel-row">
                <span class="channel-icon is-weibo"><icon-weibo /></span>
                <span class="channel-icon is-wechat"><icon-wechat /></span>
                <span class="channel-icon is-qq"><icon-qq /></span>
                <span class="channel-icon is-twitter"><icon-twitter /></span>
              </div>
            </div>
          </div>
        </div>

        <div class="agent-list-footer">
          <span>
            {{ t('agentList.card.lastLogin') }}：{{
              displayText(record.lastLoginAt)
            }}
          </span>
          <span>{{ t('agentList.card.uv') }}：{{ record.visitUv }}</span>
          <a-space class="row-actions" :size="8" @click.stop>
            <a-button type="primary" size="small" @click="emit('detail')">
              {{ t('agentList.card.detail') }}
            </a-button>
            <a-dropdown trigger="click" @select="onOperate">
              <a-button size="small">{{
                t('agentList.card.operate')
              }}</a-button>
              <template #content>
                <a-doption value="edit">
                  {{ t('agentList.card.edit') }}
                </a-doption>
                <a-doption value="toggle">{{ statusActionLabel }}</a-doption>
              </template>
            </a-dropdown>
          </a-space>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { displayText, formatAmount, formatRatio } from '../format';
  import { buildPromoLink, levelLayer } from '../options';
  import type { AgentRecord } from '../types';

  const props = defineProps<{
    record: AgentRecord;
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

  const LEVEL_COLOR: Record<string, string> = {
    A1: '#0f9f8f',
    A2: '#2f6bff',
    A3: '#7c3aed',
    A4: '#f08a24',
    A5: '#e23b8a',
    A6: '#e11d48',
    A7: '#64748b',
  };

  const layer = computed(() => levelLayer(props.record.level));
  const levelColor = computed(
    () => LEVEL_COLOR[props.record.level] || '#64748b'
  );
  const promoLink = computed(() =>
    buildPromoLink(props.record.promoHost, props.record.promoCode)
  );
  const statusActionLabel = computed(() =>
    props.record.status === 'normal'
      ? t('agentList.card.disable')
      : t('agentList.card.enable')
  );
  const commissionTone = computed(() =>
    props.record.totalCommission > 0 ? 'is-up' : ''
  );
  const unpaidTone = computed(() => {
    if (props.record.unpaidCommission < 0) return 'is-down';
    if (props.record.unpaidCommission > 0) return 'is-up';
    return '';
  });

  function onCheck(value: boolean | (string | number | boolean)[]) {
    emit('update:selected', value === true);
  }

  function toggleExpanded() {
    emit('update:expanded', !props.expanded);
  }

  function onOperate(
    value: string | number | Record<string, unknown> | undefined
  ) {
    if (value === 'edit') emit('edit');
    if (value === 'toggle') emit('toggleStatus');
  }

  async function copyLink() {
    try {
      await navigator.clipboard.writeText(promoLink.value);
      Message.success(t('agentList.message.copied'));
    } catch {
      Message.warning(t('agentList.message.copyFail'));
    }
  }
</script>

<style scoped lang="less">
  .agent-list-row {
    min-width: 980px;
    overflow: hidden;
    background: #fff;
    border: 1px solid #e4e6f0;
    border-radius: 8px;
  }

  .agent-list-row.is-selected {
    border-color: rgb(var(--primary-6));
    box-shadow: 0 0 0 1px rgb(var(--primary-6));
  }

  .agent-list-header {
    display: flex;
    flex-wrap: nowrap;
    gap: 18px;
    align-items: center;
    min-height: 44px;
    padding: 8px 14px;
    color: var(--color-text-1);
    font-size: 13px;
    line-height: 1.5;
    background: #e5e6eb;
    cursor: pointer;
  }

  .agent-list-header span {
    display: inline-flex;
    gap: 6px;
    align-items: center;
    white-space: nowrap;
  }

  .agent-list-header :deep(.arco-checkbox-icon) {
    background: #fff;
    border-color: #c9cdd4;
  }

  .agent-list-header :deep(.arco-checkbox-checked .arco-checkbox-icon),
  .agent-list-header :deep(.arco-checkbox-indeterminate .arco-checkbox-icon) {
    color: #fff;
    background: rgb(var(--primary-6));
    border-color: rgb(var(--primary-6));
  }

  .level-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 28px;
    height: 20px;
    padding: 0 6px;
    color: #fff;
    font-weight: 700;
    font-size: 12px;
    border-radius: 10px;
  }

  .status-dot {
    width: 6px;
    height: 6px;
    background: #23c343;
    border-radius: 50%;
  }

  .status-dot.is-disabled {
    background: #f76965;
  }

  .header-actions {
    flex: none;
    margin-left: auto;
  }

  .header-actions :deep(.arco-btn-secondary) {
    color: var(--color-text-1);
    background: #fff;
    border-color: #fff;
  }

  .agent-list-panel {
    display: grid;
    grid-template-rows: 0fr;
    transition: grid-template-rows 0.2s ease;
  }

  .agent-list-panel.is-open {
    grid-template-rows: 1fr;
  }

  .agent-list-panel__inner {
    min-height: 0;
    overflow: hidden;
  }

  .agent-list-body {
    display: grid;
    grid-template-columns:
      minmax(200px, 1.05fr) minmax(180px, 0.9fr) minmax(200px, 0.95fr)
      minmax(280px, 1.35fr) minmax(140px, 0.6fr);
    min-height: 168px;
    background-color: var(--color-bg-2);
  }

  .info-block {
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
  }

  .block-mark {
    display: inline-block;
    width: 10px;
    height: 10px;
    background: linear-gradient(135deg, #34d399, #10b981);
    border-radius: 3px;
  }

  .block-mark.orange {
    background: linear-gradient(135deg, #fbbf24, #f59e0b);
  }

  .block-mark.blue {
    background: linear-gradient(135deg, #60a5fa, #2563eb);
  }

  .block-mark.green {
    background: linear-gradient(135deg, #4ade80, #16a34a);
  }

  .block-mark.purple {
    background: linear-gradient(135deg, #c084fc, #8b5cf6);
  }

  .block-content {
    display: flex;
    flex-direction: column;
    gap: 4px;
    color: var(--color-text-2);
    line-height: 1.6;
  }

  .profile {
    display: flex;
    gap: 12px;
    align-items: center;
    margin-bottom: 8px;
  }

  .profile-avatar {
    flex: none;
    color: rgb(var(--primary-6));
    background: rgb(var(--primary-1));
  }

  .primary-name {
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 16px;
  }

  .compact-list {
    color: var(--color-text-2);
    font-size: 14px;
  }

  .compact-list strong {
    color: var(--color-text-1);
    font-weight: 600;
  }

  .ellipsis {
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .link-line {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    align-items: center;
  }

  .promo-link {
    max-width: 220px;
    overflow: hidden;
    color: rgb(var(--primary-6));
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .num {
    font-variant-numeric: tabular-nums;
  }

  .is-balance {
    color: rgb(var(--orange-6));
    font-weight: 600;
  }

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

  .channel-row {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .channel-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    color: #fff;
    font-size: 16px;
    border-radius: 50%;
  }

  .channel-icon.is-weibo {
    background: #e6162d;
  }

  .channel-icon.is-wechat {
    background: #2aae67;
  }

  .channel-icon.is-qq {
    background: #12b7f5;
  }

  .channel-icon.is-twitter {
    background: #1d9bf0;
  }

  .agent-list-footer {
    display: flex;
    gap: 20px;
    align-items: center;
    min-height: 46px;
    padding: 8px 16px;
    color: var(--color-text-2);
    font-size: 13px;
    background: #f7f8fc;
    border-top: 1px solid var(--color-border-2);
  }

  .row-actions {
    margin-left: auto;
  }

  .row-actions :deep(.arco-btn) {
    min-width: 52px;
  }
</style>
