package operatorservicelogic

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"oa.98ent.com/p9/operator-base/rpc/ent"
	"oa.98ent.com/p9/operator-base/rpc/ent/operator"
	"oa.98ent.com/p9/operator-base/rpc/ent/operatoragentline"
	operatorservice "oa.98ent.com/p9/operator-base/rpc/internal/operator"
	"oa.98ent.com/p9/operator-base/rpc/internal/svc"
	"oa.98ent.com/p9/operator-base/rpc/pb/operatorbaserpc/operatorpb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InitializeBasicResourcesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInitializeBasicResourcesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitializeBasicResourcesLogic {
	return &InitializeBasicResourcesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 初始化基础资源分配到指定分站
func (l *InitializeBasicResourcesLogic) InitializeBasicResources(in *operatorpb.InitializeBasicResourcesRequest) (*operatorpb.InitializeBasicResourcesResponse, error) {
	operatorCode := strings.TrimSpace(in.GetOperatorCode())
	// 克隆请求数组，后续规范化编码时不修改原 RPC 请求中的切片内容。
	languageCodes := slices.Clone(in.GetLanguageCodes())
	regionCodes := slices.Clone(in.GetRegionCodes())
	agentLineCodes := slices.Clone(in.GetAgentLineCodes())

	// 验证参数
	if err := l.validateRequest(operatorCode, languageCodes, regionCodes, agentLineCodes); err != nil {
		l.Logger.Errorw(
			"初始化基础资源分配参数验证失败",
			logx.Field("operator_code", operatorCode),
			logx.Field("error", err.Error()),
		)
		return &operatorpb.InitializeBasicResourcesResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	// 查询 operator 是否存在
	operatorData, err := l.svcCtx.DB.Operator.Query().
		Where(operator.CodeEQ(operatorCode)).
		First(l.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			errMsg := fmt.Sprintf("operator %s not found", operatorCode)
			l.Logger.Errorw("operator not found",
				logx.Field("operator_code", operatorCode),
			)
			return &operatorpb.InitializeBasicResourcesResponse{
				Success: false,
				Message: errMsg,
			}, status.Error(codes.NotFound, errMsg)
		}
		l.Logger.Errorw("query operator failed",
			logx.Field("operator_code", operatorCode),
			logx.Field("error", err.Error()),
		)
		return &operatorpb.InitializeBasicResourcesResponse{
			Success: false,
			Message: fmt.Sprintf("query operator failed: %v", err),
		}, err
	}

	// 开启初始化事务
	tx, err := l.svcCtx.DB.Tx(l.ctx)
	if err != nil {
		l.Logger.Errorw("begin transaction failed",
			logx.Field("operator_code", operatorCode),
			logx.Field("error", err.Error()),
		)
		return &operatorpb.InitializeBasicResourcesResponse{
			Success: false,
			Message: fmt.Sprintf("begin transaction failed: %v", err),
		}, err
	}

	// 初始化统计
	stats := &operatorpb.InitializeBasicResourcesResponse{
		Success: true,
		Message: "success",
		LanguageStat: &operatorpb.ResourceStat{
			Total: int64(len(languageCodes)),
		},
		RegionStat: &operatorpb.ResourceStat{
			Total: int64(len(regionCodes)),
		},
		AgentLineStat: &operatorpb.ResourceStat{
			Total: int64(len(agentLineCodes)),
		},
	}

	// 删除旧的代理线路分配（防止重复约束冲突）
	if _, err = tx.OperatorAgentLine.Delete().
		Where(operatoragentline.OperatorIDEQ(operatorData.ID)).
		Exec(l.ctx); err != nil {
		_ = tx.Rollback()
		l.Logger.Errorw("delete old operator agent line failed",
			logx.Field("operator_code", operatorCode),
			logx.Field("error", err.Error()),
		)
		return &operatorpb.InitializeBasicResourcesResponse{
			Success: false,
			Message: fmt.Sprintf("delete old agent line failed: %v", err),
		}, err
	}

	// 去除语言编码首尾空白，再复用本地分配同步逻辑，保留已有启停状态。
	for index, languageCode := range languageCodes {
		languageCodes[index] = strings.TrimSpace(languageCode)
	}
	// 复用本地 SyncLanguages，在当前事务中同步分配范围，不调用远程 RPC。
	// 保留仍获分配语言的启停状态，撤回未提交项，新增项默认启用；不是覆盖前端勾选状态。
	if err = operatorservice.SyncLanguages(l.ctx, tx, operatorData.ID, languageCodes); err != nil {
		// 同步失败则回滚整个初始化事务，包含此前对代理线路分配的删除。
		_ = tx.Rollback()
		stats.Success = false
		stats.Message = fmt.Sprintf("sync language failed: %v", err)
		stats.LanguageStat.Failed = int64(len(languageCodes))
		return stats, err
	}
	// 统计本次成功同步的分配数量，包含已有项；此时尚未提交，后续步骤仍可能回滚。
	stats.LanguageStat.Success = int64(len(languageCodes))

	// 地区编码统一大写；同步分配范围，失败则回滚当前初始化事务。
	for index, regionCode := range regionCodes {
		regionCodes[index] = strings.ToUpper(strings.TrimSpace(regionCode))
	}
	// 复用本地 SyncRegions，按同一规则同步地区分配，保留已有状态、新增项默认启用。
	// 使用同一事务写库，不调用 RPC；后续代理线路处理成功并提交后才一起生效。
	if err = operatorservice.SyncRegions(l.ctx, tx, operatorData.ID, regionCodes); err != nil {
		// 地区同步失败时，前面已完成的语言同步也随当前事务一并回滚。
		_ = tx.Rollback()
		stats.Success = false
		stats.Message = fmt.Sprintf("sync region failed: %v", err)
		stats.RegionStat.Failed = int64(len(regionCodes))
		return stats, err
	}
	// 这里统计的是同步分配数量，不是新增行数，也不代表事务已经提交。
	stats.RegionStat.Success = int64(len(regionCodes))

	// 创建代理线路分配
	if len(agentLineCodes) > 0 {
		agentBuilders := make([]*ent.OperatorAgentLineCreate, 0, len(agentLineCodes))
		for _, agentLineCode := range agentLineCodes {
			agentBuilders = append(agentBuilders,
				tx.OperatorAgentLine.
					Create().
					SetOperatorID(operatorData.ID).
					SetAgentLineCode(agentLineCode),
			)
		}

		if err = tx.OperatorAgentLine.CreateBulk(agentBuilders...).Exec(l.ctx); err != nil {
			_ = tx.Rollback()
			l.Logger.Errorw("create operator agent line failed",
				logx.Field("operator_code", operatorCode),
				logx.Field("error", err.Error()),
			)
			stats.Success = false
			stats.Message = fmt.Sprintf("create agent line failed: %v", err)
			stats.AgentLineStat.Failed = int64(len(agentLineCodes))
			return stats, err
		}
		stats.AgentLineStat.Success = int64(len(agentLineCodes))
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		l.Logger.Errorw("commit transaction failed",
			logx.Field("operator_code", operatorCode),
			logx.Field("error", err.Error()),
		)
		return &operatorpb.InitializeBasicResourcesResponse{
			Success: false,
			Message: fmt.Sprintf("commit transaction failed: %v", err),
		}, err
	}

	l.Logger.Infow("initialize basic resources success",
		logx.Field("operator_code", operatorCode),
		logx.Field("language_count", stats.LanguageStat.Success),
		logx.Field("region_count", stats.RegionStat.Success),
		logx.Field("agent_line_count", stats.AgentLineStat.Success),
	)

	return stats, nil
}

// validateRequest 验证初始化基础资源请求参数
func (l *InitializeBasicResourcesLogic) validateRequest(operatorCode string, languageCodes, regionCodes, agentLineCodes []string) error {
	// 验证 operator 编码
	if operatorCode == "" {
		return status.Error(codes.InvalidArgument, "operator_code is required")
	}

	// 验证语言编码
	if len(languageCodes) == 0 {
		return status.Error(codes.InvalidArgument, "language_codes is required")
	}
	seen := make(map[string]struct{})
	for _, code := range languageCodes {
		code = strings.TrimSpace(code)
		if code == "" {
			return status.Error(codes.InvalidArgument, "language_codes contains empty value")
		}
		if _, exists := seen[code]; exists {
			return status.Error(codes.InvalidArgument, "language_codes contains duplicate value")
		}
		seen[code] = struct{}{}
	}

	// 验证地区编码
	if len(regionCodes) == 0 {
		return status.Error(codes.InvalidArgument, "region_codes is required")
	}
	seen = make(map[string]struct{})
	for _, code := range regionCodes {
		code = strings.ToUpper(strings.TrimSpace(code))
		if code == "" {
			return status.Error(codes.InvalidArgument, "region_codes contains empty value")
		}
		if _, exists := seen[code]; exists {
			return status.Error(codes.InvalidArgument, "region_codes contains duplicate value")
		}
		seen[code] = struct{}{}
	}

	// 验证代理线路编码
	if len(agentLineCodes) == 0 {
		return status.Error(codes.InvalidArgument, "agent_line_codes is required")
	}
	seen = make(map[string]struct{})
	for _, code := range agentLineCodes {
		code = strings.TrimSpace(code)
		if code == "" {
			return status.Error(codes.InvalidArgument, "agent_line_codes contains empty value")
		}
		if _, exists := seen[code]; exists {
			return status.Error(codes.InvalidArgument, "agent_line_codes contains duplicate value")
		}
		seen[code] = struct{}{}
	}

	return nil
}
