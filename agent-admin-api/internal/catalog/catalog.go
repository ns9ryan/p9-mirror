package catalog

import (
	"oa.98ent.com/p9/core/rpc/coreclient"
)

const (
	menuTypeDir    int32 = 0
	menuTypeMenu   int32 = 1
	menuTypeButton int32 = 2
)

// CatalogReq 菜单、API目录、多语言数据种子数据
func catalogReq(code string) *coreclient.RegisterCatalogReq {
	return &coreclient.RegisterCatalogReq{
		// Menus: []*coreclient.RegisterMenuReq{},
		// Apis:  []*coreclient.CreateApiReq{},
		// I18NLangs: langSeeds(code),
		I18N: i18nSeeds(code),
	}
}
