// Package memberexport defines the versioned CSV contract shared by producer and verifier.
package memberexport

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"oa.98ent.com/p9/common/ctxdata"
)

const Path = "/member/export/v2"
const MaxRows = 5000
const MaxBytes = 32 << 20
const Missing = "[NULL]"
const Unavailable = "[UNAVAILABLE]"

var Header = []string{"member_id", "member_code", "username", "nickname", "phone", "email", "status", "agent_id", "agent_account_id", "inviter_member_id", "agent_account", "auto_level", "vip_level", "fixed_levels", "created_at", "last_login_at", "total_deposit_minor", "total_withdrawal_minor", "wallet_balance_minor", "today_deposit_minor", "today_withdrawal_minor"}

// Cell uses a literal apostrophe, never an executable ="..." expression. CSV has
// no cell type metadata: consumers should import identifiers as Text (see docs).
func Cell(value string, identifier bool) string {
	if value == "" || value == Missing || value == Unavailable || value == "[REDACTED]" {
		return value
	}
	if identifier || strings.ContainsAny(value[:1], "=+-@\t\r\n") {
		return "'" + value
	}
	// Spreadsheet applications may trim leading spaces before detecting formulas.
	trimmed := strings.TrimLeft(value, " \v\f")
	if trimmed != "" && strings.ContainsAny(trimmed[:1], "=+-@\t\r\n") {
		return "'" + value
	}
	return value
}

type Authorizer interface {
	CheckToken(context.Context, string) (*ctxdata.Claims, error)
	Enforce(context.Context, *ctxdata.Claims, string, string) (bool, error)
}

var operatorCodePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`)

type Policy struct{ Phone, Email bool }

// Authenticate again at the RPC boundary. Never trust contact flags from callers.
func Authorize(ctx context.Context, auth Authorizer) (context.Context, Policy, error) {
	var p Policy
	raw := ctxdata.RawTokenFromCtx(ctx)
	if auth == nil {
		return ctx, p, status.Error(codes.Unavailable, "export authentication unavailable")
	}
	if raw == "" {
		return ctx, p, status.Error(codes.Unauthenticated, "export credentials missing")
	}
	// Incoming RPC metadata is not automatically forwarded to the next RPC.
	// Core validates the original client IP bound to the token.
	ctx = ctxdata.WithClientIP(ctx, ctxdata.ClientIPFromCtx(ctx))
	claims, err := auth.CheckToken(ctx, raw)
	if err != nil {
		if status.Code(err) == codes.Unauthenticated || status.Code(err) == codes.PermissionDenied || status.Code(err) == codes.Code(498) {
			return ctx, p, status.Error(codes.Unauthenticated, "export authentication failed")
		}
		return ctx, p, authDependencyError(err, "export authentication dependency failed")
	}
	if claims == nil || claims.UserID <= 0 || !claims.UserType.IsOperator() || claims.IsPlatform || claims.TokenType != ctxdata.TokenTypeAccess || !operatorCodePattern.MatchString(claims.OperatorCode) {
		return ctx, p, status.Error(codes.PermissionDenied, "export forbidden")
	}
	// Incoming scope is compared with the verified credential, never accepted as authority.
	if old := ctxdata.ClaimsFromCtx(ctx); old == nil || old.OperatorCode != claims.OperatorCode || old.UserID != claims.UserID {
		return ctx, p, status.Error(codes.PermissionDenied, "export scope mismatch")
	}
	ctx = ctxdata.WithClaims(ctx, claims)
	allowed, err := auth.Enforce(ctx, claims, Path, "POST")
	if err != nil {
		return ctx, p, authDependencyError(err, "export permission unavailable")
	}
	if !allowed {
		return ctx, p, status.Error(codes.PermissionDenied, "export permission denied")
	}
	// No separate contact permissions are registered. Contacts remain redacted.
	return ctx, p, nil
}

func authDependencyError(err error, message string) error {
	code := status.Code(err)
	if code != codes.Canceled && code != codes.DeadlineExceeded {
		code = codes.Unavailable
	}
	return status.Error(code, message)
}

// ColumnIndices validates column names and preserves caller order. Empty means all.
func ColumnIndices(columns []string) ([]int, error) {
	if len(columns) == 0 {
		indices := make([]int, len(Header))
		for i := range indices {
			indices[i] = i
		}
		return indices, nil
	}
	if len(columns) > len(Header) {
		return nil, fmt.Errorf("too many export columns")
	}
	seen := map[string]bool{}
	indices := make([]int, 0, len(columns))
	for _, column := range columns {
		if seen[column] {
			return nil, fmt.Errorf("duplicate export column")
		}
		seen[column] = true
		index := -1
		for i, name := range Header {
			if column == name {
				index = i
				break
			}
		}
		if index < 0 {
			return nil, fmt.Errorf("unsupported export column")
		}
		indices = append(indices, index)
	}
	return indices, nil
}
