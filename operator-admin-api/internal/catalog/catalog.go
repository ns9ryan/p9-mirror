package catalog

import (
	"net/http"

	"oa.98ent.com/p9/core/rpc/coreclient"
)

const (
	menuTypeDir    int32 = 0
	menuTypeMenu   int32 = 1
	menuTypeButton int32 = 2

	serviceName = "operator-admin-api"
)

// CatalogReq 菜单、API目录、多语言数据种子数据
func catalogReq(code string) *coreclient.RegisterCatalogReq {
	return &coreclient.RegisterCatalogReq{
		Menus: []*coreclient.RegisterMenuReq{
			{Name: "Dashboard", Title: "menu.route.dashboard", Path: "/dashboard", MenuType: menuTypeMenu, Component: "dashboard/index", Sort: 1},

			// 会员管理
			{Name: "MemberManage", Title: "menu.memberManage", Path: "/member", MenuType: menuTypeDir, Sort: 10},
			{Name: "MemberManageList", Title: "menu.memberManage.list", Path: "/member/list", MenuType: menuTypeMenu, Component: "memberManage/index", ParentName: "MemberManage", Sort: 11},

			// 会员列表按钮权限
			{Name: "MemberSearchByCode", Title: "menu.route.memberSearchByCode", MenuType: menuTypeButton, Permission: "member:search:code", ParentName: "MemberManageList", Sort: 111},

			// 层级管理: 自动层级和固定层级在同一页面内切换
			{Name: "MemberLevel", Title: "menu.route.memberLevel", Path: "/member/level", MenuType: menuTypeMenu, Component: "memberManage/level/index", ParentName: "MemberManage", Sort: 12},

			// VIP 管理: 基础设置、等级配置和图标设置在同一页面内切换
			{Name: "MemberVip", Title: "menu.route.memberVip", Path: "/member/vip", MenuType: menuTypeMenu, Component: "memberManage/vip/index", ParentName: "MemberManage", Sort: 13},

			{Name: "MemberContactBindingCodeList", Title: "menu.route.memberContactBindingCodeList", Path: "/member/contact/binding-codes/list", MenuType: menuTypeMenu, Component: "memberManage/contactBindingCodes/index", ParentName: "MemberManage", Sort: 14},

			// 自动层级按钮权限
			{Name: "MemberAutoLevelCreate", Title: "menu.route.memberAutoLevelCreate", MenuType: menuTypeButton, Permission: "member:auto-level:create", ParentName: "MemberLevel", Sort: 121},

			// 固定层级按钮权限
			{Name: "MemberFixedLevelCreate", Title: "menu.route.memberFixedLevelCreate", MenuType: menuTypeButton, Permission: "member:fixed-level:create", ParentName: "MemberLevel", Sort: 122},
			{Name: "MemberFixedLevelUpdate", Title: "menu.route.memberFixedLevelUpdate", MenuType: menuTypeButton, Permission: "member:fixed-level:update", ParentName: "MemberLevel", Sort: 123},
			{Name: "MemberFixedLevelDelete", Title: "menu.route.memberFixedLevelDelete", MenuType: menuTypeButton, Permission: "member:fixed-level:delete", ParentName: "MemberLevel", Sort: 124},

			// 代理管理
			{Name: "Agent", Title: "menu.route.agent", Path: "/agent", MenuType: menuTypeDir, Sort: 20},
			{Name: "AgentList", Title: "menu.route.agentList", Path: "/agent/list", MenuType: menuTypeMenu, Component: "agent/list/index", ParentName: "Agent", Sort: 21},
			{Name: "AgentTree", Title: "menu.route.agentTree", HideMenu: 1, Path: "/agent/tree", MenuType: menuTypeMenu, Component: "agent/tree/index", ParentName: "Agent", Sort: 22},
			{Name: "AgentConfig", Title: "menu.route.agentConfig", HideMenu: 1, Path: "/agent/config", MenuType: menuTypeMenu, Component: "agent/config/index", ParentName: "Agent", Sort: 23},
			{Name: "AgentLevel", Title: "menu.route.agentLevel", HideMenu: 1, Path: "/agent/level", MenuType: menuTypeMenu, Component: "agent/level/index", ParentName: "Agent", Sort: 24},
			{Name: "AgentApply", Title: "menu.route.agentApply", HideMenu: 1, Path: "/agent/apply", MenuType: menuTypeMenu, Component: "agent/apply/index", ParentName: "Agent", Sort: 25},
			{Name: "AgentLoginLog", Title: "menu.route.agentLoginLog", Path: "/agent/login-log", MenuType: menuTypeMenu, Component: "agent/login-log/index", ParentName: "Agent", Sort: 26},
			{Name: "AgentLoginAssociation", Title: "menu.route.agentLoginAssociation", Path: "/agent/login-association", MenuType: menuTypeMenu, Component: "agent/login-association/index", ParentName: "Agent", Sort: 27},
			{Name: "AgentCreate", Title: "menu.route.agentCreate", MenuType: menuTypeButton, Permission: "agent:create", ParentName: "AgentList", Sort: 211},
			{Name: "AgentSearch", Title: "menu.route.agentSearch", MenuType: menuTypeButton, Permission: "agent:search", ParentName: "AgentList", Sort: 212},
			{Name: "AgentProfileGet", Title: "menu.route.agentProfileGet", MenuType: menuTypeButton, Permission: "agent:profile:get", ParentName: "AgentList", Sort: 213},
			{Name: "AgentPasswordReset", Title: "menu.route.agentPasswordReset", MenuType: menuTypeButton, Permission: "agent:password:reset", ParentName: "AgentList", Sort: 214},
			{Name: "AgentWithdrawPasswordReset", Title: "menu.route.agentWithdrawPasswordReset", MenuType: menuTypeButton, Permission: "agent:withdraw-password:reset", ParentName: "AgentList", Sort: 215},
			{Name: "AgentShareRateUpdate", Title: "menu.route.agentShareRateUpdate", MenuType: menuTypeButton, Permission: "agent:share-rate:update", ParentName: "AgentList", Sort: 216},
			{Name: "AgentPayoutUpdate", Title: "menu.route.agentPayoutUpdate", MenuType: menuTypeButton, Permission: "agent:payout:update", ParentName: "AgentList", Sort: 217},
			{Name: "AgentStatusUpdate", Title: "menu.route.agentStatusUpdate", MenuType: menuTypeButton, Permission: "agent:status:update", ParentName: "AgentList", Sort: 218},
			{Name: "AgentRemarkUpdate", Title: "menu.route.agentRemarkUpdate", MenuType: menuTypeButton, Permission: "agent:remark:update", ParentName: "AgentList", Sort: 219},
			{Name: "AgentProfileUpdate", Title: "menu.route.agentProfileUpdate", MenuType: menuTypeButton, Permission: "agent:profile:update", ParentName: "AgentList", Sort: 220},
			{Name: "AgentOpLogs", Title: "menu.route.agentOpLogs", MenuType: menuTypeButton, Permission: "agent:op-log:list", ParentName: "AgentList", Sort: 221},
			{Name: "AgentConfigUpdate", Title: "menu.route.agentConfigUpdate", MenuType: menuTypeButton, Permission: "agent:config:update", ParentName: "AgentConfig", Sort: 231},
			{Name: "AgentLevelUpdate", Title: "menu.route.agentLevelUpdate", MenuType: menuTypeButton, Permission: "agent:level:update", ParentName: "AgentLevel", Sort: 241},
			{Name: "AgentApplyReview", Title: "menu.route.agentApplyReview", HideMenu: 1, MenuType: menuTypeButton, Permission: "agent:apply:review", ParentName: "AgentApply", Sort: 251},

			// 代理存款订单审核菜单和按钮权限
			{Name: "AgentDepositOrderAudit", Title: "menu.route.agentDepositOrderAudit", Path: "/agent/deposit-order-audit", MenuType: menuTypeMenu, Component: "agent/depositOrderAudit/index", ParentName: "Agent", Sort: 28},
			{Name: "AgentDepositOrderClaim", Title: "menu.route.agentDepositOrderAudit", MenuType: menuTypeButton, Permission: "agent:deposit-order:claim", ParentName: "AgentDepositOrderAudit", Sort: 281},
			{Name: "AgentDepositOrderRelease", Title: "menu.route.agentDepositOrderAudit", MenuType: menuTypeButton, Permission: "agent:deposit-order:release", ParentName: "AgentDepositOrderAudit", Sort: 282},
			{Name: "AgentDepositOrderApprove", Title: "menu.route.agentDepositOrderAudit", MenuType: menuTypeButton, Permission: "agent:deposit-order:approve", ParentName: "AgentDepositOrderAudit", Sort: 283},
			{Name: "AgentDepositOrderReject", Title: "menu.route.agentDepositOrderAudit", MenuType: menuTypeButton, Permission: "agent:deposit-order:reject", ParentName: "AgentDepositOrderAudit", Sort: 284},
			{Name: "AgentDepositOrderFreeze", Title: "menu.route.agentDepositOrderAudit", MenuType: menuTypeButton, Permission: "agent:deposit-order:freeze", ParentName: "AgentDepositOrderAudit", Sort: 285},
			{Name: "AgentDepositOrderUnfreeze", Title: "menu.route.agentDepositOrderAudit", MenuType: menuTypeButton, Permission: "agent:deposit-order:unfreeze", ParentName: "AgentDepositOrderAudit", Sort: 286},

			// 财务中心
			{Name: "FinanceCenter", Title: "menu.route.financeCenter", Path: "/finance", MenuType: menuTypeDir, Sort: 40},
			{Name: "FinanceDepositAudit", Title: "menu.route.financeDepositAudit", Path: "/finance/deposit-audit", MenuType: menuTypeMenu, Component: "finance/depositAudit/index", ParentName: "FinanceCenter", Sort: 41},
			{Name: "FinanceMemberDepositAuditList", Title: "menu.route.financeMemberDepositAuditList", MenuType: menuTypeButton, Permission: "finance:member-deposit-audit:list", ParentName: "FinanceDepositAudit", Sort: 411},
			{Name: "FinanceSimulatedTurnoverList", Title: "menu.route.financeSimulatedTurnoverList", MenuType: menuTypeButton, Permission: "finance:simulated-turnover:list", ParentName: "FinanceDepositAudit", Sort: 412},

			// 优惠中心
			{Name: "PromoCenter", Title: "menu.route.promoCenter", Path: "/promo", MenuType: menuTypeDir, Sort: 30},
			{Name: "ActivityCenter", Title: "menu.route.activityCenter", Path: "/activity", MenuType: menuTypeMenu, Component: "activity/index", ParentName: "PromoCenter", Sort: 31},
			{Name: "TaskCenter", Title: "menu.route.taskCenter", Path: "/task", MenuType: menuTypeMenu, Component: "task/index", ParentName: "PromoCenter", Sort: 32},
		},
		Apis: []*coreclient.CreateApiReq{
			// 初始化 API
			{Path: "/init/status", Method: http.MethodGet, Description: "api.initStatus", ApiGroup: "api.group.public", ServiceName: serviceName},
			// 代理 API
			{Path: "/agent/list", Method: http.MethodPost, Description: "api.agentList", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/create", Method: http.MethodPost, Description: "api.agentCreate", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/search", Method: http.MethodPost, Description: "api.agentSearch", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/operator/list", Method: http.MethodGet, Description: "api.agentOperatorList", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/level/list", Method: http.MethodGet, Description: "api.agentLevelList", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/level/update", Method: http.MethodPost, Description: "api.agentLevelUpdate", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/config", Method: http.MethodGet, Description: "api.agentConfigGet", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/config/update", Method: http.MethodPost, Description: "api.agentConfigUpdate", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/line/defaults", Method: http.MethodGet, Description: "api.agentLineDefaults", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/tree/node", Method: http.MethodGet, Description: "api.agentTreeNode", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/tree/ancestors", Method: http.MethodGet, Description: "api.agentTreeAncestors", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/apply/list", Method: http.MethodPost, Description: "api.agentApplyList", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/apply/review", Method: http.MethodPost, Description: "api.agentApplyReview", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/profile", Method: http.MethodGet, Description: "api.agentProfileGet", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/password/reset", Method: http.MethodPost, Description: "api.agentPasswordReset", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/withdraw-password/reset", Method: http.MethodPost, Description: "api.agentWithdrawPasswordReset", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/share-rate/update", Method: http.MethodPost, Description: "api.agentShareRateUpdate", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/payout-restricted/update", Method: http.MethodPost, Description: "api.agentPayoutUpdate", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/status/update", Method: http.MethodPost, Description: "api.agentStatusUpdate", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/remark/update", Method: http.MethodPost, Description: "api.agentRemarkUpdate", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/profile/update", Method: http.MethodPost, Description: "api.agentProfileUpdate", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/op-logs", Method: http.MethodGet, Description: "api.agentOpLogs", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/login-log/list", Method: http.MethodPost, Description: "api.agentLoginLogList", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/login-log/associations", Method: http.MethodPost, Description: "api.agentLoginAssociations", ApiGroup: "api.group.agent", ServiceName: serviceName},

			//财务中心
			{Path: "/agent/deposit-order/add", Method: http.MethodPost, Description: "api.agentDepositOrderAdd", ApiGroup: "api.group.finance", ServiceName: serviceName},
			{Path: "/agent/deposit-order/list", Method: http.MethodPost, Description: "api.agentDepositOrderList", ApiGroup: "api.group.finance", ServiceName: serviceName},
			{Path: "/agent/deposit-order/claim", Method: http.MethodPost, Description: "api.agentDepositOrderClaim", ApiGroup: "api.group.finance", ServiceName: serviceName},
			{Path: "/agent/deposit-order/release", Method: http.MethodPost, Description: "api.agentDepositOrderRelease", ApiGroup: "api.group.finance", ServiceName: serviceName},
			{Path: "/agent/deposit-order/approve", Method: http.MethodPost, Description: "api.agentDepositOrderApprove", ApiGroup: "api.group.finance", ServiceName: serviceName},
			{Path: "/agent/deposit-order/reject", Method: http.MethodPost, Description: "api.agentDepositOrderReject", ApiGroup: "api.group.finance", ServiceName: serviceName},
			{Path: "/agent/deposit-order/freeze", Method: http.MethodPost, Description: "api.agentDepositOrderFreeze", ApiGroup: "api.group.finance", ServiceName: serviceName},
			{Path: "/agent/deposit-order/unfreeze", Method: http.MethodPost, Description: "api.agentDepositOrderUnfreeze", ApiGroup: "api.group.finance", ServiceName: serviceName},
			{Path: "/agent/turnover/add", Method: http.MethodPost, Description: "api.agentTurnoverAdd", ApiGroup: "api.group.finance", ServiceName: serviceName},
			{Path: "/agent/turnover/list", Method: http.MethodPost, Description: "api.agentTurnoverList", ApiGroup: "api.group.finance", ServiceName: serviceName},

			// Independent member CSV export permission.
			{Path: "/member/export/v2", Method: http.MethodPost, Description: "api.memberExport", ApiGroup: "api.group.member", ServiceName: serviceName},
			// 会员管理 API
			{Path: "/member/list", Method: http.MethodGet, Description: "api.memberList", ApiGroup: "api.group.member", ServiceName: serviceName},
			{Path: "/member/create", Method: http.MethodPost, Description: "api.memberCreate", ApiGroup: "api.group.member", ServiceName: serviceName},
			{Path: "/member/batch/create", Method: http.MethodPost, Description: "api.memberBatchCreate", ApiGroup: "api.group.member", ServiceName: serviceName},
			{Path: "/member/detail", Method: http.MethodGet, Description: "api.memberDetail", ApiGroup: "api.group.member", ServiceName: serviceName},
			{Path: "/member/search/code", Method: http.MethodGet, Description: "api.memberSearchByCode", ApiGroup: "api.group.member", ServiceName: serviceName},

			{Path: "/member/level/auto/update", Method: http.MethodPost, Description: "api.memberAutoLevelUpdate", ApiGroup: "api.group.member", ServiceName: serviceName},
			{Path: "/member/level/fixed/update", Method: http.MethodPost, Description: "api.memberFixedLevelsUpdate", ApiGroup: "api.group.member", ServiceName: serviceName},

			{Path: "/member/status/update", Method: http.MethodPost, Description: "api.memberUpdateStatus", ApiGroup: "api.group.member", ServiceName: serviceName},
			{Path: "/member/auth/errors/reset", Method: http.MethodPost, Description: "api.memberResetAuthErrors", ApiGroup: "api.group.member", ServiceName: serviceName},
			{Path: "/member/contact/binding-codes/list", Method: http.MethodGet, Description: "api.memberContactBindingCodeList", ApiGroup: "api.group.member", ServiceName: serviceName},
			{Path: "/member/contact/binding-codes/send", Method: http.MethodPost, Description: "api.memberContactBindingCodeSend", ApiGroup: "api.group.member", ServiceName: serviceName},
			{Path: "/member/contact/update", Method: http.MethodPost, Description: "api.memberUpdateContact", ApiGroup: "api.group.member", ServiceName: serviceName},
			{Path: "/member/images", Method: http.MethodGet, Description: "api.memberListImages", ApiGroup: "api.group.member", ServiceName: serviceName},
			{Path: "/member/remark/update", Method: http.MethodPost, Description: "api.memberUpdateRemark", ApiGroup: "api.group.member", ServiceName: serviceName},
			{Path: "/member/chat/update", Method: http.MethodPost, Description: "api.memberUpdateChat", ApiGroup: "api.group.member", ServiceName: serviceName},
			{Path: "/member/betting/update", Method: http.MethodPost, Description: "api.memberUpdateBetting", ApiGroup: "api.group.member", ServiceName: serviceName},
			{Path: "/member/withdrawal/update", Method: http.MethodPost, Description: "api.memberUpdateWithdrawal", ApiGroup: "api.group.member", ServiceName: serviceName},

			// 会员自动层级 API
			{Path: "/member/auto-level/list", Method: http.MethodGet, Description: "api.memberAutoLevelList", ApiGroup: "api.group.member", ServiceName: serviceName},
			{Path: "/member/auto-level/create", Method: http.MethodPost, Description: "api.memberAutoLevelCreate", ApiGroup: "api.group.member", ServiceName: serviceName},

			// 会员固定层级 API
			{Path: "/member/fixed-level/list", Method: http.MethodGet, Description: "api.memberFixedLevelList", ApiGroup: "api.group.member", ServiceName: serviceName},
			{Path: "/member/fixed-level/create", Method: http.MethodPost, Description: "api.memberFixedLevelCreate", ApiGroup: "api.group.member", ServiceName: serviceName},
			{Path: "/member/fixed-level/update", Method: http.MethodPost, Description: "api.memberFixedLevelUpdate", ApiGroup: "api.group.member", ServiceName: serviceName},
			{Path: "/member/fixed-level/delete", Method: http.MethodPost, Description: "api.memberFixedLevelDelete", ApiGroup: "api.group.member", ServiceName: serviceName},

			// VIP 基础配置 API
			{Path: "/member/vip/config", Method: http.MethodGet, Description: "api.memberVipConfigGet", ApiGroup: "api.group.member", ServiceName: serviceName},
			{Path: "/member/vip/config/update", Method: http.MethodPost, Description: "api.memberVipConfigUpdate", ApiGroup: "api.group.member", ServiceName: serviceName},

			// VIP 等级 API
			{Path: "/member/vip/level/list", Method: http.MethodGet, Description: "api.memberVipLevelList", ApiGroup: "api.group.member", ServiceName: serviceName},
			{Path: "/member/vip/level/detail", Method: http.MethodGet, Description: "api.memberVipLevelGet", ApiGroup: "api.group.member", ServiceName: serviceName},
			{Path: "/member/vip/level/create", Method: http.MethodPost, Description: "api.memberVipLevelCreate", ApiGroup: "api.group.member", ServiceName: serviceName},
			{Path: "/member/vip/level/update", Method: http.MethodPost, Description: "api.memberVipLevelUpdate", ApiGroup: "api.group.member", ServiceName: serviceName},
			{Path: "/member/vip/level/delete", Method: http.MethodPost, Description: "api.memberVipLevelDelete", ApiGroup: "api.group.member", ServiceName: serviceName},

			// 本站图片统一上传 API
			{
				Path:        "/resource/upload",   // 本站图片上传，不保存配置。
				Method:      http.MethodPost,      // multipart 单文件上传。
				Description: "api.resourceUpload", // 上传接口多语言名称。
				ApiGroup:    "resource",           // 统一图片素材接口目录。
				ServiceName: serviceName,          // 分站后台 API 服务。
			},
			// VIP 图标查询、保存和恢复 API
			{Path: "/member/vip/icon/list", Method: http.MethodGet, Description: "api.memberVipIconList", ApiGroup: "api.group.member", ServiceName: serviceName},
			{Path: "/member/vip/icon/update", Method: http.MethodPost, Description: "api.memberVipIconUpdate", ApiGroup: "api.group.member", ServiceName: serviceName},
			{Path: "/member/vip/icon/reset", Method: http.MethodPost, Description: "api.memberVipIconReset", ApiGroup: "api.group.member", ServiceName: serviceName},

			// 活动管理 API
			{Path: "/activity/create", Method: http.MethodPost, Description: "api.activityCreate", ApiGroup: "api.group.activity", ServiceName: serviceName},
			{Path: "/activity/update", Method: http.MethodPost, Description: "api.activityUpdate", ApiGroup: "api.group.activity", ServiceName: serviceName},
			{Path: "/activity/detail", Method: http.MethodGet, Description: "api.activityDetail", ApiGroup: "api.group.activity", ServiceName: serviceName},
			{Path: "/activity/list", Method: http.MethodGet, Description: "api.activityList", ApiGroup: "api.group.activity", ServiceName: serviceName},
			{Path: "/activity/close", Method: http.MethodPost, Description: "api.activityClose", ApiGroup: "api.group.activity", ServiceName: serviceName},
			{Path: "/activity/publish", Method: http.MethodPost, Description: "api.activityPublish", ApiGroup: "api.group.activity", ServiceName: serviceName},
			{Path: "/activity/enum-params", Method: http.MethodGet, Description: "api.activityEnumParams", ApiGroup: "api.group.activity", ServiceName: serviceName},

			// 活动订单 API
			{Path: "/activity/order/list", Method: http.MethodGet, Description: "api.activityOrderList", ApiGroup: "api.group.activityOrder", ServiceName: serviceName},
			{Path: "/activity/order/audit/list", Method: http.MethodGet, Description: "api.activityOrderAuditList", ApiGroup: "api.group.activityOrder", ServiceName: serviceName},
			{Path: "/activity/order/audit", Method: http.MethodPost, Description: "api.activityOrderAudit", ApiGroup: "api.group.activityOrder", ServiceName: serviceName},

			// 任务管理 API
			{Path: "/task/config/list", Method: http.MethodGet, Description: "api.taskConfigList", ApiGroup: "api.group.task", ServiceName: serviceName},
			{Path: "/task/config/update", Method: http.MethodPost, Description: "api.taskConfigUpdate", ApiGroup: "api.group.task", ServiceName: serviceName},
			{Path: "/task/reward/list", Method: http.MethodGet, Description: "api.taskRewardList", ApiGroup: "api.group.task", ServiceName: serviceName},
			{Path: "/task/reward/update-status", Method: http.MethodPost, Description: "api.taskRewardUpdateStatus", ApiGroup: "api.group.task", ServiceName: serviceName},
			{Path: "/task/global-config/list", Method: http.MethodGet, Description: "api.taskGlobalConfigList", ApiGroup: "api.group.task", ServiceName: serviceName},
			{Path: "/task/global-config/update", Method: http.MethodPost, Description: "api.taskGlobalConfigUpdate", ApiGroup: "api.group.task", ServiceName: serviceName},
		},
		I18N: i18nSeeds(code),
	}
}
