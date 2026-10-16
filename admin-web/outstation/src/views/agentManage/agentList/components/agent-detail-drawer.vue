<template>
  <a-drawer
    :visible="visible"
    :width="480"
    :title="t('agentList.detail.title')"
    :footer="false"
    unmount-on-close
    @cancel="emit('update:visible', false)"
  >
    <template v-if="record">
      <div class="drawer-profile">
        <span class="level-orb" :style="{ background: color }">
          {{ record.level }}
        </span>
        <div>
          <strong>{{ record.account }}</strong>
          <p>ID：{{ record.id }}</p>
        </div>
        <em :class="record.status">
          {{ t(`agentList.status.${record.status}`) }}
        </em>
      </div>
      <a-tabs default-active-key="basic">
        <a-tab-pane key="basic" :title="t('agentList.card.basic')">
          <a-descriptions :column="1" size="medium">
            <a-descriptions-item :label="t('agentList.card.level')">
              {{ record.level }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('agentList.card.parent')">
              {{ displayText(record.parentAgent) }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('agentList.card.region')">
              {{ t('agentList.empty') }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('agentList.card.source')">
              {{ record.registerSource }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('agentList.card.registeredAt')">
              {{ record.registeredAt }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('agentList.card.lastLogin')">
              {{ displayText(record.lastLoginAt) }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('agentList.card.remark')">
              {{ displayText(record.remark) }}
            </a-descriptions-item>
          </a-descriptions>
        </a-tab-pane>
        <a-tab-pane key="team" :title="t('agentList.card.team')">
          <a-descriptions :column="1" size="medium">
            <a-descriptions-item :label="t('agentList.card.directMembers')">
              {{ record.directMembers }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('agentList.card.directAgents')">
              {{ record.directAgents }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('agentList.card.allMembers')">
              {{ record.allMembers }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('agentList.card.allAgents')">
              {{ record.allAgents }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('agentList.card.share')">
              {{ formatRatio(record.shareRatio) }}
            </a-descriptions-item>
          </a-descriptions>
        </a-tab-pane>
        <a-tab-pane key="funds" :title="t('agentList.card.funds')">
          <a-descriptions :column="1" size="medium">
            <a-descriptions-item :label="t('agentList.card.balance')">
              ¥{{ formatAmount(record.balance) }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('agentList.card.commission')">
              ¥{{ formatAmount(record.totalCommission) }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('agentList.card.unpaid')">
              ¥{{ formatAmount(record.unpaidCommission) }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('agentList.card.withdrawLimit')">
              {{ t(`agentList.withdraw.${record.withdrawLimit}`) }}
            </a-descriptions-item>
          </a-descriptions>
        </a-tab-pane>
        <a-tab-pane key="promo" :title="t('agentList.card.promo')">
          <a-descriptions :column="1" size="medium">
            <a-descriptions-item :label="t('agentList.card.link')">
              {{ promoLink }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('agentList.card.uv')">
              {{ record.visitUv }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('agentList.card.lastLogin')">
              {{ displayText(record.lastLoginAt) }}
            </a-descriptions-item>
          </a-descriptions>
        </a-tab-pane>
        <a-tab-pane key="status" :title="t('agentList.card.status')">
          <a-descriptions :column="1" size="medium">
            <a-descriptions-item :label="t('agentList.column.status')">
              {{ t(`agentList.status.${record.status}`) }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('agentList.card.online')">
              {{ t('agentList.empty') }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('agentList.card.device')">
              {{ t('agentList.empty') }}
            </a-descriptions-item>
          </a-descriptions>
        </a-tab-pane>
      </a-tabs>
      <div class="drawer-actions">
        <a-button @click="emit('update:visible', false)">
          {{ t('agentList.detail.close') }}
        </a-button>
        <a-button type="primary" @click="emit('edit')">
          {{ t('agentList.card.edit') }}
        </a-button>
      </div>
    </template>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { displayText, formatAmount, formatRatio } from '../format';
  import { buildPromoLink, levelColor } from '../options';
  import type { AgentRecord } from '../types';

  const props = defineProps<{
    visible: boolean;
    record: AgentRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'edit'): void;
  }>();

  const { t } = useI18n();

  const color = computed(() =>
    props.record ? levelColor(props.record.level) : '#64748b'
  );

  const promoLink = computed(() =>
    props.record
      ? buildPromoLink(props.record.promoHost, props.record.promoCode)
      : ''
  );
</script>

<style scoped lang="less">
  .drawer-profile {
    display: flex;
    gap: 12px;
    align-items: center;
    margin-bottom: 16px;
  }

  .drawer-profile strong {
    color: var(--color-text-1);
    font-size: 16px;
  }

  .drawer-profile p {
    margin: 4px 0 0;
    color: var(--color-text-3);
    font-size: 12px;
  }

  .drawer-profile em {
    margin-left: auto;
    font-style: normal;
  }

  .drawer-profile em.normal {
    color: #00b42a;
  }

  .drawer-profile em.disabled {
    color: #f53f3f;
  }

  .level-orb {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 42px;
    height: 42px;
    color: #fff;
    font-weight: 700;
    border-radius: 50%;
  }

  .drawer-actions {
    display: flex;
    gap: 8px;
    justify-content: flex-end;
    margin-top: 16px;
  }
</style>
