package operatorservicelogic

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"oa.98ent.com/p9/operator-base/rpc/ent"
	"oa.98ent.com/p9/operator-base/rpc/ent/operator"
	"oa.98ent.com/p9/operator-base/rpc/ent/operatoragentline"
	"oa.98ent.com/p9/operator-base/rpc/ent/operatorlanguage"
	"oa.98ent.com/p9/operator-base/rpc/ent/operatorregion"
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

	// 删除旧的语言分配（防止重复约束冲突）
	if _, err = tx.OperatorLanguage.Delete().
		Where(operatorlanguage.OperatorIDEQ(operatorData.ID)).
		Exec(l.ctx); err != nil {
		_ = tx.Rollback()
		l.Logger.Errorw("delete old operator language failed",
			logx.Field("operator_code", operatorCode),
			logx.Field("error", err.Error()),
		)
		return &operatorpb.InitializeBasicResourcesResponse{
			Success: false,
			Message: fmt.Sprintf("delete old language failed: %v", err),
		}, err
	}

	// 删除旧的地区分配（防止重复约束冲突）
	if _, err = tx.OperatorRegion.Delete().
		Where(operatorregion.OperatorIDEQ(operatorData.ID)).
		Exec(l.ctx); err != nil {
		_ = tx.Rollback()
		l.Logger.Errorw("delete old operator region failed",
			logx.Field("operator_code", operatorCode),
			logx.Field("error", err.Error()),
		)
		return &operatorpb.InitializeBasicResourcesResponse{
			Success: false,
			Message: fmt.Sprintf("delete old region failed: %v", err),
		}, err
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

	// 创建语言分配
	if len(languageCodes) > 0 {
		langBuilders := make([]*ent.OperatorLanguageCreate, 0, len(languageCodes))
		for _, languageCode := range languageCodes {
			langBuilders = append(langBuilders,
				tx.OperatorLanguage.
					Create().
					SetOperatorID(operatorData.ID).
					SetLanguageCode(languageCode),
			)
		}

		if err = tx.OperatorLanguage.CreateBulk(langBuilders...).Exec(l.ctx); err != nil {
			_ = tx.Rollback()
			l.Logger.Errorw("create operator language failed",
				logx.Field("operator_code", operatorCode),
				logx.Field("error", err.Error()),
			)
			// 记录失败数量
			stats.Success = false
			stats.Message = fmt.Sprintf("create language failed: %v", err)
			stats.LanguageStat.Failed = int64(len(languageCodes))
			return stats, err
		}
		stats.LanguageStat.Success = int64(len(languageCodes))
	}

	// 创建地区分配
	if len(regionCodes) > 0 {
		regionBuilders := make([]*ent.OperatorRegionCreate, 0, len(regionCodes))
		for _, regionCode := range regionCodes {
			// 地区编码应该转换为大写
			regionCode = strings.ToUpper(strings.TrimSpace(regionCode))
			regionBuilders = append(regionBuilders,
				tx.OperatorRegion.
					Create().
					SetOperatorID(operatorData.ID).
					SetRegionCode(regionCode),
			)
		}

		if err = tx.OperatorRegion.CreateBulk(regionBuilders...).Exec(l.ctx); err != nil {
			_ = tx.Rollback()
			l.Logger.Errorw("create operator region failed",
				logx.Field("operator_code", operatorCode),
				logx.Field("error", err.Error()),
			)
			stats.Success = false
			stats.Message = fmt.Sprintf("create region failed: %v", err)
			stats.RegionStat.Failed = int64(len(regionCodes))
			return stats, err
		}
		stats.RegionStat.Success = int64(len(regionCodes))
	}

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
