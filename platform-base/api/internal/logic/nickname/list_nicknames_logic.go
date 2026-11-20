// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package nickname

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/platform-base/api/internal/svc"
	"oa.98ent.com/p9/platform-base/api/internal/types"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/nicknamepb"
)

type ListNicknamesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListNicknamesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListNicknamesLogic {
	return &ListNicknamesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListNicknamesLogic) ListNicknames(req *types.ListNicknamesRequest) (resp *types.ListNicknamesResponse, err error) {
	r, err := l.svcCtx.Nicknames.List(l.ctx, &nicknamepb.ListNicknamesRequest{Page: req.Page, PageSize: req.PageSize, Keyword: req.Keyword})
	if err != nil {
		return nil, err
	}
	list := make([]types.NicknameInfo, 0, len(r.List))
	for _, row := range r.List {
		list = append(list, types.NicknameInfo{Id: row.Id, Name: row.Name})
	}
	return &types.ListNicknamesResponse{Total: r.Total, List: list}, nil
}
