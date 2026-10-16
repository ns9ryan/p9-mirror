<template>
  <div class="detail-grid">
    <section>
      <h4><i class="dot is-green"></i>{{ t('agentList.card.basic') }}</h4>
      <p>
        <span>{{ t('agentList.card.level') }}</span>
        {{ record.level }}
      </p>
      <p>
        <span>{{ t('agentList.card.parent') }}</span>
        {{ displayText(record.parentAgent) }}
      </p>
      <p>
        <span>{{ t('agentList.card.directDown') }}</span>
        {{ childCount }}
      </p>
      <p>
        <span>{{ t('agentList.card.region') }}</span>
        {{ t('agentList.empty') }}
      </p>
    </section>
    <section>
      <h4><i class="dot is-blue"></i>{{ t('agentList.card.funds') }}</h4>
      <p>
        <span>{{ t('agentList.card.balance') }}</span>
        <strong class="money">¥{{ formatAmount(record.balance) }}</strong>
      </p>
      <p>
        <span>{{ t('agentList.card.commission') }}</span>
        ¥{{ formatAmount(record.totalCommission) }}
      </p>
      <p>
        <span>{{ t('agentList.card.monthCommission') }}</span>
        {{ t('agentList.empty') }}
      </p>
    </section>
    <section>
      <h4><i class="dot is-orange"></i>{{ t('agentList.card.promo') }}</h4>
      <p class="link-line">
        <span>{{ t('agentList.card.link') }}</span>
        <a :href="promoLink" target="_blank">{{ promoLink }}</a>
      </p>
      <p>
        <span>{{ t('agentList.card.online') }}</span>
        {{ t('agentList.empty') }}
      </p>
      <p>
        <span>{{ t('agentList.card.lastLogin') }}</span>
        {{ displayText(record.lastLoginAt) }}
      </p>
    </section>
    <section>
      <h4><i class="dot is-purple"></i>{{ t('agentList.card.status') }}</h4>
      <p>
        <span>{{ t('agentList.column.status') }}</span>
        <em :class="record.status">{{
          t(`agentList.status.${record.status}`)
        }}</em>
      </p>
      <p>
        <span>{{ t('agentList.card.registeredAt') }}</span>
        {{ record.registeredAt }}
      </p>
      <p>
        <span>{{ t('agentList.card.device') }}</span>
        {{ t('agentList.empty') }}
      </p>
    </section>
  </div>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { displayText, formatAmount } from '../format';
  import { buildPromoLink } from '../options';
  import type { AgentRecord } from '../types';

  const props = defineProps<{
    record: AgentRecord;
    childCount: number;
  }>();

  const { t } = useI18n();
  const promoLink = computed(() =>
    buildPromoLink(props.record.promoHost, props.record.promoCode)
  );
</script>

<style scoped lang="less">
  .detail-grid {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    margin: 0 16px 16px;
    padding: 4px 0;
    background: #f7f8fa;
    border-radius: 8px;
  }

  section {
    padding: 14px 18px;
  }

  section + section {
    border-left: 1px solid #e5e6eb;
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

  span {
    flex: none;
    color: var(--color-text-3);
  }

  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }

  .is-green {
    background: #14c9c9;
  }

  .is-blue {
    background: #3b6cff;
  }

  .is-orange {
    background: #f59a23;
  }

  .is-purple {
    background: #7c6bf2;
  }

  .money {
    font-weight: 600;
  }

  .link-line a {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    color: rgb(var(--primary-6));
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  em {
    font-style: normal;
  }

  em.normal {
    color: #00b42a;
  }

  em.disabled {
    color: #f53f3f;
  }

  @media (max-width: 1200px) {
    .detail-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }
</style>
