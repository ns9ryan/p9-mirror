package operatorservicelogic

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
	"oa.98ent.com/p9/platform-operator/rpc/ent"
	"oa.98ent.com/p9/platform-operator/rpc/ent/operator"
	"oa.98ent.com/p9/platform-operator/rpc/internal/enterror"
	"oa.98ent.com/p9/platform-operator/rpc/internal/svc"
	"oa.98ent.com/p9/platform-operator/rpc/pb/platformoperatorrpc/operatorpb"
)

const (
	dispatchTargetOperatorBase     = "operator-base"   // 分站基础服务
	dispatchTaskTypeCreateOperator = "CREATE_OPERATOR" // 创建分站任务
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

// Publish 发布分站
func (l *PublishLogic) Publish(in *operatorpb.PublishOperatorRequest) (*operatorpb.PublishOperatorResponse, error) {
	// ==================== 发布前校验 ====================

	// 分站ID必须大于0
	if in.Id <= 0 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
	}

	// 获取当前分站
	current, err := l.svcCtx.DB.Operator.Get(l.ctx, in.Id)
	if err != nil {
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// 只有已完成创建且尚未首次发布成功的分站可以发布
	if current.CreationStatus != 2 || current.PublishedAt != nil {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ConstraintError))
	}

	// 只有未发布或发布失败的分站可以发起发布
	if current.PublishStatus != 1 && current.PublishStatus != 4 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ConstraintError))
	}

	// ==================== 部署节点校验 ====================

	// 获取分站部署节点
	operatorNodeResult, err := l.svcCtx.NodeDispatchOperatorNodeRpc.Get(
		l.ctx,
		&nodedispatchoperatornodepb.GetOperatorNodeRequest{
			OperatorCode: current.Code, // 分站全局唯一业务编码
		},
	)
	if err != nil {
		// 未选择部署节点时不能发布
		if status.Code(err) == codes.NotFound {
			return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
		}

		l.Logger.Errorw(
			"获取分站部署节点失败",
			logx.Field("operator_id", current.ID),
			logx.Field("operator_code", current.Code),
			logx.Field("error", err.Error()),
		)
		return nil, err
	}

	// 部署节点信息必须完整
	if operatorNodeResult == nil || operatorNodeResult.OperatorNode == nil || operatorNodeResult.OperatorNode.Node == nil {
		l.Logger.Errorw(
			"分站部署节点信息为空",
			logx.Field("operator_id", current.ID),
			logx.Field("operator_code", current.Code),
		)
		return nil, xerr.RpcErr(xerr.InternalServerError(i18nkey.InternalError))
	}

	nodeData := operatorNodeResult.OperatorNode.Node

	// 发布时部署节点必须启用并在线
	if nodeData.Status != 1 || !nodeData.Online {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
	}

	// ==================== 创建发布轮次 ====================

	// 生成本次发布请求编号
	requestNo := uuid.NewString()

	// 编码调度任务参数
	params, err := json.Marshal(struct {
		OperatorCode string `json:"operator_code"`
	}{
		OperatorCode: current.Code,
	})
	if err != nil {
		l.Logger.Errorw(
			"编码分站发布任务参数失败",
			logx.Field("operator_id", current.ID),
			logx.Field("operator_code", current.Code),
			logx.Field("error", err.Error()),
		)
		return nil, xerr.RpcErr(xerr.InternalServerError(i18nkey.InternalError))
	}

	// 原子进入发布中状态，防止并发重复发布
	affected, err := l.svcCtx.DB.Operator.
		Update().
		Where(
			operator.IDEQ(current.ID),
			operator.CreationStatusEQ(2),
			operator.PublishStatusIn(1, 4),
		).
		SetPublishStatus(2).            // 发布状态: 2发布中
		SetPublishRequestNo(requestNo). // 当前发布请求编号
		ClearPublishTaskNo().           // 清除上一轮发布任务编号
		Save(l.ctx)
	if err != nil {
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}
	if affected != 1 {
		// 分站状态已被其他发布请求修改
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ConstraintError))
	}

	// 同步当前内存中的发布状态
	current.PublishStatus = 2
	current.PublishRequestNo = new(requestNo)
	current.PublishTaskNo = nil

	// ==================== 提交调度任务 ====================

	// 提交分站创建调度任务
	taskResult, err := l.svcCtx.NodeDispatchDispatchRpc.SubmitTask(
		l.ctx,
		&nodedispatchdispatchpb.SubmitTaskRequest{
			RequestNo: requestNo,                      // 本次发布请求编号
			Target:    dispatchTargetOperatorBase,     // 目标服务
			TaskType:  dispatchTaskTypeCreateOperator, // 任务类型
			NodeCode:  nodeData.Code,                  // 执行节点编码
			Params:    string(params),                 // 任务参数
		},
	)
	if err != nil {
		submitErr := err

		l.Logger.Errorw(
			"提交分站发布任务失败",
			logx.Field("operator_id", current.ID),
			logx.Field("operator_code", current.Code),
			logx.Field("request_no", requestNo),
			logx.Field("node_code", nodeData.Code),
			logx.Field("error", submitErr.Error()),
		)

		// 尝试根据请求编号恢复实际发布状态
		if recoveryErr := l.recoverPublishTask(current, requestNo); recoveryErr != nil {
			l.Logger.Errorw(
				"恢复分站发布状态失败",
				logx.Field("operator_id", current.ID),
				logx.Field("operator_code", current.Code),
				logx.Field("request_no", requestNo),
				logx.Field("error", recoveryErr.Error()),
			)
		}

		return nil, submitErr
	}

	// ==================== 提交结果校验 ====================

	// 调度任务结果必须完整
	if taskResult == nil {
		l.Logger.Errorw(
			"提交分站发布任务结果为空",
			logx.Field("operator_id", current.ID),
			logx.Field("operator_code", current.Code),
			logx.Field("request_no", requestNo),
		)
		return nil, xerr.RpcErr(xerr.InternalServerError(i18nkey.InternalError))
	}

	// 调度中心必须返回任务编号
	if strings.TrimSpace(taskResult.TaskNo) == "" {
		l.Logger.Errorw(
			"提交分站发布任务编号为空",
			logx.Field("operator_id", current.ID),
			logx.Field("operator_code", current.Code),
			logx.Field("request_no", requestNo),
		)
		return nil, xerr.RpcErr(xerr.InternalServerError(i18nkey.InternalError))
	}

	// task_no由任务结果回调或同步发布状态时落库
	return &operatorpb.PublishOperatorResponse{}, nil
}

// recoverPublishTask 恢复提交失败后的分站发布状态
func (l *PublishLogic) recoverPublishTask(current *ent.Operator, requestNo string) error {
	// 根据发布请求编号确认调度任务是否实际创建
	recoveryResult, err := l.svcCtx.NodeDispatchDispatchRpc.GetTask(
		l.ctx,
		&nodedispatchdispatchpb.GetTaskRequest{
			RequestNo: new(requestNo), // 本次发布请求编号
		},
	)
	if err != nil {
		// 明确不存在调度任务时标记本轮发布失败
		if status.Code(err) == codes.NotFound {
			_, updateErr := l.svcCtx.DB.Operator.
				Update().
				Where(
					operator.IDEQ(current.ID),
					operator.PublishRequestNoEQ(requestNo),
					operator.PublishStatusEQ(2),
					operator.PublishTaskNoIsNil(),
				).
				SetPublishStatus(4). // 发布状态: 4发布失败
				Save(l.ctx)
			if updateErr != nil {
				return updateErr
			}

			return nil
		}

		// 无法确认任务实际状态时保持发布中
		return err
	}

	// 调度任务信息必须完整
	if recoveryResult == nil || recoveryResult.Task == nil {
		return xerr.RpcErr(xerr.InternalServerError(i18nkey.InternalError))
	}

	// 校验实际创建的发布任务
	taskNo, _, err := validatePublishTask(
		recoveryResult.Task,
		"",
		requestNo,
	)
	if err != nil {
		return err
	}

	// 根据任务状态恢复分站发布状态
	switch recoveryResult.Task.Status {
	case 1, 2:
		// 任务已经创建并处于执行流程中，绑定实际任务编号
		_, _, err = bindPublishTask(
			l.ctx,
			l.svcCtx,
			l.Logger,
			current,
			taskNo,
			requestNo,
		)

	case 3, 4:
		// 任务已经结束，直接应用最终发布结果
		_, _, err = applyFinalPublishTask(
			l.ctx,
			l.svcCtx,
			l.Logger,
			current,
			recoveryResult.Task,
		)

	default:
		return xerr.RpcErr(xerr.BadRequest(i18nkey.ConstraintError))
	}

	return err
}
