package defaultavatarservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/platform-base/pkg/resourceurl"
	"oa.98ent.com/p9/platform-base/rpc/internal/enterror"
	"oa.98ent.com/p9/platform-base/rpc/internal/svc"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/defaultavatarpb"
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

// ListAll 读取总站默认头像，按主键稳定返回；不处理分站会员的选择或回退。
func (l *ListAllLogic) ListAll(
	in *defaultavatarpb.ListAllDefaultAvatarsRequest,
) (*defaultavatarpb.ListAllDefaultAvatarsResponse, error) {
	// 查询总站全部默认头像，按主键升序返回，不在这里补写缺失记录。
	rows, err := l.svcCtx.BaseData.DefaultAvatar.All(l.ctx)
	if err != nil {
		// 将数据库错误转换为项目统一的 gRPC 错误。
		return nil, enterror.Handle(l.Logger, err)
	}
	// 转换查询结果；没有记录时返回空数组，图片 Key 在转换过程中拼成公开 URL。
	list := make([]*defaultavatarpb.DefaultAvatarInfo, 0, len(rows))
	for _, row := range rows {
		// 复用公开地址拼接，返回可直接展示的 URL；读取不会改写原有图片记录。
		_, accessURL, err := resourceurl.Value(row.ImageURL, l.svcCtx.Config.ResourceBaseURL)
		if err != nil {
			return nil, err
		}
		list = append(list, &defaultavatarpb.DefaultAvatarInfo{
			Id:       row.ID,    // 头像记录主键，修改接口使用该值定位。
			ImageUrl: accessURL, // 由 Key 拼出的公开地址，未配置时为空串。
		})
	}
	// 返回素材主键和可展示的图片地址，具体选用哪张由分站处理。
	return &defaultavatarpb.ListAllDefaultAvatarsResponse{
		List: list, // 按主键升序返回，空列表保持为 []。
	}, nil
}
