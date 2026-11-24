// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package manage

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"
)

// SearchByCodeLogic 按会员编号/账号/昵称模糊搜索逻辑
type SearchByCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewSearchByCodeLogic 创建搜索逻辑
func NewSearchByCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchByCodeLogic {
	return &SearchByCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// SearchByCode 按关键字模糊搜索会员，返回 id、编号、账号、昵称
func (l *SearchByCodeLogic) SearchByCode(req *types.SearchMembersByCodeRequest) (resp *types.SearchMembersByCodeResponse, err error) {
	rpcResp, err := l.svcCtx.MemberRpc.SearchByCode(l.ctx, &memberpb.SearchMembersByCodeReq{
		Keyword: req.Keyword,
		Limit:   req.Limit,
	})
	if err != nil {
		return nil, err
	}

	items := make([]types.SearchMembersByCodeItem, 0, len(rpcResp.GetList()))
	for _, m := range rpcResp.GetList() {
		items = append(items, types.SearchMembersByCodeItem{
			Id:       m.Id,
			Code:     m.Code,
			Account:  m.Account,
			Nickname: m.Nickname,
		})
	}

	return &types.SearchMembersByCodeResponse{List: items}, nil
}
