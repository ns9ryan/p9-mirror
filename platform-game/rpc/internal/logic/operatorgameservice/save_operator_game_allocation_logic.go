package operatorgameservicelogic

import (
	"context"

	"oa.98ent.com/p9/platform-game/rpc/internal/svc"
	"oa.98ent.com/p9/platform-game/rpc/internal/utils"
	"oa.98ent.com/p9/platform-game/rpc/pb/platform_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveOperatorGameAllocationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveOperatorGameAllocationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveOperatorGameAllocationLogic {
	return &SaveOperatorGameAllocationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 保存分站游戏分配
func (l *SaveOperatorGameAllocationLogic) SaveOperatorGameAllocation(in *platform_game.SaveOperatorGameAllocationRequest) (*platform_game.SaveOperatorGameAllocationResp, error) {
	var resp platform_game.SaveOperatorGameAllocationResp
	resp.Total = int64(len(in.GetItems()))
	if l.svcCtx == nil || l.svcCtx.DAOManager == nil {
		l.Errorf("[SaveOperatorGameAllocation] database not available")
		return &resp, nil
	}

	opCode := in.GetOpCode()
	if opCode == "" {
		l.Errorf("[SaveOperatorGameAllocation] op_code is empty")
		return &resp, nil
	}
	// 校验opCode是否存在
	isExist, err := l.svcCtx.DAOManager.Operator.ExistByCode(l.ctx, opCode)
	if err != nil || !isExist {
		l.Errorf("[SaveOperatorGameAllocation] operator not found: op_code=%s, err=%v",
			opCode, err)
		return &resp, nil
	}

	// 如果isAllCheck为true，拉取所有游戏数据
	var items []*platform_game.SaveOperatorGameAllocationInfo
	if in.GetIsCheckAll() {
		allGames, err := l.svcCtx.DAOManager.Game.GetAllGames(l.ctx)
		if err != nil {
			l.Errorf("[SaveOperatorGameAllocation] failed to get all games: err=%v", err)
			return &resp, nil
		}

		for _, game := range allGames {
			items = append(items, &platform_game.SaveOperatorGameAllocationInfo{
				Code:        game.GameCode,
				CheckStatus: 1, // 默认为创建状态
			})
		}
		resp.Total = int64(len(items))
		l.Infof("[SaveOperatorGameAllocation] isAllCheck=true, loaded %d games", len(items))
	} else {
		items = in.GetItems()
	}

	for _, item := range items {
		if item == nil {
			resp.Failed++
			continue
		}

		gameCode := item.GetCode()
		checkStatus := item.GetCheckStatus()

		// 查询数据库
		existRecord, err := l.svcCtx.DAOManager.OperatorGame.GetByOpCodeAndGameCode(l.ctx, opCode, gameCode)
		if err != nil {
			l.Errorf("[SaveOperatorGameAllocation] query failed: op_code=%s, game_code=%s, err=%v",
				opCode, gameCode, err)
			resp.Failed++
			continue
		}

		// 数据库存在
		if existRecord != nil {
			if checkStatus == 1 {
				// exist++
				resp.Exist++
			} else if checkStatus == 2 {
				// 删除记录
				err := l.svcCtx.DAOManager.OperatorGame.DeleteAllocationByID(l.ctx, existRecord.ID)
				if err != nil {
					l.Errorf("[SaveOperatorGameAllocation] delete failed: id=%d, err=%v",
						existRecord.ID, err)
					resp.Failed++
					continue
				}
				resp.Deleted++
			}
		} else {
			// 校验是否存在game
			gameRecord, err := l.svcCtx.DAOManager.Game.GetGameByCode(l.ctx, gameCode)
			if err != nil || gameRecord == nil {
				l.Errorf("[SaveOperatorGameAllocation] game not found: game_code=%s, err=%v",
					gameCode, err)
				resp.Failed++
				continue
			}

			if checkStatus == 1 {
				l.Infof("[SaveOperatorGameAllocation] queried game record: game_code=%s, record=%s", gameCode, utils.JSON(gameRecord))
				// 创建记录
				_, err := l.svcCtx.DAOManager.OperatorGame.CreateAllocation(l.ctx, gameRecord.Name, opCode, gameCode)
				if err != nil {
					l.Errorf("[SaveOperatorGameAllocation] create failed: op_code=%s, game_code=%s, err=%v",
						opCode, gameCode, err)
					resp.Failed++
					continue
				}
				resp.Created++

				// 获取游戏的扩展信息（渠道和供应商）
				gameExtInfo, err := l.svcCtx.DAOManager.Game.GetGameExtraInfo(l.ctx, gameRecord)
				l.Infof("[SaveOperatorGameAllocation] got game ext info: game_code=%s, ext_info=%s", gameCode, utils.JSON(gameExtInfo))
				if err != nil {
					l.Errorf("[SaveOperatorGameAllocation] get game ext info failed: game_code=%s, err=%v",
						gameCode, err)
					continue
				}
				// 添加游戏分类到分配表（如果分类存在）
				if gameExtInfo.CategoryCode != "" {
					categoryExists, err := l.svcCtx.DAOManager.OperatorGameCategory.ExistByOpCodeAndCategoryCode(l.ctx, opCode, gameExtInfo.CategoryCode)
					l.Infof("[SaveOperatorGameAllocation] checked category existence: op_code=%s, category_code=%s, exists=%v", opCode, gameExtInfo.CategoryCode, categoryExists)
					if err != nil {
						l.Errorf("[SaveOperatorGameAllocation] check category existence failed: op_code=%s, category_code=%s, err=%v",
							opCode, gameExtInfo.CategoryCode, err)
						continue
					}

					if !categoryExists {
						_, err := l.svcCtx.DAOManager.OperatorGameCategory.CreateAllocation(l.ctx, opCode, gameExtInfo.CategoryCode)
						if err != nil {
							l.Errorf("[SaveOperatorGameAllocation] create operator game category failed: op_code=%s, category_code=%s, err=%v",
								opCode, gameExtInfo.CategoryCode, err)
						} else {
							l.Infof("[SaveOperatorGameAllocation] operator game category created: op_code=%s, category_code=%s",
								opCode, gameExtInfo.CategoryCode)
						}
					}
				}

				// 添加游戏供应商到分配表（如果供应商存在）
				if gameExtInfo.ProviderCode != "" {
					providerExists, err := l.svcCtx.DAOManager.OperatorGameProvider.ExistByOpCodeAndProviderCode(l.ctx, opCode, gameExtInfo.ProviderCode)
					l.Infof("[SaveOperatorGameAllocation] checked provider existence: op_code=%s, provider_code=%s, exists=%v", opCode, gameExtInfo.ProviderCode, providerExists)
					if err != nil {
						l.Errorf("[SaveOperatorGameAllocation] check provider existence failed: op_code=%s, provider_code=%s, err=%v",
							opCode, gameExtInfo.ProviderCode, err)
						continue
					}

					if !providerExists {
						_, err := l.svcCtx.DAOManager.OperatorGameProvider.CreateAllocation(l.ctx, opCode, gameExtInfo.ProviderCode)
						if err != nil {
							l.Errorf("[SaveOperatorGameAllocation] create operator game provider failed: op_code=%s, provider_code=%s, err=%v",
								opCode, gameExtInfo.ProviderCode, err)
						} else {
							l.Infof("[SaveOperatorGameAllocation] operator game provider created: op_code=%s, provider_code=%s",
								opCode, gameExtInfo.ProviderCode)
						}
					}
				}

				// 添加游戏渠道到分配表（如果渠道存在）
				if gameExtInfo.ChannelCode != "" {
					channelExists, err := l.svcCtx.DAOManager.OperatorGameChannel.ExistByOpCodeAndChannelCode(l.ctx, opCode, gameExtInfo.ChannelCode)
					l.Infof("[SaveOperatorGameAllocation] checked channel existence: op_code=%s, channel_code=%s, exists=%v", opCode, gameExtInfo.ChannelCode, channelExists)
					if err != nil {
						l.Errorf("[SaveOperatorGameAllocation] check channel existence failed: op_code=%s, channel_code=%s, err=%v",
							opCode, gameExtInfo.ChannelCode, err)
						continue
					}

					if !channelExists {
						_, err := l.svcCtx.DAOManager.OperatorGameChannel.CreateAllocation(l.ctx, opCode, gameExtInfo.ChannelCode)
						if err != nil {
							l.Errorf("[SaveOperatorGameAllocation] create operator game channel failed: op_code=%s, channel_code=%s, err=%v",
								opCode, gameExtInfo.ChannelCode, err)
						} else {
							l.Infof("[SaveOperatorGameAllocation] operator game channel created: op_code=%s, channel_code=%s",
								opCode, gameExtInfo.ChannelCode)
						}
					}
				}
			}
		}
	}

	return &resp, nil
}
