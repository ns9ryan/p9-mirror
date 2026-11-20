package operatorservicelogic

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-base/rpc/ent"
	"oa.98ent.com/p9/operator-base/rpc/ent/operatorregion"
	"oa.98ent.com/p9/operator-base/rpc/internal/svc"
	"oa.98ent.com/p9/operator-base/rpc/pb/operatorbaserpc/operatorpb"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/regionpb"
)

// 经营地区使用两个大写英文字母编码，输入会先去除首尾空白并转为大写。
var siteRegionCode = regexp.MustCompile(`^[A-Z]{2}$`)

type UpdateSiteRegionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateSiteRegionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSiteRegionsLogic {
	return &UpdateSiteRegionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateSiteRegions 保存本站经营地区完整勾选结果，不改变总网资源分配。
func (l *UpdateSiteRegionsLogic) UpdateSiteRegions(in *operatorpb.UpdateSiteRegionsRequest) (
	*operatorpb.UpdateSiteRegionsResponse, error,
) {
	// 站点身份来自认证上下文，不接受请求自行指定 operator_code。
	operatorCode := ctxdata.OperatorCodeFromCtx(l.ctx)
	if operatorCode == "" {
		return nil, status.Error(codes.PermissionDenied, "login operator is required")
	}
	// 规范化完整勾选列表并拒绝无效或重复编码；空列表合法，表示全部关闭。
	regionCodes, err := normalizeSiteRegionCodes(in.GetRegionCodes())
	if err != nil {
		return nil, err
	}
	if l.svcCtx.DB == nil {
		return nil, status.Error(codes.FailedPrecondition, "operator database is not configured")
	}
	// 先读取本站分配和选择，在事务外校验需要新开启的基础资料。
	before, err := l.loadRegionSelection(l.svcCtx.DB)
	if err != nil {
		return nil, err
	}
	// 只为本站已分配地区计算启停变化，未分配的编码不能借保存操作新增。
	plan, err := planSiteRegionSelection(before.rows, regionCodes)
	if err != nil {
		return nil, err
	}
	// 只有新开启地区才依赖总网资料；全部关闭或保留已有选择时不调用地区 RPC。
	if len(plan.activateCodes) > 0 {
		if l.svcCtx.RegionRpc == nil {
			return nil, status.Error(codes.FailedPrecondition, "region RPC is not configured")
		}
		if err := l.validateRegionActivation(plan.activateCodes); err != nil {
			return nil, err
		}
	}
	// 空列表会关闭本站全部已分配地区；事务保存只改勾选状态，不新增或删除分配。
	if err := l.saveRegionSelection(before, plan); err != nil {
		return nil, err
	}
	// 事务提交成功后返回规范化的完整勾选列表，不返回总网全部地区。
	return &operatorpb.UpdateSiteRegionsResponse{RegionCodes: regionCodes}, nil
}

type siteRegionSelection struct {
	// 事务外读取的分配与勾选快照，事务内重新读取后逐项比较，防止覆盖并发修改。
	operatorID int64
	rows       []*ent.OperatorRegion
}

type siteRegionPlan struct {
	// 只收集状态需要变化的本地记录 ID，保持原状态的记录不重复更新。
	enableIDs  []int64
	disableIDs []int64
	// 仅包含从关闭改为开启的编码，需要向总网地区 RPC 校验。
	activateCodes []string
}

// loadRegionSelection 复用登录站点归属查询，读取本站已分配选项。
func (l *UpdateSiteRegionsLogic) loadRegionSelection(db *ent.Client) (
	*siteRegionSelection, error,
) {
	// db 可为事务客户端，事务内复核继续使用同一认证站点。
	owner, err := lookupSiteOwner(l.ctx, db, l.Logger)
	if err != nil {
		return nil, err
	}
	// 读取本站全部分配记录，包含已关闭项；这些记录才是允许勾选的范围。
	current, err := db.OperatorRegion.Query().
		Where(operatorregion.OperatorIDEQ(owner.ID)).
		All(l.ctx)
	if err != nil {
		return nil, l.updateError(err)
	}
	return &siteRegionSelection{
		operatorID: owner.ID,
		rows:       current,
	}, nil
}

// planSiteRegionSelection 计算启停记录，只接受已分配选项。
func planSiteRegionSelection(current []*ent.OperatorRegion, regionCodes []string) (*siteRegionPlan, error) {
	plan := &siteRegionPlan{}
	selected := make(map[string]bool, len(regionCodes))
	for _, code := range regionCodes {
		selected[code] = true
	}
	existing := make(map[string]*ent.OperatorRegion, len(current))
	// 勾选且未启用的记录需要开启，未勾选且未关闭的记录需要关闭。
	for _, row := range current {
		existing[row.RegionCode] = row
		if selected[row.RegionCode] && row.Status != 1 {
			plan.enableIDs = append(plan.enableIDs, row.ID)
		} else if !selected[row.RegionCode] && row.Status != 2 {
			plan.disableIDs = append(plan.disableIDs, row.ID)
		}
	}
	for _, code := range regionCodes {
		row := existing[code]
		if row == nil {
			return nil, status.Errorf(codes.PermissionDenied, "region is not allocated to this operator: %s", code)
		}
		// 保留已开启但总网资料暂时不可用的勾选，只对新开启项做有效性校验。
		if row.Status != 1 {
			plan.activateCodes = append(plan.activateCodes, code)
		}
	}
	return plan, nil
}

// sameRegionSelection 对比资源分配和勾选状态，防止覆盖并发修改。
func sameRegionSelection(before, current *siteRegionSelection) bool {
	if before.operatorID != current.operatorID || len(before.rows) != len(current.rows) {
		return false
	}
	type regionState struct {
		code   string
		status int64
	}
	// 按记录 ID 比较编码与状态，不把数据库返回顺序变化误判为业务变化。
	snapshot := func(rows []*ent.OperatorRegion) map[int64]regionState {
		states := make(map[int64]regionState, len(rows))
		for _, row := range rows {
			states[row.ID] = regionState{
				code:   row.RegionCode,
				status: row.Status,
			}
		}
		return states
	}
	return reflect.DeepEqual(snapshot(before.rows), snapshot(current.rows))
}

// saveRegionSelection 在可串行化事务内复核并保存勾选状态。
func (l *UpdateSiteRegionsLogic) saveRegionSelection(
	before *siteRegionSelection, plan *siteRegionPlan,
) error {
	// 真正开启事务，后续读取和写入均通过该事务执行。
	tx, err := l.svcCtx.DB.BeginTx(l.ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return l.updateError(err)
	}
	defer func() { _ = tx.Rollback() }()
	// 重新核对归属、分配及状态；发生并发变化时拒绝覆盖。
	current, err := l.loadRegionSelection(tx.Client())
	if err != nil {
		return err
	}
	if !sameRegionSelection(before, current) {
		return status.Error(codes.Aborted, "site region allocation or selection changed; reload and retry")
	}
	// 本次未勾选的已分配选项设为关闭，不删除分配记录。
	if len(plan.disableIDs) > 0 {
		if err := tx.OperatorRegion.Update().
			Where(
				operatorregion.OperatorIDEQ(current.operatorID), operatorregion.IDIn(plan.disableIDs...),
			).
			SetStatus(2).
			Exec(l.ctx); err != nil {
			return l.updateError(err)
		}
	}
	// 仅开启本次明确勾选且已通过基础资料校验的选项。
	if len(plan.enableIDs) > 0 {
		if err := tx.OperatorRegion.Update().
			Where(
				operatorregion.OperatorIDEQ(current.operatorID), operatorregion.IDIn(plan.enableIDs...),
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
func (l *UpdateSiteRegionsLogic) updateError(err error) error {
	l.Logger.Errorw("保存本站经营地区失败", logx.Field("error", err.Error()))
	if l.ctx.Err() != nil {
		return status.FromContextError(l.ctx.Err()).
			Err()
	}
	var pgErr *pgconn.PgError
	// 序列化冲突或死锁要求重新读取后重试，其余数据库细节不直接返回调用方。
	if errors.As(err, &pgErr) && (pgErr.Code == "40001" || pgErr.Code == "40P01") {
		return status.Error(codes.Aborted, "site region allocation or selection changed; reload and retry")
	}
	return status.Error(codes.Internal, "update site regions failed")
}

// normalizeSiteRegionCodes 整理编码并校验格式和重复值。
func normalizeSiteRegionCodes(values []string) ([]string, error) {
	result := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		code := strings.ToUpper(strings.TrimSpace(value))
		if !siteRegionCode.MatchString(code) || seen[code] {
			return nil, status.Error(codes.InvalidArgument, "region_codes contains invalid or duplicate value")
		}
		seen[code] = true
		result = append(result, code)
	}
	// 稳定排序后保存和返回，避免同一组勾选因请求顺序不同产生不一致的展示。
	sort.Strings(result)
	return result, nil
}

// validateRegionActivation 校验本次新开启的选项，已开启选项保留原选择。
func (l *UpdateSiteRegionsLogic) validateRegionActivation(codesToEnable []string) error {
	if len(codesToEnable) == 0 {
		return nil
	}
	// 调用总网地区 RPC，只允许新开启仍有效的地区。
	response, err := l.svcCtx.RegionRpc.ListAll(
		l.ctx,
		&regionpb.ListAllRegionsRequest{},
	)
	if err != nil {
		if l.ctx.Err() != nil {
			return status.FromContextError(l.ctx.Err()).
				Err()
		}
		return status.Error(codes.Unavailable, "query regions failed")
	}
	if response == nil {
		return status.Error(codes.Unavailable, "query regions returned no response")
	}
	// 总网状态 1 才可用于新开启；状态 2 或资料缺失均不允许新启用。
	available := make(map[string]bool, len(response.GetList()))
	for _, row := range response.GetList() {
		if row != nil && row.GetStatus() == 1 {
			available[row.GetCode()] = true
		}
	}
	for _, code := range codesToEnable {
		if !available[code] {
			return status.Errorf(codes.InvalidArgument, "region is unavailable: %s", code)
		}
	}
	return nil
}
