package agent

import (
	"context"
	"strings"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
)

const (
	statusNormal   int64 = 1 // 正常
	statusFrozen   int64 = 2 // 冻结
	statusDisabled int64 = 3 // 禁用

	registerSourceAdmin int64 = 3 // 管理员
	registerSourceAgent int64 = 4 // 代理

	maxLevel     int64 = 7   // 最大层级
	maxShareRate int64 = 100 // 最大分享比例
)

// Service 代理配置与开户。
type Service struct {
	DB        *ent.Client
	treeCache *agentTreeCache
}

// New 创建代理服务。
func New(db *ent.Client) *Service {
	return &Service{DB: db, treeCache: newAgentTreeCache()}
}

func requireOperator(ctx context.Context) (context.Context, string, error) {
	c := ctxdata.ClaimsFromCtx(ctx)
	if c == nil || strings.TrimSpace(c.OperatorCode) == "" {
		return ctx, "", xerr.Unauthorized(i18n.Unauthorized)
	}
	code := strings.TrimSpace(c.OperatorCode)
	return ctxdata.WithOperatorCode(ctx, code), code, nil
}

func requireAgentClaims(ctx context.Context) (context.Context, *ctxdata.Claims, error) {
	ctx, _, err := requireOperator(ctx)
	if err != nil {
		return ctx, nil, err
	}
	c := ctxdata.ClaimsFromCtx(ctx)
	if c == nil || c.UserID <= 0 {
		return ctx, nil, xerr.Unauthorized(i18n.Unauthorized)
	}
	return ctx, c, nil
}

func mapEntErr(err error) error {
	if err == nil {
		return nil
	}
	if ent.IsConstraintError(err) {
		return xerr.EntConstraintError(i18n.AgentConflict, err)
	}
	if ent.IsValidationError(err) {
		return xerr.EntValidationError(i18n.InvalidParam, err)
	}
	return err
}
