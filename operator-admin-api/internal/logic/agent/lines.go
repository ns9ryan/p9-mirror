package agent

import (
	"context"
	"strings"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/core/common/i18n"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-base/rpc/pb/operatorbaserpc/operatorpb"
)

// OpenedLineCodes 返回当前分站已开通的代理子线路，顺序与 operator-base 一致。
func OpenedLineCodes(ctx context.Context, svcCtx *svc.ServiceContext) ([]string, error) {
	claims := ctxdata.ClaimsFromCtx(ctx)
	if claims == nil || strings.TrimSpace(claims.OperatorCode) == "" {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	resp, err := svcCtx.OperatorBaseRpc.GetAgentLines(ctx, &operatorpb.GetAgentLinesRequest{
		OperatorCode: strings.TrimSpace(claims.OperatorCode),
	})
	if err != nil {
		return nil, err
	}
	return resp.GetAgentLineCodes(), nil
}
