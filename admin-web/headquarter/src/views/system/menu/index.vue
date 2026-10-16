<template>
  <div class="container list-page">
    <Breadcrumb :items="['menu.route.system', 'menu.route.menu']" />
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <!-- <a-form-item field="keyword">
          <a-input-group compact>
            <a-select v-model="searchType" class="search-type">
              <a-option value="title">菜单名称</a-option>
              <a-option value="path">路由路径</a-option>
            </a-select>
            <a-input
              v-model="searchForm.keyword"
              placeholder="请输入"
              allow-clear
              class="search-input"
            />
          </a-input-group>
        </a-form-item>
        <a-form-item label="菜单类型" field="menuType">
          <a-select
            v-model="searchForm.menuType"
            placeholder="请选择"
            allow-clear
          >
            <a-option :value="0">目录</a-option>
            <a-option :value="1">菜单</a-option>
            <a-option :value="2">按钮</a-option>
          </a-select>
        </a-form-item> -->
        <a-form-item>
          <a-space>
            <a-button type="primary" @click="handleSearch">
              <template #icon><icon-search /></template>
              {{ $t('front.common.search.search') }}
            </a-button>
            <!-- <a-button @click="handleReset">
              <template #icon><icon-refresh /></template>
              重置
            </a-button> -->
          </a-space>
        </a-form-item>
      </a-form>
    </a-card>
    <a-card class="general-card table-card" :bordered="false">
      <div class="table-toolbar">
        <a-button type="primary" @click="handleCreate">
          <template #icon><icon-plus /></template>
          {{ $t('front.common.toolbar.add') }}
        </a-button>
      </div>
      <a-table
        row-key="id"
        :loading="loading"
        :columns="columns"
        :data="(displayMenus as unknown as TableData[])"
        :pagination="false"
        :bordered="true"
        :stripe="true"
        :scroll="{ x: 1500 }"
        :expanded-keys="expandedKeys"
        @update:expanded-keys="(keys) => (expandedKeys = keys)"
      >
        <template #icon="{ record }">
          <component
            :is="resolveIconName(record.icon)"
            v-if="resolveIconName(record.icon)"
          />
          <span v-else class="cell-placeholder">-</span>
        </template>
        <template #menuType="{ record }">
          <a-tag v-if="record.menu_type === MENU_TYPE.DIR">{{ $t('menuManage.type.dir') }}</a-tag>
          <a-tag v-else-if="record.menu_type === MENU_TYPE.MENU" color="arcoblue"
            >{{ $t('menuManage.type.menu') }}</a-tag
          >
          <a-tag v-else-if="record.menu_type === MENU_TYPE.BUTTON" color="orangered"
            >{{ $t('menuManage.type.button') }}</a-tag
          >
          <span v-else class="cell-placeholder">-</span>
        </template>
        <template #path="{ record }">
          {{ record.path || '-' }}
        </template>
        <template #permission="{ record }">
          {{ record.permission || '-' }}
        </template>
        <template #status="{ record }">
          <a-switch
            :model-value="!record.disabled"
            :checked-text="$t('front.common.status.enabled')"
            :unchecked-text="$t('front.common.status.disabled')"
            :loading="togglingId === record.id"
            @change="(checked) => handleToggleDisabled(record, checked as boolean)"
          />
        </template>
        <template #hideMenu="{ record }">
          <span
            v-if="record.menu_type === MENU_TYPE.BUTTON"
            class="cell-placeholder"
            >-</span
          >
          <a-switch
            v-else
            :model-value="Boolean(record.hide_menu)"
            :checked-text="$t('menuManage.hideMenu.yes')"
            :unchecked-text="$t('menuManage.hideMenu.no')"
            :loading="togglingId === record.id"
            @change="(checked) => handleToggleHideMenu(record, checked as boolean)"
          />
        </template>
        <template #operations="{ record }">
          <a-space :size="12" class="operation-links">
            <a-link @click="handleEdit(record)">
              <template #icon><icon-edit /></template>
              {{ $t('front.common.action.edit') }}
            </a-link>
            <a-link
              status="danger"
              :disabled="deletingId === record.id"
              @click="handleDelete(record)"
            >
              <template #icon><icon-delete /></template>
              {{ $t('front.common.action.delete') }}
            </a-link>
          </a-space>
        </template>
      </a-table>
    </a-card>
    <MenuModal
      :visible="modalVisible"
      :menu="editingMenu"
      :menus="allMenus"
      @update:visible="onModalVisibleChange"
      @success="fetchData"
    />
  </div>
</template>

<script lang="ts" setup>
  import { computed, getCurrentInstance, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message, Modal } from '@arco-design/web-vue';
  import type {
    TableColumnData,
    TableData,
  } from '@arco-design/web-vue/es/table/interface';
  import dayjs from 'dayjs';
  import {
    MENU_TYPE,
    deleteMenu,
    queryMenuList,
    updateMenu,
    type AdminMenuRecord,
  } from '@/api/menu';
  import useLoading from '@/hooks/loading';
  import { useAppStore } from '@/store';
  import MenuModal from './components/menu-modal.vue';

  interface MenuTreeNode extends AdminMenuRecord {
    children?: MenuTreeNode[];
  }

  const { t } = useI18n();
  const appStore = useAppStore();

  const { loading, setLoading } = useLoading(true);
  const modalVisible = ref(false);
  const editingMenu = ref<AdminMenuRecord | null>(null);
  const deletingId = ref<number | null>(null);
  const togglingId = ref<number | null>(null);
  const searchType = ref<'title' | 'path'>('title');
  const searchForm = reactive<{
    keyword: string;
    menuType: number | '';
  }>({
    keyword: '',
    menuType: '',
  });
  const appliedSearch = reactive({
    type: 'title' as 'title' | 'path',
    keyword: '',
    menuType: '' as number | '',
  });
  const allMenus = ref<AdminMenuRecord[]>([]);
  const expandedKeys = ref<(string | number)[]>([]);

  const instance = getCurrentInstance();

  const collectExpandableIds = (nodes: MenuTreeNode[]) => {
    const ids: number[] = [];
    const walk = (list: MenuTreeNode[]) => {
      list.forEach((node) => {
        if (node.children?.length) {
          ids.push(node.id);
          walk(node.children);
        }
      });
    };
    walk(nodes);
    return ids;
  };
  const toPascalCase = (value: string) =>
    value
      .replace(/^icon-/i, '')
      .replace(/(^\w|[-_]\w)/g, (s) => s.replace(/[-_]/, '').toUpperCase());
  const resolveIconName = (icon?: string) => {
    if (!icon) return '';
    const pascal = `Icon${toPascalCase(icon)}`;
    return instance?.appContext.components?.[pascal] ? pascal : '';
  };

  const formatTime = (value?: number) => {
    if (!value) return '--';
    const ms = value < 1e12 ? value * 1000 : value;
    return dayjs(ms).format('YYYY-MM-DD HH:mm:ss');
  };

  const flattenMenus = (items: AdminMenuRecord[] = []) => {
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

  const buildTree = (items: AdminMenuRecord[]): MenuTreeNode[] => {
    const nodes = new Map<number, MenuTreeNode>();
    items.forEach((item) => {
      const id = Number(item.id);
      nodes.set(id, {
        ...item,
        id,
        parent_id: Number(item.parent_id) || 0,
        children: [],
      });
    });

    const roots: MenuTreeNode[] = [];
    nodes.forEach((node) => {
      const parent = node.parent_id ? nodes.get(node.parent_id) : undefined;
      if (parent) {
        parent.children = parent.children || [];
        parent.children.push(node);
      } else {
        roots.push(node);
      }
    });

    const sortTree = (list: MenuTreeNode[]) => {
      list.sort((a, b) => (a.sort ?? 0) - (b.sort ?? 0));
      list.forEach((node) => {
        if (node.children?.length) {
          sortTree(node.children);
        } else {
          delete node.children;
        }
      });
    };
    sortTree(roots);
    return roots;
  };

  const hasActiveFilter = computed(
    () => Boolean(appliedSearch.keyword.trim()) || appliedSearch.menuType !== ''
  );

  const menuTree = computed(() => buildTree(allMenus.value));

  const flatFiltered = computed(() => {
    const keyword = appliedSearch.keyword.trim().toLowerCase();
    return allMenus.value
      .filter((item) => {
        const matchedType =
          appliedSearch.menuType === '' ||
          item.menu_type === appliedSearch.menuType;
        if (!matchedType) return false;
        if (!keyword) return true;
        const source =
          appliedSearch.type === 'path' ? item.path || '' : item.title || '';
        return source.toLowerCase().includes(keyword);
      })
      .sort((a, b) => (a.sort ?? 0) - (b.sort ?? 0));
  });

  const displayMenus = computed<MenuTreeNode[]>(() =>
    hasActiveFilter.value ? flatFiltered.value : menuTree.value
  );

  const columns = computed<TableColumnData[]>(() => [
    { title: t('menuManage.column.title'), dataIndex: 'trans_title', width: 220 },
    {
      title: t('menuManage.column.type'),
      dataIndex: 'menu_type',
      slotName: 'menuType',
      width: 90,
      align: 'center',
    },
    {
      title: t('menuManage.column.icon'),
      dataIndex: 'icon',
      slotName: 'icon',
      width: 80,
      align: 'center',
    },
    {
      title: t('menuManage.column.path'),
      dataIndex: 'path',
      slotName: 'path',
      width: 200,
    },
    {
      title: t('menuManage.column.permission'),
      dataIndex: 'permission',
      slotName: 'permission',
      width: 160,
    },
    { title: t('menuManage.column.sort'), dataIndex: 'sort', width: 90, align: 'center' },
    {
      title: t('front.common.column.status'),
      dataIndex: 'disabled',
      slotName: 'status',
      width: 110,
      align: 'center',
    },
    {
      title: t('menuManage.column.hideMenu'),
      dataIndex: 'hide_menu',
      slotName: 'hideMenu',
      width: 100,
      align: 'center',
    },
    {
      title: t('front.common.column.createdAt'),
      dataIndex: 'created_at',
      width: 180,
      align: 'center',
      render: ({ record }) => formatTime((record as MenuTreeNode).created_at),
    },
    {
      title: t('front.common.column.operations'),
      dataIndex: 'operations',
      slotName: 'operations',
      width: 140,
      align: 'center',
      fixed: 'right',
    },
  ]);

  const fetchData = async () => {
    setLoading(true);
    try {
      const { data } = await queryMenuList();
      allMenus.value = flattenMenus(data || []);
      // 异步数据加载后需手动展开，default-expand-all-rows 只对首次空数据生效
      expandedKeys.value = collectExpandableIds(buildTree(allMenus.value));
    } catch {
      allMenus.value = [];
      expandedKeys.value = [];
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = () => {
    appliedSearch.type = searchType.value;
    appliedSearch.keyword = searchForm.keyword;
    appliedSearch.menuType = searchForm.menuType;
    fetchData();
  };

  const handleReset = () => {
    searchType.value = 'title';
    Object.assign(searchForm, { keyword: '', menuType: '' });
    Object.assign(appliedSearch, { type: 'title', keyword: '', menuType: '' });
    fetchData();
  };

  const onModalVisibleChange = (visible: boolean) => {
    modalVisible.value = visible;
    if (!visible) editingMenu.value = null;
  };

  const handleCreate = () => {
    editingMenu.value = null;
    modalVisible.value = true;
  };

  const handleEdit = (record: MenuTreeNode) => {
    editingMenu.value = record;
    modalVisible.value = true;
  };

  const handleDelete = (record: MenuTreeNode) => {
    Modal.confirm({
      title: t('front.common.delete.title'),
      content: t('menuManage.delete.content', { title: record.title }),
      okText: t('front.common.delete.okText'),
      okButtonProps: { status: 'danger' },
      async onOk() {
        deletingId.value = record.id;
        try {
          await deleteMenu({ ids: [record.id] });
          Message.success(t('front.common.message.deleteSuccess'));
          await fetchData();
        } finally {
          deletingId.value = null;
        }
      },
    });
  };

  const patchMenuField = async (
    record: MenuTreeNode,
    field: 'disabled' | 'hide_menu',
    value: number
  ) => {
    const target = allMenus.value.find((item) => item.id === record.id);
    const previous = target?.[field];
    if (target) target[field] = value;
    togglingId.value = record.id;
    try {
      await updateMenu({ id: record.id, [field]: value });
      if (field === 'disabled') {
        const code = record.permission?.trim();
        if (code) appStore.setMenuPermissionDisabled(code, value === 1);
      }
      Message.success(t('front.common.message.updateSuccess'));
    } catch {
      if (target) target[field] = previous;
    } finally {
      togglingId.value = null;
    }
  };

  const handleToggleDisabled = (record: MenuTreeNode, checked: boolean) => {
    patchMenuField(record, 'disabled', checked ? 0 : 1);
  };

  const handleToggleHideMenu = (record: MenuTreeNode, checked: boolean) => {
    patchMenuField(record, 'hide_menu', checked ? 1 : 0);
  };

  fetchData();
</script>

<style scoped lang="less">
  .container {
    padding: 0 20px 20px;
  }
  .general-card {
    margin-top: 16px;
  }
  .general-card :deep(.arco-card-body) {
    padding: 20px;
  }
  .search-form :deep(.arco-form-item) {
    margin-bottom: 0;
  }
  .search-form :deep(.arco-select-view) {
    min-width: 160px;
  }
  .search-type :deep(.arco-select-view) {
    min-width: 112px;
  }
  .search-input {
    width: 170px;
  }
  .table-toolbar {
    margin-bottom: 14px;
  }
  :deep(.arco-table-th) {
    color: var(--color-text-1);
    font-weight: 600;
    background-color: var(--color-fill-2);
  }
  :deep(.arco-table-td) {
    color: var(--color-text-2);
  }
  .operation-links {
    white-space: nowrap;
  }
  .cell-placeholder {
    color: var(--color-text-4);
  }
</style>
