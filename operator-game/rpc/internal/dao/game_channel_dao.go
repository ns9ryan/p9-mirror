package dao

import (
	"context"
	"fmt"

	"oa.98ent.com/p9/operator-game/rpc/ent"
	"oa.98ent.com/p9/operator-game/rpc/ent/gamechannel"
)

type GameChannelDAO struct {
	db *ent.Client
}

func NewGameChannelDAO(db *ent.Client) *GameChannelDAO {
	return &GameChannelDAO{db: db}
}

// GetGameChannelByID 根据ID获取游戏渠道
func (d *GameChannelDAO) GetGameChannelByID(ctx context.Context, id int64) (*ent.GameChannel, error) {
	return d.db.GameChannel.Get(ctx, id)
}

// GetGameChannelList 获取游戏渠道列表
func (d *GameChannelDAO) GetGameChannelList(ctx context.Context, opts ...GameChannelListOption) ([]*ent.GameChannel, error) {
	query := d.db.GameChannel.Query()

	// 应用选项
	opt := &GameChannelListOptions{}
	for _, o := range opts {
		o(opt)
	}

	// 应用过滤条件
	if opt.ChannelCode != "" {
		query = query.Where(gamechannel.ChannelCodeEQ(opt.ChannelCode))
	}
	if opt.Status > 0 {
		query = query.Where(gamechannel.StatusEQ(opt.Status))
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
		query = query.Order(ent.Desc(gamechannel.FieldID))
	}

	return query.All(ctx)
}

// CountGameChannelList 获取游戏渠道列表总数
func (d *GameChannelDAO) CountGameChannelList(ctx context.Context, opts ...GameChannelListOption) (int, error) {
	query := d.db.GameChannel.Query()

	// 应用选项
	opt := &GameChannelListOptions{}
	for _, o := range opts {
		o(opt)
	}

	// 应用过滤条件
	if opt.ChannelCode != "" {
		query = query.Where(gamechannel.ChannelCodeEQ(opt.ChannelCode))
	}
	if opt.Status > 0 {
		query = query.Where(gamechannel.StatusEQ(opt.Status))
	}

	count, err := query.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count game channels failed: %w", err)
	}
	return count, nil
}

// UpdateGameChannel 更新游戏渠道
func (d *GameChannelDAO) UpdateGameChannel(ctx context.Context, id int64, sortNo, status int64) (*ent.GameChannel, error) {
	updater := d.db.GameChannel.UpdateOneID(id)

	if sortNo > 0 {
		updater = updater.SetSortNo(sortNo)
	}
	if status > 0 {
		updater = updater.SetStatus(status)
	}

	gc, err := updater.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update game channel failed: %w", err)
	}
	return gc, nil
}

// CreateGameChannel 创建游戏渠道
func (d *GameChannelDAO) CreateGameChannel(ctx context.Context, sourceID int64, channelCode string, sortNo, loadType, status int64) (*ent.GameChannel, error) {
	gc, err := d.db.GameChannel.Create().
		SetChannelCode(channelCode).
		SetSortNo(sortNo).
		SetLoadType(loadType).
		SetStatus(status).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create game channel failed: %w", err)
	}
	return gc, nil
}

// GetOrCreateGameChannel 获取或创建游戏渠道
func (d *GameChannelDAO) GetOrCreateGameChannel(ctx context.Context, sourceID int64, channelCode string, sortNo, loadType, status int64) (*ent.GameChannel, error) {
	// 先查询是否存在
	gc, err := d.db.GameChannel.Query().
		Where(gamechannel.ChannelCodeEQ(channelCode)).
		First(ctx)
	if err == nil {
		return gc, nil
	}

	// 创建新的
	return d.CreateGameChannel(ctx, sourceID, channelCode, sortNo, loadType, status)
}

// GameChannelListOptions 列表选项
type GameChannelListOptions struct {
	ChannelCode string
	Status      int64
	Offset      int64
	Limit       int64
	OrderBy     string
}

// GameChannelListOption 列表选项函数
type GameChannelListOption func(*GameChannelListOptions)

func WithChannelCode(code string) GameChannelListOption {
	return func(opt *GameChannelListOptions) {
		opt.ChannelCode = code
	}
}

func WithGameChannelStatus(status int64) GameChannelListOption {
	return func(opt *GameChannelListOptions) {
		opt.Status = status
	}
}

func WithGameChannelOffset(offset int64) GameChannelListOption {
	return func(opt *GameChannelListOptions) {
		opt.Offset = offset
	}
}

func WithGameChannelLimit(limit int64) GameChannelListOption {
	return func(opt *GameChannelListOptions) {
		opt.Limit = limit
	}
}

func WithGameChannelOrderBy(orderBy string) GameChannelListOption {
	return func(opt *GameChannelListOptions) {
		opt.OrderBy = orderBy
	}
}
