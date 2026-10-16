<template>
  <a-modal
    :visible="visible"
    :title="`${$t('permissionModal.title')}${roleLabel ? ` - ${roleLabel}` : ''}`"
    :ok-loading="submitting"
    :ok-text="$t('front.common.okText.save')"
    unmount-on-close
    :width="720"
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    @before-ok="handleBeforeOk"
    @cancel="handleCancel"
  >
    <a-tabs v-model:active-key="activeTab" @change="handleTabChange">
        <a-tab-pane key="menu" :title="$t('permissionModal.tab.menu')">
          <a-spin :loading="menuLoading" style="width: 100%">
            <a-tree
              v-if="menuTreeData.length"
              v-model:checked-keys="menuCheckedKeys"
              :data="menuTreeData"
              :field-names="{ key: 'id', title: 'title', children: 'children' }"
              checkable
              :check-strictly="false"
              :default-expand-all="true"
            >
              <template #title="node">
                <span class="menu-node-title">
                  <span>{{ node.title }}</span>
                  <a-button
                    v-if="node.menuType === MENU_TYPE.MENU && hasMenuButtons(node.id)"
                    type="text"
                    size="mini"
                    class="menu-assign-btn"
                    @click.stop="openButtonModal(node)"
                  >
                    {{ $t('permissionModal.menu.assign') }}
                  </a-button>
                </span>
              </template>
            </a-tree>
            <a-empty v-else-if="!menuLoading" :description="$t('permissionModal.menu.empty')" />
          </a-spin>
        </a-tab-pane>
        <a-tab-pane key="api" :title="$t('permissionModal.tab.api')">
          <a-spin :loading="apiLoading" style="width: 100%">
            <a-input
              v-model="apiKeyword"
              allow-clear
              :placeholder="$t('permissionModal.api.searchPlaceholder')"
              style="margin-bottom: 12px"
            />
            <a-tree
              v-if="filteredApiTreeData.length"
              v-model:checked-keys="apiCheckedKeys"
              :data="filteredApiTreeData"
              :field-names="{
                key: 'key',
                title: 'title',
                children: 'children',
                disableCheckbox: 'disableCheckbox',
              }"
              checkable
              :check-strictly="false"
              :default-expand-all="Boolean(apiKeyword.trim())"
            >
              <template #title="{ title }">
                <ApiTreeLabel :label="title" />
              </template>
            </a-tree>
            <a-empty v-else-if="!apiLoading" :description="$t('permissionModal.api.empty')" />
          </a-spin>
        </a-tab-pane>
      </a-tabs>
  </a-modal>

  <a-modal
    v-model:visible="buttonModalVisible"
    :title="buttonModalTitle"
    :ok-text="$t('front.common.okText.save')"
    unmount-on-close
    :width="480"
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    @before-ok="handleButtonBeforeOk"
    @cancel="closeButtonModal"
  >
    <a-checkbox-group
      v-if="buttonOptions.length"
      v-model="buttonDraftIds"
      direction="vertical"
    >
      <a-checkbox
        v-for="item in buttonOptions"
        :key="item.id"
        :value="item.id"
      >
        {{ item.title }}
      </a-checkbox>
    </a-checkbox-group>
    <a-empty v-else :description="$t('permissionModal.menu.assignEmpty')" />
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, defineComponent, h, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message, Tooltip } from '@arco-design/web-vue';
  import { MENU_TYPE, queryMenuList, type AdminMenuRecord } from '@/api/menu';
  import {
    queryApiList,
    type ApiListRes,
    type ApiRecord,
  } from '@/api/apiManage';
  import {
    queryRoleApis,
    queryRoleMenuIds,
    setRoleApis,
    setRoleMenus,
    type RoleApiItem,
    type RoleApiRes,
    type RoleRecord,
  } from '@/api/role';

  const { t, te } = useI18n();

  /** 优先用后端已翻译字段；否则把 i18n key 走 vue-i18n */
  const resolveLabel = (translated?: string, keyOrText?: string) => {
    const fromServer = translated?.trim();
    if (fromServer) return fromServer;
    const raw = keyOrText?.trim() || '';
    if (!raw) return '';
    return te(raw) ? t(raw) : raw;
  };

  interface MenuTreeNode {
    id: number;
    title: string;
    menuType: number;
    children?: MenuTreeNode[];
  }

  interface ButtonOption {
    id: number;
    title: string;
  }

  interface ApiTreeNode {
    key: string;
    title: string;
    disableCheckbox?: boolean;
    children?: ApiTreeNode[];
    /** 叶子节点才有 */
    path?: string;
    method?: string;
  }

  const props = defineProps<{
    visible: boolean;
    role?: RoleRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  const menuLoading = ref(false);
  const apiLoading = ref(false);
  const apiLoaded = ref(false);
  const submitting = ref(false);
  const activeTab = ref('menu');
  const apiKeyword = ref('');

  const menuTreeData = ref<MenuTreeNode[]>([]);
  const menuCheckedKeys = ref<(string | number)[]>([]);
  const allMenuIds = ref<number[]>([]);
  /** 菜单页 id -> 其下 menu_type=2 的按钮 */
  const menuButtonMap = ref<Map<number, ButtonOption[]>>(new Map());
  /** 已分配的按钮权限 id，树中不展示，保存时写回 */
  const selectedButtonIds = ref<number[]>([]);
  const buttonModalVisible = ref(false);
  const buttonModalMenuId = ref<number | null>(null);
  const buttonModalMenuTitle = ref('');
  const buttonDraftIds = ref<number[]>([]);
  /** 菜单 id -> parent_id，用于保存时自动补全祖先节点 */
  const menuParentMap = ref<Map<number, number>>(new Map());

  const apiTreeData = ref<ApiTreeNode[]>([]);
  const apiCheckedKeys = ref<(string | number)[]>([]);
  const apiLeafMap = ref<Map<string, RoleApiItem>>(new Map());
  const apiTipMap = ref<Map<string, string>>(new Map());
  const requiredApiKeys = ref<string[]>([]);

  const buttonOptions = computed(
    () => menuButtonMap.value.get(buttonModalMenuId.value || 0) || []
  );

  const buttonModalTitle = computed(() =>
    buttonModalMenuTitle.value
      ? `${t('permissionModal.menu.assign')} - ${buttonModalMenuTitle.value}`
      : t('permissionModal.menu.assign')
  );

  const roleLabel = computed(() => {
    if (!props.role) return '';
    return (
      props.role.role_name ||
      props.role.name ||
      props.role.display_name ||
      props.role.role_code ||
      ''
    );
  });

  const filteredApiTreeData = computed(() => {
    const keyword = apiKeyword.value.trim().toLowerCase();
    if (!keyword) return apiTreeData.value;

    return apiTreeData.value
      .map((service) => {
        const groups = (service.children || [])
          .map((group) => {
            const children = (group.children || []).filter((leaf) => {
              const text =
                `${service.title} ${group.title} ${leaf.title} ${leaf.method || ''} ${leaf.path || ''}`.toLowerCase();
              return text.includes(keyword);
            });
            if (!children.length) return null;
            return { ...group, children };
          })
          .filter(Boolean) as ApiTreeNode[];
        if (!groups.length) return null;
        return { ...service, children: groups };
      })
      .filter(Boolean) as ApiTreeNode[];
  });

  const apiKeyOf = (method: string, path: string) =>
    `${(method || '').toUpperCase()}|${path || ''}`;

  /** 悬停时再读节点 key，避免树重绘后原生 title 丢失 */
  const ApiTreeLabel = defineComponent({
    name: 'ApiTreeLabel',
    props: {
      label: { type: String, required: true },
    },
    setup(labelProps) {
      const tip = ref('');
      const visible = ref(false);
      const onEnter = (event: MouseEvent) => {
        const key =
          (event.currentTarget as HTMLElement)
            .closest('[data-key]')
            ?.getAttribute('data-key') || '';
        tip.value = apiTipMap.value.get(key) || '';
        visible.value = Boolean(tip.value);
      };
      const onLeave = () => {
        visible.value = false;
      };
      return () =>
        h(
          Tooltip,
          {
            content: tip.value,
            popupVisible: visible.value,
            popupContainer: 'body',
          },
          {
            default: () =>
              h(
                'span',
                { onMouseenter: onEnter, onMouseleave: onLeave },
                labelProps.label
              ),
          }
        );
    },
  });

  const buildMenuTree = (menus: AdminMenuRecord[]): MenuTreeNode[] => {
    const nodes = menus
      .filter((item) => item.menu_type !== MENU_TYPE.BUTTON)
      .map((item) => ({
        id: item.id,
        parent_id: item.parent_id,
        menuType: item.menu_type,
        // 优先用后端按 X-Lang 翻译的 trans_title；否则走 vue-i18n
        title: resolveLabel(item.trans_title, item.title),
        children: [] as MenuTreeNode[],
      }));
    const nodeMap = new Map(nodes.map((node) => [node.id, node]));
    const roots: MenuTreeNode[] = [];

    nodes.forEach((node) => {
      const parent = node.parent_id ? nodeMap.get(node.parent_id) : undefined;
      if (parent) {
        parent.children = parent.children || [];
        parent.children.push(node);
      } else {
        roots.push(node);
      }
    });

    return roots;
  };

  const buildMenuButtonMap = (menus: AdminMenuRecord[]) => {
    const buttonMap = new Map<number, ButtonOption[]>();
    menus.forEach((item) => {
      if (item.menu_type !== MENU_TYPE.BUTTON || !item.parent_id) return;
      const options = buttonMap.get(item.parent_id) || [];
      options.push({
        id: item.id,
        title: resolveLabel(item.trans_title, item.title),
      });
      buttonMap.set(item.parent_id, options);
    });
    menuButtonMap.value = buttonMap;
  };

  const hasMenuButtons = (id: number) =>
    (menuButtonMap.value.get(id) || []).length > 0;

  const openButtonModal = (node: MenuTreeNode) => {
    const options = menuButtonMap.value.get(node.id) || [];
    const optionIds = new Set(options.map((item) => item.id));
    buttonModalMenuId.value = node.id;
    buttonModalMenuTitle.value = node.title;
    buttonDraftIds.value = selectedButtonIds.value.filter((id) =>
      optionIds.has(id)
    );
    buttonModalVisible.value = true;
  };

  const closeButtonModal = () => {
    buttonModalVisible.value = false;
  };

  const handleButtonBeforeOk = (done: (closed: boolean) => void) => {
    const optionIds = new Set(buttonOptions.value.map((item) => item.id));
    const rest = selectedButtonIds.value.filter((id) => !optionIds.has(id));
    const draftIds = buttonDraftIds.value
      .map((id) => Number(id))
      .filter((id) => optionIds.has(id));
    selectedButtonIds.value = [...rest, ...draftIds];
    buttonModalVisible.value = false;
    done(true);
  };

  const normalizeApiList = (data?: ApiListRes | ApiRecord[]) => {
    if (!data) return [] as ApiRecord[];
    if (Array.isArray(data)) return data;
    return data.list || [];
  };

  const normalizeRoleApis = (payload: RoleApiRes | undefined | null): RoleApiItem[] => {
    if (!payload) return [];
    if (Array.isArray(payload)) return payload;
    if (Array.isArray(payload.data)) return payload.data;
    if (Array.isArray(payload.list)) return payload.list;
    if (Array.isArray(payload.apis)) return payload.apis;
    return [];
  };

  const sortByTitle = (a: ApiTreeNode, b: ApiTreeNode) =>
    String(a.title).localeCompare(String(b.title), 'zh-CN');

  const buildApiTree = (apis: ApiRecord[]) => {
    const leafMap = new Map<string, RoleApiItem>();
    const tipMap = new Map<string, string>();
    const requiredKeys: string[] = [];
    const serviceMap = new Map<string, Map<string, ApiTreeNode>>();

    apis.forEach((api) => {
      const method = (api.method || '').toUpperCase();
      const path = api.path || '';
      if (!method || !path) return;

      const key = apiKeyOf(method, path);
      leafMap.set(key, { method, path });

      const serviceName =
        api.service_name?.trim() || t('permissionModal.api.unassignedService');
      if (!serviceMap.has(serviceName)) {
        serviceMap.set(serviceName, new Map());
      }

      const groupName =
        resolveLabel(api.trans_api_group, api.api_group) ||
        t('permissionModal.api.ungrouped');
      const groupMap = serviceMap.get(serviceName)!;
      if (!groupMap.has(groupName)) {
        groupMap.set(groupName, {
          key: `service:${serviceName}|group:${groupName}`,
          title: groupName,
          children: [],
        });
      }

      const isRequired = Boolean(api.is_required);
      if (isRequired) requiredKeys.push(key);

      const desc = resolveLabel(api.trans_description, api.description);
      tipMap.set(key, `[${method}] ${path}`);
      groupMap.get(groupName)!.children!.push({
        key,
        title: desc || '—',
        path,
        method,
        disableCheckbox: isRequired,
      });
    });

    const tree = Array.from(serviceMap.entries()).map(([serviceName, groupMap]) => ({
      key: `service:${serviceName}`,
      title: serviceName,
      children: Array.from(groupMap.values())
        .map((group) => ({
          ...group,
          children: (group.children || []).sort(sortByTitle),
        }))
        .sort(sortByTitle),
    }));
    tree.sort(sortByTitle);

    apiLeafMap.value = leafMap;
    apiTipMap.value = tipMap;
    requiredApiKeys.value = requiredKeys;
    apiTreeData.value = tree;
  };

  /** 将勾选的菜单 id 补全其所有祖先 id（目录节点），保证后端能拼出完整路由树 */
  const collectMenuIdsWithAncestors = (checkedIds: number[]): number[] => {
    const result = new Set<number>();
    const parentMap = menuParentMap.value;

    checkedIds.forEach((id) => {
      let current: number | undefined = id;
      while (current) {
        result.add(current);
        const parentId = parentMap.get(current);
        if (!parentId) break;
        current = parentId;
      }
    });

    return Array.from(result).filter((id) => allMenuIds.value.includes(id));
  };

  /**
   * 后端 menu_ids 含保存时自动补全的目录节点；回显到树组件时需去掉目录 id，
   * 否则父子联动会把该目录下所有子菜单级联勾选。
   */
  const toTreeCheckedKeys = (
    savedIds: number[],
    flatMenus: AdminMenuRecord[]
  ): number[] => {
    const dirIds = new Set(
      flatMenus
        .filter((item) => item.menu_type === MENU_TYPE.DIR)
        .map((item) => item.id)
    );
    return savedIds.filter(
      (id) => allMenuIds.value.includes(id) && !dirIds.has(id)
    );
  };

  const flattenMenus = (items: AdminMenuRecord[]): AdminMenuRecord[] => {
    const result: AdminMenuRecord[] = [];
    const walk = (list: AdminMenuRecord[]) => {
      list.forEach((item) => {
        const { children, ...rest } = item;
        result.push(rest);
        if (children?.length) walk(children);
      });
    };
    walk(items);
    return result;
  };

  const resetState = () => {
    activeTab.value = 'menu';
    apiKeyword.value = '';
    menuCheckedKeys.value = [];
    selectedButtonIds.value = [];
    menuButtonMap.value = new Map();
    buttonModalVisible.value = false;
    buttonModalMenuId.value = null;
    buttonModalMenuTitle.value = '';
    buttonDraftIds.value = [];
    apiCheckedKeys.value = [];
    menuTreeData.value = [];
    menuParentMap.value = new Map();
    apiTreeData.value = [];
    apiLeafMap.value = new Map();
    apiTipMap.value = new Map();
    requiredApiKeys.value = [];
    apiLoaded.value = false;
  };

  const loadMenuData = async () => {
    if (!props.role?.id) return;
    menuLoading.value = true;
    try {
      const [menuRes, checkedMenuRes] = await Promise.all([
        queryMenuList(),
        queryRoleMenuIds({ role_id: props.role.id }).catch(() => ({
          data: { menu_ids: props.role?.menu_ids || [] },
        })),
      ]);

      const menus = (menuRes.data || []) as AdminMenuRecord[];
      const flatMenus = flattenMenus(menus);
      allMenuIds.value = flatMenus.map((item) => item.id);
      menuParentMap.value = new Map(
        flatMenus
          .filter((item) => item.parent_id)
          .map((item) => [item.id, item.parent_id])
      );
      menuTreeData.value = buildMenuTree(flatMenus);
      buildMenuButtonMap(flatMenus);
      const menuIds = checkedMenuRes.data?.menu_ids || [];
      const buttonIdSet = new Set(
        flatMenus
          .filter((item) => item.menu_type === MENU_TYPE.BUTTON)
          .map((item) => item.id)
      );
      selectedButtonIds.value = menuIds.filter((id) => buttonIdSet.has(id));
      menuCheckedKeys.value = toTreeCheckedKeys(menuIds, flatMenus).filter(
        (id) => !buttonIdSet.has(id)
      );
    } catch {
      menuTreeData.value = [];
      menuCheckedKeys.value = [];
      selectedButtonIds.value = [];
      menuButtonMap.value = new Map();
    } finally {
      menuLoading.value = false;
    }
  };

  const loadApiData = async () => {
    if (!props.role?.id || apiLoaded.value) return;
    apiLoading.value = true;
    try {
      const [apiRes, checkedApiRes] = await Promise.all([
        queryApiList({ page: 1, page_size: 10000 }),
        queryRoleApis({ role_id: props.role.id }).catch(() => ({ data: [] })),
      ]);

      const apis = normalizeApiList(apiRes.data);
      buildApiTree(apis);
      const roleApis = normalizeRoleApis(checkedApiRes.data);
      const checked = new Set<string>(requiredApiKeys.value);
      roleApis.forEach((item) => {
        const key = apiKeyOf(item.method, item.path);
        if (apiLeafMap.value.has(key)) checked.add(key);
      });
      apiCheckedKeys.value = Array.from(checked);
      apiLoaded.value = true;
    } catch {
      apiTreeData.value = [];
      apiCheckedKeys.value = [];
    } finally {
      apiLoading.value = false;
    }
  };

  const handleTabChange = (key: string | number) => {
    if (key === 'api') loadApiData();
  };

  const close = () => {
    emit('update:visible', false);
  };

  const handleCancel = () => {
    close();
  };

  const collectCheckedApis = (): RoleApiItem[] => {
    const selected = new Set<string>(
      apiCheckedKeys.value.map(String).filter((key) => apiLeafMap.value.has(key))
    );
    requiredApiKeys.value.forEach((key) => selected.add(key));
    return Array.from(selected)
      .map((key) => apiLeafMap.value.get(key))
      .filter(Boolean) as RoleApiItem[];
  };

  const handleBeforeOk = async (done: (closed: boolean) => void) => {
    if (!props.role?.id || submitting.value) {
      done(false);
      return;
    }
    submitting.value = true;
    try {
      const checkedMenuIds = menuCheckedKeys.value
        .map((key) => Number(key))
        .filter((id) => !Number.isNaN(id));
      const menuIds = Array.from(
        new Set([
          ...collectMenuIdsWithAncestors(checkedMenuIds),
          ...selectedButtonIds.value,
        ])
      );
      const saveTasks: Promise<unknown>[] = [
        setRoleMenus({ role_id: props.role.id, menu_ids: menuIds }),
      ];
      // 未打开接口 Tab 时不提交接口权限，避免用空数据覆盖已有配置
      if (apiLoaded.value) {
        saveTasks.push(
          setRoleApis({ role_id: props.role.id, data: collectCheckedApis() })
        );
      }
      await Promise.all(saveTasks);
      Message.success(t('front.common.message.saveSuccess'));
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
      resetState();
      loadMenuData();
    }
  );
</script>

<style lang="less" scoped>
  .menu-node-title {
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }

  .menu-assign-btn {
    height: 22px;
    padding: 0 4px;
  }
</style>
