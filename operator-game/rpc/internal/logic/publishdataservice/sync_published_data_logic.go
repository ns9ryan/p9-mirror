package publishdataservicelogic

import (
	"context"
	"fmt"

	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/core/rpc/coreclient"
	"oa.98ent.com/p9/operator-game/rpc/internal/svc"
	"oa.98ent.com/p9/operator-game/rpc/pb/operator_game"
	pg "oa.98ent.com/p9/platform-game/rpc/pb/platform_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncPublishedDataLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSyncPublishedDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncPublishedDataLogic {
	return &SyncPublishedDataLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 同步已发布的游戏数据到分站
func (l *SyncPublishedDataLogic) SyncPublishedData(in *operator_game.SyncPublishedDataRequest) (*operator_game.SyncPublishedDataResp, error) {
	opCode := in.OpCode
	if opCode == "" {
		return &operator_game.SyncPublishedDataResp{
			Success: false,
			Message: "分站代码不能为空",
		}, nil
	}

	resp := &operator_game.SyncPublishedDataResp{
		Success: true,
		Message: "同步成功",
	}
	l.Infof("开始同步已发布的游戏数据到分站: %s", opCode)

	// 1. 同步游戏分类
	categoryResp, err := l.syncPublishedGameCategory(opCode)
	if err != nil {
		l.Errorf("同步游戏分类失败: %v", err)
		categoryResp = &operator_game.SyncStatistic{
			Failed:       1,
			FailedReason: err.Error(),
		}
	}
	resp.CategoryStat = categoryResp

	// 2. 同步游戏供应商
	providerResp, err := l.syncPublishedGameProvider(opCode)
	if err != nil {
		l.Errorf("同步游戏供应商失败: %v", err)
		providerResp = &operator_game.SyncStatistic{
			Failed:       1,
			FailedReason: err.Error(),
		}
	}
	resp.ProviderStat = providerResp

	// 3. 同步游戏渠道
	channelResp, err := l.syncPublishedGameChannel(opCode)
	if err != nil {
		l.Errorf("同步游戏渠道失败: %v", err)
		channelResp = &operator_game.SyncStatistic{
			Failed:       1,
			FailedReason: err.Error(),
		}
	}
	resp.ChannelStat = channelResp

	// 4. 同步游戏
	gameResp, err := l.syncPublishedGame(opCode)
	if err != nil {
		l.Errorf("同步游戏失败: %v", err)
		gameResp = &operator_game.SyncStatistic{
			Failed:       1,
			FailedReason: err.Error(),
		}
	}
	resp.GameStat = gameResp

	// 同步多语言数据
	l.syncI18nNameMap()

	return resp, nil
}

// 同步游戏分类
func (l *SyncPublishedDataLogic) syncPublishedGameCategory(opCode string) (*operator_game.SyncStatistic, error) {
	stat := &operator_game.SyncStatistic{}

	// 调用platform-game RPC获取已发布的游戏分类
	pageSize := int32(1000)
	page := int32(1)
	var allItems []*pg.PublishedGameCategoryInfo

	for {
		req := &pg.GetPublishedGameCategoryListRequest{
			OpCode:   opCode,
			Page:     page,
			PageSize: pageSize,
		}

		categoryResp, err := l.svcCtx.PlatformGameGrpcClient.GetGameCategoryServiceClient().GetPublishedGameCategoryList(l.ctx, req)
		if err != nil {
			return stat, fmt.Errorf("获取已发布游戏分类列表失败: %w", err)
		}

		allItems = append(allItems, categoryResp.Items...)

		if int64(page*pageSize) >= categoryResp.Total {
			break
		}
		page++
	}

	stat.Total = int64(len(allItems))

	// 同步到数据库
	for _, item := range allItems {
		_, err := l.svcCtx.DAOManager.GameCategory.GetOrCreateGameCategory(l.ctx,
			item.CategoryCode, item.SortNo, int64(item.Status))
		if err != nil {
			l.Errorf("创建游戏分类失败: %v", err)
			stat.Failed++
			continue
		}
		stat.Success++
	}

	return stat, nil
}

// 同步游戏供应商
func (l *SyncPublishedDataLogic) syncPublishedGameProvider(opCode string) (*operator_game.SyncStatistic, error) {
	stat := &operator_game.SyncStatistic{}
	// 调用platform-game RPC获取已发布的游戏供应商
	pageSize := int32(1000)
	page := int32(1)
	var allItems []*pg.PublishedGameProviderInfo

	for {
		req := &pg.GetPublishedGameProviderListRequest{
			OpCode:   opCode,
			Page:     page,
			PageSize: pageSize,
		}

		providerResp, err := l.svcCtx.PlatformGameGrpcClient.GetGameProviderServiceClient().GetPublishedGameProviderList(l.ctx, req)
		if err != nil {
			return stat, fmt.Errorf("获取已发布游戏供应商列表失败: %w", err)
		}

		allItems = append(allItems, providerResp.Items...)

		if int64(page*pageSize) >= providerResp.Total {
			break
		}
		page++
	}

	stat.Total = int64(len(allItems))

	// 同步到数据库
	for _, item := range allItems {
		_, err := l.svcCtx.DAOManager.GameProvider.GetOrCreateGameProvider(l.ctx,
			item.ProviderCode, item.ChannelCode, &item.LogoUrl, item.SortNo, int64(item.Status))
		if err != nil {
			l.Errorf("创建游戏供应商失败: %v", err)
			stat.Failed++
			continue
		}
		stat.Success++
	}

	return stat, nil
}

// 同步游戏渠道
func (l *SyncPublishedDataLogic) syncPublishedGameChannel(opCode string) (*operator_game.SyncStatistic, error) {
	stat := &operator_game.SyncStatistic{}

	// 调用platform-game RPC获取已发布的游戏渠道
	pageSize := int32(1000)
	page := int32(1)
	var allItems []*pg.PublishedGameChannelInfo

	for {
		req := &pg.GetPublishedGameChannelListRequest{
			OpCode:   opCode,
			Page:     page,
			PageSize: pageSize,
		}

		channelResp, err := l.svcCtx.PlatformGameGrpcClient.GetGameChannelServiceClient().GetPublishedGameChannelList(l.ctx, req)
		if err != nil {
			return stat, fmt.Errorf("获取已发布游戏渠道列表失败: %w", err)
		}

		allItems = append(allItems, channelResp.Items...)

		if int64(page*pageSize) >= channelResp.Total {
			break
		}
		page++
	}

	stat.Total = int64(len(allItems))

	// 同步到数据库
	for _, item := range allItems {
		_, err := l.svcCtx.DAOManager.GameChannel.GetOrCreateGameChannel(l.ctx,
			1, item.ChannelCode, item.SortNo, item.LoadType, int64(item.Status))
		if err != nil {
			l.Errorf("创建游戏渠道失败: %v", err)
			stat.Failed++
			continue
		}
		stat.Success++
	}

	return stat, nil
}

// 同步游戏
func (l *SyncPublishedDataLogic) syncPublishedGame(opCode string) (*operator_game.SyncStatistic, error) {
	stat := &operator_game.SyncStatistic{}

	// 调用platform-game RPC获取已发布的游戏
	pageSize := int32(1000)
	page := int32(1)
	var allItems []*pg.PublishedGameInfo
	// 用于收集游戏相关的分类、渠道、厂商（使用 map 来去重）
	categoryMap := make(map[string]*pg.PublishedGameCategoryInfo)
	providerMap := make(map[string]*pg.PublishedGameProviderInfo)
	channelMap := make(map[string]*pg.PublishedGameChannelInfo)

	for {
		req := &pg.GetPublishedGameListRequest{
			OpCode:   opCode,
			Page:     page,
			PageSize: pageSize,
		}

		gameResp, err := l.svcCtx.PlatformGameGrpcClient.GetGameServiceClient().GetPublishedGameList(l.ctx, req)
		if err != nil {
			return stat, fmt.Errorf("获取已发布游戏列表失败: %w", err)
		}

		allItems = append(allItems, gameResp.GameItems...)

		// 收集分类项（使用类型码作为 key 来去重）
		for _, catItem := range gameResp.CategoryItems {
			categoryMap[catItem.CategoryCode] = catItem
		}

		// 收集厂商项（使用厂商码作为 key 来去重）
		for _, prvItem := range gameResp.ProviderItems {
			providerMap[prvItem.ProviderCode] = prvItem
		}

		// 收集渠道项（使用渠道码作为 key 来去重）
		for _, chnItem := range gameResp.ChannelItems {
			channelMap[chnItem.ChannelCode] = chnItem
		}

		if int64(page*pageSize) >= gameResp.Total {
			break
		}
		page++
	}

	stat.Total = int64(len(allItems))

	// 同步收集到的分类
	l.Infof("[游戏同步] 开始同步游戏相关的分类, 共 %d 项", len(categoryMap))
	for _, catItem := range categoryMap {
		_, err := l.svcCtx.DAOManager.GameCategory.GetOrCreateGameCategory(l.ctx,
			catItem.CategoryCode, catItem.SortNo, int64(catItem.Status))
		if err != nil {
			l.Errorf("[游戏同步] 同步游戏分类失败: %v", err)
			// 不中断游戏同步
		}
	}

	// 同步收集到的厂商
	l.Infof("[游戏同步] 开始同步游戏相关的厂商, 共 %d 项", len(providerMap))
	for _, prvItem := range providerMap {
		_, err := l.svcCtx.DAOManager.GameProvider.GetOrCreateGameProvider(l.ctx,
			prvItem.ProviderCode, prvItem.ChannelCode, &prvItem.LogoUrl, prvItem.SortNo, int64(prvItem.Status))
		if err != nil {
			l.Errorf("[游戏同步] 同步游戏厂商失败: %v", err)
			// 不中断游戏同步
		}
	}

	// 同步收集到的渠道
	l.Infof("[游戏同步] 开始同步游戏相关的渠道, 共 %d 项", len(channelMap))
	for _, chnItem := range channelMap {
		_, err := l.svcCtx.DAOManager.GameChannel.GetOrCreateGameChannel(l.ctx,
			1, chnItem.ChannelCode, chnItem.SortNo, chnItem.LoadType, int64(chnItem.Status))
		if err != nil {
			l.Errorf("[游戏同步] 同步游戏渠道失败: %v", err)
			// 不中断游戏同步
		}
	}

	// 同步游戏到数据库
	l.Infof("[游戏同步] 开始同步游戏, 共 %d 项", len(allItems))
	for _, item := range allItems {
		name := item.Name
		imageURL := item.ImageUrl

		_, err := l.svcCtx.DAOManager.Game.GetOrCreateGame(l.ctx,
			item.SourceId,
			item.GameCode, item.ProviderKey, item.CategoryCode, item.ProviderCode, item.ChannelCode,
			&name, &imageURL, item.SortNo, item.SupportsEmbed, item.SupportsRedirect, int64(item.Status), item.CurrencyCodeList)
		if err != nil {
			l.Errorf("[游戏同步] 创建游戏失败: %v", err)
			stat.Failed++
			continue
		}
		stat.Success++
	}

	return stat, nil
}

// 同步多语言名称映射
func (l *SyncPublishedDataLogic) syncI18nNameMap() {
	// 调用platform-game RPC获取多语言名称映射
	i18nResp, err := l.svcCtx.PlatformGameGrpcClient.GetGameSyncCheckpointServiceClient().
		GetI18NNameMap(l.ctx, &pg.GetI18NNameMapRequest{})
	if err != nil {
		l.Errorf("同步多语言名称映射失败: %v", err)
		return
	}

	if i18nResp.Code != 0 {
		l.Errorf("获取多语言名称映射失败: code=%d, message=%s", i18nResp.Code, i18nResp.Message)
		return
	}

	l.Infof("成功同步多语言名称映射, 共 %d 条记录", len(i18nResp.Data))

	// 记录多语言数据
	l.svcCtx.Core.RegisterCatalog(context.Background(), AppendGameI18nItems(i18nResp.Data))
	l.Infof("[API GameSyncCheckpointGet] i18n name map registered successfully")
}

func AppendGameI18nItems(nameMap map[string]string) *coreclient.RegisterCatalogReq {
	var out []*coreclient.I18NItem
	for key, value := range nameMap {
		out = append(out,
			&coreclient.I18NItem{
				I18NCode:  i18n.CodeOperator,
				I18NGroup: "game",
				TransKey:  key + ".name",
				Lang:      i18n.LangZH,
				Value:     value,
			},
		)
	}
	return &coreclient.RegisterCatalogReq{
		I18N: out,
	}
}
