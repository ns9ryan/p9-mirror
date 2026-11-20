package viplogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/vipicongrouppb"
)

// ListIconGroupsLogic VIP图标组列表逻辑
type ListIconGroupsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewListIconGroupsLogic 创建VIP图标组列表逻辑
func NewListIconGroupsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListIconGroupsLogic {
	return &ListIconGroupsLogic{
		ctx:    ctx,                   // 本次RPC请求的身份及超时上下文。
		svcCtx: svcCtx,                // 复用本站数据库访问。
		Logger: logx.WithContext(ctx), // 日志关联本次请求。
	}
}

// ListIconGroups 查询本站当前图标，不调用其他服务或保存默认资料。
func (l *ListIconGroupsLogic) ListIconGroups(in *vipicongrouppb.ListVipIconGroupsRequest) (*vipicongrouppb.ListVipIconGroupsResponse, error) {
	// 复用身份校验，只按登录分站查询，不接收前端传入的站点编码或域名。
	operatorCode, err := iconOperator(l.ctx)
	if err != nil {
		return nil, err
	}
	// 保留原有本站图标查询和排序；查询隔离继续由Ent站点拦截器执行，此处不写库。
	groups, err := l.svcCtx.DAO.MemberVipIconGroup.List(l.ctx)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	// 固定返回十组本站数据；默认图和跨服务结果由API层补充，查询不补建记录。
	list, err := toIconGroupList(groups, operatorCode, l.svcCtx.Config.ResourceBaseURL)
	if err != nil {
		return nil, err
	}

	return &vipicongrouppb.ListVipIconGroupsResponse{
		List: list, // VIP图标组列表
	}, nil
}

// toIconGroupList 仅组装本站查询响应；缺少本站记录时图标字段为空，ID和更新时间为0。
func toIconGroupList(groups []*ent.MemberVipIconGroup, operatorCode, baseURL string) ([]*vipicongrouppb.VipIconGroupInfo, error) {
	list := make([]*vipicongrouppb.VipIconGroupInfo, 10)
	for index := range list {
		groupNo := int64(index)
		list[index] = &vipicongrouppb.VipIconGroupInfo{
			GroupNo:        groupNo,        // 图标组编号，0为第一组。
			StartLevel:     groupNo * 10,   // 本组起始VIP等级。
			EndLevel:       groupNo*10 + 9, // 本组结束VIP等级。
			DefaultIconUrl: "",             // 默认图由API层调用总站后补充。
			IconUrl:        "",             // 本站未保存时为空，由API层补默认图。
		}
	}
	// 保留本站已有图片及审计信息；复用Key拼接与归属校验，不重写旧URL或识别旧初始化图。
	for _, group := range groups {
		if group.GroupNo < 0 || group.GroupNo >= int64(len(list)) {
			continue
		}
		_, accessURL, err := iconValue(group.IconURL, operatorCode, group.GroupNo, baseURL)
		if err != nil {
			return nil, err
		}
		item := list[group.GroupNo]
		item.Id = group.ID                           // 本站已保存记录的ID。
		item.UpdatedAt = group.UpdatedAt.UnixMilli() // 本站记录更新时间，单位毫秒。
		item.IconUrl = accessURL                     // 本站图片完整URL；空串由API补默认图，不写库。
	}
	return list, nil
}
