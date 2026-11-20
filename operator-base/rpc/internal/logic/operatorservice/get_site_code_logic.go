package operatorservicelogic

import (
	"context"
	"regexp"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"oa.98ent.com/p9/operator-base/rpc/ent"
	"oa.98ent.com/p9/operator-base/rpc/ent/operator"
	"oa.98ent.com/p9/operator-base/rpc/ent/operatordomain"
	"oa.98ent.com/p9/operator-base/rpc/internal/svc"
	"oa.98ent.com/p9/operator-base/rpc/pb/operatorbaserpc/operatorpb"
)

// 与域名表的字段格式一致，仅接受纯域名。
var siteCodeDomainPattern = regexp.MustCompile(
	`^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z](?:[a-z0-9-]{0,61}[a-z0-9])?$`,
)

type GetSiteCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSiteCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSiteCodeLogic {
	return &GetSiteCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetSiteCode 供登录前公开识别站点，仅查询后台域名及所属站点编码
func (l *GetSiteCodeLogic) GetSiteCode(in *operatorpb.GetSiteCodeRequest) (
	*operatorpb.GetSiteCodeResponse, error,
) {
	// 复用域名规范化规则；仅接受纯域名，不接受协议、端口、路径或通配符。
	domain := normalizeBackendDomain(in.GetDomainName())
	if len(domain) > 253 || !siteCodeDomainPattern.MatchString(domain) {
		return nil, status.Error(codes.InvalidArgument, "domain_name must be a valid domain without scheme, port or path")
	}
	// 请求已取消或超时时直接返回，避免继续访问数据库。
	if err := l.ctx.Err(); err != nil {
		return nil, status.FromContextError(err).
			Err()
	}
	if l.svcCtx.DB == nil {
		return nil, status.Error(codes.FailedPrecondition, "operator database is not configured")
	}
	// 复用现有表关联，一次查询只读取 code；不加载语言、地区或设置。
	owner, err := l.svcCtx.DB.Operator.Query().
		Where(operator.HasDomainsWith(
			operatordomain.DomainNameEqualFold(domain),
			operatordomain.DomainTypeEQ(domainTypeBackend),
		)).
		Select(operator.FieldCode).
		Only(l.ctx)
	// 没有匹配的后台域名属于资源不存在，与数据库故障区分返回。
	if ent.IsNotFound(err) {
		return nil, status.Error(codes.NotFound, "operator domain not found")
	}
	if err != nil {
		// 内部数据库原因写入日志，对外保留取消、超时或通用错误，不暴露 SQL。
		l.Logger.Errorw("按域名查询站点编码失败", logx.Field("error", err.Error()))
		if l.ctx.Err() != nil {
			return nil, status.FromContextError(l.ctx.Err()).
				Err()
		}
		return nil, status.Error(codes.Internal, "query site code failed")
	}
	// 已关联站点却没有业务编码属于数据异常，不能作为合法登录站点返回。
	if owner.Code == "" {
		l.Logger.Errorw("后台域名关联的站点编码为空", logx.Field("domain_name", domain))
		return nil, status.Error(codes.Internal, "query site code failed")
	}
	// 公开接口只返回站点编码，不读取或初始化网站设置，也不写入数据库。
	return &operatorpb.GetSiteCodeResponse{OperatorCode: owner.Code}, nil
}
