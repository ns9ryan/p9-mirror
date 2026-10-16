package operatordomainservicelogic

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
	"oa.98ent.com/p9/platform-operator/rpc/ent/operator"
	"oa.98ent.com/p9/platform-operator/rpc/ent/operatordomain"
	"oa.98ent.com/p9/platform-operator/rpc/internal/svc"
	"oa.98ent.com/p9/platform-operator/rpc/pb/platformoperatorrpc/domainpb"
)

const (
	dispatchTargetOperatorBase    = "operator-base"          // 分站基础服务
	dispatchTaskTypePublishDomain = "PUBLISH_DOMAIN"         // 发布域名任务
	dispatchTaskTypeDeleteDomain  = "DELETE_OPERATOR_DOMAIN" // 删除域名任务
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

// 发布域名
func (l *PublishLogic) Publish(in *domainpb.PublishDomainRequest) (*domainpb.PublishDomainResponse, error) {
	// ==================== 发布前校验 ====================

	// 域名ID必须大于0
	id := in.Id
	if id <= 0 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
	}

	l.Infof("收到域名发布请求: id=%d", id)

	// 根据域名ID获取域名数据中的分站ID，通过分站ID获取分站编码
	domainData, err := l.svcCtx.DB.OperatorDomain.Query().
		Where(operatordomain.ID(id)).
		Only(l.ctx)
	if err != nil {
		l.Errorf("根据域名ID获取域名数据失败: id=%d, error=%v", id, err)
		return nil, xerr.RpcErr(xerr.InternalServerError(i18nkey.InternalError))
	}

	operatorId := domainData.OperatorID
	operatorData, err := l.svcCtx.DB.Operator.Query().
		Where(operator.ID(operatorId)).
		Only(l.ctx)
	if err != nil {
		l.Errorf("根据分站ID获取分站数据失败: operator_id=%s, error=%v", operatorId, err)
		return nil, xerr.RpcErr(xerr.InternalServerError(i18nkey.InternalError))
	}
	operatorCode := operatorData.Code

	// ==================== 获取分站部署节点 ====================

	// 获取分站部署节点
	operatorNodeResult, err := l.svcCtx.NodeDispatchOperatorNodeRpc.Get(
		l.ctx,
		&nodedispatchoperatornodepb.GetOperatorNodeRequest{
			OperatorCode: operatorCode,
		},
	)
	if err != nil {
		// 如果获取失败，直接返回错误
		if status.Code(err) == codes.NotFound {
			l.Errorf("分站部署节点不存在: operator_code=%s", operatorCode)
			return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
		}

		l.Errorf("获取分站部署节点失败: operator_code=%s, error=%v", operatorCode, err)
		return nil, err
	}

	// 部署节点信息必须完整
	if operatorNodeResult == nil || operatorNodeResult.OperatorNode == nil || operatorNodeResult.OperatorNode.Node == nil {
		l.Errorf("分站部署节点信息为空: operator_code=%s", operatorCode)
		return nil, xerr.RpcErr(xerr.InternalServerError(i18nkey.InternalError))
	}

	nodeData := operatorNodeResult.OperatorNode.Node

	// 发布时部署节点必须启用并在线
	if nodeData.Status != 1 || !nodeData.Online {
		l.Errorf("分站部署节点不可用: operator_code=%s, node_code=%s, status=%d, online=%v",
			operatorCode, nodeData.Code, nodeData.Status, nodeData.Online)
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
	}

	// ==================== 创建发布请求 ====================

	// 生成本次发布请求编号
	requestNo := uuid.NewString()

	// 编码调度任务参数
	params, err := json.Marshal(struct {
		OperatorCode string `json:"operator_code"`
	}{
		OperatorCode: operatorCode,
	})
	if err != nil {
		l.Errorf("编码域名发布任务参数失败: operator_code=%s, error=%v", operatorCode, err)
		return nil, xerr.RpcErr(xerr.InternalServerError(i18nkey.InternalError))
	}

	// ==================== 提交调度任务 ====================
	l.Infof("提交域名发布调度任务: operator_code=%s, request_no=%s, node_code=%s", operatorCode, requestNo, nodeData.Code)
	// 提交域名发布调度任务
	taskResult, err := l.svcCtx.NodeDispatchDispatchRpc.SubmitTask(
		l.ctx,
		&nodedispatchdispatchpb.SubmitTaskRequest{
			RequestNo: requestNo,                     // 本次发布请求编号
			Target:    dispatchTargetOperatorBase,    // 目标服务
			TaskType:  dispatchTaskTypePublishDomain, // 任务类型
			NodeCode:  nodeData.Code,                 // 执行节点编码
			Params:    string(params),                // 任务参数
		},
	)
	if err != nil {
		l.Errorf("提交域名发布任务失败: operator_code=%s, request_no=%s, node_code=%s, error=%v",
			operatorCode, requestNo, nodeData.Code, err)
		return nil, err
	}

	// ==================== 提交结果校验 ====================

	// 调度任务结果必须完整
	if taskResult == nil {
		l.Errorf("提交域名发布任务结果为空: operator_code=%s, request_no=%s", operatorCode, requestNo)
		return nil, xerr.RpcErr(xerr.InternalServerError(i18nkey.InternalError))
	}

	if strings.TrimSpace(taskResult.TaskNo) == "" {
		l.Errorf("调度中心返回任务编号为空: operator_code=%s, request_no=%s", operatorCode, requestNo)
		return nil, xerr.RpcErr(xerr.InternalServerError(i18nkey.InternalError))
	}

	l.Infof("域名发布任务已成功提交到调度中心: operator_code=%s, task_no=%s, request_no=%s", operatorCode, taskResult.TaskNo, requestNo)

	return &domainpb.PublishDomainResponse{}, nil
}
