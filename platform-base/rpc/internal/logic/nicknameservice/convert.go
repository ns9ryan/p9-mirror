package nicknameservicelogic

import (
	"context"
	"unicode/utf8"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/platform-base/pkg/i18nkey"
	"oa.98ent.com/p9/platform-base/rpc/ent"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/nicknamepb"
)

func validName(s string) bool {
	return utf8.ValidString(s) && s != "" && utf8.RuneCountInString(s) <= 100
}
func info(row *ent.Nickname) *nicknamepb.NicknameInfo {
	return &nicknamepb.NicknameInfo{Id: row.ID, Name: row.Name}
}

func requirePlatform(ctx context.Context) error {
	c := ctxdata.ClaimsFromCtx(ctx)
	if c == nil || c.UserID <= 0 || !c.IsPlatform || c.TokenType == ctxdata.TokenTypePreview {
		return status.Error(codes.PermissionDenied, i18nkey.Forbidden)
	}
	return nil
}
