package memberservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"
)

// ImageOption 是会员创建时的临时图片选项。
type ImageOption struct {
	Key  string
	Name string
	URL  string
}

var defaultAvatars = []ImageOption{
	{Key: "avatar-01", Name: "默认头像 1", URL: "https://example.com/p9/member/avatar-01.png"},
	{Key: "avatar-02", Name: "默认头像 2", URL: "https://example.com/p9/member/avatar-02.png"},
	{Key: "avatar-03", Name: "默认头像 3", URL: "https://example.com/p9/member/avatar-03.png"},
}

var defaultAvatarFrames = []ImageOption{
	{Key: "frame-01", Name: "默认头像框 1", URL: "https://example.com/p9/member/frame-01.png"},
	{Key: "frame-02", Name: "默认头像框 2", URL: "https://example.com/p9/member/frame-02.png"},
	{Key: "frame-03", Name: "默认头像框 3", URL: "https://example.com/p9/member/frame-03.png"},
}

type GetCreateOptionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCreateOptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCreateOptionsLogic {
	return &GetCreateOptionsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// GetCreateOptions 返回临时头像与头像框选项。
func (l *GetCreateOptionsLogic) GetCreateOptions(_ *memberpb.GetMemberCreateOptionsRequest) (*memberpb.GetMemberCreateOptionsResponse, error) {
	avatars := make([]*memberpb.MemberImageOption, 0, len(defaultAvatars))
	for _, item := range defaultAvatars {
		avatars = append(avatars, &memberpb.MemberImageOption{Key: item.Key, Name: item.Name, Url: item.URL})
	}
	frames := make([]*memberpb.MemberImageOption, 0, len(defaultAvatarFrames))
	for _, item := range defaultAvatarFrames {
		frames = append(frames, &memberpb.MemberImageOption{Key: item.Key, Name: item.Name, Url: item.URL})
	}
	return &memberpb.GetMemberCreateOptionsResponse{Avatars: avatars, AvatarFrames: frames}, nil
}
