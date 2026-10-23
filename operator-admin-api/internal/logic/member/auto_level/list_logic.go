// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package auto_level

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberautolevelpb"
)

// ListLogic 自动层级列表逻辑
type ListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListLogic 创建自动层级列表逻辑
func NewListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLogic {
	return &ListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// List 获取自动层级列表
func (l *ListLogic) List(req *types.ListMemberAutoLevelsRequest) (resp *types.ListMemberAutoLevelsResponse, err error) {
	// 查询自动层级
	rpcResp, err := l.svcCtx.MemberAutoLevelRpc.List(l.ctx, &memberautolevelpb.ListMemberAutoLevelsRequest{
		Page:     req.Page,     // 当前页码, 从1开始
		PageSize: req.PageSize, // 分页数量
		Name:     &req.Name,    // 层级名称, 模糊查询
	})
	if err != nil {
		return nil, err
	}

	// 组装自动层级列表
	list := make([]types.MemberAutoLevelInfo, 0, len(rpcResp.GetList()))
	for _, level := range rpcResp.GetList() {
		list = append(list, types.MemberAutoLevelInfo{
			Id:                 level.Id,                 // 层级ID
			Name:               level.Name,               // 层级名称
			MinDeposit:         level.MinDeposit,         // 最低累计存款
			MinWater:           level.MinWater,           // 最低累计洗码量
			CurrentMemberCount: level.CurrentMemberCount, // 当前会员数
			CreatedAt:          level.CreatedAt,          // 创建时间
			UpdatedAt:          level.UpdatedAt,          // 更新时间
		})
	}

	return &types.ListMemberAutoLevelsResponse{
		Total: rpcResp.GetTotal(), // 总条数
		List:  list,               // 自动层级列表
	}, nil
}
