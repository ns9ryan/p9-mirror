<template>
  <a-modal
    :visible="visible"
    :title="isEdit ? $t('menuModal.title.edit') : $t('menuModal.title.create')"
    :ok-loading="submitting"
    :ok-text="isEdit ? $t('front.common.okText.save') : $t('front.common.okText.create')"
    :mask-closable="false"
    :esc-to-close="false"
    unmount-on-close
    :width="640"
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
      :label-col-props="{ span: 6 }"
      :wrapper-col-props="{ span: 17 }"
    >
      <a-form-item :label="$t('menuModal.field.menuType')" field="menu_type" required>
        <a-radio-group v-model="form.menu_type" :disabled="isEdit">
          <a-radio :value="0">{{ $t('menuManage.type.dir') }}</a-radio>
          <a-radio :value="1">{{ $t('menuManage.type.menu') }}</a-radio>
          <a-radio :value="2">{{ $t('menuManage.type.button') }}</a-radio>
        </a-radio-group>
      </a-form-item>
      <a-form-item :label="$t('menuModal.field.title')" field="title" required>
        <a-input
          v-model="form.title"
          :placeholder="$t('menuModal.field.titlePlaceholder')"
          allow-clear
          :max-length="64"
        />
      </a-form-item>
      <a-form-item :label="$t('menuModal.field.parentId')" field="parent_id">
        <a-tree-select
          v-model="form.parent_id"
          :data="parentTree"
          :field-names="{ key: 'id', title: 'title' }"
          :placeholder="$t('menuModal.field.parentIdPlaceholder')"
          allow-clear
          allow-search
        />
      </a-form-item>
      <a-form-item v-if="!isButton" :label="$t('menuModal.field.path')" field="path" required>
        <a-input
          v-model="form.path"
          :placeholder="$t('menuModal.field.pathPlaceholder')"
          allow-clear
          :max-length="128"
        />
      </a-form-item>
      <a-form-item :label="$t('menuModal.field.name')" field="name" required>
        <a-input
          v-model="form.name"
          :placeholder="isButton ? $t('menuModal.field.namePlaceholderButton') : $t('menuModal.field.namePlaceholderMenu')"
          allow-clear
          :max-length="64"
        />
      </a-form-item>
      <a-form-item v-if="isPage" :label="$t('menuModal.field.component')" field="component">
        <a-input
          v-model="form.component"
          :placeholder="$t('menuModal.field.componentPlaceholder')"
          allow-clear
          :max-length="128"
        />
      </a-form-item>
      <a-form-item v-if="isPage" :label="$t('menuModal.field.redirect')" field="redirect">
        <a-input
          v-model="form.redirect"
          :placeholder="$t('menuModal.field.redirectPlaceholder')"
          allow-clear
          :max-length="128"
        />
      </a-form-item>
      <a-form-item v-if="!isButton" :label="$t('menuModal.field.icon')" field="icon">
        <a-input
          v-model="form.icon"
          :placeholder="$t('menuModal.field.iconPlaceholder')"
          allow-clear
          :max-length="64"
        />
      </a-form-item>
      <a-form-item v-if="!isDir" :label="$t('menuModal.field.permission')" field="permission">
        <a-input
          v-model="form.permission"
          :placeholder="$t('menuModal.field.permissionPlaceholder')"
          allow-clear
          :max-length="128"
        />
      </a-form-item>
      <a-form-item :label="$t('menuModal.field.sort')" field="sort">
        <a-input-number
          v-model="form.sort"
          :min="0"
          :max="9999"
          :placeholder="$t('menuModal.field.sortPlaceholder')"
          style="width: 100%"
        />
      </a-form-item>
      <a-form-item v-if="!isButton" :label="$t('menuModal.field.hideMenu')" field="hide_menu">
        <a-radio-group v-model="form.hide_menu">
          <a-radio :value="0">{{ $t('menuModal.field.hideMenuShow') }}</a-radio>
          <a-radio :value="1">{{ $t('menuModal.field.hideMenuHide') }}</a-radio>
        </a-radio-group>
      </a-form-item>
      <a-form-item :label="$t('menuModal.field.status')" field="disabled">
        <a-radio-group v-model="form.disabled">
          <a-radio :value="0">{{ $t('menuModal.field.statusEnabled') }}</a-radio>
          <a-radio :value="1">{{ $t('menuModal.field.statusDisabled') }}</a-radio>
        </a-radio-group>
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { FormInstance, FieldRule } from '@arco-design/web-vue';
  import {
    MENU_TYPE,
    createMenu,
    updateMenu,
    type AdminMenuRecord,
    type SaveMenuParams,
  } from '@/api/menu';

  const { t, te } = useI18n();

  /** 优先用后端已翻译字段；否则把 i18n key 走 vue-i18n */
  const resolveLabel = (translated?: string, keyOrText?: string) => {
    const fromServer = translated?.trim();
    if (fromServer) return fromServer;
    const raw = keyOrText?.trim() || '';
    if (!raw) return '';
    return te(raw) ? t(raw) : raw;
  };

  interface TreeNode {
    id: number;
    title: string;
    children?: TreeNode[];
    disabled?: boolean;
  }

  const props = defineProps<{
    visible: boolean;
    menu?: AdminMenuRecord | null;
    menus: AdminMenuRecord[];
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  const formRef = ref<FormInstance>();
  const submitting = ref(false);

  const getDefaultForm = (): SaveMenuParams => ({
    parent_id: 0,
    menu_type: MENU_TYPE.MENU,
    path: '',
    name: '',
    component: '',
    redirect: '',
    title: '',
    icon: '',
    permission: '',
    hide_menu: 0,
    sort: 0,
    disabled: 0,
  });

  const form = reactive<SaveMenuParams>(getDefaultForm());

  const isEdit = computed(() => Boolean(props.menu?.id));
  const isDir = computed(() => form.menu_type === MENU_TYPE.DIR);
  const isPage = computed(() => form.menu_type === MENU_TYPE.MENU);
  const isButton = computed(() => form.menu_type === MENU_TYPE.BUTTON);

  const rules = computed<Record<string, FieldRule | FieldRule[]>>(() => ({
    title: [{ required: true, message: t('menuModal.rule.title') }],
    menu_type: [{ required: true, message: t('menuModal.rule.menuType') }],
    name: [{ required: true, message: t('menuModal.rule.name') }],
    permission:
      form.menu_type === MENU_TYPE.BUTTON
        ? [{ required: true, message: t('menuModal.rule.permission') }]
        : [],
  }));

  const collectDescendantIds = (id: number) => {
    const ids = new Set<number>([id]);
    const walk = (parentId: number) => {
      props.menus.forEach((item) => {
        if (item.parent_id === parentId && !ids.has(item.id)) {
          ids.add(item.id);
          walk(item.id);
        }
      });
    };
    walk(id);
    return ids;
  };

  const parentTree = computed<TreeNode[]>(() => {
    const disabledIds =
      isEdit.value && props.menu
        ? collectDescendantIds(props.menu.id)
        : new Set<number>();

    const nodes = props.menus
      .filter((item) => item.menu_type !== MENU_TYPE.BUTTON)
      .map((item) => ({
        id: item.id,
        parent_id: item.parent_id,
        title: resolveLabel(item.trans_title, item.title),
        disabled: disabledIds.has(item.id),
        children: [] as TreeNode[],
      }));

    const nodeMap = new Map(nodes.map((node) => [node.id, node]));
    const roots: TreeNode[] = [{ id: 0, title: t('menuManage.topLevel') }];

    nodes.forEach((node) => {
      const parent = node.parent_id ? nodeMap.get(node.parent_id) : undefined;
      if (parent) {
        parent.children = parent.children || [];
        parent.children.push(node);
      } else {
        roots[0].children = roots[0].children || [];
        roots[0].children.push(node);
      }
    });

    return roots;
  });

  const resetForm = () => {
    Object.assign(form, getDefaultForm());
    formRef.value?.clearValidate();
  };

  const fillForm = (menu: AdminMenuRecord) => {
    Object.assign(form, getDefaultForm(), {
      id: menu.id,
      parent_id: menu.parent_id ?? 0,
      menu_type: menu.menu_type ?? MENU_TYPE.MENU,
      path: menu.path || '',
      name: menu.name || '',
      component: menu.component || '',
      redirect: menu.redirect || '',
      title: menu.title || '',
      icon: menu.icon || '',
      permission: menu.permission || '',
      hide_menu: menu.hide_menu ?? 0,
      sort: menu.sort ?? 0,
      disabled: menu.disabled ?? 0,
    });
  };

  const close = () => {
    emit('update:visible', false);
  };

  const handleCancel = () => {
    close();
    resetForm();
  };

  const buildPayload = (): SaveMenuParams => {
    const payload: SaveMenuParams = {
      parent_id: Number(form.parent_id) || 0,
      menu_type: Number(form.menu_type),
      name: form.name?.trim(),
      title: form.title.trim(),
      sort: form.sort ?? 0,
      disabled: form.disabled ?? 0,
    };
    if (isEdit.value && props.menu) payload.id = props.menu.id;
    if (!isButton.value) {
      payload.path = form.path?.trim() || '';
      payload.icon = form.icon?.trim() || '';
      payload.hide_menu = form.hide_menu ?? 0;
    }
    if (isPage.value) {
      payload.component = form.component?.trim() || '';
      payload.redirect = form.redirect?.trim() || '';
    }
    if (!isDir.value) {
      payload.permission = form.permission?.trim() || '';
    }
    return payload;
  };

  const handleBeforeOk = async (done: (closed: boolean) => void) => {
    const errors = await formRef.value?.validate();
    if (errors || submitting.value) {
      done(false);
      return;
    }

    submitting.value = true;
    try {
      const payload = buildPayload();
      if (isEdit.value && props.menu) {
        await updateMenu({ ...payload, id: props.menu.id });
        Message.success(t('front.common.message.saveSuccess'));
      } else {
        await createMenu(payload);
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
      if (props.menu) fillForm(props.menu);
    }
  );
</script>
