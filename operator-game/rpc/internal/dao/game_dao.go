package dao

import (
	"context"
	"fmt"

	"oa.98ent.com/p9/operator-game/rpc/ent"
	"oa.98ent.com/p9/operator-game/rpc/ent/game"
	"oa.98ent.com/p9/operator-game/rpc/ent/schema"
)

type GameDAO struct {
	db *ent.Client
}

func NewGameDAO(db *ent.Client) *GameDAO {
	return &GameDAO{db: db}
}

// GetGameByID 根据ID获取游戏
func (d *GameDAO) GetGameByID(ctx context.Context, id int64) (*ent.Game, error) {
	return d.db.Game.Get(ctx, id)
}

// GetGameList 获取游戏列表
func (d *GameDAO) GetGameList(ctx context.Context, opts ...GameListOption) ([]*ent.Game, error) {
	query := d.db.Game.Query()

	// 应用选项
	opt := &GameListOptions{}
	for _, o := range opts {
		o(opt)
	}

	// 应用过滤条件
	if opt.OpCode != "" {
		query = query.Where(game.OpCodeEQ(opt.OpCode))
	}
	if opt.GameCode != "" {
		query = query.Where(game.GameCodeEQ(opt.GameCode))
	}
	if opt.Name != "" {
		query = query.Where(game.NameContains(opt.Name))
	}
	if opt.CategoryCode != "" {
		query = query.Where(game.CategoryCodeEQ(opt.CategoryCode))
	}
	if opt.ProviderCode != "" {
		query = query.Where(game.ProviderCodeEQ(opt.ProviderCode))
	}
	if opt.ChannelCode != "" {
		query = query.Where(game.ChannelCodeEQ(opt.ChannelCode))
	}
	if opt.Status > 0 {
		query = query.Where(game.StatusEQ(opt.Status))
	}

	// 分页
	if opt.Offset > 0 {
		query = query.Offset(int(opt.Offset))
	}
	if opt.Limit > 0 {
		query = query.Limit(int(opt.Limit))
	}

	// 排序
	if opt.OrderBy != "" {
		query = query.Order(ent.Asc(opt.OrderBy))
	} else {
		query = query.Order(ent.Desc(game.FieldID))
	}

	return query.All(ctx)
}

// CountGameList 获取游戏列表总数
func (d *GameDAO) CountGameList(ctx context.Context, opts ...GameListOption) (int, error) {
	query := d.db.Game.Query()

	// 应用选项
	opt := &GameListOptions{}
	for _, o := range opts {
		o(opt)
	}

	// 应用过滤条件
	if opt.OpCode != "" {
		query = query.Where(game.OpCodeEQ(opt.OpCode))
	}
	if opt.GameCode != "" {
		query = query.Where(game.GameCodeEQ(opt.GameCode))
	}
	if opt.Name != "" {
		query = query.Where(game.NameContains(opt.Name))
	}
	if opt.CategoryCode != "" {
		query = query.Where(game.CategoryCodeEQ(opt.CategoryCode))
	}
	if opt.ProviderCode != "" {
		query = query.Where(game.ProviderCodeEQ(opt.ProviderCode))
	}
	if opt.ChannelCode != "" {
		query = query.Where(game.ChannelCodeEQ(opt.ChannelCode))
	}
	if opt.Status > 0 {
		query = query.Where(game.StatusEQ(opt.Status))
	}

	count, err := query.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count games failed: %w", err)
	}
	return count, nil
}

// UpdateGame 更新游戏
func (d *GameDAO) UpdateGame(ctx context.Context, id int64, name string, sortNo, status int64) (*ent.Game, error) {
	updater := d.db.Game.UpdateOneID(id)

	if name != "" {
		updater = updater.SetName(name)
	}
	if sortNo > 0 {
		updater = updater.SetSortNo(sortNo)
	}
	if status > 0 {
		updater = updater.SetStatus(status)
	}

	g, err := updater.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update game failed: %w", err)
	}
	return g, nil
}

// CreateGame 创建游戏
func (d *GameDAO) CreateGame(ctx context.Context, opCode string, sourceID int64, gameCode string, providerKey string, categoryCode, providerCode string, channelCode string,
	name, imageURL *string, sortNo int64, supportsEmbed, supportsRedirect bool, status int64, currencyCodeList []string) (*ent.Game, error) {
	currencyInfoList := make([]schema.CurrencyInfo, len(currencyCodeList))
	for i, code := range currencyCodeList {
		currencyInfoList[i] = schema.CurrencyInfo{Code: code}
	}

	creator := d.db.Game.Create().
		SetOpCode(opCode).
		SetSourceID(sourceID).
		SetGameCode(gameCode).
		SetCategoryCode(categoryCode).
		SetProviderCode(providerCode).
		SetChannelCode(channelCode).
		SetProviderKey(providerKey).
		SetSortNo(sortNo).
		SetSupportsEmbed(supportsEmbed).
		SetSupportsRedirect(supportsRedirect).
		SetStatus(status).
		SetCurrencyList(currencyInfoList)

	if name != nil && *name != "" {
		creator = creator.SetName(*name)
	}
	if imageURL != nil && *imageURL != "" {
		creator = creator.SetImageURL(*imageURL)
	}

	g, err := creator.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create game failed: %w", err)
	}
	return g, nil
}

// GetOrUpdateGame 获取或更新游戏（更新除了sortNo和status之外的所有字段）
func (d *GameDAO) GetOrUpdateGame(ctx context.Context, opCode string, sourceID int64, gameCode string, providerKey string, categoryCode, providerCode string, channelCode string,
	name, imageURL *string, sortNo int64, supportsEmbed, supportsRedirect bool, status int64, currencyCodeList []string) (*ent.Game, error) {

	// 先查询是否存在
	g, err := d.db.Game.Query().
		Where(game.OpCodeEQ(opCode)).
		Where(game.GameCodeEQ(gameCode)).
		First(ctx)
	if err == nil {
		// 存在则更新其他字段（除了sortNo和status）
		updater := d.db.Game.UpdateOneID(g.ID).
			SetProviderKey(providerKey).
			SetCategoryCode(categoryCode).
			SetProviderCode(providerCode).
			SetChannelCode(channelCode).
			SetSupportsEmbed(supportsEmbed).
			SetSupportsRedirect(supportsRedirect)

		if name != nil && *name != "" {
			updater = updater.SetName(*name)
		}
		if imageURL != nil && *imageURL != "" {
			updater = updater.SetImageURL(*imageURL)
		}

		currencyInfoList := make([]schema.CurrencyInfo, len(currencyCodeList))
		for i, code := range currencyCodeList {
			currencyInfoList[i] = schema.CurrencyInfo{Code: code}
		}
		updater = updater.SetCurrencyList(currencyInfoList)

		return updater.Save(ctx)
	}

	// 不存在则创建新的
	return d.CreateGame(ctx, opCode, sourceID, gameCode, providerKey, categoryCode, providerCode, channelCode, name, imageURL, sortNo, supportsEmbed, supportsRedirect, status, currencyCodeList)
}

func (d *GameDAO) CountGameByChannel(ctx context.Context, channelCode string) (int, error) {
	return d.db.Game.Query().
		Where(game.ChannelCodeEQ(channelCode)).
		Count(ctx)
}

// GameListOptions 列表选项
type GameListOptions struct {
	OpCode       string
	GameCode     string
	Name         string
	CategoryCode string
	ProviderCode string
	ChannelCode  string
	Status       int64
	Offset       int64
	Limit        int64
	OrderBy      string
}

// GameListOption 列表选项函数
type GameListOption func(*GameListOptions)

func WithGameOpCode(code string) GameListOption {
	return func(opt *GameListOptions) {
		opt.OpCode = code
	}
}

func WithGameCode(code string) GameListOption {
	return func(opt *GameListOptions) {
		opt.GameCode = code
	}
}

func WithGameName(name string) GameListOption {
	return func(opt *GameListOptions) {
		opt.Name = name
	}
}

func WithGameChannelID(code string) GameListOption {
	return func(opt *GameListOptions) {
		opt.ChannelCode = code
	}
}

func WithGameStatus(status int64) GameListOption {
	return func(opt *GameListOptions) {
		opt.Status = status
	}
}

func WithGameOffset(offset int64) GameListOption {
	return func(opt *GameListOptions) {
		opt.Offset = offset
	}
}

func WithGameLimit(limit int64) GameListOption {
	return func(opt *GameListOptions) {
		opt.Limit = limit
	}
}

func WithGameOrderBy(orderBy string) GameListOption {
	return func(opt *GameListOptions) {
		opt.OrderBy = orderBy
	}
}

func WithGameCategoryCode(code string) GameListOption {
	return func(opt *GameListOptions) {
		opt.CategoryCode = code
	}
}

func WithGameProviderCode(code string) GameListOption {
	return func(opt *GameListOptions) {
		opt.ProviderCode = code
	}
}

func WithGameChannelCode(code string) GameListOption {
	return func(opt *GameListOptions) {
		opt.ChannelCode = code
	}
}
