package catalog

import (
	"net/http"

	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/core/rpc/coreclient"
)

// avatarResourceMenus 构造默认头像和头像框的页面菜单、修改及上传按钮权限。
func avatarResourceMenus() []*coreclient.RegisterMenuReq {
	return []*coreclient.RegisterMenuReq{
		{
			Name:       "DefaultAvatar",                  // 菜单或按钮的名称。
			Title:      "menu.route.iconManagement",      // 图标管理标题，使用新键兼容已登记的旧翻译。
			Path:       "/base-data/default-avatar",      // 前端页面路由。
			MenuType:   menuTypeMenu,                     // 页面菜单。
			Component:  "base-data/default-avatar/index", // 前端页面组件。
			ParentName: "BaseData",                       // 归属的上级菜单名称。
			Sort:       25,                               // 同级菜单或按钮的排序值。
		},
		{
			Name:       "DefaultAvatarUpdate",            // 菜单或按钮的名称。
			Title:      "menu.route.defaultAvatarUpdate", // 菜单或按钮标题的翻译键。
			MenuType:   menuTypeButton,                   // 按钮权限项。
			Permission: "default-avatar:update",          // 按钮操作的权限标识。
			ParentName: "DefaultAvatar",                  // 归属的上级菜单名称。
			Sort:       251,                              // 同级菜单或按钮的排序值。
		},
		{
			Name:       "DefaultAvatarUpload",            // 菜单或按钮的名称。
			Title:      "menu.route.defaultAvatarUpload", // 菜单或按钮标题的翻译键。
			MenuType:   menuTypeButton,                   // 按钮权限项。
			Permission: "default-avatar:upload",          // 按钮操作的权限标识。
			ParentName: "DefaultAvatar",                  // 归属的上级菜单名称。
			Sort:       252,                              // 同级菜单或按钮的排序值。
		},
		{
			Name:       "DefaultAvatarFrame",                   // 菜单或按钮的名称。
			Title:      "menu.route.defaultAvatarFrame",        // 菜单或按钮标题的翻译键。
			Path:       "/base-data/default-avatar-frame",      // 前端页面路由。
			MenuType:   menuTypeMenu,                           // 页面菜单。
			HideMenu:   1,                                      // 不下发前端导航，保留原页面及按钮权限。
			Component:  "base-data/default-avatar-frame/index", // 前端页面组件。
			ParentName: "BaseData",                             // 归属的上级菜单名称。
			Sort:       26,                                     // 同级菜单或按钮的排序值。
		},
		{
			Name:       "DefaultAvatarFrameUpdate",            // 菜单或按钮的名称。
			Title:      "menu.route.defaultAvatarFrameUpdate", // 菜单或按钮标题的翻译键。
			MenuType:   menuTypeButton,                        // 按钮权限项。
			Permission: "default-avatar-frame:update",         // 按钮操作的权限标识。
			ParentName: "DefaultAvatarFrame",                  // 归属的上级菜单名称。
			Sort:       261,                                   // 同级菜单或按钮的排序值。
		},
		{
			Name:       "DefaultAvatarFrameUpload",            // 菜单或按钮的名称。
			Title:      "menu.route.defaultAvatarFrameUpload", // 菜单或按钮标题的翻译键。
			MenuType:   menuTypeButton,                        // 按钮权限项。
			Permission: "default-avatar-frame:upload",         // 按钮操作的权限标识。
			ParentName: "DefaultAvatarFrame",                  // 归属的上级菜单名称。
			Sort:       262,                                   // 同级菜单或按钮的排序值。
		},
	}
}

// avatarResourceApis 构造默认头像和头像框的查询、修改接口目录。
// 此处不注册 HTTP 路由，也不校验 Token；实际请求仍走原有路由和认证中间件。
func avatarResourceApis() []*coreclient.CreateApiReq {
	return []*coreclient.CreateApiReq{
		{
			Path:        "/admin/default-avatar/list-all", // 查询全部默认图片的接口路径。
			Method:      http.MethodGet,                   // GET 查询请求。
			Description: "api.defaultAvatarListAll",       // 接口说明的翻译键。
			ApiGroup:    "default-avatar",                 // 接口目录分组。
			ServiceName: "platform-base-api",              // 接口所属服务。
		},
		{
			Path:        "/admin/default-avatar/update", // 修改已有图片配置的接口路径。
			Method:      http.MethodPost,                // POST 保存、修改或上传请求。
			Description: "api.defaultAvatarUpdate",      // 接口说明的翻译键。
			ApiGroup:    "default-avatar",               // 接口目录分组。
			ServiceName: "platform-base-api",            // 接口所属服务。
		},

		{
			Path:        "/admin/default-avatar-frame/list-all", // 查询全部默认图片的接口路径。
			Method:      http.MethodGet,                         // GET 查询请求。
			Description: "api.defaultAvatarFrameListAll",        // 接口说明的翻译键。
			ApiGroup:    "default-avatar-frame",                 // 接口目录分组。
			ServiceName: "platform-base-api",                    // 接口所属服务。
		},
		{
			Path:        "/admin/default-avatar-frame/update", // 修改已有图片配置的接口路径。
			Method:      http.MethodPost,                      // POST 保存、修改或上传请求。
			Description: "api.defaultAvatarFrameUpdate",       // 接口说明的翻译键。
			ApiGroup:    "default-avatar-frame",               // 接口目录分组。
			ServiceName: "platform-base-api",                  // 接口所属服务。
		},
	}
}

// avatarResourceI18n 提供头像素材管理的菜单和接口翻译。
func avatarResourceI18n() []*coreclient.I18NItem {
	var out []*coreclient.I18NItem
	// 菜单分组：页面标题及修改、上传按钮名称，提供简体、繁体和英文文本。
	addI18n(&out, i18n.GroupMenu, "menu.route.iconManagement", "图标管理", "圖示管理", "Icon management")
	addI18n(&out, i18n.GroupMenu, "menu.route.defaultAvatarUpdate", "修改默认头像", "修改預設頭像", "Update default avatar")
	addI18n(&out, i18n.GroupMenu, "menu.route.defaultAvatarUpload", "上传默认头像", "上傳預設頭像", "Upload default avatar")
	addI18n(&out, i18n.GroupMenu, "menu.route.defaultAvatarFrame", "默认头像框", "預設頭像框", "Default avatar frames")
	addI18n(
		&out,
		i18n.GroupMenu,
		"menu.route.defaultAvatarFrameUpdate",
		"修改默认头像框",
		"修改預設頭像框",
		"Update default avatar frame",
	)
	addI18n(
		&out,
		i18n.GroupMenu,
		"menu.route.defaultAvatarFrameUpload",
		"上传默认头像框",
		"上傳預設頭像框",
		"Upload default avatar frame",
	)
	// 接口分组：翻译键与接口目录的 Description 对应。
	addI18n(&out, i18n.GroupAPI, "api.defaultAvatarListAll", "默认头像列表", "預設頭像列表", "Default avatar list")
	addI18n(&out, i18n.GroupAPI, "api.defaultAvatarUpdate", "修改默认头像", "修改預設頭像", "Update default avatar")
	addI18n(&out, i18n.GroupAPI, "api.defaultAvatarFrameListAll", "默认头像框列表", "預設頭像框列表", "Default avatar frame list")
	addI18n(&out, i18n.GroupAPI, "api.defaultAvatarFrameUpdate", "修改默认头像框", "修改預設頭像框", "Update default avatar frame")
	return out
}
