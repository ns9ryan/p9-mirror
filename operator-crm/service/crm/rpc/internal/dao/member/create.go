package member

import (
	"context"
	"fmt"

	"oa.98ent.com/p9/operator-crm/ent"
	entmember "oa.98ent.com/p9/operator-crm/ent/member"
	"oa.98ent.com/p9/operator-crm/ent/memberauth"
)

// CreateParams 会员创建参数
type CreateParams struct {
	// 会员标识与归属
	Code            string // 12位会员编号
	InviteCode      string // 会员邀请码
	InviterMemberID *int64 // 邀请人会员ID
	AgentID         int64  // 所属分层代理ID

	// 登录认证
	Account         *string // 登录账号
	PasswordHash    *string // 登录密码哈希
	PhoneCiphertext *string // 手机号密文
	PhoneLookupHash *string // 手机号检索摘要, 使用HMAC-SHA256生成
	EmailCiphertext *string // 邮箱密文
	EmailLookupHash *string // 邮箱检索摘要, 使用HMAC-SHA256生成

	// 个人资料
	Nickname       string  // 会员昵称
	AvatarURL      string  // 头像地址, 使用系统预设资源
	AvatarFrameURL string  // 头像框地址, 使用系统预设资源
	Remark         *string // 运营备注

	// 注册来源
	RegisterMethod     string  // 注册方式
	RegisterChannel    string  // 注册渠道
	RegisterIP         *string // 注册IP
	RegisterDeviceCode *string // 注册设备标识
	RegisterDomain     *string // 注册域名
}

// CodeExistsTx 检查当前厅的会员编号是否存在
func (*Service) CodeExistsTx(ctx context.Context, tx *ent.Tx, code string) (bool, error) {
	return tx.Member.Query().
		Where(entmember.CodeEQ(code)).
		Exist(ctx)
}

// InviteCodeExistsTx 检查当前厅的会员邀请码是否存在
func (*Service) InviteCodeExistsTx(ctx context.Context, tx *ent.Tx, code string) (bool, error) {
	return tx.Member.Query().
		Where(entmember.InviteCodeEQ(code)).
		Exist(ctx)
}

// UsernameExistsTx 检查当前厅的会员登录账号是否存在
func (*Service) UsernameExistsTx(ctx context.Context, tx *ent.Tx, account string) (bool, error) {
	return tx.MemberAuth.Query().
		Where(memberauth.UsernameEQ(account)).
		Exist(ctx)
}

// CreateTx 在事务中创建会员及初始业务记录
func (s *Service) CreateTx(ctx context.Context, tx *ent.Tx, in CreateParams) (*ent.Member, error) {
	// 获取默认自动层级
	autoLevel, err := s.autoLevel.GetDefaultTx(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("查询默认自动层级失败: %w", err)
	}

	// 获取默认VIP等级
	vipLevel, err := s.vipLevel.GetDefaultTx(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("查询默认VIP等级失败: %w", err)
	}

	// 创建会员主体
	m, err := tx.Member.Create().
		SetCode(in.Code).                                     // 会员对外编号, 12位数字
		SetAgentID(in.AgentID).                               // 会员所属分层代理ID
		SetNillableInviterMemberID(in.InviterMemberID).       // 邀请当前会员注册的会员ID
		SetInviteCode(in.InviteCode).                         // 会员专属邀请码
		SetAutoLevelID(autoLevel.ID).                         // 当前自动层级ID
		SetVipLevelID(vipLevel.ID).                           // 当前VIP等级ID
		SetRegisterMethod(in.RegisterMethod).                 // 注册方式
		SetRegisterChannel(in.RegisterChannel).               // 注册渠道
		SetNillableRegisterIP(in.RegisterIP).                 // 注册IP
		SetNillableRegisterDeviceCode(in.RegisterDeviceCode). // 注册设备标识
		SetNillableRegisterDomain(in.RegisterDomain).         // 注册域名
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建会员主体失败: %w", err)
	}

	// 创建登录认证信息
	if _, err := tx.MemberAuth.Create().
		SetMemberID(m.ID).                              // 所属会员ID
		SetNillableUsername(in.Account).                // 会员登录账号, 保留原始大小写
		SetNillablePhoneCiphertext(in.PhoneCiphertext). // 手机号加密密文
		SetNillablePhoneLookupHash(in.PhoneLookupHash). // 手机号检索摘要
		SetNillableEmailCiphertext(in.EmailCiphertext). // 邮箱加密密文
		SetNillableEmailLookupHash(in.EmailLookupHash). // 邮箱检索摘要
		SetNillableLoginPasswordHash(in.PasswordHash).  // 登录密码单向哈希
		Save(ctx); err != nil {
		return nil, fmt.Errorf("创建会员认证信息失败: %w", err)
	}

	// 创建个人资料
	if _, err := tx.MemberProfile.Create().
		SetMemberID(m.ID).                    // 所属会员ID
		SetAvatarURL(in.AvatarURL).           // 会员头像URL, 使用系统预设资源
		SetAvatarFrameURL(in.AvatarFrameURL). // 会员头像框URL, 使用系统预设资源
		SetNickname(in.Nickname).             // 会员昵称
		SetNillableRemark(in.Remark).         // 后台会员备注
		Save(ctx); err != nil {
		return nil, fmt.Errorf("创建会员个人资料失败: %w", err)
	}

	// 初始化VIP晋级进度
	if _, err := tx.MemberVipProgress.Create().
		SetMemberID(m.ID). // 所属会员ID
		Save(ctx); err != nil {
		return nil, fmt.Errorf("创建会员VIP晋级进度失败: %w", err)
	}

	// 会员统计由统计业务维护, 创建时不生成统计记录
	return m, nil
}
