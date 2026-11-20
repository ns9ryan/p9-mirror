package dao

import "oa.98ent.com/p9/operator-game/rpc/ent"

// Manager 数据库访问对象管理器
type Manager struct {
	DB           *ent.Client
	Game         *GameDAO
	GameCategory *GameCategoryDAO
	GameProvider *GameProviderDAO
	GameChannel  *GameChannelDAO
}

// NewManager 创建 DAO 管理器
func NewManager(db *ent.Client) *Manager {
	return &Manager{
		DB:           db,
		Game:         NewGameDAO(db),
		GameCategory: NewGameCategoryDAO(db),
		GameProvider: NewGameProviderDAO(db),
		GameChannel:  NewGameChannelDAO(db),
	}
}
