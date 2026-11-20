// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package manage

import (
	"context"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListImagesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListImagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListImagesLogic {
	return &ListImagesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListImagesLogic) ListImages() (resp *types.ListMemberImagesResponse, err error) {
	out, err := l.svcCtx.MemberRpc.ListImages(l.ctx, &memberpb.ListMemberImagesRequest{})
	if err != nil {
		return nil, err
	}
	resp = &types.ListMemberImagesResponse{Avatars: []types.MemberImageResource{}, AvatarFrames: []types.MemberImageResource{}}
	for _, v := range out.Avatars {
		resp.Avatars = append(resp.Avatars, types.MemberImageResource{Id: v.Id, ImageUrl: v.ImageUrl})
	}
	for _, v := range out.AvatarFrames {
		resp.AvatarFrames = append(resp.AvatarFrames, types.MemberImageResource{Id: v.Id, ImageUrl: v.ImageUrl})
	}
	return resp, nil
}
