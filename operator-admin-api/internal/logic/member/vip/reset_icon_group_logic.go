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

// ResetIconGroupLogic VIP图标恢复默认逻辑
type ResetIconGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewResetIconGroupLogic 绑定恢复请求的上下文和日志，此时不调用RPC或保存配置。
func NewResetIconGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetIconGroupLogic {
	return &ResetIconGroupLogic{
		Logger: logx.WithContext(ctx), // 日志关联当前恢复请求。
		ctx:    ctx,                   // 透传登录身份和请求期限。
		svcCtx: svcCtx,                // 复用CRM VIP客户端。
	}
}

// ResetIconGroup 恢复默认VIP图标
func (l *ResetIconGroupLogic) ResetIconGroup(req *types.ResetVipIconGroupRequest) (resp *types.EmptyResponse, err error) {
	// 调用CRM恢复接口，将本站指定组图片清空；后续查询随总站实时默认图展示。
	_, err = l.svcCtx.MemberVipRpc.ResetIconGroup(l.ctx, &vipicongrouppb.ResetVipIconGroupRequest{
		GroupNo: *req.GroupNo, // 请求恢复的组编号，0为第一组。
	})
	if err != nil {
		return nil, iconRequestError(err)
	}

	// CRM保存成功后返回空结果；前端应重新查询图标列表刷新当前图片。
	return &types.EmptyResponse{}, nil
}
