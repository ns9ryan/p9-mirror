<template>
  <a-modal
    v-model:visible="visible"
    width="min(1960px, calc(100vw - 48px))"
    unmount-on-close
    :footer="false"
    class="activity-create-modal"
    :maskClosable="false"
    :on-before-cancel="handleBeforeCancel"
  >
    <template #title>
      <div class="dialog-header-title">
        {{ isEdit ? '编辑活动' : '新增活动' }}
      </div>
    </template>
    <div v-if="preparing" class="detail-loading">
      <a-spin :tip="isEdit ? '正在加载活动详情' : '加载中'" />
    </div>
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
        <a-col :xs="24" :xl="12" class="form-panel-left">
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
                      :key="item.value"
                      class="type-item"
                      :class="{ active: form.type === item.value }"
                      @click="form.type = item.value"
                    >
                      {{ item.label }}
                    </div>
                  </div>
                </a-form-item>

                <!-- 货币 -->
                <a-form-item label="货币">
                  <span class="currency-text">{{
                    currency || '--'
                  }}</span>
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
              <div class="section-body settings-body">
                <!-- 是否在活动列表内展示 -->
                <a-form-item
                  label="是否在活动列表内展示"
                  :class="leadingItemClass"
                >
                  <a-radio-group v-model="form.showInList">
                    <a-radio :value="true">是</a-radio>
                    <a-radio :value="false">否</a-radio>
                  </a-radio-group>
                </a-form-item>

                <!-- 活动所属 -->
                <a-form-item label="活动所属" required>
                  <a-radio-group v-model="form.section">
                    <a-radio
                      v-for="opt in getEnumOptions('section')"
                      :key="opt.value"
                      :value="opt.value"
                      >{{ opt.label }}</a-radio
                    >
                  </a-radio-group>
                </a-form-item>

                <!-- 侧边栏个人中心入口显示（VIP奖金） -->
                <a-form-item
                  v-if="form.type === ACTIVITY_TYPE.VIP"
                  label="侧边栏个人中心入口显示"
                >
                  <a-radio-group v-model="form.sidebarEntry">
                    <a-radio :value="true">是</a-radio>
                    <a-radio :value="false">否</a-radio>
                  </a-radio-group>
                </a-form-item>

                <!-- 活动时间 -->
                <a-form-item
                  v-if="!activityTimeHidden"
                  label="活动时间"
                  required
                  :class="leadingItemClass"
                >
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

                <!-- 活动条件（新人彩金） -->
                <a-form-item
                  v-if="form.type === ACTIVITY_TYPE.NEWCOMER"
                  label="活动条件"
                  required
                >
                  <a-select
                    v-model="form.activityCondition"
                    class="assist-amount-input"
                  >
                    <a-option
                      v-for="opt in activityConditionOptions"
                      :key="opt"
                      :value="opt"
                    >
                      {{ opt }}
                    </a-option>
                  </a-select>
                </a-form-item>

                <!-- 充值类型 -->
                <a-form-item
                  v-if="form.type === ACTIVITY_TYPE.DEPOSIT"
                  label="充值类型"
                  required
                >
                  <a-radio-group v-model="form.rechargeType">
                    <a-radio
                      v-for="opt in getEnumOptions('deposit_reward_type')"
                      :key="opt.value"
                      :value="opt.value"
                    >
                      {{ opt.label }}
                    </a-radio>
                  </a-radio-group>
                </a-form-item>

                <!-- 参与模式 -->
                <a-form-item
                  v-if="!isFieldHidden('participation')"
                  label="参与模式"
                  required
                >
                  <a-radio-group
                    v-model="form.participationMode"
                    class="wrap-radio-group"
                  >
                    <a-radio
                      v-for="opt in getEnumOptions('apply_mode')"
                      :key="opt.value"
                      :value="opt.value"
                      :disabled="
                        isFieldOptionDisabled('participationMode', opt.value)
                      "
                      >{{ opt.label }}</a-radio
                    >
                  </a-radio-group>
                </a-form-item>
                <template v-if="form.type === ACTIVITY_TYPE.VIP">
                  <!-- 活动唯一性（VIP奖金） -->
                  <a-form-item label="活动唯一性">
                    <a-radio-group v-model="form.vipUniqueness">
                      <a-radio value="系统默认">系统默认</a-radio>
                    </a-radio-group>
                  </a-form-item>

                  <!-- VIP配置（VIP奖金） -->
                  <a-form-item label="VIP配置" required>
                    <table class="vip-config-table">
                      <thead>
                        <tr>
                          <th>VIP等级</th>
                          <th v-for="col in vipConfigColumns" :key="col.key">
                            {{ col.title }}
                          </th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr v-for="level in vipLevels" :key="level">
                          <td>{{ level }}</td>
                          <template v-if="level === 0">
                            <td v-for="col in vipConfigColumns" :key="col.key"
                              >-</td
                            >
                          </template>
                          <template v-else>
                            <td v-for="col in vipConfigColumns" :key="col.key">
                              <a-input-number
                                v-model="form.vipConfigs[level][col.key]"
                                :min="0"
                                hide-button
                                class="vip-config-input"
                              />
                            </td>
                          </template>
                        </tr>
                        <tr v-if="!vipLevels.length">
                          <td :colspan="vipConfigColumns.length + 1">
                            暂无VIP等级，请先在 VIP 管理中配置
                          </td>
                        </tr>
                      </tbody>
                    </table>
                  </a-form-item>

                  <!-- 领取规则（VIP奖金） -->
                  <a-form-item label="领取规则">
                    <a-radio-group v-model="form.vipClaimRule">
                      <a-radio value="跨天、跨周、跨月"
                        >跨天、跨周、跨月</a-radio
                      >
                    </a-radio-group>
                  </a-form-item>
                </template>

                <!-- 奖励模式（全民代理） -->
                <a-form-item
                  v-if="form.type === ACTIVITY_TYPE.AGENT"
                  label="奖励模式"
                  required
                >
                  <div class="bonus-mode-group">
                    <a-radio-group v-model="form.agentRewardMode">
                      <a-radio value="3级抽成">3级抽成</a-radio>
                    </a-radio-group>
                    <div
                      v-for="(level, index) in agentCommissionLevels"
                      :key="level"
                      class="agent-commission-item"
                    >
                      <a-input-number
                        v-model="form.agentCommissionRates[index]"
                        :min="0"
                        :max="100"
                        hide-button
                        placeholder="有效投注佣金率设置"
                        class="offer-amount-input"
                      >
                        <template #prefix>第{{ level }}级会员</template>
                        <template #suffix>%</template>
                      </a-input-number>
                      <div class="agent-commission-formula">
                        上级收益 = 有效投注 {{ getAgentFormula(level) }}
                      </div>
                    </div>
                  </div>
                </a-form-item>
                <!-- 循环模式 -->
                <a-form-item v-if="showLoopMode" :label="loopLabel" required>
                  <a-radio-group
                    v-model="form.loopMode"
                    class="wrap-radio-group"
                  >
                    <a-radio
                      v-for="opt in loopOptions"
                      :key="opt"
                      :value="opt"
                      :disabled="isLoopModeDisabled(opt)"
                      >{{ loopModeLabel(opt) }}</a-radio
                    >
                  </a-radio-group>
                </a-form-item>
                <!-- 派发奖励模式 -->
                <a-form-item
                  v-if="!isFieldHidden('distribution')"
                  label="派发奖励模式"
                  required
                >
                  <a-radio-group
                    v-model="form.distributionMode"
                    class="wrap-radio-group"
                  >
                    <a-radio
                      v-for="opt in getEnumOptions('payout_mode')"
                      :key="opt.value"
                      :value="opt.value"
                      :disabled="
                        isFieldOptionDisabled('distributionMode', opt.value)
                      "
                      >{{ opt.label }}</a-radio
                    >
                  </a-radio-group>
                </a-form-item>
                <!-- 领取过期天数（紧跟派发奖励模式） -->
                <a-form-item v-if="showClaimExpireDays" label="领取过期天数">
                  <a-input-number
                    v-model="form.claimExpireDays"
                    :min="1"
                    :precision="0"
                    hide-button
                    style="width: 400px"
                  >
                    <template #suffix>天</template>
                  </a-input-number>
                </a-form-item>

                <!-- 派发奖励时间 -->
                <a-form-item
                  v-if="!isFieldHidden('distribution')"
                  label="派发奖励时间"
                  required
                >
                  <a-radio-group
                    v-model="form.distributionTime"
                    class="wrap-radio-group"
                  >
                    <a-radio
                      v-for="opt in getEnumOptions('payout_timing')"
                      :key="opt.value"
                      :value="opt.value"
                      :disabled="
                        isFieldOptionDisabled('distributionTime', opt.value)
                      "
                      >{{ opt.label }}</a-radio
                    >
                  </a-radio-group>
                </a-form-item>

                <!-- 奖励模式（推荐会员） -->
                <a-form-item
                  v-if="form.type === ACTIVITY_TYPE.REFER"
                  label="奖励模式"
                  required
                >
                  <div class="bonus-mode-group">
                    <a-radio-group v-model="form.referralRewardMode">
                      <a-radio value="单独人数奖励">单独人数奖励</a-radio>
                      <a-radio value="累计人数奖励">累计人数奖励</a-radio>
                    </a-radio-group>

                    <template v-if="form.referralRewardMode === '单独人数奖励'">
                      <a-input-number
                        v-model="form.referralSingleAmount"
                        :min="0"
                        hide-button
                        placeholder="赠送金额"
                        class="assist-amount-input"
                        :error="
                          referralValidated && form.referralSingleAmount == null
                        "
                      >
                        <template #prefix>每个新增会员奖励</template>
                        <template #suffix>{{ currency }}</template>
                      </a-input-number>
                    </template>

                    <template v-else-if="form.referralRewardMode === '累计人数奖励'">
                      <div
                        v-for="(tier, index) in form.bonusTiers"
                        :key="tier.uid"
                        class="bonus-tier-row"
                      >
                        <span class="bonus-tier-index">{{ index + 1 }}</span>
                        <a-input-number
                          v-model="tier.threshold"
                          :min="1"
                          :precision="0"
                          hide-button
                          placeholder="推荐人数"
                          class="bonus-tier-input"
                          :error="
                            referralValidated && tier.threshold == null
                          "
                        />
                        <a-input-number
                          v-model="tier.reward"
                          :min="0"
                          hide-button
                          placeholder="赠送金额"
                          class="bonus-tier-input"
                          :error="referralValidated && tier.reward == null"
                        />
                        <a-button
                          type="text"
                          status="danger"
                          shape="circle"
                          :disabled="form.bonusTiers.length <= 1"
                          @click="removeBonusTier(index)"
                        >
                          <template #icon><icon-minus-circle /></template>
                        </a-button>
                      </div>
                      <a-button
                        type="primary"
                        class="bonus-tier-add-btn"
                        @click="addBonusTier"
                      >
                        <template #icon><icon-plus /></template>
                      </a-button>
                    </template>

                    <div
                      v-if="referralValidated && referralRewardError"
                      class="agent-id-error"
                    >
                      {{ referralRewardError }}
                    </div>
                  </div>
                </a-form-item>

                <!-- 被邀请会员有效判定（推荐会员） -->
                <a-form-item
                  v-if="form.type === ACTIVITY_TYPE.REFER"
                  label="被邀请会员有效判定"
                  required
                >
                  <div class="invitee-rule-grid">
                    <div
                      v-for="rule in inviteeRuleOptions"
                      :key="rule.key"
                      class="invitee-rule-item"
                    >
                      <a-checkbox v-model="form.inviteeRules[rule.key].checked">
                        {{ rule.label }}
                      </a-checkbox>
                      <a-input-number
                        v-model="form.inviteeRules[rule.key].value"
                        :min="0"
                        :precision="rule.unit === '人' ? 0 : undefined"
                        hide-button
                        class="invitee-rule-input"
                      />
                      <span class="invitee-rule-unit">{{ rule.unit || currency }}</span>
                      <a-tooltip :content="rule.tip">
                        <icon-info-circle-fill class="claim-mode-tip" />
                      </a-tooltip>
                    </div>
                  </div>
                </a-form-item>

                <!-- 奖金设置（幸运轮盘） -->
                <a-form-item
                  v-if="form.type === ACTIVITY_TYPE.LUCK_WHEEL"
                  label="奖金设置"
                  required
                >
                  <div class="bonus-mode-group">
                    <a-radio-group v-model="form.wheelSize">
                      <a-radio :value="6">6格轮盘</a-radio>
                      <a-radio :value="12">12格轮盘</a-radio>
                    </a-radio-group>
                    <div
                      v-for="(prize, index) in form.wheelPrizes"
                      :key="prize.uid"
                      class="bonus-tier-row"
                    >
                      <span class="bonus-tier-index">{{ index + 1 }}</span>
                      <a-select
                        v-model="prize.prizeType"
                        class="wheel-prize-type"
                      >
                        <a-option
                          v-for="opt in wheelPrizeTypeOptions"
                          :key="opt"
                          :value="opt"
                        >
                          {{ opt }}
                        </a-option>
                      </a-select>
                      <a-input-number
                        v-model="prize.amount"
                        :min="0"
                        hide-button
                        :disabled="prize.prizeType === '谢谢参与'"
                        :placeholder="
                          prize.prizeType === '比例奖金' ? '比例%' : '金额'
                        "
                        class="wheel-prize-input"
                      />
                      <a-input-number
                        v-model="prize.probability"
                        :min="0"
                        :max="100"
                        hide-button
                        placeholder="输入概率%"
                        class="wheel-prize-input wheel-prize-prob"
                      />
                    </div>
                  </div>
                </a-form-item>

                <!-- 抽奖需要条件（幸运轮盘） -->
                <a-form-item
                  v-if="form.type === ACTIVITY_TYPE.LUCK_WHEEL"
                  label="抽奖需要条件"
                  required
                >
                  <div class="draw-condition-grid">
                    <div class="draw-condition-col">
                      <a-checkbox
                        :model-value="
                          form.drawConditions.includes(WHEEL_BY_BET)
                        "
                        @change="toggleDrawCondition(WHEEL_BY_BET, $event)"
                      >
                        达标有效投注赠送1次机会
                      </a-checkbox>
                      <template
                        v-if="form.drawConditions.includes(WHEEL_BY_BET)"
                      >
                        <div
                          v-for="(tier, index) in form.wheelBetTiers"
                          :key="tier.uid"
                          class="bonus-tier-row"
                        >
                          <span class="bonus-tier-index">{{ index + 1 }}</span>
                          <a-input-number
                            v-model="tier.threshold"
                            :min="0"
                            hide-button
                            placeholder="要求有效投注金额"
                            class="draw-condition-input"
                          />
                          <a-button
                            type="text"
                            status="danger"
                            shape="circle"
                            :disabled="form.wheelBetTiers.length <= 1"
                            @click="removeWheelBetTier(index)"
                          >
                            <template #icon><icon-minus-circle /></template>
                          </a-button>
                        </div>
                        <a-button
                          type="primary"
                          class="bonus-tier-add-btn"
                          @click="addWheelBetTier"
                        >
                          <template #icon><icon-plus /></template>
                        </a-button>
                      </template>
                    </div>

                    <div class="draw-condition-col">
                      <a-checkbox
                        :model-value="
                          form.drawConditions.includes(WHEEL_BY_DEPOSIT)
                        "
                        @change="toggleDrawCondition(WHEEL_BY_DEPOSIT, $event)"
                      >
                        达标充值金额赠送1次机会
                      </a-checkbox>
                      <template
                        v-if="form.drawConditions.includes(WHEEL_BY_DEPOSIT)"
                      >
                        <div
                          v-for="(tier, index) in form.wheelDepositTiers"
                          :key="tier.uid"
                          class="bonus-tier-row"
                        >
                          <span class="bonus-tier-index">{{ index + 1 }}</span>
                          <a-input-number
                            v-model="tier.threshold"
                            :min="0"
                            hide-button
                            placeholder="要求充值金额"
                            class="draw-condition-input"
                          />
                          <a-button
                            type="text"
                            status="danger"
                            shape="circle"
                            :disabled="form.wheelDepositTiers.length <= 1"
                            @click="removeWheelDepositTier(index)"
                          >
                            <template #icon><icon-minus-circle /></template>
                          </a-button>
                        </div>
                        <a-button
                          type="primary"
                          class="bonus-tier-add-btn"
                          @click="addWheelDepositTier"
                        >
                          <template #icon><icon-plus /></template>
                        </a-button>
                      </template>
                    </div>
                  </div>
                </a-form-item>

                <!-- 额外条件（幸运轮盘） -->
                <a-form-item
                  v-if="form.type === ACTIVITY_TYPE.LUCK_WHEEL"
                  label="额外条件"
                  required
                >
                  <a-checkbox-group v-model="form.extraConditions">
                    <div class="checkbox-row">
                      <a-checkbox value="达标每日首充赠送一次机会"
                        >达标每日首充赠送一次机会</a-checkbox
                      >
                      <a-checkbox value="达标账户首充赠送一次机会"
                        >达标账户首充赠送一次机会</a-checkbox
                      >
                    </div>
                  </a-checkbox-group>
                </a-form-item>

                <!-- 本次实际奖金（助力） -->
                <a-form-item
                  v-if="form.type === ACTIVITY_TYPE.ASSIST"
                  label="本次实际奖金"
                  required
                >
                  <div class="bonus-mode-group">
                    <div class="assist-ref-panel">
                      <div class="assist-ref-title">参考图示</div>
                      <div class="assist-ref-list">
                        <div
                          v-for="refImage in assistReferenceImages"
                          :key="refImage.caption"
                          class="assist-ref-item"
                        >
                          <div class="assist-ref-image">
                            <img
                              v-if="refImage.src"
                              :src="refImage.src"
                              :alt="refImage.caption"
                            />
                            <icon-image v-else class="assist-ref-placeholder" />
                          </div>
                          <div class="assist-ref-caption">{{
                            refImage.caption
                          }}</div>
                        </div>
                      </div>
                    </div>

                    <a-input-number
                      v-for="field in assistAmountFields"
                      :key="field.key"
                      v-model="form.assistAmounts[field.key]"
                      :min="0"
                      hide-button
                      placeholder="金额"
                      class="assist-amount-input"
                    >
                      <template #prefix>{{ field.label }}</template>
                      <template #suffix>{{ currency }}</template>
                    </a-input-number>

                    <div
                      v-for="(item, index) in form.assistRepeatAmounts"
                      :key="item.uid"
                      class="bonus-tier-row"
                    >
                      <a-input-number
                        v-model="item.amount"
                        :min="0"
                        hide-button
                        placeholder="金额"
                        class="assist-amount-input"
                      >
                        <template #prefix>再次助力金额</template>
                        <template #suffix>{{ currency }}</template>
                      </a-input-number>
                      <a-button
                        type="text"
                        status="danger"
                        shape="circle"
                        @click="removeAssistRepeatAmount(index)"
                      >
                        <template #icon><icon-minus-circle /></template>
                      </a-button>
                    </div>
                    <a-button
                      type="primary"
                      class="bonus-tier-add-btn"
                      @click="addAssistRepeatAmount"
                    >
                      <template #icon><icon-plus /></template>
                    </a-button>
                  </div>
                </a-form-item>

                <!-- 完成所需人数（助力） -->
                <a-form-item
                  v-if="form.type === ACTIVITY_TYPE.ASSIST"
                  label="完成所需人数"
                  required
                >
                  <div class="bonus-mode-group">
                    <a-input-number
                      v-model="form.assistRequiredCount"
                      :min="1"
                      :precision="0"
                      hide-button
                      placeholder="完成所需人数"
                      class="assist-amount-input"
                    >
                      <template #suffix>人</template>
                    </a-input-number>
                    <span class="assist-helper"
                      >剩余金额将按计划，给予指定人数全部领取完毕</span
                    >
                  </div>
                </a-form-item>

                <!-- 盲盒奖金（盲盒） -->
                <a-form-item
                  v-if="form.type === ACTIVITY_TYPE.BLIND_BOX"
                  label="盲盒奖金"
                  required
                >
                  <div class="bonus-mode-group">
                    <div
                      v-for="(prize, index) in form.blindBoxPrizes"
                      :key="prize.uid"
                      class="bonus-tier-row"
                    >
                      <span class="bonus-tier-index">{{ index + 1 }}</span>
                      <a-input-number
                        v-model="prize.inviteCount"
                        :min="1"
                        :precision="0"
                        hide-button
                        placeholder="累计邀请人数"
                        class="bonus-tier-input"
                      >
                        <template #suffix>人</template>
                      </a-input-number>
                      <a-input-number
                        v-model="prize.amount"
                        :min="0"
                        hide-button
                        placeholder="奖金金额"
                        class="bonus-tier-input"
                      >
                        <template #suffix>{{ currency }}</template>
                      </a-input-number>
                      <a-button
                        type="text"
                        status="danger"
                        shape="circle"
                        @click="removeBlindBoxPrize(index)"
                      >
                        <template #icon><icon-minus-circle /></template>
                      </a-button>
                    </div>
                    <a-button
                      type="primary"
                      class="bonus-tier-add-btn"
                      @click="addBlindBoxPrize"
                    >
                      <template #icon><icon-plus /></template>
                    </a-button>
                  </div>
                </a-form-item>

                <!-- 认定助力条件（助力）/ 认定有效条件（盲盒） -->
                <a-form-item
                  v-if="form.type === ACTIVITY_TYPE.ASSIST || form.type === ACTIVITY_TYPE.BLIND_BOX"
                  :label="validConditionLabel"
                  required
                >
                  <a-select
                    v-model="form.validCondition"
                    :placeholder="`请选择${validConditionLabel}`"
                    class="assist-amount-input"
                  >
                    <template #prefix>有效条件</template>
                    <a-option
                      v-for="opt in validConditionOptions"
                      :key="opt"
                      :value="opt"
                    >
                      {{ opt }}
                    </a-option>
                  </a-select>
                </a-form-item>

                <template v-if="form.type === ACTIVITY_TYPE.NEWCOMER">
                  <!-- 优惠类型（新人彩金） -->
                  <a-form-item label="优惠类型" required>
                    <a-select
                      v-model="form.offerType"
                      placeholder="请选择"
                      class="assist-amount-input"
                    >
                      <a-option
                        v-for="opt in offerTypeOptions"
                        :key="opt"
                        :value="opt"
                      >
                        {{ opt }}
                      </a-option>
                    </a-select>
                  </a-form-item>

                  <!-- 限制会员注册时间（新人彩金） -->
                  <a-form-item label="限制会员注册时间" required>
                    <a-range-picker
                      v-model="form.registerTimeRange"
                      show-time
                      format="YYYY-MM-DD HH:mm:ss"
                      style="width: 420px"
                    />
                  </a-form-item>

                  <!-- 优惠金额（新人彩金） -->
                  <a-form-item label="优惠金额" required>
                    <div class="bonus-mode-group">
                      <a-radio-group v-model="form.offerAmountMode">
                        <a-radio value="fixed">固定彩金</a-radio>
                        <a-radio value="random">随机彩金</a-radio>
                      </a-radio-group>
                      <div
                        v-for="(item, index) in form.offerAmounts"
                        :key="item.uid"
                        class="bonus-tier-row"
                      >
                        <div
                          v-if="form.offerAmountMode === 'random'"
                          class="offer-range-box"
                        >
                          <span class="offer-range-affix"
                            >第{{ index + 1 }}天</span
                          >
                          <a-input-number
                            v-model="item.min"
                            :min="0"
                            :precision="2"
                            hide-button
                            placeholder="0.00"
                            class="offer-range-input"
                          />
                          <span class="offer-range-affix">-</span>
                          <a-input-number
                            v-model="item.max"
                            :min="0"
                            :precision="2"
                            hide-button
                            placeholder="0.00"
                            class="offer-range-input"
                          />
                          <span class="offer-range-affix">{{ currency }}</span>
                        </div>
                        <a-input-number
                          v-else
                          v-model="item.amount"
                          :min="0"
                          :precision="2"
                          hide-button
                          placeholder="金额"
                          class="offer-amount-input"
                        >
                          <template #prefix>第{{ index + 1 }}天</template>
                          <template #suffix>{{ currency }}</template>
                        </a-input-number>
                        <a-button
                          type="text"
                          status="danger"
                          shape="circle"
                          @click="removeOfferAmount(index)"
                        >
                          <template #icon><icon-minus-circle /></template>
                        </a-button>
                      </div>
                      <a-button
                        type="primary"
                        class="bonus-tier-add-btn"
                        :disabled="
                          form.offerAmountMode !== 'fixed' &&
                          form.offerAmountMode !== 'random'
                        "
                        @click="addOfferAmount"
                      >
                        <template #icon><icon-plus /></template>
                      </a-button>
                    </div>
                  </a-form-item>

                  <!-- 指定代理（新人彩金） -->
                  <a-form-item label="指定代理">
                    <div class="agent-id-group">
                      <a-switch v-model="form.specifiedAgent" />
                      <template v-if="form.specifiedAgent">
                        <a-textarea
                          v-model="form.agentIds"
                          :error="Boolean(agentIdsError)"
                          placeholder="请输入ID，多ID请使用逗号隔开，最多支持200个"
                          :auto-size="{ minRows: 5, maxRows: 8 }"
                        />
                        <div v-if="agentIdsError" class="agent-id-error">
                          {{ agentIdsError }}
                        </div>
                      </template>
                    </div>
                  </a-form-item>
                </template>

                <!-- 统计亏盈（救援金） -->
                <a-form-item v-if="showProfitStat" label="统计亏盈" required>
                  <a-radio-group v-model="form.profitStatMode">
                    <a-radio value="含优惠金">含优惠金</a-radio>
                    <a-radio value="不含优惠金">不含优惠金</a-radio>
                  </a-radio-group>
                </a-form-item>

                <template v-if="form.type === ACTIVITY_TYPE.SIGN_IN">
                  <!-- 参与条件（签到） -->
                  <a-form-item label="参与条件">
                    <a-radio-group v-model="form.signInCondition">
                      <a-radio value="全部会员">全部会员</a-radio>
                      <a-radio value="新注册会员"
                        >新注册会员（活动有效期内注册）</a-radio
                      >
                    </a-radio-group>
                  </a-form-item>

                  <!-- 签到方式（签到） -->
                  <a-form-item label="签到方式">
                    <a-radio-group v-model="form.signInMode">
                      <a-radio value="连续签到">连续签到</a-radio>
                      <a-radio value="累计签到">
                        累计签到
                        <a-tooltip>
                          <template #content>
                            <div>连续签到：断签后，需要从第一天重新开始。</div>
                            <div>累计签到：断签后，可从原有天数继续签到。</div>
                          </template>
                          <icon-info-circle class="claim-mode-tip" />
                        </a-tooltip>
                      </a-radio>
                    </a-radio-group>
                  </a-form-item>

                  <!-- 签到周期（签到） -->
                  <a-form-item label="签到周期">
                    <a-radio-group
                      v-model="form.signInCycle"
                      class="sign-in-inline"
                    >
                      <a-radio :value="7">7天</a-radio>
                      <a-radio :value="15">15天</a-radio>
                      <a-radio :value="30">30天</a-radio>
                      <a-radio value="custom">
                        <a-input-number
                          v-model="form.signInCustomDays"
                          :min="1"
                          :max="365"
                          :precision="0"
                          hide-button
                          class="sign-in-days-input"
                          @focus="form.signInCycle = 'custom'"
                        >
                          <template #append>天</template>
                        </a-input-number>
                      </a-radio>
                    </a-radio-group>
                  </a-form-item>

                  <!-- 强制重置日（签到） -->
                  <a-form-item label="强制重置日">
                    <div class="sign-in-inline">
                      <a-radio-group v-model="form.signInResetMode">
                        <a-radio value="不重置">不重置</a-radio>
                        <a-radio value="每月">每月</a-radio>
                      </a-radio-group>
                      <a-input-number
                        v-model="form.signInResetDay"
                        :min="1"
                        :max="31"
                        :precision="0"
                        hide-button
                        :disabled="form.signInResetMode !== '每月'"
                        class="sign-in-days-input"
                      >
                        <template #append>日</template>
                      </a-input-number>
                      <span>强制重置</span>
                    </div>
                  </a-form-item>

                  <!-- 奖励方式（签到） -->
                  <a-form-item label="奖励方式">
                    <div class="bonus-mode-group">
                      <a-radio-group v-model="form.signInRewardMode">
                        <a-radio value="按阶梯奖励">按阶梯奖励</a-radio>
                        <a-radio value="VIP等级奖励">VIP等级奖励</a-radio>
                      </a-radio-group>
                      <div
                        v-for="table in signInTables"
                        :key="table.key"
                        :class="{ 'sign-in-vip-group': table.group }"
                      >
                        <div v-if="table.group" class="sign-in-vip-header">
                          <span>VIP等级：</span>
                          <a-select
                            v-model="table.group.levels"
                            multiple
                            placeholder="请选择"
                            class="sign-in-vip-select"
                          >
                            <a-option
                              v-for="level in vipLevels"
                              :key="level"
                              :value="level"
                              :disabled="
                                isSignInVipLevelUsed(table.group, level)
                              "
                            >
                              VIP{{ level }}
                            </a-option>
                          </a-select>
                          <a-link
                            status="success"
                            @click="
                              table.group.collapsed = !table.group.collapsed
                            "
                          >
                            {{ table.group.collapsed ? '展开' : '收缩' }}
                          </a-link>
                          <a-button
                            shape="circle"
                            size="small"
                            @click="addSignInVipGroup(table.group)"
                          >
                            +
                          </a-button>
                          <a-button
                            shape="circle"
                            size="small"
                            :disabled="form.signInVipGroups.length <= 1"
                            @click="removeSignInVipGroup(table.group)"
                          >
                            -
                          </a-button>
                        </div>
                        <table
                          v-show="!table.group?.collapsed"
                          class="sign-in-table"
                        >
                          <thead>
                            <tr>
                              <th>天</th>
                              <th>类型</th>
                              <th>奖励金额</th>
                              <th>完成充值金额</th>
                              <th>完成有效投注</th>
                              <th>额外奖励</th>
                            </tr>
                          </thead>
                          <tbody>
                            <tr v-for="row in table.rows" :key="row.day">
                              <td>{{ row.day }}</td>
                              <td>
                                <a-select
                                  v-model="row.type"
                                  class="sign-in-type"
                                >
                                  <a-option value="固定">固定</a-option>
                                  <a-option value="随机">随机</a-option>
                                </a-select>
                              </td>
                              <td>
                                <div
                                  v-if="row.type === '随机'"
                                  class="sign-in-range"
                                >
                                  <a-input-number
                                    v-model="row.min"
                                    :min="0"
                                    hide-button
                                    placeholder="最小"
                                  />
                                  <span>~</span>
                                  <a-input-number
                                    v-model="row.max"
                                    :min="0"
                                    hide-button
                                    placeholder="最大"
                                  />
                                </div>
                                <a-input-number
                                  v-else
                                  v-model="row.amount"
                                  :min="0"
                                  hide-button
                                  placeholder="100"
                                />
                              </td>
                              <td>
                                <a-input-number
                                  v-model="row.recharge"
                                  :min="0"
                                  hide-button
                                  placeholder="50"
                                />
                              </td>
                              <td>
                                <a-input-number
                                  v-model="row.bet"
                                  :min="0"
                                  hide-button
                                  placeholder="20"
                                />
                              </td>
                              <td>
                                <a-input-number
                                  v-model="row.extra"
                                  :min="0"
                                  hide-button
                                  placeholder="0"
                                />
                              </td>
                            </tr>
                          </tbody>
                        </table>
                      </div>
                    </div>
                  </a-form-item>
                </template>

                <!-- 代理层级（全民代理） -->
                <a-form-item v-if="form.type === ACTIVITY_TYPE.AGENT" label="代理层级">
                  <span class="currency-text">无限层级</span>
                </a-form-item>

                <!-- 会员层级 -->
                <a-form-item
                  v-if="!memberLevelsHidden"
                  label="会员层级"
                  :class="{ 'settings-item-early': memberLevelsEarly }"
                >
                  <div class="checkbox-stack">
                    <a-checkbox-group
                      v-model="form.memberLevels"
                      :disabled="memberLevelsLocked"
                    >
                      <div
                        class="checkbox-row full-row"
                        @click="toggleMemberLevels"
                      >
                        <a-checkbox value="全部">全部</a-checkbox>
                      </div>
                      <template
                        v-for="group in memberLevelGroups"
                        :key="group.label"
                      >
                        <div v-if="group.options.length" class="checkbox-row">
                          <span class="level-group-label">{{
                            group.label
                          }}</span>
                          <a-checkbox
                            v-for="level in group.options"
                            :key="level.value"
                            :value="level.value"
                          >
                            {{ level.label }}
                          </a-checkbox>
                        </div>
                      </template>
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
                <!-- 领取模式（与奖金模式相邻，便于看清联动） -->
                <a-form-item v-if="showClaimMode" label="领取模式" required>
                  <a-radio-group
                    v-model="form.claimMode"
                    class="wrap-radio-group"
                  >
                    <a-radio
                      v-for="opt in getEnumOptions('bet_claim_mode')"
                      :key="opt.value"
                      :value="opt.value"
                    >
                      {{ opt.label }}
                      <a-tooltip
                        v-if="opt.value === BET_CLAIM_MODE.CUMULATIVE"
                        content="按活动期间内的累计有效投注额计算奖励"
                      >
                        <icon-info-circle-fill class="claim-mode-tip" />
                      </a-tooltip>
                    </a-radio>
                  </a-radio-group>
                </a-form-item>

                <!-- 奖金模式 -->
                <a-form-item
                  v-if="showBonusMode"
                  :label="bonusConfig.label"
                  required
                >
                  <div class="bonus-mode-group">
                    <a-radio-group v-model="form.bonusMode">
                      <a-radio value="fixed" :disabled="isBonusFixedDisabled">{{
                        getEnumLabel('bonus_mode', BONUS_MODE.FIXED_LADDER_MAX)
                      }}</a-radio>
                      <a-radio value="ratio">{{
                        getEnumLabel('bonus_mode', BONUS_MODE.RATE_LADDER_MAX)
                      }}</a-radio>
                    </a-radio-group>
                    <div v-if="isBonusFixedDisabled" class="inline-helper">
                      累计领取仅支持比例金额
                    </div>

                    <template v-if="form.bonusMode">
                      <div
                        v-for="(tier, index) in form.bonusTiers"
                        :key="tier.uid"
                        class="bonus-tier-row"
                      >
                        <span class="bonus-tier-index">{{ index + 1 }}</span>
                        <a-input-number
                          v-model="tier.threshold"
                          :min="0"
                          hide-button
                          :error="bonusValidated && tier.threshold == null"
                          :placeholder="bonusConfig.thresholdLabel"
                          class="bonus-tier-input"
                        />
                        <a-input-number
                          v-model="tier.reward"
                          :min="0"
                          hide-button
                          :error="bonusValidated && tier.reward == null"
                          :placeholder="bonusRewardLabel"
                          class="bonus-tier-input"
                        >
                          <template v-if="form.bonusMode === 'ratio'" #suffix
                            >%</template
                          >
                        </a-input-number>
                        <a-button
                          type="text"
                          status="danger"
                          shape="circle"
                          :disabled="
                            form.bonusTiers.length <= bonusConfig.minTiers
                          "
                          @click="removeBonusTier(index)"
                        >
                          <template #icon><icon-minus-circle /></template>
                        </a-button>
                      </div>
                      <a-button
                        type="primary"
                        class="bonus-tier-add-btn"
                        @click="addBonusTier"
                      >
                        <template #icon><icon-plus /></template>
                      </a-button>
                      <div
                        v-if="bonusValidated && bonusTierError"
                        class="agent-id-error"
                      >
                        {{ bonusTierError }}
                      </div>
                    </template>
                  </div>
                </a-form-item>

                <template v-if="form.type === ACTIVITY_TYPE.LUCKY_BET">
                  <!-- 奖励模式（幸运注单） -->
                  <a-form-item label="奖励模式" required>
                    <div class="bonus-mode-group">
                      <a-radio-group v-model="form.luckyRewardMode">
                        <a-radio value="充值获得次数">充值获得次数</a-radio>
                        <a-radio value="洗码获得次数">洗码获得次数</a-radio>
                      </a-radio-group>
                      <div class="bonus-tier-row">
                        <a-input-number
                          v-model="form.luckyAmountPerTime"
                          :min="0"
                          hide-button
                          :placeholder="
                            form.luckyRewardMode === '洗码获得次数'
                              ? '当日洗码金额/次'
                              : '当日充值金额/次'
                          "
                          class="bonus-tier-input"
                        />
                        <a-input-number
                          v-model="form.luckyMaxTimes"
                          :min="1"
                          :precision="0"
                          hide-button
                          placeholder="最多获得次数"
                          class="bonus-tier-input"
                        />
                      </div>
                    </div>
                  </a-form-item>

                  <!-- 总领取次数（幸运注单） -->
                  <a-form-item label="总领取次数" required>
                    <a-input-number
                      v-model="form.luckyTotalClaimTimes"
                      :min="1"
                      :precision="0"
                      hide-button
                      placeholder="请输入"
                      class="assist-amount-input"
                    >
                      <template #suffix>次</template>
                    </a-input-number>
                  </a-form-item>

                  <!-- 有效注单条件（幸运注单） -->
                  <a-form-item label="有效注单条件" required>
                    <div class="bonus-mode-group">
                      <a-input-number
                        v-model="form.luckyValidBetAmount"
                        :min="0"
                        hide-button
                        placeholder="有效注单条件金额"
                        class="assist-amount-input"
                      />
                      <div class="lucky-profit-row">
                        <span>是否盈利</span>
                        <a-radio-group v-model="form.luckyProfitMode">
                          <a-radio value="不限制">不限制</a-radio>
                          <a-radio value="盈利">盈利</a-radio>
                          <a-radio value="亏损">亏损</a-radio>
                        </a-radio-group>
                      </div>
                    </div>
                  </a-form-item>

                  <!-- 订单中奖方式（幸运注单） -->
                  <a-form-item label="订单中奖方式" required>
                    <div class="bonus-mode-group">
                      <a-radio-group v-model="form.luckyWinMode">
                        <a-radio
                          v-for="opt in luckyWinModeOptions"
                          :key="opt.value"
                          :value="opt.value"
                        >
                          {{ opt.value }}
                        </a-radio>
                      </a-radio-group>
                      <div class="lucky-win-tips">
                        <div v-for="tip in currentLuckyWinMode.tips" :key="tip">
                          {{ tip }}
                        </div>
                      </div>
                      <div
                        v-for="(row, index) in form.luckyWinRows"
                        :key="row.uid"
                        class="bonus-tier-row"
                      >
                        <span class="bonus-tier-index">{{ index + 1 }}</span>
                        <a-input
                          :model-value="row.number"
                          :max-length="8"
                          :error="
                            luckyValidated && isLuckyRowNumberInvalid(row)
                          "
                          :placeholder="currentLuckyWinMode.placeholder"
                          class="bonus-tier-input"
                          @update:model-value="
                            row.number = $event.replace(/\D/g, '')
                          "
                        />
                        <a-input-number
                          v-model="row.reward"
                          :min="0"
                          hide-button
                          :error="luckyValidated && row.reward == null"
                          placeholder="奖励金额"
                          class="bonus-tier-input"
                        >
                          <template #suffix>{{ currency }}</template>
                        </a-input-number>
                        <a-button
                          type="text"
                          status="danger"
                          shape="circle"
                          @click="removeLuckyWinRow(index)"
                        >
                          <template #icon><icon-minus-circle /></template>
                        </a-button>
                      </div>
                      <a-button
                        type="primary"
                        class="bonus-tier-add-btn"
                        @click="addLuckyWinRow"
                      >
                        <template #icon><icon-plus /></template>
                      </a-button>
                      <div
                        v-if="luckyValidated && luckyWinError"
                        class="agent-id-error"
                      >
                        {{ luckyWinError }}
                      </div>
                    </div>
                  </a-form-item>
                </template>

                <template v-if="form.type === ACTIVITY_TYPE.RED_PACKET">
                  <!-- 派奖次数（红包雨） -->
                  <a-form-item label="派奖次数">
                    <div class="red-wave-list">
                      <div class="red-wave-row red-wave-head">
                        <span class="red-wave-hour">派发奖励时间</span>
                        <span class="red-wave-duration">持续时间（分）</span>
                      </div>
                      <div
                        v-for="(wave, index) in form.redPacketWaves"
                        :key="wave.uid"
                        class="red-wave-row"
                      >
                        <a-select v-model="wave.hour" class="red-wave-hour">
                          <a-option
                            v-for="hour in redPacketHourOptions"
                            :key="hour"
                            :value="hour"
                          >
                            {{ hour }} o'clock
                          </a-option>
                        </a-select>
                        <a-input-number
                          v-model="wave.duration"
                          :min="1"
                          :max="59"
                          :precision="0"
                          class="red-wave-duration"
                        >
                          <template #append>min</template>
                        </a-input-number>
                        <div class="red-wave-actions">
                          <a-button
                            v-if="index === form.redPacketWaves.length - 1"
                            type="primary"
                            size="small"
                            @click="addRedPacketWave"
                          >
                            +
                          </a-button>
                          <a-button
                            v-if="index > 0"
                            type="primary"
                            size="small"
                            @click="removeRedPacketWave(index)"
                          >
                            -
                          </a-button>
                        </div>
                      </div>
                    </div>
                  </a-form-item>

                  <!-- 奖励限制（红包雨） -->
                  <a-form-item label="奖励限制">
                    <div class="red-limit-list">
                      <template
                        v-for="item in redPacketLimitFields"
                        :key="item.key"
                      >
                        <span class="red-limit-label">{{ item.label }}</span>
                        <a-input-number
                          v-model="form.redPacketLimits[item.key]"
                          :min="0"
                          :precision="item.integer ? 0 : undefined"
                          hide-button
                          class="red-limit-input"
                        />
                      </template>
                    </div>
                  </a-form-item>

                  <!-- 有效用户（红包雨） -->
                  <a-form-item label="有效用户">
                    <div class="bonus-mode-group">
                      <a-radio-group v-model="form.redPacketValidUser">
                        <a-radio
                          v-for="opt in redPacketValidUserOptions"
                          :key="opt"
                          :value="opt"
                        >
                          {{ opt }}
                        </a-radio>
                      </a-radio-group>
                      <template v-if="redPacketValidUserNeedsThreshold">
                        <a-input-number
                          v-model="form.redPacketValidDeposit"
                          :min="0"
                          hide-button
                          class="assist-amount-input"
                          :error="
                            redPacketValidated &&
                            !(form.redPacketValidDeposit > 0)
                          "
                        >
                          <template #prefix>充值金额</template>
                        </a-input-number>
                        <a-input-number
                          v-model="form.redPacketValidBet"
                          :min="0"
                          hide-button
                          class="assist-amount-input"
                          :error="
                            redPacketValidated && !(form.redPacketValidBet > 0)
                          "
                        >
                          <template #prefix>有效洗码</template>
                        </a-input-number>
                        <div
                          v-if="
                            redPacketValidated && redPacketValidUserError
                          "
                          class="agent-id-error"
                        >
                          {{ redPacketValidUserError }}
                        </div>
                      </template>
                    </div>
                  </a-form-item>

                  <!-- 奖励类型（红包雨）；无限制只需固定金额，不展示 -->
                  <a-form-item
                    v-if="!redPacketRewardIsUnlimited"
                    label="奖励类型"
                    required
                  >
                    <a-radio-group v-model="form.redPacketRewardType">
                      <a-radio
                        v-for="opt in redPacketRewardTypeOptions"
                        :key="opt"
                        :value="opt"
                      >
                        {{ opt }}
                      </a-radio>
                    </a-radio-group>
                  </a-form-item>

                  <!-- 奖励金额（红包雨）；无限制只填一笔固定金额 -->
                  <a-form-item label="奖励金额">
                    <div class="bonus-mode-group">
                      <a-radio-group v-model="form.redPacketRewardBasis">
                        <a-radio
                          v-for="opt in redPacketRewardBasisOptions"
                          :key="opt.value"
                          :value="opt.value"
                        >
                          {{ opt.value }}
                        </a-radio>
                      </a-radio-group>
                      <div
                        v-if="redPacketBasis.conditionLabel"
                        class="red-history-row"
                      >
                        <span>历史时间</span>
                        <a-select
                          v-model="form.redPacketHistoryTime"
                          class="red-history-select"
                        >
                          <a-option
                            v-for="opt in redPacketHistoryOptions"
                            :key="opt"
                            :value="opt"
                          >
                            {{ opt }}
                          </a-option>
                        </a-select>
                      </div>
                      <div
                        v-for="(row, index) in visibleRedPacketRewardRows"
                        :key="row.uid"
                        class="red-reward-row"
                      >
                        <div class="red-reward-index-row">
                          <span class="bonus-tier-index">{{ index + 1 }}</span>
                          <a-button
                            v-if="index > 0"
                            type="text"
                            status="danger"
                            shape="circle"
                            @click="removeRedPacketRewardRow(index)"
                          >
                            <template #icon><icon-minus-circle /></template>
                          </a-button>
                        </div>
                        <a-input-number
                          v-if="redPacketBasis.conditionLabel"
                          v-model="row.condition"
                          :min="redPacketBasis.allowNegative ? undefined : 0"
                          hide-button
                          :error="redPacketValidated && row.condition == null"
                          :placeholder="redPacketBasis.conditionLabel"
                          class="red-reward-input"
                          :class="{
                            'red-reward-input--wide':
                              redPacketBasis.allowNegative,
                          }"
                        />
                        <template v-if="redPacketRewardIsRandom">
                          <a-input-number
                            v-model="row.min"
                            :min="0"
                            hide-button
                            :error="redPacketValidated && row.min == null"
                            :placeholder="`最小${redPacketRewardUnitLabel}`"
                            class="red-reward-input"
                          >
                            <template v-if="redPacketRewardIsRatio" #suffix
                              >%</template
                            >
                          </a-input-number>
                          <a-input-number
                            v-model="row.max"
                            :min="0"
                            hide-button
                            :error="redPacketValidated && row.max == null"
                            :placeholder="`最大${redPacketRewardUnitLabel}`"
                            class="red-reward-input"
                          >
                            <template v-if="redPacketRewardIsRatio" #suffix
                              >%</template
                            >
                          </a-input-number>
                        </template>
                        <a-input-number
                          v-else
                          v-model="row.reward"
                          :min="0"
                          hide-button
                          :error="redPacketValidated && row.reward == null"
                          :placeholder="
                            redPacketRewardIsUnlimited
                              ? '固定金额'
                              : form.redPacketRewardType
                          "
                          class="red-reward-input"
                        >
                          <template v-if="redPacketRewardIsRatio" #suffix
                            >%</template
                          >
                        </a-input-number>
                      </div>
                      <a-button
                        v-if="redPacketBasis.conditionLabel"
                        type="primary"
                        class="bonus-tier-add-btn"
                        @click="addRedPacketRewardRow"
                      >
                        <template #icon><icon-plus /></template>
                      </a-button>
                      <div
                        v-if="redPacketValidated && redPacketRewardError"
                        class="agent-id-error"
                      >
                        {{ redPacketRewardError }}
                      </div>
                    </div>
                  </a-form-item>
                </template>

                <!-- 出款限制 -->
                <a-form-item v-if="!isFieldHidden('withdraw')" label="出款限制">
                  <div class="time-group">
                    <a-input
                      v-model="form.withdrawMultiple"
                      placeholder="请输入洗码倍数(输入0代表不需要洗码)"
                      style="width: 400px"
                    >
                      <template #suffix>
                        <span class="inline-helper">洗码倍数</span>
                      </template>
                    </a-input>
                    <a-radio-group v-model="form.giftMode">
                      <a-radio
                        v-for="opt in getEnumOptions('wagering_base')"
                        :key="opt.value"
                        :value="opt.value"
                        :disabled="isFieldOptionDisabled('giftMode', opt.value)"
                        >{{ opt.label }}</a-radio
                      >
                    </a-radio-group>
                  </div>
                </a-form-item>

                <!-- 指定洗码游戏 -->
                <a-form-item v-if="!isFieldHidden('specifiedGame')">
                  <template #label>
                    指定洗码游戏
                    <a-tooltip
                      content="会员领取的赠送奖励，需在指定游戏中完成洗码（有效投注）要求；不开启则全部游戏均可完成洗码"
                    >
                      <icon-info-circle class="claim-mode-tip" />
                    </a-tooltip>
                  </template>
                  <div class="specified-game-section">
                    <div class="specified-game-header">
                      <div>
                        <div class="specified-game-title">按平台定向选择</div>
                        <div class="specified-game-desc">
                          开启后，赠送奖励仅能通过所选游戏的有效投注完成洗码；关闭则不限游戏。
                        </div>
                      </div>
                      <a-switch v-model="form.specifiedGame" />
                    </div>

                    <GamePicker
                      v-if="form.specifiedGame"
                      v-model="form.gameIds"
                      class="specified-game-picker"
                    />
                  </div>
                </a-form-item>

                <!-- 参与客户端 -->
                <a-form-item v-if="!clientListHidden" label="参与客户端">
                  <a-checkbox-group v-model="form.clientList">
                    <div
                      class="checkbox-row full-row"
                      @click="toggleClientList"
                    >
                      <a-checkbox :value="ALL_OPTION">全部</a-checkbox>
                    </div>
                    <div class="checkbox-row">
                      <a-checkbox
                        v-for="opt in getEnumOptions('client_type')"
                        :key="opt.value"
                        :value="opt.value"
                        >{{ opt.label }}</a-checkbox
                      >
                    </div>
                  </a-checkbox-group>
                </a-form-item>

                <!-- 参与活动限制 -->
                <a-form-item v-if="!restrictionsHidden" label="参与活动限制">
                  <a-checkbox-group v-model="form.activityRestrictions">
                    <div class="checkbox-row">
                      <a-checkbox
                        v-for="opt in getEnumOptions('apply_limit_type')"
                        :key="opt.value"
                        :value="opt.value"
                        :disabled="
                          isFieldOptionDisabled(
                            'activityRestrictions',
                            opt.value
                          )
                        "
                      >
                        {{ opt.label }}
                      </a-checkbox>
                    </div>
                  </a-checkbox-group>
                </a-form-item>

                <!-- 领取奖励限制 -->
                <a-form-item v-if="!restrictionsHidden" label="领取奖励限制">
                  <a-checkbox-group v-model="form.claimRestrictions">
                    <div class="checkbox-row">
                      <a-checkbox
                        v-for="opt in getEnumOptions('claim_limit_type')"
                        :key="opt.value"
                        :value="opt.value"
                        :disabled="
                          isFieldOptionDisabled('claimRestrictions', opt.value)
                        "
                      >
                        {{ opt.label }}
                      </a-checkbox>
                    </div>
                  </a-checkbox-group>
                </a-form-item>
              </div>
            </div>
          </section>
        </a-col>

        <!-- 右侧表单区域 -->
        <a-col :xs="24" :xl="12" class="form-panel-right">
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
              <span>4. 活动内容</span>
              <icon-down
                class="section-title-icon"
                :class="{ collapsed: !isSectionExpanded('rules') }"
              />
            </div>
            <div v-show="isSectionExpanded('rules')" class="section-card">
              <div class="section-body">
                <!-- 内容图片（图文(广告)）：每组按语言分别上传图片并配置点击跳转 -->
                <a-form-item v-if="form.type === ACTIVITY_TYPE.IMAGE_AD" label="内容图片">
                  <div class="bonus-mode-group">
                    <div
                      v-for="(group, groupIndex) in form.contentImageGroups"
                      :key="group.uid"
                      class="content-image-group"
                    >
                      <div class="content-image-card">
                        <div class="content-image-header">
                          <span>内容图片 #{{ groupIndex + 1 }}</span>
                          <a-link
                            :disabled="
                              !getContentImageAvailableLangs(group).length
                            "
                            @click="addContentImageItem(group)"
                          >
                            新增语言
                          </a-link>
                        </div>
                        <div
                          v-for="(item, itemIndex) in group.items"
                          :key="item.uid"
                          class="content-image-item"
                        >
                          <div class="content-image-lang-row">
                            <a-select
                              v-model="item.lang"
                              placeholder="请选择语言"
                              class="content-image-lang"
                            >
                              <a-option
                                v-for="opt in langOptions"
                                :key="opt.value"
                                :value="opt.value"
                                :disabled="
                                  isContentImageLangUsed(
                                    group,
                                    opt.value,
                                    itemIndex
                                  )
                                "
                              >
                                {{ opt.label }}
                              </a-option>
                            </a-select>
                            <a-link
                              status="danger"
                              :disabled="group.items.length <= 1"
                              @click="removeContentImageItem(group, itemIndex)"
                            >
                              删除当前语言
                            </a-link>
                          </div>
                          <a-upload
                            v-model:file-list="item.files"
                            list-type="picture-card"
                            accept="image/*"
                            :auto-upload="false"
                            :limit="1"
                          />
                          <div class="content-image-link-label">链接模块</div>
                          <a-select
                            v-model="item.link"
                            class="content-image-link"
                            @change="resetContentImageLinkValue(item)"
                          >
                            <a-option
                              v-for="opt in contentImageLinkOptions"
                              :key="opt.value"
                              :value="opt.value"
                            >
                              {{ opt.value }}
                            </a-option>
                          </a-select>
                          <a-select
                            v-if="
                              getContentImageLinkOption(item.link).kind ===
                              'activity'
                            "
                            v-model="item.linkValue"
                            allow-search
                            placeholder="请选择活动"
                            :options="contentImageActivityOptions"
                            class="content-image-link"
                          />
                          <a-select
                            v-else-if="
                              getContentImageLinkOption(item.link).kind ===
                              'message'
                            "
                            v-model="item.linkValue"
                            allow-search
                            placeholder="请选择消息"
                            :options="contentImageMessageOptions"
                            class="content-image-link"
                          />
                          <a-input
                            v-else-if="
                              getContentImageLinkOption(item.link).kind ===
                              'fixed'
                            "
                            :model-value="item.linkValue"
                            disabled
                            class="content-image-link"
                          />
                          <a-input
                            v-else-if="
                              getContentImageLinkOption(item.link).kind ===
                              'text'
                            "
                            v-model="item.linkValue"
                            placeholder="请输入内容"
                            class="content-image-link"
                          />
                        </div>
                      </div>
                      <a-button
                        type="text"
                        status="danger"
                        shape="circle"
                        @click="removeContentImageGroup(groupIndex)"
                      >
                        <template #icon><icon-minus-circle /></template>
                      </a-button>
                    </div>
                    <a-button
                      type="primary"
                      class="bonus-tier-add-btn"
                      @click="addContentImageGroup"
                    >
                      <template #icon><icon-plus /></template>
                    </a-button>
                    <div
                      v-if="contentImageValidated && contentImageError"
                      class="agent-id-error"
                    >
                      {{ contentImageError }}
                    </div>
                  </div>
                </a-form-item>

                <!-- 活动名称与规则：按语言成套配置 -->
                <div class="locale-block">
                  <div class="locale-block-title">
                    活动名称、图片与规则<span class="locale-required">*</span>
                  </div>

                  <a-tabs
                    v-model:active-key="activeLocaleUid"
                    type="card-gutter"
                    editable
                    show-add-button
                    class="locale-tabs"
                    @add="addLocale"
                    @delete="removeLocale"
                  >
                    <a-tab-pane
                      v-for="(item, index) in form.locales"
                      :key="item.uid"
                      :closable="form.locales.length > 1"
                    >
                      <template #title>
                        <span class="locale-tab-title">
                          {{ getLangLabel(item.lang) }}
                          <span
                            v-if="isLocaleIncomplete(item)"
                            class="locale-tab-dot"
                            title="未填写完整"
                          />
                        </span>
                      </template>

                      <div class="locale-card-body">
                        <div class="locale-field">
                          <span class="locale-field-label">语言</span>
                          <a-select
                            v-model="item.lang"
                            class="locale-lang-select"
                            placeholder="请选择语言"
                            @change="fillLocalePreset(item)"
                          >
                            <a-option
                              v-for="opt in langOptions"
                              :key="opt.value"
                              :value="opt.value"
                              :disabled="isLangUsed(opt.value, index)"
                            >
                              {{ opt.label }}
                            </a-option>
                          </a-select>
                        </div>
                        <div class="locale-field">
                          <span class="locale-field-label">活动名称</span>
                          <a-input
                            v-model="item.name"
                            placeholder="请输入活动名称，最多 60 个字符"
                            :max-length="60"
                            show-word-limit
                          />
                        </div>
                        <div class="locale-field">
                          <span
                            class="locale-field-label locale-field-label--optional"
                            >副标题</span
                          >
                          <a-input
                            v-model="item.description"
                            placeholder="请输入副标题，最多 1024 个字符"
                            :max-length="1024"
                            show-word-limit
                          />
                        </div>
                        <!-- 活动图片随语言页签成套配置，语言与名称、规则始终一致 -->
                        <div class="locale-field locale-field--top">
                          <span
                            class="locale-field-label locale-field-label--optional"
                            >活动图片</span
                          >
                          <div class="cover-image-item">
                            <a-upload
                              :auto-upload="false"
                              :show-file-list="false"
                              :disabled="item.imageUploading"
                              accept="image/jpeg,image/png,image/gif"
                              @change="
                                (_, fileItem) =>
                                  uploadLocaleImage(item, fileItem)
                              "
                            >
                              <template #upload-button>
                                <div class="cover-image-box">
                                  <a-spin v-if="item.imageUploading" />
                                  <template v-else-if="item.image">
                                    <img :src="item.image" alt="" />
                                    <span class="cover-image-mask"
                                      >更换图片</span
                                    >
                                  </template>
                                  <div v-else class="cover-image-empty">
                                    <icon-plus />
                                    <span>上传图片</span>
                                  </div>
                                </div>
                              </template>
                            </a-upload>
                            <div class="upload-tip">
                              尺寸 668px * 250px，JPG、JPEG、PNG、GIF
                            </div>
                          </div>
                        </div>
                        <div class="locale-field locale-field--top">
                          <span class="locale-field-label">活动规则</span>
                          <!-- QuillEditor 会把工具栏渲染为编辑区的兄弟节点，需包一层占据 grid 单列 -->
                          <div class="quill-editor-wrap">
                            <QuillEditor
                              v-model:content="item.content"
                              theme="snow"
                              :toolbar="quillToolbar"
                              content-type="html"
                              class="quill-editor"
                            />
                          </div>
                        </div>
                      </div>
                    </a-tab-pane>
                  </a-tabs>
                </div>
              </div>
            </div>
          </section>
        </a-col>
      </a-row>
    </a-form>

    <div class="dialog-footer-actions">
      <a-button class="cancel-btn" @click="requestClose">取消</a-button>
      <a-button
        type="primary"
        class="save-draft-btn"
        :loading="saving"
        :disabled="preparing"
        @click="handleSaveDraft"
      >
        保存草稿
      </a-button>
    </div>
  </a-modal>
</template>

<script setup lang="ts">
  import { ref, reactive, computed, watch, nextTick } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { DATE_TIME, siteNow } from '../site-time';
  import { Message, Modal } from '@arco-design/web-vue';
  import type { FileItem } from '@arco-design/web-vue';
  import {
    IconClose,
    IconDown,
    IconRight,
    IconPlus,
    IconMinusCircle,
    IconInfoCircleFill,
    IconInfoCircle,
    IconImage,
  } from '@arco-design/web-vue/es/icon';
  import { QuillEditor } from '@vueup/vue-quill';
  import '@vueup/vue-quill/dist/vue-quill.snow.css';
  import { queryEnabledLangs } from '@/api/i18n';
  import { getCachedOperatorCode } from '@/utils/operator-site';
  import { queryVipLevelList } from '@/api/member';
  import { uploadOssFile } from '@/api/oss';
  import {
    ACTIVITY_APPLY_LIMIT as APPLY_LIMIT,
    ACTIVITY_APPLY_MODE as APPLY_MODE,
    ACTIVITY_BET_CLAIM_MODE as BET_CLAIM_MODE,
    ACTIVITY_BONUS_MODE as BONUS_MODE,
    ACTIVITY_CLAIM_LIMIT as CLAIM_LIMIT,
    ACTIVITY_CYCLE_MODE as LOOP_MODE,
    ACTIVITY_DEPOSIT_REWARD as RECHARGE_TYPE,
    ACTIVITY_PAYOUT_MODE as PAYOUT_MODE,
    ACTIVITY_PAYOUT_TIMING as PAYOUT_TIMING,
    ACTIVITY_SECTION,
    ACTIVITY_TYPE,
    ACTIVITY_WAGERING_BASE as WAGERING_BASE,
    createActivity,
    getActivityDetail,
    queryActivityList,
    updateActivity,
    type ActivityDetail,
  } from '@/api/activity';
  import { getActivityLocalePreset } from '../locale-presets';
  import { useActivityStore } from '../store';
  import { enumLabel, enumOptions, type EnumParamGroup } from '../options';
  import buildCreatePayload, {
    ALL_OPTION,
    buildUpdatePayload,
    detailToForm,
    validateCreateForm,
    type CreateFormVisibility,
    type FormSection,
  } from '../create-payload';
  import GamePicker from './game-picker.vue';

  const props = defineProps<{
    visible: boolean;
    /** 传入时为编辑该草稿活动，否则为新增 */
    activityId?: number | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    /** 新增或编辑保存成功，参数为活动 ID */
    (e: 'saved', id: number): void;
  }>();

  const visible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value),
  });

  type AccordionSection = FormSection;

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

  const { t, locale: uiLocale } = useI18n();

  const { autoLevels, fixedLevels, loadLevels, enumParams, loadEnumParams } =
    useActivityStore();

  /** 枚举选项和文案来自枚举参数接口 */
  const getEnumOptions = (group: EnumParamGroup) =>
    enumOptions(enumParams.value, group);

  const getEnumLabel = (group: EnumParamGroup, value: number) =>
    enumLabel(enumParams.value, group, value);

  /** 活动类型按钮来自枚举参数接口 */
  const typeOptions = computed(() => getEnumOptions('type'));

  /** 本站结算币种，金额输入框的单位 */
  const currency = computed(() => enumParams.value?.currency || '');

  const isEdit = computed(() => props.activityId != null);

  /** 值带 auto: / fixed: 前缀，自动层级与固定层级的 ID 可能重复 */
  const memberLevelGroups = computed(() => [
    {
      label: '自动层级',
      options: autoLevels.value.map((level) => ({
        value: `auto:${level.id}`,
        label: level.name,
      })),
    },
    {
      label: '固定层级',
      options: fixedLevels.value.map((level) => ({
        value: `fixed:${level.id}`,
        label: level.name,
      })),
    },
  ]);

  const memberLevelOptions = computed(() =>
    memberLevelGroups.value.flatMap((group) =>
      group.options.map((option) => option.value)
    )
  );

  const clientCodes = computed(() =>
    getEnumOptions('client_type').map((option) => option.value)
  );

  const defaultLoopOptions: number[] = [
    LOOP_MODE.NONE,
    LOOP_MODE.DAILY,
    LOOP_MODE.WEEKLY,
    LOOP_MODE.MONTHLY,
    LOOP_MODE.INFINITE,
  ];

  type InviteeRuleKey = 'recharge' | 'turnover' | 'sameIp' | 'sameDevice';

  const inviteeRuleOptions: {
    key: InviteeRuleKey;
    label: string;
    /** 为空时是金额，显示本站币种 */
    unit: string;
    tip: string;
  }[] = [
    {
      key: 'recharge',
      label: '充值达到',
      unit: '',
      tip: '被邀请会员累计充值达到该金额，才算一次有效邀请',
    },
    {
      key: 'turnover',
      label: '有效流水达到',
      unit: '',
      tip: '被邀请会员累计有效投注达到该金额，才算一次有效邀请',
    },
    {
      key: 'sameIp',
      label: '同IP只允许',
      unit: '人',
      tip: '同一 IP 下最多计入的有效被邀请人数',
    },
    {
      key: 'sameDevice',
      label: '同设备号只允许',
      unit: '人',
      tip: '同一设备号下最多计入的有效被邀请人数',
    },
  ];

  const createInviteeRules = () =>
    Object.fromEntries(
      inviteeRuleOptions.map((rule) => [
        rule.key,
        { checked: false, value: undefined as number | undefined },
      ])
    ) as Record<InviteeRuleKey, { checked: boolean; value?: number }>;

  /** 各活动类型均暂不支持的领取限制 */
  const unavailableClaimLimits: number[] = [
    CLAIM_LIMIT.BANK_CARD,
    CLAIM_LIMIT.REAL_NAME,
  ];

  /** 不少类型都不允许按设备、IP 限制参与 */
  const deviceAndIpLimits: number[] = [APPLY_LIMIT.DEVICE, APPLY_LIMIT.IP];

  const manualPayoutModes: number[] = [
    PAYOUT_MODE.MANUAL_EXPIRE_VOID,
    PAYOUT_MODE.MANUAL_EXPIRE_AUTO,
  ];

  /**
   * 每种充值类型能选的循环模式，不在 allowed 里的选项置灰；
   * 切换到该充值类型时，循环模式自动选中 default
   */
  const rechargeLoopRules: Record<
    number,
    { allowed: number[]; default: number }
  > = {
    [RECHARGE_TYPE.FIRST]: {
      allowed: [LOOP_MODE.NONE],
      default: LOOP_MODE.NONE,
    },
    [RECHARGE_TYPE.SECOND]: {
      allowed: [LOOP_MODE.NONE],
      default: LOOP_MODE.NONE,
    },
    [RECHARGE_TYPE.CUMULATIVE]: {
      allowed: [LOOP_MODE.DAILY],
      default: LOOP_MODE.DAILY,
    },
    [RECHARGE_TYPE.SINGLE]: {
      allowed: [
        LOOP_MODE.NONE,
        LOOP_MODE.DAILY,
        LOOP_MODE.WEEKLY,
        LOOP_MODE.MONTHLY,
        LOOP_MODE.INFINITE,
      ],
      default: LOOP_MODE.DAILY,
    },
  };

  type HideableField =
    | 'specifiedGame'
    | 'participation'
    | 'distribution'
    | 'withdraw';

  type TypePreset = {
    showLoopMode: boolean;
    /** 该类型不展示的通用表单项 */
    hiddenFields?: HideableField[];
    /** 会员层级排在活动时间之后、参与模式之前 */
    memberLevelsEarly?: boolean;
    /** 自动派发时也展示领取过期天数 */
    alwaysShowExpireDays?: boolean;
    /** 统计亏盈（救援金） */
    showProfitStat?: boolean;
    /** 奖金模式标题，不传时为「奖金模式」 */
    bonusLabel?: string;
    /** 阶梯达标金额的提示文字，不传时为「达标充值金额」 */
    bonusThresholdLabel?: string;
    /** 固定金额模式奖励值的提示文字，不传时为「奖励固定金额」 */
    bonusFixedRewardLabel?: string;
    /** 比例金额模式奖励值的提示文字，不传时为「返奖比例」 */
    bonusRatioRewardLabel?: string;
    /** 奖金模式默认值，不传时默认不选 */
    bonusMode?: string;
    /** 阶梯最少行数，不传时为 1（选中模式后自动加第一行） */
    minBonusTiers?: number;
    /** 不传时为「循环模式」 */
    loopLabel?: string;
    /** 会员层级固定为全部，不可修改 */
    lockMemberLevels?: boolean;
    /** 不展示参与活动限制与领取奖励限制 */
    hideRestrictions?: boolean;
    /** 不传时使用 defaultLoopOptions */
    loopOptions?: number[];
    /** 循环模式在该类型下的特殊叫法（语言包 key），不传时用通用叫法 */
    loopOptionLabels?: Record<number, string>;
    showBonusMode: boolean;
    showClaimMode: boolean;
    /** 以下默认值和禁用项均为接口枚举编码 */
    claimMode: number | undefined;
    giftMode: number;
    disabledGiftValues: number[];
    participationMode: number;
    loopMode: number | undefined;
    distributionMode: number;
    distributionTime: number;
    rechargeType: number | undefined;
    disabledParticipationValues: number[];
    disabledLoopValues: number[];
    disabledDistributionValues: number[];
    disabledDistributionTimeValues: number[];
    disabledRestrictionValues: number[];
  };

  /**
   * 大多数类型共用的配置；
   * typePresets 里每个类型展开它，只写和它不一样的字段
   */
  const defaultPreset: TypePreset = {
    showLoopMode: true,
    showBonusMode: false,
    showClaimMode: false,
    claimMode: undefined,
    giftMode: WAGERING_BASE.BONUS_ONLY,
    disabledGiftValues: [WAGERING_BASE.PRINCIPAL_AND_BONUS],
    participationMode: APPLY_MODE.AUTO,
    loopMode: LOOP_MODE.NONE,
    distributionMode: PAYOUT_MODE.AUTO,
    distributionTime: PAYOUT_TIMING.REALTIME,
    rechargeType: RECHARGE_TYPE.FIRST,
    disabledParticipationValues: [APPLY_MODE.MANUAL],
    disabledLoopValues: [
      LOOP_MODE.DAILY,
      LOOP_MODE.WEEKLY,
      LOOP_MODE.MONTHLY,
      LOOP_MODE.INFINITE,
    ],
    disabledDistributionValues: [],
    disabledDistributionTimeValues: [],
    disabledRestrictionValues: deviceAndIpLimits,
  };

  const typePresets: Record<number, TypePreset> = {
    [ACTIVITY_TYPE.DEPOSIT]: {
      ...defaultPreset,
      showBonusMode: true,
      giftMode: WAGERING_BASE.PRINCIPAL_AND_BONUS,
      disabledGiftValues: [],
      disabledLoopValues: [],
    },
    [ACTIVITY_TYPE.BET]: {
      ...defaultPreset,
      showLoopMode: false,
      showClaimMode: true,
      showBonusMode: true,
      claimMode: BET_CLAIM_MODE.EACH,
      bonusMode: 'fixed',
      bonusThresholdLabel: '达标有效投注额',
      bonusFixedRewardLabel: '奖励固定金额',
      bonusRatioRewardLabel: '返奖比例',
      participationMode: APPLY_MODE.MANUAL,
      loopMode: undefined,
      distributionMode: PAYOUT_MODE.MANUAL_EXPIRE_VOID,
      rechargeType: undefined,
      disabledParticipationValues: [APPLY_MODE.AUTO],
      disabledLoopValues: defaultLoopOptions,
      disabledRestrictionValues: [],
    },
    [ACTIVITY_TYPE.REFER]: {
      ...defaultPreset,
      loopOptions: [LOOP_MODE.NONE, LOOP_MODE.DAILY, LOOP_MODE.WEEKLY],
      loopOptionLabels: {
        [LOOP_MODE.NONE]:
          'discountCenter.activityCenter.create.loopMode.noReset',
      },
      disabledLoopValues: defaultLoopOptions,
      disabledDistributionTimeValues: [PAYOUT_TIMING.NEXT_DAY],
    },
    [ACTIVITY_TYPE.LUCK_WHEEL]: {
      ...defaultPreset,
      disabledLoopValues: [LOOP_MODE.INFINITE],
    },
    [ACTIVITY_TYPE.ASSIST]: defaultPreset,
    [ACTIVITY_TYPE.BLIND_BOX]: defaultPreset,
    [ACTIVITY_TYPE.NEWCOMER]: {
      ...defaultPreset,
      showLoopMode: false,
      distributionMode: PAYOUT_MODE.MANUAL_EXPIRE_AUTO,
      disabledLoopValues: defaultLoopOptions,
    },
    [ACTIVITY_TYPE.AGENT]: {
      ...defaultPreset,
      loopLabel: '循环模式(返奖周期)',
      lockMemberLevels: true,
      hideRestrictions: true,
    },
    [ACTIVITY_TYPE.RESCUE]: {
      ...defaultPreset,
      showBonusMode: true,
      showProfitStat: true,
      loopMode: LOOP_MODE.DAILY,
      disabledLoopValues: [
        LOOP_MODE.NONE,
        LOOP_MODE.MONTHLY,
        LOOP_MODE.INFINITE,
      ],
      disabledDistributionTimeValues: [PAYOUT_TIMING.NEXT_DAY],
      bonusLabel: '奖励模式',
      bonusThresholdLabel: '损失金额',
      bonusFixedRewardLabel: '奖励金额',
      bonusRatioRewardLabel: '奖励比例',
      bonusMode: 'fixed',
      minBonusTiers: 0,
    },
    [ACTIVITY_TYPE.LUCKY_BET]: {
      ...defaultPreset,
      disabledDistributionTimeValues: [PAYOUT_TIMING.NEXT_DAY],
    },
    [ACTIVITY_TYPE.VIP]: {
      ...defaultPreset,
      showLoopMode: false,
      hideRestrictions: true,
      disabledDistributionTimeValues: [PAYOUT_TIMING.NEXT_DAY],
    },
    [ACTIVITY_TYPE.RED_PACKET]: {
      ...defaultPreset,
      memberLevelsEarly: true,
      alwaysShowExpireDays: true,
      loopOptions: [LOOP_MODE.NONE, LOOP_MODE.DAILY, LOOP_MODE.WEEKLY],
      loopMode: LOOP_MODE.DAILY,
      disabledLoopValues: [LOOP_MODE.NONE, LOOP_MODE.WEEKLY],
      disabledDistributionValues: manualPayoutModes,
      disabledDistributionTimeValues: [PAYOUT_TIMING.NEXT_DAY],
    },
    [ACTIVITY_TYPE.IMAGE_AD]: {
      ...defaultPreset,
      showLoopMode: false,
      giftMode: WAGERING_BASE.PRINCIPAL_AND_BONUS,
      disabledGiftValues: [],
      disabledLoopValues: defaultLoopOptions,
      disabledDistributionValues: manualPayoutModes,
      disabledRestrictionValues: [],
      hideRestrictions: true,
      hiddenFields: [
        'specifiedGame',
        'participation',
        'distribution',
        'withdraw',
      ],
    },
    [ACTIVITY_TYPE.SIGN_IN]: {
      ...defaultPreset,
      loopOptions: [LOOP_MODE.NONE, LOOP_MODE.INFINITE],
      loopOptionLabels: {
        [LOOP_MODE.INFINITE]:
          'discountCenter.activityCenter.create.loopMode.loop',
      },
      disabledLoopValues: [],
      disabledDistributionValues: manualPayoutModes,
      disabledDistributionTimeValues: [PAYOUT_TIMING.NEXT_DAY],
    },
  };

  /** 没有专属预设的类型（如后端新增的类型）按充值的表单展示 */
  const presetOf = (type: number) =>
    typePresets[type] ?? typePresets[ACTIVITY_TYPE.DEPOSIT];

  type LangOption = { value: string; label: string };

  const langOptions = ref<LangOption[]>([]);
  let langOptionsLoaded = false;

  /** 成功后复用，失败的下次打开弹窗重试 */
  const loadLangOptions = async () => {
    if (langOptionsLoaded) return;
    const operatorCode = getCachedOperatorCode();
    try {
      const { data } = await queryEnabledLangs({
        params: operatorCode ? { operator_code: operatorCode } : undefined,
      });
      langOptions.value = [...(data?.list || [])]
        .filter((item) => !item.disabled)
        .sort((a, b) => (a.sort_no ?? a.id) - (b.sort_no ?? b.id))
        .map((item) => ({
          value: item.lang,
          label: item.i18n_name || item.name,
        }));
      langOptionsLoaded = true;
    } catch {
      // 失败时按界面语言兜底，提示由全局拦截器弹出
    }
  };

  /**
   * 优先当前后台界面语言，与活动列表接口的 lang 一致，保证列表能取到名称；
   * 分站未启用该语言时取排序第一的语言
   */
  const defaultLang = computed(() => {
    if (!langOptions.value.length) return uiLocale.value;
    return (
      langOptions.value.find((opt) => opt.value === uiLocale.value)?.value ??
      langOptions.value[0].value
    );
  });

  type ActivityLocale = {
    uid: number;
    lang: string;
    name: string;
    /** 副标题，对应 i18n.description */
    description: string;
    content: string;
    /** 已上传的活动图片地址，选图后立即上传 */
    image: string;
    imageUploading: boolean;
  };

  let localeUid = 0;

  /**
   * kind 决定第二个控件：activity / message 为下拉选择，fixed 为只读固定值，
   * text 为自由输入，none 不需要跳转目标
   */
  const contentImageLinkOptions: {
    value: string;
    kind: 'none' | 'activity' | 'message' | 'fixed' | 'text';
    fixedValue?: string;
  }[] = [
    { value: '无法点击', kind: 'none' },
    { value: '活动ID', kind: 'activity' },
    { value: '消息ID', kind: 'message' },
    { value: '充值模块（输入0）', kind: 'fixed', fixedValue: '0' },
    { value: '会员模块（输入1）', kind: 'fixed', fixedValue: '1' },
    { value: '其他模块', kind: 'text' },
  ];

  /** 分站暂无推送消息接口，先用固定数据占位 */
  const contentImageMessages = [
    { id: 211, title: '测试极光推播4(关闭应用程式不关闭浏览器)' },
    { id: 206, title: '测试极光推播3(关闭浏览器)' },
    { id: 205, title: '测试极光推播11091750' },
    { id: 204, title: '5555252' },
  ];

  const contentImageMessageOptions = contentImageMessages.map((item) => ({
    value: String(item.id),
    label: `${item.title} (ID:${item.id})`,
  }));

  type ContentImageItem = {
    uid: number;
    lang: string;
    files: FileItem[];
    link: string;
    linkValue: string;
  };

  type ContentImageGroup = { uid: number; items: ContentImageItem[] };

  let contentImageUid = 0;

  const createContentImageItem = (lang: string): ContentImageItem => {
    contentImageUid += 1;
    return {
      uid: contentImageUid,
      lang,
      files: [],
      link: '无法点击',
      linkValue: '',
    };
  };

  const createContentImageGroup = (): ContentImageGroup => {
    contentImageUid += 1;
    return {
      uid: contentImageUid,
      items: [createContentImageItem(defaultLang.value)],
    };
  };

  type BonusTier = {
    uid: number;
    /** 达标充值金额 */
    threshold?: number;
    /** 固定金额模式为奖励金额，比例模式为返奖百分比 */
    reward?: number;
  };

  let bonusTierUid = 0;

  const WHEEL_BY_BET = '达标有效投注赠送1次机会';
  const WHEEL_BY_DEPOSIT = '达标充值金额赠送1次机会';

  type WheelTriggerTier = {
    uid: number;
    /** 要求有效投注 / 充值金额 */
    threshold?: number;
  };

  let wheelTriggerTierUid = 0;

  const createWheelTriggerTier = (): WheelTriggerTier => {
    wheelTriggerTierUid += 1;
    return { uid: wheelTriggerTierUid, threshold: undefined };
  };

  const wheelPrizeTypeOptions = ['固定奖金', '比例奖金', '谢谢参与'];

  type WheelPrize = {
    uid: number;
    prizeType: string;
    /** 固定奖金为金额，比例奖金为百分比 */
    amount?: number;
    /** 中奖概率（%） */
    probability?: number;
  };

  /** src 为空时显示占位图 */
  const assistReferenceImages: { caption: string; src?: string }[] = [
    { caption: '助力总金额' },
    { caption: '进度条展示金额' },
    { caption: '第一次助力金额' },
  ];

  type AssistAmountKey = 'total' | 'progress' | 'first';

  const assistAmountFields: { key: AssistAmountKey; label: string }[] = [
    { key: 'total', label: '助力总金额' },
    { key: 'progress', label: '进度条展示金额' },
    { key: 'first', label: '第一次助力金额' },
  ];

  const validConditionOptions = ['仅注册', '注册并充值'];

  type BlindBoxPrize = {
    uid: number;
    /** 累计邀请人数 */
    inviteCount?: number;
    /** 奖金金额（本站币种） */
    amount?: number;
  };

  let blindBoxPrizeUid = 0;

  const redPacketHourOptions = Array.from({ length: 24 }, (_, hour) => hour);

  type RedPacketWave = { uid: number; hour: number; duration?: number };

  let redPacketWaveUid = 0;

  const createRedPacketWave = (hour = 0): RedPacketWave => {
    redPacketWaveUid += 1;
    return { uid: redPacketWaveUid, hour, duration: 59 };
  };

  type RedPacketLimitKey =
    | 'waveTotalDisplay'
    | 'waveTotalActual'
    | 'singleDisplay'
    | 'singleActual'
    | 'dailyClaimTimes';

  const redPacketLimitFields: {
    key: RedPacketLimitKey;
    label: string;
    integer?: boolean;
  }[] = [
    { key: 'waveTotalDisplay', label: '每波派奖总额上限(前端显示)' },
    { key: 'waveTotalActual', label: '每波派奖总额上限(实际生效)' },
    { key: 'singleDisplay', label: '单红包金额上限(前端显示)' },
    { key: 'singleActual', label: '单红包金额上限(实际生效)' },
    { key: 'dailyClaimTimes', label: '单用户每日领取次数上限', integer: true },
  ];

  const redPacketValidUserOptions = [
    '无限制',
    '符合一个即可',
    '全部满足',
    '已充值会员可领取',
  ];

  const redPacketRewardTypeOptions = [
    '固定金额',
    '随机金额',
    '固定比例',
    '随机比例',
  ];

  /** conditionLabel 为空时不需要填门槛和历史时间，且只保留一行奖励 */
  const redPacketRewardBasisOptions = [
    { value: '充值', conditionLabel: '充值金额', allowNegative: false },
    {
      value: '盈亏',
      conditionLabel: '盈亏（正数盈利负数亏损）',
      allowNegative: true,
    },
    { value: '无限制', conditionLabel: '', allowNegative: false },
    { value: '有效投注', conditionLabel: '有效投注', allowNegative: false },
  ];

  const redPacketHistoryOptions = [
    '当日',
    '本周',
    '本月',
    '昨日',
    '上周',
    '上月',
    '近七天(含当日)',
  ];

  type RedPacketRewardRow = {
    uid: number;
    /** 达标条件金额（充值 / 盈亏 / 有效投注） */
    condition?: number;
    /** 固定金额或固定比例 */
    reward?: number;
    /** 随机金额或随机比例的区间 */
    min?: number;
    max?: number;
  };

  let redPacketRewardRowUid = 0;

  const createRedPacketRewardRow = (): RedPacketRewardRow => {
    redPacketRewardRowUid += 1;
    return { uid: redPacketRewardRowUid };
  };

  type SignInRewardRow = {
    day: number;
    type: string;
    /** 固定类型的奖励金额 */
    amount?: number;
    /** 随机类型的奖励区间 */
    min?: number;
    max?: number;
    /** 完成充值金额 */
    recharge?: number;
    /** 完成有效投注 */
    bet?: number;
    extra?: number;
  };

  /** VIP 等级奖励的表格默认填 0，阶梯奖励只给提示文字 */
  const createSignInRewards = (
    days: number,
    startDay = 1,
    fillZero = false
  ): SignInRewardRow[] =>
    Array.from({ length: days }, (_, index) => ({
      day: startDay + index,
      type: '固定',
      ...(fillZero ? { amount: 0, recharge: 0, bet: 0, extra: 0 } : {}),
    }));

  const resizeSignInRewards = (
    rows: SignInRewardRow[],
    days: number,
    fillZero: boolean
  ) => {
    if (rows.length > days) {
      rows.splice(days);
    } else {
      rows.push(
        ...createSignInRewards(days - rows.length, rows.length + 1, fillZero)
      );
    }
  };

  type SignInVipGroup = {
    uid: number;
    levels: number[];
    collapsed: boolean;
    rows: SignInRewardRow[];
  };

  let signInVipGroupUid = 0;

  const createSignInVipGroup = (days: number): SignInVipGroup => {
    signInVipGroupUid += 1;
    return {
      uid: signInVipGroupUid,
      levels: [],
      collapsed: false,
      rows: createSignInRewards(days, 1, true),
    };
  };

  type VipConfigKey =
    | 'dailyTurnover'
    | 'dailyReward'
    | 'weeklyTurnover'
    | 'weeklyReward'
    | 'monthlyTurnover'
    | 'monthlyReward';

  const vipConfigColumns: { key: VipConfigKey; title: string }[] = [
    { key: 'dailyTurnover', title: '日奖励所需流水' },
    { key: 'dailyReward', title: '日奖励' },
    { key: 'weeklyTurnover', title: '周奖励所需流水' },
    { key: 'weeklyReward', title: '周奖励' },
    { key: 'monthlyTurnover', title: '月奖励所需流水' },
    { key: 'monthlyReward', title: '月奖励' },
  ];

  const luckyWinModeOptions = [
    {
      value: '尾号(后八位)',
      placeholder: '8位数字',
      tips: ['填入8位0-9的一串数字，当注单的后8位与该串数字完全一致时，即中奖'],
      isValid: (num: string) => /^\d{8}$/.test(num),
      invalidText: '需为8位数字',
    },
    {
      value: '递增(后八位)',
      placeholder: '递增数字串',
      tips: [
        '递增数字串最少2位，最多8位。比如：01、234、6789；',
        '每个数字串后一个数字比前一个数字多1；',
        '数字串从注单的右边开始匹配；',
        '每条订单仅能匹配全部数字串的其中1个。',
      ],
      isValid: (num: string) =>
        /^\d{2,8}$/.test(num) &&
        [...num].every(
          (digit, i) => i === 0 || Number(digit) === Number(num[i - 1]) + 1
        ),
      invalidText: '需为2-8位递增数字串',
    },
    {
      value: '连号(后八位)',
      placeholder: '连号数字串',
      tips: [
        '连号数字串最少2位，最多8位，每位数字均相同。比如：33、777、9999；',
        '注单的后8位和连号数字串的关系是包含，可中奖多次。比如：一条注单为79348478888812，注单的后8位为：78888812。如果输入：88奖励10U，则该条订单可中奖4次。',
      ],
      isValid: (num: string) =>
        /^\d{2,8}$/.test(num) && [...num].every((digit) => digit === num[0]),
      invalidText: '需为2-8位相同数字',
    },
  ];

  type LuckyWinRow = {
    uid: number;
    /** 中奖数字串，格式随订单中奖方式而定 */
    number: string;
    reward?: number;
  };

  let luckyWinRowUid = 0;

  const agentCommissionLevels = [1, 2, 3];

  /** 第 n 级的上级收益公式，如 *1级抽成*2级抽成 */
  const getAgentFormula = (level: number) =>
    agentCommissionLevels
      .slice(0, level)
      .map((item) => `*${item}级抽成`)
      .join('');

  const activityConditionOptions = ['无条件', '完成首充', '完成手机验证'];

  const offerTypeOptions = ['注册彩金', '首充彩金'];

  type OfferAmount = {
    uid: number;
    /** 固定彩金金额 */
    amount?: number;
    /** 随机彩金区间 */
    min?: number;
    max?: number;
  };

  let offerAmountUid = 0;

  const getCurrentMonthRange = () => [
    siteNow().startOf('month').format(DATE_TIME),
    siteNow().endOf('month').format(DATE_TIME),
  ];

  let assistRepeatUid = 0;

  const createAssistRepeatAmount = () => {
    assistRepeatUid += 1;
    return { uid: assistRepeatUid, amount: undefined as number | undefined };
  };

  let wheelPrizeUid = 0;

  const createWheelPrize = (): WheelPrize => {
    wheelPrizeUid += 1;
    return { uid: wheelPrizeUid, prizeType: '固定奖金' };
  };

  const createWheelPrizes = (size: number) =>
    Array.from({ length: size }, createWheelPrize);

  const createLocale = (type: number, lang: string): ActivityLocale => {
    localeUid += 1;
    return {
      uid: localeUid,
      lang,
      ...getActivityLocalePreset(type, lang, presetOf(type).rechargeType),
      description: '',
      image: '',
      imageUploading: false,
    };
  };

  const createDefaultLocales = (type: number): ActivityLocale[] => [
    createLocale(type, defaultLang.value),
  ];

  const createDefaultForm = (type: number = ACTIVITY_TYPE.DEPOSIT) => ({
    type,
    locales: createDefaultLocales(type),
    showInList: true,
    section: ACTIVITY_SECTION.LOBBY,
    timeType: 'permanent',
    timeRange: [],
    specifiedGame: false,
    gameIds: [] as number[],
    rechargeType: presetOf(type).rechargeType,
    loopMode: presetOf(type).loopMode,
    participationMode: presetOf(type).participationMode,
    distributionMode: presetOf(type).distributionMode,
    distributionTime: presetOf(type).distributionTime,
    claimExpireDays: 1,
    claimMode: presetOf(type).claimMode,
    memberLevels: ['全部', ...memberLevelOptions.value],
    bonusMode: typePresets[type]?.bonusMode ?? '',
    profitStatMode: '含优惠金',
    withdrawMultiple: '',
    giftMode: presetOf(type).giftMode,
    clientList: [ALL_OPTION] as (number | typeof ALL_OPTION)[],
    activityRestrictions: [] as number[],
    claimRestrictions: [] as number[],
    bonusTiers: [] as BonusTier[],
    referralRewardMode: '',
    referralSingleAmount: undefined as number | undefined,
    inviteeRules: createInviteeRules(),
    wheelSize: 6,
    wheelPrizes: createWheelPrizes(6),
    drawConditions: [] as string[],
    wheelBetTiers: [] as WheelTriggerTier[],
    wheelDepositTiers: [] as WheelTriggerTier[],
    extraConditions: [] as string[],
    assistAmounts: {} as Partial<Record<AssistAmountKey, number>>,
    assistRepeatAmounts: [createAssistRepeatAmount()],
    assistRequiredCount: undefined as number | undefined,
    validCondition: undefined as string | undefined,
    blindBoxPrizes: [] as BlindBoxPrize[],
    activityCondition: '无条件',
    offerType: undefined as string | undefined,
    registerTimeRange: getCurrentMonthRange(),
    offerAmountMode: '',
    offerAmounts: [] as OfferAmount[],
    specifiedAgent: false,
    agentIds: '',
    agentRewardMode: '3级抽成',
    agentCommissionRates: [] as (number | undefined)[],
    luckyRewardMode: '充值获得次数',
    luckyAmountPerTime: undefined as number | undefined,
    luckyMaxTimes: undefined as number | undefined,
    luckyTotalClaimTimes: undefined as number | undefined,
    luckyValidBetAmount: undefined as number | undefined,
    luckyProfitMode: '亏损',
    luckyWinMode: '尾号(后八位)',
    luckyWinRows: [] as LuckyWinRow[],
    sidebarEntry: true,
    vipUniqueness: '系统默认',
    vipClaimRule: '跨天、跨周、跨月',
    vipConfigs: {} as Record<number, Partial<Record<VipConfigKey, number>>>,
    redPacketWaves: [createRedPacketWave()],
    redPacketLimits: {} as Partial<Record<RedPacketLimitKey, number>>,
    redPacketValidUser: '无限制',
    redPacketValidDeposit: undefined as number | undefined,
    redPacketValidBet: undefined as number | undefined,
    redPacketRewardType: '固定金额',
    redPacketRewardBasis: '充值',
    redPacketHistoryTime: '当日',
    redPacketRewardRows: [createRedPacketRewardRow()],
    contentImageGroups: [createContentImageGroup()],
    signInCondition: '全部会员',
    signInMode: '累计签到',
    signInCycle: 7 as number | 'custom',
    signInCustomDays: 100 as number | undefined,
    signInResetMode: '不重置',
    signInResetDay: 1 as number | undefined,
    signInRewardMode: '按阶梯奖励',
    signInRewards: createSignInRewards(7),
    signInVipGroups: [createSignInVipGroup(7)],
  });

  const form = reactive(createDefaultForm());

  /** 回填详情期间暂停切换类型重置表单、切换奖金模式清空奖励的联动 */
  let applyingDetail = false;

  const bonusConfig = computed(() => {
    const preset = presetOf(form.type);
    return {
      label: preset.bonusLabel ?? '奖金模式',
      thresholdLabel: preset.bonusThresholdLabel ?? '达标充值金额',
      fixedRewardLabel: preset.bonusFixedRewardLabel ?? '奖励固定金额',
      ratioRewardLabel: preset.bonusRatioRewardLabel ?? '返奖比例',
      minTiers: preset.minBonusTiers ?? 1,
    };
  });

  const bonusRewardLabel = computed(() =>
    form.bonusMode === 'ratio'
      ? bonusConfig.value.ratioRewardLabel
      : bonusConfig.value.fixedRewardLabel
  );

  /** 来自 VIP 管理的等级编号，按升序 */
  const vipLevels = ref<number[]>([]);

  /** 表格里每个可编辑等级都需要有对应对象，v-model 才能写入 */
  const ensureVipConfigs = () => {
    vipLevels.value.forEach((level) => {
      if (level > 0 && !form.vipConfigs[level]) form.vipConfigs[level] = {};
    });
  };

  let vipLevelsLoaded = false;

  /** 成功后复用，失败的下次打开弹窗重试 */
  const loadVipLevels = async () => {
    if (vipLevelsLoaded) return;
    try {
      const { data } = await queryVipLevelList({ page: 1, page_size: 100 });
      vipLevels.value = (data?.list || []).map((item) => item.level);
      vipLevelsLoaded = true;
    } catch {
      vipLevels.value = [];
    }
  };

  /** 点过保存后才展示阶梯的必填校验 */
  const bonusValidated = ref(false);

  const currentLuckyWinMode = computed(
    () =>
      luckyWinModeOptions.find((opt) => opt.value === form.luckyWinMode) ??
      luckyWinModeOptions[0]
  );

  /** 点过保存后才展示订单中奖方式的校验 */
  const luckyValidated = ref(false);

  /** 点过保存后才展示红包雨奖励金额的校验 */
  const redPacketValidated = ref(false);

  /** 点过保存后才展示内容图片的校验 */
  const contentImageValidated = ref(false);

  const COVER_IMAGE_WIDTH = 668;
  const COVER_IMAGE_HEIGHT = 250;
  const COVER_IMAGE_TYPES = ['image/jpeg', 'image/png', 'image/gif'];

  const readImageSize = (file: File) =>
    new Promise<{ width: number; height: number } | null>((resolve) => {
      const url = URL.createObjectURL(file);
      const image = new Image();
      image.onload = () => {
        resolve({ width: image.naturalWidth, height: image.naturalHeight });
        URL.revokeObjectURL(url);
      };
      image.onerror = () => {
        resolve(null);
        URL.revokeObjectURL(url);
      };
      image.src = url;
    });

  /** 尺寸不符只提示不拦截，类型不符直接拒绝 */
  const uploadLocaleImage = async (
    item: ActivityLocale,
    fileItem: FileItem
  ) => {
    const { file } = fileItem;
    if (!file) return;
    if (!COVER_IMAGE_TYPES.includes(file.type)) {
      Message.warning('仅支持 JPG、JPEG、PNG、GIF 格式的图片');
      return;
    }
    const size = await readImageSize(file);
    if (
      size &&
      (size.width !== COVER_IMAGE_WIDTH || size.height !== COVER_IMAGE_HEIGHT)
    ) {
      Message.warning(
        `建议尺寸 ${COVER_IMAGE_WIDTH}×${COVER_IMAGE_HEIGHT}，当前图片为 ${size.width}×${size.height}`
      );
    }
    item.imageUploading = true;
    try {
      const data = new FormData();
      data.append('file', file);
      data.append('scene', 'public');
      const { data: result } = await uploadOssFile(data);
      item.image = result.url;
    } catch {
      // 失败提示由全局拦截器弹出，保留原图
    } finally {
      item.imageUploading = false;
    }
  };

  const getContentImageLinkOption = (link: string) =>
    contentImageLinkOptions.find((opt) => opt.value === link) ??
    contentImageLinkOptions[0];

  const ACTIVITY_OPTION_PAGE_SIZE = 100;

  /** 可跳转的活动，取最新的前 100 个 */
  const contentImageActivityOptions = ref<{ value: string; label: string }[]>(
    []
  );
  let activityOptionsLoaded = false;

  const loadActivityOptions = async () => {
    if (activityOptionsLoaded) return;
    activityOptionsLoaded = true;
    try {
      const { data } = await queryActivityList({
        page: 1,
        page_size: ACTIVITY_OPTION_PAGE_SIZE,
        lang: uiLocale.value,
      });
      contentImageActivityOptions.value = (data?.list || []).map((item) => ({
        value: String(item.id),
        label: `${item.name} (ID:${item.id})`,
      }));
    } catch {
      activityOptionsLoaded = false;
    }
  };

  const resetContentImageLinkValue = (item: ContentImageItem) => {
    item.linkValue = getContentImageLinkOption(item.link).fixedValue ?? '';
  };

  const isContentImageLangUsed = (
    group: ContentImageGroup,
    lang: string,
    selfIndex: number
  ) =>
    group.items.some(
      (item, index) => index !== selfIndex && item.lang === lang
    );

  const getContentImageAvailableLangs = (group: ContentImageGroup) =>
    langOptions.value.filter(
      (opt) => !group.items.some((item) => item.lang === opt.value)
    );

  const addContentImageItem = (group: ContentImageGroup) => {
    const nextLang = getContentImageAvailableLangs(group)[0];
    if (!nextLang) return;
    group.items.push(createContentImageItem(nextLang.value));
  };

  const removeContentImageItem = (group: ContentImageGroup, index: number) => {
    if (group.items.length <= 1) return;
    group.items.splice(index, 1);
  };

  const addContentImageGroup = () => {
    form.contentImageGroups.push(createContentImageGroup());
  };

  const removeContentImageGroup = (index: number) => {
    form.contentImageGroups.splice(index, 1);
  };

  const contentImageError = computed(() => {
    let message = '';
    form.contentImageGroups.some((group, groupIndex) =>
      group.items.some((item, itemIndex) => {
        const position = `第${groupIndex + 1}组第${itemIndex + 1}行`;
        const { kind } = getContentImageLinkOption(item.link);
        const missingLink = !item.linkValue.trim();
        if (!item.lang) message = `${position}请选择语言`;
        else if (!item.files.length) message = `${position}请上传图片`;
        else if (kind === 'activity' && missingLink) {
          message = `${position}请选择活动`;
        } else if (kind === 'message' && missingLink) {
          message = `${position}请选择消息`;
        } else if (kind === 'text' && missingLink) {
          message = `${position}请输入内容`;
        }
        return Boolean(message);
      })
    );
    return message;
  });

  const isLuckyRowNumberInvalid = (row: LuckyWinRow) =>
    !currentLuckyWinMode.value.isValid(row.number);

  const luckyWinError = computed(() => {
    if (!form.luckyWinRows.length) return '请填写订单中奖方式具体规则';
    const mode = currentLuckyWinMode.value;
    const index = form.luckyWinRows.findIndex(
      (row) => isLuckyRowNumberInvalid(row) || row.reward == null
    );
    if (index < 0) return '';
    const row = form.luckyWinRows[index];
    if (!row.number) return `请填写第${index + 1}项${mode.placeholder}`;
    if (isLuckyRowNumberInvalid(row)) {
      return `第${index + 1}项${mode.invalidText}`;
    }
    return `请填写第${index + 1}项奖励金额`;
  });

  const bonusTierError = computed(() => {
    const index = form.bonusTiers.findIndex(
      (tier) => tier.threshold == null || tier.reward == null
    );
    if (index < 0) return '';
    const tier = form.bonusTiers[index];
    const label =
      tier.threshold == null
        ? bonusConfig.value.thresholdLabel
        : bonusRewardLabel.value;
    return `请填写第${index + 1}项${label}`;
  });

  const referralValidated = ref(false);

  const referralRewardError = computed(() => {
    if (!form.referralRewardMode) return '请选择奖励模式';
    if (form.referralRewardMode === '单独人数奖励') {
      return form.referralSingleAmount == null
        ? '请填写奖励模式具体金额'
        : '';
    }
    if (form.referralRewardMode === '累计人数奖励') {
      if (!form.bonusTiers.length) return '请填写奖励模式具体金额';
      const incomplete = form.bonusTiers.some(
        (tier) => tier.threshold == null || tier.reward == null
      );
      return incomplete ? '请填写奖励模式具体金额' : '';
    }
    return '';
  });

  const createBonusTier = (): BonusTier => {
    bonusTierUid += 1;
    return { uid: bonusTierUid, threshold: undefined, reward: undefined };
  };

  const addBonusTier = () => {
    form.bonusTiers.push(createBonusTier());
  };

  const removeBonusTier = (index: number) => {
    const minTiers =
      form.type === ACTIVITY_TYPE.REFER ? 1 : bonusConfig.value.minTiers;
    if (form.bonusTiers.length <= minTiers) return;
    form.bonusTiers.splice(index, 1);
  };

  watch(
    () => form.bonusMode,
    (mode, prevMode) => {
      if (!mode || applyingDetail) return;
      if (!form.bonusTiers.length) {
        if (bonusConfig.value.minTiers > 0) addBonusTier();
        return;
      }
      // 固定金额与返奖比例含义不同，切换模式时清空奖励值，保留达标金额
      if (prevMode) {
        form.bonusTiers.forEach((tier) => {
          tier.reward = undefined;
        });
      }
    }
  );

  watch(
    () => form.claimMode,
    (claimMode) => {
      if (applyingDetail || !showClaimMode.value) return;
      // 累计领取不允许固定金额，已选固定时自动切到比例
      if (
        claimMode === BET_CLAIM_MODE.CUMULATIVE &&
        form.bonusMode === 'fixed'
      ) {
        form.bonusMode = 'ratio';
      }
    }
  );

  watch(
    () => form.referralRewardMode,
    (mode) => {
      if (applyingDetail || form.type !== ACTIVITY_TYPE.REFER) return;
      referralValidated.value = false;
      if (mode === '累计人数奖励' && !form.bonusTiers.length) {
        addBonusTier();
      }
      if (mode === '单独人数奖励') {
        form.bonusTiers = [];
      }
    }
  );

  const signInDays = computed(() =>
    form.signInCycle === 'custom'
      ? form.signInCustomDays ?? 0
      : form.signInCycle
  );

  /** 阶梯奖励表与各 VIP 组的表都按签到周期的天数逐天配置 */
  watch(signInDays, (days) => {
    resizeSignInRewards(form.signInRewards, days, false);
    form.signInVipGroups.forEach((group) => {
      resizeSignInRewards(group.rows, days, true);
    });
  });

  const signInTables = computed(() =>
    form.signInRewardMode === 'VIP等级奖励'
      ? form.signInVipGroups.map((group) => ({
          key: group.uid,
          rows: group.rows,
          group,
        }))
      : [{ key: 'ladder', rows: form.signInRewards, group: null }]
  );

  const isSignInVipLevelUsed = (group: SignInVipGroup, level: number) =>
    form.signInVipGroups.some(
      (item) => item !== group && item.levels.includes(level)
    );

  const addSignInVipGroup = (after: SignInVipGroup) => {
    const index = form.signInVipGroups.indexOf(after);
    form.signInVipGroups.splice(
      index + 1,
      0,
      createSignInVipGroup(signInDays.value)
    );
  };

  const removeSignInVipGroup = (group: SignInVipGroup) => {
    if (form.signInVipGroups.length <= 1) return;
    form.signInVipGroups.splice(form.signInVipGroups.indexOf(group), 1);
  };

  watch(
    () => form.wheelSize,
    (size) => {
      if (form.wheelPrizes.length > size) {
        form.wheelPrizes.splice(size);
      } else {
        form.wheelPrizes.push(
          ...createWheelPrizes(size - form.wheelPrizes.length)
        );
      }
    }
  );

  watch(
    () => form.wheelPrizes.map((prize) => prize.prizeType),
    () => {
      form.wheelPrizes.forEach((prize) => {
        if (prize.prizeType === '谢谢参与') prize.amount = undefined;
      });
    }
  );

  const toggleDrawCondition = (value: string, checked: boolean) => {
    if (checked) {
      if (!form.drawConditions.includes(value)) {
        form.drawConditions.push(value);
      }
      if (value === WHEEL_BY_BET && !form.wheelBetTiers.length) {
        form.wheelBetTiers.push(createWheelTriggerTier());
      }
      if (value === WHEEL_BY_DEPOSIT && !form.wheelDepositTiers.length) {
        form.wheelDepositTiers.push(createWheelTriggerTier());
      }
      return;
    }
    form.drawConditions = form.drawConditions.filter((item) => item !== value);
  };

  const addWheelBetTier = () => {
    form.wheelBetTiers.push(createWheelTriggerTier());
  };

  const removeWheelBetTier = (index: number) => {
    if (form.wheelBetTiers.length <= 1) return;
    form.wheelBetTiers.splice(index, 1);
  };

  const addWheelDepositTier = () => {
    form.wheelDepositTiers.push(createWheelTriggerTier());
  };

  const removeWheelDepositTier = (index: number) => {
    if (form.wheelDepositTiers.length <= 1) return;
    form.wheelDepositTiers.splice(index, 1);
  };

  const addAssistRepeatAmount = () => {
    form.assistRepeatAmounts.push(createAssistRepeatAmount());
  };

  const removeAssistRepeatAmount = (index: number) => {
    form.assistRepeatAmounts.splice(index, 1);
  };

  const addBlindBoxPrize = () => {
    blindBoxPrizeUid += 1;
    form.blindBoxPrizes.push({ uid: blindBoxPrizeUid });
  };

  const removeBlindBoxPrize = (index: number) => {
    form.blindBoxPrizes.splice(index, 1);
  };

  const addRedPacketWave = () => {
    const lastHour = form.redPacketWaves[form.redPacketWaves.length - 1]?.hour;
    const nextHour = lastHour == null ? 0 : (lastHour + 1) % 24;
    form.redPacketWaves.push(createRedPacketWave(nextHour));
  };

  const removeRedPacketWave = (index: number) => {
    form.redPacketWaves.splice(index, 1);
  };

  const addRedPacketRewardRow = () => {
    form.redPacketRewardRows.push(createRedPacketRewardRow());
  };

  const removeRedPacketRewardRow = (index: number) => {
    form.redPacketRewardRows.splice(index, 1);
  };

  const redPacketBasis = computed(
    () =>
      redPacketRewardBasisOptions.find(
        (opt) => opt.value === form.redPacketRewardBasis
      ) ?? redPacketRewardBasisOptions[0]
  );

  /** 无限制不需要门槛，奖励只填一笔固定金额 */
  const redPacketRewardIsUnlimited = computed(
    () => !redPacketBasis.value.conditionLabel
  );

  const redPacketRewardIsRandom = computed(
    () =>
      !redPacketRewardIsUnlimited.value &&
      form.redPacketRewardType.startsWith('随机')
  );

  const redPacketRewardIsRatio = computed(
    () =>
      !redPacketRewardIsUnlimited.value &&
      form.redPacketRewardType.endsWith('比例')
  );

  const redPacketRewardUnitLabel = computed(() =>
    redPacketRewardIsRatio.value ? '比例' : '金额'
  );

  /** 无门槛（无限制）时只有一档奖励 */
  const visibleRedPacketRewardRows = computed(() =>
    redPacketRewardIsUnlimited.value
      ? form.redPacketRewardRows.slice(0, 1)
      : form.redPacketRewardRows
  );

  const redPacketRewardError = computed(() => {
    const needCondition = !redPacketRewardIsUnlimited.value;
    const random = redPacketRewardIsRandom.value;
    const unit = redPacketRewardUnitLabel.value;
    let message = '';
    visibleRedPacketRewardRows.value.some((row, index) => {
      const no = `(${index + 1})`;
      if (needCondition && row.condition == null) {
        message = `奖励依据门槛 ${no}`;
      } else if (random && row.min == null) {
        message = `最小${unit} ${no}`;
      } else if (random && row.max == null) {
        message = `最大${unit} ${no}`;
      } else if (!random && row.reward == null) {
        message = redPacketRewardIsUnlimited.value
          ? '固定金额'
          : `${form.redPacketRewardType} ${no}`;
      }
      return Boolean(message);
    });
    return message;
  });

  const redPacketValidUserNeedsThreshold = computed(
    () =>
      form.redPacketValidUser === '符合一个即可' ||
      form.redPacketValidUser === '全部满足'
  );

  const redPacketValidUserError = computed(() => {
    if (!redPacketValidUserNeedsThreshold.value) return '';
    if (
      !(form.redPacketValidDeposit > 0) ||
      !(form.redPacketValidBet > 0)
    ) {
      return '充值金额门槛和有效洗码门槛必须都大于0';
    }
    return '';
  });

  const addLuckyWinRow = () => {
    luckyWinRowUid += 1;
    form.luckyWinRows.push({ uid: luckyWinRowUid, number: '' });
  };

  const removeLuckyWinRow = (index: number) => {
    form.luckyWinRows.splice(index, 1);
  };

  const addOfferAmount = () => {
    if (
      form.offerAmountMode !== 'fixed' &&
      form.offerAmountMode !== 'random'
    ) {
      return;
    }
    offerAmountUid += 1;
    form.offerAmounts.push({ uid: offerAmountUid });
  };

  const removeOfferAmount = (index: number) => {
    form.offerAmounts.splice(index, 1);
  };

  const MAX_AGENT_IDS = 200;

  /** 中英文逗号都可作为分隔符 */
  const parseAgentIds = (value: string) =>
    value
      .split(/[,，]/)
      .map((id) => id.trim())
      .filter(Boolean);

  const agentIdsError = computed(() => {
    if (!form.specifiedAgent) return '';
    const ids = parseAgentIds(form.agentIds);
    if (!ids.length) return '请填写ID';
    if (ids.length > MAX_AGENT_IDS) return `最多支持${MAX_AGENT_IDS}个ID`;
    return '';
  });

  const getCurrentTypePreset = () => presetOf(form.type);

  /**
   * 多个类型共用的表单项，是否显示由类型预设决定；
   * 只属于某一个类型的区块直接按 form.type 判断
   */
  const showLoopMode = computed(() => getCurrentTypePreset().showLoopMode);

  const showBonusMode = computed(() => getCurrentTypePreset().showBonusMode);

  const showClaimMode = computed(() => getCurrentTypePreset().showClaimMode);

  /** 累计领取只能选比例金额 */
  const isBonusFixedDisabled = computed(
    () =>
      showClaimMode.value && form.claimMode === BET_CLAIM_MODE.CUMULATIVE
  );

  const loopOptions = computed(
    () => getCurrentTypePreset().loopOptions ?? defaultLoopOptions
  );

  const loopLabel = computed(
    () => getCurrentTypePreset().loopLabel ?? '循环模式'
  );

  const activityTimeHidden = computed(() => form.type === ACTIVITY_TYPE.VIP);

  const memberLevelsHidden = computed(() => form.type === ACTIVITY_TYPE.VIP);

  const clientListHidden = computed(() => form.type === ACTIVITY_TYPE.VIP);

  const showProfitStat = computed(() =>
    Boolean(getCurrentTypePreset().showProfitStat)
  );

  const memberLevelsLocked = computed(() =>
    Boolean(getCurrentTypePreset().lockMemberLevels)
  );

  const restrictionsHidden = computed(() =>
    Boolean(getCurrentTypePreset().hideRestrictions)
  );

  const isFieldHidden = (field: HideableField) =>
    Boolean(getCurrentTypePreset().hiddenFields?.includes(field));

  /** 领取过期天数一般仅手动领取时需要 */
  const showClaimExpireDays = computed(
    () =>
      !isFieldHidden('distribution') &&
      (Boolean(getCurrentTypePreset().alwaysShowExpireDays) ||
        form.distributionMode !== PAYOUT_MODE.AUTO)
  );

  const memberLevelsEarly = computed(() =>
    Boolean(getCurrentTypePreset().memberLevelsEarly)
  );

  /** 会员层级提前时，排在它前面的几项需要更靠前的 order */
  const leadingItemClass = computed(() =>
    memberLevelsEarly.value ? 'settings-item-leading' : ''
  );

  const validConditionLabel = computed(() =>
    form.type === ACTIVITY_TYPE.ASSIST ? '认定助力条件' : '认定有效条件'
  );

  const currentRechargeLoopRule = computed(() =>
    form.type === ACTIVITY_TYPE.DEPOSIT && form.rechargeType != null
      ? rechargeLoopRules[form.rechargeType]
      : undefined
  );

  /** 当前能选的循环模式，其余选项置灰 */
  const selectableLoopModes = computed(() => {
    // 推荐会员只展示当前值，不允许修改
    if (form.type === ACTIVITY_TYPE.REFER) return [form.loopMode];
    // 签到：连续签到只能循环；累计签到默认不循环，循环也可选
    if (form.type === ACTIVITY_TYPE.SIGN_IN) {
      if (form.signInMode === '连续签到') return [LOOP_MODE.INFINITE];
      return [LOOP_MODE.NONE, LOOP_MODE.INFINITE];
    }
    // 有充值类型时按充值类型的规则
    if (currentRechargeLoopRule.value) {
      return currentRechargeLoopRule.value.allowed;
    }
    // 其余按活动类型预设
    const { disabledLoopValues } = getCurrentTypePreset();
    return loopOptions.value.filter(
      (mode) => !disabledLoopValues.includes(mode)
    );
  });

  const isLoopModeDisabled = (mode: number) =>
    !selectableLoopModes.value.includes(mode);

  const loopModeLabel = (mode: number) => {
    // 个别类型对循环模式有自己的叫法，优先于接口文案
    const specialLabel = getCurrentTypePreset().loopOptionLabels?.[mode];
    return specialLabel ? t(specialLabel) : getEnumLabel('cycle_mode', mode);
  };

  /**
   * 充值活动何时禁用「次日」：
   * - 累计：始终禁用
   * - 单笔：每日循环 / 无限循环禁用；不循环、每周、每月可选
   */
  const isDepositNextDayDisabled = () => {
    if (form.type !== ACTIVITY_TYPE.DEPOSIT) return false;
    if (form.rechargeType === RECHARGE_TYPE.CUMULATIVE) return true;
    if (form.rechargeType === RECHARGE_TYPE.SINGLE) {
      return (
        form.loopMode === LOOP_MODE.DAILY ||
        form.loopMode === LOOP_MODE.INFINITE
      );
    }
    return false;
  };

  const resetNextDayIfDisabled = () => {
    if (
      form.distributionTime === PAYOUT_TIMING.NEXT_DAY &&
      isDepositNextDayDisabled()
    ) {
      form.distributionTime = PAYOUT_TIMING.REALTIME;
    }
  };

  const isFieldOptionDisabled = (field: string, value: number) => {
    const preset = getCurrentTypePreset();
    if (field === 'activityRestrictions') {
      return preset.disabledRestrictionValues.includes(value);
    }

    if (field === 'claimRestrictions') {
      return unavailableClaimLimits.includes(value);
    }

    if (field === 'giftMode') {
      return preset.disabledGiftValues.includes(value);
    }

    if (field === 'participationMode') {
      return preset.disabledParticipationValues.includes(value);
    }

    if (field === 'distributionMode') {
      return preset.disabledDistributionValues.includes(value);
    }

    if (field === 'distributionTime') {
      if (value === PAYOUT_TIMING.NEXT_DAY && isDepositNextDayDisabled()) {
        return true;
      }
      return preset.disabledDistributionTimeValues.includes(value);
    }

    return false;
  };

  const resetForm = (type: number = ACTIVITY_TYPE.DEPOSIT) => {
    Object.assign(form, createDefaultForm(type));
    bonusValidated.value = false;
    referralValidated.value = false;
    luckyValidated.value = false;
    redPacketValidated.value = false;
    contentImageValidated.value = false;
    ensureVipConfigs();
  };

  /** 打开后表单是否被编辑过，用于关闭前确认 */
  const formDirty = ref(false);

  watch(
    form,
    () => {
      formDirty.value = true;
    },
    { deep: true }
  );

  /** 富文本编辑器挂载时会规范化一次内容，等它完成后再开始记录改动 */
  const markPristine = () => {
    setTimeout(() => {
      formDirty.value = false;
    }, 0);
  };

  /** 编辑时打开弹窗那一刻的详情，保存时以它为底拼请求体 */
  const editingDetail = ref<ActivityDetail | null>(null);
  /** 打开后加载选项数据和详情期间为 true，表单此时被遮住 */
  const preparing = ref(false);
  let openTicket = 0;

  const applyDetail = async (detail: ActivityDetail) => {
    const result = detailToForm(
      detail,
      (type) => typePresets[type]?.loopOptions ?? defaultLoopOptions
    );
    if (!result.ok) {
      Message.warning(result.message);
      return false;
    }
    const { type, values } = result;
    const {
      locales,
      bonusTiers,
      wheelBetTiers,
      wheelDepositTiers,
      ...rest
    } = values;
    applyingDetail = true;
    resetForm(type);
    Object.assign(form, rest);
    if (locales?.length) {
      form.locales = locales.map((item) => {
        localeUid += 1;
        return { ...item, uid: localeUid, imageUploading: false };
      });
    }
    if (bonusTiers) {
      form.bonusTiers = bonusTiers.map((tier) => ({
        ...createBonusTier(),
        ...tier,
      }));
    }
    if (wheelBetTiers) {
      form.wheelBetTiers = wheelBetTiers.map((tier) => ({
        ...createWheelTriggerTier(),
        ...tier,
      }));
    }
    if (wheelDepositTiers) {
      form.wheelDepositTiers = wheelDepositTiers.map((tier) => ({
        ...createWheelTriggerTier(),
        ...tier,
      }));
    }
    if (
      form.drawConditions.includes(WHEEL_BY_BET) &&
      !form.wheelBetTiers.length
    ) {
      form.wheelBetTiers.push(createWheelTriggerTier());
    }
    if (
      form.drawConditions.includes(WHEEL_BY_DEPOSIT) &&
      !form.wheelDepositTiers.length
    ) {
      form.wheelDepositTiers.push(createWheelTriggerTier());
    }
    // 等类型、奖金模式的 watcher 在暂停状态下跑完再恢复联动
    await nextTick();
    applyingDetail = false;
    return true;
  };

  /**
   * 默认语言、VIP 配置、「全部」层级和客户端都依赖选项数据，
   * 必须等它们就绪后再初始化表单，否则后到的数据改动表单会被记为未保存
   */
  const prepareForm = async (id: number | null) => {
    openTicket += 1;
    const current = openTicket;
    preparing.value = true;
    try {
      const [, detail] = await Promise.all([
        Promise.all([
          loadLangOptions(),
          loadVipLevels(),
          loadLevels(),
          loadEnumParams(),
        ]),
        id != null ? getActivityDetail(id).then(({ data }) => data) : null,
      ]);
      if (current !== openTicket) return;
      if (detail) {
        if (!(await applyDetail(detail))) {
          visible.value = false;
          return;
        }
        editingDetail.value = detail;
      } else {
        resetForm();
      }
      markPristine();
    } catch {
      // 详情加载失败，提示由全局拦截器弹出
      if (current === openTicket) visible.value = false;
    } finally {
      if (current === openTicket) preparing.value = false;
    }
  };

  watch(
    () => visible.value,
    (isVisible) => {
      if (!isVisible) {
        openTicket += 1;
        preparing.value = false;
        editingDetail.value = null;
        // 下次打开重新拉取，带上期间新建的活动
        activityOptionsLoaded = false;
        resetForm();
        return;
      }
      prepareForm(props.activityId ?? null);
    }
  );

  watch(
    () => visible.value && form.type === ACTIVITY_TYPE.IMAGE_AD,
    (needed) => {
      if (needed) loadActivityOptions();
    }
  );

  const confirmDiscard = (onDiscard: () => void) => {
    Modal.confirm({
      title: '放弃未保存的内容？',
      content: '关闭后，本次填写的内容将不会保存。',
      okText: '放弃',
      cancelText: '继续编辑',
      onOk: onDiscard,
    });
  };

  const hasUnsavedChanges = () => formDirty.value && !preparing.value;

  const requestClose = () => {
    if (!hasUnsavedChanges()) {
      visible.value = false;
      return;
    }
    confirmDiscard(() => {
      visible.value = false;
    });
  };

  /** 右上角关闭和 Esc 也走同样的确认 */
  const handleBeforeCancel = () => {
    if (!hasUnsavedChanges()) return true;
    requestClose();
    return false;
  };

  const fillLocalePreset = (item: ActivityLocale) => {
    if (!item.lang) return;
    Object.assign(
      item,
      getActivityLocalePreset(form.type, item.lang, form.rechargeType)
    );
  };

  watch(
    () => form.type,
    (type) => {
      if (applyingDetail) return;
      const { locales } = form;
      resetForm(type || ACTIVITY_TYPE.DEPOSIT);
      // 切换活动类型时保留已添加的语言页签，只按新类型重填文案
      form.locales = locales;
      locales.forEach(fillLocalePreset);
    },
    { immediate: true }
  );

  // 切换充值类型时，循环模式跳到该类型的默认值
  watch(currentRechargeLoopRule, (rule) => {
    // 编辑回填时保留已保存的循环模式
    if (applyingDetail || !rule) return;
    form.loopMode = rule.default;
  });

  // 签到方式联动循环模式：连续→只能循环；累计→默认不循环
  watch(
    () => form.signInMode,
    (mode) => {
      if (applyingDetail || form.type !== ACTIVITY_TYPE.SIGN_IN) return;
      form.loopMode =
        mode === '连续签到' ? LOOP_MODE.INFINITE : LOOP_MODE.NONE;
    }
  );

  // 切换充值类型：按规则校正「次日」；并重填活动规则首条
  watch(
    () => form.rechargeType,
    () => {
      if (applyingDetail) return;
      resetNextDayIfDisabled();
      if (form.type === ACTIVITY_TYPE.DEPOSIT) {
        form.locales.forEach((item) => {
          if (!item.lang) return;
          item.content = getActivityLocalePreset(
            form.type,
            item.lang,
            form.rechargeType
          ).content;
        });
      }
    }
  );

  // 单笔下切换循环模式时，同步校正「次日」是否可用
  watch(
    () => form.loopMode,
    () => {
      if (applyingDetail) return;
      resetNextDayIfDisabled();
    }
  );

  const quillToolbar = [
    ['bold', 'italic', 'underline', 'strike'],
    [{ list: 'ordered' }, { list: 'bullet' }],
    ['blockquote', 'code-block'],
    [{ header: [1, 2, 3, false] }],
    [{ color: [] }, { background: [] }],
    ['clean'],
  ];

  const isLangUsed = (lang: string, selfIndex: number) =>
    form.locales.some(
      (item, index) => index !== selfIndex && item.lang === lang
    );

  const availableLangOptions = computed(() =>
    langOptions.value.filter(
      (opt) => !form.locales.some((item) => item.lang === opt.value)
    )
  );

  const activeLocaleUid = ref<number>(form.locales[0].uid);

  watch(
    () => form.locales.map((item) => item.uid),
    (uids) => {
      if (!uids.includes(activeLocaleUid.value)) {
        [activeLocaleUid.value] = uids;
      }
    }
  );

  const getLangLabel = (lang: string) => {
    if (!lang) return '未选择语言';
    return langOptions.value.find((opt) => opt.value === lang)?.label ?? lang;
  };

  const isLocaleIncomplete = (item: ActivityLocale) =>
    !item.lang ||
    !item.name.trim() ||
    !item.content.replace(/<[^>]*>/g, '').trim();

  const addLocale = () => {
    const nextLang = availableLangOptions.value[0];
    let locale: ActivityLocale;
    if (nextLang) {
      locale = createLocale(form.type, nextLang.value);
    } else {
      localeUid += 1;
      locale = {
        uid: localeUid,
        lang: '',
        name: '',
        description: '',
        content: '',
        image: '',
        imageUploading: false,
      };
    }
    form.locales.push(locale);
    activeLocaleUid.value = locale.uid;
  };

  const saving = ref(false);

  const getFormVisibility = (): CreateFormVisibility => ({
    activityTime: !activityTimeHidden.value,
    specifiedGame: !isFieldHidden('specifiedGame'),
    rechargeType: form.type === ACTIVITY_TYPE.DEPOSIT,
    loopMode: showLoopMode.value,
    participation: !isFieldHidden('participation'),
    distribution: !isFieldHidden('distribution'),
    claimExpireDays: showClaimExpireDays.value,
    claimMode: showClaimMode.value,
    bonusMode: showBonusMode.value,
    memberLevels: !memberLevelsHidden.value,
    withdraw: !isFieldHidden('withdraw'),
    clientList: !clientListHidden.value,
    restrictions: !restrictionsHidden.value,
  });

  const submitDraft = async (show: CreateFormVisibility) => {
    const payload = buildCreatePayload(form, show);
    const detail = editingDetail.value;
    saving.value = true;
    let savedId: number;
    try {
      if (detail) {
        await updateActivity(buildUpdatePayload(detail, payload));
        savedId = detail.id;
      } else {
        const { data } = await createActivity(payload);
        savedId = data.id;
      }
    } catch {
      return;
    } finally {
      saving.value = false;
    }
    Message.success(detail ? '活动已保存' : '活动已保存为草稿');
    visible.value = false;
    emit('saved', savedId);
  };

  const handleSaveDraft = () => {
    if (saving.value || preparing.value) return;
    // 详情加载失败时弹窗已关闭，这里兜底防止把编辑误存成新活动
    if (isEdit.value && !editingDetail.value) return;
    const blankLocale = form.locales.find((item) => !item.lang);
    if (blankLocale) {
      expandedSections.rules = true;
      activeLocaleUid.value = blankLocale.uid;
      Message.warning('存在未选择语言的页签，请选择语言或删除该页签');
      return;
    }
    if (form.locales.some((item) => item.imageUploading)) {
      Message.warning('活动图片上传中，请稍候');
      return;
    }
    const incompleteLocale = form.locales.find(isLocaleIncomplete);
    if (incompleteLocale) {
      expandedSections.rules = true;
      activeLocaleUid.value = incompleteLocale.uid;
      Message.warning(
        `请填写「${getLangLabel(incompleteLocale.lang)}」的活动名称和规则`
      );
      return;
    }
    if (form.type === ACTIVITY_TYPE.NEWCOMER && agentIdsError.value) {
      expandedSections.settings = true;
      Message.warning(`指定代理：${agentIdsError.value}`);
      return;
    }
    if (showBonusMode.value && form.bonusMode && bonusTierError.value) {
      bonusValidated.value = true;
      expandedSections.rewards = true;
      Message.warning(bonusTierError.value);
      return;
    }
    if (form.type === ACTIVITY_TYPE.REFER && referralRewardError.value) {
      referralValidated.value = true;
      expandedSections.settings = true;
      Message.warning(referralRewardError.value);
      return;
    }
    if (form.type === ACTIVITY_TYPE.LUCKY_BET && luckyWinError.value) {
      luckyValidated.value = true;
      expandedSections.rewards = true;
      Message.warning(luckyWinError.value);
      return;
    }
    if (form.type === ACTIVITY_TYPE.RED_PACKET && redPacketValidUserError.value) {
      redPacketValidated.value = true;
      expandedSections.rewards = true;
      Message.warning(redPacketValidUserError.value);
      return;
    }
    if (form.type === ACTIVITY_TYPE.RED_PACKET && redPacketRewardError.value) {
      redPacketValidated.value = true;
      expandedSections.rewards = true;
      Message.warning(`请填写${redPacketRewardError.value}`);
      return;
    }
    if (form.type === ACTIVITY_TYPE.IMAGE_AD && contentImageError.value) {
      contentImageValidated.value = true;
      expandedSections.rules = true;
      Message.warning(contentImageError.value);
      return;
    }
    const show = getFormVisibility();
    const error = validateCreateForm(form, show);
    if (error) {
      expandedSections[error.section] = true;
      Message.warning(error.message);
      return;
    }
    submitDraft(show);
  };

  const removeLocale = (uid: string | number) => {
    if (form.locales.length <= 1) return;
    const index = form.locales.findIndex((item) => item.uid === uid);
    if (index < 0) return;
    form.locales.splice(index, 1);
    if (activeLocaleUid.value === uid) {
      activeLocaleUid.value =
        form.locales[Math.min(index, form.locales.length - 1)].uid;
    }
  };

  const isAllMemberLevelsSelected = computed(() =>
    memberLevelOptions.value.every((level) => form.memberLevels.includes(level))
  );

  const toggleMemberLevels = () => {
    if (memberLevelsLocked.value) return;
    form.memberLevels = isAllMemberLevelsSelected.value
      ? []
      : ['全部', ...memberLevelOptions.value];
  };

  // 层级异步加载，加载完成前选中的「全部」需要补齐到具体层级
  watch(memberLevelOptions, (options) => {
    if (form.memberLevels.includes('全部')) {
      form.memberLevels = ['全部', ...options];
    }
  });

  const isAllClientListSelected = computed(() =>
    clientCodes.value.every((client) => form.clientList.includes(client))
  );

  const toggleClientList = () => {
    form.clientList = isAllClientListSelected.value
      ? []
      : [ALL_OPTION, ...clientCodes.value];
  };

  watch(
    () => form.memberLevels.slice(),
    (levels) => {
      const filtered = levels.filter((level) => level !== '全部');
      const allSelected = memberLevelOptions.value.every((level) =>
        filtered.includes(level)
      );
      const hasAll = levels.includes('全部');

      if (hasAll && !allSelected) {
        form.memberLevels = filtered.length
          ? filtered
          : ['全部', ...memberLevelOptions.value];
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
      const filtered = clients.filter((client) => client !== ALL_OPTION);
      const allSelected = clientCodes.value.every((client) =>
        filtered.includes(client)
      );
      const hasAll = clients.includes(ALL_OPTION);

      if (hasAll && !allSelected) {
        form.clientList = filtered.length
          ? filtered
          : [ALL_OPTION, ...clientCodes.value];
        return;
      }

      if (allSelected && !hasAll) {
        form.clientList = [ALL_OPTION, ...filtered];
      }
    },
    { immediate: true }
  );
</script>

<style scoped>
  .activity-create-modal :deep(.arco-modal) {
    border-radius: 16px;
    overflow: hidden;
    background: var(--color-bg-2);
    box-shadow: 0 18px 46px rgba(29, 33, 41, 0.12);
  }

  .activity-create-modal :deep(.arco-modal-header) {
    padding: 0;
    border-bottom: 1px solid var(--color-border-1);
    background: var(--color-bg-2);
  }

  .activity-create-modal :deep(.arco-modal-content) {
    background: var(--color-bg-1);
  }

  .activity-create-modal :deep(.arco-modal-body) {
    position: relative;
    padding: 0;
    background: var(--color-bg-1);
  }

  /* 编辑时详情返回前盖住表单，避免先闪出默认值 */
  .detail-loading {
    position: absolute;
    inset: 0;
    z-index: 20;
    display: flex;
    justify-content: center;
    padding-top: 160px;
    background: var(--color-bg-1);
  }

  /* 苹果 macOS 窗口弹出与关闭微交互动效 */
  .activity-create-modal :deep(.zoom-modal-enter-active) {
    transition: transform 0.32s cubic-bezier(0.16, 1, 0.3, 1),
      opacity 0.28s cubic-bezier(0.16, 1, 0.3, 1), filter 0.28s ease-out;
    transform-origin: center center;
  }

  .activity-create-modal :deep(.zoom-modal-leave-active) {
    transition: transform 0.24s cubic-bezier(0.32, 0.72, 0, 1),
      opacity 0.2s cubic-bezier(0.32, 0.72, 0, 1), filter 0.2s ease-in;
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
    color: var(--color-text-1);
    letter-spacing: 0.2px;
  }

  .activity-dialog-toolbar {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    padding: 10px 18px 12px;
    background: var(--color-bg-2);
    border-bottom: 1px solid var(--color-border-1);
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
    background: var(--color-fill-2);
    color: var(--color-text-2);
    border: 1px solid var(--color-border-2);
  }

  .toolbar-btn.primary {
    background: rgb(var(--primary-6));
    border: 1px solid rgb(var(--primary-6));
    color: #fff;
  }

  .toolbar-close {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    border-radius: 6px;
    color: var(--color-text-2);
  }

  .custom-form {
    padding: 18px 18px 0;
    background: var(--color-bg-1);
  }

  .activity-form-grid {
    align-items: stretch;
    margin: 0;
  }

  .section-card {
    margin-bottom: 16px;
    padding: 18px 18px 0;
    border: 1px solid var(--color-border-2);
    border-radius: 14px;
    background: var(--color-bg-2);
    box-shadow: 0 2px 10px rgba(22, 93, 255, 0.02);
  }

  .section-title {
    margin: 0 0 10px;
    padding-left: 10px;
    font-size: 14px;
    font-weight: 700;
    color: var(--color-text-1);
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
    border: 1px solid var(--color-border-2);
    border-radius: 12px;
    background: var(--color-bg-2);
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
    border-color: rgb(var(--primary-3));
    background: var(--color-fill-1);
    box-shadow: 0 4px 12px rgba(22, 93, 255, 0.08);
  }

  .section-title--collapsible.active {
    margin-bottom: 0;
    color: rgb(var(--primary-6));
    border-color: rgb(var(--primary-3));
    border-bottom-color: rgb(var(--primary-2));
    border-radius: 12px 12px 0 0;
    background: linear-gradient(
      90deg,
      rgb(var(--primary-1)) 0%,
      var(--color-bg-2) 72%
    );
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
    background: var(--color-fill-2);
    color: var(--color-text-3);
    font-size: 14px;
    transition: transform 0.2s ease;
  }

  .section-title--collapsible.active .section-title-icon {
    color: rgb(var(--primary-6));
    background: rgb(var(--primary-1));
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
    background: rgb(var(--primary-6));
  }

  .section-body {
    display: block;
  }

  .form-panel-left,
  .form-panel-right {
    min-height: 560px;
    background: transparent;
  }

  /* 左右两栏各自滚动，避免一侧过长时另一侧大片留白；240px 为弹窗标题、底部按钮与上下边距 */
  .form-panel-left,
  .form-panel-right {
    display: flex;
    flex-direction: column;
    max-height: calc(100vh - 240px);
    overflow-y: auto;
    overscroll-behavior: contain;
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
    border-right: 1px solid var(--color-border-1);
  }

  .form-panel-right {
    padding-left: 18px;
  }

  /* 窄于 1200px（arco xl 断点，与 a-col :xl 对应）时两栏上下堆叠，改为整个表单区统一滚动 */
  @media (max-width: 1199px) {
    .activity-form-grid {
      max-height: calc(100vh - 240px);
      overflow-y: auto;
      overscroll-behavior: contain;
    }

    .form-panel-left,
    .form-panel-right {
      min-height: 0;
      max-height: none;
      padding: 0;
      overflow: visible;
      border-right: none;
    }

    .form-panel-left {
      margin-bottom: 16px;
    }
  }

  .custom-form :deep(.arco-form-item) {
    margin-bottom: 18px;
  }

  .custom-form :deep(.arco-form-item-label) {
    color: var(--color-text-1);
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
    color: var(--color-text-2);
    --checkbox-border-color: var(--color-border-3);
    display: inline-flex;
    align-items: center;
  }

  .custom-form :deep(.arco-radio-checked .arco-radio-inner),
  .custom-form :deep(.arco-checkbox-checked .arco-checkbox-inner) {
    border-color: rgb(var(--primary-6));
    background-color: rgb(var(--primary-6));
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
    background: var(--color-bg-2);
    border-color: var(--color-border-2);
    border-radius: 6px;
    box-shadow: none;
    transition: border-color 0.2s ease, box-shadow 0.2s ease;
  }

  .custom-form :deep(.arco-select-view):focus-within,
  .custom-form :deep(.arco-input-wrapper):focus-within,
  .custom-form :deep(.arco-picker):focus-within {
    border-color: rgb(var(--primary-4));
    box-shadow: 0 0 0 2px rgba(22, 93, 255, 0.08);
  }

  .custom-form :deep(.arco-select-view),
  .custom-form :deep(.arco-input-wrapper) {
    height: 32px;
  }

  .custom-form :deep(.arco-select-view-value),
  .custom-form :deep(.arco-input) {
    color: var(--color-text-1);
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
    color: var(--color-text-2);
    background-color: var(--color-fill-2);
    border: 1px solid transparent;
    border-radius: 4px;
    cursor: pointer;
    user-select: none;
    transition: all 0.2s ease;
  }

  .type-item:hover {
    background-color: var(--color-fill-3);
    border-color: rgb(var(--primary-2));
  }

  .type-item.active {
    color: #fff;
    background-color: rgb(var(--primary-6));
    border-color: rgb(var(--primary-6));
    box-shadow: 0 4px 10px rgba(22, 93, 255, 0.18);
  }

  /* 2. 活动名称与规则（按语言成套） */
  .locale-block {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding-bottom: 18px;
  }

  .locale-block-title {
    position: relative;
    padding-left: 10px;
    font-size: 14px;
    font-weight: 700;
    line-height: 16px;
    color: var(--color-text-1);
  }

  .locale-block-title::before {
    content: '';
    position: absolute;
    left: 0;
    top: 1px;
    width: 3px;
    height: 14px;
    border-radius: 2px;
    background: rgb(var(--primary-6));
  }

  .locale-required {
    margin-left: 4px;
    color: rgb(var(--danger-6));
  }

  .locale-tabs :deep(.arco-tabs-content) {
    padding: 16px;
    border: 1px solid var(--color-border-2);
    border-top: 0;
    border-radius: 0 0 10px 10px;
    background: var(--color-fill-1);
  }

  .locale-tab-title {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }

  .locale-tab-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: rgb(var(--danger-6));
  }

  .locale-lang-select {
    width: 220px;
  }

  .locale-card-body {
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  .locale-field {
    display: grid;
    grid-template-columns: 76px minmax(0, 1fr);
    align-items: center;
    column-gap: 12px;
  }

  .locale-field--top {
    align-items: start;
  }

  .locale-field--top .locale-field-label {
    padding-top: 9px;
  }

  .locale-field-label {
    text-align: right;
    font-size: 13px;
    color: var(--color-text-1);
    white-space: nowrap;
  }

  .locale-field-label--optional::before {
    visibility: hidden;
  }

  .locale-field-label::before {
    content: '*';
    margin-right: 4px;
    color: rgb(var(--danger-6));
  }

  .claim-mode-tip {
    margin-left: 4px;
    color: var(--color-text-2);
    cursor: help;
  }

  .lucky-win-tips {
    display: flex;
    flex-direction: column;
    gap: 8px;
    font-size: 13px;
    line-height: 1.6;
    color: var(--color-text-2);
  }

  /* 活动设置区块用 flex 排列，个别类型可通过 order 调整表单项位置 */
  .settings-body {
    display: flex;
    flex-direction: column;
  }

  .settings-item-leading {
    order: -2;
  }

  .settings-item-early {
    order: -1;
  }

  /* 签到 */
  .sign-in-inline {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .sign-in-days-input {
    width: 240px;
  }

  .sign-in-table {
    border-collapse: separate;
    border-spacing: 10px 8px;
    margin-left: -10px;
    font-size: 13px;
    color: var(--color-text-1);
  }

  .sign-in-table th {
    font-weight: 600;
    text-align: left;
    white-space: nowrap;
  }

  .sign-in-table td :deep(.arco-input-wrapper) {
    width: 100px;
  }

  .sign-in-table td:nth-child(3) :deep(.arco-input-wrapper) {
    width: 160px;
  }

  .sign-in-type {
    width: 100px;
  }

  .sign-in-vip-group {
    padding: 12px 0 4px;
  }

  .sign-in-vip-group + .sign-in-vip-group {
    margin-top: 16px;
  }

  .sign-in-vip-header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 8px;
    color: var(--color-text-1);
  }

  .sign-in-vip-header > span,
  .sign-in-vip-header :deep(.arco-link) {
    flex-shrink: 0;
    white-space: nowrap;
  }

  .sign-in-vip-select {
    flex: 1;
    min-width: 0;
    max-width: 350px;
  }

  .sign-in-range {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .sign-in-range :deep(.arco-input-wrapper) {
    width: 74px !important;
  }

  /* 图文(广告)：内容图片 */
  .content-image-group {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .content-image-card {
    display: flex;
    flex-direction: column;
    gap: 12px;
    width: 360px;
    padding: 14px 16px;
    border: 1px solid var(--color-border-2);
    border-radius: 10px;
    background: var(--color-bg-2);
  }

  .content-image-header,
  .content-image-lang-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }

  .content-image-header {
    font-weight: 600;
    color: var(--color-text-1);
  }

  .content-image-item {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .content-image-item + .content-image-item {
    padding-top: 12px;
    border-top: 1px dashed var(--color-border-2);
  }

  .content-image-lang {
    width: 220px;
  }

  .content-image-link-label {
    color: var(--color-text-1);
  }

  .content-image-link {
    width: 100%;
  }

  /* 红包雨 */
  .red-wave-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .red-wave-row {
    display: flex;
    align-items: center;
    gap: 20px;
  }

  .red-wave-head {
    font-weight: 600;
    color: var(--color-text-1);
  }

  .red-wave-hour {
    width: 270px;
  }

  .red-wave-duration {
    width: 270px;
  }

  .red-limit-list {
    display: grid;
    grid-template-columns: max-content 380px;
    align-items: center;
    gap: 12px 24px;
  }

  .red-limit-label {
    color: var(--color-text-1);
  }

  .red-history-row {
    display: flex;
    align-items: center;
    gap: 24px;
    color: var(--color-text-1);
  }

  .red-history-select {
    width: 380px;
  }

  .red-reward-row {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .red-reward-index-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .red-reward-input {
    width: 325px;
  }

  .red-reward-input--wide {
    width: 395px;
  }

  .red-wave-actions {
    display: flex;
    gap: 12px;
    min-width: 108px;
  }

  /* VIP奖金：VIP配置 */
  .vip-config-table {
    width: 100%;
    border-collapse: collapse;
    text-align: center;
    font-size: 13px;
    color: var(--color-text-1);
  }

  .vip-config-table th,
  .vip-config-table td {
    padding: 8px 10px;
    border: 1px solid var(--color-border-2);
  }

  .vip-config-table th {
    font-weight: 500;
    background: var(--color-fill-1);
  }

  .vip-config-input {
    width: 100%;
  }

  /* 幸运注单：是否盈利 */
  .lucky-profit-row {
    display: flex;
    align-items: center;
    gap: 12px;
    color: var(--color-text-1);
  }

  /* 全民代理：3级抽成 */
  .agent-commission-item {
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin-bottom: 8px;
  }

  .agent-commission-formula {
    font-size: 13px;
    color: rgb(var(--primary-6));
  }

  /* 新人彩金：优惠金额 */
  .offer-amount-input {
    width: 450px;
  }

  .offer-range-box {
    display: flex;
    align-items: center;
    width: 450px;
    height: 32px;
    padding: 0 12px;
    border: 1px solid var(--color-border-2);
    border-radius: var(--border-radius-small);
    background: var(--color-bg-2);
    transition: border-color 0.2s ease;
  }

  .offer-range-box:focus-within {
    border-color: rgb(var(--primary-6));
  }

  .offer-range-affix {
    flex: none;
    color: var(--color-text-3);
  }

  .offer-range-input {
    flex: 1;
    min-width: 0;
  }

  .offer-range-input,
  .offer-range-input:hover,
  .offer-range-input.arco-input-focus {
    border-color: transparent;
    background: transparent;
    box-shadow: none;
  }

  .offer-range-input :deep(.arco-input) {
    text-align: center;
  }

  /* 新人彩金：指定代理 */
  .agent-id-group {
    display: flex;
    flex-direction: column;
    gap: 10px;
    width: 100%;
  }

  .agent-id-error {
    margin-top: -6px;
    font-size: 12px;
    color: rgb(var(--danger-6));
  }

  /* 助力：本次实际奖金 */
  .assist-ref-panel {
    padding: 14px 16px 16px;
    border-radius: 10px;
    background: var(--color-fill-1);
  }

  .assist-ref-title {
    margin-bottom: 10px;
    font-size: 14px;
    font-weight: 600;
    color: var(--color-text-1);
  }

  .assist-ref-list {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 12px;
  }

  .assist-ref-image {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 140px;
    border-radius: 6px;
    background: var(--color-bg-2);
    overflow: hidden;
  }

  .assist-ref-image img {
    height: 100%;
    object-fit: contain;
  }

  .assist-ref-placeholder {
    font-size: 32px;
    color: var(--color-text-4);
  }

  .assist-ref-caption {
    margin-top: 8px;
    text-align: center;
    font-size: 13px;
    color: var(--color-text-1);
  }

  .assist-amount-input {
    width: 400px;
  }

  .assist-helper {
    font-size: 13px;
    color: var(--color-text-2);
  }

  /* 幸运轮盘奖金设置 */
  .wheel-prize-type {
    width: 280px;
  }

  .wheel-prize-input {
    width: 110px;
  }

  .wheel-prize-prob {
    width: 160px;
  }

  .draw-condition-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 16px 24px;
    width: 100%;
  }

  .draw-condition-col {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .draw-condition-input {
    width: 200px;
  }

  /* 被邀请会员有效判定 */
  .invitee-rule-grid {
    display: grid;
    grid-template-columns: repeat(2, max-content);
    gap: 12px 24px;
  }

  .invitee-rule-item {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .invitee-rule-input {
    width: 110px;
  }

  .invitee-rule-unit {
    color: var(--color-text-1);
    font-weight: 500;
  }

  .invitee-rule-item .claim-mode-tip {
    margin-left: 0;
  }

  /* 奖金模式阶梯 */
  .bonus-mode-group {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .bonus-tier-row {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .bonus-tier-index {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex: none;
    width: 28px;
    height: 28px;
    border: 1px solid rgb(var(--primary-2));
    border-radius: 4px;
    background: rgb(var(--primary-1));
    color: rgb(var(--primary-6));
    font-size: 13px;
  }

  .bonus-tier-input {
    width: 170px;
  }

  .bonus-tier-add-btn {
    align-self: flex-start;
    width: 52px;
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
  .upload-tip {
    font-size: 12px;
    color: var(--color-text-3);
    line-height: 1.6;
  }

  .cover-image-item {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  /* 按 668 × 250 的比例缩到一半预览 */
  .cover-image-box {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 334px;
    height: 125px;
    overflow: hidden;
    background: var(--color-fill-1);
    border: 1px dashed var(--color-border-3);
    border-radius: 8px;
    cursor: pointer;
    transition: border-color 0.2s;
  }

  .cover-image-box:hover {
    border-color: rgb(var(--primary-6));
  }

  .cover-image-box.is-error {
    border-color: rgb(var(--danger-6));
  }

  .cover-image-box img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .cover-image-empty {
    display: flex;
    flex-direction: column;
    gap: 6px;
    align-items: center;
    color: var(--color-text-3);
    font-size: 12px;
  }

  .cover-image-empty :deep(.arco-icon) {
    font-size: 20px;
  }

  .cover-image-mask {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    font-size: 13px;
    background: rgba(0, 0, 0, 0.45);
    opacity: 0;
    transition: opacity 0.2s;
  }

  .cover-image-box:hover .cover-image-mask {
    opacity: 1;
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

  .inline-helper {
    color: var(--color-text-3);
    white-space: nowrap;
  }

  .level-group-label {
    min-width: 56px;
    color: var(--color-text-3);
    white-space: nowrap;
  }

  .dialog-footer-actions {
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 16px;
    padding: 20px 0 22px;
    background: var(--color-bg-1);
    border-top: 1px solid var(--color-border-1);
  }

  .cancel-btn,
  .save-draft-btn {
    min-width: 104px;
    height: 36px;
    border-radius: 8px;
  }

  .quill-editor-wrap {
    min-width: 0;
  }

  .quill-editor {
    width: 100%;
  }

  :deep(.ql-toolbar) {
    border: 1px solid var(--color-border-2);
    border-bottom: none;
    border-radius: 8px 8px 0 0;
    background: var(--color-fill-2);
    min-height: 38px;
  }

  :deep(.ql-container) {
    min-height: 180px;
    max-height: 260px;
    overflow-y: auto;
    border: 1px solid var(--color-border-2);
    border-radius: 0 0 8px 8px;
    font-size: 14px;
    line-height: 1.8;
    background: var(--color-bg-2);
  }

  :deep(.ql-editor) {
    min-height: 180px;
    color: var(--color-text-1);
    font-size: 13px;
  }

  :deep(.ql-editor.ql-blank::before) {
    color: var(--color-text-4);
    font-style: normal;
  }

  :deep(.ql-editor blockquote) {
    margin: 8px 0;
    padding-left: 12px;
    border-left: 3px solid rgb(var(--primary-6));
    color: var(--color-text-2);
  }

  .currency-text {
    font-weight: 500;
    color: var(--color-text-1);
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
    border: 1px solid var(--color-border-2);
    border-radius: 10px;
    background: linear-gradient(
      180deg,
      var(--color-fill-1) 0%,
      var(--color-bg-2) 100%
    );
    box-shadow: inset 0 0 0 1px rgba(22, 93, 255, 0.02);
  }

  .specified-game-title {
    font-size: 14px;
    font-weight: 600;
    color: var(--color-text-1);
  }

  .specified-game-desc {
    margin-top: 4px;
    font-size: 12px;
    color: var(--color-text-3);
    line-height: 1.6;
  }

  .specified-game-picker {
    margin-top: 12px;
  }
</style>
