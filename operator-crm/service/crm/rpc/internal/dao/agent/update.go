package agent

import (
	"context"
	"strings"
	"unicode/utf8"

	"entgo.io/ent/dialect/sql"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/utils"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/agentuser"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"
)

const (
	payoutNormal          int64 = 1 // 不限制出款
	payoutOwnerRestricted int64 = 2 // 限制本人出款
	payoutLineRestricted  int64 = 3 // 限制全线出款
)

// 重置登录密码 随机生成密码并返回
func (s *Service) ResetLoginPassword(ctx context.Context, req *agentpb.ResetLoginPasswordReq) (*agentpb.ResetLoginPasswordResp, error) {
	if req == nil {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	agentID := req.GetAgentId()
	// 操作日志参数
	opParams := &RecordOpParams{
		targetType: opTargetTypeAgent,
		targetID:   agentID,
		action:     opActionResetLoginPassword,
		after:      opPasswordMask,
		before:     opPasswordMask,
	}
	// 加载代理
	row, err := s.LoadAgent(ctx, agentID)
	if err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	// 操作日志参数 代理账号
	opParams.targetAccount = row.Account
	// 生成随机密码
	plain, err := utils.GenRandomPassword(16)
	if err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	hashed, err := dao.HashPassword(plain)
	if err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	if err = s.DB.AgentUser.UpdateOneID(row.ID).
		SetLoginPasswordHash(hashed).
		Exec(ctx); err != nil {
		err = mapEntErr(err)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	s.RecordOpOK(ctx, opParams)
	return &agentpb.ResetLoginPasswordResp{Password: plain}, nil
}

// UpdateLoginPassword 当前代理修改登录密码。校验旧密码后只更新哈希，不改令牌盐。
func (s *Service) UpdateLoginPassword(ctx context.Context, req *agentpb.UpdateLoginPasswordReq) (*agentpb.Empty, error) {
	if req == nil {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	oldPw := strings.TrimSpace(req.GetOldPassword())
	newPw := strings.TrimSpace(req.GetNewPassword())
	if oldPw == "" || newPw == "" {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	c := ctxdata.ClaimsFromCtx(ctx)
	if c == nil || c.UserID <= 0 || strings.TrimSpace(c.OperatorCode) == "" || c.Salt == "" {
		return nil, xerr.Unauthorized(i18n.Unauthorized)
	}
	agentID := c.UserID
	// 操作日志参数
	opParams := &RecordOpParams{
		targetType: opTargetTypeAgent,
		targetID:   agentID,
		action:     opActionUpdateLoginPassword,
		after:      opPasswordMask,
		before:     opPasswordMask,
	}
	// 加载代理
	row, err := s.LoadAgent(ctx, agentID)
	if err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	// 操作日志参数 代理账号
	opParams.targetAccount = row.Account
	// 校验代理状态和盐值
	if row.Status != statusNormal || row.Salt != c.Salt {
		err = xerr.Unauthorized(i18n.Unauthorized)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	if !dao.CheckPassword(row.LoginPasswordHash, oldPw) {
		err = xerr.Unauthorized(i18n.AuthPasswordIncorrect)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	hashed, err := dao.HashPassword(newPw)
	if err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	if err = s.DB.AgentUser.UpdateOneID(row.ID).
		SetLoginPasswordHash(hashed).
		Exec(ctx); err != nil {
		err = mapEntErr(err)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	s.RecordOpOK(ctx, opParams)
	return &agentpb.Empty{}, nil
}

// 重置提现密码
func (s *Service) UpdateWithdrawPassword(ctx context.Context, req *agentpb.UpdateWithdrawPasswordReq) (*agentpb.Empty, error) {
	if req == nil || strings.TrimSpace(req.GetPassword()) == "" {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	agentID := req.GetAgentId()
	// 操作日志参数
	opParams := &RecordOpParams{
		targetType: opTargetTypeAgent,
		targetID:   agentID,
		action:     opActionUpdateWithdrawPassword,
		after:      opPasswordMask,
		before:     opPasswordMask,
	}
	// 加载代理
	row, err := s.LoadAgent(ctx, agentID)
	if err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	// 操作日志参数 代理账号
	opParams.targetAccount = row.Account
	// 生成提现密码哈希
	hashed, err := dao.HashPassword(strings.TrimSpace(req.GetPassword()))
	if err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	if err = s.DB.AgentUser.UpdateOneID(row.ID).
		SetWithdrawPasswordHash(hashed).
		Exec(ctx); err != nil {
		err = mapEntErr(err)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	s.RecordOpOK(ctx, opParams)
	return &agentpb.Empty{}, nil
}

// 更新代理占成比例
func (s *Service) UpdateShareRate(ctx context.Context, req *agentpb.UpdateAgentShareRateReq) (*agentpb.Empty, error) {
	if req == nil {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	agentID := req.GetAgentId()
	share := int64(req.GetShareRate())
	// 操作日志参数
	opParams := &RecordOpParams{
		targetType: opTargetTypeAgent,
		targetID:   agentID,
		action:     opActionShareRate,
		after:      formatInt64(share),
	}
	if share < 0 || share > maxShareRate {
		err := xerr.BadRequest(i18n.AgentShareRateInvalid)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	row, err := s.LoadAgent(ctx, agentID)
	if err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	// 操作日志参数 更新前占成比例
	opParams.before = formatInt64(row.ShareRate)
	// 操作日志参数 代理账号
	opParams.targetAccount = row.Account
	if row.ParentID != nil {
		parent, err := s.FecthAgentByID(ctx, *row.ParentID)
		if err != nil {
			if ent.IsNotFound(err) {
				err = xerr.BadRequest(i18n.AgentParentNotFound)
			}
			s.RecordOpFail(ctx, opParams, err)
			return nil, err
		}
		if share > parent.ShareRate {
			err = xerr.BadRequest(i18n.AgentShareRateExceedParent)
			s.RecordOpFail(ctx, opParams, err)
			return nil, err
		}
	}
	var childMax []struct {
		Max sql.NullInt64 `json:"max"`
	}
	if err = s.DB.AgentUser.Query().
		Where(agentuser.ParentIDEQ(row.ID)).
		Aggregate(ent.As(ent.Max(agentuser.FieldShareRate), "max")).
		Scan(ctx, &childMax); err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	if len(childMax) > 0 && childMax[0].Max.Valid && share < childMax[0].Max.Int64 {
		err = xerr.BadRequest(i18n.AgentShareRateBelowChild)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	if err = s.DB.AgentUser.UpdateOneID(row.ID).
		SetShareRate(share).
		Exec(ctx); err != nil {
		err = mapEntErr(err)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	if _, op, oerr := requireOperator(ctx); oerr == nil {
		s.invalidateTree(op)
	}
	s.RecordOpOK(ctx, opParams)
	return &agentpb.Empty{}, nil
}

// 更新代理提现限制
func (s *Service) UpdatePayoutRestricted(ctx context.Context, req *agentpb.UpdateAgentPayoutRestrictedReq) (*agentpb.Empty, error) {
	if req == nil {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	agentID := req.GetAgentId()
	v := int64(req.GetPayoutRestricted())
	// 操作日志参数
	opParams := &RecordOpParams{
		targetType: opTargetTypeAgent,
		targetID:   agentID,
		action:     opActionPayoutRestricted,
		after:      formatInt64(v),
	}
	if v < payoutNormal || v > payoutLineRestricted {
		err := xerr.BadRequest(i18n.InvalidParam)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	row, err := s.LoadAgent(ctx, agentID)
	if err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	// 操作日志参数 更新前提现限制
	opParams.before = formatInt64(row.PayoutRestricted)
	// 操作日志参数 代理账号
	opParams.targetAccount = row.Account
	if err = s.DB.AgentUser.UpdateOneID(row.ID).
		SetPayoutRestricted(v).
		Exec(ctx); err != nil {
		err = mapEntErr(err)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	s.RecordOpOK(ctx, opParams)
	return &agentpb.Empty{}, nil
}

// UpdateStatus 更新代理账号状态：1正常 2冻结 3禁用。
func (s *Service) UpdateStatus(ctx context.Context, req *agentpb.UpdateAgentStatusReq) (*agentpb.Empty, error) {
	if req == nil {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	agentID := req.GetAgentId()
	v := int64(req.GetStatus())
	// 操作日志参数
	opParams := &RecordOpParams{
		targetType: opTargetTypeAgent,
		targetID:   agentID,
		action:     opActionStatus,
		after:      formatInt64(v),
	}
	if v < statusNormal || v > statusDisabled {
		err := xerr.BadRequest(i18n.InvalidParam)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	row, err := s.LoadAgent(ctx, agentID)
	if err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	// 操作日志参数 更新前状态
	opParams.before = formatInt64(row.Status)
	// 操作日志参数 代理账号
	opParams.targetAccount = row.Account
	// 更新代理状态
	if err = s.DB.AgentUser.UpdateOneID(row.ID).
		SetStatus(v).
		Exec(ctx); err != nil {
		err = mapEntErr(err)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	s.RecordOpOK(ctx, opParams)
	return &agentpb.Empty{}, nil
}

// 更新代理备注
func (s *Service) UpdateRemark(ctx context.Context, req *agentpb.UpdateAgentRemarkReq) (*agentpb.Empty, error) {
	if req == nil {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	agentID := req.GetAgentId()
	remark := strings.TrimSpace(req.GetRemark())
	// 操作日志参数
	opParams := &RecordOpParams{
		targetType: opTargetTypeAgent,
		targetID:   agentID,
		action:     opActionRemark,
		after:      remark,
	}
	if utf8.RuneCountInString(remark) > 1000 {
		err := xerr.BadRequest(i18n.InvalidParam)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	row, err := s.LoadAgent(ctx, agentID)
	if err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	// 操作日志参数 更新前备注
	opParams.before = ptrStr(row.Remark)
	// 操作日志参数 代理账号
	opParams.targetAccount = row.Account
	// 更新代理备注
	upd := s.DB.AgentUser.UpdateOneID(row.ID)
	if remark == "" {
		upd.ClearRemark()
	} else {
		upd.SetRemark(remark)
	}
	if err = upd.Exec(ctx); err != nil {
		err = mapEntErr(err)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	s.RecordOpOK(ctx, opParams)
	return &agentpb.Empty{}, nil
}

// TouchLogin 更新代理登录信息
func (s *Service) TouchLogin(ctx context.Context, u *ent.AgentUser) error {
	upd := s.DB.AgentUser.UpdateOneID(u.ID).
		AddLoginCount(1)
	// 更新登录IP
	if u.LastLoginIP != nil {
		upd.SetLastLoginIP(*u.LastLoginIP)
	}
	// 更新登录UDID
	if u.LastLoginUdid != nil {
		upd.SetLastLoginUdid(*u.LastLoginUdid)
	}
	// 更新登录时间
	if u.LastLoginAt != nil {
		upd.SetLastLoginAt(*u.LastLoginAt)
	}

	return upd.Exec(ctx)
}

// RotateSalt 旋转重置代理用户盐值
func (s *Service) RotateSalt(ctx context.Context, userID int64) error {
	salt, err := utils.GenShortID()
	if err != nil {
		return err
	}
	n, err := s.DB.AgentUser.Update().
		Where(agentuser.ID(userID)).
		SetSalt(salt).
		Save(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return xerr.NotFound(i18n.AuthUserNotFound)
	}
	return nil
}
