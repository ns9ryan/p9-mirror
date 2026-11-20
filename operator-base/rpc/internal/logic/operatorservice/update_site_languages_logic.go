package operatorservicelogic

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-base/rpc/ent"
	"oa.98ent.com/p9/operator-base/rpc/ent/operatorlanguage"
	"oa.98ent.com/p9/operator-base/rpc/internal/svc"
	"oa.98ent.com/p9/operator-base/rpc/pb/operatorbaserpc/operatorpb"
)

type UpdateSiteLanguagesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateSiteLanguagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSiteLanguagesLogic {
	return &UpdateSiteLanguagesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateSiteLanguages 保存本站开放语言完整勾选结果，不改变总网资源分配。
func (l *UpdateSiteLanguagesLogic) UpdateSiteLanguages(in *operatorpb.UpdateSiteLanguagesRequest) (
	*operatorpb.UpdateSiteLanguagesResponse, error,
) {
	// 归属取自认证上下文，旧 RPC 域名字段不参与站点选择。
	operatorCode := ctxdata.OperatorCodeFromCtx(l.ctx)
	if operatorCode == "" {
		return nil, status.Error(codes.PermissionDenied, "login operator is required")
	}
	languageCodes, err := normalizeSiteLanguageCodes(in.GetLanguageCodes())
	if err != nil {
		return nil, err
	}
	if l.svcCtx.DB == nil {
		return nil, status.Error(codes.FailedPrecondition, "operator database is not configured")
	}
	// 先读取本站分配和选择，在事务外校验需要新开启的基础资料。
	before, err := l.loadLanguageSelection(l.svcCtx.DB)
	if err != nil {
		return nil, err
	}
	plan, err := planSiteLanguageSelection(before.rows, languageCodes)
	if err != nil {
		return nil, err
	}
	if len(plan.activateCodes) > 0 {
		if l.svcCtx.CoreRpc == nil {
			return nil, status.Error(codes.FailedPrecondition, "language RPC is not configured")
		}
		if err := l.validateLanguageActivation(operatorCode, plan.activateCodes); err != nil {
			return nil, err
		}
	}
	// 空列表会关闭本站全部已分配语言；事务保存只改勾选状态，不新增或删除分配。
	if err := l.saveLanguageSelection(before, plan); err != nil {
		return nil, err
	}
	return &operatorpb.UpdateSiteLanguagesResponse{LanguageCodes: languageCodes}, nil
}

type siteLanguageSelection struct {
	// 事务外读取的分配与勾选快照，事务内重新读取后逐项比较，防止覆盖并发修改。
	operatorID int64
	rows       []*ent.OperatorLanguage
}

type siteLanguagePlan struct {
	// 只收集状态需要变化的本地记录 ID，保持原状态的记录不重复更新。
	enableIDs  []int64
	disableIDs []int64
	// 仅包含从关闭改为开启的编码，需要向本站 Core 校验；已开启项不重新拦截。
	activateCodes []string
}

// loadLanguageSelection 复用登录站点归属查询，读取本站已分配选项。
func (l *UpdateSiteLanguagesLogic) loadLanguageSelection(db *ent.Client) (*siteLanguageSelection, error) {
	owner, err := lookupSiteOwner(l.ctx, db, l.Logger)
	if err != nil {
		return nil, err
	}
	current, err := db.OperatorLanguage.Query().
		Where(operatorlanguage.OperatorIDEQ(owner.ID)).
		All(l.ctx)
	if err != nil {
		return nil, l.updateError(err)
	}
	return &siteLanguageSelection{
		operatorID: owner.ID,
		rows:       current,
	}, nil
}

// planSiteLanguageSelection 计算启停记录，只接受已分配选项。
func planSiteLanguageSelection(current []*ent.OperatorLanguage, languageCodes []string) (*siteLanguagePlan, error) {
	plan := &siteLanguagePlan{}
	selected := make(map[string]bool, len(languageCodes))
	for _, code := range languageCodes {
		selected[code] = true
	}
	existing := make(map[string]*ent.OperatorLanguage, len(current))
	for _, row := range current {
		existing[row.LanguageCode] = row
		if selected[row.LanguageCode] && row.Status != 1 {
			plan.enableIDs = append(plan.enableIDs, row.ID)
		} else if !selected[row.LanguageCode] && row.Status != 2 {
			plan.disableIDs = append(plan.disableIDs, row.ID)
		}
	}
	for _, code := range languageCodes {
		row := existing[code]
		if row == nil {
			return nil, status.Errorf(codes.PermissionDenied, "language is not allocated to this operator: %s", code)
		}
		// 保留已开启但基础资料暂时不可用的勾选，只对新开启项做有效性校验。
		if row.Status != 1 {
			plan.activateCodes = append(plan.activateCodes, code)
		}
	}
	return plan, nil
}

// sameLanguageSelection 对比资源分配和勾选状态，防止覆盖并发修改。
func sameLanguageSelection(before, current *siteLanguageSelection) bool {
	if before.operatorID != current.operatorID || len(before.rows) != len(current.rows) {
		return false
	}
	type languageState struct {
		code   string
		status int64
	}
	snapshot := func(rows []*ent.OperatorLanguage) map[int64]languageState {
		states := make(map[int64]languageState, len(rows))
		for _, row := range rows {
			states[row.ID] = languageState{
				code:   row.LanguageCode,
				status: row.Status,
			}
		}
		return states
	}
	return reflect.DeepEqual(snapshot(before.rows), snapshot(current.rows))
}

// saveLanguageSelection 在可串行化事务内复核并保存勾选状态。
func (l *UpdateSiteLanguagesLogic) saveLanguageSelection(before *siteLanguageSelection, plan *siteLanguagePlan) error {
	// 真正开启事务，后续读取和写入均通过该事务执行。
	tx, err := l.svcCtx.DB.BeginTx(l.ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return l.updateError(err)
	}
	defer func() { _ = tx.Rollback() }()
	// 重新核对归属、分配及状态；发生并发变化时拒绝覆盖。
	current, err := l.loadLanguageSelection(tx.Client())
	if err != nil {
		return err
	}
	if !sameLanguageSelection(before, current) {
		return status.Error(codes.Aborted, "site language allocation or selection changed; reload and retry")
	}
	// 本次未勾选的已分配选项设为关闭，不删除分配记录。
	if len(plan.disableIDs) > 0 {
		if err := tx.OperatorLanguage.Update().
			Where(
				operatorlanguage.OperatorIDEQ(current.operatorID), operatorlanguage.IDIn(plan.disableIDs...),
			).
			SetStatus(2).
			Exec(l.ctx); err != nil {
			return l.updateError(err)
		}
	}
	// 仅开启本次明确勾选且已通过基础资料校验的选项。
	if len(plan.enableIDs) > 0 {
		if err := tx.OperatorLanguage.Update().
			Where(
				operatorlanguage.OperatorIDEQ(current.operatorID), operatorlanguage.IDIn(plan.enableIDs...),
			).
			SetStatus(1).
			Exec(l.ctx); err != nil {
			return l.updateError(err)
		}
	}
	// 所有更新成功后提交，任一步失败由延迟回滚撤销。
	if err := tx.Commit(); err != nil {
		return l.updateError(err)
	}
	return nil
}

// updateError 记录数据库错误，将并发冲突转换为可重试的 gRPC 错误。
func (l *UpdateSiteLanguagesLogic) updateError(err error) error {
	l.Logger.Errorw("保存本站开放语言失败", logx.Field("error", err.Error()))
	if l.ctx.Err() != nil {
		return status.FromContextError(l.ctx.Err()).
			Err()
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && (pgErr.Code == "40001" || pgErr.Code == "40P01") {
		return status.Error(codes.Aborted, "site language allocation or selection changed; reload and retry")
	}
	return status.Error(codes.Internal, "update site languages failed")
}

// normalizeSiteLanguageCodes 整理编码并校验格式和重复值。
func normalizeSiteLanguageCodes(values []string) ([]string, error) {
	result := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		code := strings.TrimSpace(value)
		if code == "" || len(code) > 16 || seen[code] {
			return nil, status.Error(codes.InvalidArgument, "language_codes contains invalid or duplicate value")
		}
		seen[code] = true
		result = append(result, code)
	}
	// 稳定排序后保存和返回，避免同一组勾选因请求顺序不同产生不一致的展示。
	sort.Strings(result)
	return result, nil
}

// validateLanguageActivation 校验本次新开启的选项，已开启选项保留原选择。
func (l *UpdateSiteLanguagesLogic) validateLanguageActivation(operatorCode string, codesToEnable []string) error {
	if len(codesToEnable) == 0 {
		return nil
	}
	// 复用本站 Core RPC 分页查询，只核对本次由关闭改为开启的语言是否仍可用。
	rows, err := listSiteLanguages(l.ctx, l.svcCtx, operatorCode)
	if err != nil {
		if l.ctx.Err() != nil {
			return status.FromContextError(l.ctx.Err()).
				Err()
		}
		return status.Error(codes.Unavailable, "query languages failed")
	}
	for _, code := range codesToEnable {
		row := rows[code]
		if row == nil || row.GetDisabled() != 0 {
			return status.Errorf(codes.InvalidArgument, "language is unavailable: %s", code)
		}
	}
	return nil
}
