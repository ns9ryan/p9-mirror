package vipdefaulticonservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/platform-base/pkg/resourceurl"
	"oa.98ent.com/p9/platform-base/rpc/ent"
	"oa.98ent.com/p9/platform-base/rpc/internal/enterror"
	"oa.98ent.com/p9/platform-base/rpc/internal/svc"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/vipdefaulticonpb"
)

type ListAllLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAllLogic {
	return &ListAllLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListAll 获取固定十组总网默认图标，分站通过该查询自行处理默认值回退。
func (l *ListAllLogic) ListAll(
	in *vipdefaulticonpb.ListAllVipDefaultIconsRequest,
) (*vipdefaulticonpb.ListAllVipDefaultIconsResponse, error) {
	// 只读取总网默认图标，不读取或修改分站自定义配置。
	rows, err := l.svcCtx.BaseData.VipDefaultIconGroup.All(l.ctx)
	if err != nil {
		// 复用 Ent 错误转换，避免直接向调用方暴露数据库内部错误。
		return nil, enterror.Handle(l.Logger, err)
	}
	// 在响应中补齐十组，并把数据库中的 Key 拼成完整公开地址。
	list, err := toIconGroupList(rows, l.svcCtx.Config.ResourceBaseURL)
	if err != nil {
		return nil, err
	}
	// 返回固定顺序的等级组列表，不处理分站自己的图片覆盖和默认值回退。
	return &vipdefaulticonpb.ListAllVipDefaultIconsResponse{
		List: list, // 固定十组，缺失组的图片地址为空。
	}, nil
}

const (
	vipDefaultIconGroupCount = 10 // 固定十组默认图标
	vipDefaultIconLevelStep  = 10 // 每组包含十个等级
)

// toIconGroupList 在响应中补齐缺失组，把已存 Key 拼成公开地址；查询不初始化或改写数据库。
func toIconGroupList(
	rows []*ent.VipDefaultIconGroup, baseURL string,
) ([]*vipdefaulticonpb.VipDefaultIconGroupInfo, error) {
	// 先建立 0–9 至 90–99 的十个空图片组，缺少配置也保持响应结构完整。
	list := make([]*vipdefaulticonpb.VipDefaultIconGroupInfo, vipDefaultIconGroupCount)
	for index := range list {
		minLevel := int64(index * vipDefaultIconLevelStep)
		list[index] = &vipdefaulticonpb.VipDefaultIconGroupInfo{
			MinLevel: minLevel,                               // 本组等级下限，从 0 级起每组递增 10 级。
			MaxLevel: minLevel + vipDefaultIconLevelStep - 1, // 本组等级上限，为下限加 9。
		}
	}
	// 只把有效等级组填回对应位置，不借查询修复或插入数据库记录。
	for _, row := range rows {
		if row == nil || row.MinLevel < 0 || row.MinLevel > 90 || row.MinLevel%vipDefaultIconLevelStep != 0 {
			continue
		}
		// 复用公开地址转换：Key 按配置拼接，旧完整 URL 保持兼容，空值仍为空。
		_, accessURL, err := resourceurl.Value(row.IconURL, baseURL)
		if err != nil {
			return nil, err
		}
		list[row.MinLevel/vipDefaultIconLevelStep].IconUrl = accessURL
	}
	return list, nil
}
