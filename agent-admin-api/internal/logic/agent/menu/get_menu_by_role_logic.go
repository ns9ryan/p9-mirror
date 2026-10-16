// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package menu

import (
	"context"

	"oa.98ent.com/p9/agent-admin-api/internal/svc"
	"oa.98ent.com/p9/agent-admin-api/internal/types"
	"oa.98ent.com/p9/common/i18n"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMenuByRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 当前菜单树
func NewGetMenuByRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMenuByRoleLogic {
	return &GetMenuByRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetMenuByRole 返回写死的代理后台菜单树，标题按当前语言翻译。
func (l *GetMenuByRoleLogic) GetMenuByRole() (resp []types.MenuNode, err error) {
	return translateMenus(l.ctx, agentMenus()), nil
}

func translateMenus(ctx context.Context, nodes []types.MenuNode) []types.MenuNode {
	for i := range nodes {
		nodes[i].Title = i18n.TG(ctx, i18n.CodeAgent, i18n.GroupMenu, nodes[i].Title)
		nodes[i].Children = translateMenus(ctx, nodes[i].Children)
	}
	return nodes
}

const (
	menuTypeDir     int32 = 0
	menuTypeMenu    int32 = 1
	layoutComponent       = "LAYOUT"
)

func agentMenus() []types.MenuNode {
	return []types.MenuNode{
		page(1, 0, 1, "Dashboard", "menu.route.dashboard", "/dashboard"),
		dir(2, 0, 2, "Account", "menu.route.account", "/account", "/account/member",
			page(3, 2, 1, "MemberList", "menu.route.member", "/account/member"),
			page(4, 2, 2, "AgentList", "menu.route.agent", "/account/agent"),
		),
		dir(5, 0, 3, "Report", "menu.route.report", "/report", "/report/win-loss",
			page(6, 5, 1, "WinLossReport", "menu.route.winLoss", "/report/win-loss"),
			page(7, 5, 2, "GameReport", "menu.route.gameReport", "/report/game"),
			page(8, 5, 3, "CommissionReport", "menu.route.commission", "/report/commission"),
		),
		dir(9, 0, 4, "Finance", "menu.route.finance", "/finance", "/finance/withdraw-apply",
			page(10, 9, 1, "WithdrawApply", "menu.route.withdrawApply", "/finance/withdraw-apply"),
			page(11, 9, 2, "DepositOrder", "menu.route.depositOrder", "/finance/deposit-order"),
			page(12, 9, 3, "WithdrawOrder", "menu.route.withdrawOrder", "/finance/withdraw-order"),
		),
		dir(13, 0, 5, "System", "menu.route.system", "/system", "/system/profile",
			page(14, 13, 1, "Profile", "menu.route.profile", "/system/profile"),
			page(15, 13, 2, "AdminLog", "menu.route.adminLog", "/system/log"),
		),
	}
}

func dir(id, parent int64, sort int32, name, title, path, redirect string, children ...types.MenuNode) types.MenuNode {
	if children == nil {
		children = []types.MenuNode{}
	}
	return types.MenuNode{
		Id:        id,
		ParentId:  parent,
		MenuType:  menuTypeDir,
		Path:      path,
		Name:      name,
		Component: layoutComponent,
		Redirect:  redirect,
		Title:     title,
		Sort:      sort,
		Children:  children,
	}
}

func page(id, parent int64, sort int32, name, title, path string) types.MenuNode {
	return types.MenuNode{
		Id:        id,
		ParentId:  parent,
		MenuType:  menuTypeMenu,
		Path:      path,
		Name:      name,
		Component: path,
		Title:     title,
		Sort:      sort,
		Children:  []types.MenuNode{},
	}
}
