// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package operator_domain

import (
	"context"

	"oa.98ent.com/p9/platform-operator/api/internal/svc"
	"oa.98ent.com/p9/platform-operator/api/internal/types"
	domainpb "oa.98ent.com/p9/platform-operator/rpc/pb/platformoperatorrpc/domainpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishOperatorDomainLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPublishOperatorDomainLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishOperatorDomainLogic {
	return &PublishOperatorDomainLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PublishOperatorDomainLogic) PublishOperatorDomain(req *types.PublishOperatorDomainRequest) (resp *types.PublishOperatorDomainResponse, err error) {
	// 验证分站编码
	id := req.Id
	if id <= 0 {
		l.Errorf("分站ID必须大于0")
		return nil, err
	}

	l.Infof("开始发布分站域名, id=%d", id)

	// 调用 RPC 服务发布域名
	_, err = l.svcCtx.OperatorDomainRpc.Publish(l.ctx, &domainpb.PublishDomainRequest{
		Id: id,
	})
	if err != nil {
		l.Errorf("调用RPC发布分站域名失败: id=%d, error=%v", id, err)
		return nil, err
	}

	l.Infof("分站域名发布成功, id=%d", id)
	return &types.PublishOperatorDomainResponse{}, nil
}
