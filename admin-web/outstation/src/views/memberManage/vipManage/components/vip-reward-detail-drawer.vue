<template>
  <a-drawer
    :visible="record !== null"
    :width="440"
    :title="t('vipManage.detail.title')"
    unmount-on-close
    @cancel="close"
  >
    <template v-if="record">
      <section class="detail-section">
        <div class="section-title">
          {{ t('vipManage.detail.section.member') }}
        </div>
        <dl class="detail-list">
          <dt>{{ t('vipManage.record.account') }}</dt>
          <dd>{{ record.account }}</dd>
          <dt>{{ t('vipManage.record.nickname') }}</dt>
          <dd>{{ record.nickname }}</dd>
          <dt>{{ t('vipManage.detail.agent') }}</dt>
          <dd>{{ record.agentPath.join(' / ') }}</dd>
          <dt>{{ t('vipManage.detail.autoLayer') }}</dt>
          <dd>{{ record.autoLayer }}</dd>
          <dt>{{ t('vipManage.detail.fixedLayer') }}</dt>
          <dd>
            <a-space v-if="record.fixedLayers.length" :size="6" wrap>
              <a-tag
                v-for="layer in record.fixedLayers"
                :key="layer"
                size="small"
                color="arcoblue"
              >
                {{ layer }}
              </a-tag>
            </a-space>
            <span v-else>-</span>
          </dd>
          <dt>{{ t('vipManage.detail.currentLevel') }}</dt>
          <dd>
            <a-tag size="small" :color="levelColor(record.currentLevel)">
              VIP{{ record.currentLevel }}
            </a-tag>
          </dd>
        </dl>
      </section>

      <section class="detail-section">
        <div class="section-title">
          {{ t('vipManage.detail.section.reward') }}
        </div>
        <dl class="detail-list">
          <dt>{{ t('vipManage.record.level') }}</dt>
          <dd>
            <a-tag size="small" :color="levelColor(record.level)">
              VIP{{ record.level }}
            </a-tag>
          </dd>
          <dt>{{ t('vipManage.record.reward') }}</dt>
          <dd>{{ formatAmount(record.reward) }}</dd>
          <dt>{{ t('vipManage.record.mode') }}</dt>
          <dd>{{ t(`vipManage.record.modeText.${record.mode}`) }}</dd>
          <dt>{{ t('vipManage.record.status') }}</dt>
          <dd>
            <a-tag size="small" bordered :color="statusColor[record.status]">
              {{ t(`vipManage.record.statusText.${record.status}`) }}
            </a-tag>
          </dd>
          <dt>{{ t('vipManage.record.washRequired') }}</dt>
          <dd>
            {{ formatAmount(record.washRequired) }}
            <span class="muted">
              {{
                t('vipManage.detail.multiple', {
                  multiple: record.washMultiple,
                })
              }}
            </span>
          </dd>
          <dt>{{ t('vipManage.detail.createdAt') }}</dt>
          <dd>{{ record.promotedAt }}</dd>
          <dt>{{ t('vipManage.record.time') }}</dt>
          <dd>{{ record.claimedAt || '-' }}</dd>
        </dl>
      </section>

      <section class="detail-section">
        <div class="section-title">
          {{ t('vipManage.detail.section.change') }}
        </div>
        <dl class="detail-list">
          <dt>{{ t('vipManage.detail.fromLevel') }}</dt>
          <dd>
            <a-tag size="small" :color="levelColor(record.level - 1)">
              VIP{{ record.level - 1 }}
            </a-tag>
          </dd>
          <dt>{{ t('vipManage.detail.toLevel') }}</dt>
          <dd>
            <a-tag size="small" :color="levelColor(record.level)">
              VIP{{ record.level }}
            </a-tag>
          </dd>
          <dt>{{ t('vipManage.detail.changeType') }}</dt>
          <dd>{{ t('vipManage.detail.changeTypeValue') }}</dd>
          <dt>{{ t('vipManage.detail.changeTime') }}</dt>
          <dd>{{ record.promotedAt }}</dd>
          <dt>{{ t('vipManage.detail.changeReason') }}</dt>
          <dd>{{ t('vipManage.detail.changeReasonValue') }}</dd>
          <dt />
          <dd>
            <a-link class="change-link" @click="viewChanges">
              {{ t('vipManage.detail.viewChanges') }}
            </a-link>
          </dd>
        </dl>
      </section>

      <section class="detail-section">
        <div class="section-title">
          {{ t('vipManage.detail.section.log') }}
        </div>
        <div class="log-list">
          <div
            v-for="(item, index) in logs"
            :key="index"
            class="log-item"
            :class="{ 'is-done': item.done }"
          >
            <span class="log-dot" />
            <span class="log-time">{{ item.time || '-' }}</span>
            <div class="log-body">
              <div class="log-title">{{ item.title }}</div>
              <div v-if="item.desc" class="log-desc">{{ item.desc }}</div>
            </div>
          </div>
        </div>
      </section>
    </template>
    <template #footer>
      <a-button @click="close">{{ t('vipManage.detail.close') }}</a-button>
    </template>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { formatAmount } from '../../memberList/format';
  import { levelColor, statusColor } from '../format';
  import type { VipRewardRecord } from '../types';

  const props = defineProps<{
    record: VipRewardRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'close'): void;
    (e: 'viewChanges', account: string): void;
  }>();

  const { t } = useI18n();

  interface LogItem {
    time: string;
    title: string;
    desc: string;
    done: boolean;
  }

  const logs = computed<LogItem[]>(() => {
    const item = props.record;
    if (!item) return [];
    const washDesc = t('vipManage.detail.log.washDesc', {
      amount: formatAmount(item.washRequired),
    });
    const created: LogItem = {
      time: item.promotedAt,
      title: t('vipManage.detail.log.created'),
      desc: t('vipManage.detail.log.createdDesc', {
        level: item.level,
        amount: formatAmount(item.reward),
      }),
      done: true,
    };
    if (item.status === 'pending') {
      return [
        created,
        {
          time: '',
          title: t('vipManage.detail.log.waiting'),
          desc: washDesc,
          done: false,
        },
        {
          time: '',
          title: t('vipManage.detail.log.notClaimed'),
          desc: '',
          done: false,
        },
      ];
    }
    let finishKey = 'vipManage.detail.log.claimed';
    if (item.status === 'dispatched') {
      finishKey =
        item.mode === 'auto'
          ? 'vipManage.detail.log.autoDispatched'
          : 'vipManage.detail.log.manualDispatched';
    }
    return [
      created,
      { time: item.claimedAt, title: t(finishKey), desc: washDesc, done: true },
    ];
  });

  function close() {
    emit('close');
  }

  function viewChanges() {
    if (props.record) emit('viewChanges', props.record.account);
  }
</script>

<style scoped lang="less">
  .detail-section {
    margin-bottom: 12px;
    padding: 12px 16px;
    border: 1px solid var(--color-neutral-3);
    border-radius: 6px;

    &:last-child {
      margin-bottom: 0;
    }
  }

  .section-title {
    margin-bottom: 10px;
    color: var(--color-text-1);
    font-weight: 600;
  }

  .detail-list {
    display: grid;
    grid-template-columns: 96px 1fr;
    gap: 8px 12px;
    margin: 0;
    font-size: 13px;
    line-height: 22px;

    dt {
      color: var(--color-text-3);
    }

    dd {
      min-width: 0;
      margin: 0;
      color: var(--color-text-1);
      word-break: break-all;
    }
  }

  .change-link {
    padding: 0;
    font-size: 13px;
  }

  .muted {
    margin-left: 4px;
    color: var(--color-text-3);
  }

  .log-list {
    position: relative;
    font-size: 13px;
  }

  .log-item {
    position: relative;
    display: flex;
    gap: 12px;
    padding: 0 0 14px 18px;

    &:last-child {
      padding-bottom: 0;
    }

    &:not(:last-child)::before {
      position: absolute;
      top: 14px;
      bottom: -6px;
      left: 3px;
      width: 1px;
      background: var(--color-neutral-3);
      content: '';
    }
  }

  .log-dot {
    position: absolute;
    top: 7px;
    left: 0;
    width: 7px;
    height: 7px;
    background: var(--color-neutral-4);
    border-radius: 50%;
  }

  .log-item.is-done .log-dot {
    background: rgb(var(--primary-6));
  }

  .log-time {
    flex: none;
    width: 130px;
    color: var(--color-text-3);
    line-height: 22px;
  }

  .log-body {
    flex: 1;
    min-width: 0;
  }

  .log-title {
    color: var(--color-text-1);
    font-weight: 500;
    line-height: 22px;
  }

  .log-desc {
    color: var(--color-text-3);
    font-size: 12px;
    line-height: 18px;
  }
</style>
