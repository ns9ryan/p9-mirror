package operatordomainservicelogic

import (
	"context"
	"strings"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/platform-operator/pkg/i18nkey"
	"oa.98ent.com/p9/platform-operator/rpc/ent"
	"oa.98ent.com/p9/platform-operator/rpc/ent/operatordomain"
	"oa.98ent.com/p9/platform-operator/rpc/internal/enterror"
	"oa.98ent.com/p9/platform-operator/rpc/internal/svc"
	"oa.98ent.com/p9/platform-operator/rpc/pb/platformoperatorrpc/domainpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLogic {
	return &CreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Create 创建分站域名
func (l *CreateLogic) Create(in *domainpb.CreateDomainRequest) (*domainpb.CreateDomainResponse, error) {
	// 分站ID必须大于0
	if in.OperatorId <= 0 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
	}

	// 确认分站存在
	_, err := l.svcCtx.DB.Operator.Get(l.ctx, in.OperatorId)
	if err != nil {
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// 标准化域名
	domainName := strings.ToLower(strings.TrimSpace(in.DomainName))

	// 校验域名是否重复
	existingDomain, err := l.svcCtx.DB.OperatorDomain.
		Query().
		Where(
			operatordomain.DomainNameEQ(domainName),
		).
		Only(l.ctx)
	if err == nil && existingDomain != nil {
		l.Errorf("域名已存在: domain_name=%s", domainName)
		return nil, xerr.RpcErr(xerr.BadRequest("The domain name already exists"))
	}
	if err != nil && !ent.IsNotFound(err) {
		l.Errorf("查询域名是否存在失败: domain_name=%s, error=%v", domainName, err)
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// 创建分站域名
	data, err := l.svcCtx.DB.OperatorDomain.
		Create().
		SetOperatorID(in.OperatorId).                     // 分站ID
		SetDomainName(domainName).                        // 域名, 不包含协议和端口
		SetDomainType(in.DomainType).                     // 域名类型: 1分站后台, 2代理后台, 3会员H5
		SetNillableStatus(in.Status).                     // 域名状态: 1启用, 2停用
		SetNillableRemark(trimOptionalString(in.Remark)). // 总网内部备注
		Save(l.ctx)
	if err != nil {
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// 返回创建结果
	return &domainpb.CreateDomainResponse{
		Id: data.ID, // 域名ID
	}, nil
}
