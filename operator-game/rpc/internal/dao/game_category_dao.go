package dao

import (
	"context"
	"fmt"

	"oa.98ent.com/p9/operator-game/rpc/ent"
	"oa.98ent.com/p9/operator-game/rpc/ent/gamecategory"
)

type GameCategoryDAO struct {
	db *ent.Client
}

func NewGameCategoryDAO(db *ent.Client) *GameCategoryDAO {
	return &GameCategoryDAO{db: db}
}

// GetGameCategoryByID 根据ID获取游戏分类
func (d *GameCategoryDAO) GetGameCategoryByID(ctx context.Context, id int64) (*ent.GameCategory, error) {
	return d.db.GameCategory.Get(ctx, id)
}

// GetGameCategoryList 获取游戏分类列表
func (d *GameCategoryDAO) GetGameCategoryList(ctx context.Context, opts ...GameCategoryListOption) ([]*ent.GameCategory, error) {
	query := d.db.GameCategory.Query()

	// 应用选项
	opt := &GameCategoryListOptions{}
	for _, o := range opts {
		o(opt)
	}

	// 应用过滤条件
	if opt.OpCode != "" {
		query = query.Where(gamecategory.OpCodeEQ(opt.OpCode))
	}
	if opt.CategoryCode != "" {
		query = query.Where(gamecategory.CategoryCodeEQ(opt.CategoryCode))
	}
	if opt.Status > 0 {
		query = query.Where(gamecategory.StatusEQ(opt.Status))
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
		query = query.Order(ent.Desc(gamecategory.FieldID))
	}

	return query.All(ctx)
}

// CountGameCategoryList 获取游戏分类列表总数
func (d *GameCategoryDAO) CountGameCategoryList(ctx context.Context, opts ...GameCategoryListOption) (int, error) {
	query := d.db.GameCategory.Query()

	// 应用选项
	opt := &GameCategoryListOptions{}
	for _, o := range opts {
		o(opt)
	}

	// 应用过滤条件
	if opt.OpCode != "" {
		query = query.Where(gamecategory.OpCodeEQ(opt.OpCode))
	}
	if opt.CategoryCode != "" {
		query = query.Where(gamecategory.CategoryCodeEQ(opt.CategoryCode))
	}
	if opt.Status > 0 {
		query = query.Where(gamecategory.StatusEQ(opt.Status))
	}

	count, err := query.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count game categories failed: %w", err)
	}
	return count, nil
}

// UpdateGameCategory 更新游戏分类
func (d *GameCategoryDAO) UpdateGameCategory(ctx context.Context, id int64, sortNo, status int64) (*ent.GameCategory, error) {
	updater := d.db.GameCategory.UpdateOneID(id)

	if sortNo > 0 {
		updater = updater.SetSortNo(sortNo)
	}
	if status > 0 {
		updater = updater.SetStatus(status)
	}

	gc, err := updater.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update game category failed: %w", err)
	}
	return gc, nil
}

// CreateGameCategory 创建游戏分类
func (d *GameCategoryDAO) CreateGameCategory(ctx context.Context, opCode, categoryCode string, sortNo, status int64) (*ent.GameCategory, error) {
	gc, err := d.db.GameCategory.Create().
		SetOpCode(opCode).
		SetCategoryCode(categoryCode).
		SetSortNo(sortNo).
		SetStatus(status).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create game category failed: %w", err)
	}
	return gc, nil
}

// GetOrUpdateGameCategory 获取或更新游戏分类（存在则保留，不存在则创建）
func (d *GameCategoryDAO) GetOrUpdateGameCategory(ctx context.Context, opCode, categoryCode string, sortNo, status int64) (*ent.GameCategory, error) {
	// 先查询是否存在
	gc, err := d.db.GameCategory.Query().
		Where(gamecategory.OpCodeEQ(opCode)).
		Where(gamecategory.CategoryCodeEQ(categoryCode)).
		First(ctx)
	if err == nil {
		// 存在则直接返回，不做更新
		return gc, nil
	}

	// 不存在则创建新的
	return d.CreateGameCategory(ctx, opCode, categoryCode, sortNo, status)
}

// GameCategoryListOptions 列表选项
type GameCategoryListOptions struct {
	OpCode       string
	CategoryCode string
	Status       int64
	Offset       int64
	Limit        int64
	OrderBy      string
}

// GameCategoryListOption 列表选项函数
type GameCategoryListOption func(*GameCategoryListOptions)

func WithCategoryOpCode(code string) GameCategoryListOption {
	return func(opt *GameCategoryListOptions) {
		opt.OpCode = code
	}
}

func WithCategoryCode(code string) GameCategoryListOption {
	return func(opt *GameCategoryListOptions) {
		opt.CategoryCode = code
	}
}

func WithGameCategoryStatus(status int64) GameCategoryListOption {
	return func(opt *GameCategoryListOptions) {
		opt.Status = status
	}
}

func WithGameCategoryOffset(offset int64) GameCategoryListOption {
	return func(opt *GameCategoryListOptions) {
		opt.Offset = offset
	}
}

func WithGameCategoryLimit(limit int64) GameCategoryListOption {
	return func(opt *GameCategoryListOptions) {
		opt.Limit = limit
	}
}

func WithGameCategoryOrderBy(orderBy string) GameCategoryListOption {
	return func(opt *GameCategoryListOptions) {
		opt.OrderBy = orderBy
	}
}
