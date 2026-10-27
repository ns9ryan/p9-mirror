package memberfixedlevellogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/common/paging"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberfixedlevelpb"
)

type ListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLogic {
	return &ListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// List 获取固定层级列表
func (l *ListLogic) List(in *memberfixedlevelpb.ListMemberFixedLevelsRequest) (*memberfixedlevelpb.ListMemberFixedLevelsResponse, error) {
	// 分页参数
	offset, limit, ok := paging.OffsetLimit(in.GetPage(), in.GetPageSize())
	if !ok {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	// 获取固定层级分页列表
	levels, total, err := l.svcCtx.DAO.MemberFixedLevel.List(l.ctx, in.GetName(), offset, limit)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}

	// 获取当前页层级ID
	levelIDs := make([]int64, 0, len(levels))
	for _, level := range levels {
		levelIDs = append(levelIDs, level.ID)
	}

	// 统计指定固定层级的会员数
	memberCounts, err := l.svcCtx.DAO.MemberFixedLevel.MemberCounts(l.ctx, levelIDs)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}

	// 组装层级列表
	list := make([]*memberfixedlevelpb.MemberFixedLevelInfo, 0, len(levels))
	for _, level := range levels {
		list = append(list, &memberfixedlevelpb.MemberFixedLevelInfo{
			Id:                 level.ID,                    // 层级ID
			Name:               level.Name,                  // 层级名称
			Description:        level.Description,           // 层级说明
			WithdrawalLimit:    level.WithdrawalLimit,       // 单笔提款金额上限
			CurrentMemberCount: memberCounts[level.ID],      // 当前会员数
			CreatedAt:          level.CreatedAt.UnixMilli(), // 创建时间
			UpdatedAt:          level.UpdatedAt.UnixMilli(), // 更新时间
		})
	}

	return &memberfixedlevelpb.ListMemberFixedLevelsResponse{
		Total: total, // 数据总数
		List:  list,  // 层级列表
	}, nil
}
