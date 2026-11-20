package taskservicelogic

import (
	"context"
	"time"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-activity/ent"
	"oa.98ent.com/p9/operator-activity/ent/membertaskreward"
	"oa.98ent.com/p9/operator-activity/internal/svc"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/taskpb"
	"oa.98ent.com/p9/operator-activity/pkg/i18n"

	"github.com/zeromicro/go-zero/core/logx"
)

type MemberTaskRewardListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMemberTaskRewardListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MemberTaskRewardListLogic {
	return &MemberTaskRewardListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// group: member_task_reward 会员任务奖励查询
func (l *MemberTaskRewardListLogic) MemberTaskRewardList(in *taskpb.ListMemberTaskRewardReq) (*taskpb.ListMemberTaskRewardResp, error) {
	// 验证操作员权限
	claims, ok := ctxdata.OperatorClaimsFromCtx(l.ctx)
	if !ok {
		return nil, xerr.RpcErr(xerr.Unauthorized(i18n.Unauthorized))
	}

	// 验证操作员代码一致
	if in.OperatorCode != claims.OperatorCode {
		l.Infof("starting member task reward list for operator: %s, operator code in claims: %v", in.OperatorCode, claims.OperatorCode)
		return nil, xerr.RpcErr(xerr.Unauthorized(i18n.Unauthorized))
	}

	// 设置分页参数
	page := in.GetPage()
	if page <= 0 {
		page = 1
	}
	pageSize := in.GetPageSize()
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// 构建查询条件
	query := l.svcCtx.DB.MemberTaskReward.Query().
		Where(membertaskreward.OperatorCode(claims.OperatorCode))

	// 添加会员 ID 过滤
	if in.MemberId != nil && *in.MemberId > 0 {
		query = query.Where(membertaskreward.MemberID(*in.MemberId))
	}

	// 添加任务编码过滤
	if in.TaskCode != nil && *in.TaskCode != "" {
		query = query.Where(membertaskreward.TaskCode(*in.TaskCode))
	}

	// 添加任务分类过滤
	if in.Category != nil && *in.Category != "" {
		query = query.Where(membertaskreward.Category(*in.Category))
	}

	// 添加订单状态过滤
	if in.Status != nil {
		query = query.Where(membertaskreward.Status(int(*in.Status)))
	}

	// 添加订单号过滤
	if in.OrderNo != nil && *in.OrderNo != "" {
		query = query.Where(membertaskreward.OrderNo(*in.OrderNo))
	}

	// 添加创建时间范围过滤
	if in.CreatedAtStart != nil && *in.CreatedAtStart > 0 {
		startTime := time.UnixMilli(*in.CreatedAtStart)
		query = query.Where(membertaskreward.CreatedAtGTE(startTime))
	}
	if in.CreatedAtEnd != nil && *in.CreatedAtEnd > 0 {
		endTime := time.UnixMilli(*in.CreatedAtEnd)
		query = query.Where(membertaskreward.CreatedAtLTE(endTime))
	}

	// 获取总数
	total, err := query.Count(l.ctx)
	if err != nil {
		l.Errorf("failed to count member task reward: %v", err)
		return nil, xerr.RpcErr(err)
	}

	// 按创建时间降序排序
	rewards, err := query.
		Order(ent.Desc(membertaskreward.FieldCreatedAt), ent.Desc(membertaskreward.FieldID)).
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		All(l.ctx)
	if err != nil {
		l.Errorf("failed to query member task reward: %v", err)
		return nil, xerr.RpcErr(err)
	}

	// 转换为 proto 消息
	list := make([]*taskpb.MemberTaskReward, 0, len(rewards))
	for _, reward := range rewards {
		payoutAt := int64(0)
		if reward.PayoutAt != nil {
			payoutAt = reward.PayoutAt.UnixMilli()
		}
		list = append(list, &taskpb.MemberTaskReward{
			Id:                  reward.ID,
			OperatorCode:        reward.OperatorCode,
			MemberId:            reward.MemberID,
			TaskCode:            reward.TaskCode,
			KafkaId:             reward.KafkaID,
			Category:            reward.Category,
			PeriodKey:           reward.PeriodKey,
			PayoutMode:          reward.PayoutMode,
			Amount:              reward.Amount,
			Status:              int32(reward.Status),
			OrderNo:             reward.OrderNo,
			WalletTransactionId: reward.WalletTransactionID,
			DeviceId:            reward.DeviceID,
			WalletId:            reward.WalletID,
			ClaimIp:             reward.ClaimIP,
			PayoutAt:            payoutAt,
			CreatedAt:           reward.CreatedAt.UnixMilli(),
			UpdatedAt:           reward.UpdatedAt.UnixMilli(),
		})
	}

	return &taskpb.ListMemberTaskRewardResp{
		List:     list,
		Total:    int64(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}
