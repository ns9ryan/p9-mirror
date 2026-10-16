<template>
  <a-modal
    v-model:visible="visible"
    width="1960px"
    unmount-on-close
    :footer="false"
    class="activity-create-modal"
    :maskClosable="false"
  >
    <template #title>
      <div class="dialog-header-title">新增活动</div>
    </template>
    <!-- 使用 flex 固定 Label 宽度 120px，彻底解决文字垂直挤压变形问题 -->
    <a-form
      :model="form"
      layout="horizontal"
      :label-col-props="{ flex: '120px' }"
      :wrapper-col-props="{ flex: '1' }"
      label-align="right"
      class="custom-form"
    >
      <a-row :gutter="32" class="activity-form-grid">
        <!-- 左侧表单区域 -->
        <a-col :span="12" class="form-panel-left">
          <section class="accordion-section">
          <div
            class="section-title section-title--collapsible"
            role="button"
            tabindex="0"
            :class="{ active: isSectionExpanded('basic') }"
            :aria-expanded="isSectionExpanded('basic')"
            @click="toggleSection('basic')"
            @keydown.enter="toggleSection('basic')"
            @keydown.space.prevent="toggleSection('basic')"
          >
            <span>1. 基本信息</span>
            <icon-down
              class="section-title-icon"
              :class="{ collapsed: !isSectionExpanded('basic') }"
            />
          </div>
          <div v-show="isSectionExpanded('basic')" class="section-card">
            <div class="section-body">
              <!-- 活动类型：按钮网格 -->
              <a-form-item label="活动类型" required>
            <div class="type-grid">
              <div
                v-for="item in typeOptions"
                :key="item"
                class="type-item"
                :class="{ active: form.type === item }"
                @click="form.type = item"
              >
                {{ item }}
              </div>
            </div>
          </a-form-item>

          <!-- 货币 -->
          <a-form-item label="货币">
            <span class="currency-text">USDT</span>
          </a-form-item>

          <!-- 活动名称（多语言组） -->
          <a-form-item label="活动名称" required>
            <div class="dynamic-group">
              <div
                v-for="(item, index) in form.names"
                :key="index"
                class="dynamic-row"
              >
                <a-select v-model="item.lang" style="width: 150px">
                  <a-option value="en">English</a-option>
                  <a-option value="en-US">English (United...)</a-option>
                  <a-option value="zh-TW">繁體中文</a-option>
                  <a-option value="pt-BR">Português (Br...</a-option>
                  <a-option value="vi">Tiếng Việt</a-option>
                </a-select>
                <a-input
                  v-model="item.value"
                  placeholder="请输入"
                  style="flex: 1"
                />
                <a-button
                  type="primary"
                  status="danger"
                  shape="circle"
                  size="small"
                  class="remove-btn"
                  @click="removeName(index)"
                >
                  <icon-minus />
                </a-button>
              </div>
              <a-button
                type="primary"
                size="small"
                class="add-btn"
                @click="addName"
              >
                <icon-plus />
              </a-button>
            </div>
              </a-form-item>
            </div>
          </div>
          </section>

          <section class="accordion-section">
          <div
            class="section-title section-title--collapsible"
            role="button"
            tabindex="0"
            :class="{ active: isSectionExpanded('settings') }"
            :aria-expanded="isSectionExpanded('settings')"
            @click="toggleSection('settings')"
            @keydown.enter="toggleSection('settings')"
            @keydown.space.prevent="toggleSection('settings')"
          >
            <span>2. 活动设置</span>
            <icon-down
              class="section-title-icon"
              :class="{ collapsed: !isSectionExpanded('settings') }"
            />
          </div>
          <div v-show="isSectionExpanded('settings')" class="section-card">
            <div class="section-body">
              <!-- 是否在活动列表内展示 -->
              <a-form-item label="是否在活动列表内展示">
            <a-radio-group v-model="form.showInList">
              <a-radio :value="true">是</a-radio>
              <a-radio :value="false">否</a-radio>
            </a-radio-group>
          </a-form-item>

          <!-- 活动时间 -->
          <a-form-item label="活动时间" required>
            <div class="time-group">
              <div class="time-row">
                <a-radio
                  :model-value="form.timeType === 'range'"
                  @change="form.timeType = 'range'"
                />
                <a-range-picker
                  v-model="form.timeRange"
                  show-time
                  format="YYYY-MM-DD HH:mm:ss"
                  :disabled="form.timeType !== 'range'"
                  style="width: 260px"
                />
              </div>
              <div class="time-row">
                <a-radio
                  :model-value="form.timeType === 'permanent'"
                  @change="form.timeType = 'permanent'"
                >
                  永久
                </a-radio>
              </div>
            </div>
          </a-form-item>

          <!-- 指定游戏 -->
          <a-form-item label="指定游戏">
            <div class="specified-game-section">
              <div class="specified-game-header">
                <div>
                  <div class="specified-game-title">按平台定向选择</div>
                  <div class="specified-game-desc">
                    开启后可按分类筛选平台，支持多选并实时查看已选数量。
                  </div>
                </div>
                <a-switch v-model="form.specifiedGame" />
              </div>

              <div v-if="form.specifiedGame" class="specified-game-panel">
                <div class="specified-game-tabs">
                  <div
                    v-for="item in gameCategoryOptions"
                    :key="item.value"
                    class="specified-game-tab"
                    :class="{ active: activeGameCategory === item.value }"
                    @click="activeGameCategory = item.value"
                  >
                    {{ item.label }}
                  </div>
                </div>

                <div class="specified-game-toolbar">
                  <a-checkbox
                    :model-value="currentCategoryAllSelected"
                    :indeterminate="currentCategoryIndeterminate"
                    @update:model-value="toggleCurrentCategorySelection"
                  >
                    全部平台
                  </a-checkbox>
                  <div class="specified-game-toolbar-hint">
                    当前分类已选 {{ currentCategorySummary.gameCount }} 个游戏
                  </div>
                </div>

                <div class="specified-game-list">
                  <div
                    v-for="platform in activeGamePlatforms"
                    :key="platform.name"
                    class="specified-game-card"
                    :class="{
                      selected: hasSelectedGames(platform),
                      expanded: platform.expanded,
                    }"
                  >
                    <div
                      class="specified-game-card-header"
                      @click="togglePlatformExpanded(platform)"
                    >
                      <a-checkbox
                        :model-value="isPlatformAllSelected(platform)"
                        :indeterminate="isPlatformIndeterminate(platform)"
                        @update:model-value="
                          setPlatformSelected(platform, $event)
                        "
                        @click.stop
                      />

                      <div class="specified-game-card-body">
                        <div class="specified-game-card-top">
                          <div class="specified-game-card-name">
                            {{ platform.name }}
                            <span class="specified-game-count-hint">
                              已选 {{ getPlatformSelectedCount(platform) }} 个游戏
                            </span>
                          </div>
                        </div>
                      </div>

                      <span
                        class="specified-game-card-arrow"
                        :class="{ expanded: platform.expanded }"
                      >
                        <icon-down v-if="platform.expanded" />
                        <icon-right v-else />
                      </span>
                    </div>

                    <div
                      v-if="platform.expanded"
                      class="specified-game-subpanel"
                    >
                      <div class="specified-game-subtoolbar">
                        <a-input
                          v-model="platform.keyword"
                          allow-clear
                          size="small"
                          placeholder="请输入游戏名称"
                          class="specified-game-search"
                          @click.stop
                        />
                        <div class="specified-game-subtoolbar-actions">
                          <a-button
                            type="text"
                            size="mini"
                            @click.stop="togglePlatformSelection(platform)"
                          >
                            {{
                              isPlatformAllSelected(platform)
                                ? '取消全选'
                                : '全选'
                            }}
                          </a-button>
                          <a-button
                            type="text"
                            size="mini"
                            @click.stop="clearPlatformSelection(platform)"
                          >
                            清空
                          </a-button>
                        </div>
                      </div>

                      <div
                        v-if="filterPlatformGames(platform).length"
                        class="specified-game-options"
                      >
                        <label
                          v-for="game in filterPlatformGames(platform)"
                          :key="game.name"
                          class="specified-game-option"
                        >
                          <a-checkbox v-model="game.selected" @click.stop />
                          <span class="specified-game-option-name">{{
                            game.name
                          }}</span>
                        </label>
                      </div>
                      <div v-else class="specified-game-empty"
                        >暂无匹配游戏</div
                      >
                    </div>
                  </div>
                </div>

                <div class="specified-game-summary">
                  已选 {{ selectedGameSummary.platformCount }} 个平台，
                  {{ selectedGameSummary.gameCount }} 个游戏
                </div>
              </div>
            </div>
          </a-form-item>

          <!-- 充值类型 -->
          <a-form-item
            v-if="showField('rechargeType')"
            label="充值类型"
            required
          >
            <a-radio-group v-model="form.rechargeType">
              <a-radio value="首充奖励">首充奖励</a-radio>
              <a-radio value="次存奖励">次存奖励</a-radio>
              <a-radio value="累计奖励">累计奖励</a-radio>
              <a-radio value="单笔奖励">单笔奖励</a-radio>
            </a-radio-group>
          </a-form-item>

          <!-- 循环模式 -->
          <a-form-item v-if="showField('loopMode')" label="循环模式" required>
            <a-radio-group v-model="form.loopMode" class="wrap-radio-group">
              <a-radio
                value="不指定"
                :disabled="isFieldOptionDisabled('loopMode', '不指定')"
                >不指定</a-radio
              >
              <a-radio
                value="不循环"
                :disabled="isFieldOptionDisabled('loopMode', '不循环')"
                >不循环</a-radio
              >
              <a-radio
                value="每日循环"
                :disabled="isFieldOptionDisabled('loopMode', '每日循环')"
                >每日循环</a-radio
              >
              <a-radio
                value="每周循环"
                :disabled="isFieldOptionDisabled('loopMode', '每周循环')"
                >每周循环</a-radio
              >
              <a-radio
                value="每月循环"
                :disabled="isFieldOptionDisabled('loopMode', '每月循环')"
                >每月循环</a-radio
              >
              <a-radio
                value="无限循环"
                :disabled="isFieldOptionDisabled('loopMode', '无限循环')"
                >无限循环</a-radio
              >
            </a-radio-group>
          </a-form-item>
          <!-- 参与模式 -->
          <a-form-item label="参与模式" required>
            <a-radio-group
              v-model="form.participationMode"
              class="wrap-radio-group"
            >
              <a-radio
                value="默认参与"
                :disabled="
                  isFieldOptionDisabled('participationMode', '默认参与')
                "
                >默认参与</a-radio
              >
              <a-radio
                value="手动参与"
                :disabled="
                  isFieldOptionDisabled('participationMode', '手动参与')
                "
                >手动参与</a-radio
              >
            </a-radio-group>
          </a-form-item>
          <!-- 派发奖励模式 -->
          <a-form-item label="派发奖励模式" required>
            <a-radio-group
              v-model="form.distributionMode"
              class="wrap-radio-group"
            >
              <a-radio
                value="手动领取(过期作废)"
                :disabled="
                  isFieldOptionDisabled(
                    'distributionMode',
                    '手动领取(过期作废)'
                  )
                "
                >手动领取(过期作废)</a-radio
              >
              <a-radio
                value="手动领取(过期自动领取)"
                :disabled="
                  isFieldOptionDisabled(
                    'distributionMode',
                    '手动领取(过期自动领取)'
                  )
                "
                >手动领取(过期自动领取)</a-radio
              >
              <a-radio
                value="自动派发"
                :disabled="
                  isFieldOptionDisabled('distributionMode', '自动派发')
                "
                >自动派发</a-radio
              >
            </a-radio-group>
          </a-form-item>
          <!-- 派发奖励时间 -->
          <a-form-item label="派发奖励时间" required>
            <a-radio-group
              v-model="form.distributionTime"
              class="wrap-radio-group"
            >
              <a-radio
                value="手动领取(过期作废)"
                :disabled="
                  isFieldOptionDisabled(
                    'distributionTime',
                    '手动领取(过期作废)'
                  )
                "
                >手动领取(过期作废)</a-radio
              >
              <a-radio
                value="手动领取(过期自动领取)"
                :disabled="
                  isFieldOptionDisabled(
                    'distributionTime',
                    '手动领取(过期自动领取)'
                  )
                "
                >手动领取(过期自动领取)</a-radio
              >
              <a-radio
                value="实时(影响留存)"
                :disabled="
                  isFieldOptionDisabled('distributionTime', '实时(影响留存)')
                "
                >实时(影响留存)</a-radio
              >
            </a-radio-group>
          </a-form-item>

          <!-- 会员等级 -->
          <a-form-item label="会员等级">
            <div class="checkbox-stack">
              <a-checkbox-group v-model="form.memberLevels">
                <div class="checkbox-row full-row" @click="toggleMemberLevels">
                  <a-checkbox value="全部">全部</a-checkbox>
                </div>
                <div class="checkbox-row">
                  <a-checkbox value="默认会员">默认会员</a-checkbox>
                  <a-checkbox value="初级会员">初级会员</a-checkbox>
                  <a-checkbox value="中级会员">中级会员</a-checkbox>
                  <a-checkbox value="高级会员">高级会员</a-checkbox>
                  <a-checkbox value="超级会员">超级会员</a-checkbox>
                </div>
                <div class="checkbox-row">
                  <a-checkbox value="钻石会员">钻石会员</a-checkbox>
                  <a-checkbox value="铂金会员">铂金会员</a-checkbox>
                  <a-checkbox value="黑金会员">黑金会员</a-checkbox>
                  <a-checkbox value="王者会员">王者会员</a-checkbox>
                  <a-checkbox value="荣耀会员">荣耀会员</a-checkbox>
                </div>
                <div class="checkbox-row">
                  <a-checkbox value="贵宾会员">贵宾会员</a-checkbox>
                  <a-checkbox value="VIP 1">VIP 1</a-checkbox>
                  <a-checkbox value="VIP 2">VIP 2</a-checkbox>
                  <a-checkbox value="VIP 3">VIP 3</a-checkbox>
                </div>
              </a-checkbox-group>
            </div>
              </a-form-item>
            </div>
          </div>
          </section>

          <section class="accordion-section">
          <div
            class="section-title section-title--collapsible"
            role="button"
            tabindex="0"
            :class="{ active: isSectionExpanded('rewards') }"
            :aria-expanded="isSectionExpanded('rewards')"
            @click="toggleSection('rewards')"
            @keydown.enter="toggleSection('rewards')"
            @keydown.space.prevent="toggleSection('rewards')"
          >
            <span>3. 奖励与参与</span>
            <icon-down
              class="section-title-icon"
              :class="{ collapsed: !isSectionExpanded('rewards') }"
            />
          </div>
          <div v-show="isSectionExpanded('rewards')" class="section-card">
            <div class="section-body">
              <!-- 奖金模式 -->
              <a-form-item label="奖金模式" required>
            <a-radio-group v-model="form.bonusMode">
              <a-radio value="fixed">固定金额</a-radio>
              <a-radio value="ratio">比例金额</a-radio>
            </a-radio-group>
          </a-form-item>

          <!-- 出款限制 -->
          <a-form-item label="出款限制">
            <div class="inline-input-group">
              <a-input
                v-model="form.withdrawMultiple"
                placeholder="请输入洗码倍数（输0代表不需要洗码）"
                style="width: 310px"
              />
              <span class="inline-helper">洗码倍数</span>
            </div>
          </a-form-item>

          <!-- 赠送方式 -->
          <a-form-item label="赠送方式">
            <a-radio-group v-model="form.giftMode">
              <a-radio value="part">赠送部分</a-radio>
              <a-radio value="all">赠送全部</a-radio>
            </a-radio-group>
          </a-form-item>

          <!-- 参与客户端 -->
          <a-form-item label="参与客户端">
            <a-checkbox-group v-model="form.clientList">
              <div class="checkbox-row full-row" @click="toggleClientList">
                <a-checkbox value="全部">全部</a-checkbox>
              </div>
              <div class="checkbox-row">
                <a-checkbox value="PC网页">PC网页</a-checkbox>
                <a-checkbox value="移动网页">移动网页</a-checkbox>
                <a-checkbox value="Android-APP">Android-APP</a-checkbox>
                <a-checkbox value="IOS-APP">IOS-APP</a-checkbox>
              </div>
            </a-checkbox-group>
              </a-form-item>

              <!-- 参与活动限制 -->
              <a-form-item label="参与活动限制">
            <a-checkbox-group v-model="form.activityRestrictions">
              <div class="checkbox-row">
                <a-checkbox value="同设备账号参与一次"
                  >同设备账号参与一次</a-checkbox
                >
                <a-checkbox value="同浏览器账号参与一次"
                  >同浏览器账号参与一次</a-checkbox
                >
              </div>
            </a-checkbox-group>
              </a-form-item>
            </div>
          </div>
          </section>
        </a-col>

        <!-- 右侧表单区域 -->
        <a-col :span="12" class="form-panel-right">
          <section class="accordion-section">
            <div
              class="section-title section-title--collapsible"
              role="button"
              tabindex="0"
              :class="{ active: isSectionExpanded('rules') }"
              :aria-expanded="isSectionExpanded('rules')"
              @click="toggleSection('rules')"
              @keydown.enter="toggleSection('rules')"
              @keydown.space.prevent="toggleSection('rules')"
            >
              <span>4. 活动规则</span>
              <icon-down
                class="section-title-icon"
                :class="{ collapsed: !isSectionExpanded('rules') }"
              />
            </div>
            <div v-show="isSectionExpanded('rules')" class="section-card">
              <div class="section-body">
                <!-- 活动图片 -->
                <a-form-item label="活动图片" required>
              <div class="upload-container">
                <div class="upload-tip"
                  >上传尺寸 668px * 250px, JPG、JPEG、PNG、GIF</div
                >
                <a-upload action="/" list-type="picture-card" :limit="1" />
              </div>
            </a-form-item>

            <!-- 活动规则（多语言富文本组） -->
            <a-form-item label="活动规则" required>
              <div class="dynamic-group">
                <div
                  v-for="(rule, index) in form.rules"
                  :key="index"
                  class="rule-block"
                >
                  <a-select
                    v-model="rule.lang"
                    style="width: 150px; margin-bottom: 8px"
                  >
                    <a-option value="en">English</a-option>
                    <a-option value="en-US">English (United...)</a-option>
                    <a-option value="zh-TW">繁體中文</a-option>
                    <a-option value="pt-BR">Português (Br...)</a-option>
                    <a-option value="vi">Tiếng Việt</a-option>
                  </a-select>

                  <div class="editor-wrapper">
                    <QuillEditor
                      v-model:content="rule.content"
                      theme="snow"
                      :toolbar="quillToolbar"
                      content-type="html"
                      class="quill-editor"
                    />

                    <a-button
                      type="primary"
                      status="danger"
                      shape="circle"
                      size="small"
                      class="remove-rule-btn"
                      @click="removeRule(index)"
                    >
                      <icon-minus />
                    </a-button>
                  </div>
                </div>

                <a-button
                  type="primary"
                  size="small"
                  class="add-btn"
                  @click="addRule"
                >
                  <icon-plus />
                </a-button>
              </div>
                </a-form-item>
              </div>
            </div>
          </section>
        </a-col>
      </a-row>
    </a-form>

    <div class="dialog-footer-actions">
      <a-button class="cancel-btn" @click="closeDialog">取消</a-button>
      <a-button type="primary" class="save-draft-btn">保存草稿</a-button>
    </div>
  </a-modal>
</template>

<script setup lang="ts">
  import { ref, reactive, computed, watch } from 'vue';
  import {
    IconPlus,
    IconMinus,
    IconClose,
    IconDown,
    IconRight,
  } from '@arco-design/web-vue/es/icon';
  import { QuillEditor } from '@vueup/vue-quill';
  import '@vueup/vue-quill/dist/vue-quill.snow.css';

  const visible = ref(true);

  type AccordionSection = 'basic' | 'settings' | 'rules' | 'rewards';

  const expandedSections = reactive<Record<AccordionSection, boolean>>({
    basic: true,
    settings: true,
    rules: true,
    rewards: true,
  });

  const isSectionExpanded = (section: AccordionSection) =>
    expandedSections[section];

  const toggleSection = (section: AccordionSection) => {
    expandedSections[section] = !expandedSections[section];
  };

  const typeOptions = [
    '充值',
    '洗码',
    '推荐会员',
    '幸运轮盘',
    '助力',
    '盲盒',
    '新人彩金',
    '全民代理',
    '救援金',
    '幸运注单',
    'VIP奖金',
    '红包雨',
    '图文(广告)',
    '签到',
  ];

  const memberLevelOptions = [
    '默认会员',
    '初级会员',
    '中级会员',
    '高级会员',
    '超级会员',
    '钻石会员',
    '铂金会员',
    '黑金会员',
    '王者会员',
    '荣耀会员',
    '贵宾会员',
    'VIP 1',
    'VIP 2',
    'VIP 3',
  ];

  const clientOptions = ['PC网页', '移动网页', 'Android-APP', 'IOS-APP'];

  const gameCategoryOptions = [
    { label: '视讯', value: '视讯' },
    { label: '体育', value: '体育' },
    { label: '电子', value: '电子' },
    { label: '彩票', value: '彩票' },
    { label: '棋牌', value: '棋牌' },
    { label: '捕鱼', value: '捕鱼' },
    { label: '其他', value: '其他' },
  ];

  type GameItem = {
    name: string;
    selected: boolean;
  };

  type GamePlatformItem = {
    name: string;
    description?: string;
    expanded: boolean;
    keyword: string;
    games: GameItem[];
  };

  const eeaiVideoGames = [
    'GirlGroup',
    '10-Minute Lottery',
    'Hourly Lottery',
    'Lace-2',
    'Golden Lily',
    'JiSoo Roulette',
    'Baccarat TEST-4',
    'Baccarat TEST-3',
    'Baccarat TEST-2',
    'Baccarat TEST-1',
    'Goal',
    'Space Crash',
    '10-Minute Lottery',
    'Sicbo',
    'Thunder',
    'LUCKOAI',
  ];

  const defaultVideoPlatforms: GamePlatformItem[] = [
    {
      name: 'EEAI-真人',
      expanded: true,
      keyword: '',
      games: eeaiVideoGames.map((name) => ({
        name,
        selected: false,
      })),
    },
    {
      name: 'DBLIVE-真人',
      expanded: false,
      keyword: '',
      games: [
        { name: 'DB 百家乐', selected: false },
        { name: 'DB 轮盘', selected: false },
        { name: 'DB 骰宝', selected: false },
        { name: 'DB 龙虎', selected: false },
        { name: 'DB 牛牛', selected: false },
      ],
    },
    {
      name: 'Amusnet-真人',
      expanded: false,
      keyword: '',
      games: [
        { name: 'Amusnet Roulette', selected: false },
        { name: 'Amusnet Blackjack', selected: false },
        { name: 'Amusnet Baccarat', selected: false },
        { name: 'Amusnet Dynamic Roulette', selected: false },
      ],
    },
    {
      name: 'Evolution Live-真人',
      expanded: false,
      keyword: '',
      games: [
        { name: 'Crazy Time', selected: false },
        { name: 'Lightning Roulette', selected: false },
        { name: 'Monopoly Live', selected: false },
        { name: 'Speed Baccarat', selected: false },
        { name: 'Mega Ball', selected: false },
      ],
    },
    {
      name: 'Ezugi-真人',
      expanded: false,
      keyword: '',
      games: [
        { name: 'Ezugi Live Baccarat', selected: false },
        { name: 'Ezugi Live Roulette', selected: false },
        { name: 'Ezugi Live Blackjack', selected: false },
        { name: 'Andar Bahar', selected: false },
        { name: 'Teen Patti', selected: false },
      ],
    },
  ];

  const createDefaultCategoryPlatforms = (category: string): GamePlatformItem[] => {
    return [
      {
        name: `${category}-平台A`,
        expanded: false,
        keyword: '',
        games: [
          { name: `${category} 热门推荐 1`, selected: false },
          { name: `${category} 热门推荐 2`, selected: false },
          { name: `${category} 经典专区 1`, selected: false },
          { name: `${category} 经典专区 2`, selected: false },
        ],
      },
      {
        name: `${category}-平台B`,
        expanded: false,
        keyword: '',
        games: [
          { name: `${category} 高爆率专区`, selected: false },
          { name: `${category} 新人专享`, selected: false },
        ],
      },
    ];
  };

  const gamePlatformGroups: Record<string, GamePlatformItem[]> = reactive({
    视讯: defaultVideoPlatforms,
    体育: createDefaultCategoryPlatforms('体育'),
    电子: createDefaultCategoryPlatforms('电子'),
    彩票: createDefaultCategoryPlatforms('彩票'),
    棋牌: createDefaultCategoryPlatforms('棋牌'),
    捕鱼: createDefaultCategoryPlatforms('捕鱼'),
    其他: createDefaultCategoryPlatforms('其他'),
  });

  const activeGameCategory = ref(gameCategoryOptions[0].value);

  const activeGamePlatforms = computed(
    () => gamePlatformGroups[activeGameCategory.value] ?? []
  );

  const filterPlatformGames = (platform: GamePlatformItem) => {
    const keyword = platform.keyword.trim().toLowerCase();

    if (!keyword) {
      return platform.games;
    }

    return platform.games.filter((game) =>
      game.name.toLowerCase().includes(keyword)
    );
  };

  const getPlatformSelectedCount = (platform: GamePlatformItem) =>
    platform.games.filter((game) => game.selected).length;

  const hasSelectedGames = (platform: GamePlatformItem) =>
    getPlatformSelectedCount(platform) > 0;

  const isPlatformAllSelected = (platform: GamePlatformItem) =>
    platform.games.length > 0 && platform.games.every((game) => game.selected);

  const isPlatformIndeterminate = (platform: GamePlatformItem) => {
    const selectedCount = getPlatformSelectedCount(platform);
    return selectedCount > 0 && selectedCount < platform.games.length;
  };

  const togglePlatformExpanded = (platform: GamePlatformItem) => {
    platform.expanded = !platform.expanded;
  };

  const togglePlatformSelection = (platform: GamePlatformItem) => {
    const nextValue = !isPlatformAllSelected(platform);

    platform.games.forEach((game) => {
      game.selected = nextValue;
    });
  };

  const clearPlatformSelection = (platform: GamePlatformItem) => {
    platform.games.forEach((game) => {
      game.selected = false;
    });
  };

  const currentCategorySummary = computed(() => ({
    platformCount: activeGamePlatforms.value.filter((item) =>
      hasSelectedGames(item)
    ).length,
    gameCount: activeGamePlatforms.value.reduce(
      (total, item) => total + getPlatformSelectedCount(item),
      0
    ),
  }));

  const selectedGameSummary = computed(() => {
    const selectedPlatforms = Object.values(gamePlatformGroups)
      .flat()
      .filter((item) => hasSelectedGames(item));

    return {
      platformCount: selectedPlatforms.length,
      gameCount: selectedPlatforms.reduce(
        (total, item) => total + getPlatformSelectedCount(item),
        0
      ),
    };
  });

  const currentCategoryAllSelected = computed(
    () =>
      activeGamePlatforms.value.length > 0 &&
      activeGamePlatforms.value.every((platform) =>
        isPlatformAllSelected(platform)
      )
  );

  const currentCategoryIndeterminate = computed(() => {
    const hasAny = activeGamePlatforms.value.some(
      (platform) => hasSelectedGames(platform)
    );
    return hasAny && !currentCategoryAllSelected.value;
  });

  const setPlatformSelected = (platform: GamePlatformItem, value: unknown) => {
    const nextValue = Boolean(value);

    platform.games.forEach((game) => {
      game.selected = nextValue;
    });
  };

  const toggleCurrentCategorySelection = () => {
    const nextValue = !currentCategoryAllSelected.value;

    activeGamePlatforms.value.forEach((platform) => {
      platform.games.forEach((game) => {
        game.selected = nextValue;
      });
    });
  };

  const baseRechargePreset = {
    showRechargeType: true,
    showLoopMode: true,
    names: [
      { lang: 'en', value: 'First Deposit Extra Bonus' },
      { lang: 'en-US', value: '' },
      { lang: 'zh-TW', value: '首充額外獎勵' },
      { lang: 'pt-BR', value: 'Prêmio extra para o primeiro recarregamer' },
      { lang: 'vi', value: 'Thưởng thêm nạp tiền lần đầu' },
    ],
    participationMode: '默认参与',
    loopMode: '不循环',
    distributionMode: '手动领取(过期作废)',
    distributionTime: '手动领取(过期自动领取)',
    rechargeType: '首充奖励',
    disabledParticipationValues: ['手动参与'],
    disabledLoopValues: [],
    disabledDistributionValues: [],
    disabledDistributionTimeValues: [],
  };

  const baseWashPreset = {
    showRechargeType: false,
    showLoopMode: false,
    names: [
      { lang: 'en', value: 'Wash Code Rewards' },
      { lang: 'en-US', value: '' },
      { lang: 'zh-TW', value: '洗碼獎勵' },
      { lang: 'pt-BR', value: 'Recompensa de lavagem de código' },
      { lang: 'vi', value: 'Phần thưởng rửa mã' },
    ],
    participationMode: '手动参与',
    loopMode: '',
    distributionMode: '手动领取(过期作废)',
    distributionTime: '手动领取(过期自动领取)',
    rechargeType: '',
    disabledParticipationValues: ['默认参与'],
    disabledLoopValues: [
      '不指定',
      '不循环',
      '每日循环',
      '每周循环',
      '每月循环',
      '无限循环',
    ],
    disabledDistributionValues: [],
    disabledDistributionTimeValues: [],
  };

  const baseRecommendPreset = {
    showRechargeType: false,
    showLoopMode: true,
    names: [
      { lang: 'en', value: 'Recommended Member Rewards' },
      { lang: 'en-US', value: '' },
      { lang: 'zh-TW', value: '推薦會員獎勵' },
      { lang: 'pt-BR', value: 'Recompensa para membro recomendado' },
      { lang: 'vi', value: 'Phần thưởng thành viên được giới thiệu' },
    ],
    participationMode: '默认参与',
    loopMode: '不循环',
    distributionMode: '自动派发',
    distributionTime: '实时(影响留存)',
    rechargeType: '首充奖励',
    disabledParticipationValues: ['手动参与'],
    disabledLoopValues: [
      '不指定',
      '不循环',
      '每日循环',
      '每周循环',
      '每月循环',
      '无限循环',
    ],
    disabledDistributionValues: [
      '手动领取(过期作废)',
      '手动领取(过期自动领取)',
    ],
    disabledDistributionTimeValues: [
      '手动领取(过期作废)',
      '手动领取(过期自动领取)',
    ],
  };

  const baseGeneralPreset = {
    ...baseRecommendPreset,
    showRechargeType: false,
    showLoopMode: true,
    names: [
      { lang: 'en', value: 'General Activity Rewards' },
      { lang: 'en-US', value: '' },
      { lang: 'zh-TW', value: '一般活動獎勵' },
      { lang: 'pt-BR', value: 'Recompensas gerais de atividade' },
      { lang: 'vi', value: 'Phần thưởng hoạt động chung' },
    ],
    participationMode: '默认参与',
    loopMode: '不循环',
    distributionMode: '自动派发',
    distributionTime: '实时(影响留存)',
    rechargeType: '首充奖励',
    disabledParticipationValues: ['手动参与'],
    disabledLoopValues: [
      '不指定',
      '不循环',
      '每日循环',
      '每周循环',
      '每月循环',
      '无限循环',
    ],
    disabledDistributionValues: [
      '手动领取(过期作废)',
      '手动领取(过期自动领取)',
    ],
    disabledDistributionTimeValues: [
      '手动领取(过期作废)',
      '手动领取(过期自动领取)',
    ],
  };

  const typePresets: Record<
    string,
    {
      showRechargeType: boolean;
      showLoopMode: boolean;
      names: { lang: string; value: string }[];
      participationMode: string;
      loopMode: string;
      distributionMode: string;
      distributionTime: string;
      rechargeType: string;
      disabledParticipationValues: string[];
      disabledLoopValues: string[];
      disabledDistributionValues: string[];
      disabledDistributionTimeValues: string[];
    }
  > = {
    充值: baseRechargePreset,
    洗码: baseWashPreset,
    推荐会员: baseRecommendPreset,
    幸运轮盘: { ...baseWashPreset, names: [
      { lang: 'en', value: 'Lucky Wheel Rewards' },
      { lang: 'en-US', value: '' },
      { lang: 'zh-TW', value: '幸運輪盤獎勵' },
      { lang: 'pt-BR', value: 'Recompensa da Roleta da Sorte' },
      { lang: 'vi', value: 'Phần thưởng vòng quay may mắn' },
    ] },
    助力: { ...baseRecommendPreset, names: [
      { lang: 'en', value: 'Assistance Rewards' },
      { lang: 'en-US', value: '' },
      { lang: 'zh-TW', value: '助力獎勵' },
      { lang: 'pt-BR', value: 'Recompensa de ajuda' },
      { lang: 'vi', value: 'Phần thưởng hỗ trợ' },
    ] },
    盲盒: { ...baseGeneralPreset, names: [
      { lang: 'en', value: 'Mystery Box Rewards' },
      { lang: 'en-US', value: '' },
      { lang: 'zh-TW', value: '盲盒獎勵' },
      { lang: 'pt-BR', value: 'Recompensa de caixa misteriosa' },
      { lang: 'vi', value: 'Phần thưởng hộp bí mật' },
    ] },
    新人彩金: { ...baseRechargePreset, names: [
      { lang: 'en', value: 'New User Bonus' },
      { lang: 'en-US', value: '' },
      { lang: 'zh-TW', value: '新人彩金' },
      { lang: 'pt-BR', value: 'Bônus de usuário novo' },
      { lang: 'vi', value: 'Thưởng người dùng mới' },
    ] },
    全民代理: { ...baseGeneralPreset, names: [
      { lang: 'en', value: 'Universal Agent Rewards' },
      { lang: 'en-US', value: '' },
      { lang: 'zh-TW', value: '全民代理獎勵' },
      { lang: 'pt-BR', value: 'Recompensa de agente universal' },
      { lang: 'vi', value: 'Phần thưởng đại lý toàn dân' },
    ] },
    救援金: { ...baseRechargePreset, names: [
      { lang: 'en', value: 'Rescue Fund Rewards' },
      { lang: 'en-US', value: '' },
      { lang: 'zh-TW', value: '救援金獎勵' },
      { lang: 'pt-BR', value: 'Recompensa do fundo de resgate' },
      { lang: 'vi', value: 'Phần thưởng quỹ cứu trợ' },
    ] },
    幸运注单: { ...baseWashPreset, names: [
      { lang: 'en', value: 'Lucky Bet Rewards' },
      { lang: 'en-US', value: '' },
      { lang: 'zh-TW', value: '幸運注單獎勵' },
      { lang: 'pt-BR', value: 'Recompensa de aposta sortuda' },
      { lang: 'vi', value: 'Phần thưởng cược may mắn' },
    ] },
    'VIP奖金': { ...baseRechargePreset, names: [
      { lang: 'en', value: 'VIP Bonus' },
      { lang: 'en-US', value: '' },
      { lang: 'zh-TW', value: 'VIP獎金' },
      { lang: 'pt-BR', value: 'Bônus VIP' },
      { lang: 'vi', value: 'Thưởng VIP' },
    ] },
    红包雨: { ...baseWashPreset, names: [
      { lang: 'en', value: 'Rain of Rewards' },
      { lang: 'en-US', value: '' },
      { lang: 'zh-TW', value: '紅包雨獎勵' },
      { lang: 'pt-BR', value: 'Chuva de recompensas' },
      { lang: 'vi', value: 'Mưa phần thưởng' },
    ] },
    '图文(广告)': { ...baseGeneralPreset, names: [
      { lang: 'en', value: 'Graphic Advertisement Rewards' },
      { lang: 'en-US', value: '' },
      { lang: 'zh-TW', value: '圖文廣告獎勵' },
      { lang: 'pt-BR', value: 'Recompensa publicitária gráfica' },
      { lang: 'vi', value: 'Phần thưởng quảng cáo hình ảnh' },
    ] },
    签到: { ...baseGeneralPreset, names: [
      { lang: 'en', value: 'Check-in Rewards' },
      { lang: 'en-US', value: '' },
      { lang: 'zh-TW', value: '簽到獎勵' },
      { lang: 'pt-BR', value: 'Recompensa de check-in' },
      { lang: 'vi', value: 'Phần thưởng điểm danh' },
    ] },
  };

  const getDefaultNamesByType = (type: string) =>
    (typePresets[type]?.names ?? typePresets.充值.names).map((item) => ({
      ...item,
    }));

  const createDefaultForm = (type = '充值') => ({
    type,
    names: getDefaultNamesByType(type),
    showInList: true,
    timeType: 'permanent',
    timeRange: [],
    specifiedGame: false,
    rechargeType:
      typePresets[type]?.rechargeType ?? typePresets.充值.rechargeType,
    loopMode: typePresets[type]?.loopMode ?? typePresets.充值.loopMode,
    participationMode:
      typePresets[type]?.participationMode ??
      typePresets.充值.participationMode,
    distributionMode:
      typePresets[type]?.distributionMode ?? typePresets.充值.distributionMode,
    distributionTime:
      typePresets[type]?.distributionTime ?? typePresets.充值.distributionTime,
    memberLevels: ['全部'],
    bonusMode: 'fixed',
    withdrawMultiple: '0',
    giftMode: 'part',
    clientList: ['全部'],
    activityRestrictions: [],
    rules: [
      {
        lang: 'en',
        content:
          '<p>Members can get corresponding bonus for first deposit</p><p>Bonus will be automatically claimed and distributed after expiration</p><p>Bonus (excluding principal) requires 1x valid bet to withdraw</p><p>Only the account holder can perform normal operations, otherwise the reward will be cancelled or deducted, frozen, or even blacklisted</p><p>To avoid differences in text understanding, the platform reserves the final right to interpret this promotion</p>',
      },
      {
        lang: 'en-US',
        content:
          '<p>Members can get corresponding bonus for first deposit</p><p>Bonus will be automatically claimed and distributed after expiration</p><p>Bonus (excluding principal) requires 1x valid bet to withdraw</p><p>Only the account holder can perform normal operations, otherwise the reward will be cancelled or deducted, frozen, or even blacklisted</p><p>To avoid differences in text understanding, the platform reserves the final right to interpret this promotion</p>',
      },
      {
        lang: 'zh-TW',
        content:
          '<p>會員首次存款，可獲得相應彩金</p><p>彩金自動領取，過期後彩金將自動派發</p><p>彩金（不含本金）需1倍有效投注才能提現</p><p>僅限帳號本人進行正常的人為操作，否則將取消或扣除獎勵、凍結、甚至拉入黑名單</p><p>為避免文字理解差異，平台將保留本活動最終解釋權</p>',
      },
      {
        lang: 'pt-BR',
        content:
          '<p>Membros podem receber bônus correspondente no primeiro depósito</p><p>Bônus será automaticamente reivindicado e distribuído após expiração</p><p>Bônus (excluindo principal) requer 1x aposta válida para saque</p><p>Apenas o titular da conta pode realizar operações normais, caso contrário a recompensa será cancelada ou deduzida, congelada, ou até na lista negra</p><p>Para evitar diferenças de entendimento de texto, a plataforma reserva o direito final de interpretar esta promotion</p>',
      },
      {
        lang: 'vi',
        content:
          '<p>Thành viên nhận thưởng tương ứng khi nạp lần đầu</p><p>Thưởng được nhận và phát tự động sau khi hết hạn</p><p>Thưởng (không gồm vốn gốc) cần cược hợp lệ 1 lần mới rút</p><p>Chỉ chủ tài khoản mới có thể thực hiện các thao tác bình thường, nếu không phần thưởng sẽ bị hủy hoặc bị khấu trừ, bị đóng băng, hoặc thậm chí bị đưa vào danh sách đen</p><p>Để tránh sự khác biệt trong việc hiểu văn bản, nền tảng giữ quyền giải thích cuối cùng cho chương trình khuyến mãi này</p>',
      },
    ],
  });

  const form = reactive(createDefaultForm());

  const closeDialog = () => {
    visible.value = false;
  };

  const getCurrentTypePreset = () => typePresets[form.type] ?? typePresets.充值;

  const showField = (field: 'rechargeType' | 'loopMode') => {
    const preset = getCurrentTypePreset();
    if (field === 'rechargeType') return preset.showRechargeType;
    return preset.showLoopMode;
  };

  const isFieldOptionDisabled = (field: string, value: string) => {
    const preset = getCurrentTypePreset();
    if (field === 'loopMode') {
      if (form.type === '推荐会员') {
        return form.loopMode !== value;
      }
      return preset.disabledLoopValues.includes(value);
    }

    if (field === 'participationMode') {
      return preset.disabledParticipationValues.includes(value);
    }

    if (field === 'distributionMode') {
      return preset.disabledDistributionValues.includes(value);
    }

    if (field === 'distributionTime') {
      return preset.disabledDistributionTimeValues.includes(value);
    }

    return false;
  };

  const resetGamePlatformState = () => {
    activeGameCategory.value = gameCategoryOptions[0].value;

    Object.values(gamePlatformGroups)
      .flat()
      .forEach((platform, index) => {
        platform.expanded = false;
        platform.keyword = '';
        platform.games.forEach((game) => {
          game.selected = false;
        });
      });
  };

  const resetForm = (type = '充值') => {
    Object.assign(form, createDefaultForm(type));
    resetGamePlatformState();
  };

  watch(
    () => visible.value,
    (isVisible) => {
      if (!isVisible) {
        resetForm();
      }
    }
  );

  watch(
    () => form.type,
    (type) => {
      const safeType = type && typePresets[type] ? type : '充值';
      resetForm(safeType);
    },
    { immediate: true }
  );

  const quillToolbar = [
    ['bold', 'italic', 'underline', 'strike'],
    [{ list: 'ordered' }, { list: 'bullet' }],
    ['blockquote', 'code-block'],
    [{ header: [1, 2, 3, false] }],
    [{ color: [] }, { background: [] }],
    ['clean'],
  ];

  const addName = () => form.names.push({ lang: 'en', value: '' });
  const removeName = (i: number) =>
    form.names.length > 1 && form.names.splice(i, 1);

  const addRule = () => form.rules.push({ lang: 'en', content: '' });
  const removeRule = (i: number) => form.rules.splice(i, 1);

  const isAllMemberLevelsSelected = computed(() =>
    memberLevelOptions.every((level) => form.memberLevels.includes(level))
  );

  const toggleMemberLevels = () => {
    form.memberLevels = isAllMemberLevelsSelected.value
      ? []
      : ['全部', ...memberLevelOptions];
  };

  const isAllClientListSelected = computed(() =>
    clientOptions.every((client) => form.clientList.includes(client))
  );

  const toggleClientList = () => {
    form.clientList = isAllClientListSelected.value
      ? []
      : ['全部', ...clientOptions];
  };

  watch(
    () => form.memberLevels.slice(),
    (levels) => {
      const filtered = levels.filter((level) => level !== '全部');
      const allSelected = memberLevelOptions.every((level) =>
        filtered.includes(level)
      );
      const hasAll = levels.includes('全部');

      if (hasAll && !allSelected) {
        form.memberLevels = filtered.length
          ? filtered
          : ['全部', ...memberLevelOptions];
        return;
      }

      if (allSelected && !hasAll) {
        form.memberLevels = ['全部', ...filtered];
      }
    },
    { immediate: true }
  );

  watch(
    () => form.clientList.slice(),
    (clients) => {
      const filtered = clients.filter((client) => client !== '全部');
      const allSelected = clientOptions.every((client) =>
        filtered.includes(client)
      );
      const hasAll = clients.includes('全部');

      if (hasAll && !allSelected) {
        form.clientList = filtered.length
          ? filtered
          : ['全部', ...clientOptions];
        return;
      }

      if (allSelected && !hasAll) {
        form.clientList = ['全部', ...filtered];
      }
    },
    { immediate: true }
  );
</script>

<style scoped>
  .activity-create-modal :deep(.arco-modal) {
    border-radius: 16px;
    overflow: hidden;
    background: #fff;
    box-shadow: 0 18px 46px rgba(29, 33, 41, 0.12);
  }

  .activity-create-modal :deep(.arco-modal-header) {
    padding: 0;
    border-bottom: 1px solid #edf1f5;
    background: #fff;
  }

  .activity-create-modal :deep(.arco-modal-content) {
    background: #f3f5f7;
  }

  .activity-create-modal :deep(.arco-modal-body) {
    padding: 0;
    background: #f3f5f7;
  }

  /* 苹果 macOS 窗口弹出与关闭微交互动效 */
  .activity-create-modal :deep(.zoom-modal-enter-active) {
    transition: transform 0.32s cubic-bezier(0.16, 1, 0.3, 1),
      opacity 0.28s cubic-bezier(0.16, 1, 0.3, 1),
      filter 0.28s ease-out;
    transform-origin: center center;
  }

  .activity-create-modal :deep(.zoom-modal-leave-active) {
    transition: transform 0.24s cubic-bezier(0.32, 0.72, 0, 1),
      opacity 0.2s cubic-bezier(0.32, 0.72, 0, 1),
      filter 0.2s ease-in;
    transform-origin: center center;
  }

  .activity-create-modal :deep(.zoom-modal-enter-from) {
    opacity: 0;
    transform: scale(0.86) translateY(12px);
    filter: blur(4px);
  }

  .activity-create-modal :deep(.zoom-modal-leave-to) {
    opacity: 0;
    transform: scale(0.82) translateY(16px);
    filter: blur(6px);
  }

  .activity-create-modal :deep(.fade-modal-enter-active),
  .activity-create-modal :deep(.fade-modal-leave-active) {
    transition: opacity 0.24s cubic-bezier(0.32, 0.72, 0, 1);
  }

  .activity-create-modal :deep(.fade-modal-enter-from),
  .activity-create-modal :deep(.fade-modal-leave-to) {
    opacity: 0;
  }

  .dialog-header-title {
    padding: 14px 20px 12px;
    font-size: 18px;
    font-weight: 700;
    color: #1d2129;
    letter-spacing: 0.2px;
  }

  .activity-dialog-toolbar {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    padding: 10px 18px 12px;
    background: #fff;
    border-bottom: 1px solid #edf1f5;
  }

  .toolbar-title {
    flex: 1;
    min-height: 1px;
  }

  .toolbar-actions {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .toolbar-btn {
    height: 32px;
    min-width: 88px;
    border-radius: 6px;
  }

  .toolbar-btn.secondary {
    background: #f5f7fb;
    color: #4e5969;
    border: 1px solid #e5e6eb;
  }

  .toolbar-btn.primary {
    background: #165dff;
    border: 1px solid #165dff;
    color: #fff;
  }

  .toolbar-close {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    border-radius: 6px;
    color: #4e5969;
  }

  .custom-form {
    padding: 18px 18px 0;
    background: #f3f5f7;
  }

  .activity-form-grid {
    align-items: stretch;
    margin: 0;
  }

  .section-card {
    margin-bottom: 16px;
    padding: 18px 18px 0;
    border: 1px solid #e8ebf0;
    border-radius: 14px;
    background: #fff;
    box-shadow: 0 2px 10px rgba(22, 93, 255, 0.02);
  }

  .section-title {
    margin: 0 0 10px;
    padding-left: 10px;
    font-size: 14px;
    font-weight: 700;
    color: #1d2129;
    position: relative;
    letter-spacing: 0.2px;
  }

  .accordion-section {
    margin-bottom: 12px;
  }

  .accordion-section .section-card {
    margin: 0;
    padding-top: 20px;
    border-top: 0;
    border-radius: 0 0 12px 12px;
    box-shadow: 0 8px 20px rgba(29, 33, 41, 0.035);
  }

  .section-title--collapsible {
    min-height: 52px;
    padding: 0 14px 0 16px;
    border: 1px solid #e5e8ef;
    border-radius: 12px;
    background: #fff;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    cursor: pointer;
    user-select: none;
    transition: border-color 0.2s ease, background-color 0.2s ease,
      box-shadow 0.2s ease;
  }

  .section-title--collapsible:hover {
    border-color: #bfd4ff;
    background: #f8fbff;
    box-shadow: 0 4px 12px rgba(22, 93, 255, 0.08);
  }

  .section-title--collapsible.active {
    margin-bottom: 0;
    color: #165dff;
    border-color: #c8dbff;
    border-bottom-color: #e7efff;
    border-radius: 12px 12px 0 0;
    background: linear-gradient(90deg, #f2f7ff 0%, #fff 72%);
    box-shadow: 0 4px 12px rgba(22, 93, 255, 0.06);
  }

  .section-title--collapsible:focus-visible {
    outline: 2px solid rgba(22, 93, 255, 0.35);
    outline-offset: 3px;
    border-radius: 4px;
  }

  .section-title-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    border-radius: 50%;
    background: #f2f3f5;
    color: #86909c;
    font-size: 14px;
    transition: transform 0.2s ease;
  }

  .section-title--collapsible.active .section-title-icon {
    color: #165dff;
    background: #e8f1ff;
  }

  .section-title-icon.collapsed {
    transform: rotate(-90deg);
  }

  .section-title::before {
    content: '';
    position: absolute;
    left: 0;
    top: 2px;
    width: 3px;
    height: 14px;
    border-radius: 2px;
    background: #165dff;
  }

  .section-body {
    display: block;
  }

  .form-panel-left,
  .form-panel-right {
    min-height: 560px;
    background: transparent;
  }

  .form-panel-left,
  .form-panel-right {
    display: flex;
    flex-direction: column;
  }

  .form-panel-right > .accordion-section:last-child {
    margin-bottom: 0;
    display: flex;
    flex: 1 1 auto;
    flex-direction: column;
  }

  .form-panel-right > .accordion-section:last-child .section-card {
    flex: 1 1 auto;
  }

  .form-panel-left > .accordion-section:last-child {
    display: flex;
    flex: 1 1 auto;
    flex-direction: column;
    margin-bottom: 0;
  }

  .form-panel-left > .accordion-section:last-child .section-card {
    flex: 1 1 auto;
  }

  .form-panel-left {
    padding-right: 18px;
    border-right: 1px solid #edf1f5;
  }

  .form-panel-right {
    padding-left: 18px;
  }

  .custom-form :deep(.arco-form-item) {
    margin-bottom: 18px;
  }

  .custom-form :deep(.arco-form-item-label) {
    color: #1d2129;
    font-size: 13px;
    font-weight: 600;
    line-height: 1.4;
    padding-right: 12px;
  }

  .custom-form :deep(.arco-form-item-label-col) {
    justify-content: flex-end;
    flex: 0 0 120px;
    max-width: 120px;
    padding-right: 4px;
  }

  .custom-form :deep(.arco-form-item-content) {
    min-width: 0;
    padding-left: 4px;
  }

  .custom-form :deep(.arco-radio-group),
  .custom-form :deep(.arco-checkbox-group) {
    gap: 8px 14px;
    align-items: center;
  }

  .custom-form :deep(.arco-radio),
  .custom-form :deep(.arco-checkbox) {
    color: #4e5969;
    --checkbox-border-color: #c9cdd4;
    display: inline-flex;
    align-items: center;
  }

  .custom-form :deep(.arco-radio-checked .arco-radio-inner),
  .custom-form :deep(.arco-checkbox-checked .arco-checkbox-inner) {
    border-color: #165dff;
    background-color: #165dff;
  }

  .custom-form :deep(.arco-radio-label),
  .custom-form :deep(.arco-checkbox-label) {
    font-size: 13px;
  }

  .custom-form :deep(.arco-select-view),
  .custom-form :deep(.arco-input),
  .custom-form :deep(.arco-input-wrapper),
  .custom-form :deep(.arco-range-picker),
  .custom-form :deep(.arco-picker) {
    background: #fff;
    border-color: #dfe3e8;
    border-radius: 6px;
    box-shadow: none;
    transition: border-color 0.2s ease, box-shadow 0.2s ease;
  }

  .custom-form :deep(.arco-select-view):focus-within,
  .custom-form :deep(.arco-input-wrapper):focus-within,
  .custom-form :deep(.arco-picker):focus-within {
    border-color: #93b6ff;
    box-shadow: 0 0 0 2px rgba(22, 93, 255, 0.08);
  }

  .custom-form :deep(.arco-select-view),
  .custom-form :deep(.arco-input-wrapper) {
    height: 32px;
  }

  .custom-form :deep(.arco-select-view-value),
  .custom-form :deep(.arco-input) {
    color: #1d2129;
    font-size: 13px;
  }

  .type-grid {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .type-item {
    padding: 4px 12px;
    min-width: 72px;
    height: 30px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    font-size: 13px;
    color: #4e5969;
    background-color: #f2f3f5;
    border: 1px solid transparent;
    border-radius: 4px;
    cursor: pointer;
    user-select: none;
    transition: all 0.2s ease;
  }

  .type-item:hover {
    background-color: #e5e6eb;
    border-color: #d9e8ff;
  }

  .type-item.active {
    color: #fff;
    background-color: #165dff;
    border-color: #165dff;
    box-shadow: 0 4px 10px rgba(22, 93, 255, 0.18);
  }

  /* 2. 多语言行样式 */
  .dynamic-group {
    width: 100%;
  }

  .dynamic-row {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;
    padding: 8px 8px 8px 10px;
    border: 1px solid #edf1f5;
    border-radius: 8px;
    background: #fafbfc;
  }

  .dynamic-row :deep(.arco-select-view),
  .dynamic-row :deep(.arco-input-wrapper) {
    height: 30px;
  }

  .add-btn {
    width: 28px;
    height: 28px;
    padding: 0;
    border-radius: 4px;
    background: #e8f0ff;
    border-color: #bfd4ff;
    color: #165dff;
  }

  /* 3. 时间配置行 */
  .time-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .time-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  /* 4. 单选组换行排版 */
  .wrap-radio-group {
    display: flex;
    flex-wrap: wrap;
    gap: 8px 16px;
  }

  /* 5. 右侧上传与规则 */
  .upload-container {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 12px 14px;
    border: 1px solid #edf1f5;
    border-radius: 10px;
    background: #fafbfc;
  }

  .upload-tip {
    font-size: 12px;
    color: #86909c;
    line-height: 1.6;
  }

  .upload-container :deep(.arco-upload-list) {
    margin-top: 4px;
  }

  .upload-container :deep(.arco-upload-picture-card) {
    width: 100%;
    min-height: 120px;
  }

  .checkbox-stack {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .checkbox-row {
    display: flex;
    flex-wrap: wrap;
    gap: 12px 16px;
    align-items: center;
    padding: 2px 0;
  }

  .full-row {
    margin-bottom: 4px;
    cursor: pointer;
    padding-top: 2px;
  }

  .full-row :deep(.arco-checkbox) {
    pointer-events: none;
  }

  .inline-input-group {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .inline-helper {
    color: #86909c;
    white-space: nowrap;
  }

  .rule-block {
    margin-bottom: 12px;
    padding: 12px 12px 0;
    border: 1px solid #edf1f5;
    border-radius: 10px;
    background: #fafbfc;
  }

  .rule-block :deep(.arco-select-view) {
    height: 30px;
    border-radius: 6px;
  }

  .dialog-footer-actions {
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 16px;
    padding: 20px 0 22px;
    background: #f3f5f7;
    border-top: 1px solid #edf1f5;
  }

  .cancel-btn,
  .save-draft-btn {
    min-width: 104px;
    height: 36px;
    border-radius: 8px;
  }

  .editor-wrapper {
    position: relative;
    padding-right: 36px;
  }

  .quill-editor {
    width: 100%;
  }

  :deep(.ql-toolbar) {
    border: 1px solid #e5e6eb;
    border-bottom: none;
    border-radius: 8px 8px 0 0;
    background: #f7f8fa;
    min-height: 38px;
  }

  :deep(.ql-container) {
    min-height: 180px;
    max-height: 260px;
    overflow-y: auto;
    border: 1px solid #e5e6eb;
    border-radius: 0 0 8px 8px;
    font-size: 14px;
    line-height: 1.8;
    background: #fff;
  }

  :deep(.ql-editor) {
    min-height: 180px;
    color: #1d2129;
    font-size: 13px;
  }

  :deep(.ql-editor.ql-blank::before) {
    color: #c9cdd4;
    font-style: normal;
  }

  :deep(.ql-editor blockquote) {
    margin: 8px 0;
    padding-left: 12px;
    border-left: 3px solid #165dff;
    color: #4e5969;
  }

  .remove-rule-btn {
    position: absolute;
    right: 4px;
    top: 50%;
    transform: translateY(-50%);
  }

  .currency-text {
    font-weight: 500;
    color: #1d2129;
  }

  .specified-game-section {
    width: 100%;
  }

  .specified-game-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
    padding: 14px 16px;
    border: 1px solid #e5e6eb;
    border-radius: 10px;
    background: linear-gradient(180deg, #f8fbff 0%, #ffffff 100%);
    box-shadow: inset 0 0 0 1px rgba(22, 93, 255, 0.02);
  }

  .specified-game-title {
    font-size: 14px;
    font-weight: 600;
    color: #1d2129;
  }

  .specified-game-desc {
    margin-top: 4px;
    font-size: 12px;
    color: #86909c;
    line-height: 1.6;
  }

  .specified-game-panel {
    margin-top: 12px;
    padding: 16px;
    border: 1px solid #e5e6eb;
    border-radius: 12px;
    background: #fff;
  }

  .specified-game-tabs {
    display: flex;
    align-items: center;
    gap: 28px;
    padding-bottom: 8px;
    border-bottom: 1px solid #e5e6eb;
  }

  .specified-game-tab {
    position: relative;
    padding: 6px 4px;
    font-size: 14px;
    font-weight: 500;
    color: #4e5969;
    cursor: pointer;
    transition: color 0.2s ease;
    user-select: none;
  }

  .specified-game-tab:hover {
    color: #165dff;
  }

  .specified-game-tab.active {
    color: #165dff;
    font-weight: 600;
  }

  .specified-game-tab.active::after {
    content: '';
    position: absolute;
    bottom: -9px;
    left: 0;
    width: 100%;
    height: 2px;
    background: #165dff;
    border-radius: 2px;
  }

  .specified-game-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-top: 14px;
    padding: 0 4px;
  }

  .specified-game-toolbar-hint {
    font-size: 12px;
    color: #86909c;
  }

  .specified-game-list {
    display: grid;
    grid-template-columns: 1fr;
    gap: 10px;
    margin-top: 14px;
  }

  .specified-game-card {
    padding: 0;
    border: 1px solid #e5e6eb;
    border-radius: 8px;
    background: #fff;
    transition: all 0.2s ease;
  }

  .specified-game-card:hover {
    border-color: #bfd4ff;
  }

  .specified-game-card.selected {
    border-color: #e5e6eb;
    background: #fff;
  }

  .specified-game-card-header {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 14px;
    cursor: pointer;
  }

  .specified-game-card-body {
    flex: 1;
    min-width: 0;
  }

  .specified-game-card-top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }

  .specified-game-card-name {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    font-weight: 500;
    color: #1d2129;
  }

  .specified-game-count-hint {
    font-size: 12px;
    color: #86909c;
    font-weight: normal;
  }

  .specified-game-card-arrow {
    color: #86909c;
    font-size: 12px;
    flex-shrink: 0;
  }

  .specified-game-subpanel {
    padding: 0 16px 16px 44px;
  }

  .specified-game-subtoolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 10px 0;
  }

  .specified-game-search {
    width: 100%;
    max-width: 100%;
  }

  .specified-game-subtoolbar-actions {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .specified-game-options {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(170px, 1fr));
    gap: 12px 16px;
    padding: 6px 0;
  }

  .specified-game-option {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
  }

  .specified-game-option-name {
    font-size: 13px;
    color: #1d2129;
    line-height: 1.4;
  }

  .specified-game-empty {
    padding: 18px 0 6px;
    color: #86909c;
    font-size: 12px;
    text-align: center;
  }

  .specified-game-summary {
    margin-top: 12px;
    font-size: 12px;
    color: #4e5969;
  }
</style>
