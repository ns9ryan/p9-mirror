package operatorservicelogic

import (
	"context"

	"oa.98ent.com/p9/operator-base/rpc/internal/svc"
	"oa.98ent.com/p9/operator-base/rpc/pb/operatorbaserpc/operatorpb"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	domainTypeBackend int64 = 1
	domainTypeAgent   int64 = 2
	domainTypeMember  int64 = 3
)

// GetSiteConfigLogic 按登录站点查询配置。
type GetSiteConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewGetSiteConfigLogic 创建登录站点配置查询逻辑。
func NewGetSiteConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSiteConfigLogic {
	return &GetSiteConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetSiteConfig 查询登录站点的配置编码，保留旧请求类型以兼容已生成的调用方。
func (l *GetSiteConfigLogic) GetSiteConfig(in *operatorpb.GetSiteConfigRequest) (
	*operatorpb.GetSiteConfigResponse, error,
) {
	// 复用认证上下文确定站点，旧请求中的域名和类型不参与归属判断。
	config, _, err := lookupSiteConfigWithOwner(l.ctx, l.svcCtx.DB, l.Logger)
	if err != nil {
		return nil, err
	}
	return &operatorpb.GetSiteConfigResponse{SiteConfig: config}, nil
}
