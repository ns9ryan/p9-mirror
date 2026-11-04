package memberlogic

import (
	"context"
	"sort"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"
)

// GetDetailLogic 会员详情逻辑
type GetDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewGetDetailLogic 创建会员详情逻辑
func NewGetDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDetailLogic {
	return &GetDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetDetail 获取会员详情
func (l *GetDetailLogic) GetDetail(in *memberpb.GetMemberDetailRequest) (*memberpb.GetMemberDetailResponse, error) {
	// 校验分站运营身份和当前厅
	if _, ok := ctxdata.OperatorClaimsFromCtx(l.ctx); !ok {
		return nil, xerr.RpcErr(xerr.Unauthorized(i18n.Unauthorized))
	}

	// 校验会员ID
	if in.GetMemberId() <= 0 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	// 查询会员详情及关联信息
	m, err := l.svcCtx.Member.GetDetail(l.ctx, in.GetMemberId())
	if err != nil {
		if ent.IsNotFound(err) {
			// 会员不存在或不属于当前厅
			return nil, xerr.RpcErr(xerr.EntNotFound(i18n.InvalidParam, err))
		}
		return nil, xerr.RpcErr(err)
	}

	return memberDetailResponse(m), nil
}

// memberDetailResponse 组装会员详情响应
func memberDetailResponse(m *ent.Member) *memberpb.GetMemberDetailResponse {
	resp := &memberpb.GetMemberDetailResponse{
		// ==================== 会员基本信息 ====================

		Id:   m.ID,   // 会员ID
		Code: m.Code, // 会员编号

		// 代理归属
		AgentId: m.AgentID, // 所属分层代理ID

		// 会员等级
		Levels: &memberpb.MemberLevelInfo{
			AutoLevelId: m.AutoLevelID, // 当前自动层级ID
			VipLevelId:  m.VipLevelID,  // 当前VIP等级ID
		},

		// 状态与业务限制
		Status:             m.Status,             // 会员状态
		ChatDisabled:       m.ChatDisabled,       // 是否禁止发言
		BettingDisabled:    m.BettingDisabled,    // 是否禁止投注
		WithdrawalDisabled: m.WithdrawalDisabled, // 是否禁止提款

		// 注册信息
		RegisterMethod:     m.RegisterMethod,        // 注册方式
		RegisterChannel:    m.RegisterChannel,       // 注册渠道
		CreatedAt:          m.CreatedAt.UnixMilli(), // 注册时间
		RegisterIp:         m.RegisterIP,            // 注册IP
		RegisterDeviceCode: m.RegisterDeviceCode,    // 注册设备标识
		RegisterDomain:     m.RegisterDomain,        // 注册域名

		// 最后登录信息
		LastLoginIp: m.LastLoginIP, // 最后登录IP

		// 运营信息
		InviterMemberId: m.InviterMemberID, // 邀请人会员ID

		// 统计信息
		Stats: listMemberStats(m.Edges.Stats),
	}

	// ==================== 关联账号与个人资料 ====================

	// 所属分层代理
	if agent := m.Edges.Agent; agent != nil {
		resp.AgentAccount = agent.Account        // 所属分层代理账号
		resp.AgentLineCode = agent.AgentLineCode // 所属代理线路编码
	}

	// 会员登录账号
	if auth := m.Edges.Auth; auth != nil {
		resp.Account = auth.Username
	}

	// 会员个人资料
	if profile := m.Edges.Profile; profile != nil {
		resp.Nickname = profile.Nickname             // 会员昵称
		resp.AvatarUrl = profile.AvatarURL           // 会员头像URL
		resp.AvatarFrameUrl = profile.AvatarFrameURL // 会员头像框URL
		resp.Gender = profile.Gender                 // 性别
		resp.RegionCode = profile.RegionCode         // 国家或地区编码
		resp.Remark = profile.Remark                 // 后台会员备注

		// 会员生日, YYYY-MM-DD
		if profile.Birthday != nil {
			resp.Birthday = new(profile.Birthday.Format("2006-01-02"))
		}
	}

	// 最后登录时间, Unix毫秒时间戳
	if m.LastLoginAt != nil {
		resp.LastLoginAt = new(m.LastLoginAt.UnixMilli())
	}

	// ==================== 会员等级信息 ====================

	// 当前自动层级名称
	if level := m.Edges.AutoLevel; level != nil {
		resp.Levels.AutoLevelName = level.Name
	}

	// 当前VIP等级编号
	if level := m.Edges.VipLevel; level != nil {
		resp.Levels.VipLevel = level.Level
	}

	// 已分配的固定层级
	for _, assignment := range m.Edges.FixedLevelAssignments {
		level := assignment.Edges.FixedLevel
		if level == nil {
			continue
		}

		resp.Levels.FixedLevels = append(resp.Levels.FixedLevels, &memberpb.MemberFixedLevelInfo{
			Id:   level.ID,   // 固定层级ID
			Name: level.Name, // 固定层级名称
		})
	}

	// 固定层级按ID排序, 保证响应顺序稳定
	sort.Slice(resp.Levels.FixedLevels, func(i, j int) bool {
		return resp.Levels.FixedLevels[i].Id < resp.Levels.FixedLevels[j].Id
	})

	return resp
}
