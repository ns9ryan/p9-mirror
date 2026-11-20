package catalog

import (
	"net/http"
	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/core/rpc/coreclient"
)

func nicknameMenus() []*coreclient.RegisterMenuReq {
	out := []*coreclient.RegisterMenuReq{{Name: "Nickname", Title: "menu.route.nickname", Path: "/base-data/nickname", MenuType: menuTypeMenu, Component: "base-data/nickname/index", ParentName: "BaseData", Sort: 27}}
	for i, op := range []string{"Create", "Update", "Delete"} {
		out = append(out, &coreclient.RegisterMenuReq{Name: "Nickname" + op, Title: "menu.route.nickname" + op, MenuType: menuTypeButton, Permission: "nickname:" + []string{"create", "update", "delete"}[i], ParentName: "Nickname", Sort: int32(271 + i)})
	}
	return out
}
func nicknameAPIs() []*coreclient.CreateApiReq {
	var out []*coreclient.CreateApiReq
	for _, op := range []string{"create", "update", "delete", "get", "list"} {
		method := http.MethodPost
		if op == "get" || op == "list" {
			method = http.MethodGet
		}
		out = append(out, &coreclient.CreateApiReq{Path: "/admin/nickname/" + op, Method: method, Description: "api.nickname." + op, ApiGroup: "nickname", ServiceName: "platform-base-api"})
	}
	return out
}
func nicknameI18n() []*coreclient.I18NItem {
	var out []*coreclient.I18NItem
	addI18n(&out, i18n.GroupMenu, "menu.route.nickname", "昵称管理", "暱稱管理", "Nicknames")
	for _, v := range []struct{ op, zh, tw, en string }{{"Create", "新增昵称", "新增暱稱", "Create nickname"}, {"Update", "修改昵称", "修改暱稱", "Update nickname"}, {"Delete", "删除昵称", "刪除暱稱", "Delete nickname"}} {
		addI18n(&out, i18n.GroupMenu, "menu.route.nickname"+v.op, v.zh, v.tw, v.en)
	}
	for _, v := range []struct{ op, zh, tw, en string }{{"create", "新增昵称", "新增暱稱", "Create nickname"}, {"update", "修改昵称", "修改暱稱", "Update nickname"}, {"delete", "删除昵称", "刪除暱稱", "Delete nickname"}, {"get", "昵称详情", "暱稱詳情", "Get nickname"}, {"list", "昵称列表", "暱稱列表", "List nicknames"}} {
		addI18n(&out, i18n.GroupAPI, "api.nickname."+v.op, v.zh, v.tw, v.en)
	}
	return out
}
