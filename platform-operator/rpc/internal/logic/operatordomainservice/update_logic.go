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

type UpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLogic {
	return &UpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Update 修改分站域名
func (l *UpdateLogic) Update(in *domainpb.UpdateDomainRequest) (*domainpb.UpdateDomainResponse, error) {
	// 域名ID必须大于0
	if in.Id <= 0 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
	}

	// 至少需要修改一个字段
	if in.DomainName == nil &&
		in.DomainType == nil &&
		in.Status == nil &&
		in.Remark == nil {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
	}

	// 获取当前分站域名
	current, err := l.svcCtx.DB.OperatorDomain.Get(l.ctx, in.Id)
	if err != nil {
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// 标准化域名
	domainName := strings.ToLower(strings.TrimSpace(in.GetDomainName()))

	// 校验域名是否重复
	existingDomain, err := l.svcCtx.DB.OperatorDomain.
		Query().
		Where(
			operatordomain.DomainNameEQ(domainName),
		).
		Only(l.ctx)

	// 等于原域名时不报错（检查ID是否相同）
	if err == nil && existingDomain != nil && existingDomain.ID != current.ID {
		l.Errorf("域名已存在: domain_name=%s", domainName)
		return nil, xerr.RpcErr(xerr.BadRequest("The domain name already exists"))
	}
	if err != nil && !ent.IsNotFound(err) {
		l.Errorf("查询域名是否存在失败: domain_name=%s, error=%v", domainName, err)
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// 修改分站域名
	err = current.
		Update().
		SetNillableDomainName(&domainName).               // 域名, 不包含协议和端口
		SetNillableDomainType(in.DomainType).             // 域名类型: 1分站后台, 2代理后台, 3会员H5
		SetNillableStatus(in.Status).                     // 域名状态: 1启用, 2停用
		SetNillableRemark(trimOptionalString(in.Remark)). // 总网内部备注
		Exec(l.ctx)
	if err != nil {
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// 返回修改结果
	return &domainpb.UpdateDomainResponse{}, nil
}
