<template>
  <div class="locales">
    <a-radio-group
      v-if="items.length > 1"
      v-model="active"
      type="button"
      size="small"
      class="lang-switch"
      @click.stop
    >
      <a-radio v-for="item in items" :key="item.lang" :value="item.lang">
        {{ langLabel(item.lang) }}
      </a-radio>
    </a-radio-group>
    <div v-else class="lang-single">{{ langLabel(current.lang) }}</div>

    <dl class="locale-body">
      <dt>{{ t('discountCenter.activityCenter.columns.activityName') }}</dt>
      <dd class="locale-name">{{ displayText(current.name) }}</dd>
      <dt>{{ t('discountCenter.activityCenter.detail.description') }}</dt>
      <dd class="locale-description">
        {{ displayText(current.description) }}
      </dd>
      <dt>{{ t('discountCenter.activityCenter.expand.rule') }}</dt>
      <!-- eslint-disable-next-line vue/no-v-html -- 已经过 sanitizeRichText 白名单过滤 -->
      <dd v-if="ruleHtml" class="locale-rule is-rich" v-html="ruleHtml"></dd>
      <dd v-else-if="current.rule" class="locale-rule">{{ current.rule }}</dd>
      <dd v-else class="locale-rule muted">
        {{ t('discountCenter.activityCenter.expand.noRule') }}
      </dd>
    </dl>
  </div>
</template>

<script lang="ts" setup>
  import { computed, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type { ActivityI18nItem } from '@/api/activity';
  import { displayText } from '../options';
  import { isRichText, sanitizeRichText } from '../rich-text';
  import { useActivityStore } from '../store';

  const props = defineProps<{
    /** 至少一项 */
    items: ActivityI18nItem[];
  }>();

  const { t, locale } = useI18n();
  const { langLabel, loadLangs } = useActivityStore();

  const pickDefault = () =>
    props.items.find((item) => item.lang === locale.value)?.lang ??
    props.items[0]?.lang ??
    '';

  const active = ref(pickDefault());

  watch(
    () => props.items.map((item) => item.lang).join(),
    () => {
      if (!props.items.some((item) => item.lang === active.value)) {
        active.value = pickDefault();
      }
    }
  );

  const current = computed(
    () =>
      props.items.find((item) => item.lang === active.value) ?? props.items[0]
  );

  /** 新增弹窗保存的是 Quill HTML，后端初始化数据是纯文本，两种都要兼容 */
  const ruleHtml = computed(() => {
    const rule = current.value?.rule || '';
    return isRichText(rule) ? sanitizeRichText(rule) : '';
  });

  loadLangs();
</script>

<style scoped lang="less">
  .locales {
    display: flex;
    flex-direction: column;
    gap: 10px;
    min-width: 0;
  }

  .lang-switch {
    align-self: flex-start;
  }

  .lang-single {
    color: var(--color-text-3);
    font-size: 12px;
  }

  .locale-body {
    display: grid;
    grid-template-columns: max-content minmax(0, 1fr);
    gap: 8px 16px;
    margin: 0;
    padding: 12px 14px;
    background: var(--color-fill-1);
    border-radius: 6px;

    dt {
      color: var(--color-text-3);
      font-size: 12px;
      line-height: 22px;
    }

    dd {
      margin: 0;
    }
  }

  .locale-name {
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 14px;
    line-height: 22px;
  }

  .locale-description {
    color: var(--color-text-2);
    font-size: 13px;
    line-height: 22px;
    white-space: pre-wrap;
    word-break: break-word;
  }

  .locale-rule {
    color: var(--color-text-1);
    font-size: 13px;
    line-height: 22px;
    white-space: pre-wrap;
    word-break: break-word;
  }

  .locale-rule.is-rich {
    white-space: normal;

    :deep(p),
    :deep(ol),
    :deep(ul),
    :deep(blockquote),
    :deep(pre) {
      margin: 0 0 4px;
    }

    :deep(ol),
    :deep(ul) {
      padding-left: 20px;
    }
  }

  .muted {
    color: var(--color-text-3);
  }
</style>
