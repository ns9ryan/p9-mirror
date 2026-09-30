package gamechannelservicelogic

import (
	"context"
	"fmt"

	"oa.98ent.com/p9/platform-game/rpc/internal/svc"
	"oa.98ent.com/p9/platform-game/rpc/pb/platform_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPublishedGameChannelListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPublishedGameChannelListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPublishedGameChannelListLogic {
	return &GetPublishedGameChannelListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取已发布的游戏渠道列表（供分站同步）
func (l *GetPublishedGameChannelListLogic) GetPublishedGameChannelList(in *platform_game.GetPublishedGameChannelListRequest) (*platform_game.GetPublishedGameChannelListResp, error) {
	// 验证请求参数
	if in.OpCode == "" {
		return nil, fmt.Errorf("op_code is required")
	}

	if in.Page <= 0 {
		in.Page = 1
	}
	if in.PageSize <= 0 {
		in.PageSize = 10
	}

	offset := (in.Page - 1) * in.PageSize
	limit := in.PageSize

	// 获取分站对应的所有渠道代码
	channelCodes, err := l.svcCtx.DAOManager.OperatorGameChannel.GetChannelCodesByOpCode(l.ctx, in.OpCode)
	if err != nil {
		l.Errorf("GetChannelCodesByOpCode failed: %v", err)
		return nil, err
	}

	// 获取渠道信息
	channels, total, err := l.svcCtx.DAOManager.GameChannel.GetPublishedGameChannelList(l.ctx, channelCodes, int64(offset), int64(limit))
	if err != nil {
		l.Errorf("GetPublishedGameChannelList failed: %v", err)
		return nil, err
	}

	// 转换为proto消息
	items := make([]*platform_game.PublishedGameChannelInfo, 0, len(channels))
	for _, channel := range channels {
		items = append(items, &platform_game.PublishedGameChannelInfo{
			Id:          channel.ID,
			ChannelCode: channel.SourceChannelCode,
			SortNo:      channel.SortNo,
			LoadType:    channel.LoadType,
			Status:      int32(channel.Status),
			CreatedAt:   channel.CreatedAt.Unix() * 1000, // 转换为毫秒时间戳
			UpdatedAt:   channel.UpdatedAt.Unix() * 1000,
		})
	}

	return &platform_game.GetPublishedGameChannelListResp{
		Items:    items,
		Total:    int64(total),
		Page:     in.Page,
		PageSize: in.PageSize,
	}, nil
}
