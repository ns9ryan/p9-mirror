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

	l.Infof("[同步总结] 分站代码: %s, 分类(成功:%d, 失败:%d), 供应商(成功:%d, 失败:%d), 渠道(成功:%d, 失败:%d), 游戏(成功:%d, 失败:%d)",
		opCode,
		resp.CategoryStat.Success, resp.CategoryStat.Failed,
		resp.ProviderStat.Success, resp.ProviderStat.Failed,
		resp.ChannelStat.Success, resp.ChannelStat.Failed,
		resp.GameStat.Success, resp.GameStat.Failed,
	)

	return resp, nil
}

// 同步游戏分类
func (l *SyncPublishedDataLogic) syncPublishedGameCategory(opCode string) (*operator_game.SyncStatistic, error) {
	stat := &operator_game.SyncStatistic{}
	l.Infof("[分类同步] 开始同步分站 %s 的游戏分类", opCode)

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
			l.Errorf("[分类同步] 获取分类列表失败: %v", err)
			return stat, fmt.Errorf("获取已发布游戏分类列表失败: %w", err)
		}

		l.Debugf("[分类同步] 第%d页获取分类 %d 条, 总计 %d 条", page, len(categoryResp.Items), categoryResp.Total)
		allItems = append(allItems, categoryResp.Items...)

		if int64(page*pageSize) >= categoryResp.Total {
			break
		}
		page++
	}

	stat.Total = int64(len(allItems))
	l.Infof("[分类同步] 共获取 %d 个分类", stat.Total)

	// 收集返回的分类编码
	publishedCategoryCodeSet := make(map[string]bool)
	for _, item := range allItems {
		publishedCategoryCodeSet[item.CategoryCode] = true
	}
	l.Debugf("[分类同步] 已发布的分类编码集合: %v", publishedCategoryCodeSet)

	// 同步到数据库
	l.Infof("[分类同步] 开始同步 %d 个分类到数据库", len(allItems))
	for _, item := range allItems {
		_, err := l.svcCtx.DAOManager.GameCategory.GetOrUpdateGameCategory(l.ctx,
			opCode, item.CategoryCode, item.SortNo, int64(item.Status))
		if err != nil {
			l.Errorf("[分类同步] 同步分类失败: code=%s, err=%v", item.CategoryCode, err)
			stat.Failed++
			continue
		}
		l.Debugf("[分类同步] 成功同步分类: code=%s", item.CategoryCode)
		stat.Success++
	}

	// 删除不在发布列表中的分类
	l.Infof("[分类同步] 准备删除不存在的分类")
	deleteCount, err := l.svcCtx.DAOManager.GameCategory.DeleteGameCategoryNotIn(l.ctx, opCode, publishedCategoryCodeSet)
	if err != nil {
		l.Errorf("[分类同步] 删除过期分类失败: %v", err)
	} else if deleteCount > 0 {
		l.Infof("[分类同步] 已删除 %d 个过期分类", deleteCount)
	}

	l.Infof("[分类同步] 完成, 成功: %d, 失败: %d, 已删除: %d", stat.Success, stat.Failed, deleteCount)

	return stat, nil
}

// 同步游戏供应商
func (l *SyncPublishedDataLogic) syncPublishedGameProvider(opCode string) (*operator_game.SyncStatistic, error) {
	stat := &operator_game.SyncStatistic{}
	l.Infof("[供应商同步] 开始同步分站 %s 的游戏供应商", opCode)
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
			l.Errorf("[供应商同步] 获取供应商列表失败: %v", err)
			return stat, fmt.Errorf("获取已发布游戏供应商列表失败: %w", err)
		}

		l.Debugf("[供应商同步] 第%d页获取供应商 %d 条, 总计 %d 条", page, len(providerResp.Items), providerResp.Total)
		allItems = append(allItems, providerResp.Items...)

		if int64(page*pageSize) >= providerResp.Total {
			break
		}
		page++
	}

	stat.Total = int64(len(allItems))
	l.Infof("[供应商同步] 共获取 %d 个供应商", stat.Total)

	// 收集返回的供应商编码
	publishedProviderCodeSet := make(map[string]bool)
	for _, item := range allItems {
		publishedProviderCodeSet[item.ProviderCode] = true
	}
	l.Debugf("[供应商同步] 已发布的供应商编码集合: %v", publishedProviderCodeSet)

	// 同步到数据库
	l.Infof("[供应商同步] 开始同步 %d 个供应商到数据库", len(allItems))
	for _, item := range allItems {
		_, err := l.svcCtx.DAOManager.GameProvider.GetOrUpdateGameProvider(l.ctx,
			opCode, item.ProviderCode, item.ChannelCode, &item.LogoUrl, item.SortNo, int64(item.Status))
		if err != nil {
			l.Errorf("[供应商同步] 同步供应商失败: code=%s, err=%v", item.ProviderCode, err)
			stat.Failed++
			continue
		}
		l.Debugf("[供应商同步] 成功同步供应商: code=%s", item.ProviderCode)
		stat.Success++
	}

	// 删除不在发布列表中的供应商
	l.Infof("[供应商同步] 准备删除不存在的供应商")
	deleteCount, err := l.svcCtx.DAOManager.GameProvider.DeleteGameProviderNotIn(l.ctx, opCode, publishedProviderCodeSet)
	if err != nil {
		l.Errorf("[供应商同步] 删除过期供应商失败: %v", err)
	} else if deleteCount > 0 {
		l.Infof("[供应商同步] 已删除 %d 个过期供应商", deleteCount)
	}

	l.Infof("[供应商同步] 完成, 成功: %d, 失败: %d, 已删除: %d", stat.Success, stat.Failed, deleteCount)

	return stat, nil
}

// 同步游戏渠道
func (l *SyncPublishedDataLogic) syncPublishedGameChannel(opCode string) (*operator_game.SyncStatistic, error) {
	stat := &operator_game.SyncStatistic{}
	l.Infof("[渠道同步] 开始同步分站 %s 的游戏渠道", opCode)

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
			l.Errorf("[渠道同步] 获取渠道列表失败: %v", err)
			return stat, fmt.Errorf("获取已发布游戏渠道列表失败: %w", err)
		}

		l.Debugf("[渠道同步] 第%d页获取渠道 %d 条, 总计 %d 条", page, len(channelResp.Items), channelResp.Total)
		allItems = append(allItems, channelResp.Items...)

		if int64(page*pageSize) >= channelResp.Total {
			break
		}
		page++
	}

	stat.Total = int64(len(allItems))
	l.Infof("[渠道同步] 共获取 %d 个渠道", stat.Total)

	// 收集返回的渠道编码
	publishedChannelCodeSet := make(map[string]bool)
	for _, item := range allItems {
		publishedChannelCodeSet[item.ChannelCode] = true
	}
	l.Debugf("[渠道同步] 已发布的渠道编码集合: %v", publishedChannelCodeSet)

	// 同步到数据库
	l.Infof("[渠道同步] 开始同步 %d 个渠道到数据库", len(allItems))
	for _, item := range allItems {
		_, err := l.svcCtx.DAOManager.GameChannel.GetOrUpdateGameChannel(l.ctx,
			opCode, item.ChannelCode, item.SortNo, item.LoadType, int64(item.Status))
		if err != nil {
			l.Errorf("[渠道同步] 同步渠道失败: code=%s, err=%v", item.ChannelCode, err)
			stat.Failed++
			continue
		}
		l.Debugf("[渠道同步] 成功同步渠道: code=%s", item.ChannelCode)
		stat.Success++
	}

	// 删除不在发布列表中的渠道
	l.Infof("[渠道同步] 准备删除不存在的渠道")
	deleteCount, err := l.svcCtx.DAOManager.GameChannel.DeleteGameChannelNotIn(l.ctx, opCode, publishedChannelCodeSet)
	if err != nil {
		l.Errorf("[渠道同步] 删除过期渠道失败: %v", err)
	} else if deleteCount > 0 {
		l.Infof("[渠道同步] 已删除 %d 个过期渠道", deleteCount)
	}

	l.Infof("[渠道同步] 完成, 成功: %d, 失败: %d, 已删除: %d", stat.Success, stat.Failed, deleteCount)

	return stat, nil
}

// 同步游戏
func (l *SyncPublishedDataLogic) syncPublishedGame(opCode string) (*operator_game.SyncStatistic, error) {
	stat := &operator_game.SyncStatistic{}
	l.Infof("[游戏同步] 开始同步分站 %s 的游戏", opCode)

	// 调用platform-game RPC获取已发布的游戏
	pageSize := int32(1000)
	page := int32(1)
	var allItems []*pg.PublishedGameInfo

	for {
		req := &pg.GetPublishedGameListRequest{
			OpCode:   opCode,
			Page:     page,
			PageSize: pageSize,
		}

		gameResp, err := l.svcCtx.PlatformGameGrpcClient.GetGameServiceClient().GetPublishedGameList(l.ctx, req)
		if err != nil {
			l.Errorf("[游戏同步] 获取游戏列表失败: %v", err)
			return stat, fmt.Errorf("获取已发布游戏列表失败: %w", err)
		}

		l.Debugf("[游戏同步] 第%d页获取游戏 %d 条, 总计 %d 条", page, len(gameResp.GameItems), gameResp.Total)

		allItems = append(allItems, gameResp.GameItems...)
		if int64(page*pageSize) >= gameResp.Total {
			break
		}
		page++
	}

	stat.Total = int64(len(allItems))
	l.Infof("[游戏同步] 共获取 %d 个游戏", stat.Total)

	// 同步游戏到数据库
	l.Infof("[游戏同步] 开始同步游戏, 共 %d 项", len(allItems))

	// 收集返回的游戏编码
	publishedGameCodeSet := make(map[string]bool)
	for _, item := range allItems {
		publishedGameCodeSet[item.GameCode] = true
	}
	l.Debugf("[游戏同步] 已发布的游戏编码集合(前100个): %v", func() []string {
		var codes []string
		for k := range publishedGameCodeSet {
			codes = append(codes, k)
			if len(codes) >= 100 {
				break
			}
		}
		return codes
	}())

	l.Infof("[游戏同步] 开始同步 %d 个游戏到数据库", len(allItems))
	for _, item := range allItems {
		name := item.Name
		imageURL := item.ImageUrl

		_, err := l.svcCtx.DAOManager.Game.GetOrUpdateGame(l.ctx,
			opCode, item.SourceId,
			item.GameCode, item.ProviderKey, item.CategoryCode, item.ProviderCode, item.ChannelCode,
			&name, &imageURL, item.SortNo, item.SupportsEmbed, item.SupportsRedirect, int64(item.Status), item.CurrencyCodeList)
		if err != nil {
			l.Errorf("[游戏同步] 同步游戏失败: code=%s, err=%v", item.GameCode, err)
			stat.Failed++
			continue
		}
		l.Debugf("[游戏同步] 成功同步游戏: code=%s", item.GameCode)
		stat.Success++
	}

	// 删除不在发布列表中的游戏
	l.Infof("[游戏同步] 准备删除不存在的游戏")
	deleteCount, err := l.svcCtx.DAOManager.Game.DeleteGameNotIn(l.ctx, opCode, publishedGameCodeSet)
	if err != nil {
		l.Errorf("[游戏同步] 删除过期游戏失败: %v", err)
	} else if deleteCount > 0 {
		l.Infof("[游戏同步] 已删除 %d 个过期游戏", deleteCount)
	}

	l.Infof("[游戏同步] 完成, 成功: %d, 失败: %d, 已删除: %d", stat.Success, stat.Failed, deleteCount)
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
