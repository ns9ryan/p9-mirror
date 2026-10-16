<template>
  <a-modal
    :visible="visible"
    :title="
      isEdit ? $t('ipgeoModal.title.edit') : $t('ipgeoModal.title.create')
    "
    :ok-loading="submitting"
    :ok-text="
      isEdit ? $t('front.common.okText.save') : $t('front.common.okText.create')
    "
    :mask-closable="false"
    :esc-to-close="false"
    unmount-on-close
    :width="760"
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    @before-ok="handleBeforeOk"
    @cancel="handleCancel"
  >
    <a-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-align="right"
      :label-col-props="{ span: 7 }"
      :wrapper-col-props="{ span: 16 }"
    >
      <a-row :gutter="12">
        <a-col :span="12">
          <a-form-item :label="$t('ipgeoModal.field.ip')" field="ip" required>
            <a-input
              v-model="form.ip"
              :placeholder="$t('ipgeoModal.field.ipPlaceholder')"
              allow-clear
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="$t('ipgeoModal.field.isp')" field="isp">
            <a-input
              v-model="form.isp"
              :placeholder="$t('ipgeoModal.field.ispPlaceholder')"
              allow-clear
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="$t('ipgeoModal.field.as')" field="as">
            <a-input
              v-model="form.as"
              :placeholder="$t('ipgeoModal.field.asPlaceholder')"
              allow-clear
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="$t('ipgeoModal.field.org')" field="org">
            <a-input
              v-model="form.org"
              :placeholder="$t('ipgeoModal.field.orgPlaceholder')"
              allow-clear
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item
            :label="$t('ipgeoModal.field.continent')"
            field="continent"
          >
            <a-input
              v-model="form.continent"
              :placeholder="$t('ipgeoModal.field.continentPlaceholder')"
              allow-clear
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item
            :label="$t('ipgeoModal.field.continentCode')"
            field="continent_code"
          >
            <a-input
              v-model="form.continent_code"
              :placeholder="$t('ipgeoModal.field.continentCodePlaceholder')"
              allow-clear
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="$t('ipgeoModal.field.country')" field="country">
            <a-input
              v-model="form.country"
              :placeholder="$t('ipgeoModal.field.countryPlaceholder')"
              allow-clear
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item
            :label="$t('ipgeoModal.field.countryCode')"
            field="country_code"
          >
            <a-input
              v-model="form.country_code"
              :placeholder="$t('ipgeoModal.field.countryCodePlaceholder')"
              allow-clear
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item
            :label="$t('ipgeoModal.field.province')"
            field="province"
          >
            <a-input
              v-model="form.province"
              :placeholder="$t('ipgeoModal.field.provincePlaceholder')"
              allow-clear
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item
            :label="$t('ipgeoModal.field.provinceCode')"
            field="province_iso_code"
          >
            <a-input
              v-model="form.province_iso_code"
              :placeholder="$t('ipgeoModal.field.provinceCodePlaceholder')"
              allow-clear
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="$t('ipgeoModal.field.city')" field="city">
            <a-input
              v-model="form.city"
              :placeholder="$t('ipgeoModal.field.cityPlaceholder')"
              allow-clear
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item
            :label="$t('ipgeoModal.field.timezone')"
            field="timezone"
          >
            <a-input
              v-model="form.timezone"
              :placeholder="$t('ipgeoModal.field.timezonePlaceholder')"
              allow-clear
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="$t('ipgeoModal.field.zip')" field="zip">
            <a-input
              v-model="form.zip"
              :placeholder="$t('ipgeoModal.field.zipPlaceholder')"
              allow-clear
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="$t('ipgeoModal.field.postal')" field="postal">
            <a-input
              v-model="form.postal"
              :placeholder="$t('ipgeoModal.field.postalPlaceholder')"
              allow-clear
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="$t('ipgeoModal.field.lat')" field="lat">
            <a-input-number
              v-model="form.lat"
              :placeholder="$t('ipgeoModal.field.latPlaceholder')"
              :precision="6"
              hide-button
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="$t('ipgeoModal.field.lon')" field="lon">
            <a-input-number
              v-model="form.lon"
              :placeholder="$t('ipgeoModal.field.lonPlaceholder')"
              :precision="6"
              hide-button
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
      </a-row>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { FormInstance, FieldRule } from '@arco-design/web-vue';
  import {
    createIpgeoLocal,
    updateIpgeoLocal,
    type IpgeoLocalRecord,
  } from '@/api/ipgeo';

  const props = defineProps<{
    visible: boolean;
    record?: IpgeoLocalRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  const { t } = useI18n();
  const formRef = ref<FormInstance>();
  const submitting = ref(false);
  const isEdit = computed(() => Boolean(props.record?.id));

  interface IpgeoForm {
    ip: string;
    isp: string;
    as: string;
    org: string;
    continent: string;
    continent_code: string;
    country: string;
    country_code: string;
    province: string;
    province_iso_code: string;
    city: string;
    zip: string;
    postal: string;
    lat: number | undefined;
    lon: number | undefined;
    timezone: string;
  }

  const getDefaultForm = (): IpgeoForm => ({
    ip: '',
    isp: '',
    as: '',
    org: '',
    continent: '',
    continent_code: '',
    country: '',
    country_code: '',
    province: '',
    province_iso_code: '',
    city: '',
    zip: '',
    postal: '',
    lat: undefined,
    lon: undefined,
    timezone: '',
  });

  const form = reactive<IpgeoForm>(getDefaultForm());

  const rules = computed<Record<string, FieldRule | FieldRule[]>>(() => ({
    ip: [{ required: true, message: t('ipgeoModal.rule.ip') }],
  }));

  const resetForm = () => {
    Object.assign(form, getDefaultForm());
    formRef.value?.clearValidate();
  };

  const fillForm = (record: IpgeoLocalRecord) => {
    Object.assign(form, {
      ip: record.ip || '',
      isp: record.isp || '',
      as: record.as || '',
      org: record.org || '',
      continent: record.continent || '',
      continent_code: record.continent_code || '',
      country: record.country || '',
      country_code: record.country_code || '',
      province: record.province || '',
      province_iso_code: record.province_iso_code || '',
      city: record.city || '',
      zip: record.zip || '',
      postal: record.postal || '',
      lat: record.lat ?? undefined,
      lon: record.lon ?? undefined,
      timezone: record.timezone || '',
    });
  };

  const close = () => {
    emit('update:visible', false);
  };

  const handleCancel = () => {
    close();
    resetForm();
  };

  const buildPayload = () => ({
    ip: form.ip.trim(),
    isp: form.isp.trim(),
    as: form.as.trim(),
    org: form.org.trim(),
    continent: form.continent.trim(),
    continent_code: form.continent_code.trim(),
    country: form.country.trim(),
    country_code: form.country_code.trim(),
    province: form.province.trim(),
    province_iso_code: form.province_iso_code.trim(),
    city: form.city.trim(),
    zip: form.zip.trim(),
    postal: form.postal.trim(),
    lat: form.lat ?? 0,
    lon: form.lon ?? 0,
    timezone: form.timezone.trim(),
  });

  const handleBeforeOk = async (done: (closed: boolean) => void) => {
    const errors = await formRef.value?.validate();
    if (errors || submitting.value) {
      done(false);
      return;
    }

    submitting.value = true;
    try {
      const payload = buildPayload();
      if (isEdit.value && props.record?.id) {
        await updateIpgeoLocal({
          id: props.record.id,
          ...payload,
        });
        Message.success(t('front.common.message.saveSuccess'));
      } else {
        await createIpgeoLocal(payload);
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
    (visible) => {
      if (!visible) return;
      resetForm();
      if (isEdit.value && props.record) {
        fillForm(props.record);
      }
    }
  );
</script>
