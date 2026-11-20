package types

// FirstID 优先读取 id，未传时沿用 ids 的第一项；两者都未传时返回 0。
// 辅助方法独立保存，避免重新生成 types.go 时被覆盖。
func (r IDReq) FirstID() uint64 {
	if r.ID != 0 {
		return r.ID
	}
	if len(r.IDs) > 0 {
		return r.IDs[0]
	}
	return 0
}

// FileReply 保留原有接口文档使用的响应类型。
type FileReply struct {
	Code int      `json:"code"`
	Msg  string   `json:"msg"`
	Data FileData `json:"data,omitempty"`
}

// PresignReply 保留原有预签名响应类型。
type PresignReply struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data PresignData `json:"data,omitempty"`
}

// FileListReply 保留原有文件列表响应类型。
type FileListReply struct {
	Code int          `json:"code"`
	Msg  string       `json:"msg"`
	Data FileListData `json:"data,omitempty"`
}

// ProviderReply 保留原有服务商响应类型。
type ProviderReply struct {
	Code int          `json:"code"`
	Msg  string       `json:"msg"`
	Data ProviderView `json:"data,omitempty"`
}

// ProviderListReply 保留原有服务商列表响应类型。
type ProviderListReply struct {
	Code int              `json:"code"`
	Msg  string           `json:"msg"`
	Data ProviderListData `json:"data,omitempty"`
}
