<template>
  <a-modal
    :visible="visible"
    :title="$t('ipManageModal.title')"
    :ok-loading="submitting"
    :ok-text="$t('ipManageModal.okText')"
    :cancel-text="$t('ipManageModal.cancelText')"
    unmount-on-close
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    @before-ok="handleBeforeOk"
    @cancel="handleCancel"
  >
    <a-spin :loading="detailLoading" style="width: 100%">
      <a-form
        :model="form"
        label-align="right"
        :label-col-props="{ span: 5 }"
        :wrapper-col-props="{ span: 18 }"
      >
        <a-form-item :label="$t('ipManageModal.field.switch')">
          <a-switch
            v-model="form.enabled"
            :checked-text="$t('front.common.switch.on')"
            :unchecked-text="$t('front.common.switch.off')"
          />
        </a-form-item>
        <a-form-item :label="$t('ipManageModal.field.whitelist')">
          <div class="ip-list">
            <div
              v-for="(item, index) in form.ipList"
              :key="item.key"
              class="ip-row"
            >
              <a-input
                v-model="item.value"
                :placeholder="$t('ipManageModal.field.placeholder')"
                allow-clear
              />
              <a-button
                shape="circle"
                status="danger"
                @click="removeIp(index)"
              >
                <template #icon><icon-minus /></template>
              </a-button>
            </div>
            <a-button type="primary" class="add-btn" @click="addIp">
              <template #icon><icon-plus /></template>
            </a-button>
          </div>
        </a-form-item>
      </a-form>
    </a-spin>
  </a-modal>
</template>

<script lang="ts" setup>
  import { reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { queryUserDetail, updateUserIpWhitelist } from '@/api/user';

  const props = defineProps<{
    visible: boolean;
    userId?: number | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  const { t } = useI18n();
  const submitting = ref(false);
  const detailLoading = ref(false);
  let rowSeq = 0;

  interface IpRow {
    key: number;
    value: string;
  }

  const form = reactive({
    enabled: false,
    ipList: [] as IpRow[],
  });

  const createRow = (value = ''): IpRow => {
    rowSeq += 1;
    return {
      key: rowSeq,
      value,
    };
  };

  const resetForm = () => {
    form.enabled = false;
    form.ipList = [];
  };

  const loadDetail = async (id: number) => {
    detailLoading.value = true;
    try {
      const { data } = await queryUserDetail(id);
      form.enabled = data?.ip_whitelist_enabled === 1;
      const list = (data?.ip_whitelist || []).filter(Boolean);
      form.ipList = list.map((ip) => createRow(ip));
    } catch {
      resetForm();
    } finally {
      detailLoading.value = false;
    }
  };

  watch(
    () => props.visible,
    (visible) => {
      if (!visible) return;
      resetForm();
      if (props.userId) loadDetail(props.userId);
    }
  );

  const addIp = () => {
    form.ipList.push(createRow());
  };

  const removeIp = (index: number) => {
    form.ipList.splice(index, 1);
  };

  const handleCancel = () => {
    emit('update:visible', false);
  };

  const handleBeforeOk = async (done: (closed: boolean) => void) => {
    if (!props.userId || submitting.value) {
      done(false);
      return;
    }

    const ipWhitelist = form.ipList
      .map((item) => item.value.trim())
      .filter(Boolean);

    submitting.value = true;
    try {
      await updateUserIpWhitelist({
        id: props.userId,
        ip_whitelist: ipWhitelist,
        ip_whitelist_enabled: form.enabled ? 1 : 0,
      });
      Message.success(t('ipManageModal.message.success'));
      emit('success');
      emit('update:visible', false);
      done(true);
    } catch {
      done(false);
    } finally {
      submitting.value = false;
    }
  };
</script>

<style scoped lang="less">
  .ip-list {
    width: 100%;
  }
  .ip-row {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;
  }
  .ip-row :deep(.arco-input-wrapper) {
    flex: 1;
  }
  .add-btn {
    width: 32px;
    padding: 0;
  }
</style>
