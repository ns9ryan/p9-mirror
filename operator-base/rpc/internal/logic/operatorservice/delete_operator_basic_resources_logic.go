package operatorservicelogic

import (
	"context"
	"fmt"
	"strings"

	"oa.98ent.com/p9/operator-base/rpc/ent/operator"
	"oa.98ent.com/p9/operator-base/rpc/ent/operatoragentline"
	"oa.98ent.com/p9/operator-base/rpc/ent/operatordomain"
	"oa.98ent.com/p9/operator-base/rpc/ent/operatorlanguage"
	"oa.98ent.com/p9/operator-base/rpc/ent/operatorregion"
	"oa.98ent.com/p9/operator-base/rpc/internal/svc"
	"oa.98ent.com/p9/operator-base/rpc/pb/operatorbaserpc/operatorpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteOperatorBasicResourcesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteOperatorBasicResourcesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteOperatorBasicResourcesLogic {
	return &DeleteOperatorBasicResourcesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除分站基础资源数据
func (l *DeleteOperatorBasicResourcesLogic) DeleteOperatorBasicResources(in *operatorpb.DeleteOperatorBasicResourcesRequest) (*operatorpb.DeleteOperatorBasicResourcesResponse, error) {
	// ==================== 参数校验 ====================

	operatorCode := strings.TrimSpace(in.OperatorCode)
	if operatorCode == "" {
		return &operatorpb.DeleteOperatorBasicResourcesResponse{
			Success: false,
			Message: "operator_code is required",
		}, nil
	}

	l.Logger.Infof("开始删除分站数据: operator_code=%s", operatorCode)

	// ==================== 查询分站 ====================

	operatorData, err := l.svcCtx.DB.Operator.Query().
		Where(operator.CodeEQ(operatorCode)).
		First(l.ctx)
	if err != nil {
		l.Logger.Errorf("查询分站失败: operator_code=%s, error=%v", operatorCode, err)
		return &operatorpb.DeleteOperatorBasicResourcesResponse{
			Success: false,
			Message: fmt.Sprintf("query operator failed: %v", err),
		}, nil
	}

	// ==================== 删除分站数据 ====================

	// 统计删除数量
	var domainDeleted, languageDeleted, regionDeleted, agentLineDeleted int64

	// 删除域名
	if count, err := l.svcCtx.DB.OperatorDomain.Delete().
		Where(operatordomain.OperatorIDEQ(operatorData.ID)).
		Exec(l.ctx); err != nil {
		l.Logger.Errorf("删除分站域名失败: operator_code=%s, error=%v", operatorCode, err)
		return &operatorpb.DeleteOperatorBasicResourcesResponse{
			Success: false,
			Message: fmt.Sprintf("delete domains failed: %v", err),
		}, nil
	} else {
		domainDeleted = int64(count)
		if count > 0 {
			l.Logger.Infof("已删除分站域名: operator_code=%s, count=%d", operatorCode, count)
		}
	}

	// 删除语言分配
	if count, err := l.svcCtx.DB.OperatorLanguage.Delete().
		Where(operatorlanguage.OperatorIDEQ(operatorData.ID)).
		Exec(l.ctx); err != nil {
		l.Logger.Errorf("删除分站语言分配失败: operator_code=%s, error=%v", operatorCode, err)
		return &operatorpb.DeleteOperatorBasicResourcesResponse{
			Success: false,
			Message: fmt.Sprintf("delete languages failed: %v", err),
		}, nil
	} else {
		languageDeleted = int64(count)
		if count > 0 {
			l.Logger.Infof("已删除分站语言分配: operator_code=%s, count=%d", operatorCode, count)
		}
	}

	// 删除地区分配
	if count, err := l.svcCtx.DB.OperatorRegion.Delete().
		Where(operatorregion.OperatorIDEQ(operatorData.ID)).
		Exec(l.ctx); err != nil {
		l.Logger.Errorf("删除分站地区分配失败: operator_code=%s, error=%v", operatorCode, err)
		return &operatorpb.DeleteOperatorBasicResourcesResponse{
			Success: false,
			Message: fmt.Sprintf("delete regions failed: %v", err),
		}, nil
	} else {
		regionDeleted = int64(count)
		if count > 0 {
			l.Logger.Infof("已删除分站地区分配: operator_code=%s, count=%d", operatorCode, count)
		}
	}

	// 删除代理线路分配
	if count, err := l.svcCtx.DB.OperatorAgentLine.Delete().
		Where(operatoragentline.OperatorIDEQ(operatorData.ID)).
		Exec(l.ctx); err != nil {
		l.Logger.Errorf("删除分站代理线路分配失败: operator_code=%s, error=%v", operatorCode, err)
		return &operatorpb.DeleteOperatorBasicResourcesResponse{
			Success: false,
			Message: fmt.Sprintf("delete agent lines failed: %v", err),
		}, nil
	} else {
		agentLineDeleted = int64(count)
		if count > 0 {
			l.Logger.Infof("已删除分站代理线路分配: operator_code=%s, count=%d", operatorCode, count)
		}
	}

	// ==================== 返回删除结果 ====================

	l.Logger.Infof("成功删除分站数据: operator_code=%s, domain=%d, language=%d, region=%d, agent_line=%d",
		operatorCode, domainDeleted, languageDeleted, regionDeleted, agentLineDeleted)

	return &operatorpb.DeleteOperatorBasicResourcesResponse{
		Success:          true,
		Message:          "success",
		DomainDeleted:    domainDeleted,
		LanguageDeleted:  languageDeleted,
		RegionDeleted:    regionDeleted,
		AgentLineDeleted: agentLineDeleted,
	}, nil
}
