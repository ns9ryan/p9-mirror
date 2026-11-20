package constant

// 错误码定义 - 遵循Simple Admin规范

// 标准HTTP状态码 - 遵循 core-api 规范
const (
	CodeSuccess         = 0     // 成功
	CodeBadRequest      = 400   // 参数错误
	CodeUnauthorized    = 401   // 未登录/无效token
	CodeForbidden       = 403   // 无权限或资源被禁用
	CodeNotFound        = 404   // 不存在
	CodeAccessExpired   = 498   // access token 过期
	CodeInternalError   = 500   // 内部错误
	CodeInvalidArgument = 40001 // 无效参数
)
