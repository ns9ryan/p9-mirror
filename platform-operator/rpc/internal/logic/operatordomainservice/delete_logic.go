package operatordomainservicelogic

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"oa.98ent.com/p9/common/xerr"
	nodedispatchdispatchpb "oa.98ent.com/p9/node-dispatch/rpc/pb/nodedispatchrpc/dispatchpb"
	nodedispatchoperatornodepb "oa.98ent.com/p9/node-dispatch/rpc/pb/nodedispatchrpc/operatornodepb"
	"oa.98ent.com/p9/platform-operator/pkg/i18nkey"
	"oa.98ent.com/p9/platform-operator/rpc/ent/operator"
	"oa.98ent.com/p9/platform-operator/rpc/ent/operatordomain"
	"oa.98ent.com/p9/platform-operator/rpc/internal/enterror"
	"oa.98ent.com/p9/platform-operator/rpc/internal/svc"
	"oa.98ent.com/p9/platform-operator/rpc/pb/platformoperatorrpc/domainpb"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteLogic {
	return &DeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Delete 删除分站域名
func (l *DeleteLogic) Delete(in *domainpb.DeleteDomainRequest) (*domainpb.DeleteDomainResponse, error) {
	// 域名ID必须大于0
	if in.Id <= 0 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
	}

	// 根据ID查找域名对应分站ID，通过分站ID查找分站编码
	operatorDomain, err := l.svcCtx.DB.OperatorDomain.Query().
		Where(operatordomain.IDEQ(in.Id)).
		Only(l.ctx)
	if err != nil {
		l.Errorf("根据ID查询分站域名失败: id=%d, error=%v", in.Id, err)
		return nil, enterror.Handle(l.Logger, err)
	}
	operatorID := operatorDomain.OperatorID
	operator, err := l.svcCtx.DB.Operator.Query().
		Where(operator.IDEQ(operatorID)).
		Only(l.ctx)
	if err != nil {
		l.Errorf("根据ID查询分站编码失败: operator_id=%d, error=%v", operatorID, err)
		return nil, enterror.Handle(l.Logger, err)
	}
	operatorCode := operator.Code
	// 获取当前启用域名
	domainDataList, err := l.svcCtx.DB.OperatorDomain.
		Query().
		Where(
			operatordomain.OperatorIDEQ(operator.ID),
			operatordomain.DomainTypeEQ(operatorDomain.DomainType),
		).
		Order(operatordomain.ByDomainType()).
		All(l.ctx)
	if err != nil {
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}
	if len(domainDataList) <= 1 {
		l.Errorf("删除后当前类型没有启用的域名: operator_code=%s, domain_type=%d", operatorCode, operatorDomain.DomainType)
		return nil, xerr.RpcErr(xerr.BadRequest("删除后当前类型没有启用的域名"))
	}

	// 删除分站域名
	err = l.svcCtx.DB.OperatorDomain.
		DeleteOneID(in.Id).
		Exec(l.ctx)
	if err != nil {
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}
	// 获取分站部署节点
	operatorNodeResult, err := l.svcCtx.NodeDispatchOperatorNodeRpc.Get(
		l.ctx,
		&nodedispatchoperatornodepb.GetOperatorNodeRequest{
			OperatorCode: operatorCode,
		},
	)
	if err != nil {
		// 如果无法获取节点，记录警告但继续，因为本地数据已删除
		if status.Code(err) == codes.NotFound {
			l.Errorf("无法获取分站部署节点，跳过删除任务提交: operator_code=%s", operatorCode)
			return &domainpb.DeleteDomainResponse{}, nil
		}

		l.Errorf("获取分站部署节点失败，跳过删除任务提交: operator_code=%s, error=%v", operatorCode, err)
		return &domainpb.DeleteDomainResponse{}, nil
	}

	// 检查节点信息完整性
	if operatorNodeResult == nil || operatorNodeResult.OperatorNode == nil || operatorNodeResult.OperatorNode.Node == nil {
		l.Errorf("分站部署节点信息为空，跳过删除任务提交: operator_code=%s", operatorCode)
		return &domainpb.DeleteDomainResponse{}, nil
	}
	// 编码删除任务参数，包含 operator_code 与 domain_names（单个域名）
	deleteTaskPayload := map[string]any{
		"operator_code": operatorCode,
		"domain_names":  []string{operatorDomain.DomainName},
	}
	deleteParams, err := json.Marshal(deleteTaskPayload)
	if err != nil {
		l.Errorf("编码删除任务参数失败，跳过删除任务提交: operator_code=%s, error=%v", operatorCode, err)
		return &domainpb.DeleteDomainResponse{}, nil
	}
	nodeData := operatorNodeResult.OperatorNode.Node
	requestNo := uuid.NewString()
	// 提交删除域名任务
	submitResp, err := l.svcCtx.NodeDispatchDispatchRpc.SubmitTask(
		l.ctx,
		&nodedispatchdispatchpb.SubmitTaskRequest{
			RequestNo: requestNo,                    // 请求编号
			Target:    dispatchTargetOperatorBase,   // 目标服务
			TaskType:  dispatchTaskTypeDeleteDomain, // 任务类型
			NodeCode:  nodeData.Code,                // 执行节点编码
			Params:    string(deleteParams),         // 任务参数
		},
	)
	if err != nil {
		l.Errorf("提交DELETE_OPERATOR_DOMAIN任务失败: operator_code=%s, error=%v", operatorCode, err)
	} else if submitResp != nil {
		l.Infof("DELETE_OPERATOR_DOMAIN任务已提交: operator_code=%s, task_no=%s", operatorCode, submitResp.TaskNo)
	}

	// 返回删除结果
	return &domainpb.DeleteDomainResponse{}, nil
}
