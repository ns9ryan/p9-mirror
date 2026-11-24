package memberlogic

import (
	"context"
	"strings"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchByCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchByCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchByCodeLogic {
	return &SearchByCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 按会员编号/账号/昵称模糊搜索会员（返回 id、编号、账号、昵称）
func (l *SearchByCodeLogic) SearchByCode(in *memberpb.SearchMembersByCodeReq) (*memberpb.SearchMembersByCodeResp, error) {
	if in == nil || strings.TrimSpace(in.Keyword) == "" {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}

	rows, err := l.svcCtx.Member.SearchByCode(l.ctx, strings.TrimSpace(in.Keyword), int(in.Limit))
	if err != nil {
		return nil, xerr.RpcErr(err)
	}

	items := make([]*memberpb.SearchMembersByCodeItem, 0, len(rows))
	for _, m := range rows {
		item := &memberpb.SearchMembersByCodeItem{
			Id:   m.ID,
			Code: m.Code,
		}
		if auth := m.Edges.Auth; auth != nil && auth.Username != nil {
			item.Account = *auth.Username
		}
		if profile := m.Edges.Profile; profile != nil {
			item.Nickname = profile.Nickname
		}
		items = append(items, item)
	}

	return &memberpb.SearchMembersByCodeResp{List: items}, nil
}

