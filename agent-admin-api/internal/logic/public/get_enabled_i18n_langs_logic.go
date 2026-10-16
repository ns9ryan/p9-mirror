// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package public

import (
	"context"

	"oa.98ent.com/p9/agent-admin-api/internal/svc"
	"oa.98ent.com/p9/agent-admin-api/internal/types"
	"oa.98ent.com/p9/core/rpc/coreclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetEnabledI18nLangsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 已开启的语言列表
func NewGetEnabledI18nLangsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetEnabledI18nLangsLogic {
	return &GetEnabledI18nLangsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetEnabledI18nLangsLogic) GetEnabledI18nLangs(req *types.GetEnabledI18nLangsReq) (resp *types.I18nLangListResp, err error) {
	out, err := l.svcCtx.OperatorCoreRpc.GetEnabledI18NLangs(
		l.ctx,
		&coreclient.GetEnabledI18NLangsReq{
			OperatorCode: req.OperatorCode,
		},
	)
	if err != nil {
		return nil, err
	}
	list := make([]types.I18nLangInfo, 0, len(out.List))
	for _, item := range out.List {
		list = append(list, types.I18nLangInfo{
			Id:   item.Id,
			Lang: item.Lang,
			Name: item.Name,
		})
	}
	return &types.I18nLangListResp{List: list}, nil
}
