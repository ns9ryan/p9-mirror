<template>
  <a-modal
    :visible="visible"
    :width="560"
    :footer="false"
    unmount-on-close
    title-align="start"
    :title="t('memberManage.ops.restriction.title')"
    @cancel="close"
  >
    <a-spin :loading="loading" class="ops-spin">
      <a-result
        v-if="!loading && !flags"
        status="error"
        :subtitle="t('memberManage.ops.restriction.loadFailed')"
      >
        <template #extra>
          <a-button type="primary" @click="load">
            {{ t('memberManage.ops.retry') }}
          </a-button>
        </template>
      </a-result>

      <template v-else>
        <section class="ops-section">
          <h4>{{ t('memberManage.ops.restriction.sectionBan') }}</h4>
          <div v-for="item in RESTRICTIONS" :key="item.kind" class="ops-row">
            <span>{{ t(item.label) }}</span>
            <a-switch
              :model-value="flags?.[item.kind] ?? false"
              :disabled="!flags"
              :before-change="(value) => toggle(item.kind, value === true)"
            />
          </div>
        </section>

        <section class="ops-section">
          <h4>{{ t('memberManage.ops.restriction.sectionReset') }}</h4>
          <p class="ops-tip">{{
            t('memberManage.ops.restriction.resetTip')
          }}</p>
          <div v-for="item in AUTH_ERRORS" :key="item.field" class="ops-row">
            <span>{{ t(item.label) }}</span>
            <a-popconfirm
              :content="
                t('memberManage.ops.restriction.resetConfirm', {
                  name: t(item.label),
                })
              "
              :on-before-ok="() => reset(item.field)"
            >
              <a-button size="small">
                {{ t('memberManage.ops.restriction.reset') }}
              </a-button>
            </a-popconfirm>
          </div>
        </section>
      </template>
    </a-spin>
  </a-modal>
</template>

<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import {
    getMemberDetail,
    type MemberAuthErrorField,
    type MemberRestriction,
  } from '@/api/member';
  import { useMemberStore } from '../../store';
  import type { MemberRecord } from '../../types';

  const RESTRICTIONS: { kind: MemberRestriction; label: string }[] = [
    { kind: 'chat', label: 'memberManage.restriction.chat' },
    { kind: 'betting', label: 'memberManage.restriction.bet' },
    { kind: 'withdrawal', label: 'memberManage.restriction.withdraw' },
  ];

  const AUTH_ERRORS: { field: MemberAuthErrorField; label: string }[] = [
    {
      field: 'sms_code_error_count',
      label: 'memberManage.ops.restriction.smsCode',
    },
    {
      field: 'email_code_error_count',
      label: 'memberManage.ops.restriction.emailCode',
    },
    {
      field: 'login_password_error_count',
      label: 'memberManage.ops.restriction.loginPassword',
    },
    {
      field: 'withdrawal_password_error_count',
      label: 'memberManage.ops.restriction.withdrawalPassword',
    },
  ];

  const props = defineProps<{
    visible: boolean;
    record: MemberRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
  }>();

  const { t } = useI18n();
  const { setRestriction, resetAuthErrors } = useMemberStore();

  /** 当前开关状态取自详情接口，列表只有投注开关 */
  const flags = ref<Record<MemberRestriction, boolean> | null>(null);
  const loading = ref(false);
  let ticket = 0;

  async function load() {
    const id = props.record?.id;
    if (id == null) return;
    ticket += 1;
    const current = ticket;
    loading.value = true;
    flags.value = null;
    try {
      const { data } = await getMemberDetail(id);
      if (current !== ticket || !data) return;
      flags.value = {
        chat: !!data.chat_disabled,
        betting: !!data.betting_disabled,
        withdrawal: !!data.withdrawal_disabled,
      };
    } catch {
      // 错误提示由请求拦截器统一处理
    } finally {
      if (current === ticket) loading.value = false;
    }
  }

  watch(
    () => props.visible,
    (open) => {
      if (open) load();
    }
  );

  async function toggle(kind: MemberRestriction, disabled: boolean) {
    const id = props.record?.id;
    if (id == null || !flags.value) return false;
    try {
      await setRestriction(id, kind, disabled);
      flags.value = { ...flags.value, [kind]: disabled };
      Message.success(
        t(
          disabled
            ? 'memberManage.ops.restriction.banned'
            : 'memberManage.ops.restriction.unbanned'
        )
      );
      return true;
    } catch {
      return false;
    }
  }

  async function reset(field: MemberAuthErrorField) {
    const id = props.record?.id;
    if (id == null) return false;
    try {
      await resetAuthErrors(id, field);
      Message.success(t('memberManage.ops.restriction.resetSuccess'));
      return true;
    } catch {
      return false;
    }
  }

  function close() {
    emit('update:visible', false);
  }
</script>

<style scoped lang="less">
  .ops-spin {
    display: block;
    min-height: 160px;
  }

  .ops-section {
    & + & {
      margin-top: 20px;
    }

    h4 {
      margin: 0 0 8px;
      color: var(--color-text-1);
      font-weight: 600;
      font-size: 14px;
    }
  }

  .ops-tip {
    margin: 0 0 8px;
    color: var(--color-text-3);
    font-size: 12px;
  }

  .ops-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 0;
    color: var(--color-text-1);
    font-size: 13px;
    border-bottom: 1px solid var(--color-neutral-2);

    &:last-child {
      border-bottom: none;
    }
  }
</style>
