package viplogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/vipicongrouppb"
)

// ResetIconGroupLogic VIP图标恢复默认逻辑
type ResetIconGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewResetIconGroupLogic 创建VIP图标恢复默认逻辑
func NewResetIconGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetIconGroupLogic {
	return &ResetIconGroupLogic{
		ctx:    ctx,                   // 本次RPC请求的身份及超时上下文。
		svcCtx: svcCtx,                // 复用本站图标保存DAO。
		Logger: logx.WithContext(ctx), // 日志关联本次恢复请求。
	}
}

// ResetIconGroup 清空本站指定组的图片配置，后续查询使用总站实时默认图。
func (l *ResetIconGroupLogic) ResetIconGroup(in *vipicongrouppb.ResetVipIconGroupRequest) (*vipicongrouppb.ResetVipIconGroupResponse, error) {
	// 恢复归属取登录身份，拒绝总站、代理和预览身份操作本站图标。
	if _, err := iconOperator(l.ctx); err != nil {
		return nil, err
	}
	// 0是合法第一组，允许范围为0~9；请求缺失或越界时不写库。
	if in == nil || in.GetGroupNo() < 0 || in.GetGroupNo() > 9 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}
	// 复用保存事务及错误分类，缺组时补齐十组，仅将目标组图片保存为空；失败时回滚。
	// 恢复不调用总站RPC、不要求公开前缀，也不上传或删除OSS对象；默认图由后续查询获取。
	if err := l.svcCtx.DAO.MemberVipIconGroup.Update(l.ctx, in.GetGroupNo(), ""); err != nil {
		return nil, iconDBError(err)
	}

	return &vipicongrouppb.ResetVipIconGroupResponse{}, nil
}
