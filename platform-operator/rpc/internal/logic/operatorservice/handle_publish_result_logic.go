package operatorservicelogic

import (
	"context"
	"strings"
	"time"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/platform-operator/pkg/i18nkey"
	"oa.98ent.com/p9/platform-operator/rpc/ent/operator"
	"oa.98ent.com/p9/platform-operator/rpc/internal/enterror"
	"oa.98ent.com/p9/platform-operator/rpc/internal/svc"
	"oa.98ent.com/p9/platform-operator/rpc/pb/platformoperatorrpc/operatorpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandlePublishResultLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHandlePublishResultLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandlePublishResultLogic {
	return &HandlePublishResultLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// HandlePublishResult 处理分站发布结果
func (l *HandlePublishResultLogic) HandlePublishResult(in *operatorpb.HandlePublishResultRequest) (*operatorpb.HandlePublishResultResponse, error) {
	// 整理请求参数
	taskNo := strings.TrimSpace(in.TaskNo)

	// 校验请求参数
	if taskNo == "" {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
	}
	if in.Status != 3 && in.Status != 4 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
	}
	if in.FinishedAt <= 0 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
	}

	// 根据发布任务编号获取分站
	current, err := l.svcCtx.DB.Operator.
		Query().
		Where(operator.PublishTaskNoEQ(taskNo)).
		Only(l.ctx)
	if err != nil {
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// 相同结果重复回调时直接返回
	if current.PublishStatus == in.Status {
		return &operatorpb.HandlePublishResultResponse{
			OperatorId:    current.ID,            // 分站ID
			OperatorCode:  current.Code,          // 分站业务编码
			PublishStatus: current.PublishStatus, // 发布状态
		}, nil
	}

	// 只有发布中的分站允许接收最终发布结果
	if current.PublishStatus != 2 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ConstraintError))
	}

	// 创建发布结果更新
	update := current.
		Update().
		SetPublishStatus(in.Status) // 发布状态: 3已发布, 4发布失败

	// 首次发布成功时记录实际任务完成时间
	if in.Status == 3 && current.PublishedAt == nil {
		update.SetPublishedAt(time.UnixMilli(in.FinishedAt))
	}

	// 保存发布结果
	if err = update.Exec(l.ctx); err != nil {
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// 返回发布结果
	return &operatorpb.HandlePublishResultResponse{
		OperatorId:    current.ID,   // 分站ID
		OperatorCode:  current.Code, // 分站业务编码
		PublishStatus: in.Status,    // 发布状态: 3已发布, 4发布失败
	}, nil
}
