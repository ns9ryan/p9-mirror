package memberautolevelservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/common/paging"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberautolevelpb"
)

// ListLogic 自动层级列表逻辑
type ListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewListLogic 创建自动层级列表逻辑
func NewListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLogic {
	return &ListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// List 获取自动层级列表
func (l *ListLogic) List(in *memberautolevelpb.ListMemberAutoLevelsRequest) (*memberautolevelpb.ListMemberAutoLevelsResponse, error) {
	// 分页参数
	offset, limit, ok := paging.OffsetLimit(in.GetPage(), in.GetPageSize())
	if !ok {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	// 获取自动层级分页列表
	levels, total, err := l.svcCtx.DAO.MemberAutoLevel.List(l.ctx, in.GetName(), offset, limit)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}

	// 获取当前页层级ID
	levelIDs := make([]int64, 0, len(levels))
	for _, level := range levels {
		levelIDs = append(levelIDs, level.ID)
	}

	// 统计指定自动层级的会员数
	memberCounts, err := l.svcCtx.DAO.MemberAutoLevel.MemberCounts(l.ctx, levelIDs)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}

	// 组装层级列表
	list := make([]*memberautolevelpb.MemberAutoLevelInfo, 0, len(levels))
	for _, level := range levels {
		list = append(list, &memberautolevelpb.MemberAutoLevelInfo{
			Id:                 level.ID,                    // 层级ID
			Name:               level.Name,                  // 层级名称
			MinDeposit:         level.MinDeposit,            // 最低累计存款
			MinWater:           level.MinWater,              // 最低累计洗码量
			CurrentMemberCount: memberCounts[level.ID],      // 当前会员数
			CreatedAt:          level.CreatedAt.UnixMilli(), // 创建时间
			UpdatedAt:          level.UpdatedAt.UnixMilli(), // 更新时间
		})
	}

	return &memberautolevelpb.ListMemberAutoLevelsResponse{
		Total: total, // 数据总数
		List:  list,  // 层级列表
	}, nil
}
