package regionservicelogic

import (
	"context"

	"oa.98ent.com/p9/platform-base/pkg/i18nkey"
	"oa.98ent.com/p9/platform-base/pkg/rpc/grpcerror"
	"oa.98ent.com/p9/platform-base/rpc/internal/enterror"
	"oa.98ent.com/p9/platform-base/rpc/internal/svc"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/regionpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAllLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAllLogic {
	return &ListAllLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListAll 获取全部国家地区
func (l *ListAllLogic) ListAll(in *regionpb.ListAllRegionsRequest) (*regionpb.ListAllRegionsResponse, error) {
	// 校验状态
	if in.Status != nil && (*in.Status < 1 || *in.Status > 2) {
		return nil, grpcerror.InvalidArgument(i18nkey.ValidationError)
	}

	results, err := l.svcCtx.BaseData.Region.All(l.ctx)
	if err != nil {
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// 转换国家地区列表
	list := make([]*regionpb.RegionInfo, 0, len(results))
	for _, result := range results {
		if in.Status != nil && result.Status != *in.Status {
			continue
		}
		list = append(list, toRegionInfo(result))
	}

	return &regionpb.ListAllRegionsResponse{
		List: list, // 国家地区列表
	}, nil
}
