package operatorservicelogic

import (
	"context"

	"oa.98ent.com/p9/operator-base/rpc/internal/svc"
	"oa.98ent.com/p9/operator-base/rpc/pb/operatorbaserpc/operatorpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAllOperatorTimezoneInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAllOperatorTimezoneInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAllOperatorTimezoneInfoLogic {
	return &GetAllOperatorTimezoneInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取所有分站的编码和时区信息
func (l *GetAllOperatorTimezoneInfoLogic) GetAllOperatorTimezoneInfo(in *operatorpb.GetAllOperatorTimezoneInfoRequest) (*operatorpb.GetAllOperatorTimezoneInfoResponse, error) {
	// 查询所有启用状态的分站
	operators, err := l.svcCtx.DB.Operator.Query().
		Where().
		All(l.ctx)

	if err != nil {
		l.Errorf("获取分站时区信息失败，err=%v", err)
		return &operatorpb.GetAllOperatorTimezoneInfoResponse{
			Operators: []*operatorpb.OperatorTimezoneInfoItem{},
		}, nil
	}

	// 构建响应
	items := make([]*operatorpb.OperatorTimezoneInfoItem, 0, len(operators))
	for _, op := range operators {
		items = append(items, &operatorpb.OperatorTimezoneInfoItem{
			Code:         op.Code,
			TimezoneCode: op.TimezoneCode,
		})
	}

	l.Infof("成功获取分站时区信息，总数=%d", len(items))

	return &operatorpb.GetAllOperatorTimezoneInfoResponse{
		Operators: items,
	}, nil
}
