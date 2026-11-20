// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package fixed_level

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberfixedlevelpb"
)

// ListLogic 会员固定层级列表逻辑
type ListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListLogic 创建固定层级列表逻辑
func NewListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLogic {
	return &ListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// List 获取固定层级列表
func (l *ListLogic) List(req *types.ListMemberFixedLevelsRequest) (resp *types.ListMemberFixedLevelsResponse, err error) {
	rpcResp, err := l.svcCtx.MemberFixedLevelRpc.List(l.ctx, &memberfixedlevelpb.ListMemberFixedLevelsRequest{
		Page:     req.Page,     // 当前页码
		PageSize: req.PageSize, // 每页数量
		Name:     &req.Name,    // 层级名称，模糊查询
	})
	if err != nil {
		return nil, err
	}

	// 组装固定层级列表
	list := make([]types.MemberFixedLevelInfo, 0, len(rpcResp.GetList()))
	for _, level := range rpcResp.GetList() {
		list = append(list, types.MemberFixedLevelInfo{
			Id:                 level.Id,                 // 层级ID
			Name:               level.Name,               // 层级名称
			Description:        level.Description,        // 层级说明
			WithdrawalLimit:    level.WithdrawalLimit,    // 单笔提款金额上限
			CurrentMemberCount: level.CurrentMemberCount, // 当前会员数
			CreatedAt:          level.CreatedAt,          // 创建时间
			UpdatedAt:          level.UpdatedAt,          // 更新时间
		})
	}

	return &types.ListMemberFixedLevelsResponse{
		Total: rpcResp.GetTotal(), // 数据总数
		List:  list,               // 固定层级列表
	}, nil
}
