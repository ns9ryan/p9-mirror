package member

import (
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"
)

func CreateMemberFromPB(in *memberpb.CreateMemberResponse) *types.CreateMemberResp {
	if in == nil {
		return &types.CreateMemberResp{}
	}
	return &types.CreateMemberResp{
		Id:        in.Id,
		Code:      in.Code,
		CreatedAt: in.CreatedAt,
	}
}

func CreateOptionsFromPB(in *memberpb.GetMemberCreateOptionsResponse) *types.GetMemberCreateOptionsResp {
	out := &types.GetMemberCreateOptionsResp{
		Avatars:      make([]types.MemberImageOption, 0),
		AvatarFrames: make([]types.MemberImageOption, 0),
	}
	if in == nil {
		return out
	}
	for _, item := range in.GetAvatars() {
		if item != nil {
			out.Avatars = append(out.Avatars, types.MemberImageOption{Key:item.GetKey(), Name:item.GetName(), Url:item.GetUrl()})
		}
	}
	for _, item := range in.GetAvatarFrames() {
		if item != nil {
			out.AvatarFrames = append(out.AvatarFrames, types.MemberImageOption{Key:item.GetKey(), Name:item.GetName(), Url:item.GetUrl()})
		}
	}
	return out
}
