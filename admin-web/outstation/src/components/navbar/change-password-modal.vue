<template>
  <a-modal
    :visible="visible"
    title="修改密码"
    :ok-loading="submitting"
    unmount-on-close
    @before-ok="handleBeforeOk"
    @cancel="handleCancel"
  >
    <a-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-align="right"
      :label-col-props="{ span: 6 }"
      :wrapper-col-props="{ span: 17 }"
    >
      <a-form-item label="旧密码" field="old_password" required>
        <a-input-password
          v-model="form.old_password"
          placeholder="请输入密码"
          allow-clear
          :max-length="64"
        />
      </a-form-item>
      <a-form-item label="新密码" field="password" required>
        <a-input-password
          v-model="form.password"
          placeholder="请输入密码"
          allow-clear
          :max-length="64"
        />
      </a-form-item>
      <a-form-item label="再次输入" field="confirm_password" required>
        <a-input-password
          v-model="form.confirm_password"
          placeholder="请输入密码"
          allow-clear
          :max-length="64"
        />
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { reactive, ref, watch } from 'vue';
  import { useRouter } from 'vue-router';
  import { Message } from '@arco-design/web-vue';
  import type { FormInstance } from '@arco-design/web-vue';
  import { changeSelfPassword } from '@/api/user';
  import { useUserStore } from '@/store';

  const props = defineProps<{
    visible: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
  }>();

  const router = useRouter();
  const userStore = useUserStore();
  const formRef = ref<FormInstance>();
  const submitting = ref(false);

  const form = reactive({
    old_password: '',
    password: '',
    confirm_password: '',
  });

  const resetForm = () => {
    form.old_password = '';
    form.password = '';
    form.confirm_password = '';
    formRef.value?.clearValidate();
  };

  watch(
    () => props.visible,
    (visible) => {
      if (visible) resetForm();
    }
  );

  const rules = {
    old_password: [{ required: true, message: '请输入旧密码' }],
    password: [
      { required: true, message: '请输入新密码' },
      {
        validator: (value: string, callback: (error?: string) => void) => {
          if (value && value === form.old_password) {
            callback('新密码不能与旧密码相同');
            return;
          }
          callback();
        },
      },
    ],
    confirm_password: [
      { required: true, message: '请再次输入新密码' },
      {
        validator: (value: string, callback: (error?: string) => void) => {
          if (value && value !== form.password) {
            callback('两次输入的密码不一致');
            return;
          }
          callback();
        },
      },
    ],
  };

  const handleCancel = () => {
    emit('update:visible', false);
  };

  const handleBeforeOk = async (done: (closed: boolean) => void) => {
    submitting.value = true;
    try {
      const errors = await formRef.value?.validate();
      if (errors) {
        done(false);
        return;
      }
      await changeSelfPassword({
        old_password: form.old_password,
        password: form.password,
      });
      Message.success('密码修改成功，请重新登录');
      emit('update:visible', false);
      done(true);
      userStore.logoutCallBack();
      router.push({ name: 'login' });
    } catch {
      done(false);
    } finally {
      submitting.value = false;
    }
  };
</script>
