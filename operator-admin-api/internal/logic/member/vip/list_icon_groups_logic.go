// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package vip

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/vipicongrouppb"
	"oa.98ent.com/p9/platform-base/pkg/resourceurl"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/vipdefaulticonpb"
)

// ListIconGroupsLogic VIP图标组列表逻辑
type ListIconGroupsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListIconGroupsLogic 绑定查询请求的上下文和日志，此时不调用RPC。
func NewListIconGroupsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListIconGroupsLogic {
	return &ListIconGroupsLogic{
		Logger: logx.WithContext(ctx), // 日志关联当前查询请求。
		ctx:    ctx,                   // 透传登录身份和请求期限。
		svcCtx: svcCtx,                // 复用分站CRM及总站默认图客户端。
	}
}

// ListIconGroups 获取VIP图标组列表
func (l *ListIconGroupsLogic) ListIconGroups(req *types.ListVipIconGroupsRequest) (resp *types.ListVipIconGroupsResponse, err error) {
	// 调用CRM获取本站当前图标；CRM只查分站数据，不调用总站。
	rpcResp, err := l.svcCtx.MemberVipRpc.ListIconGroups(l.ctx, &vipicongrouppb.ListVipIconGroupsRequest{})
	if err != nil {
		// 复用图标接口的错误转换，超时和服务不可用分别返回504、503。
		return nil, iconRequestError(err)
	}

	// API直接调用总站获取实时默认图；每次查询最多等待3秒，不自动重试。
	if l.svcCtx.VipDefaultIconRpc == nil {
		return nil, iconRequestError(status.Error(codes.Unavailable, "common.service_unavailable"))
	}
	ctx, cancel := context.WithTimeout(l.ctx, 3*time.Second)
	defer cancel()
	defaults, err := l.svcCtx.VipDefaultIconRpc.ListAll(ctx, &vipdefaulticonpb.ListAllVipDefaultIconsRequest{})
	if err != nil {
		return nil, l.defaultIconError(err)
	}
	defaultURLs, err := defaultIconURLs(defaults.GetList())
	if err != nil {
		return nil, l.defaultIconError(err)
	}

	// 分站图标优先；分站为空时使用总站默认图。
	list := make([]types.VipIconGroupInfo, 0, len(rpcResp.GetList()))
	for _, group := range rpcResp.GetList() {
		defaultURL := defaultURLs[group.GetGroupNo()]
		iconURL := group.GetIconUrl()
		if iconURL == "" {
			iconURL = defaultURL
		}
		list = append(list, types.VipIconGroupInfo{
			Id:             group.GetId(),         // 本站记录ID，尚未建组时为0。
			GroupNo:        group.GetGroupNo(),    // 图标组编号
			StartLevel:     group.GetStartLevel(), // 起始VIP等级
			EndLevel:       group.GetEndLevel(),   // 结束VIP等级
			DefaultIconUrl: defaultURL,            // API查询总站返回的最新默认图片URL。
			IconUrl:        iconURL,               // 分站图片优先，否则使用总站默认图。
			UpdatedAt:      group.GetUpdatedAt(),  // 本站更新时间，Unix毫秒；尚未建组时为0。
		})
	}

	return &types.ListVipIconGroupsResponse{
		List: list, // VIP图标组列表
	}, nil
}

// defaultIconURLs 校验总站十组默认图，按图标组编号返回公开URL。
func defaultIconURLs(list []*vipdefaulticonpb.VipDefaultIconGroupInfo) (map[int64]string, error) {
	if len(list) != 10 {
		return nil, status.Error(codes.Unavailable, "common.service_unavailable")
	}
	result := make(map[int64]string, 10)
	// 核对等级范围、重复组和永久URL，防止默认图匹配到其他等级组。
	for _, item := range list {
		if item == nil || item.MinLevel < 0 || item.MinLevel > 90 ||
			item.MinLevel%10 != 0 || item.MaxLevel != item.MinLevel+9 {
			return nil, status.Error(codes.Unavailable, "common.service_unavailable")
		}
		groupNo := item.MinLevel / 10
		if _, exists := result[groupNo]; exists {
			return nil, status.Error(codes.Unavailable, "common.service_unavailable")
		}
		if item.IconUrl != "" {
			if _, err := resourceurl.PublicURL(item.IconUrl); err != nil {
				return nil, status.Error(codes.Unavailable, "common.service_unavailable")
			}
		}
		result[groupNo] = item.IconUrl // 按等级组编号匹配默认图，不依赖总站返回顺序。
	}
	return result, nil
}

// defaultIconError 记录总站查询或响应校验失败，保留超时语义，不向前端暴露下游内部错误。
func (l *ListIconGroupsLogic) defaultIconError(err error) error {
	operatorCode := ""
	if claims, ok := ctxdata.OperatorClaimsFromCtx(l.ctx); ok {
		operatorCode = claims.OperatorCode
	}
	l.Errorw("查询总站VIP默认图标失败",
		logx.Field("operator_code", operatorCode),
		logx.Field("error", err),
	)
	switch status.Code(err) {
	case codes.Canceled:
		return status.Error(codes.Canceled, "common.request_timeout")
	case codes.DeadlineExceeded:
		return iconRequestError(status.Error(codes.DeadlineExceeded, "common.request_timeout"))
	default:
		return iconRequestError(status.Error(codes.Unavailable, "common.service_unavailable"))
	}
}
