package operatorservicelogic

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"oa.98ent.com/p9/common/xerr"
	nodedispatchdispatchpb "oa.98ent.com/p9/node-dispatch/rpc/pb/nodedispatchrpc/dispatchpb"
	nodedispatchoperatornodepb "oa.98ent.com/p9/node-dispatch/rpc/pb/nodedispatchrpc/operatornodepb"
	"oa.98ent.com/p9/platform-operator/pkg/i18nkey"
	"oa.98ent.com/p9/platform-operator/rpc/ent/operatoradmin"
	"oa.98ent.com/p9/platform-operator/rpc/ent/operatoragentlineallocation"
	"oa.98ent.com/p9/platform-operator/rpc/ent/operatordomain"
	"oa.98ent.com/p9/platform-operator/rpc/ent/operatorlanguageallocation"
	"oa.98ent.com/p9/platform-operator/rpc/ent/operatorprofile"
	"oa.98ent.com/p9/platform-operator/rpc/ent/operatorregionallocation"
	"oa.98ent.com/p9/platform-operator/rpc/internal/enterror"
	"oa.98ent.com/p9/platform-operator/rpc/internal/svc"
	"oa.98ent.com/p9/platform-operator/rpc/pb/platformoperatorrpc/operatorpb"
)

const (
	// 删除任务类型
	dispatchTaskTypeDeleteBasicResources = "DELETE_OPERATOR_BASIC_RESOURCES" // 删除基础资源任务
	dispatchTaskTypeDeleteDomain         = "DELETE_OPERATOR_DOMAIN"          // 删除域名任务
	dispatchTaskTypeDeleteGameResources  = "DELETE_OPERATOR_GAME_ALLOCATION" // 删除游戏资源任务
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

// Delete 删除分站
func (l *DeleteLogic) Delete(in *operatorpb.DeleteOperatorRequest) (*operatorpb.DeleteOperatorResponse, error) {
	l.Infof("收到分站删除请求: operator_id=%d", in.Id)
	// 分站ID必须大于0
	if in.Id <= 0 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
	}
	// 获取当前分站
	current, err := l.svcCtx.DB.Operator.Get(l.ctx, in.Id)
	if err != nil {
		l.Errorf("根据分站ID获取当前分站失败: operator_id=%d, error=%v", in.Id, err)
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// 仅未发布的分站允许删除
	if current.PublishStatus != 1 {
		l.Errorf("分站未发布，禁止删除: operator_id=%d, publish_status=%d", in.Id, current.PublishStatus)
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ConstraintError))
	}

	// 开启数据库事务
	tx, err := l.svcCtx.DB.Tx(l.ctx)
	if err != nil {
		l.Errorf("开启数据库事务失败: operator_id=%d, error=%v", in.Id, err)
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}
	defer func() {
		l.Errorf("回滚数据库事务: operator_id=%d", in.Id)
		_ = tx.Rollback()
	}()

	// 删除语言分配
	_, err = tx.OperatorLanguageAllocation.
		Delete().
		Where(operatorlanguageallocation.OperatorIDEQ(in.Id)).
		Exec(l.ctx)
	if err != nil {
		l.Errorf("删除语言分配失败: operator_id=%d, error=%v", in.Id, err)
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// 删除经营地区分配
	_, err = tx.OperatorRegionAllocation.
		Delete().
		Where(operatorregionallocation.OperatorIDEQ(in.Id)).
		Exec(l.ctx)
	if err != nil {
		l.Errorf("删除经营地区分配失败: operator_id=%d, error=%v", in.Id, err)
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// 删除代理子线路分配
	_, err = tx.OperatorAgentLineAllocation.
		Delete().
		Where(operatoragentlineallocation.OperatorIDEQ(in.Id)).
		Exec(l.ctx)
	if err != nil {
		l.Errorf("删除代理子线路分配失败: operator_id=%d, error=%v", in.Id, err)
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// 在删除数据库记录之前，先收集当前分站的域名列表（用于后续通知节点清理 nginx 配置）
	var domainNames []string
	if domains, err := l.svcCtx.DB.OperatorDomain.Query().Where(operatordomain.OperatorIDEQ(in.Id)).All(l.ctx); err == nil {
		for _, d := range domains {
			if d != nil && d.DomainName != "" {
				domainNames = append(domainNames, d.DomainName)
			}
		}
	} else {
		l.Errorf("获取分站域名列表失败（继续删除），operator_id=%d, error=%v", in.Id, err)
	}

	// 删除分站域名
	_, err = tx.OperatorDomain.
		Delete().
		Where(operatordomain.OperatorIDEQ(in.Id)).
		Exec(l.ctx)
	if err != nil {
		l.Errorf("删除分站域名失败: operator_id=%d, error=%v", in.Id, err)
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// 删除分站档案
	_, err = tx.OperatorProfile.
		Delete().
		Where(operatorprofile.OperatorIDEQ(in.Id)).
		Exec(l.ctx)
	if err != nil {
		l.Errorf("删除分站档案失败: operator_id=%d, error=%v", in.Id, err)
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// 按分站清理管理员账号
	_, err = tx.OperatorAdmin.
		Delete().
		Where(operatoradmin.OperatorIDEQ(in.Id)).
		Exec(l.ctx)
	if err != nil {
		l.Errorf("按分站清理管理员账号失败: operator_id=%d, error=%v", in.Id, err)
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// 删除分站
	err = tx.Operator.
		DeleteOneID(in.Id).
		Exec(l.ctx)
	if err != nil {
		l.Errorf("删除分站失败: operator_id=%d, error=%v", in.Id, err)
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// 提交数据库事务
	if err = tx.Commit(); err != nil {
		l.Errorf("提交数据库事务失败: operator_id=%d, error=%v", in.Id, err)
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	l.Infof("分站本地数据已删除: operator_id=%d, operator_code=%s", in.Id, current.Code)

	// ==================== 提交删除任务到调度中心 ====================

	// 获取分站部署节点
	operatorNodeResult, err := l.svcCtx.NodeDispatchOperatorNodeRpc.Get(
		l.ctx,
		&nodedispatchoperatornodepb.GetOperatorNodeRequest{
			OperatorCode: current.Code,
		},
	)
	if err != nil {
		// 如果无法获取节点，记录警告但继续，因为本地数据已删除
		if status.Code(err) == codes.NotFound {
			l.Errorf("无法获取分站部署节点，跳过删除任务提交: operator_code=%s", current.Code)
			return &operatorpb.DeleteOperatorResponse{}, nil
		}

		l.Errorf("获取分站部署节点失败，跳过删除任务提交: operator_code=%s, error=%v", current.Code, err)
		return &operatorpb.DeleteOperatorResponse{}, nil
	}

	// 检查节点信息完整性
	if operatorNodeResult == nil || operatorNodeResult.OperatorNode == nil || operatorNodeResult.OperatorNode.Node == nil {
		l.Errorf("分站部署节点信息为空，跳过删除任务提交: operator_code=%s", current.Code)
		return &operatorpb.DeleteOperatorResponse{}, nil
	}

	nodeData := operatorNodeResult.OperatorNode.Node

	// 检查节点是否在线
	if nodeData.Status != 1 || !nodeData.Online {
		l.Errorf("分站部署节点不在线，跳过删除任务提交: operator_code=%s, node_code=%s", current.Code, nodeData.Code)
		return &operatorpb.DeleteOperatorResponse{}, nil
	}

	// 生成删除请求编号
	deleteRequestNo := uuid.NewString()

	// 编码删除任务参数，包含 operator_code 与 domain_names（如有）
	deleteTaskPayload := map[string]any{
		"operator_code": current.Code,
	}
	if len(domainNames) > 0 {
		deleteTaskPayload["domain_names"] = domainNames
	}
	deleteParams, err := json.Marshal(deleteTaskPayload)
	if err != nil {
		l.Errorf("编码删除任务参数失败，跳过删除任务提交: operator_code=%s, error=%v", current.Code, err)
		return &operatorpb.DeleteOperatorResponse{}, nil
	}

	// 提交删除任务到调度中心
	l.Infof("提交分站删除任务到调度中心: operator_code=%s, request_no=%s, node_code=%s",
		current.Code, deleteRequestNo, nodeData.Code)

	// ==================== 提交三个删除任务 ====================

	// 1. 删除基础资源（语言、地区、代理线）
	submitResp1, err := l.svcCtx.NodeDispatchDispatchRpc.SubmitTask(
		l.ctx,
		&nodedispatchdispatchpb.SubmitTaskRequest{
			RequestNo: deleteRequestNo + "-1",               // 请求编号1
			Target:    dispatchTargetOperatorBase,           // 目标服务
			TaskType:  dispatchTaskTypeDeleteBasicResources, // 任务类型
			NodeCode:  nodeData.Code,                        // 执行节点编码
			Params:    string(deleteParams),                 // 任务参数
		},
	)
	if err != nil {
		l.Errorf("提交DELETE_OPERATOR_BASIC_RESOURCES任务失败: operator_code=%s, error=%v", current.Code, err)
	} else if submitResp1 != nil {
		l.Infof("DELETE_OPERATOR_BASIC_RESOURCES任务已提交: operator_code=%s, task_no=%s", current.Code, submitResp1.TaskNo)
	}

	// 2. 删除域名
	submitResp2, err := l.svcCtx.NodeDispatchDispatchRpc.SubmitTask(
		l.ctx,
		&nodedispatchdispatchpb.SubmitTaskRequest{
			RequestNo: deleteRequestNo + "-2",       // 请求编号2
			Target:    dispatchTargetOperatorBase,   // 目标服务
			TaskType:  dispatchTaskTypeDeleteDomain, // 任务类型
			NodeCode:  nodeData.Code,                // 执行节点编码
			Params:    string(deleteParams),         // 任务参数
		},
	)
	if err != nil {
		l.Errorf("提交DELETE_OPERATOR_DOMAIN任务失败: operator_code=%s, error=%v", current.Code, err)
	} else if submitResp2 != nil {
		l.Infof("DELETE_OPERATOR_DOMAIN任务已提交: operator_code=%s, task_no=%s", current.Code, submitResp2.TaskNo)
	}

	// 3. 删除游戏资源
	submitResp3, err := l.svcCtx.NodeDispatchDispatchRpc.SubmitTask(
		l.ctx,
		&nodedispatchdispatchpb.SubmitTaskRequest{
			RequestNo: deleteRequestNo + "-3",              // 请求编号3
			Target:    dispatchTargetOperatorGame,          // 目标服务
			TaskType:  dispatchTaskTypeDeleteGameResources, // 任务类型
			NodeCode:  nodeData.Code,                       // 执行节点编码
			Params:    string(deleteParams),                // 任务参数
		},
	)
	if err != nil {
		l.Errorf("提交DELETE_OPERATOR_GAME_ALLOCATION任务失败: operator_code=%s, error=%v", current.Code, err)
	} else if submitResp3 != nil {
		l.Infof("DELETE_OPERATOR_GAME_ALLOCATION任务已提交: operator_code=%s, task_no=%s", current.Code, submitResp3.TaskNo)
	}
	// 返回删除结果
	return &operatorpb.DeleteOperatorResponse{}, nil
}
