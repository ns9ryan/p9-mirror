<template>
  <a-modal
    :visible="visible"
    :title="t('indieGameModal.title.edit')"
    :ok-loading="submitting"
    :ok-button-props="{ disabled: detailLoading || !detailGame }"
    :ok-text="t('front.common.okText.save')"
    :mask-closable="false"
    :esc-to-close="false"
    unmount-on-close
    :width="680"
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
        <a-form-item :label="t('indieGameModal.field.gameCode')" field="gameCode">
          <a-input v-model="form.gameCode" disabled />
        </a-form-item>
        <a-form-item
          :label="t('indieGameModal.field.name')"
          field="name"
          required
        >
          <a-input
            v-model="form.name"
            :placeholder="t('indieGameModal.field.namePlaceholder')"
            allow-clear
          />
        </a-form-item>
        <a-form-item :label="t('indieGameModal.field.imageUrl')" field="image_url">
          <a-input
            v-model="form.image_url"
            disabled
            :placeholder="t('indieGameModal.field.imageUrlPlaceholder')"
            allow-clear
          />
        </a-form-item>
        <a-form-item
          v-if="form.image_url"
          :label="t('indieGameModal.field.imagePreview')"
        >
          <a-image :src="form.image_url" width="60" height="60" fit="contain" />
        </a-form-item>
        <a-form-item
          :label="t('indieGameModal.field.providerKey')"
          field="provider_key"
        >
          <a-input
            v-model="form.provider_key"
            :placeholder="t('indieGameModal.field.providerKeyPlaceholder')"
            allow-clear
            disabled
          />
        </a-form-item>
        <a-form-item :label="t('indieGameModal.field.sortNo')" field="sort_no">
          <a-input-number
            v-model="form.sort_no"
            :min="1"
            :max="9999"
            :placeholder="t('indieGameModal.field.sortNoPlaceholder')"
            style="width: 100%"
          />
        </a-form-item>
        <a-form-item
          :label="t('indieGameModal.field.supportsEmbed')"
          field="supports_embed"
        >
          <a-switch v-model="form.supports_embed" />
        </a-form-item>
        <a-form-item
          :label="t('indieGameModal.field.supportsRedirect')"
          field="supports_redirect"
        >
          <a-switch v-model="form.supports_redirect" />
        </a-form-item>
        <a-form-item :label="t('indieGameModal.field.status')" field="status">
          <a-radio-group v-model="form.status">
            <a-radio :value="1">{{ t('front.common.status.enabled') }}</a-radio>
            <a-radio :value="2"
              >{{ t('front.common.status.disabled')
              }}<span style="color: red">{{
                t('indieGameModal.field.forceLogoutHint')
              }}</span></a-radio
            >
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
  import type { FieldRule, FormInstance } from '@arco-design/web-vue';
  import {
    getGame,
    updateGame,
    type GameRecord,
    type UpdateGameParams,
  } from '@/api/game';

  const { t } = useI18n();

  const props = defineProps<{
    visible: boolean;
    game?: GameRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  interface GameForm {
    gameCode: string;
    name: string;
    image_url: string;
    provider_key: string;
    sort_no: number;
    supports_embed: boolean;
    supports_redirect: boolean;
    status: number;
  }

  const formRef = ref<FormInstance>();
  const submitting = ref(false);
  const detailLoading = ref(false);
  const detailGame = ref<GameRecord | null>(null);
  const isEdit = computed(() => Boolean(props.game?.id));

  const getDefaultForm = (): GameForm => ({
    gameCode: '',
    name: '',
    image_url: '',
    provider_key: '',
    sort_no: 1,
    supports_embed: false,
    supports_redirect: false,
    status: 1,
  });

  const form = reactive<GameForm>(getDefaultForm());

  const rules = computed<Record<string, FieldRule | FieldRule[]>>(() => ({
    name: [
      { required: true, message: t('indieGameModal.rule.nameRequired') },
    ],
  }));

  const resetForm = () => {
    Object.assign(form, getDefaultForm());
    formRef.value?.clearValidate();
  };

  const fillForm = (game: GameRecord) => {
    Object.assign(form, getDefaultForm(), {
      gameCode: game.game_code || '--',
      name: game.name || '',
      image_url: game.image_url || '',
      provider_key: game.provider_key || '',
      sort_no: game.sort_no && game.sort_no > 0 ? game.sort_no : 1,
      supports_embed: Boolean(game.supports_embed),
      supports_redirect: Boolean(game.supports_redirect),
      status: game.status ?? 1,
    });
  };

  const fetchDetail = async (id: number) => {
    detailLoading.value = true;
    detailGame.value = null;
    resetForm();
    try {
      const { data } = await getGame(id);
      detailGame.value = data ?? null;
      if (data) {
        fillForm(data);
      }
    } catch {
      detailGame.value = null;
    } finally {
      detailLoading.value = false;
    }
  };

  const close = () => {
    emit('update:visible', false);
  };

  const handleCancel = () => {
    close();
    detailGame.value = null;
    resetForm();
  };

  const buildPayload = (): UpdateGameParams => ({
    id: detailGame.value!.id,
    name: form.name.trim(),
    image_url: form.image_url.trim(),
    provider_key: form.provider_key.trim(),
    sort_no: form.sort_no && form.sort_no > 0 ? form.sort_no : 1,
    supports_embed: form.supports_embed,
    supports_redirect: form.supports_redirect,
    status: form.status ?? 1,
    force_logout: form.status === 2,
  });

  const handleBeforeOk = async (done: (closed: boolean) => void) => {
    const errors = await formRef.value?.validate();
    if (
      errors ||
      submitting.value ||
      detailLoading.value ||
      !isEdit.value ||
      !detailGame.value?.id
    ) {
      done(false);
      return;
    }

    submitting.value = true;
    try {
      await updateGame(buildPayload());
      Message.success(t('front.common.message.saveSuccess'));
      detailGame.value = null;
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
    (visible) => {
      if (!visible) return;
      if (props.game?.id) {
        fetchDetail(props.game.id);
      } else {
        detailGame.value = null;
        resetForm();
      }
    }
  );
</script>
