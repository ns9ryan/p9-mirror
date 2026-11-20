package nicknameservicelogic

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/platform-base/pkg/i18nkey"
	"oa.98ent.com/p9/platform-base/pkg/rpc/grpcerror"
	"oa.98ent.com/p9/platform-base/rpc/internal/cache/basedata"
	nicknamecache "oa.98ent.com/p9/platform-base/rpc/internal/cache/nickname"
	"oa.98ent.com/p9/platform-base/rpc/internal/enterror"
	"oa.98ent.com/p9/platform-base/rpc/internal/svc"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/nicknamepb"
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

func (l *ListLogic) List(in *nicknamepb.ListNicknamesRequest) (*nicknamepb.ListNicknamesResponse, error) {
	if in.GetPage() < 1 || in.GetPageSize() < 1 || in.GetPageSize() > 100 || in.GetPage() > 2147483647/in.GetPageSize() || utf8.RuneCountInString(in.GetKeyword()) > 100 {
		return nil, grpcerror.InvalidArgument(i18nkey.ValidationError)
	}
	rows, err := l.svcCtx.Nicknames.All(l.ctx)
	if err != nil {
		return nil, enterror.Handle(l.Logger, err)
	}
	var matches func(*nicknamecache.Item) bool
	if keyword := strings.ToLower(strings.TrimSpace(in.GetKeyword())); keyword != "" {
		matches = func(row *nicknamecache.Item) bool {
			return strings.Contains(strings.ToLower(row.Name), keyword)
		}
	}
	rows, total := basedata.Page(rows, in.GetPage(), in.GetPageSize(), matches)
	list := make([]*nicknamepb.NicknameInfo, 0, len(rows))
	for _, row := range rows {
		list = append(list, &nicknamepb.NicknameInfo{Id: row.ID, Name: row.Name})
	}
	return &nicknamepb.ListNicknamesResponse{Total: total, List: list}, nil
}
