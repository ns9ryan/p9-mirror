package catalog

import (
	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/core/rpc/coreclient"
)

func menuI18n() []*coreclient.I18NItem {
	var out []*coreclient.I18NItem
	add := func(key, zh, hk, en string) {
		out = append(out,
			&coreclient.I18NItem{I18NGroup: i18n.GroupMenu, TransKey: key, Lang: i18n.LangZH, Value: zh},
			&coreclient.I18NItem{I18NGroup: i18n.GroupMenu, TransKey: key, Lang: i18n.LangHK, Value: hk},
			&coreclient.I18NItem{I18NGroup: i18n.GroupMenu, TransKey: key, Lang: i18n.LangEN, Value: en},
		)
	}
	add("menu.route.oss", "对象存储", "對象存儲", "OSS")
	add("menu.route.ossFile", "文件", "文件", "Files")
	add("menu.route.ossProvider", "服务商", "服務商", "Providers")
	return out
}

func apiI18n() []*coreclient.I18NItem {
	var out []*coreclient.I18NItem
	add := func(key, zh, hk, en string) {
		out = append(out,
			&coreclient.I18NItem{I18NCode: i18n.CodePlatform, I18NGroup: i18n.GroupAPI, TransKey: key, Lang: i18n.LangZH, Value: zh},
			&coreclient.I18NItem{I18NCode: i18n.CodePlatform, I18NGroup: i18n.GroupAPI, TransKey: key, Lang: i18n.LangHK, Value: hk},
			&coreclient.I18NItem{I18NCode: i18n.CodePlatform, I18NGroup: i18n.GroupAPI, TransKey: key, Lang: i18n.LangEN, Value: en},
		)
	}

	// 接口组多语言
	add("api.group.oss", "对象存储", "對象存儲", "OSS")

	// 接口描述多语言
	add("api.ossUrlDetail", "换取访问地址", "換取訪問地址", "Resolve URL")
	add("api.ossFileUpload", "上传文件", "上傳文件", "Upload file")
	add("api.ossUploadPresign", "预签名上传", "預簽名上傳", "Presign upload")
	add("api.ossUploadComplete", "完成预签名上传", "完成預簽名上傳", "Complete presign upload")
	add("api.ossFileList", "文件列表", "文件列表", "File list")
	add("api.ossFileDelete", "删除文件", "刪除文件", "Delete file")
	add("api.ossProviderList", "服务商列表", "服務商列表", "Provider list")
	add("api.ossProviderCreate", "创建服务商", "創建服務商", "Create provider")
	add("api.ossProviderActive", "当前生效服务商", "當前生效服務商", "Active provider")
	add("api.ossProviderDetail", "服务商详情", "服務商詳情", "Provider detail")
	add("api.ossProviderUpdate", "更新服务商", "更新服務商", "Update provider")
	add("api.ossProviderDelete", "删除服务商", "刪除服務商", "Delete provider")
	add("api.ossProviderEnable", "启用服务商", "啟用服務商", "Enable provider")
	return out
}
