package activityservicelogic

import (
	"context"
	"strings"

	"oa.98ent.com/p9/operator-activity/internal/svc"
	"oa.98ent.com/p9/operator-activity/modules/activity/repository/admin"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/activitypb"
	"oa.98ent.com/p9/operator-activity/pkg/i18n"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
)

type ListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLogic {
	return &ListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// List 活动列表
func (l *ListLogic) List(in *activitypb.ListActivitiesReq) (*activitypb.ListActivitiesResp, error) {
	if _, ok := ctxdata.OperatorClaimsFromCtx(l.ctx); !ok {
		return nil, xerr.RpcErr(xerr.Unauthorized(i18n.Unauthorized))
	}

	params := admin.ListActivityParams{
		Page:     in.GetPage(),
		PageSize: in.GetPageSize(),
		Name:     strings.TrimSpace(in.GetName()),
		Lang:     strings.TrimSpace(in.GetLang()),
	}
	if in.Type != nil {
		v := in.GetType()
		params.Type = &v
	}
	if in.Client != nil {
		v := in.GetClient()
		params.Client = &v
	}
	if in.Status != nil {
		v := in.GetStatus()
		params.Status = &v
	}

	adminServ := admin.NewService(l.svcCtx.DB)
	rows, total, err := adminServ.ListActivities(l.ctx, params)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}

	list := make([]*activitypb.ActivityListItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, toActivityListItem(row))
	}
	return &activitypb.ListActivitiesResp{
		Total: total,
		List:  list,
	}, nil
}
