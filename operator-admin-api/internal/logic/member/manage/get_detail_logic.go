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

// GetDetailLogic 会员详情逻辑
type GetDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetDetailLogic 创建会员详情逻辑
func NewGetDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDetailLogic {
	return &GetDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetDetail 获取会员详情
func (l *GetDetailLogic) GetDetail(req *types.GetMemberDetailRequest) (resp *types.GetMemberDetailResponse, err error) {
	// 调用CRM会员详情RPC
	rpcResp, err := l.svcCtx.MemberRpc.GetDetail(l.ctx, &memberpb.GetMemberDetailRequest{
		MemberId: req.MemberId, // 会员数据库主键
	})
	if err != nil {
		return nil, err
	}

	// 转换会员详情
	return &types.GetMemberDetailResponse{
		// ==================== 会员基本信息 ====================

		Id:             rpcResp.Id,             // 会员ID
		Code:           rpcResp.Code,           // 会员对外编号
		Account:        rpcResp.Account,        // 会员登录账号
		Nickname:       rpcResp.Nickname,       // 会员昵称
		AvatarUrl:      rpcResp.AvatarUrl,      // 会员头像URL
		AvatarFrameUrl: rpcResp.AvatarFrameUrl, // 会员头像框URL
		Birthday:       rpcResp.Birthday,       // 会员生日
		Gender:         rpcResp.Gender,         // 性别
		RegionCode:     rpcResp.RegionCode,     // 国家或地区编码

		// ==================== 代理归属 ====================

		AgentId:       rpcResp.AgentId,       // 所属分层代理ID
		AgentAccount:  rpcResp.AgentAccount,  // 所属分层代理账号
		AgentLineCode: rpcResp.AgentLineCode, // 所属代理线路编码

		// ==================== 会员等级 ====================

		Levels: listMemberLevels(rpcResp.GetLevels()), // 会员当前等级信息

		// ==================== 状态与业务限制 ====================

		Status:             rpcResp.Status,             // 会员状态
		ChatDisabled:       rpcResp.ChatDisabled,       // 是否禁止发言
		BettingDisabled:    rpcResp.BettingDisabled,    // 是否禁止投注
		WithdrawalDisabled: rpcResp.WithdrawalDisabled, // 是否禁止提款

		// ==================== 注册信息 ====================

		RegisterMethod:     rpcResp.RegisterMethod,     // 注册方式
		RegisterChannel:    rpcResp.RegisterChannel,    // 注册渠道
		CreatedAt:          rpcResp.CreatedAt,          // 注册时间
		RegisterIp:         rpcResp.RegisterIp,         // 注册IP
		RegisterDeviceCode: rpcResp.RegisterDeviceCode, // 注册设备标识
		RegisterDomain:     rpcResp.RegisterDomain,     // 注册域名

		// ==================== 最后登录信息 ====================

		LastLoginAt: rpcResp.LastLoginAt, // 最后登录时间
		LastLoginIp: rpcResp.LastLoginIp, // 最后登录IP

		// ==================== 运营信息 ====================

		InviterMemberId: rpcResp.InviterMemberId, // 邀请人会员ID
		Remark:          rpcResp.Remark,          // 后台会员备注

		// ==================== 统计信息 ====================

		Stats: listMemberStats(rpcResp.GetStats()), // 会员累计及当前周期统计
	}, nil
}
