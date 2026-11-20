// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package public

import (
	"context"

	"oa.98ent.com/p9/agent-admin-api/internal/svc"
	"oa.98ent.com/p9/agent-admin-api/internal/types"
	"oa.98ent.com/p9/common/i18n"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetI18nDictLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 词条下发
func NewGetI18nDictLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetI18nDictLogic {
	return &GetI18nDictLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetI18nDictLogic) GetI18nDict(req *types.GetI18nDictReq) (resp *types.I18nDictResp, err error) {
	items := i18n.Dict(l.ctx, req.I18nCode, req.I18nGroup, req.Lang)
	if items == nil {
		items = map[string]string{}
	}

	return &types.I18nDictResp{Items: items}, nil
}
