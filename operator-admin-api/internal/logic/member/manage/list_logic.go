// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package manage

import (
	"context"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"

	"github.com/zeromicro/go-zero/core/logx"
)

// ListLogic 会员列表逻辑
type ListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListLogic 创建会员列表逻辑
func NewListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLogic {
	return &ListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// List 获取会员列表
func (l *ListLogic) List(req *types.ListMembersRequest) (resp *types.ListMembersResponse, err error) {
	// 调用CRM会员列表RPC
	rpcResp, err := l.svcCtx.MemberRpc.List(l.ctx, &memberpb.ListMembersRequest{
		// 分页与关键字
		Page:     req.Page,     // 当前页码
		PageSize: req.PageSize, // 每页数量
		Keyword:  req.Keyword,  // 会员账号, 编号或昵称

		// 时间筛选
		TimeType: memberpb.MemberTimeType(req.TimeType), // 时间筛选类型
		StartAt:  req.StartAt,                           // 起始时间, 包含
		EndAt:    req.EndAt,                             // 结束时间, 不包含

		// 等级与注册信息
		AutoLevelId:     req.AutoLevelId,     // 自动层级ID
		FixedLevelId:    req.FixedLevelId,    // 固定层级ID
		RegisterMethod:  req.RegisterMethod,  // 注册方式
		RegisterChannel: req.RegisterChannel, // 注册渠道

		// 会员状态与代理归属
		Status:  req.Status,  // 会员状态
		AgentId: req.AgentId, // 所属分层代理ID

		// 其他筛选条件
		BettingDisabled: req.BettingDisabled, // 禁止投注
		NoPhone:         req.NoPhone,         // 未绑定手机号
		NoEmail:         req.NoEmail,         // 未绑定邮箱
		Deposited:       req.Deposited,       // 已充值
		NotDeposited:    req.NotDeposited,    // 未充值

		// 排序
		SortBy:    memberpb.MemberSortField(req.SortBy),    // 排序字段
		SortOrder: memberpb.MemberSortOrder(req.SortOrder), // 排序方向
	})
	if err != nil {
		return nil, err
	}

	// 组装会员列表
	list := make([]types.MemberListItem, 0, len(rpcResp.GetList()))
	for _, m := range rpcResp.GetList() {
		list = append(list, listMemberItem(m))
	}

	return &types.ListMembersResponse{
		Total: rpcResp.GetTotal(), // 符合条件的会员总数
		List:  list,               // 当前页会员列表
	}, nil
}

// listMemberItem 转换会员列表项
func listMemberItem(m *memberpb.MemberListItem) types.MemberListItem {
	return types.MemberListItem{
		// ==================== 会员基础信息 ====================

		Id:       m.Id,       // 会员ID
		Code:     m.Code,     // 12位会员编号
		Account:  m.Account,  // 会员登录账号
		Nickname: m.Nickname, // 会员昵称

		// ==================== 代理归属与等级 ====================

		AgentId:      m.AgentId,                       // 所属分层代理ID
		AgentAccount: m.AgentAccount,                  // 所属分层代理账号
		Levels:       listMemberLevels(m.GetLevels()), // 会员等级信息

		// ==================== 状态与注册信息 ====================

		Status:             m.Status,             // 会员状态
		ChatDisabled:       m.ChatDisabled,       // 禁止发言
		BettingDisabled:    m.BettingDisabled,    // 禁止投注
		WithdrawalDisabled: m.WithdrawalDisabled, // 禁止提款
		RegisterMethod:     m.RegisterMethod,     // 注册方式
		RegisterChannel:    m.RegisterChannel,    // 注册渠道
		CreatedAt:          m.CreatedAt,          // 注册时间
		LastLoginAt:        m.LastLoginAt,        // 最后登录时间
		LastLoginIp:        m.LastLoginIp,        // 最后登录IP

		// ==================== 联系方式与运营信息 ====================

		Phone:           m.Phone,           // 已脱敏手机号
		Email:           m.Email,           // 已脱敏邮箱
		InviterMemberId: m.InviterMemberId, // 邀请人会员ID
		Remark:          m.Remark,          // 后台会员备注

		// ==================== 会员统计 ====================

		Stats: listMemberStats(m.GetStats()), // 累计及当前周期统计
	}
}

// listMemberLevels 转换会员等级信息
func listMemberLevels(levels *memberpb.MemberLevelInfo) types.MemberLevelInfo {
	info := types.MemberLevelInfo{
		// 自动层级
		AutoLevelId:   levels.GetAutoLevelId(),   // 自动层级ID
		AutoLevelName: levels.GetAutoLevelName(), // 自动层级名称

		// 固定层级
		FixedLevels: make([]types.MemberListFixedLevelInfo, 0, len(levels.GetFixedLevels())),

		// VIP等级
		VipLevelId: levels.GetVipLevelId(), // VIP等级ID
		VipLevel:   levels.GetVipLevel(),   // VIP等级编号
	}

	// 组装已分配的固定层级
	for _, level := range levels.GetFixedLevels() {
		info.FixedLevels = append(info.FixedLevels, types.MemberListFixedLevelInfo{
			Id:   level.GetId(),   // 固定层级ID
			Name: level.GetName(), // 固定层级名称
		})
	}

	return info
}

// listMemberStats 转换会员统计信息, 无统计记录时返回零值
func listMemberStats(stats *memberpb.MemberStatsInfo) types.MemberStatsInfo {
	if stats == nil {
		return types.MemberStatsInfo{}
	}

	return types.MemberStatsInfo{
		// 累计资金统计
		TotalDeposit:               stats.TotalDeposit,               // 累计充值金额
		TotalDepositCount:          stats.TotalDepositCount,          // 累计充值成功次数
		TotalWithdrawal:            stats.TotalWithdrawal,            // 累计提现金额
		TotalWithdrawalCount:       stats.TotalWithdrawalCount,       // 累计提现成功次数
		TotalDepositWithdrawalDiff: stats.TotalDepositWithdrawalDiff, // 累计存提差额

		// 累计游戏统计
		TotalBet:      stats.TotalBet,      // 累计投注金额
		TotalValidBet: stats.TotalValidBet, // 累计有效投注金额
		TotalWater:    stats.TotalWater,    // 累计洗码量
		TotalWinLoss:  stats.TotalWinLoss,  // 累计会员净输赢

		// 当前结算周期统计
		CurrentPeriodDeposit:    stats.CurrentPeriodDeposit,    // 当前周期充值金额
		CurrentPeriodWithdrawal: stats.CurrentPeriodWithdrawal, // 当前周期提现金额
		CurrentPeriodBet:        stats.CurrentPeriodBet,        // 当前周期投注金额
		CurrentPeriodValidBet:   stats.CurrentPeriodValidBet,   // 当前周期有效投注金额
		CurrentPeriodWater:      stats.CurrentPeriodWater,      // 当前周期洗码量
		CurrentPeriodWinLoss:    stats.CurrentPeriodWinLoss,    // 当前周期会员净输赢

		// 当前结算周期开始时间
		CurrentPeriodStartedAt: stats.CurrentPeriodStartedAt,
	}
}
