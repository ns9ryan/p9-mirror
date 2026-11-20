package memberlogic

import (
	"context"
	"fmt"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/defaultavatarframepb"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/defaultavatarpb"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListImagesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListImagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListImagesLogic {
	return &ListImagesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListImagesLogic) ListImages(in *memberpb.ListMemberImagesRequest) (*memberpb.ListMemberImagesResponse, error) {
	if _, ok := ctxdata.OperatorClaimsFromCtx(l.ctx); !ok {
		return nil, xerr.RpcErr(xerr.Unauthorized("common.unauthorized"))
	}
	if l.svcCtx.Avatars == nil || l.svcCtx.AvatarFrames == nil {
		return nil, fmt.Errorf("platform base RPC is not configured")
	}
	avatars, err := l.svcCtx.Avatars.ListAll(l.ctx, &defaultavatarpb.ListAllDefaultAvatarsRequest{})
	if err != nil {
		return nil, err
	}
	frames, err := l.svcCtx.AvatarFrames.ListAll(l.ctx, &defaultavatarframepb.ListAllDefaultAvatarFramesRequest{})
	if err != nil {
		return nil, err
	}
	out := &memberpb.ListMemberImagesResponse{Avatars: []*memberpb.MemberImageResource{}, AvatarFrames: []*memberpb.MemberImageResource{}}
	for _, v := range avatars.GetList() {
		if v.GetImageUrl() != "" {
			out.Avatars = append(out.Avatars, &memberpb.MemberImageResource{Id: v.GetId(), ImageUrl: v.GetImageUrl()})
		}
	}
	for _, v := range frames.GetList() {
		if v.GetImageUrl() != "" {
			out.AvatarFrames = append(out.AvatarFrames, &memberpb.MemberImageResource{Id: v.GetId(), ImageUrl: v.GetImageUrl()})
		}
	}
	return out, nil
}
