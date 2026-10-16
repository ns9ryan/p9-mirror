<template>
  <a-drawer
    :visible="visible"
    :width="760"
    :title="title"
    :footer="false"
    unmount-on-close
    @cancel="close"
  >
    <template v-if="record">
      <a-descriptions
        :column="2"
        bordered
        size="medium"
        :title="t('memberManage.card.basic')"
        class="block"
      >
        <a-descriptions-item :label="t('memberManage.card.memberId')">
          {{ record.id }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.card.vip')">
          VIP {{ record.vipLevel }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.card.account')">
          {{ displayText(record.account) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.card.phone')">
          {{ displayText(record.phone) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.card.email')" :span="2">
          {{ displayText(record.email) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.card.registerType')">
          {{ t(`memberManage.registerType.${record.registerType}`) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.card.status')">
          {{ t(`memberManage.status.${record.status}`) }}
        </a-descriptions-item>
        <a-descriptions-item
          :label="t('memberManage.card.registeredAt')"
          :span="2"
        >
          {{ record.registeredAt }}
        </a-descriptions-item>
      </a-descriptions>

      <a-descriptions
        :column="2"
        bordered
        size="medium"
        :title="t('memberManage.card.agent')"
        class="block"
      >
        <a-descriptions-item :label="t('memberManage.card.level')">
          {{ record.level }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.modal.levelKind')">
          {{ t(`memberManage.levelKind.${record.levelKind}`) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.card.source')">
          {{ record.registerSource }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.card.channel')">
          {{ displayText(record.channelSource) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.modal.parent')">
          {{ displayText(record.parentAgent) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.modal.parentId')">
          {{ displayText(record.parentAgentId) }}
        </a-descriptions-item>
      </a-descriptions>

      <a-descriptions
        :column="2"
        bordered
        size="medium"
        :title="t('memberManage.card.funds')"
        class="block"
      >
        <a-descriptions-item :label="t('memberManage.card.balance')">
          <span class="num is-balance">{{ formatAmount(record.balance) }}</span>
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.card.unsettled')">
          {{ formatAmount(record.unsettled) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.detail.totalDeposit')">
          <span class="num is-up">
            {{
              formatAmountCount(
                record.totalDepositAmount,
                record.totalDepositCount
              )
            }}
          </span>
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.detail.totalWithdraw')">
          {{
            formatAmountCount(
              record.totalWithdrawAmount,
              record.totalWithdrawCount
            )
          }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.detail.diff')">
          {{ formatAmount(record.depositWithdrawDiff) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.card.firstDeposit')">
          {{ formatAmount(record.firstDepositAmount) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.card.todayDeposit')">
          {{ formatAmount(record.todayDeposit) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.card.todayWithdraw')">
          {{ formatAmount(record.todayWithdraw) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.detail.todayBet')">
          {{ formatAmount(record.todayValidBet) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.card.todayWin')">
          <span :class="record.todayWinLoss < 0 ? 'is-down' : 'is-up'">
            {{ formatAmount(record.todayWinLoss) }}
          </span>
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.card.todayBonus')">
          {{ formatAmount(record.todayBonus) }}
        </a-descriptions-item>
      </a-descriptions>

      <a-descriptions
        :column="2"
        bordered
        size="medium"
        :title="t('memberManage.card.risk')"
        class="block"
      >
        <a-descriptions-item :label="t('memberManage.card.lastLogin')">
          {{ displayText(record.lastLoginAt) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.card.loginIp')">
          {{ displayText(record.lastLoginIp) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.card.referral')">
          {{
            record.referralStatus === 'normal'
              ? t('memberManage.referral.normal')
              : t('memberManage.empty')
          }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.search.betBanned')">
          {{ record.betBanned ? t('front.common.yes') : t('front.common.no') }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.card.tags')" :span="2">
          <a-space wrap>
            <a-tag v-for="tag in tags" :key="tag" size="small">
              {{ t(`memberManage.tag.${tag}`) }}
            </a-tag>
          </a-space>
        </a-descriptions-item>
      </a-descriptions>

      <div class="drawer-actions">
        <a-button @click="close">{{ t('memberManage.detail.close') }}</a-button>
        <a-button type="primary" @click="emit('edit')">
          {{ t('memberManage.card.edit') }}
        </a-button>
      </div>
    </template>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import {
    displayText,
    formatAmount,
    formatAmountCount,
    memberTagKeys,
  } from '../format';
  import type { MemberRecord } from '../types';

  const props = defineProps<{
    visible: boolean;
    record: MemberRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'edit'): void;
  }>();

  const { t } = useI18n();

  const title = computed(() => {
    if (!props.record) return t('memberManage.detail.title');
    return `${t('memberManage.detail.title')} · ${props.record.id}`;
  });

  const tags = computed(() =>
    props.record ? memberTagKeys(props.record) : []
  );

  function close() {
    emit('update:visible', false);
  }
</script>

<style scoped lang="less">
  .block {
    margin-bottom: 16px;
  }

  .num {
    font-variant-numeric: tabular-nums;
  }

  .is-balance {
    color: #d48806;
    font-weight: 600;
  }

  .is-up {
    color: #009a29;
    font-weight: 600;
  }

  .is-down {
    color: #cb272d;
    font-weight: 600;
  }

  .drawer-actions {
    display: flex;
    gap: 8px;
    justify-content: flex-end;
  }
</style>
