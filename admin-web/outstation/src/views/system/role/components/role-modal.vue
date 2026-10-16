<template>
  <a-modal
    :visible="visible"
    :title="isEdit ? $t('roleModal.title.edit') : $t('roleModal.title.create')"
    :ok-loading="submitting"
    :ok-text="isEdit ? $t('front.common.okText.save') : $t('front.common.okText.create')"
    :mask-closable="false"
    :esc-to-close="false"
    unmount-on-close
    :width="560"
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    @before-ok="handleBeforeOk"
    @cancel="handleCancel"
  >
    <a-spin :loading="detailLoading" style="width: 100%">
      <a-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-align="right"
        :label-col-props="{ span: 6 }"
        :wrapper-col-props="{ span: 17 }"
      >
        <a-form-item :label="$t('roleModal.field.roleName')" field="role_name" required>
          <a-input
            v-model="form.role_name"
            :placeholder="$t('roleModal.field.roleNamePlaceholder')"
            allow-clear
            :max-length="64"
          />
        </a-form-item>
        <a-form-item :label="$t('roleModal.field.roleCode')" field="role_code" required>
          <a-input
            v-model="form.role_code"
            :placeholder="$t('roleModal.field.roleCodePlaceholder')"
            allow-clear
            :max-length="64"
            :disabled="isEdit"
          />
        </a-form-item>
        <a-form-item :label="$t('roleModal.field.description')" field="description">
          <a-textarea
            v-model="form.description"
            :placeholder="$t('roleModal.field.descriptionPlaceholder')"
            :max-length="200"
            :auto-size="{ minRows: 2, maxRows: 4 }"
            allow-clear
          />
        </a-form-item>
        <a-form-item :label="$t('roleModal.field.sortNo')" field="sort_no">
          <a-input-number
            v-model="form.sort_no"
            :min="0"
            :max="9999"
            :placeholder="$t('roleModal.field.sortNoPlaceholder')"
            style="width: 100%"
          />
        </a-form-item>
        <a-form-item :label="$t('roleModal.field.status')" field="status">
          <a-radio-group v-model="form.status" :disabled="isSystem">
            <a-radio :value="1">{{ $t('roleModal.field.statusEnabled') }}</a-radio>
            <a-radio :value="2">{{ $t('roleModal.field.statusDisabled') }}</a-radio>
          </a-radio-group>
        </a-form-item>
      </a-form>
    </a-spin>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { FormInstance, FieldRule } from '@arco-design/web-vue';
  import {
    createRole,
    queryRoleDetail,
    updateRole,
    type CreateRoleParams,
    type RoleRecord,
    type UpdateRoleParams,
  } from '@/api/role';

  const { t } = useI18n();

  const props = defineProps<{
    visible: boolean;
    role?: RoleRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  const formRef = ref<FormInstance>();
  const submitting = ref(false);
  const detailLoading = ref(false);
  const isSystem = ref(false);

  const isEdit = computed(() => Boolean(props.role?.id));

  const getDefaultForm = (): CreateRoleParams => ({
    role_name: '',
    role_code: '',
    description: '',
    status: 1,
    sort_no: 0,
  });

  const form = reactive<CreateRoleParams>(getDefaultForm());

  const rules = computed<Record<string, FieldRule | FieldRule[]>>(() => ({
    role_name: [{ required: true, message: t('roleModal.rule.roleName') }],
    role_code: isEdit.value
      ? []
      : [
          { required: true, message: t('roleModal.rule.roleCode') },
          {
            match: /^[a-zA-Z][a-zA-Z0-9_]*$/,
            message: t('roleModal.rule.roleCodeFormat'),
          },
        ],
  }));

  const roleName = (item: RoleRecord) =>
    item.role_name || item.name || item.display_name || '';

  const roleCode = (item: RoleRecord) => item.role_code || item.code || '';

  const resetForm = () => {
    Object.assign(form, getDefaultForm());
    isSystem.value = false;
    formRef.value?.clearValidate();
  };

  const fillForm = (role: RoleRecord) => {
    Object.assign(form, getDefaultForm(), {
      role_name: roleName(role),
      role_code: roleCode(role),
      description: role.description || '',
      status: role.status ?? 1,
      sort_no: role.sort_no ?? 0,
    });
    isSystem.value = Boolean(role.is_system);
  };

  const fillFormFromDetail = async (id: number) => {
    detailLoading.value = true;
    try {
      const { data } = await queryRoleDetail(id);
      if (data) fillForm(data);
      else if (props.role) fillForm(props.role);
    } catch {
      if (props.role) fillForm(props.role);
    } finally {
      detailLoading.value = false;
    }
  };

  const close = () => {
    emit('update:visible', false);
  };

  const handleCancel = () => {
    close();
    resetForm();
  };

  const handleBeforeOk = async (done: (closed: boolean) => void) => {
    const errors = await formRef.value?.validate();
    if (errors || submitting.value) {
      done(false);
      return;
    }

    submitting.value = true;
    try {
      if (isEdit.value && props.role?.id) {
        const payload: UpdateRoleParams = {
          id: props.role.id,
          role_name: form.role_name.trim(),
          description: form.description?.trim() || '',
          status: form.status ?? 1,
          sort_no: form.sort_no ?? 0,
        };
        await updateRole(payload);
        Message.success(t('front.common.message.saveSuccess'));
      } else {
        const payload: CreateRoleParams = {
          role_name: form.role_name.trim(),
          role_code: form.role_code.trim(),
          description: form.description?.trim() || '',
          status: form.status ?? 1,
          sort_no: form.sort_no ?? 0,
        };
        await createRole(payload);
        Message.success(t('front.common.message.createSuccess'));
      }
      resetForm();
      emit('success');
      done(true);
      close();
    } catch {
      done(false);
    } finally {
      submitting.value = false;
    }
  };

  watch(
    () => props.visible,
    async (visible) => {
      if (!visible) return;
      resetForm();
      if (isEdit.value && props.role?.id) {
        await fillFormFromDetail(props.role.id);
      }
    }
  );
</script>
