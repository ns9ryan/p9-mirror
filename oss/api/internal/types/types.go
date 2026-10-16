package types

import "oa.98ent.com/p9/oss/common/oss"

type FileData struct {
	URL          string `json:"url"`
	Key          string `json:"key"`
	Scene        string `json:"scene,omitempty"`
	Visibility   string `json:"visibility,omitempty"`
	URLExpiresIn int    `json:"url_expires_in,omitempty"`
}

type PresignRequest struct {
	Filename string `json:"filename" example:"photo.jpg"`
	Size     int64  `json:"size" example:"204800"`
	Scene    string `json:"scene" example:"public"`
}

type CompleteRequest struct {
	Key   string `json:"key" example:"public/20260825/ab12cd.jpg"`
	Scene string `json:"scene"`
}

type PresignData struct {
	PutURL       string            `json:"put_url"`
	Method       string            `json:"method"`
	Headers      map[string]string `json:"headers"`
	Key          string            `json:"key"`
	URL          string            `json:"url"`
	ExpiresIn    int               `json:"expires_in"`
	Scene        string            `json:"scene"`
	Visibility   string            `json:"visibility,omitempty"`
	URLExpiresIn int               `json:"url_expires_in,omitempty"`
}

type FileListReq struct {
	Scene    string `json:"scene"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type FileListData struct {
	List  []FileInfo `json:"list"`
	Total int64      `json:"total"`
}

type FileInfo struct {
	ID           uint64 `json:"id"`
	Key          string `json:"key"`
	Scene        string `json:"scene"`
	Bucket       string `json:"bucket"`
	Name         string `json:"name"`
	Size         int64  `json:"size"`
	ContentType  string `json:"content_type"`
	URL          string `json:"url"`
	Visibility   string `json:"visibility,omitempty"`
	URLExpiresIn int    `json:"url_expires_in,omitempty"`
	CreatedAt    int64  `json:"created_at"`
}

type FileDeleteReq struct {
	Key string `json:"key" example:"public/20260824/aaa.png"`
}

type ResolveReq struct {
	Key string `form:"key"`
	URL string `form:"url"`
}

type IDReq struct {
	ID  uint64   `form:"id" json:"id"`
	IDs []uint64 `json:"ids"`
}

func (r IDReq) FirstID() uint64 {
	if r.ID != 0 {
		return r.ID
	}
	if len(r.IDs) > 0 {
		return r.IDs[0]
	}
	return 0
}

type ProviderView struct {
	ID        uint64                `json:"id,omitempty"`
	Name      string                `json:"name"`
	IsDefault bool                  `json:"is_default"`
	Driver    string                `json:"driver"`
	Endpoint  string                `json:"endpoint"`
	FileURL   string                `json:"file_url"`
	AccessKey string                `json:"access_key"`
	SecretKey string                `json:"secret_key,omitempty"`
	PathStyle bool                  `json:"path_style"`
	Region    string                `json:"region"`
	Buckets   map[string]oss.Bucket `json:"buckets"`
	Source    string                `json:"source,omitempty"`
}

type ProviderListData struct {
	List  []ProviderView `json:"list"`
	Total int64          `json:"total"`
}

type FileReply struct {
	Code int      `json:"code"`
	Msg  string   `json:"msg"`
	Data FileData `json:"data,omitempty"`
}

type PresignReply struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data PresignData `json:"data,omitempty"`
}

type FileListReply struct {
	Code int          `json:"code"`
	Msg  string       `json:"msg"`
	Data FileListData `json:"data,omitempty"`
}

type ProviderReply struct {
	Code int          `json:"code"`
	Msg  string       `json:"msg"`
	Data ProviderView `json:"data,omitempty"`
}

type ProviderListReply struct {
	Code int              `json:"code"`
	Msg  string           `json:"msg"`
	Data ProviderListData `json:"data,omitempty"`
}
