package viplogic

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/vipicongrouppb"
)

// UpdateIconGroupLogic VIP图标修改逻辑
type UpdateIconGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewUpdateIconGroupLogic 创建VIP图标修改逻辑
func NewUpdateIconGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateIconGroupLogic {
	return &UpdateIconGroupLogic{
		ctx:    ctx,                   // 本次RPC请求的身份及超时上下文。
		svcCtx: svcCtx,                // 复用本站图标DAO和公开前缀配置。
		Logger: logx.WithContext(ctx), // 日志关联本次保存请求。
	}
}

// UpdateIconGroup 保存本站当前图标Key或兼容URL，不上传、复制或删除OSS对象。
func (l *UpdateIconGroupLogic) UpdateIconGroup(in *vipicongrouppb.UpdateVipIconGroupRequest) (*vipicongrouppb.UpdateVipIconGroupResponse, error) {
	// 复用登录分站身份校验，本站图片目录中的站点编码必须与该身份一致。
	operatorCode, err := iconOperator(l.ctx)
	if err != nil {
		return nil, err
	}
	// 0组合法；图片不能为空，Key和完整URL继续使用原icon_url请求字段。
	iconURL := strings.TrimSpace(in.GetIconUrl())
	if in == nil || in.GetGroupNo() < 0 || in.GetGroupNo() > 9 || iconURL == "" {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}
	// 复用公开地址转换及目录校验：本站、本组Key直接入库，已知前缀URL转Key，旧外链兼容保存。
	stored, _, err := iconValue(iconURL, operatorCode, in.GetGroupNo(), l.svcCtx.Config.ResourceBaseURL)
	if err != nil {
		return nil, err
	}

	// 复用保存DAO，在同一事务内补齐本站十组空记录并保存目标图片；失败时复用数据库错误分类。
	if err := l.svcCtx.DAO.MemberVipIconGroup.Update(l.ctx, in.GetGroupNo(), stored); err != nil {
		return nil, iconDBError(err)
	}

	return &vipicongrouppb.UpdateVipIconGroupResponse{}, nil
}
