package operatorgameallocationservicelogic

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
	"oa.98ent.com/p9/platform-game/rpc/internal/svc"
	"oa.98ent.com/p9/platform-game/rpc/pb/platform_game"
)

const (
	dispatchTargetOperatorGame            = "operator-game"
	dispatchTaskTypePublishGameAllocation = "PUBLISH_GAME_ALLOCATION" // 发布游戏资源分配任务
)

type PublishOperatorGameAllocationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublishOperatorGameAllocationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishOperatorGameAllocationLogic {
	return &PublishOperatorGameAllocationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// PublishOperatorGameAllocation 发布游戏资源分配到指定分站
func (l *PublishOperatorGameAllocationLogic) PublishOperatorGameAllocation(in *platform_game.PublishOperatorGameAllocationRequest) (*platform_game.PublishOperatorGameAllocationResponse, error) {
	// ==================== 发布前校验 ====================
	l.Infof("开始发布游戏资源分配", logx.Field("op_code", in.OpCode))
	// 分站编码必须不为空
	opCode := strings.TrimSpace(in.OpCode)
	if opCode == "" {
		return nil, xerr.RpcErr(xerr.BadRequest("op_code不能为空"))
	}

	// 校验分站是否存在
	operatorExists, err := l.svcCtx.DAOManager.Operator.ExistByCode(l.ctx, opCode)
	if err != nil {
		l.Errorf("检查分站是否存在失败: op_code=%s, error=%v", opCode, err)
		return nil, xerr.RpcErr(xerr.InternalServerError("检查分站状态失败"))
	}

	if !operatorExists {
		l.Errorf("分站不存在", logx.Field("op_code", opCode))
		return nil, xerr.RpcErr(xerr.BadRequest("分站不存在"))
	}

	// ==================== 部署节点校验 ====================
	// 从 node-dispatch 获取分站部署节点信息
	operatorNodeResult, err := l.svcCtx.NodeDispatchOperatorNodeRpc.Get(
		l.ctx,
		&nodedispatchoperatornodepb.GetOperatorNodeRequest{
			OperatorCode: opCode,
		},
	)
	if err != nil {
		// 如果获取失败，直接返回错误
		if status.Code(err) == codes.NotFound {
			l.Errorf("分站部署节点不存在", logx.Field("op_code", opCode))
			return nil, xerr.RpcErr(xerr.BadRequest("分站尚未部署到节点"))
		}

		l.Errorf("获取分站部署节点失败: op_code=%s, error=%v", opCode, err)
		return nil, err
	}

	// 部署节点信息必须完整
	if operatorNodeResult == nil || operatorNodeResult.OperatorNode == nil || operatorNodeResult.OperatorNode.Node == nil {
		l.Errorf("分站部署节点信息为空: op_code=%s", opCode)
		return nil, xerr.RpcErr(xerr.InternalServerError("分站部署节点信息为空"))
	}

	nodeData := operatorNodeResult.OperatorNode.Node

	// 发布时部署节点必须启用并在线
	if nodeData.Status != 1 || !nodeData.Online {
		l.Errorf("分站部署节点不可用",
			logx.Field("op_code", opCode),
			logx.Field("node_code", nodeData.Code),
			logx.Field("node_status", nodeData.Status),
			logx.Field("node_online", nodeData.Online),
		)
		return nil, xerr.RpcErr(xerr.BadRequest("分站部署节点不可用"))
	}

	// ==================== 创建发布轮次 ====================

	// 生成本次发布请求编号
	requestNo := uuid.NewString()

	// 编码调度任务参数
	params, err := json.Marshal(struct {
		OperatorCode string `json:"operator_code"`
	}{
		OperatorCode: opCode,
	})
	if err != nil {
		l.Errorf("编码游戏资源分配发布任务参数失败: op_code=%s, error=%v", opCode, err)
		return nil, xerr.RpcErr(xerr.InternalServerError("编码任务参数失败"))
	}

	// ==================== 提交调度任务 ====================
	l.Infof("提交游戏资源分配发布调度任务: op_code=%s, request_no=%s, node_code=%s", opCode, requestNo, nodeData.Code)
	// 提交游戏资源分配发布调度任务
	taskResult, err := l.svcCtx.NodeDispatchDispatchRpc.SubmitTask(
		l.ctx,
		&nodedispatchdispatchpb.SubmitTaskRequest{
			RequestNo: requestNo,                             // 本次发布请求编号
			Target:    dispatchTargetOperatorGame,            // 目标服务
			TaskType:  dispatchTaskTypePublishGameAllocation, // 任务类型
			NodeCode:  nodeData.Code,                         // 执行节点编码
			Params:    string(params),                        // 任务参数
		},
	)
	if err != nil {
		l.Errorf("提交游戏资源分配发布任务失败: op_code=%s, request_no=%s, node_code=%s, error=%v",
			opCode, requestNo, nodeData.Code, err)
		return nil, err
	}

	// ==================== 提交结果校验 ====================

	// 调度任务结果必须完整
	if taskResult == nil {
		l.Errorf("提交游戏资源分配发布任务结果为空: op_code=%s, request_no=%s", opCode, requestNo)
		return nil, xerr.RpcErr(xerr.InternalServerError("任务提交失败"))
	}

	// 调度中心必须返回任务编号
	if strings.TrimSpace(taskResult.TaskNo) == "" {
		l.Errorf("提交游戏资源分配发布任务编号为空: op_code=%s, request_no=%s", opCode, requestNo)
		return nil, xerr.RpcErr(xerr.InternalServerError("任务编号为空"))
	}

	l.Infow("游戏资源分配发布任务已提交",
		logx.Field("op_code", opCode),
		logx.Field("request_no", requestNo),
		logx.Field("task_no", taskResult.TaskNo),
		logx.Field("node_code", nodeData.Code),
	)

	return &platform_game.PublishOperatorGameAllocationResponse{}, nil
}
