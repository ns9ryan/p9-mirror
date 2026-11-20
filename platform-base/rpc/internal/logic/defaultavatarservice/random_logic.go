package defaultavatarservicelogic

import (
	"context"
	"math/rand/v2"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/platform-base/pkg/i18nkey"
	"oa.98ent.com/p9/platform-base/pkg/resourceurl"
	"oa.98ent.com/p9/platform-base/pkg/rpc/grpcerror"
	"oa.98ent.com/p9/platform-base/rpc/internal/enterror"
	"oa.98ent.com/p9/platform-base/rpc/internal/svc"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/defaultavatarpb"
)

type RandomLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRandomLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RandomLogic {
	return &RandomLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Random 为分站随机选取新会员初始头像，不写素材表或会员资料，不调用 OSS。
func (l *RandomLogic) Random(
	in *defaultavatarpb.RandomDefaultAvatarRequest,
) (*defaultavatarpb.RandomDefaultAvatarResponse, error) {
	// 固定八个槽位数据量很小，一次读取，避免先计数再分页时被并发修改影响。
	rows, err := l.svcCtx.DB.DefaultAvatar.
		Query().
		All(l.ctx)
	if err != nil {
		return nil, enterror.Handle(l.Logger, err)
	}
	// 未配置图片的槽位不参与抽取；没有可用头像时明确返回 NotFound。
	configured := rows[:0]
	for _, row := range rows {
		if strings.TrimSpace(row.ImageURL) != "" {
			configured = append(configured, row)
		}
	}
	if len(configured) == 0 {
		return nil, grpcerror.NotFound(i18nkey.DataNotFound)
	}
	// 每条已配置记录机会相同，允许不同会员选中同一张；配置错误不能伪装成成功。
	selected := configured[rand.IntN(len(configured))]
	_, accessURL, err := resourceurl.Value(selected.ImageURL, l.svcCtx.Config.ResourceBaseURL)
	if err != nil {
		return nil, err
	}
	// 复用原有头像资料结构和 Key 拼接规则，分站在创建会员时保存本次选择。
	return &defaultavatarpb.RandomDefaultAvatarResponse{
		Avatar: &defaultavatarpb.DefaultAvatarInfo{ // 本次随机选中的一张已配置头像。
			Id:       selected.ID, // 选中头像的记录主键，不是会员 ID。
			ImageUrl: accessURL,   // 可直接展示的公开地址，不写入会员资料。
		},
	}, nil
}
