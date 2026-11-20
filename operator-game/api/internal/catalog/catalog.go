package catalog

import (
	"net/http"

	"oa.98ent.com/p9/core/rpc/coreclient"
)

const (
	menuTypeDir    int32 = 0
	menuTypeMenu   int32 = 1
	menuTypeButton int32 = 2
)

// AdminReq 管理员请求(菜单相关种子数据)
func AdminReq(code string) *coreclient.RegisterCatalogReq {
	return &coreclient.RegisterCatalogReq{
		Menus: []*coreclient.RegisterMenuReq{
			{Name: "Game", Title: "menu.route.game", Path: "/gameManage", MenuType: menuTypeDir, Sort: 40},
			{Name: "GameCategory", Title: "menu.route.gameCategory", Path: "/gameManage/category", MenuType: menuTypeMenu, Component: "gameManage/category/index", ParentName: "Game", Sort: 1},
			{Name: "GameChannel", Title: "menu.route.gameChannel", Path: "/gameManage/channel", MenuType: menuTypeMenu, Component: "gameManage/channel/index", ParentName: "Game", Sort: 2},
			{Name: "GameManufacturer", Title: "menu.route.gameManufacturer", Path: "/gameManage/manufacturer", MenuType: menuTypeMenu, Component: "gameManage/manufacturer/index", ParentName: "Game", Sort: 3},
			{Name: "IndieGame", Title: "menu.route.gameIndie", Path: "/gameManage/indieGame", MenuType: menuTypeMenu, Component: "gameManage/indieGame/index", ParentName: "Game", Sort: 4},
		},
		Apis: []*coreclient.CreateApiReq{
			{Path: "/operator-game/game-category/list", Method: http.MethodGet, Description: "api.gameCategoryList", ApiGroup: "game", ServiceName: "operator-game-api"},
			{Path: "/operator-game/game-channel/list", Method: http.MethodGet, Description: "api.gameChannelList", ApiGroup: "game", ServiceName: "operator-game-api"},
			{Path: "/operator-game/game-provider/list", Method: http.MethodGet, Description: "api.gameProviderList", ApiGroup: "game", ServiceName: "operator-game-api"},
			{Path: "/operator-game/game/list", Method: http.MethodGet, Description: "api.gameList", ApiGroup: "game", ServiceName: "operator-game-api"},

			{Path: "/operator-game/game/get", Method: http.MethodGet, Description: "api.gameGet", ApiGroup: "game", ServiceName: "operator-game-api"},
			{Path: "/operator-game/game-category/get", Method: http.MethodGet, Description: "api.gameCategoryGet", ApiGroup: "game", ServiceName: "operator-game-api"},
			{Path: "/operator-game/game-channel/get", Method: http.MethodGet, Description: "api.gameChannelGet", ApiGroup: "game", ServiceName: "operator-game-api"},
			{Path: "/operator-game/game-provider/get", Method: http.MethodGet, Description: "api.gameProviderGet", ApiGroup: "game", ServiceName: "operator-game-api"},

			{Path: "/operator-game/game-category/update", Method: http.MethodPost, Description: "api.gameCategoryUpdate", ApiGroup: "game", ServiceName: "operator-game-api"},
			{Path: "/operator-game/game-channel/update", Method: http.MethodPost, Description: "api.gameChannelUpdate", ApiGroup: "game", ServiceName: "operator-game-api"},
			{Path: "/operator-game/game-provider/update", Method: http.MethodPost, Description: "api.gameProviderUpdate", ApiGroup: "game", ServiceName: "operator-game-api"},
			{Path: "/operator-game/game/update", Method: http.MethodPost, Description: "api.gameUpdate", ApiGroup: "game", ServiceName: "operator-game-api"},
		},
		I18N:      append(append(menuI18n(code), apiI18n(code)...)),
		I18NLangs: langSeeds(),
	}
}
