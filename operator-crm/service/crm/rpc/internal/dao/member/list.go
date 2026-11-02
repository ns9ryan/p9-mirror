package member

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/agentuser"
	entmember "oa.98ent.com/p9/operator-crm/ent/member"
	"oa.98ent.com/p9/operator-crm/ent/memberauth"
	"oa.98ent.com/p9/operator-crm/ent/memberfixedlevelassignment"
	"oa.98ent.com/p9/operator-crm/ent/memberprofile"
	"oa.98ent.com/p9/operator-crm/ent/memberstats"
	"oa.98ent.com/p9/operator-crm/ent/predicate"
)

// ListTimeField 会员列表时间筛选字段
type ListTimeField uint8

const (
	ListTimeNone      ListTimeField = iota // 不筛选时间
	ListTimeRegister                       // 注册时间
	ListTimeLastLogin                      // 最后登录时间
)

// ListSortField 会员列表排序字段
type ListSortField uint8

const (
	ListSortRegister             ListSortField = iota // 注册时间
	ListSortTotalDeposit                              // 累计充值金额
	ListSortTotalDepositCount                         // 累计充值次数
	ListSortTotalWithdrawal                           // 累计提现金额
	ListSortTotalWithdrawalCount                      // 累计提现次数
)

// ListParams 会员列表查询参数
type ListParams struct {
	// 分页与关键字
	Offset  int    // 查询偏移量
	Limit   int    // 查询数量
	Keyword string // 会员账号, 编号或昵称

	// 时间筛选
	TimeField ListTimeField // 时间筛选字段
	StartAt   *time.Time    // 起始时间, 包含
	EndAt     *time.Time    // 结束时间, 不包含

	// 等级与注册信息
	AutoLevelID     *int64  // 自动层级ID
	FixedLevelID    *int64  // 固定层级ID
	RegisterMethod  *string // 注册方式
	RegisterChannel *string // 注册渠道

	// 状态与代理归属
	Status  *int64 // 会员状态
	AgentID *int64 // 所属分层代理ID

	// 其他筛选条件
	BettingDisabled bool  // 仅查询禁止投注的会员
	NoPhone         bool  // 仅查询未绑定手机号的会员
	NoEmail         bool  // 仅查询未绑定邮箱的会员
	Deposited       *bool // nil不筛选, true已充值, false未充值

	// 排序
	SortField ListSortField // 排序字段
	SortDesc  bool          // 统计字段是否降序
}

// List 查询当前厅会员列表及关联信息
func (s *Service) List(ctx context.Context, in ListParams) ([]*ent.Member, int64, error) {
	// 校验分页参数, 防止无效分页进入数据库查询
	if in.Offset < 0 || in.Limit <= 0 || in.Limit > 100 {
		return nil, 0, fmt.Errorf("会员列表分页参数无效")
	}

	// 构建会员筛选条件
	filters, err := listWhere(in)
	if err != nil {
		return nil, 0, err
	}

	query := s.db.Member.Query().Where(filters...)

	// 统计符合条件的会员总数
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("统计会员数量失败: %w", err)
	}
	if total == 0 {
		return nil, 0, nil
	}

	// 设置会员列表排序
	if err := listOrder(query, in.SortField, in.SortDesc); err != nil {
		return nil, 0, err
	}

	// 加载所属分层代理, 用于显示代理账号
	query.WithAgent(func(q *ent.AgentUserQuery) {
		q.Select(
			agentuser.FieldID,      // 代理ID
			agentuser.FieldAccount, // 代理登录账号
		)
	})

	// 加载会员认证信息, 不读取密码和联系方式密文
	query.WithAuth(func(q *ent.MemberAuthQuery) {
		q.Select(
			memberauth.FieldID,       // 认证记录ID
			memberauth.FieldMemberID, // 所属会员ID
			memberauth.FieldUsername, // 会员登录账号
		)
	})

	// 加载会员个人资料
	query.WithProfile(func(q *ent.MemberProfileQuery) {
		q.Select(
			memberprofile.FieldID,       // 个人资料ID
			memberprofile.FieldMemberID, // 所属会员ID
			memberprofile.FieldNickname, // 会员昵称
			memberprofile.FieldRemark,   // 运营备注
		)
	})

	// 加载会员自动层级
	query.WithAutoLevel()

	// 加载会员VIP等级
	query.WithVipLevel()

	// 加载会员固定层级及分配信息
	query.WithFixedLevelAssignments(func(q *ent.MemberFixedLevelAssignmentQuery) {
		q.WithFixedLevel()
	})

	// 加载会员累计及当前周期统计, 允许统计记录不存在
	query.WithStats()

	// 查询当前页会员
	rows, err := query.
		Offset(in.Offset).
		Limit(in.Limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("查询会员列表失败: %w", err)
	}

	return rows, int64(total), nil

}

// listWhere 构建会员列表查询条件
func listWhere(in ListParams) ([]predicate.Member, error) {
	filters := make([]predicate.Member, 0, 12)

	// 关键字匹配会员账号, 编号或昵称
	if in.Keyword != "" {
		filters = append(filters, entmember.Or(
			entmember.CodeContains(in.Keyword),                                       // 会员编号, 模糊匹配
			entmember.HasAuthWith(memberauth.UsernameContainsFold(in.Keyword)),       // 登录账号, 不区分大小写
			entmember.HasProfileWith(memberprofile.NicknameContainsFold(in.Keyword)), // 会员昵称, 不区分大小写
		))
	}

	// 时间筛选, 使用左闭右开区间
	switch in.TimeField {
	case ListTimeNone:
		// 不筛选时间

	case ListTimeRegister:
		// 按会员注册时间筛选
		if in.StartAt != nil {
			filters = append(filters, entmember.CreatedAtGTE(*in.StartAt))
		}
		if in.EndAt != nil {
			filters = append(filters, entmember.CreatedAtLT(*in.EndAt))
		}

	case ListTimeLastLogin:
		// 按会员最后登录时间筛选
		if in.StartAt != nil {
			filters = append(filters, entmember.LastLoginAtGTE(*in.StartAt))
		}
		if in.EndAt != nil {
			filters = append(filters, entmember.LastLoginAtLT(*in.EndAt))
		}

	default:
		return nil, fmt.Errorf("不支持的会员时间筛选字段: %d", in.TimeField)
	}

	// 自动层级筛选
	if in.AutoLevelID != nil {
		filters = append(filters, entmember.AutoLevelIDEQ(*in.AutoLevelID))
	}

	// 固定层级筛选
	if in.FixedLevelID != nil {
		filters = append(filters, entmember.HasFixedLevelAssignmentsWith(
			memberfixedlevelassignment.FixedLevelIDEQ(*in.FixedLevelID),
		))
	}

	// 注册方式筛选
	if in.RegisterMethod != nil {
		filters = append(filters, entmember.RegisterMethodEQ(*in.RegisterMethod))
	}

	// 注册渠道筛选
	if in.RegisterChannel != nil {
		filters = append(filters, entmember.RegisterChannelEQ(*in.RegisterChannel))
	}

	// 会员状态筛选
	if in.Status != nil {
		filters = append(filters, entmember.StatusEQ(*in.Status))
	}

	// 所属分层代理筛选
	if in.AgentID != nil {
		filters = append(filters, entmember.AgentIDEQ(*in.AgentID))
	}

	// 禁止投注筛选
	if in.BettingDisabled {
		filters = append(filters, entmember.BettingDisabled(true))
	}

	// 未绑定手机号筛选, 以手机号检索摘要判断
	if in.NoPhone {
		filters = append(filters, entmember.Not(
			entmember.HasAuthWith(memberauth.PhoneLookupHashNotNil()),
		))
	}

	// 未绑定邮箱筛选, 以邮箱检索摘要判断
	if in.NoEmail {
		filters = append(filters, entmember.Not(
			entmember.HasAuthWith(memberauth.EmailLookupHashNotNil()),
		))
	}

	// 充值筛选, 无统计记录的会员视为未充值
	if in.Deposited != nil {
		deposited := entmember.HasStatsWith(
			memberstats.TotalDepositCountGT(0),
		)
		if *in.Deposited {
			filters = append(filters, deposited)
		} else {
			filters = append(filters, entmember.Not(deposited))
		}
	}

	return filters, nil
}

// listOrder 设置会员列表排序, 使用会员ID保证分页顺序稳定
func listOrder(query *ent.MemberQuery, field ListSortField, desc bool) error {
	// 注册时间排序, 默认最新注册的会员在前
	if field == ListSortRegister {
		query.Order(
			entmember.ByCreatedAt(sql.OrderDesc()), // 注册时间倒序
			entmember.ByID(sql.OrderDesc()),        // 会员ID倒序, 保证顺序稳定
		)
		return nil
	}

	// 选择会员统计排序字段
	var statsField string
	switch field {
	case ListSortTotalDeposit:
		statsField = memberstats.FieldTotalDeposit // 累计充值金额

	case ListSortTotalDepositCount:
		statsField = memberstats.FieldTotalDepositCount // 累计充值成功次数

	case ListSortTotalWithdrawal:
		statsField = memberstats.FieldTotalWithdrawal // 累计提现金额

	case ListSortTotalWithdrawalCount:
		statsField = memberstats.FieldTotalWithdrawalCount // 累计提现成功次数

	default:
		return fmt.Errorf("不支持的会员列表排序字段: %d", field)
	}

	// 升序排列, 无统计记录的会员排在前面
	opts := []sql.OrderTermOption{
		sql.OrderAsc(),
		sql.OrderNullsFirst(),
	}

	// 降序排列, 无统计记录的会员排在后面
	if desc {
		opts = []sql.OrderTermOption{
			sql.OrderDesc(),
			sql.OrderNullsLast(),
		}
	}

	// 按统计字段排序, 会员ID作为次级排序
	query.Order(
		entmember.ByStatsField(statsField, opts...), // 指定统计字段及排序方向
		entmember.ByID(sql.OrderDesc()),             // 会员ID倒序, 保证顺序稳定
	)

	return nil
}
