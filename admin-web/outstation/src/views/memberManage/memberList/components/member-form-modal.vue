<template>
  <a-modal
    :visible="visible"
    :width="800"
    :mask-closable="false"
    unmount-on-close
    title-align="start"
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel member-edit-modal"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    :ok-text="t('front.common.okText.save')"
    :ok-button-props="{ disabled: !original }"
    @before-ok="handleBeforeOk"
    @cancel="close"
  >
    <template #title>
      <div class="edit-title">
        <div class="edit-title-main">{{ t('memberManage.modal.edit') }}</div>
        <div class="edit-title-sub">
          {{ t('memberManage.modal.editSubtitle') }}
        </div>
      </div>
    </template>

    <a-spin :loading="loading" class="edit-spin">
      <a-result
        v-if="!loading && !original"
        status="error"
        :subtitle="t('memberManage.contact.loadFailed')"
      >
        <template #extra>
          <a-button type="primary" @click="load">
            {{ t('memberManage.contact.retry') }}
          </a-button>
        </template>
      </a-result>

      <a-form
        v-else
        ref="formRef"
        :model="form"
        :rules="rules"
        layout="vertical"
        class="edit-form"
      >
        <section class="edit-section">
          <div class="section-head">
            <span class="section-index">1</span>
            {{ t('memberManage.contact.sectionProfile') }}
          </div>
          <div class="profile-row">
            <div class="avatar-preview">
              <img
                v-if="avatarUrl"
                class="avatar-preview-img"
                :src="avatarUrl"
                alt=""
              />
              <span v-else class="avatar-preview-img is-empty">
                {{ initial }}
              </span>
              <img
                v-if="frameUrl"
                class="avatar-preview-frame"
                :src="frameUrl"
                alt=""
              />
            </div>
            <a-form-item
              field="nickname"
              :label="t('memberManage.contact.nickname')"
              class="nickname-item"
            >
              <a-input
                v-model="form.nickname"
                :max-length="100"
                :placeholder="t('memberManage.contact.nicknamePlaceholder')"
              />
            </a-form-item>
          </div>

          <a-form-item :label="t('memberManage.contact.avatar')">
            <div v-if="images.avatars.length" class="image-grid">
              <button
                v-for="item in images.avatars"
                :key="item.id"
                type="button"
                class="image-tile"
                :class="{ 'is-active': form.avatarId === item.id }"
                @click="form.avatarId = item.id"
              >
                <img :src="item.image_url" alt="" />
              </button>
            </div>
            <span v-else class="image-empty">
              {{ t('memberManage.contact.noAvatar') }}
            </span>
          </a-form-item>

          <a-form-item :label="t('memberManage.contact.avatarFrame')">
            <div v-if="images.avatar_frames.length" class="image-grid">
              <button
                v-for="item in images.avatar_frames"
                :key="item.id"
                type="button"
                class="image-tile"
                :class="{ 'is-active': form.avatarFrameId === item.id }"
                @click="form.avatarFrameId = item.id"
              >
                <img :src="item.image_url" alt="" />
              </button>
            </div>
            <span v-else class="image-empty">
              {{ t('memberManage.contact.noAvatarFrame') }}
            </span>
          </a-form-item>
        </section>

        <section class="edit-section">
          <div class="section-head">
            <span class="section-index">2</span>
            {{ t('memberManage.contact.sectionIdentity') }}
            <span class="section-tip">
              {{ t('memberManage.contact.keepTip') }}
            </span>
          </div>
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item :label="t('memberManage.contact.realName')">
                <a-input
                  v-model="form.realName"
                  :max-length="100"
                  :disabled="clear.realName"
                  :placeholder="t('memberManage.contact.keepPlaceholder')"
                  allow-clear
                />
                <template #extra>
                  <a-checkbox v-model="clear.realName">
                    {{ t('memberManage.contact.clear') }}
                  </a-checkbox>
                </template>
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item :label="t('memberManage.contact.identityNumber')">
                <a-input
                  v-model="form.identityNumber"
                  :max-length="100"
                  :disabled="clear.identityNumber"
                  :placeholder="t('memberManage.contact.keepPlaceholder')"
                  allow-clear
                />
                <template #extra>
                  <a-checkbox v-model="clear.identityNumber">
                    {{ t('memberManage.contact.clear') }}
                  </a-checkbox>
                </template>
              </a-form-item>
            </a-col>
          </a-row>
        </section>

        <section class="edit-section">
          <div class="section-head">
            <span class="section-index">3</span>
            {{ t('memberManage.contact.sectionContact') }}
            <span class="section-tip">
              {{ t('memberManage.contact.verifyTip') }}
            </span>
          </div>
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item
                field="phone"
                :label="t('memberManage.contact.phone')"
              >
                <a-input
                  v-model="form.phone"
                  :max-length="25"
                  :placeholder="t('memberManage.contact.phonePlaceholder')"
                  allow-clear
                />
              </a-form-item>
              <a-form-item
                v-if="form.phone.trim()"
                field="phoneCode"
                :label="t('memberManage.contact.verificationCode')"
              >
                <div class="verify-row">
                  <a-input
                    v-model="verify.phone.code"
                    :max-length="6"
                    :placeholder="t('memberManage.contact.verificationPlaceholder')"
                    allow-clear
                  />
                  <a-button
                    :loading="verify.phone.sending"
                    :disabled="verify.phone.cooldown > 0"
                    @click="sendCode('phone')"
                  >
                    {{ sendButtonText('phone') }}
                  </a-button>
                </div>
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item
                field="email"
                :label="t('memberManage.contact.email')"
              >
                <a-input
                  v-model="form.email"
                  :max-length="254"
                  :placeholder="t('memberManage.contact.keepPlaceholder')"
                  allow-clear
                />
              </a-form-item>
              <a-form-item
                v-if="form.email.trim()"
                field="emailCode"
                :label="t('memberManage.contact.verificationCode')"
              >
                <div class="verify-row">
                  <a-input
                    v-model="verify.email.code"
                    :max-length="6"
                    :placeholder="t('memberManage.contact.verificationPlaceholder')"
                    allow-clear
                  />
                  <a-button
                    :loading="verify.email.sending"
                    :disabled="verify.email.cooldown > 0"
                    @click="sendCode('email')"
                  >
                    {{ sendButtonText('email') }}
                  </a-button>
                </div>
              </a-form-item>
            </a-col>
          </a-row>
        </section>
      </a-form>
    </a-spin>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { FieldRule, FormInstance } from '@arco-design/web-vue';
  import {
    CONTACT_CHANNEL,
    getMemberDetail,
    sendMemberContactBindingCode,
    type ContactVerification,
    type UpdateMemberContactParams,
  } from '@/api/member';
  import { useMemberStore } from '../store';
  import type { MemberRecord } from '../types';

  const props = defineProps<{
    visible: boolean;
    record: MemberRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'saved'): void;
  }>();

  const { t } = useI18n();
  const { images, loadImages, updateContact } = useMemberStore();

  const PHONE_PATTERN = /^\d{5,20}$/;
  const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  const CODE_PATTERN = /^\d{6}$/;
  /** 发送验证码后的冷却秒数，与服务端 60 秒限制一致 */
  const COOLDOWN_SECONDS = 60;

  type ClearField = 'realName' | 'identityNumber';
  type ContactChannel = 'phone' | 'email';

  function createForm() {
    return {
      avatarId: 0,
      avatarFrameId: 0,
      nickname: '',
      realName: '',
      identityNumber: '',
      phone: '',
      email: '',
    };
  }

  function createClear(): Record<ClearField, boolean> {
    return {
      realName: false,
      identityNumber: false,
    };
  }

  interface VerifyState {
    /** 0 表示没有可用记录 */
    recordId: number;
    /** 发送时使用的规范化目标，目标变化后记录失效 */
    target: string;
    code: string;
    sending: boolean;
    /** 剩余冷却秒数 */
    cooldown: number;
  }

  function createVerify(): VerifyState {
    return { recordId: 0, target: '', code: '', sending: false, cooldown: 0 };
  }

  const formRef = ref<FormInstance>();
  const form = reactive(createForm());
  const clear = reactive(createClear());
  const verify = reactive<Record<ContactChannel, VerifyState>>({
    phone: createVerify(),
    email: createVerify(),
  });
  const loading = ref(false);
  /** 打开时从详情接口取得的当前值，用于只提交改动的字段 */
  const original = ref<{
    nickname: string;
    avatarId: number;
    avatarFrameId: number;
  } | null>(null);
  let ticket = 0;

  /** 服务端同样会去空格及前导 + */
  function normalizePhone(value: string) {
    return value.replace(/\s+/g, '').replace(/^\+/, '');
  }

  function targetOf(channel: ContactChannel) {
    return channel === 'phone'
      ? normalizePhone(form.phone)
      : form.email.trim().toLowerCase();
  }

  const avatarUrl = computed(
    () =>
      images.value.avatars.find((item) => item.id === form.avatarId)?.image_url
  );
  const frameUrl = computed(
    () =>
      images.value.avatar_frames.find((item) => item.id === form.avatarFrameId)
        ?.image_url
  );
  const initial = computed(() =>
    (form.nickname || props.record?.account || 'U').charAt(0).toUpperCase()
  );

  const rules = computed<Record<string, FieldRule[]>>(() => ({
    nickname: [
      {
        validator: (value: string, callback: (error?: string) => void) => {
          if (!(value || '').trim()) {
            callback(t('memberManage.contact.nicknameRequired'));
            return;
          }
          callback();
        },
      },
    ],
    phone: [
      {
        validator: (value: string, callback: (error?: string) => void) => {
          const phone = normalizePhone(value || '');
          if (phone && !PHONE_PATTERN.test(phone)) {
            callback(t('memberManage.contact.phoneInvalid'));
            return;
          }
          callback();
        },
      },
    ],
    email: [
      {
        validator: (value: string, callback: (error?: string) => void) => {
          const email = (value || '').trim();
          if (email && !EMAIL_PATTERN.test(email)) {
            callback(t('memberManage.contact.emailInvalid'));
            return;
          }
          callback();
        },
      },
    ],
  }));

  async function load() {
    const id = props.record?.id;
    if (id == null) return;
    ticket += 1;
    const current = ticket;
    loading.value = true;
    original.value = null;
    try {
      const [{ data }] = await Promise.all([
        getMemberDetail(id),
        loadImages(true),
      ]);
      if (current !== ticket) return;
      const value = {
        nickname: data?.nickname || '',
        avatarId: data?.avatar_id || 0,
        avatarFrameId: data?.avatar_frame_id || 0,
      };
      Object.assign(form, value);
      original.value = value;
    } catch {
      // 错误提示由请求拦截器统一处理
    } finally {
      if (current === ticket) loading.value = false;
    }
  }

  /* ---------- 验证码冷却计时 ---------- */

  let cooldownTimer: ReturnType<typeof setInterval> | null = null;

  function tickCooldown() {
    let running = false;
    (['phone', 'email'] as const).forEach((channel) => {
      const state = verify[channel];
      if (state.cooldown <= 0) return;
      state.cooldown -= 1;
      if (state.cooldown > 0) running = true;
    });
    if (!running && cooldownTimer) {
      clearInterval(cooldownTimer);
      cooldownTimer = null;
    }
  }

  function startCooldown(channel: ContactChannel) {
    verify[channel].cooldown = COOLDOWN_SECONDS;
    if (!cooldownTimer) cooldownTimer = setInterval(tickCooldown, 1000);
  }

  onBeforeUnmount(() => {
    if (cooldownTimer) clearInterval(cooldownTimer);
    cooldownTimer = null;
  });

  function sendButtonText(channel: ContactChannel) {
    const { cooldown } = verify[channel];
    return cooldown > 0
      ? t('memberManage.contact.resendAfter', { seconds: cooldown })
      : t('memberManage.contact.sendCode');
  }

  /** 目标变化后旧验证码记录不再适用，冷却时间保留以免绕过频率限制 */
  function resetRecord(channel: ContactChannel) {
    const state = verify[channel];
    state.recordId = 0;
    state.target = '';
    state.code = '';
  }

  watch(
    () => form.phone,
    () => resetRecord('phone')
  );
  watch(
    () => form.email,
    () => resetRecord('email')
  );

  async function sendCode(channel: ContactChannel) {
    const memberId = props.record?.id;
    if (memberId == null) return;
    const target = targetOf(channel);
    if (!target) {
      Message.warning(t('memberManage.contact.targetRequired'));
      return;
    }
    const valid =
      channel === 'phone'
        ? PHONE_PATTERN.test(target)
        : EMAIL_PATTERN.test(target);
    if (!valid) {
      Message.warning(
        t(
          channel === 'phone'
            ? 'memberManage.contact.phoneInvalid'
            : 'memberManage.contact.emailInvalid'
        )
      );
      return;
    }

    const state = verify[channel];
    state.sending = true;
    try {
      const { data } = await sendMemberContactBindingCode({
        member_id: memberId,
        channel:
          channel === 'phone' ? CONTACT_CHANNEL.SMS : CONTACT_CHANNEL.EMAIL,
        target,
      });
      if (!data?.record_id) return;
      state.recordId = data.record_id;
      state.target = target;
      state.code = '';
      Message.success(t('memberManage.contact.codeSent'));
      startCooldown(channel);
    } catch {
      // 错误提示由请求拦截器统一处理
    } finally {
      state.sending = false;
    }
  }

  /** 取得本次修改可用的验证信息，不可用时提示并返回 null */
  function checkVerification(channel: ContactChannel): ContactVerification | null {
    const state = verify[channel];
    if (!state.recordId || state.target !== targetOf(channel)) {
      Message.warning(t('memberManage.contact.needSendCode'));
      return null;
    }
    const code = state.code.trim();
    if (!CODE_PATTERN.test(code)) {
      Message.warning(t('memberManage.contact.verificationInvalid'));
      return null;
    }
    return { record_id: state.recordId, verification_code: code };
  }

  watch(
    () => [props.visible, props.record?.id] as const,
    ([open]) => {
      if (!open) return;
      Object.assign(form, createForm());
      Object.assign(clear, createClear());
      Object.assign(verify.phone, createVerify());
      Object.assign(verify.email, createVerify());
      load();
    }
  );

  (Object.keys(clear) as ClearField[]).forEach((key) => {
    watch(
      () => clear[key],
      (checked) => {
        if (!checked) return;
        form[key] = '';
        formRef.value?.clearValidate(key);
      }
    );
  });

  /** 勾选清空时传空字符串，留空不传以保持原值 */
  function sensitiveValue(key: ClearField, value: string) {
    if (clear[key]) return '';
    return value || undefined;
  }

  function buildParams(memberId: number): UpdateMemberContactParams | null {
    const base = original.value;
    if (!base) return null;
    const params: UpdateMemberContactParams = { member_id: memberId };
    const nickname = form.nickname.trim();
    if (nickname !== base.nickname) params.nickname = nickname;
    if (form.avatarId && form.avatarId !== base.avatarId) {
      params.avatar_id = form.avatarId;
    }
    if (form.avatarFrameId && form.avatarFrameId !== base.avatarFrameId) {
      params.avatar_frame_id = form.avatarFrameId;
    }
    const optional: Partial<UpdateMemberContactParams> = {
      real_name: sensitiveValue('realName', form.realName.trim()),
      identity_number: sensitiveValue(
        'identityNumber',
        form.identityNumber.trim()
      ),
    };
    Object.entries(optional).forEach(([key, value]) => {
      if (value !== undefined) Object.assign(params, { [key]: value });
    });
    // 文档：修改手机号、邮箱不能清空，留空即不修改
    const phone = normalizePhone(form.phone);
    if (phone) params.phone = phone;
    const email = form.email.trim().toLowerCase();
    if (email) params.email = email;
    return params;
  }

  function close() {
    emit('update:visible', false);
  }

  async function handleBeforeOk() {
    const id = props.record?.id;
    if (id == null) return false;
    const errors = await formRef.value?.validate();
    if (errors) return false;
    const params = buildParams(id);
    if (!params) return false;

    // 修改手机号或邮箱必须携带对应的验证码
    if (params.phone) {
      const verification = checkVerification('phone');
      if (!verification) return false;
      params.phone_verification = verification;
    }
    if (params.email) {
      const verification = checkVerification('email');
      if (!verification) return false;
      params.email_verification = verification;
    }

    if (Object.keys(params).length === 1) {
      Message.info(t('memberManage.contact.noChange'));
      return false;
    }
    try {
      await updateContact(params);
      emit('saved');
      close();
      return true;
    } catch {
      return false;
    }
  }
</script>

<style scoped lang="less">
  .edit-title-main {
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 18px;
    line-height: 26px;
  }

  .edit-title-sub {
    margin-top: 2px;
    color: var(--color-text-3);
    font-weight: 400;
    font-size: 13px;
    line-height: 20px;
  }

  .edit-spin {
    display: block;
    min-height: 200px;
  }

  .edit-section {
    padding: 18px 20px 4px;
    background: var(--color-bg-2);
    border: 1px solid var(--color-neutral-2);
    border-radius: 10px;

    & + & {
      margin-top: 12px;
    }
  }

  .section-head {
    display: flex;
    gap: 10px;
    align-items: center;
    margin-bottom: 14px;
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 15px;
  }

  .section-tip {
    color: var(--color-text-3);
    font-weight: 400;
    font-size: 12px;
  }

  .section-index {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 22px;
    color: #fff;
    font-weight: 600;
    font-size: 13px;
    background: rgb(var(--primary-6));
    border-radius: 50%;
  }

  .profile-row {
    display: flex;
    gap: 20px;
    align-items: flex-start;
  }

  .avatar-preview {
    position: relative;
    flex: none;
    width: 72px;
    height: 72px;
  }

  .avatar-preview-img {
    position: absolute;
    inset: 8px;
    width: 56px;
    height: 56px;
    object-fit: cover;
    border-radius: 50%;

    &.is-empty {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      color: rgb(var(--primary-6));
      font-weight: 600;
      font-size: 22px;
      background: rgb(var(--primary-1));
    }
  }

  .avatar-preview-frame {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: contain;
    pointer-events: none;
  }

  .nickname-item {
    flex: 1;
    min-width: 0;
  }

  .verify-row {
    display: flex;
    gap: 8px;
    width: 100%;

    > :first-child {
      flex: 1;
      min-width: 0;
    }
  }

  .image-grid {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
  }

  .image-tile {
    width: 56px;
    height: 56px;
    padding: 3px;
    background: var(--color-fill-1);
    border: 2px solid transparent;
    border-radius: 8px;
    cursor: pointer;
    transition: border-color 0.15s;

    &:hover {
      border-color: rgb(var(--primary-3));
    }

    &.is-active {
      border-color: rgb(var(--primary-6));
    }

    img {
      width: 100%;
      height: 100%;
      object-fit: contain;
    }
  }

  .image-empty {
    color: var(--color-text-3);
    font-size: 13px;
  }

  .edit-form :deep(.arco-form-item-label) {
    color: var(--color-text-1);
    font-weight: 500;
  }
</style>

<style lang="less">
  .member-edit-modal {
    .arco-modal-header {
      height: auto;
      padding: 18px 24px 14px;
      border-bottom: none;
    }

    .arco-modal-body {
      max-height: 64vh;
      padding: 0 24px 16px;
      overflow: auto;
    }

    .arco-modal-footer {
      padding: 14px 24px;
    }
  }
</style>
