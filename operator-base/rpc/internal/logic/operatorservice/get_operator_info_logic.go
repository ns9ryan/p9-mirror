package operatorservicelogic

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-base/rpc/ent"
	"oa.98ent.com/p9/operator-base/rpc/ent/operator"
	"oa.98ent.com/p9/operator-base/rpc/internal/svc"
	"oa.98ent.com/p9/operator-base/rpc/pb/operatorbaserpc/operatorpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOperatorInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOperatorInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOperatorInfoLogic {
	return &GetOperatorInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取分站信息
func (l *GetOperatorInfoLogic) GetOperatorInfo(in *operatorpb.GetOperatorInfoRequest) (*operatorpb.GetOperatorInfoResponse, error) {
	operatorCode := ctxdata.OperatorCodeFromCtx(l.ctx)
	if operatorCode == "" {
		return nil, status.Error(codes.Unauthenticated, "缺少分站上下文")
	}

	operatorData, err := l.svcCtx.DB.Operator.Query().
		Select(operator.FieldName, operator.FieldStatus, operator.FieldTimezoneCode, operator.FieldSettlementCurrencyCode).
		Where(operator.CodeEQ(operatorCode)).
		First(l.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "operator not found")
		}

		l.Logger.Errorw(
			"查询分站失败",
			logx.Field("operator_code", operatorCode),
			logx.Field("error", err.Error()),
		)
		return nil, err
	}

	return &operatorpb.GetOperatorInfoResponse{
		Name:                   operatorData.Name,
		Status:                 operatorData.Status,
		TimezoneCode:           operatorData.TimezoneCode,
		SettlementCurrencyCode: operatorData.SettlementCurrencyCode,
	}, nil
}
