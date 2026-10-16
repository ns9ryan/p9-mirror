<template>
  <a-modal
    :visible="visible"
    :title="isEdit ? $t('userModal.title.edit') : $t('userModal.title.create')"
    :ok-loading="submitting"
    :ok-text="isEdit ? $t('front.common.okText.save') : $t('front.common.okText.create')"
    :mask-closable="false"
    :esc-to-close="false"
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
        ref="formRef"
        :model="form"
        :rules="rules"
        label-align="right"
        :label-col-props="{ span: 6 }"
        :wrapper-col-props="{ span: 17 }"
      >
        <a-form-item :label="$t('userModal.field.username')" field="username" required>
          <a-input
            v-model="form.username"
            :placeholder="$t('userModal.field.usernamePlaceholder')"
            allow-clear
            :max-length="64"
            :disabled="isEdit"
          />
        </a-form-item>
        <a-form-item v-if="!isEdit" :label="$t('userModal.field.password')" field="password" required>
          <a-input-password
            v-model="form.password"
            :placeholder="$t('userModal.field.passwordPlaceholder')"
            allow-clear
            :max-length="64"
          />
        </a-form-item>
        <a-form-item :label="$t('userModal.field.displayName')" field="display_name" required>
          <a-input
            v-model="form.display_name"
            :placeholder="$t('userModal.field.displayNamePlaceholder')"
            allow-clear
            :max-length="64"
          />
        </a-form-item>
        <a-form-item :label="$t('userModal.field.mobile')" field="mobile">
          <a-input
            v-model="form.mobile"
            :placeholder="$t('userModal.field.mobilePlaceholder')"
            allow-clear
            :max-length="20"
          />
        </a-form-item>
        <a-form-item :label="$t('userModal.field.email')" field="email">
          <a-input
            v-model="form.email"
            :placeholder="$t('userModal.field.emailPlaceholder')"
            allow-clear
            :max-length="128"
          />
        </a-form-item>
        <a-form-item :label="$t('userModal.field.status')" field="status">
          <a-radio-group v-model="form.status">
            <a-radio :value="1">{{ $t('userModal.field.statusEnabled') }}</a-radio>
            <a-radio :value="2">{{ $t('userModal.field.statusDisabled') }}</a-radio>
          </a-radio-group>
        </a-form-item>
        <a-form-item :label="$t('userModal.field.roles')" field="role_ids" required>
          <a-select
            v-model="form.role_ids"
            :placeholder="$t('userModal.field.rolesPlaceholder')"
            multiple
            allow-clear
            :disabled="roleSelectDisabled"
            :loading="roleLoading"
            :options="roleOptions"
          />
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
    createUser,
    queryRoleList,
    queryUserDetail,
    updateUser,
    updateUserRoles,
    type CreateUserParams,
    type RoleRecord,
    type UpdateUserParams,
  } from '@/api/user';

  const { t } = useI18n();

  const props = defineProps<{
    visible: boolean;
    /** 传入用户 ID 时为编辑模式，打开弹窗后拉取详情回填 */
    userId?: number | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  const formRef = ref<FormInstance>();
  const submitting = ref(false);
  const detailLoading = ref(false);
  const roleLoading = ref(false);
  const roleLoaded = ref(false);
  const isSuperAdmin = ref(false);
  const roleOptions = ref<{ label: string; value: number; code?: string }[]>(
    []
  );
  const pendingRoleCodes = ref<string[]>([]);

  const isEdit = computed(() => props.userId != null && props.userId > 0);
  /** 超管不可改绑定角色；新建或普通用户可改 */
  const roleSelectDisabled = computed(
    () => isEdit.value && isSuperAdmin.value
  );

  const getDefaultForm = (): CreateUserParams => ({
    username: '',
    password: '',
    display_name: '',
    mobile: '',
    email: '',
    status: 1,
    role_ids: [],
  });

  const form = reactive<CreateUserParams>(getDefaultForm());

  const rules = computed<Record<string, FieldRule | FieldRule[]>>(() => ({
    username: [{ required: true, message: t('userModal.rule.username') }],
    password: isEdit.value
      ? []
      : [{ required: true, message: t('userModal.rule.password') }],
    display_name: [{ required: true, message: t('userModal.rule.displayName') }],
    email: [
      {
        validator: (value: string, callback: (error?: string) => void) => {
          if (!value) {
            callback();
            return;
          }
          if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) {
            callback(t('userModal.rule.emailInvalid'));
            return;
          }
          callback();
        },
      },
    ],
  }));

  const roleLabel = (item: RoleRecord) =>
    item.name ||
    item.role_name ||
    item.display_name ||
    item.role_code ||
    item.code ||
    t('userModal.roleFallback', { id: item.id });

  const applyRoleCodes = (codes: string[]) => {
    if (!codes.length || !roleOptions.value.length) {
      pendingRoleCodes.value = codes;
      return;
    }
    const codeSet = new Set(codes);
    form.role_ids = roleOptions.value
      .filter((opt) => opt.code && codeSet.has(opt.code))
      .map((opt) => opt.value);
    pendingRoleCodes.value = [];
  };

  const fetchRoles = async () => {
    if (roleLoaded.value || roleLoading.value) return;
    roleLoading.value = true;
    try {
      const { data } = await queryRoleList(
        { page: 1, page_size: 200 },
        { silentError: true }
      );
      roleOptions.value = (data?.list || []).map((item) => ({
        label: roleLabel(item),
        value: item.id,
        code: item.role_code || item.code || '',
      }));
      roleLoaded.value = true;
      if (pendingRoleCodes.value.length) {
        applyRoleCodes(pendingRoleCodes.value);
      }
    } catch {
      roleOptions.value = [];
    } finally {
      roleLoading.value = false;
    }
  };

  const resetForm = () => {
    pendingRoleCodes.value = [];
    isSuperAdmin.value = false;
    Object.assign(form, getDefaultForm());
    formRef.value?.clearValidate();
  };

  const fillFormFromDetail = async (id: number) => {
    detailLoading.value = true;
    try {
      const { data } = await queryUserDetail(id);
      isSuperAdmin.value = !!data?.is_super_admin;
      Object.assign(form, getDefaultForm(), {
        username: data?.username || '',
        display_name: data?.display_name || '',
        mobile: data?.mobile || '',
        email: data?.email || '',
        status: data?.status ?? 1,
        role_ids: [],
      });
      applyRoleCodes(data?.role_codes || []);
    } catch {
      resetForm();
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
      if (isEdit.value && props.userId) {
        const payload: UpdateUserParams = {
          id: props.userId,
          display_name: form.display_name.trim(),
          mobile: form.mobile?.trim() || '',
          email: form.email?.trim() || '',
          status: form.status ?? 1,
        };
        await updateUser(payload);
        if (!isSuperAdmin.value) {
          await updateUserRoles({
            user_id: props.userId,
            role_ids: form.role_ids || [],
          });
        }
        Message.success(t('front.common.message.saveSuccess'));
      } else {
        const payload: CreateUserParams = {
          username: form.username.trim(),
          password: form.password,
          display_name: form.display_name.trim(),
          status: form.status ?? 1,
        };
        if (form.mobile?.trim()) payload.mobile = form.mobile.trim();
        if (form.email?.trim()) payload.email = form.email.trim();
        if (form.role_ids?.length) payload.role_ids = form.role_ids;

        const { data } = await createUser(payload);
        Message.success(
          t('front.common.message.createSuccess', {
            username: data?.username || payload.username,
          })
        );
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
      await fetchRoles();
      if (isEdit.value && props.userId) {
        await fillFormFromDetail(props.userId);
      }
    }
  );
</script>
