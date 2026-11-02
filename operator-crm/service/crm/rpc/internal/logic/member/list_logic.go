package memberlogic

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/paging"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	memberdao "oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao/member"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"
)

// ListLogic 会员列表逻辑
type ListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewListLogic 创建会员列表逻辑
func NewListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLogic {
	return &ListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// List 获取会员列表
func (l *ListLogic) List(in *memberpb.ListMembersRequest) (*memberpb.ListMembersResponse, error) {
	// 校验分站运营身份和当前厅
	claims := ctxdata.ClaimsFromCtx(l.ctx)
	if claims == nil ||
		claims.UserID <= 0 ||
		claims.UserType != ctxdata.UserTypeOperator ||
		claims.OperatorCode == "" {
		return nil, xerr.RpcErr(xerr.Unauthorized(i18n.Unauthorized))
	}

	// 校验并转换列表查询参数
	params, err := parseListParams(in)
	if err != nil {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	// 查询当前厅会员及关联信息
	members, total, err := l.svcCtx.Member.List(l.ctx, params)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}

	// 组装会员列表
	list := make([]*memberpb.MemberListItem, 0, len(members))
	for _, m := range members {
		list = append(list, listMemberItem(m))
	}

	return &memberpb.ListMembersResponse{
		Total: total, // 符合条件的会员总数
		List:  list,  // 当前页会员列表
	}, nil
}

// parseListParams 校验并转换会员列表查询参数
func parseListParams(in *memberpb.ListMembersRequest) (memberdao.ListParams, error) {
	var params memberdao.ListParams

	// ==================== 分页与关键字 ====================

	// 规范化分页参数
	offset, limit, ok := paging.OffsetLimit(in.GetPage(), in.GetPageSize())
	if !ok {
		return params, errors.New("会员列表分页参数无效")
	}

	params.Offset = offset
	params.Limit = limit
	params.Keyword = strings.TrimSpace(in.GetKeyword())

	// ==================== 时间筛选 ====================

	// 时间筛选类型
	switch in.GetTimeType() {
	case memberpb.MemberTimeType_MEMBER_TIME_TYPE_UNSPECIFIED:
		params.TimeField = memberdao.ListTimeNone // 不筛选时间

	case memberpb.MemberTimeType_MEMBER_TIME_TYPE_REGISTER:
		params.TimeField = memberdao.ListTimeRegister // 注册时间

	case memberpb.MemberTimeType_MEMBER_TIME_TYPE_LAST_LOGIN:
		params.TimeField = memberdao.ListTimeLastLogin // 最后登录时间

	default:
		return params, fmt.Errorf("不支持的时间筛选类型: %d", in.GetTimeType())
	}

	// 校验时间筛选条件
	if params.TimeField == memberdao.ListTimeNone {
		if in.StartAt != nil || in.EndAt != nil {
			return params, errors.New("未指定时间筛选类型")
		}
	} else if in.StartAt == nil && in.EndAt == nil {
		return params, errors.New("时间筛选范围不能为空")
	}

	// 校验起止时间顺序
	if in.StartAt != nil && in.EndAt != nil && *in.StartAt >= *in.EndAt {
		return params, errors.New("起始时间必须早于结束时间")
	}

	// 转换起始时间, 包含该时刻
	if in.StartAt != nil {
		start := time.UnixMilli(*in.StartAt)
		if start.Year() < 1 || start.Year() > 9999 {
			return params, errors.New("起始时间超出有效范围")
		}
		params.StartAt = &start
	}

	// 转换结束时间, 不包含该时刻
	if in.EndAt != nil {
		end := time.UnixMilli(*in.EndAt)
		if end.Year() < 1 || end.Year() > 9999 {
			return params, errors.New("结束时间超出有效范围")
		}
		params.EndAt = &end
	}

	// ==================== 等级与注册信息 ====================

	// 自动层级ID
	if in.AutoLevelId != nil {
		if *in.AutoLevelId <= 0 {
			return params, errors.New("自动层级ID无效")
		}
		params.AutoLevelID = in.AutoLevelId
	}

	// 固定层级ID
	if in.FixedLevelId != nil {
		if *in.FixedLevelId <= 0 {
			return params, errors.New("固定层级ID无效")
		}
		params.FixedLevelID = in.FixedLevelId
	}

	// 注册方式
	if in.RegisterMethod != nil {
		method := strings.TrimSpace(*in.RegisterMethod)
		if method == "" || len(method) > 32 {
			return params, errors.New("注册方式无效")
		}
		params.RegisterMethod = &method
	}

	// 注册渠道
	if in.RegisterChannel != nil {
		channel := strings.TrimSpace(*in.RegisterChannel)
		if channel == "" || len(channel) > 32 {
			return params, errors.New("注册渠道无效")
		}
		params.RegisterChannel = &channel
	}

	// ==================== 会员状态与其他筛选 ====================

	// 会员状态
	if in.Status != nil {
		if *in.Status < 1 || *in.Status > 3 {
			return params, errors.New("会员状态无效")
		}
		params.Status = in.Status
	}

	// 所属分层代理ID
	if in.AgentId != nil {
		if *in.AgentId <= 0 {
			return params, errors.New("分层代理ID无效")
		}
		params.AgentID = in.AgentId
	}

	// 业务限制和联系方式筛选
	params.BettingDisabled = in.GetBettingDisabled()
	params.NoPhone = in.GetNoPhone()
	params.NoEmail = in.GetNoEmail()

	// 充值筛选, 两个条件不能同时启用
	if in.GetDeposited() && in.GetNotDeposited() {
		return params, errors.New("充值筛选条件冲突")
	}

	if in.GetDeposited() {
		deposited := true
		params.Deposited = &deposited
	} else if in.GetNotDeposited() {
		deposited := false
		params.Deposited = &deposited
	}

	// ==================== 排序 ====================

	// 会员列表排序字段
	switch in.GetSortBy() {
	case memberpb.MemberSortField_MEMBER_SORT_FIELD_UNSPECIFIED:
		params.SortField = memberdao.ListSortRegister // 注册时间

	case memberpb.MemberSortField_MEMBER_SORT_FIELD_TOTAL_DEPOSIT:
		params.SortField = memberdao.ListSortTotalDeposit // 累计充值金额

	case memberpb.MemberSortField_MEMBER_SORT_FIELD_TOTAL_DEPOSIT_COUNT:
		params.SortField = memberdao.ListSortTotalDepositCount // 累计充值成功次数

	case memberpb.MemberSortField_MEMBER_SORT_FIELD_TOTAL_WITHDRAWAL:
		params.SortField = memberdao.ListSortTotalWithdrawal // 累计提现金额

	case memberpb.MemberSortField_MEMBER_SORT_FIELD_TOTAL_WITHDRAWAL_COUNT:
		params.SortField = memberdao.ListSortTotalWithdrawalCount // 累计提现成功次数

	default:
		return params, fmt.Errorf("不支持的会员排序字段: %d", in.GetSortBy())
	}

	// 排序方向, 未指定时默认降序
	switch in.GetSortOrder() {
	case memberpb.MemberSortOrder_MEMBER_SORT_ORDER_UNSPECIFIED,
		memberpb.MemberSortOrder_MEMBER_SORT_ORDER_DESC:
		params.SortDesc = true // 降序

	case memberpb.MemberSortOrder_MEMBER_SORT_ORDER_ASC:
		params.SortDesc = false // 升序

	default:
		return params, fmt.Errorf("不支持的会员排序方向: %d", in.GetSortOrder())
	}

	return params, nil
}

// listMemberItem 组装会员列表项
func listMemberItem(m *ent.Member) *memberpb.MemberListItem {
	// ==================== 会员基础信息 ====================

	item := &memberpb.MemberListItem{
		// 会员基本信息
		Id:   m.ID,   // 会员ID
		Code: m.Code, // 12位会员编号

		// 代理归属
		AgentId: m.AgentID, // 所属分层代理ID

		// 会员等级
		Levels: &memberpb.MemberLevelInfo{
			AutoLevelId: m.AutoLevelID, // 当前自动层级ID
			VipLevelId:  m.VipLevelID,  // 当前VIP等级ID
		},

		// 状态与注册信息
		Status:             m.Status,                // 会员状态
		ChatDisabled:       m.ChatDisabled,          // 禁止发言
		BettingDisabled:    m.BettingDisabled,       // 禁止投注
		WithdrawalDisabled: m.WithdrawalDisabled,    // 禁止提款
		RegisterMethod:     m.RegisterMethod,        // 注册方式
		RegisterChannel:    m.RegisterChannel,       // 注册渠道
		CreatedAt:          m.CreatedAt.UnixMilli(), // 注册时间
		LastLoginIp:        m.LastLoginIP,           // 最后登录IP
		InviterMemberId:    m.InviterMemberID,       // 邀请人会员ID

		// 会员统计
		Stats: listMemberStats(m.Edges.Stats),
	}

	// 最后登录时间
	if m.LastLoginAt != nil {
		lastLoginAt := m.LastLoginAt.UnixMilli()
		item.LastLoginAt = &lastLoginAt
	}

	// ==================== 关联账号与资料 ====================

	// 所属分层代理账号
	if agent := m.Edges.Agent; agent != nil {
		item.AgentAccount = agent.Account
	}

	// 会员登录账号
	if auth := m.Edges.Auth; auth != nil && auth.Username != nil {
		item.Account = *auth.Username
	}

	// 会员昵称和运营备注
	if profile := m.Edges.Profile; profile != nil {
		item.Nickname = profile.Nickname
		item.Remark = profile.Remark
	}

	// ==================== 会员等级信息 ====================

	// 当前自动层级名称
	if level := m.Edges.AutoLevel; level != nil {
		item.Levels.AutoLevelName = level.Name
	}

	// 当前VIP等级编号
	if level := m.Edges.VipLevel; level != nil {
		item.Levels.VipLevel = level.Level
	}

	// 已分配的固定层级
	for _, assignment := range m.Edges.FixedLevelAssignments {
		level := assignment.Edges.FixedLevel
		if level == nil {
			continue
		}
		item.Levels.FixedLevels = append(item.Levels.FixedLevels, &memberpb.MemberFixedLevelInfo{
			Id:   level.ID,   // 固定层级ID
			Name: level.Name, // 固定层级名称
		})
	}

	// 固定层级按ID排序, 保证响应顺序稳定
	sort.Slice(item.Levels.FixedLevels, func(i, j int) bool {
		return item.Levels.FixedLevels[i].Id < item.Levels.FixedLevels[j].Id
	})

	// 手机号和邮箱暂不返回, 避免泄露密文或检索摘要
	return item
}

// listMemberStats 组装会员统计信息, 无统计记录时返回零值
func listMemberStats(stats *ent.MemberStats) *memberpb.MemberStatsInfo {
	if stats == nil {
		return &memberpb.MemberStatsInfo{}
	}

	info := &memberpb.MemberStatsInfo{
		// 累计资金统计
		TotalDeposit:               stats.TotalDeposit,                         // 累计充值金额
		TotalDepositCount:          stats.TotalDepositCount,                    // 累计充值成功次数
		TotalWithdrawal:            stats.TotalWithdrawal,                      // 累计提现金额
		TotalWithdrawalCount:       stats.TotalWithdrawalCount,                 // 累计提现成功次数
		TotalDepositWithdrawalDiff: stats.TotalDeposit - stats.TotalWithdrawal, // 累计存提差额

		// 累计游戏统计
		TotalBet:      stats.TotalBet,      // 累计投注金额
		TotalValidBet: stats.TotalValidBet, // 累计有效投注金额
		TotalWater:    stats.TotalWater,    // 累计洗码量
		TotalWinLoss:  stats.TotalWinLoss,  // 累计会员净输赢

		// 当前结算周期统计
		CurrentPeriodDeposit:    stats.LastDeposit,    // 当前周期充值金额
		CurrentPeriodWithdrawal: stats.LastWithdrawal, // 当前周期提现金额
		CurrentPeriodBet:        stats.LastBet,        // 当前周期投注金额
		CurrentPeriodValidBet:   stats.LastValidBet,   // 当前周期有效投注金额
		CurrentPeriodWater:      stats.LastWater,      // 当前周期洗码量
		CurrentPeriodWinLoss:    stats.LastWinLoss,    // 当前周期会员净输赢
	}

	// 当前结算周期开始时间
	if !stats.PeriodStartedAt.IsZero() {
		startedAt := stats.PeriodStartedAt.UnixMilli()
		info.CurrentPeriodStartedAt = &startedAt
	}

	return info
}
