package manage

import (
	"context"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/taskpb"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMemberTaskRewardLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListMemberTaskRewardLogic 会员任务奖励订单查询
func NewListMemberTaskRewardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMemberTaskRewardLogic {
	return &ListMemberTaskRewardLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMemberTaskRewardLogic) ListMemberTaskReward(req *types.ListMemberTaskRewardReq) (resp *types.ListMemberTaskRewardResp, err error) {
	// 获取操作员信息
	claims := ctxdata.ClaimsFromCtx(l.ctx)
	operatorCode := claims.OperatorCode

	// 构建 RPC 请求
	rpcReq := &taskpb.ListMemberTaskRewardReq{
		OperatorCode:   operatorCode,
		MemberGlobalId: req.MemberGlobalId,
		MemberId:       req.MemberId,
		TaskCode:       req.TaskCode,
		Category:       req.Category,
		Status:         req.Status,
		OrderNo:        req.OrderNo,
		CreatedAtStart: req.CreatedAtStart,
		CreatedAtEnd:   req.CreatedAtEnd,
		Page:           &req.Page,
		PageSize:       &req.PageSize,
	}

	// 调用 RPC
	rpcResp, err := l.svcCtx.TaskServiceRpc.MemberTaskRewardList(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}

	// 转换响应
	list := make([]types.MemberTaskReward, 0, len(rpcResp.GetList()))
	for _, item := range rpcResp.GetList() {
		list = append(list, types.MemberTaskReward{
			Id:                  item.Id,
			OperatorCode:        item.OperatorCode,
			MemberGlobalId:      item.MemberGlobalId,
			MemberId:            item.MemberId,
			TaskCode:            item.TaskCode,
			KafkaId:             item.KafkaId,
			Category:            item.Category,
			PeriodKey:           item.PeriodKey,
			PayoutMode:          item.PayoutMode,
			Amount:              item.Amount,
			Status:              item.Status,
			OrderNo:             item.OrderNo,
			WalletTransactionId: item.WalletTransactionId,
			DeviceId:            item.DeviceId,
			WalletId:            item.WalletId,
			ClaimIp:             item.ClaimIp,
			PayoutAt:            item.PayoutAt,
			CreatedAt:           item.CreatedAt,
			UpdatedAt:           item.UpdatedAt,
		})
	}

	return &types.ListMemberTaskRewardResp{
		Total:    rpcResp.GetTotal(),
		List:     list,
		Page:     rpcResp.GetPage(),
		PageSize: rpcResp.GetPageSize(),
	}, nil
}
