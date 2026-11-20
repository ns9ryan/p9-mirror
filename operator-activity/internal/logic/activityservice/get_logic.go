package activityservicelogic

import (
	"context"

	"oa.98ent.com/p9/operator-activity/ent"
	"oa.98ent.com/p9/operator-activity/internal/svc"
	"oa.98ent.com/p9/operator-activity/modules/activity/repository/admin"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/activitypb"
	"oa.98ent.com/p9/operator-activity/pkg/i18n"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
)

type GetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLogic {
	return &GetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Get 活动详情
func (l *GetLogic) Get(in *activitypb.GetActivityReq) (*activitypb.GetActivityResp, error) {
	if _, ok := ctxdata.OperatorClaimsFromCtx(l.ctx); !ok {
		return nil, xerr.RpcErr(xerr.Unauthorized(i18n.Unauthorized))
	}
	if in.GetId() <= 0 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidID))
	}

	adminServ := admin.NewService(l.svcCtx.DB)
	model, i18nList, err := adminServ.GetActivityDetail(l.ctx, in.GetId())
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, xerr.RpcErr(xerr.EntNotFound(i18n.NotFound, err))
		}
		return nil, xerr.RpcErr(err)
	}

	detail, err := toActivityDetail(model, i18nList)
	if err != nil {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}
	return &activitypb.GetActivityResp{Activity: detail}, nil
}
