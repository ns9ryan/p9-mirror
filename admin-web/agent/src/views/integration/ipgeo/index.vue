<template>
  <div class="container list-page">
    <Breadcrumb :items="['menu.threePart', 'menu.threePart.ipManage']" />
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item :label="$t('ipManage.search.ip')" field="ip">
          <a-input
            v-model="searchForm.ip"
            :placeholder="$t('ipManage.search.ipPlaceholder')"
            allow-clear
            class="search-input"
            @press-enter="handleSearch"
          />
        </a-form-item>
        <a-form-item>
          <a-space>
            <a-button type="primary" @click="handleSearch">
              <template #icon><icon-search /></template>
              {{ $t('front.common.search.search') }}
            </a-button>
            <a-button @click="handleReset">
              <template #icon><icon-refresh /></template>
              {{ $t('front.common.search.reset') }}
            </a-button>
          </a-space>
        </a-form-item>
      </a-form>
    </a-card>
    <a-card class="general-card table-card" :bordered="false">
      <div class="table-toolbar">
        <a-space>
          <a-button type="primary" @click="handleCreate">
            <template #icon><icon-plus /></template>
            {{ $t('front.common.toolbar.add') }}
          </a-button>
          <a-button @click="() => fetchData()">
            <template #icon><icon-refresh /></template>
            {{ $t('front.common.toolbar.refresh') }}
          </a-button>
        </a-space>
      </div>
      <a-table
        row-key="id"
        :loading="loading"
        :columns="columns"
        :data="displayList"
        :pagination="pagination"
        :bordered="true"
        :stripe="true"
        :scroll="{ x: 2200 }"
        @page-change="onPageChange"
        @page-size-change="onPageSizeChange"
      >
        <template #index="{ rowIndex }">
          {{ rowIndex + 1 + (pagination.current - 1) * pagination.pageSize }}
        </template>
        <template #operations="{ record }">
          <a-space :size="12" class="operation-links">
            <a-link @click="handleEdit(record)">{{
              $t('front.common.action.edit')
            }}</a-link>
            <a-link status="danger" @click="handleDelete(record)">{{
              $t('front.common.action.delete')
            }}</a-link>
          </a-space>
        </template>
      </a-table>
    </a-card>
    <IpgeoModal
      :visible="modalVisible"
      :record="editingRecord"
      @update:visible="onModalVisibleChange"
      @success="fetchData"
    />
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message, Modal } from '@arco-design/web-vue';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import {
    deleteIpgeoLocal,
    getIpgeoLocalDetail,
    queryIpgeoLocalList,
    type IpgeoLocalListParams,
    type IpgeoLocalRecord,
  } from '@/api/ipgeo';
  import useLoading from '@/hooks/loading';
  import IpgeoModal from './components/ipgeo-modal.vue';

  interface IpgeoLocalRow {
    id: number;
    ip: string;
    isp: string;
    as: string;
    org: string;
    country: string;
    countryCode: string;
    continent: string;
    continentCode: string;
    province: string;
    provinceIsoCode: string;
    city: string;
    zip: string;
    postal: string;
    lat: string;
    lon: string;
    timezone: string;
    raw: IpgeoLocalRecord;
  }

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(true);
  const modalVisible = ref(false);
  const editingRecord = ref<IpgeoLocalRecord | null>(null);

  const searchForm = reactive({
    ip: '',
  });
  const appliedSearch = reactive({
    ip: '',
  });

  const list = ref<IpgeoLocalRow[]>([]);
  const pagination = reactive({
    current: 1,
    pageSize: 20,
    total: 0,
    showTotal: true,
    showPageSize: true,
    pageSizeOptions: [10, 20, 30, 50, 100],
    pageSizeProps: {
      popupContainer: 'body',
    },
  });

  const displayText = (value?: string | number | null) => {
    if (value === null || value === undefined || value === '') return '--';
    return String(value);
  };

  const mapRecord = (item: IpgeoLocalRecord): IpgeoLocalRow => ({
    id: item.id,
    ip: displayText(item.ip),
    isp: displayText(item.isp),
    as: displayText(item.as),
    org: displayText(item.org),
    country: displayText(item.country),
    countryCode: displayText(item.country_code),
    continent: displayText(item.continent),
    continentCode: displayText(item.continent_code),
    province: displayText(item.province),
    provinceIsoCode: displayText(item.province_iso_code),
    city: displayText(item.city),
    zip: displayText(item.zip),
    postal: displayText(item.postal),
    lat: displayText(item.lat),
    lon: displayText(item.lon),
    timezone: displayText(item.timezone),
    raw: item,
  });

  const displayList = computed(() => list.value);

  const buildQueryParams = (
    page = pagination.current,
    pageSize = pagination.pageSize
  ): IpgeoLocalListParams => {
    const params: IpgeoLocalListParams = {
      page,
      page_size: pageSize,
    };
    const ip = appliedSearch.ip.trim();
    if (ip) params.ip = ip;
    return params;
  };

  const fetchData = async (
    page = pagination.current,
    pageSize = pagination.pageSize
  ) => {
    setLoading(true);
    try {
      const { data } = await queryIpgeoLocalList(
        buildQueryParams(page, pageSize)
      );
      list.value = (data?.list || []).map(mapRecord);
      pagination.current = page;
      pagination.pageSize = pageSize;
      pagination.total = data?.total ?? 0;
    } catch {
      list.value = [];
      pagination.total = 0;
    } finally {
      setLoading(false);
    }
  };

  const columns = computed<TableColumnData[]>(() => [
    {
      title: t('front.common.column.index'),
      dataIndex: 'index',
      slotName: 'index',
      width: 70,
      align: 'center',
      fixed: 'left',
    },
    {
      title: t('ipManage.column.id'),
      dataIndex: 'id',
      width: 80,
      align: 'center',
      fixed: 'left',
    },
    {
      title: t('ipManage.column.ip'),
      dataIndex: 'ip',
      width: 140,
      align: 'center',
      fixed: 'left',
    },
    {
      title: t('ipManage.column.isp'),
      dataIndex: 'isp',
      width: 140,
      align: 'center',
    },
    {
      title: t('ipManage.column.as'),
      dataIndex: 'as',
      width: 140,
      align: 'center',
    },
    {
      title: t('ipManage.column.org'),
      dataIndex: 'org',
      width: 160,
      align: 'center',
    },
    {
      title: t('ipManage.column.country'),
      dataIndex: 'country',
      width: 120,
      align: 'center',
    },
    {
      title: t('ipManage.column.countryCode'),
      dataIndex: 'countryCode',
      width: 90,
      align: 'center',
    },
    {
      title: t('ipManage.column.continent'),
      dataIndex: 'continent',
      width: 100,
      align: 'center',
    },
    {
      title: t('ipManage.column.continentCode'),
      dataIndex: 'continentCode',
      width: 90,
      align: 'center',
    },
    {
      title: t('ipManage.column.province'),
      dataIndex: 'province',
      width: 120,
      align: 'center',
    },
    {
      title: t('ipManage.column.provinceCode'),
      dataIndex: 'provinceIsoCode',
      width: 100,
      align: 'center',
    },
    {
      title: t('ipManage.column.city'),
      dataIndex: 'city',
      width: 120,
      align: 'center',
    },
    {
      title: t('ipManage.column.zip'),
      dataIndex: 'zip',
      width: 100,
      align: 'center',
    },
    {
      title: t('ipManage.column.postal'),
      dataIndex: 'postal',
      width: 100,
      align: 'center',
    },
    {
      title: t('ipManage.column.lat'),
      dataIndex: 'lat',
      width: 100,
      align: 'center',
    },
    {
      title: t('ipManage.column.lon'),
      dataIndex: 'lon',
      width: 100,
      align: 'center',
    },
    {
      title: t('ipManage.column.timezone'),
      dataIndex: 'timezone',
      width: 140,
      align: 'center',
    },
    {
      title: t('front.common.column.operations'),
      dataIndex: 'operations',
      slotName: 'operations',
      width: 120,
      align: 'center',
      fixed: 'right',
    },
  ]);

  const onModalVisibleChange = (visible: boolean) => {
    modalVisible.value = visible;
    if (!visible) {
      editingRecord.value = null;
    }
  };

  const handleCreate = () => {
    editingRecord.value = null;
    modalVisible.value = true;
  };

  const handleEdit = async (record: IpgeoLocalRow) => {
    setLoading(true);
    try {
      const { data } = await getIpgeoLocalDetail({ id: record.id });
      editingRecord.value = data || record.raw;
      modalVisible.value = true;
    } catch {
      // 全局拦截器已提示错误
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = (record: IpgeoLocalRow) => {
    Modal.confirm({
      title: t('front.common.delete.title'),
      content: t('ipManage.delete.content', { ip: record.ip }),
      okText: t('front.common.delete.okText'),
      okButtonProps: { status: 'danger' },
      async onOk() {
        await deleteIpgeoLocal({ id: record.id });
        Message.success(t('front.common.message.deleteSuccess'));
        const nextPage =
          list.value.length <= 1 && pagination.current > 1
            ? pagination.current - 1
            : pagination.current;
        await fetchData(nextPage, pagination.pageSize);
      },
    });
  };

  const handleSearch = () => {
    appliedSearch.ip = searchForm.ip;
    fetchData(1, pagination.pageSize);
  };

  const handleReset = () => {
    searchForm.ip = '';
    appliedSearch.ip = '';
    fetchData(1, pagination.pageSize);
  };

  const onPageChange = (current: number) => {
    fetchData(current, pagination.pageSize);
  };

  const onPageSizeChange = (pageSize: number) => {
    fetchData(1, pageSize);
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
  .search-form {
    row-gap: 16px;
    column-gap: 16px;
  }
  .search-form :deep(.arco-form-item) {
    margin-bottom: 0;
  }
  .search-input {
    width: 200px;
  }
  .table-toolbar {
    margin-bottom: 14px;
  }
  .operation-links {
    white-space: nowrap;
  }
  :deep(.arco-table-th) {
    color: var(--color-text-1);
    font-weight: 600;
    background-color: var(--color-fill-2);
  }
  :deep(.arco-table-td) {
    color: var(--color-text-2);
  }
  :deep(.arco-pagination-options .arco-select) {
    min-width: 100px;
  }
</style>
