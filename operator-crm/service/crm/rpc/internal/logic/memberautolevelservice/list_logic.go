package memberautolevelservicelogic

import (
	"context"
	"math"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberautolevelpb"
)

const (
	defaultPageSize int64 = 20  // 默认每页数量
	maxPageSize     int64 = 100 // 最大每页数量
)

type ListLogic struct {
	ctx         context.Context     // 请求上下文
	svcCtx      *svc.ServiceContext // 服务依赖
	logx.Logger                     // 日志
}

func NewListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLogic {
	return &ListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// List 获取自动层级列表
func (l *ListLogic) List(in *memberautolevelpb.ListMemberAutoLevelsRequest) (*memberautolevelpb.ListMemberAutoLevelsResponse, error) {
	if in == nil {
		in = &memberautolevelpb.ListMemberAutoLevelsRequest{}
	}

	// 页码
	page := max(in.GetPage(), int64(1))

	// 每页数量
	pageSize := in.GetPageSize()
	switch {
	case pageSize <= 0:
		pageSize = defaultPageSize
	case pageSize > maxPageSize:
		pageSize = maxPageSize
	}

	// 计算分页偏移量
	if page-1 > int64(math.MaxInt)/pageSize {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}
	offset := int((page - 1) * pageSize)

	// 查询自动层级
	levels, total, err := l.svcCtx.DAO.MemberAutoLevel.List(l.ctx, in.GetName(), offset, int(pageSize))
	if err != nil {
		return nil, xerr.RpcErr(err)
	}

	// 批量统计当前页的会员数
	levelIDs := make([]int64, 0, len(levels))
	for _, level := range levels {
		levelIDs = append(levelIDs, level.ID)
	}

	// 批量统计各层级会员数
	memberCounts, err := l.svcCtx.DAO.MemberAutoLevel.CountMembersByLevelIDs(l.ctx, levelIDs)
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
