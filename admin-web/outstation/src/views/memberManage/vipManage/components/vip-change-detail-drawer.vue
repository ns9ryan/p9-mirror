<template>
  <a-drawer
    :visible="record !== null"
    :width="440"
    :title="t('vipManage.change.detailTitle')"
    unmount-on-close
    @cancel="close"
  >
    <template v-if="record">
      <div class="section-title">{{ t('vipManage.change.section.base') }}</div>
      <dl class="detail-box">
        <div class="detail-row">
          <dt>{{ t('vipManage.record.account') }}</dt>
          <dd>{{ record.account }}</dd>
        </div>
        <div class="detail-row">
          <dt>{{ t('vipManage.change.memberId') }}</dt>
          <dd>{{ record.memberId }}</dd>
        </div>
      </dl>

      <div class="section-title">{{ t('vipManage.change.section.info') }}</div>
      <dl class="detail-box">
        <div class="detail-row">
          <dt>{{ t('vipManage.change.from') }}</dt>
          <dd><VipLevelBadge :level="record.fromLevel" :size="18" /></dd>
        </div>
        <div class="detail-row">
          <dt>{{ t('vipManage.change.to') }}</dt>
          <dd><VipLevelBadge :level="record.toLevel" :size="18" /></dd>
        </div>
        <div class="detail-row">
          <dt>{{ t('vipManage.change.type') }}</dt>
          <dd>
            <a-tag bordered :color="changeColor(record)">
              {{ t(`vipManage.change.typeText.${record.type}`) }}
            </a-tag>
          </dd>
        </div>
        <div class="detail-row">
          <dt>{{ t('vipManage.change.reason') }}</dt>
          <dd>{{ record.reason }}</dd>
        </div>
        <div class="detail-row">
          <dt>{{ t('vipManage.change.operator') }}</dt>
          <dd>{{ record.operator || t('vipManage.change.system') }}</dd>
        </div>
        <div class="detail-row">
          <dt>{{ t('vipManage.change.time') }}</dt>
          <dd>{{ record.changedAt }}</dd>
        </div>
        <div class="detail-row">
          <dt>{{ t('vipManage.change.reward') }}</dt>
          <dd>{{ record.rewardOrderNo || t('vipManage.change.none') }}</dd>
        </div>
        <div class="detail-row">
          <dt>{{ t('vipManage.change.remark') }}</dt>
          <dd>{{ record.remark || '-' }}</dd>
        </div>
      </dl>
    </template>
    <template #footer>
      <a-button @click="close">{{ t('vipManage.detail.close') }}</a-button>
    </template>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { useI18n } from 'vue-i18n';
  import { changeColor } from '../format';
  import type { VipLevelChangeRecord } from '../types';
  import VipLevelBadge from './vip-level-badge.vue';

  defineProps<{
    record: VipLevelChangeRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'close'): void;
  }>();

  const { t } = useI18n();

  function close() {
    emit('close');
  }
</script>

<style scoped lang="less">
  .section-title {
    position: relative;
    margin: 4px 0 12px;
    padding-left: 12px;
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 15px;
    line-height: 22px;

    &::before {
      position: absolute;
      top: 3px;
      bottom: 3px;
      left: 0;
      width: 3px;
      background: rgb(var(--primary-6));
      border-radius: 2px;
      content: '';
    }
  }

  .detail-box {
    margin: 0 0 24px;
    padding: 0 16px;
    background: var(--color-fill-1);
    border: 1px solid var(--color-neutral-2);
    border-radius: 6px;

    &:last-child {
      margin-bottom: 0;
    }
  }

  .detail-row {
    display: flex;
    gap: 16px;
    align-items: center;
    min-height: 44px;
    padding: 10px 0;
    font-size: 13px;
    line-height: 22px;

    & + & {
      border-top: 1px solid var(--color-neutral-3);
    }

    dt {
      flex: none;
      width: 80px;
      color: var(--color-text-3);
    }

    dd {
      flex: 1;
      min-width: 0;
      margin: 0;
      color: var(--color-text-1);
      word-break: break-all;
    }
  }
</style>
