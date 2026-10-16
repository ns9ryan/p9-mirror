<template>
  <div class="container list-page">
    <Breadcrumb
      :items="['menu.discountCenter', 'menu.discountCenter.activityCenter']"
    />
    <a-card class="general-card" :bordered="false">
      <a-form :model="formModel" layout="inline" class="search-form">
        <a-form-item field="keyword">
          <a-input-group compact>
            <a-select
              v-model="formModel.keywordField"
              class="search-type"
              :options="keywordFieldOptions"
            />
            <a-input
              v-model="formModel.keyword"
              class="search-input"
              allow-clear
              :placeholder="
                $t('discountCenter.activityCenter.form.keyword.placeholder')
              "
            />
          </a-input-group>
        </a-form-item>
        <a-form-item
          :label="$t('discountCenter.activityCenter.form.activityType')"
          field="activityType"
        >
          <a-select
            v-model="formModel.activityType"
            :options="activityTypeOptions"
            allow-clear
            :placeholder="
              $t('discountCenter.activityCenter.form.selectDefault')
            "
          />
        </a-form-item>
        <a-form-item
          :label="$t('discountCenter.activityCenter.form.client')"
          field="client"
        >
          <a-select
            v-model="formModel.client"
            :options="clientOptions"
            allow-clear
            :placeholder="
              $t('discountCenter.activityCenter.form.selectDefault')
            "
          />
        </a-form-item>
        <a-form-item
          :label="$t('discountCenter.activityCenter.form.status')"
          field="status"
        >
          <a-select
            v-model="formModel.status"
            :options="statusOptions"
            allow-clear
            :placeholder="
              $t('discountCenter.activityCenter.form.selectDefault')
            "
          />
        </a-form-item>
        <a-form-item>
          <a-space>
            <a-button type="primary" @click="search">
              <template #icon><icon-search /></template>
              {{ $t('discountCenter.activityCenter.form.search') }}
            </a-button>
            <a-button @click="reset">
              <template #icon><icon-refresh /></template>
              {{ $t('discountCenter.activityCenter.form.reset') }}
            </a-button>
          </a-space>
        </a-form-item>
      </a-form>
    </a-card>
    <a-card class="general-card table-card" :bordered="false">
      <div class="table-toolbar">
        <a-space>
          <a-button type="primary" @click="createDialogVisible = true">
            <template #icon><icon-plus /></template>
            {{ $t('discountCenter.activityCenter.operation.create') }}
          </a-button>
          <a-dropdown @select="handleSelectDensity">
            <a-tooltip
              :content="$t('discountCenter.activityCenter.actions.density')"
            >
              <div class="action-icon"><icon-line-height size="18" /></div>
            </a-tooltip>
            <template #content>
              <a-doption
                v-for="item in densityList"
                :key="item.value"
                :value="item.value"
                :class="{ active: item.value === size }"
              >
                <span>{{ item.name }}</span>
              </a-doption>
            </template>
          </a-dropdown>
          <a-tooltip
            :content="$t('discountCenter.activityCenter.actions.columnSetting')"
          >
            <a-popover
              trigger="click"
              position="bl"
              @popup-visible-change="popupVisibleChange"
            >
              <div class="action-icon"><icon-settings size="18" /></div>
              <template #content>
                <div id="tableSetting">
                  <div
                    v-for="(item, index) in showColumns"
                    :key="item.dataIndex"
                    class="setting"
                  >
                    <div style="margin-right: 4px; cursor: move">
                      <icon-drag-arrow />
                    </div>
                    <div>
                      <a-checkbox
                        v-model="item.checked"
                        @change="handleChange($event, item as Column, index)"
                      />
                    </div>
                    <div class="title">
                      {{ item.title }}
                    </div>
                  </div>
                </div>
              </template>
            </a-popover>
          </a-tooltip>
        </a-space>
      </div>

      <div class="activity-table-wrap">
        <a-table
          row-key="activityId"
          class="activity-table"
          :data="paginatedTableData"
          :pagination="false"
          :bordered="{ cell: true }"
          :show-header="false"
        >
          <template #columns>
            <a-table-column data-index="activityId">
              <template #cell="{ record }">
                <div class="activity-list-row">
                  <div class="activity-list-header">
                    <span>排序：{{ record.sort }}</span>
                    <span>活动 ID：{{ record.activityId }}</span>
                    <span>货币：{{ record.currency }}</span>
                    <span
                      >在列表显示：<span class="show-tag">{{
                        record.listed || '是'
                      }}</span></span
                    >
                    <span>
                      自动审核配置：
                      <a-popover
                        trigger="click"
                        position="bottom"
                        :popup-visible="
                          autoReviewPopoverVisibleId === record.activityId
                        "
                        @popup-visible-change="
                          handleAutoReviewPopoverVisibleChange(
                            $event,
                            record.activityId
                          )
                        "
                      >
                        <button
                          type="button"
                          class="review-tag review-config-trigger"
                        >
                          {{ record.autoReviewConfig }}
                        </button>
                        <template #content>
                          <a-space
                            direction="vertical"
                            :size="12"
                            class="auto-review-form"
                          >
                            <a-switch
                              v-model="record.autoReviewEnabled"
                              checked-text="开"
                              unchecked-text="关"
                            />
                            <a-space v-if="record.autoReviewEnabled" :size="8">
                              <a-select
                                v-model="record.autoReviewOperator"
                                :options="autoReviewOperatorOptions"
                                placeholder="请选择"
                                :style="{ width: '96px' }"
                              />
                              <a-input-number
                                v-model="record.autoReviewValue"
                                :min="0"
                                placeholder="0.00"
                                :style="{ width: '112px' }"
                              />
                            </a-space>
                            <a-space :size="8">
                              <a-button
                                type="primary"
                                size="mini"
                                @click="confirmAutoReview(record)"
                                >确认</a-button
                              >
                              <a-button
                                size="mini"
                                @click="cancelAutoReview(record)"
                                >取消</a-button
                              >
                            </a-space>
                          </a-space>
                        </template>
                      </a-popover>
                    </span>
                  </div>
                  <div class="activity-list-body">
                    <div class="info-block">
                      <div class="block-title"
                        ><span class="block-mark"></span>活动图片 &amp;
                        名称</div
                      >
                      <a-space :size="12" align="start">
                        <a-image
                          class="activity-image"
                          :src="activityImageSrc"
                          :width="72"
                          :height="72"
                          :preview="false"
                          fit="cover"
                        />
                        <div class="block-content"
                          ><div class="primary-name">{{
                            record.activityName
                          }}</div
                          ><div class="sub-value">{{ record.activityType }}</div
                          ><div class="sub-value muted">{{
                            record.activityCondition || '无条件啊'
                          }}</div></div
                        >
                      </a-space>
                    </div>
                    <div class="info-block"
                      ><div class="block-title"
                        ><span class="block-mark orange"></span>活动设置</div
                      ><div class="block-content compact-list"
                        ><div
                          ><strong>参与类型：</strong
                          >{{ record.activityCondition || '无条件' }}</div
                        ><div
                          ><strong>活动时间：</strong
                          >{{ record.activityTime }}</div
                        ><div
                          ><strong>活动方式：</strong
                          >{{ record.listed || '是' }}</div
                        ></div
                      ></div
                    >
                    <div class="info-block"
                      ><div class="block-title"
                        ><span class="block-mark blue"></span>奖金模式 &amp;
                        洗码限制</div
                      ><div class="block-content compact-list"
                        ><div
                          ><strong>奖金模式：</strong
                          >{{ record.claimCount }}</div
                        ><div
                          ><strong>洗码倍数：</strong>
                          <strong :style="{ color: 'rgb(var(--red-6))' }">
                            {{ record.washMultiplier || '---' }}
                          </strong> </div
                        ><div
                          ><strong>参与人数：</strong
                          >{{ record.participantMembers }}</div
                        ><div
                          ><strong>参与层级：</strong
                          >{{ record.participantLevel }}</div
                        ></div
                      ></div
                    >
                    <div class="info-block"
                      ><div class="block-title"
                        ><span class="block-mark green"></span>参与/客户端</div
                      ><div class="block-content compact-list"
                        ><div
                          ><strong>客户端：</strong
                          >{{ record.participantClient }}</div
                        ><div
                          ><strong>指定游戏：</strong
                          >{{ record.specifiedGames }}</div
                        ></div
                      ></div
                    >
                    <div class="info-block status-block"
                      ><div class="block-title"
                        ><span class="block-mark purple"></span>状态</div
                      ><div class="block-content status-wrap"
                        ><a-tag
                          size="medium"
                          :color="getStatusTagColor(record.status)"
                          >{{ $t(getStatusLabelKey(record.status)) }}</a-tag
                        ><span class="status-description">活动状态</span></div
                      ></div
                    >
                  </div>
                  <div class="activity-list-footer">
                    <span>创建人：{{ record.createdBy }}</span>
                    <span>创建时间：{{ record.createdTime }}</span>
                    <a-space class="row-actions" :size="8">
                      <a-button type="primary" size="small">编辑</a-button>
                      <a-button size="small">明细</a-button>
                      <a-button status="danger" size="small">删除</a-button>
                    </a-space>
                  </div>
                </div>
              </template>
            </a-table-column>
          </template>
        </a-table>

        <div class="table-pagination">
          <a-pagination
            :total="pagination.total"
            :current="pagination.current"
            :page-size="pagination.pageSize"
            :page-size-options="pagination.pageSizeOptions"
            :show-total="pagination.showTotal"
            :show-page-size="pagination.showPageSize"
            :page-size-props="pagination.pageSizeProps"
            @change="handlePageChange"
            @page-size-change="handlePageSizeChange"
          />
        </div>
      </div>
    </a-card>

    <CreateActivityDialog v-model:visible="createDialogVisible" />
  </div>
</template>

<script lang="ts" setup>
  import { computed, nextTick, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import type { SelectOptionData } from '@arco-design/web-vue/es/select/interface';
  import cloneDeep from 'lodash/cloneDeep';
  import Sortable from 'sortablejs';
  import activityImageSrc from '@/assets/images/a.jpg';
  import CreateActivityDialog from './components/create-activity-dialog.vue';

  type SizeProps = 'mini' | 'small' | 'medium' | 'large';
  type Column = TableColumnData & { checked?: boolean };

  type ActivityStatus = 'draft' | 'published' | 'running';

  interface ActivityRecord {
    sort: number;
    activityId: string;
    activityName: string;
    activityType: string;
    activityCondition: string;
    autoReviewConfig: string;
    autoReviewEnabled: boolean;
    autoReviewOperator: string;
    autoReviewValue: number | null;
    listed: string;
    currency: string;
    claimCount: string;
    participantMembers: string;
    participantLevel: string;
    washMultiplier: string;
    specifiedGames: string;
    participantClient: string;
    activityTime: string;
    createdTime: string;
    createdBy: string;
    createdInfo: string;
    updatedTime?: string;
    updatedBy?: string;
    updatedInfo?: string;
    status: ActivityStatus;
  }

  const statusLabelMap: Record<ActivityStatus, string> = {
    draft: 'discountCenter.activityCenter.status.draft',
    published: 'discountCenter.activityCenter.status.published',
    running: 'discountCenter.activityCenter.status.running',
  };

  const statusColorMap: Record<ActivityStatus, 'gray' | 'green' | 'orange'> = {
    draft: 'gray',
    published: 'green',
    running: 'orange',
  };

  const getStatusLabelKey = (status: ActivityStatus) => statusLabelMap[status];
  const getStatusTagColor = (status: ActivityStatus) => statusColorMap[status];

  const autoReviewOperatorOptions: SelectOptionData[] = [
    { label: '<=', value: '<=' },
    { label: '<', value: '<' },
    { label: '>', value: '>' },
    { label: '>=', value: '>=' },
  ];
  const autoReviewPopoverVisibleId = ref<string | null>(null);

  const resetAutoReviewInputs = (record: ActivityRecord) => {
    record.autoReviewOperator = '';
    record.autoReviewValue = null;
  };

  const confirmAutoReview = (record: ActivityRecord) => {
    if (!record.autoReviewEnabled) {
      record.autoReviewConfig = '关闭';
      resetAutoReviewInputs(record);
      autoReviewPopoverVisibleId.value = null;
      return;
    }
    const operator = record.autoReviewOperator || '请选择';
    if (operator === '请选择' || record.autoReviewValue === null) {
      record.autoReviewConfig = '请选择';
      return;
    }
    record.autoReviewConfig = `${operator} ${record.autoReviewValue}`;
    autoReviewPopoverVisibleId.value = null;
  };

  const cancelAutoReview = (record: ActivityRecord) => {
    record.autoReviewEnabled = false;
    resetAutoReviewInputs(record);
    record.autoReviewConfig = '请选择';
    autoReviewPopoverVisibleId.value = null;
  };

  const handleAutoReviewPopoverVisibleChange = (
    visible: boolean,
    activityId: string
  ) => {
    autoReviewPopoverVisibleId.value = visible ? activityId : null;
  };

  const { t } = useI18n();

  const generateFormModel = () => {
    return {
      keywordField: 'activityName',
      keyword: '',
      activityType: '',
      client: '',
      status: '',
    };
  };

  const formModel = ref(generateFormModel());
  const tableData = ref<ActivityRecord[]>([]);
  const size = ref<SizeProps>('medium');

  const allowedSpecifiedGames = [
    'Crime Empire',
    'Mines',
    'Dice',
    'Plinko',
    'Hilo',
    'Keno',
    'Coin Flip',
    'Penalty',
    'Blackjack',
    'Poker',
    'Air Jet',
    'StarX',
    'Fortune Crash',
    'Meta Crash',
    'Tropicana',
  ];

  const formatSpecifiedGames = (selectedGames: string[]) => {
    if (!selectedGames.length) {
      return 'PC网页，移动网页，Android-APP，IOS-APP';
    }
    return selectedGames.join(' / ');
  };

  const mockActivities = (): ActivityRecord[] => [
    {
      sort: 1,
      activityId: '553',
      activityName: 'Vòng quay may mắn',
      activityType: '幸运轮盘',
      activityCondition: '无条件',
      autoReviewConfig: '请选择',
      autoReviewEnabled: false,
      autoReviewOperator: '',
      autoReviewValue: 0,
      listed: '是',
      currency: 'USDT',
      claimCount: '---',
      participantMembers: '2',
      participantLevel: '显示层级 (16)',
      washMultiplier: '3倍',
      specifiedGames: formatSpecifiedGames(['Dice']),
      participantClient: 'PC网页，移动网页，Android-APP，IOS-APP',
      activityTime: '永久',
      createdTime: '2026-08-18 18:05:58',
      createdBy: 'betatest2',
      createdInfo: 'betatest2',
      updatedInfo: 'betatest2',
      status: 'running',
    },
    {
      sort: 2,
      activityId: '551',
      activityName: 'Thưởng thẻ nạp tiền lần đầu',
      activityType: '新人彩金',
      activityCondition: '充值',
      autoReviewConfig: '请选择',
      autoReviewEnabled: false,
      autoReviewOperator: '',
      autoReviewValue: 0,
      listed: '是',
      currency: 'USDT',
      claimCount: '---',
      participantMembers: '2',
      participantLevel: '显示层级 (16)',
      washMultiplier: '---',
      specifiedGames: formatSpecifiedGames(['Mines']),
      participantClient: 'PC网页，移动网页，Android-APP，IOS-APP',
      activityTime: '2026-08-18 18:05:58',
      createdTime: '2026-08-18 18:05:58',
      createdBy: 'betatest2',
      createdInfo: 'betatest2',
      updatedInfo: 'betatest2',
      status: 'published',
    },
    {
      sort: 3,
      activityId: '386',
      activityName: 'Mùa Lì Xi',
      activityType: '红包雨',
      activityCondition: '无条件',
      autoReviewConfig: '请选择',
      autoReviewEnabled: false,
      autoReviewOperator: '',
      autoReviewValue: 0,
      listed: '是',
      currency: 'USDT',
      claimCount: '---',
      participantMembers: '2',
      participantLevel: '显示层级 (17)',
      washMultiplier: '---',
      specifiedGames: formatSpecifiedGames([
        'Tropicana',
        'Meta Crash',
        'Fortune Crash',
        'StarX',
        'Air Jet',
        'Dice',
        'Plinko',
        'Hilo',
        'Keno',
        'Coin Flip',
        'Penalty',
        'Blackjack',
        'Poker',
      ]),
      participantClient: 'PC网页，移动网页，Android-APP，IOS-APP',
      activityTime: '2026-08-18 18:05:58',
      createdTime: '2026-08-18 18:05:58',
      createdBy: 'betatest2',
      createdInfo: 'betatest2',
      updatedInfo: 'betatest2',
      status: 'running',
    },
    {
      sort: 4,
      activityId: '550',
      activityName: 'Quý cửu trừ',
      activityType: '推荐会员',
      activityCondition: '救援金',
      autoReviewConfig: '请选择',
      autoReviewEnabled: false,
      autoReviewOperator: '',
      autoReviewValue: 0,
      listed: '是',
      currency: 'USDT',
      claimCount: '---',
      participantMembers: '1',
      participantLevel: '显示层级 (16)',
      washMultiplier: '---',
      specifiedGames: formatSpecifiedGames(['Coin Flip']),
      participantClient: 'PC网页，移动网页，Android-APP，IOS-APP',
      activityTime: '2026-08-18 18:05:58',
      createdTime: '2026-08-18 18:05:58',
      createdBy: 'betatest2',
      createdInfo: 'betatest2',
      updatedInfo: 'betatest2',
      status: 'draft',
    },
    {
      sort: 5,
      activityId: '539',
      activityName: 'Mời bạn bè họ trợ',
      activityType: '助力',
      activityCondition: '无条件',
      autoReviewConfig: '请选择',
      autoReviewEnabled: false,
      autoReviewOperator: '',
      autoReviewValue: 0,
      listed: '是',
      currency: 'USDT',
      claimCount: '---',
      participantMembers: '1',
      participantLevel: '显示层级 (16)',
      washMultiplier: '---',
      specifiedGames: formatSpecifiedGames(['Air Jet', 'StarX']),
      participantClient: 'PC网页，移动网页，Android-APP，IOS-APP',
      activityTime: '2026-08-18 18:05:58',
      createdTime: '2026-08-18 18:05:58',
      createdBy: 'betatest2',
      createdInfo: 'betatest2',
      updatedInfo: 'betatest2',
      status: 'running',
    },
    {
      sort: 6,
      activityId: '533',
      activityName: 'Phần thưởng VIP',
      activityType: 'VIP奖金',
      activityCondition: '无条件',
      autoReviewConfig: '请选择',
      autoReviewEnabled: false,
      autoReviewOperator: '',
      autoReviewValue: 0,
      listed: '是',
      currency: 'USDT',
      claimCount: '---',
      participantMembers: '2',
      participantLevel: '显示层级 (16)',
      washMultiplier: '---',
      specifiedGames: formatSpecifiedGames(allowedSpecifiedGames),
      participantClient: 'PC网页，移动网页，Android-APP，IOS-APP',
      activityTime: '2026-08-18 18:05:58',
      createdTime: '2026-08-18 18:05:58',
      createdBy: 'betatest2',
      createdInfo: 'betatest2',
      updatedInfo: 'betatest2',
      status: 'published',
    },
    {
      sort: 7,
      activityId: '483',
      activityName: 'Thưởng nạp tiền lần đầu',
      activityType: '推荐会员',
      activityCondition: '充值',
      autoReviewConfig: '请选择',
      autoReviewEnabled: false,
      autoReviewOperator: '',
      autoReviewValue: 0,
      listed: '是',
      currency: 'USDT',
      claimCount: '---',
      participantMembers: '2',
      participantLevel: '显示层级 (16)',
      washMultiplier: '---',
      specifiedGames: formatSpecifiedGames([
        'PC网页',
        '移动网页',
        'Android-APP',
        'IOS-APP',
      ]),
      participantClient: 'PC网页，移动网页，Android-APP，IOS-APP',
      activityTime: '2026-08-18 18:05:58',
      createdTime: '2026-08-18 18:05:58',
      createdBy: 'betatest2',
      createdInfo: 'betatest2',
      updatedInfo: 'betatest2',
      status: 'draft',
    },
    {
      sort: 8,
      activityId: '462',
      activityName: 'Điểm danh',
      activityType: '签到',
      activityCondition: '无条件',
      autoReviewConfig: '请选择',
      autoReviewEnabled: false,
      autoReviewOperator: '',
      autoReviewValue: 0,
      listed: '是',
      currency: 'USDT',
      claimCount: '---',
      participantMembers: '2',
      participantLevel: '显示层级 (16)',
      washMultiplier: '---',
      specifiedGames: formatSpecifiedGames([
        'PC网页',
        '移动网页',
        'Android-APP',
        'IOS-APP',
      ]),
      participantClient: 'PC网页，移动网页，Android-APP，IOS-APP',
      activityTime: '2026-08-18 18:05:58',
      createdTime: '2026-08-18 18:05:58',
      createdBy: 'betatest2',
      createdInfo: 'betatest2',
      updatedInfo: 'betatest2',
      status: 'running',
    },
    {
      sort: 9,
      activityId: '454',
      activityName: 'Tất cả đại lý',
      activityType: '全民代理',
      activityCondition: '无条件',
      autoReviewConfig: '请选择',
      autoReviewEnabled: false,
      autoReviewOperator: '',
      autoReviewValue: 0,
      listed: '是',
      currency: 'USDT',
      claimCount: '---',
      participantMembers: '2',
      participantLevel: '显示层级 (16)',
      washMultiplier: '---',
      specifiedGames: formatSpecifiedGames(allowedSpecifiedGames),
      participantClient: 'PC网页，移动网页，Android-APP，IOS-APP',
      activityTime: '2026-08-18 18:05:58',
      createdTime: '2026-08-18 18:05:58',
      createdBy: 'betatest2',
      createdInfo: 'betatest2',
      updatedInfo: 'betatest2',
      status: 'published',
    },
    {
      sort: 10,
      activityId: '220',
      activityName: 'Thưởng thẻ nạp tiền lần hai',
      activityType: '推荐会员',
      activityCondition: '充值',
      autoReviewConfig: '请选择',
      autoReviewEnabled: false,
      autoReviewOperator: '',
      autoReviewValue: 0,
      listed: '是',
      currency: 'USDT',
      claimCount: '---',
      participantMembers: '1',
      participantLevel: '显示层级 (16)',
      washMultiplier: '---',
      specifiedGames: formatSpecifiedGames(['Air Jet', 'StarX']),
      participantClient: 'PC网页，移动网页，Android-APP，IOS-APP',
      activityTime: '2026-08-18 18:05:58',
      createdTime: '2026-08-18 18:05:58',
      createdBy: 'betatest2',
      createdInfo: 'betatest2',
      updatedInfo: 'betatest2',
      status: 'draft',
    },
    {
      sort: 11,
      activityId: '290',
      activityName: 'Thưởng nạp tiền tích lũy',
      activityType: '新人彩金',
      activityCondition: '充值',
      autoReviewConfig: '请选择',
      autoReviewEnabled: false,
      autoReviewOperator: '',
      autoReviewValue: 0,
      listed: '是',
      currency: 'USDT',
      claimCount: '---',
      participantMembers: '2',
      participantLevel: '显示层级 (16)',
      washMultiplier: '---',
      specifiedGames: formatSpecifiedGames([
        'Dice',
        'Plinko',
        'Hilo',
        'Keno',
        'Coin Flip',
        'Penalty',
        'Blackjack',
        'Poker',
      ]),
      participantClient: 'PC网页，移动网页，Android-APP，IOS-APP',
      activityTime: '2026-08-18 18:05:58',
      createdTime: '2026-08-18 18:05:58',
      createdBy: 'betatest2',
      createdInfo: 'betatest2',
      updatedInfo: 'betatest2',
      status: 'running',
    },
  ];

  const createDialogVisible = ref(false);

  const keywordFieldOptions = computed<SelectOptionData[]>(() => [
    {
      label: t('discountCenter.activityCenter.form.keyword.field.activityName'),
      value: 'activityName',
    },
    {
      label: t('discountCenter.activityCenter.form.keyword.field.activityId'),
      value: 'activityId',
    },
  ]);

  const activityTypeOptions = computed<SelectOptionData[]>(() => [
    {
      label: t('discountCenter.activityCenter.activityType.discount'),
      value: 'discount',
    },
    {
      label: t('discountCenter.activityCenter.activityType.coupon'),
      value: 'coupon',
    },
    {
      label: t('discountCenter.activityCenter.activityType.rebate'),
      value: 'rebate',
    },
    {
      label: t('discountCenter.activityCenter.activityType.luckyDraw'),
      value: 'luckyDraw',
    },
    {
      label: t('discountCenter.activityCenter.activityType.assist'),
      value: 'assist',
    },
    {
      label: t('discountCenter.activityCenter.activityType.blindBox'),
      value: 'blindBox',
    },
    {
      label: t('discountCenter.activityCenter.activityType.newcomerBonus'),
      value: 'newcomerBonus',
    },
    {
      label: t('discountCenter.activityCenter.activityType.nationalAgent'),
      value: 'nationalAgent',
    },
    {
      label: t('discountCenter.activityCenter.activityType.rescueFund'),
      value: 'rescueFund',
    },
    {
      label: t('discountCenter.activityCenter.activityType.luckyBet'),
      value: 'luckyBet',
    },
    {
      label: t('discountCenter.activityCenter.activityType.vipBonus'),
      value: 'vipBonus',
    },
    {
      label: t('discountCenter.activityCenter.activityType.redPacketRain'),
      value: 'redPacketRain',
    },
    {
      label: t('discountCenter.activityCenter.activityType.imageAd'),
      value: 'imageAd',
    },
    {
      label: t('discountCenter.activityCenter.activityType.signIn'),
      value: 'signIn',
    },
  ]);

  const clientOptions = computed<SelectOptionData[]>(() => [
    {
      label: t('discountCenter.activityCenter.client.web'),
      value: 'web',
    },
    {
      label: t('discountCenter.activityCenter.client.mobile'),
      value: 'mobile',
    },
    {
      label: t('discountCenter.activityCenter.client.android'),
      value: 'android',
    },
    {
      label: t('discountCenter.activityCenter.client.ios'),
      value: 'ios',
    },
  ]);

  const statusOptions = computed<SelectOptionData[]>(() => [
    {
      label: t('discountCenter.activityCenter.status.all'),
      value: '',
    },
    {
      label: t('discountCenter.activityCenter.status.draft'),
      value: 'draft',
    },
    {
      label: t('discountCenter.activityCenter.status.published'),
      value: 'published',
    },
    {
      label: t('discountCenter.activityCenter.status.running'),
      value: 'running',
    },
  ]);

  const densityList = computed(() => [
    {
      name: t('discountCenter.activityCenter.size.mini'),
      value: 'mini',
    },
    {
      name: t('discountCenter.activityCenter.size.small'),
      value: 'small',
    },
    {
      name: t('discountCenter.activityCenter.size.medium'),
      value: 'medium',
    },
    {
      name: t('discountCenter.activityCenter.size.large'),
      value: 'large',
    },
  ]);

  const baseColumns = computed<TableColumnData[]>(() => [
    {
      title: t('discountCenter.activityCenter.columns.sort'),
      dataIndex: 'sort',
      width: 70,
      align: 'center',
      fixed: 'left',
    },
    {
      title: t('discountCenter.activityCenter.columns.activityId'),
      dataIndex: 'activityId',
      width: 100,
      align: 'center',
    },
    {
      title: t('discountCenter.activityCenter.columns.activityName'),
      dataIndex: 'activityName',
      width: 140,
      align: 'center',
    },
    {
      title: '活动类型/活动条件',
      dataIndex: 'activityType',
      slotName: 'activityMeta',
      width: 200,
      align: 'center',
    },
    {
      title: t('discountCenter.activityCenter.columns.autoReviewConfig'),
      dataIndex: 'autoReviewConfig',
      slotName: 'autoConfig',
      width: 390,
      align: 'center',
    },
    {
      title: t('discountCenter.activityCenter.columns.listed'),
      dataIndex: 'listed',
      width: 120,
      align: 'center',
    },
    {
      title: t('discountCenter.activityCenter.columns.currency'),
      dataIndex: 'currency',
      width: 100,
      align: 'center',
    },
    {
      title: t('discountCenter.activityCenter.columns.claimCount'),
      dataIndex: 'claimCount',
      width: 120,
      align: 'center',
    },
    {
      title: t('discountCenter.activityCenter.columns.participantMembers'),
      dataIndex: 'participantMembers',
      width: 120,
      align: 'center',
    },
    {
      title: '参与层级',
      dataIndex: 'participantLevel',
      width: 140,
      align: 'center',
    },
    {
      title: t('discountCenter.activityCenter.columns.washMultiplier'),
      dataIndex: 'washMultiplier',
      width: 120,
      align: 'center',
    },
    {
      title: '指定游戏',
      dataIndex: 'specifiedGames',
      slotName: 'specifiedGames',
      width: 160,
      align: 'center',
    },
    {
      title: '参与客户端',
      dataIndex: 'participantClient',
      slotName: 'participantClient',
      width: 180,
      align: 'center',
    },
    {
      title: t('discountCenter.activityCenter.columns.activityTime'),
      dataIndex: 'activityTime',
      slotName: 'activityTime',
      width: 180,
      align: 'center',
    },
    {
      title: '创建时间/创建人',
      dataIndex: 'createdInfo',
      slotName: 'activeScope',
      width: 170,
      align: 'center',
    },
    {
      title: t('discountCenter.activityCenter.columns.updatedInfo'),
      dataIndex: 'updatedInfo',
      slotName: 'updatedInfo',
      width: 170,
      align: 'center',
    },
    {
      title: t('discountCenter.activityCenter.columns.status'),
      dataIndex: 'status',
      slotName: 'status',
      width: 120,
      align: 'center',
      fixed: 'right',
    },
    {
      title: t('discountCenter.activityCenter.columns.operations'),
      dataIndex: 'operations',
      slotName: 'operations',
      width: 100,
      align: 'center',
      fixed: 'right',
    },
  ]);

  const cloneColumns = ref<Column[]>([]);
  const showColumns = ref<Column[]>([]);

  const pagination = reactive({
    total: 0,
    current: 1,
    pageSize: 10,
    size: 'small' as const,
    showTotal: true,
    showPageSize: true,
    pageSizeProps: {
      popupContainer: 'body',
    },
    pageSizeOptions: [10, 15, 30, 50],
  });

  const paginatedTableData = computed(() => {
    const start = (pagination.current - 1) * pagination.pageSize;
    const end = start + pagination.pageSize;
    return tableData.value.slice(start, end);
  });

  const handlePageChange = (current: number) => {
    pagination.current = current;
  };

  const handlePageSizeChange = (pageSize: number) => {
    pagination.pageSize = pageSize;
    pagination.current = 1;
  };

  tableData.value = mockActivities().map((item) => ({
    ...item,
    updatedTime: item.updatedTime ?? item.createdTime,
    updatedBy: item.updatedBy ?? item.createdBy,
  }));
  pagination.total = tableData.value.length;

  const search = () => {
    pagination.current = 1;
  };

  const reset = () => {
    formModel.value = generateFormModel();
    pagination.current = 1;
  };

  const handleSelectDensity = (
    val: string | number | Record<string, any> | undefined
  ) => {
    size.value = val as SizeProps;
  };

  const handleChange = (
    checked: boolean | (string | boolean | number)[],
    _column: Column,
    index: number
  ) => {
    showColumns.value[index].checked = Boolean(checked);
    cloneColumns.value = showColumns.value.filter((item) => item.checked);
  };

  const exchangeArray = <T extends Array<any>>(
    array: T,
    beforeIdx: number,
    newIdx: number,
    isDeep = false
  ): T => {
    const newArray = isDeep ? cloneDeep(array) : array;
    if (beforeIdx > -1 && newIdx > -1) {
      newArray.splice(
        beforeIdx,
        1,
        newArray.splice(newIdx, 1, newArray[beforeIdx]).pop()
      );
    }
    return newArray;
  };

  const popupVisibleChange = (val: boolean) => {
    if (val) {
      nextTick(() => {
        const el = document.getElementById('tableSetting') as HTMLElement;
        const sortable = new Sortable(el, {
          onEnd(e: any) {
            const { oldIndex, newIndex } = e;
            exchangeArray(cloneColumns.value, oldIndex, newIndex);
            exchangeArray(showColumns.value, oldIndex, newIndex);
          },
        });
        return sortable;
      });
    }
  };

  watch(
    () => baseColumns.value,
    (val) => {
      cloneColumns.value = cloneDeep(val);
      cloneColumns.value.forEach((item) => {
        item.checked = true;
      });
      showColumns.value = cloneDeep(cloneColumns.value);
    },
    { deep: true, immediate: true }
  );
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

  .action-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    border-radius: 4px;
    cursor: pointer;
    transition: background-color 0.2s ease;

    &:hover {
      background-color: var(--color-fill-2);
    }
  }

  .status-tag {
    min-width: 52px;
    justify-content: center;
  }

  .setting {
    display: flex;
    align-items: center;
    width: 220px;

    .title {
      margin-left: 12px;
      cursor: pointer;
    }
  }

  .active {
    color: #0960bd;
    background-color: #e3f4fc;
  }

  .activity-table-wrap {
    width: 100%;
  }

  .activity-table {
    width: 100%;

    :deep(.arco-table) {
      width: 100%;
    }

    :deep(.arco-table-container) {
      width: 100%;
    }

    :deep(.arco-table-element) {
      width: 100%;
    }

    :deep(.arco-table-td) {
      padding: 0;
      vertical-align: top;
    }

    :deep(.arco-table-tr) {
      background-color: var(--color-bg-2);
    }

    :deep(.arco-table-tr + .arco-table-tr .arco-table-td) {
      border-top: 12px solid var(--color-fill-1);
    }
  }

  .activity-list-row {
    overflow: hidden;
    background-color: var(--color-bg-2);
  }

  .activity-list-header {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 16px;
    min-height: 38px;
    padding: 8px 16px;
    background-color: var(--color-fill-1);
    border-bottom: 1px solid var(--color-border-2);
    color: var(--color-text-2);
    font-size: 13px;
    line-height: 1.6;
  }

  .activity-list-header span {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    white-space: nowrap;
  }

  .show-tag,
  .review-tag {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 34px;
    padding: 2px 10px;
    border-radius: 999px;
    background: rgba(16, 185, 129, 0.12);
    color: #0f766e;
    border: 1px solid rgba(16, 185, 129, 0.22);
    font-size: 13px;
  }

  .review-tag {
    background: rgba(59, 130, 246, 0.08);
    color: #1d4ed8;
    border-color: rgba(96, 165, 250, 0.22);
  }

  .review-config-trigger {
    cursor: pointer;
  }

  .auto-review-form {
    min-width: 240px;
  }

  .activity-list-body {
    display: grid;
    grid-template-columns:
      minmax(280px, 1.2fr) minmax(210px, 1fr) minmax(210px, 1fr)
      minmax(270px, 1.15fr) minmax(150px, 0.65fr);
    gap: 0;
    min-height: 132px;
    background-color: var(--color-bg-2);
  }

  .info-block {
    position: relative;
    padding: 14px 16px;
    min-height: 132px;
    background-color: var(--color-bg-2);
  }

  .info-block:not(:last-child) {
    border-right: 1px solid var(--color-border-2);
  }

  .block-title {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 10px;
    color: var(--color-text-2);
    font-size: 14px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .block-mark {
    display: inline-block;
    width: 10px;
    height: 10px;
    border-radius: 3px;
    background: linear-gradient(135deg, #34d399, #10b981);
    box-shadow: 0 0 12px rgba(52, 211, 153, 0.45);
  }

  .block-mark.orange {
    background: linear-gradient(135deg, #fbbf24, #f59e0b);
    box-shadow: 0 0 12px rgba(251, 191, 36, 0.45);
  }

  .block-mark.blue {
    background: linear-gradient(135deg, #60a5fa, #2563eb);
    box-shadow: 0 0 12px rgba(96, 165, 250, 0.45);
  }

  .block-mark.green {
    background: linear-gradient(135deg, #4ade80, #16a34a);
    box-shadow: 0 0 12px rgba(74, 222, 128, 0.45);
  }

  .block-mark.purple {
    background: linear-gradient(135deg, #c084fc, #8b5cf6);
    box-shadow: 0 0 12px rgba(192, 132, 252, 0.45);
  }

  .block-content {
    display: flex;
    flex-direction: column;
    gap: 4px;
    line-height: 1.6;
    color: var(--color-text-2);
  }

  :deep(.activity-image) {
    flex: 0 0 auto;
    width: 72px;
    height: 72px;
    overflow: hidden;
    border-radius: var(--border-radius-small);

    .arco-image-img {
      width: 72px;
      height: 72px;
      object-fit: cover;
    }
  }

  .primary-name {
    font-size: 16px;
    font-weight: 600;
    color: var(--color-text-1);
    line-height: 1.4;
  }

  .sub-value {
    color: var(--color-text-2);
    font-size: 13px;
  }

  .sub-value.muted {
    color: var(--color-text-3);
  }

  .compact-list {
    font-size: 14px;
    color: var(--color-text-2);
  }

  .compact-list strong {
    font-weight: 600;
    color: var(--color-text-1);
  }

  .status-block {
    display: flex;
    flex-direction: column;
  }

  .status-wrap {
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    height: 100%;
  }

  .status-wrap :deep(.arco-tag) {
    align-self: flex-start;
  }

  .status-description {
    color: var(--color-text-3);
    font-size: 13px;
  }

  .activity-list-footer {
    display: flex;
    align-items: center;
    gap: 20px;
    min-height: 46px;
    padding: 8px 16px;
    border-top: 1px solid var(--color-border-2);
    color: var(--color-text-2);
    font-size: 13px;
  }

  .row-actions {
    display: flex;
    flex-wrap: wrap;
    margin-top: 0;
    margin-left: auto;
  }

  .row-actions :deep(.arco-btn) {
    min-width: 52px;
  }

  .table-pagination {
    display: flex;
    justify-content: flex-end;
    padding-top: 16px;
  }

  :deep(.arco-pagination-options .arco-select) {
    min-width: 100px;
  }

  .create-activity-modal {
    :deep(.arco-modal-header) {
      border-bottom: 1px solid rgb(var(--gray-3));
      margin-bottom: 0;
    }

    :deep(.arco-modal-body) {
      padding: 0;
    }
  }

  .create-activity-dialog {
    display: flex;
    min-height: 600px;
    background: #fff;
  }

  .dialog-left {
    width: 430px;
    padding: 16px 20px 12px;
    border-right: 1px solid rgb(var(--gray-3));
    background: rgb(var(--gray-1));
  }

  .section-label {
    margin-bottom: 12px;
    font-size: 14px;
    font-weight: 600;
    color: rgb(var(--gray-10));
  }

  .activity-tabs {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 16px;

    .tab-btn {
      min-width: 84px;
      font-size: 12px;
    }
  }

  .language-wrap {
    margin-top: 6px;
  }

  .language-row {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 12px;

    .locale-select {
      width: 170px;
    }

    .arco-input {
      flex: 1;
    }

    .remove-btn {
      flex-shrink: 0;
      width: 26px;
      height: 26px;
      border: none;
      border-radius: 50%;
      background: rgb(var(--red-5));
      color: #fff;
    }
  }

  .add-language {
    width: 100%;
    margin-top: 4px;
    border: 1px solid rgb(var(--arcoblue-5));
    color: rgb(var(--arcoblue-6));
  }

  .switch-row,
  .time-row,
  .status-row {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-top: 18px;
    color: rgb(var(--gray-8));
  }

  .switch-row {
    justify-content: space-between;
  }

  .time-row {
    .date-separator {
      color: rgb(var(--gray-6));
    }
  }

  .dialog-right {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
    padding: 18px 20px 16px;
    background: #fff;
  }

  .editor-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 8px 12px;
    border: 1px solid rgb(var(--gray-3));
    border-radius: 8px 8px 0 0;
    background: rgb(var(--gray-1));
  }

  .editor-tools {
    display: flex;
    align-items: center;
    gap: 12px;
    color: rgb(var(--gray-8));
    font-size: 12px;
  }

  .editor-body {
    flex: 1;
    border: 1px solid rgb(var(--gray-3));
    border-top: none;
    border-radius: 0 0 8px 8px;
    background: #fff;
    padding: 20px 18px;
    font-size: 14px;
    line-height: 2;
    color: rgb(var(--gray-8));
  }
</style>
