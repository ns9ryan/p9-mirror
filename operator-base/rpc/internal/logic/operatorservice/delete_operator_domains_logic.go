package operatorservicelogic

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/operator-base/rpc/ent/operator"
	"oa.98ent.com/p9/operator-base/rpc/ent/operatordomain"
	"oa.98ent.com/p9/operator-base/rpc/internal/svc"
	"oa.98ent.com/p9/operator-base/rpc/pb/operatorbaserpc/operatorpb"
)

type DeleteOperatorDomainsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteOperatorDomainsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteOperatorDomainsLogic {
	return &DeleteOperatorDomainsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除分站域名
func (l *DeleteOperatorDomainsLogic) DeleteOperatorDomains(in *operatorpb.DeleteOperatorDomainsRequest) (*operatorpb.DeleteOperatorDomainsResponse, error) {
	// ==================== 参数校验 ====================

	operatorCode := strings.TrimSpace(in.OperatorCode)
	if operatorCode == "" {
		return &operatorpb.DeleteOperatorDomainsResponse{
			Success: false,
			Message: "operator_code is required",
		}, nil
	}

	l.Logger.Infof("开始删除分站域名: operator_code=%s", operatorCode)

	// ==================== 查询分站 ====================

	operatorData, err := l.svcCtx.DB.Operator.Query().
		Where(operator.CodeEQ(operatorCode)).
		First(l.ctx)
	if err != nil {
		l.Logger.Errorf("查询分站失败: operator_code=%s, error=%v", operatorCode, err)
		return &operatorpb.DeleteOperatorDomainsResponse{
			Success: false,
			Message: fmt.Sprintf("query operator failed: %v", err),
		}, nil
	}

	// ==================== 删除分站域名 ====================

	deleteCount, err := l.svcCtx.DB.OperatorDomain.Delete().
		Where(operatordomain.OperatorIDEQ(operatorData.ID)).
		Where(operatordomain.DomainNameIn(in.DomainNames...)).
		Exec(l.ctx)
	if err != nil {
		l.Logger.Errorf("删除分站域名失败: operator_code=%s, error=%v", operatorCode, err)
		return &operatorpb.DeleteOperatorDomainsResponse{
			Success: false,
			Message: fmt.Sprintf("delete domains failed: %v", err),
		}, nil
	}

	domainDeleted := int64(deleteCount)
	if deleteCount > 0 {
		l.Logger.Infof("成功删除分站域名: operator_code=%s, domain_count=%d", operatorCode, deleteCount)
	}

	return &operatorpb.DeleteOperatorDomainsResponse{
		Success:       true,
		Message:       "delete operator domains successfully",
		DomainDeleted: domainDeleted,
	}, nil
}
