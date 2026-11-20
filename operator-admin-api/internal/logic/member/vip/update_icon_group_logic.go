// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package vip

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/vipicongrouppb"
)

// UpdateIconGroupLogic VIP图标修改逻辑
type UpdateIconGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdateIconGroupLogic 绑定保存请求的上下文和日志，此时不调用RPC或写库。
func NewUpdateIconGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateIconGroupLogic {
	return &UpdateIconGroupLogic{
		Logger: logx.WithContext(ctx), // 日志关联当前保存请求。
		ctx:    ctx,                   // 透传登录身份和请求期限。
		svcCtx: svcCtx,                // 复用CRM VIP客户端。
	}
}

// UpdateIconGroup 修改VIP图标
func (l *UpdateIconGroupLogic) UpdateIconGroup(req *types.UpdateVipIconGroupRequest) (resp *types.EmptyResponse, err error) {
	// 调用CRM校验并保存当前图标；上传结果中的Key仍使用原icon_url字段，不在API再次上传。
	_, err = l.svcCtx.MemberVipRpc.UpdateIconGroup(l.ctx, &vipicongrouppb.UpdateVipIconGroupRequest{
		GroupNo: *req.GroupNo, // 图标组编号
		IconUrl: req.IconUrl,  // 原字段接收Key或完整URL，由CRM校验并保存。
	})
	if err != nil {
		return nil, iconRequestError(err)
	}

	return &types.EmptyResponse{}, nil
}
