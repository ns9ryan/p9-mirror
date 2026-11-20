package activityservicelogic

import (
	"context"
	"errors"

	"oa.98ent.com/p9/operator-activity/ent"
	"oa.98ent.com/p9/operator-activity/internal/svc"
	"oa.98ent.com/p9/operator-activity/modules/activity/repository/admin"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/activitypb"
	"oa.98ent.com/p9/operator-activity/pkg/i18n"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
)

type PublishLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublishLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishLogic {
	return &PublishLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Publish 发布活动（仅草稿）
func (l *PublishLogic) Publish(in *activitypb.PublishActivityReq) (*activitypb.PublishActivityResp, error) {
	claims, ok := ctxdata.OperatorClaimsFromCtx(l.ctx)
	if !ok {
		return nil, xerr.RpcErr(xerr.Unauthorized(i18n.Unauthorized))
	}
	if in.GetId() <= 0 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidID))
	}

	adminServ := admin.NewService(l.svcCtx.DB)
	if err := adminServ.PublishActivity(l.ctx, in.GetId(), claims.OperatorCode, claims.UserID); err != nil {
		switch {
		case errors.Is(err, admin.ErrActivityNotDraft):
			return nil, xerr.RpcErr(xerr.BadRequest(i18n.NotDraft))
		case ent.IsNotFound(err):
			return nil, xerr.RpcErr(xerr.EntNotFound(i18n.NotFound, err))
		case ent.IsConstraintError(err):
			return nil, xerr.RpcErr(xerr.EntConstraintError(i18n.InvalidParam, err))
		case ent.IsValidationError(err):
			return nil, xerr.RpcErr(xerr.EntValidationError(i18n.InvalidParam, err))
		default:
			return nil, xerr.RpcErr(err)
		}
	}

	return &activitypb.PublishActivityResp{Id: in.GetId()}, nil
}
