package timezoneservicelogic

import (
	"context"

	"oa.98ent.com/p9/platform-base/pkg/i18nkey"
	"oa.98ent.com/p9/platform-base/pkg/rpc/grpcerror"
	"oa.98ent.com/p9/platform-base/rpc/ent"
	"oa.98ent.com/p9/platform-base/rpc/internal/cache/basedata"
	"oa.98ent.com/p9/platform-base/rpc/internal/enterror"
	"oa.98ent.com/p9/platform-base/rpc/internal/svc"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/timezonepb"

	"github.com/zeromicro/go-zero/core/logx"
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

// List 获取时区管理列表
func (l *ListLogic) List(in *timezonepb.ListTimezonesRequest) (*timezonepb.ListTimezonesResponse, error) {
	// 校验分页参数
	if in.Page < 1 || in.PageSize < 1 || in.PageSize > 100 {
		return nil, grpcerror.InvalidArgument(i18nkey.ValidationError)
	}

	// 校验状态
	if in.Status != nil && (*in.Status < 1 || *in.Status > 2) {
		return nil, grpcerror.InvalidArgument(i18nkey.ValidationError)
	}

	results, err := l.svcCtx.BaseData.Timezone.All(l.ctx)
	if err != nil {
		return nil, enterror.Handle(l.Logger, err)
	}

	var matches func(*ent.Timezone) bool
	if in.Status != nil {
		status := *in.Status
		matches = func(row *ent.Timezone) bool { return row.Status == status }
	}
	results, total := basedata.Page(results, in.Page, in.PageSize, matches)

	// 转换时区列表
	list := make([]*timezonepb.TimezoneInfo, 0, len(results))
	for _, result := range results {
		list = append(list, toTimezoneInfo(result))
	}

	return &timezonepb.ListTimezonesResponse{
		Total: total, // 数据总数
		List:  list,  // 时区列表
	}, nil
}
