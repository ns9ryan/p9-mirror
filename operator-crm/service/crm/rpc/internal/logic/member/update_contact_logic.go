package memberlogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateContactLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateContactLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateContactLogic {
	return &UpdateContactLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateContactLogic) UpdateContact(in *memberpb.UpdateMemberContactRequest) (*memberpb.MemberOperationResponse, error) {
	if in.AvatarId != nil || in.AvatarFrameId != nil {
		images, err := NewListImagesLogic(l.ctx, l.svcCtx).ListImages(&memberpb.ListMemberImagesRequest{})
		if err != nil {
			return nil, err
		}
		valid := func(id int64, list []*memberpb.MemberImageResource) bool {
			for _, v := range list {
				if id > 0 && v.Id == id {
					return true
				}
			}
			return false
		}
		if (in.AvatarId != nil && !valid(*in.AvatarId, images.Avatars)) || (in.AvatarFrameId != nil && !valid(*in.AvatarFrameId, images.AvatarFrames)) {
			return nil, xerr.RpcErr(xerr.BadRequest("common.invalidParam"))
		}
	}

	// 更新前获取原始数据
	beforeProfile, beforeAuth, beforeProfileErr, beforeAuthErr := l.svcCtx.Member.GetMemberContact(l.ctx, in.MemberId)
	if beforeProfileErr != nil {
		l.Errorf("获取更新前的会员资料失败: member_id=%d, err=%v", in.MemberId, beforeProfileErr)
	}
	if beforeAuthErr != nil {
		l.Errorf("获取更新前的会员认证信息失败: member_id=%d, err=%v", in.MemberId, beforeAuthErr)
	}

	// 执行更新
	if err := l.svcCtx.Member.UpdateContact(l.ctx, in, l.svcCtx.Cryptor); err != nil {
		return nil, xerr.RpcErr(err)
	}

	// 更新后获取新数据
	afterProfile, afterAuth, afterProfileErr, afterAuthErr := l.svcCtx.Member.GetMemberContact(l.ctx, in.MemberId)

	// 发送会员资料更新事件
	if afterProfileErr != nil {
		l.Errorf("获取更新后的会员资料失败: member_id=%d, err=%v", in.MemberId, afterProfileErr)
	} else {
		l.svcCtx.KafkaEventHandler.PublishMemberProfileUpdatedEvent(l.ctx, beforeProfile, afterProfile)
	}

	// 发送会员认证信息更新事件
	if afterAuthErr != nil {
		l.Errorf("获取更新后的会员认证信息失败: member_id=%d, err=%v", in.MemberId, afterAuthErr)
	} else {
		l.svcCtx.KafkaEventHandler.PublishMemberAuthUpdatedEvent(
			l.ctx,
			in.MemberId,
			beforeProfile.OperatorCode,
			afterProfile.UpdatedAt,
			beforeAuth,
			afterAuth,
		)
	}

	return &memberpb.MemberOperationResponse{}, nil
}
