<template>
  <a-drawer
    :visible="visible"
    :width="760"
    :title="title"
    :footer="false"
    unmount-on-close
    @cancel="close"
  >
    <a-spin v-if="record" :loading="loading" class="drawer-body">
      <a-descriptions
        :column="2"
        bordered
        size="medium"
        :title="t('memberManage.card.basic')"
        class="block"
      >
        <a-descriptions-item :label="t('memberManage.column.memberId')">
          {{ displayText(record.code) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.card.vip')">
          VIP {{ record.vipLevel }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.card.account')">
          {{ displayText(record.account) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.detail.nickname')">
          {{ displayText(record.nickname) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.detail.gender')">
          {{ detail ? t(`memberManage.gender.${detail.gender}`) : '--' }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.detail.birthday')">
          {{ displayText(detail?.birthday || '') }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.detail.region')">
          {{ displayText(detail?.regionCode || '') }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.card.phone')">
          {{ displayText(record.phone) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.card.email')" :span="2">
          {{ displayText(record.email) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.card.registerType')">
          {{ registerMethodText(record.registerType, t) }}
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
        <a-descriptions-item
          :label="t('memberManage.v2.field.remark')"
          :span="2"
        >
          <span class="remark-text">{{ displayText(record.remark) }}</span>
        </a-descriptions-item>
      </a-descriptions>

      <a-descriptions
        :column="2"
        bordered
        size="medium"
        :title="t('memberManage.card.agent')"
        class="block"
      >
        <a-descriptions-item :label="t('memberManage.levelKind.auto')">
          {{ displayText(record.autoLevelName) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.levelKind.fixed')">
          {{ displayText(fixedLevelsText(record)) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.card.source')" :span="2">
          {{ registerChannelText(record.registerSource, t) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.modal.parent')">
          {{ displayText(record.parentAgent) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.modal.parentId')">
          {{ displayText(record.parentAgentId) }}
        </a-descriptions-item>
        <a-descriptions-item
          :label="t('memberManage.detail.agentLine')"
          :span="2"
        >
          {{ displayText(detail?.agentLineCode || '') }}
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
          {{ formatAmount(record.periodDeposit) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.card.todayWithdraw')">
          {{ formatAmount(record.periodWithdraw) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.detail.todayBet')">
          {{ formatAmount(record.periodValidBet) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.card.todayWin')">
          <span :class="record.periodWinLoss < 0 ? 'is-down' : 'is-up'">
            {{ formatAmount(record.periodWinLoss) }}
          </span>
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.card.todayBonus')">
          {{ formatAmount(record.periodBonus) }}
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
        <a-descriptions-item :label="t('memberManage.detail.registerIp')">
          {{ displayText(detail?.registerIp || '') }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.detail.registerDevice')">
          {{ displayText(detail?.registerDeviceCode || '') }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.detail.registerDomain')">
          {{ displayText(detail?.registerDomain || '') }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.detail.inviter')">
          {{ displayText(detail?.inviterMemberId || '') }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.card.referral')">
          {{
            record.referralStatus === 'normal'
              ? t('memberManage.referral.normal')
              : t('memberManage.empty')
          }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('memberManage.detail.restriction')">
          {{ detail ? restrictionText(detail, t) : '--' }}
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
        <a-dropdown position="top" @select="onOperation">
          <a-button>
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
          </template>
        </a-dropdown>
        <a-button type="primary" @click="emit('edit')">
          {{ t('memberManage.card.edit') }}
        </a-button>
      </div>
    </a-spin>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import {
    displayText,
    fixedLevelsText,
    formatAmount,
    formatAmountCount,
    memberTagKeys,
    registerChannelText,
    registerMethodText,
    restrictionText,
  } from '../format';
  import type { MemberDetailRecord, MemberRecord } from '../types';
  import {
    OPERATION_COMMANDS,
    isOperationCommand,
    type OperationCommand,
  } from './operations/commands';

  const props = defineProps<{
    visible: boolean;
    record: MemberRecord | null;
    /** 详情接口数据，加载完成前为 null，此时补充字段显示 -- */
    detail: MemberDetailRecord | null;
    loading: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'edit'): void;
    (e: 'operation', command: OperationCommand): void;
  }>();

  const { t } = useI18n();

  const title = computed(() => {
    if (!props.record) return t('memberManage.detail.title');
    return `${t('memberManage.detail.title')} · ${props.record.code}`;
  });

  const tags = computed(() =>
    props.record ? memberTagKeys(props.record) : []
  );

  function close() {
    emit('update:visible', false);
  }

  function onOperation(
    value: string | number | Record<string, unknown> | undefined
  ) {
    if (isOperationCommand(value)) emit('operation', value);
  }
</script>

<style scoped lang="less">
  .drawer-body {
    display: block;
  }

  .block {
    margin-bottom: 16px;
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

  .remark-text {
    white-space: pre-wrap;
    word-break: break-all;
  }

  .drawer-actions {
    display: flex;
    gap: 8px;
    justify-content: flex-end;
  }
</style>
