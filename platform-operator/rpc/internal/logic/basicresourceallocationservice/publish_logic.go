package basicresourceallocationservicelogic

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"oa.98ent.com/p9/common/xerr"
	nodedispatchdispatchpb "oa.98ent.com/p9/node-dispatch/rpc/pb/nodedispatchrpc/dispatchpb"
	nodedispatchoperatornodepb "oa.98ent.com/p9/node-dispatch/rpc/pb/nodedispatchrpc/operatornodepb"
	"oa.98ent.com/p9/platform-operator/pkg/i18nkey"
	operatorent "oa.98ent.com/p9/platform-operator/rpc/ent/operator"
	operatoragentlineallocationent "oa.98ent.com/p9/platform-operator/rpc/ent/operatoragentlineallocation"
	operatorlanguageallocationent "oa.98ent.com/p9/platform-operator/rpc/ent/operatorlanguageallocation"
	operatorregionallocationent "oa.98ent.com/p9/platform-operator/rpc/ent/operatorregionallocation"
	"oa.98ent.com/p9/platform-operator/rpc/internal/enterror"
	"oa.98ent.com/p9/platform-operator/rpc/internal/svc"
	"oa.98ent.com/p9/platform-operator/rpc/pb/platformoperatorrpc/basicresourceallocationpb"
)

const (
	dispatchTargetOperatorBase           = "operator-base"        // 分站基础服务
	dispatchTaskTypePublishBasicResource = "PUBLISH_BASIC_ALLOCATION" // 发布基础资源分配任务
)

type PublishLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublishLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishLogic {
	return &PublishLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Publish 发布基础资源分配到指定分站
func (l *PublishLogic) Publish(in *basicresourceallocationpb.PublishBasicResourceAllocationRequest) (*basicresourceallocationpb.PublishBasicResourceAllocationResponse, error) {
	// ==================== 发布前校验 ====================
	l.Infof("开始发布基础资源分配", logx.Field("operator_code", in.OperatorCode))
	
	// 分站编码必须不为空
	opCode := strings.TrimSpace(in.OperatorCode)
	if opCode == "" {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
	}

	// 校验分站是否存在
	operatorData, err := l.svcCtx.DB.Operator.
		Query().
		Where(operatorent.CodeEQ(opCode)).
		Only(l.ctx)
	if err != nil {
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// ==================== 基础资源配置验证 ====================
	
	// 获取当前语言分配
	languageCodes, err := l.svcCtx.DB.OperatorLanguageAllocation.
		Query().
		Where(operatorlanguageallocationent.OperatorIDEQ(operatorData.ID)).
		Select(operatorlanguageallocationent.FieldLanguageCode).
		Strings(l.ctx)
	if err != nil {
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// 至少需要一种语言
	if len(languageCodes) == 0 {
		l.Errorf("分站基础资源配置不完整: 未分配语言", logx.Field("operator_code", opCode))
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
	}

	// 获取当前经营地区分配
	regionCodes, err := l.svcCtx.DB.OperatorRegionAllocation.
		Query().
		Where(operatorregionallocationent.OperatorIDEQ(operatorData.ID)).
		Select(operatorregionallocationent.FieldRegionCode).
		Strings(l.ctx)
	if err != nil {
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// 至少需要一个经营地区
	if len(regionCodes) == 0 {
		l.Errorf("分站基础资源配置不完整: 未分配经营地区", logx.Field("operator_code", opCode))
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
	}

	// 获取当前代理子线路分配
	agentLineCodes, err := l.svcCtx.DB.OperatorAgentLineAllocation.
		Query().
		Where(operatoragentlineallocationent.OperatorIDEQ(operatorData.ID)).
		Select(operatoragentlineallocationent.FieldAgentLineCode).
		Strings(l.ctx)
	if err != nil {
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// 至少需要一条代理子线路
	if len(agentLineCodes) == 0 {
		l.Errorf("分站基础资源配置不完整: 未分配代理子线路", logx.Field("operator_code", opCode))
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
	}

	// ==================== 部署节点校验 ====================

	// 获取分站部署节点
	operatorNodeResult, err := l.svcCtx.NodeDispatchOperatorNodeRpc.Get(
		l.ctx,
		&nodedispatchoperatornodepb.GetOperatorNodeRequest{
			OperatorCode: opCode,
		},
	)
	if err != nil {
		// 如果获取失败，直接返回错误
		if status.Code(err) == codes.NotFound {
			l.Errorf("分站部署节点不存在", logx.Field("operator_code", opCode))
			return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
		}

		l.Errorf("获取分站部署节点失败: operator_code=%s, error=%v", opCode, err)
		return nil, err
	}

	// 部署节点信息必须完整
	if operatorNodeResult == nil || operatorNodeResult.OperatorNode == nil || operatorNodeResult.OperatorNode.Node == nil {
		l.Errorf("分站部署节点信息为空", logx.Field("operator_code", opCode))
		return nil, xerr.RpcErr(xerr.InternalServerError(i18nkey.InternalError))
	}

	nodeData := operatorNodeResult.OperatorNode.Node

	// 发布时部署节点必须启用并在线
	if nodeData.Status != 1 || !nodeData.Online {
		l.Errorf("分站部署节点不可用",
			logx.Field("operator_code", opCode),
			logx.Field("node_code", nodeData.Code),
			logx.Field("node_status", nodeData.Status),
			logx.Field("node_online", nodeData.Online),
		)
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
	}

	// ==================== 创建发布请求 ====================

	// 生成本次发布请求编号
	requestNo := uuid.NewString()

	// 编码调度任务参数
	params, err := json.Marshal(struct {
		OperatorCode string `json:"operator_code"`
	}{
		OperatorCode: opCode,
	})
	if err != nil {
		l.Errorf("编码基础资源分配发布任务参数失败: operator_code=%s, error=%v", opCode, err)
		return nil, xerr.RpcErr(xerr.InternalServerError(i18nkey.InternalError))
	}

	// ==================== 提交调度任务 ====================
	l.Infof("提交基础资源分配发布调度任务: operator_code=%s, request_no=%s, node_code=%s", opCode, requestNo, nodeData.Code)
	// 提交基础资源分配发布调度任务
	taskResult, err := l.svcCtx.NodeDispatchDispatchRpc.SubmitTask(
		l.ctx,
		&nodedispatchdispatchpb.SubmitTaskRequest{
			RequestNo: requestNo,                         // 本次发布请求编号
			Target:    dispatchTargetOperatorBase,        // 目标服务
			TaskType:  dispatchTaskTypePublishBasicResource, // 任务类型
			NodeCode:  nodeData.Code,                     // 执行节点编码
			Params:    string(params),                    // 任务参数
		},
	)
	if err != nil {
		l.Errorf("提交基础资源分配发布任务失败: operator_code=%s, request_no=%s, node_code=%s, error=%v",
			opCode, requestNo, nodeData.Code, err)
		return nil, err
	}

	// ==================== 提交结果校验 ====================

	// 调度任务结果必须完整
	if taskResult == nil {
		l.Errorf("提交基础资源分配发布任务结果为空: operator_code=%s, request_no=%s", opCode, requestNo)
		return nil, xerr.RpcErr(xerr.InternalServerError(i18nkey.InternalError))
	}

	// 调度中心必须返回任务编号
	if strings.TrimSpace(taskResult.TaskNo) == "" {
		l.Errorf("提交基础资源分配发布任务编号为空: operator_code=%s, request_no=%s", opCode, requestNo)
		return nil, xerr.RpcErr(xerr.InternalServerError(i18nkey.InternalError))
	}

	l.Infow("基础资源分配发布任务已提交",
		logx.Field("operator_code", opCode),
		logx.Field("request_no", requestNo),
		logx.Field("task_no", taskResult.TaskNo),
		logx.Field("node_code", nodeData.Code),
	)

	return &basicresourceallocationpb.PublishBasicResourceAllocationResponse{}, nil
}
