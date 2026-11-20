package operatorservicelogic

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"oa.98ent.com/p9/operator-base/rpc/ent"
	"oa.98ent.com/p9/operator-base/rpc/ent/operatordomain"
	"oa.98ent.com/p9/operator-base/rpc/ent/operator"
	"oa.98ent.com/p9/operator-base/rpc/internal/svc"
	"oa.98ent.com/p9/operator-base/rpc/pb/operatorbaserpc/operatorpb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InitializeDomainsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInitializeDomainsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitializeDomainsLogic {
	return &InitializeDomainsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 初始化域名到指定分站
func (l *InitializeDomainsLogic) InitializeDomains(in *operatorpb.InitializeDomainsRequest) (*operatorpb.InitializeDomainsResponse, error) {
	operatorCode := strings.TrimSpace(in.GetOperatorCode())
	domainList := slices.Clone(in.GetDomainList())

	// 验证参数
	if err := l.validateRequest(operatorCode, domainList); err != nil {
		l.Logger.Errorw(
			"初始化域名参数验证失败",
			logx.Field("operator_code", operatorCode),
			logx.Field("error", err.Error()),
		)
		return &operatorpb.InitializeDomainsResponse{
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
			return &operatorpb.InitializeDomainsResponse{
				Success: false,
				Message: errMsg,
			}, status.Error(codes.NotFound, errMsg)
		}
		l.Logger.Errorw("query operator failed",
			logx.Field("operator_code", operatorCode),
			logx.Field("error", err.Error()),
		)
		return &operatorpb.InitializeDomainsResponse{
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
		return &operatorpb.InitializeDomainsResponse{
			Success: false,
			Message: fmt.Sprintf("begin transaction failed: %v", err),
		}, err
	}

	// 删除现有域名记录
	deleteCount, err := tx.OperatorDomain.Delete().
		Where(operatordomain.OperatorIDEQ(operatorData.ID)).
		Exec(l.ctx)
	if err != nil {
		_ = tx.Rollback()
		l.Logger.Errorw("delete existing domains failed",
			logx.Field("operator_code", operatorCode),
			logx.Field("error", err.Error()),
		)
		return &operatorpb.InitializeDomainsResponse{
			Success: false,
			Message: fmt.Sprintf("delete existing domains failed: %v", err),
		}, err
	}

	if deleteCount > 0 {
		l.Logger.Infow("deleted existing domains",
			logx.Field("operator_code", operatorCode),
			logx.Field("delete_count", deleteCount),
		)
	}

	// 初始化统计
	stats := &operatorpb.InitializeDomainsResponse{
		Success: true,
		Message: "success",
		DomainStat: &operatorpb.ResourceStat{
			Total: int64(len(domainList)),
		},
	}

	// 创建域名记录
	if len(domainList) > 0 {
		domainBuilders := make([]*ent.OperatorDomainCreate, 0, len(domainList))
		for _, domain := range domainList {
			if domain == nil {
				continue
			}
			domainBuilders = append(domainBuilders,
				tx.OperatorDomain.
					Create().
					SetOperatorID(operatorData.ID).
					SetDomainName(domain.DomainName).
					SetDomainType(domain.DomainType),
			)
		}

		if err = tx.OperatorDomain.CreateBulk(domainBuilders...).Exec(l.ctx); err != nil {
			_ = tx.Rollback()
			l.Logger.Errorw("create operator domain failed",
				logx.Field("operator_code", operatorCode),
				logx.Field("error", err.Error()),
			)
			// 记录失败数量
			stats.Success = false
			stats.Message = fmt.Sprintf("create domain failed: %v", err)
			stats.DomainStat.Failed = int64(len(domainList))
			return stats, err
		}
		stats.DomainStat.Success = int64(len(domainBuilders))
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		l.Logger.Errorw("commit transaction failed",
			logx.Field("operator_code", operatorCode),
			logx.Field("error", err.Error()),
		)
		return &operatorpb.InitializeDomainsResponse{
			Success: false,
			Message: fmt.Sprintf("commit transaction failed: %v", err),
		}, err
	}

	l.Logger.Infow("initialize domains success",
		logx.Field("operator_code", operatorCode),
		logx.Field("domain_count", stats.DomainStat.Success),
	)

	return stats, nil
}

// validateRequest 验证初始化域名请求参数
func (l *InitializeDomainsLogic) validateRequest(operatorCode string, domainList []*operatorpb.DomainModel) error {
	// 验证 operator 编码
	if operatorCode == "" {
		return status.Error(codes.InvalidArgument, "operator_code is required")
	}

	// 验证域名列表
	if len(domainList) == 0 {
		return status.Error(codes.InvalidArgument, "domain_list is required")
	}

	// 验证每个域名
	for i, domain := range domainList {
		if domain == nil {
			return status.Error(codes.InvalidArgument, fmt.Sprintf("domain_list[%d] is nil", i))
		}

		domain.DomainName = strings.TrimSpace(domain.DomainName)
		if domain.DomainName == "" {
			return status.Error(codes.InvalidArgument, fmt.Sprintf("domain_list[%d].domain_name is required", i))
		}

		// 验证域名类型: 1分站后台, 2代理后台, 3会员H5
		if domain.DomainType < 1 || domain.DomainType > 3 {
			return status.Error(codes.InvalidArgument, fmt.Sprintf("domain_list[%d].domain_type must be 1, 2, or 3", i))
		}
	}

	return nil
}
