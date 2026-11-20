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
            />
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
              :default-expand-all="false"
            />
            <a-empty v-else-if="!apiLoading" :description="$t('permissionModal.api.empty')" />
          </a-spin>
        </a-tab-pane>
      </a-tabs>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import {
    MENU_TYPE,
    queryMenuList,
    type AdminMenuRecord,
  } from '@/api/menu';
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
    children?: MenuTreeNode[];
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
  /** 菜单 id -> parent_id，用于保存时自动补全祖先节点 */
  const menuParentMap = ref<Map<number, number>>(new Map());

  const apiTreeData = ref<ApiTreeNode[]>([]);
  const apiCheckedKeys = ref<(string | number)[]>([]);
  const apiLeafMap = ref<Map<string, RoleApiItem>>(new Map());
  const requiredApiKeys = ref<string[]>([]);

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
      .map((group) => {
        const children = (group.children || []).filter((leaf) => {
          const text = `${group.title} ${leaf.title} ${leaf.path || ''}`.toLowerCase();
          return text.includes(keyword);
        });
        if (!children.length) return null;
        return { ...group, children };
      })
      .filter(Boolean) as ApiTreeNode[];
  });

  const apiKeyOf = (method: string, path: string) =>
    `${(method || '').toUpperCase()}|${path || ''}`;

  const buildMenuTree = (menus: AdminMenuRecord[]): MenuTreeNode[] => {
    const nodes = menus.map((item) => ({
      id: item.id,
      parent_id: item.parent_id,
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

  const buildApiTree = (apis: ApiRecord[]) => {
    const leafMap = new Map<string, RoleApiItem>();
    const requiredKeys: string[] = [];
    const groupMap = new Map<string, ApiTreeNode>();

    apis.forEach((api) => {
      const method = (api.method || '').toUpperCase();
      const path = api.path || '';
      if (!method || !path) return;

      const key = apiKeyOf(method, path);
      leafMap.set(key, { method, path });

      const groupName = api.api_group?.trim() || t('permissionModal.api.ungrouped');
      if (!groupMap.has(groupName)) {
        groupMap.set(groupName, {
          key: `group:${groupName}`,
          title: groupName,
          children: [],
        });
      }

      const isRequired = Boolean(api.is_required);
      if (isRequired) requiredKeys.push(key);

      const desc = resolveLabel(api.trans_description, api.description);
      groupMap.get(groupName)!.children!.push({
        key,
        title: desc ? `[${method}] ${path}（${desc}）` : `[${method}] ${path}`,
        path,
        method,
        disableCheckbox: isRequired,
      });
    });

    const tree = Array.from(groupMap.values()).map((group) => ({
      ...group,
      children: (group.children || []).sort((a, b) =>
        String(a.title).localeCompare(String(b.title), 'zh-CN')
      ),
    }));
    tree.sort((a, b) => String(a.title).localeCompare(String(b.title), 'zh-CN'));

    apiLeafMap.value = leafMap;
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
    apiCheckedKeys.value = [];
    menuTreeData.value = [];
    menuParentMap.value = new Map();
    apiTreeData.value = [];
    apiLeafMap.value = new Map();
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
      const menuIds = checkedMenuRes.data?.menu_ids || [];
      menuCheckedKeys.value = toTreeCheckedKeys(menuIds, flatMenus);
    } catch {
      menuTreeData.value = [];
      menuCheckedKeys.value = [];
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
      const menuIds = collectMenuIdsWithAncestors(checkedMenuIds);
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
