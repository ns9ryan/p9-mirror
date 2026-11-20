<template>
  <div class="container list-page">
    <Breadcrumb :items="['menu.threePart', 'menu.threePart.strategy']" />
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item :label="$t('strategyManage.search.ip')" field="ip">
          <a-input
            v-model="searchForm.ip"
            :placeholder="$t('strategyManage.search.ipPlaceholder')"
            allow-clear
            class="search-input"
            @press-enter="handleSearch"
          />
        </a-form-item>
        <a-form-item>
          <a-space>
            <a-button type="primary" :loading="searching" @click="handleSearch">
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

    <a-card
      v-if="showResult"
      class="general-card table-card"
      :bordered="false"
      :title="$t('strategyManage.result.title')"
    >
      <a-table
        row-key="query"
        :loading="searching"
        :columns="resultColumns"
        :data="resultList"
        :pagination="false"
        :bordered="true"
        :stripe="true"
        :scroll="{ x: 1400 }"
      >
        <template #index="{ rowIndex }">
          {{ rowIndex + 1 }}
        </template>
        <template #source="{ record }">
          <a-tag v-if="record.source !== '--'" color="arcoblue">
            {{ record.source }}
          </a-tag>
          <span v-else>--</span>
        </template>
      </a-table>
    </a-card>

    <a-card class="general-card table-card" :bordered="false">
      <div class="table-toolbar">
        <a-space>
          <a-button
            type="primary"
            :loading="saving"
            :disabled="!isDirty"
            @click="handleSave"
          >
            {{ $t('front.common.okText.save') }}
          </a-button>
          <a-button @click="fetchData">
            <template #icon><icon-refresh /></template>
            {{ $t('front.common.toolbar.refresh') }}
          </a-button>
        </a-space>
        <a-alert type="info" :show-icon="false" class="tip">
          {{ $t('strategyManage.tip') }}
        </a-alert>
      </div>
      <a-table
        row-key="providerId"
        :loading="loading"
        :columns="columns"
        :data="displayList"
        :pagination="false"
        :bordered="true"
        :stripe="true"
      >
        <template #index="{ rowIndex }">
          {{ rowIndex + 1 }}
        </template>
        <template #tag="{ record }">
          <a-tag v-if="record.tag !== '--'" color="arcoblue">
            {{ record.tag }}
          </a-tag>
          <span v-else>--</span>
        </template>
        <template #isEnabled="{ record }">
          <a-switch
            v-model="record.isEnabled"
            :checked-value="true"
            :unchecked-value="false"
            @change="markDirty"
          />
        </template>
        <template #priority="{ record }">
          <a-input-number
            v-model="record.priority"
            :min="1"
            :precision="0"
            hide-button
            class="priority-input"
            @change="markDirty"
          />
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import {
    searchIpgeoIp,
    type IpgeoIpSearchRecord,
  } from '@/api/ipgeo';
  import {
    queryIpgeoPolicyList,
    saveIpgeoPolicy,
    type IpgeoPolicyRecord,
  } from '@/api/ipgeoPolicy';
  import useLoading from '@/hooks/loading';
  import useLocale from '@/hooks/locale';

  interface PolicyRow {
    providerId: number;
    name: string;
    tag: string;
    isEnabled: boolean;
    priority: number;
  }

  interface IpSearchRow {
    query: string;
    country: string;
    countryCode: string;
    province: string;
    city: string;
    isp: string;
    org: string;
    as: string;
    source: string;
  }

  const { t } = useI18n();
  const { currentLocale } = useLocale();
  const { loading, setLoading } = useLoading(true);
  const saving = ref(false);
  const searching = ref(false);
  const showResult = ref(false);
  const isDirty = ref(false);
  const list = ref<PolicyRow[]>([]);
  const resultList = ref<IpSearchRow[]>([]);
  const snapshot = ref('');

  const searchForm = reactive({
    ip: '',
  });

  const displayText = (value?: string | number | null) => {
    if (value === null || value === undefined || value === '') return '--';
    return String(value);
  };

  const mapRecord = (item: IpgeoPolicyRecord): PolicyRow => ({
    providerId: item.provider_id,
    name: displayText(item.name),
    tag: displayText(item.tag),
    isEnabled: Boolean(item.is_enabled),
    priority: item.priority ?? 1,
  });

  const mapSearchRecord = (item: IpgeoIpSearchRecord): IpSearchRow => ({
    query: displayText(item.query),
    country: displayText(item.country),
    countryCode: displayText(item.country_code),
    province: displayText(item.province),
    city: displayText(item.city),
    isp: displayText(item.isp),
    org: displayText(item.org),
    as: displayText(item.as),
    source: displayText(item.source),
  });

  const parseIps = (value: string) =>
    value
      .split(/[\s,，;；]+/)
      .map((ip) => ip.trim())
      .filter(Boolean);

  const buildSnapshot = (rows: PolicyRow[]) =>
    JSON.stringify(
      rows.map((row) => ({
        providerId: row.providerId,
        isEnabled: row.isEnabled,
        priority: row.priority,
      }))
    );

  const displayList = computed(() =>
    [...list.value].sort((a, b) => a.priority - b.priority)
  );

  const markDirty = () => {
    isDirty.value = buildSnapshot(list.value) !== snapshot.value;
  };

  const fetchData = async () => {
    setLoading(true);
    try {
      const { data } = await queryIpgeoPolicyList();
      const rows = (data?.list || []).map(mapRecord);
      list.value = rows;
      snapshot.value = buildSnapshot(rows);
      isDirty.value = false;
    } catch {
      list.value = [];
      snapshot.value = '';
      isDirty.value = false;
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = async () => {
    const ips = parseIps(searchForm.ip);
    if (!ips.length) {
      Message.warning(t('strategyManage.search.ipRequired'));
      return;
    }
    searching.value = true;
    showResult.value = true;
    try {
      const { data } = await searchIpgeoIp({
        ips,
        lang: currentLocale.value,
      });
      resultList.value = (data?.list || []).map(mapSearchRecord);
    } catch {
      resultList.value = [];
    } finally {
      searching.value = false;
    }
  };

  const handleReset = () => {
    searchForm.ip = '';
    resultList.value = [];
    showResult.value = false;
  };

  const columns = computed<TableColumnData[]>(() => [
    {
      title: t('front.common.column.index'),
      dataIndex: 'index',
      slotName: 'index',
      width: 80,
      align: 'center',
    },
    {
      title: t('strategyManage.column.providerId'),
      dataIndex: 'providerId',
      width: 120,
      align: 'center',
    },
    {
      title: t('strategyManage.column.provider'),
      dataIndex: 'name',
      width: 180,
      align: 'center',
    },
    {
      title: t('strategyManage.column.tag'),
      dataIndex: 'tag',
      slotName: 'tag',
      width: 140,
      align: 'center',
    },
    {
      title: t('strategyManage.column.enabled'),
      dataIndex: 'isEnabled',
      slotName: 'isEnabled',
      width: 120,
      align: 'center',
    },
    {
      title: t('strategyManage.column.priority'),
      dataIndex: 'priority',
      slotName: 'priority',
      width: 160,
      align: 'center',
    },
  ]);

  const resultColumns = computed<TableColumnData[]>(() => [
    {
      title: t('front.common.column.index'),
      dataIndex: 'index',
      slotName: 'index',
      width: 70,
      align: 'center',
    },
    {
      title: t('strategyManage.column.query'),
      dataIndex: 'query',
      width: 140,
      align: 'center',
    },
    {
      title: t('strategyManage.column.country'),
      dataIndex: 'country',
      width: 120,
      align: 'center',
    },
    {
      title: t('strategyManage.column.countryCode'),
      dataIndex: 'countryCode',
      width: 100,
      align: 'center',
    },
    {
      title: t('strategyManage.column.province'),
      dataIndex: 'province',
      width: 120,
      align: 'center',
    },
    {
      title: t('strategyManage.column.city'),
      dataIndex: 'city',
      width: 120,
      align: 'center',
    },
    {
      title: t('strategyManage.column.isp'),
      dataIndex: 'isp',
      width: 140,
      align: 'center',
    },
    {
      title: t('strategyManage.column.org'),
      dataIndex: 'org',
      width: 160,
      align: 'center',
    },
    {
      title: t('strategyManage.column.as'),
      dataIndex: 'as',
      width: 140,
      align: 'center',
    },
    {
      title: t('strategyManage.column.source'),
      dataIndex: 'source',
      slotName: 'source',
      width: 120,
      align: 'center',
    },
  ]);

  const handleSave = async () => {
    if (!list.value.length) {
      Message.warning(t('strategyManage.message.empty'));
      return;
    }
    saving.value = true;
    try {
      await saveIpgeoPolicy({
        list: list.value.map((row) => ({
          provider_id: row.providerId,
          is_enabled: row.isEnabled,
          priority: Number(row.priority) || 1,
        })),
      });
      Message.success(t('front.common.message.saveSuccess'));
      await fetchData();
    } catch {
      // 全局拦截器已提示错误
    } finally {
      saving.value = false;
    }
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
    width: 320px;
  }
  .table-toolbar {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 12px;
    margin-bottom: 14px;
  }
  .tip {
    padding: 4px 12px;
  }
  .priority-input {
    width: 100px;
  }
  :deep(.arco-table-th) {
    color: var(--color-text-1);
    font-weight: 600;
    background-color: var(--color-fill-2);
  }
  :deep(.arco-table-td) {
    color: var(--color-text-2);
  }
</style>
