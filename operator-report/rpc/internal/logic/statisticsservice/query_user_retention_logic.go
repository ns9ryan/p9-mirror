package statisticsservicelogic

import (
	"context"

	"oa.98ent.com/p9/operator-report/rpc/internal/svc"
	"oa.98ent.com/p9/operator-report/rpc/pb/operator_report"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryUserRetentionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryUserRetentionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryUserRetentionLogic {
	return &QueryUserRetentionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询用户留存数据
func (l *QueryUserRetentionLogic) QueryUserRetention(in *operator_report.UserRetentionQueryRequest) (*operator_report.UserRetentionQueryResponse, error) {
	// todo: add your logic here and delete this line

	return &operator_report.UserRetentionQueryResponse{}, nil
}
