package dao

import (
	"context"
	"fmt"

	"oa.98ent.com/p9/operator-game/rpc/ent"
	"oa.98ent.com/p9/operator-game/rpc/ent/gameprovider"
)

type GameProviderDAO struct {
	db *ent.Client
}

func NewGameProviderDAO(db *ent.Client) *GameProviderDAO {
	return &GameProviderDAO{db: db}
}

// GetGameProviderByID 根据ID获取游戏供应商
func (d *GameProviderDAO) GetGameProviderByID(ctx context.Context, id int64) (*ent.GameProvider, error) {
	return d.db.GameProvider.Get(ctx, id)
}

// GetGameProviderList 获取游戏供应商列表
func (d *GameProviderDAO) GetGameProviderList(ctx context.Context, opts ...GameProviderListOption) ([]*ent.GameProvider, error) {
	query := d.db.GameProvider.Query()

	// 应用选项
	opt := &GameProviderListOptions{}
	for _, o := range opts {
		o(opt)
	}

	// 应用过滤条件
	if opt.ProviderCode != "" {
		query = query.Where(gameprovider.ProviderCodeEQ(opt.ProviderCode))
	}
	if opt.Status > 0 {
		query = query.Where(gameprovider.StatusEQ(opt.Status))
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
		query = query.Order(ent.Desc(gameprovider.FieldID))
	}

	return query.All(ctx)
}

// CountGameProviderList 获取游戏供应商列表总数
func (d *GameProviderDAO) CountGameProviderList(ctx context.Context, opts ...GameProviderListOption) (int, error) {
	query := d.db.GameProvider.Query()

	// 应用选项
	opt := &GameProviderListOptions{}
	for _, o := range opts {
		o(opt)
	}

	// 应用过滤条件
	if opt.OpCode != "" {
		query = query.Where(gameprovider.OpCodeEQ(opt.OpCode))
	}
	if opt.ProviderCode != "" {
		query = query.Where(gameprovider.ProviderCodeEQ(opt.ProviderCode))
	}
	if opt.Status > 0 {
		query = query.Where(gameprovider.StatusEQ(opt.Status))
	}

	count, err := query.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count game providers failed: %w", err)
	}
	return count, nil
}

// UpdateGameProvider 更新游戏供应商
func (d *GameProviderDAO) UpdateGameProvider(ctx context.Context, id int64, sortNo, status int64) (*ent.GameProvider, error) {
	updater := d.db.GameProvider.UpdateOneID(id)

	if sortNo > 0 {
		updater = updater.SetSortNo(sortNo)
	}
	if status > 0 {
		updater = updater.SetStatus(status)
	}

	gp, err := updater.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update game provider failed: %w", err)
	}
	return gp, nil
}

// CreateGameProvider 创建游戏供应商
func (d *GameProviderDAO) CreateGameProvider(ctx context.Context, opCode, providerCode string, channelCode string, logoURL *string, sortNo, status int64) (*ent.GameProvider, error) {
	creator := d.db.GameProvider.Create().
		SetOpCode(opCode).
		SetProviderCode(providerCode).
		SetChannelCode(channelCode).
		SetSortNo(sortNo).
		SetStatus(status)

	if logoURL != nil && *logoURL != "" {
		creator = creator.SetLogoURL(*logoURL)
	}

	gp, err := creator.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create game provider failed: %w", err)
	}
	return gp, nil
}

// GetOrUpdateGameProvider 获取或更新游戏供应商（更新除了sortNo和status之外的所有字段）
func (d *GameProviderDAO) GetOrUpdateGameProvider(ctx context.Context, opCode, providerCode string, channelCode string, logoURL *string, sortNo, status int64) (*ent.GameProvider, error) {
	// 先查询是否存在
	gp, err := d.db.GameProvider.Query().
		Where(gameprovider.OpCodeEQ(opCode)).
		Where(gameprovider.ProviderCodeEQ(providerCode)).
		First(ctx)
	if err == nil {
		// 存在则更新其他字段（除了sortNo和status）
		updater := d.db.GameProvider.UpdateOneID(gp.ID).SetChannelCode(channelCode)
		if logoURL != nil && *logoURL != "" {
			updater = updater.SetLogoURL(*logoURL)
		}
		return updater.Save(ctx)
	}

	// 不存在则创建新的
	return d.CreateGameProvider(ctx, opCode, providerCode, channelCode, logoURL, sortNo, status)
}

// CountGameProviderByChannel 统计指定渠道的游戏供应商数量
func (d *GameProviderDAO) CountGameProviderByChannel(ctx context.Context, channelCode string) (int, error) {
	return d.db.GameProvider.Query().
		Where(gameprovider.ChannelCodeEQ(channelCode)).
		Count(ctx)
}

// GameProviderListOptions 列表选项
type GameProviderListOptions struct {
	OpCode       string
	ProviderCode string
	Status       int64
	Offset       int64
	Limit        int64
	OrderBy      string
}

// GameProviderListOption 列表选项函数
type GameProviderListOption func(*GameProviderListOptions)

func WithProviderOpCode(code string) GameProviderListOption {
	return func(opt *GameProviderListOptions) {
		opt.OpCode = code
	}
}

func WithProviderCode(code string) GameProviderListOption {
	return func(opt *GameProviderListOptions) {
		opt.ProviderCode = code
	}
}

func WithGameProviderStatus(status int64) GameProviderListOption {
	return func(opt *GameProviderListOptions) {
		opt.Status = status
	}
}

func WithGameProviderOffset(offset int64) GameProviderListOption {
	return func(opt *GameProviderListOptions) {
		opt.Offset = offset
	}
}

func WithGameProviderLimit(limit int64) GameProviderListOption {
	return func(opt *GameProviderListOptions) {
		opt.Limit = limit
	}
}

func WithGameProviderOrderBy(orderBy string) GameProviderListOption {
	return func(opt *GameProviderListOptions) {
		opt.OrderBy = orderBy
	}
}
