package crm

import (
	"context"

	"oa.98ent.com/p9/operator-activity/modules/activity/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"
)

// LoadMemberInfo 加载会员信息
func (s *CrmServ) LoadMember(ctx context.Context, memberID int64) (*types.MemberInfo, error) {
	resp, err := s.memberRpc.GetBasicInfo(ctx, &memberpb.GetMemberContextRequest{MemberId: memberID})
	if err != nil {
		return nil, err
	}
	// 会员信息
	memberInfo := &types.MemberInfo{
		ID:           memberID,
		OperatorCode: resp.OperatorCode,
		FixedLevels:  resp.FixedLevels,
		AutoLevel:    resp.AutoLevel,
		AgentID:      resp.AgentId,
	}
	// 获取推荐人ID
	if resp.InviterMemberId != nil {
		memberInfo.InviterID = *resp.InviterMemberId
	}
	// 注册IP
	if resp.RegisterIp != nil {
		memberInfo.RegisterIP = *resp.RegisterIp
	}
	// 注册设备码
	if resp.RegisterDeviceCode != nil {
		memberInfo.RegisterUdid = *resp.RegisterDeviceCode
	}

	return memberInfo, nil
}
