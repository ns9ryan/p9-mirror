package dao

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	
	"entgo.io/ent/dialect/sql"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/agentclosure"
	"oa.98ent.com/p9/operator-crm/ent/agentstats"
	"oa.98ent.com/p9/operator-crm/ent/agentuser"
	"oa.98ent.com/p9/operator-crm/ent/member"
	"oa.98ent.com/p9/operator-crm/ent/memberautolevel"
	"oa.98ent.com/p9/operator-crm/ent/memberviplevel"
)

// Member 负责会员开户及其关联数据。
type Member struct {
	db *ent.Client // Ent 数据库客户端
}

// CreateInput 是创建会员所需的业务字段。
type CreateInput struct {
	Account        string
	PasswordHash   string
	Nickname       string
	AgentID        int64
	AvatarURL      string
	AvatarFrameURL string
	Remark         string
}

// Create 创建会员，并在同一事务中初始化认证、资料和VIP进度。
func (d *Member) Create(ctx context.Context, in CreateInput) (*ent.Member, error) {
	var created *ent.Member

	err := WithTx(ctx, d.db, func(tx *ent.Tx) error {
		// 代理、自动层级和 VIP 等级都必须属于当前厅。
		agent, err := tx.AgentUser.Query().
			Where(agentuser.IDEQ(in.AgentID), agentuser.StatusEQ(1)).
			Only(ctx)
		if err != nil {
			return err
		}

		autoLevel, err := tx.MemberAutoLevel.Query().
			Where(memberautolevel.MinDepositEQ(0), memberautolevel.MinWaterEQ(0)).
			Order(memberautolevel.ByID(sql.OrderAsc())).
			First(ctx)
		if err != nil {
			return err
		}

		vipLevel, err := tx.MemberVipLevel.Query().
			Where(memberviplevel.LevelEQ(0)).
			Only(ctx)
		if err != nil {
			return err
		}

		code, err := newMemberCode(ctx, tx)
		if err != nil {
			return err
		}
		inviteCode, err := randomHex(8)
		if err != nil {
			return err
		}

		created, err = tx.Member.Create().
			SetCode(code).
			SetAgentID(agent.ID).
			SetInviteCode(inviteCode).
			SetAutoLevelID(autoLevel.ID).
			SetVipLevelID(vipLevel.ID).
			SetRegisterMethod("admin").
			SetRegisterChannel("operator_admin").
			Save(ctx)
		if err != nil {
			return err
		}

		if _, err = tx.MemberAuth.Create().
			SetMemberID(created.ID).
			SetUsername(in.Account).
			SetLoginPasswordHash(in.PasswordHash).
			Save(ctx); err != nil {
			return err
		}

		profile := tx.MemberProfile.Create().
			SetMemberID(created.ID).
			SetAvatarURL(in.AvatarURL).
			SetAvatarFrameURL(in.AvatarFrameURL).
			SetNickname(in.Nickname)
		if in.Remark != "" {
			profile.SetRemark(in.Remark)
		}
		if _, err = profile.Save(ctx); err != nil {
			return err
		}

		if _, err = tx.MemberVipProgress.Create().
			SetMemberID(created.ID).
			Save(ctx); err != nil {
			return err
		}

		return incrementAgentMemberCounts(ctx, tx, agent.ID)
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// newMemberCode 生成当前厅内唯一的12位会员编号。
func newMemberCode(ctx context.Context, tx *ent.Tx) (string, error) {
	for range 5 {
		n, err := rand.Int(rand.Reader, big.NewInt(900_000_000_000))
		if err != nil {
			return "", err
		}
		code := fmt.Sprintf("%012d", n.Int64()+100_000_000_000)
		exists, err := tx.Member.Query().Where(member.CodeEQ(code)).Exist(ctx)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}
	return "", fmt.Errorf("generate unique member code: retries exhausted")
}

func randomHex(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// incrementAgentMemberCounts 更新代理直属会员数和代理链总会员数。
func incrementAgentMemberCounts(ctx context.Context, tx *ent.Tx, agentID int64) error {
	if _, err := tx.AgentStats.Update().
		Where(agentstats.AgentIDEQ(agentID)).
		AddDirectMemberCount(1).
		Save(ctx); err != nil {
		return err
	}

	ancestors, err := tx.AgentClosure.Query().
		Where(agentclosure.DescendantIDEQ(agentID)).
		All(ctx)
	if err != nil {
		return err
	}

	ids := make([]int64, 0, len(ancestors))
	for _, item := range ancestors {
		ids = append(ids, item.AncestorID)
	}
	if len(ids) == 0 {
		ids = append(ids, agentID)
	}
	_, err = tx.AgentStats.Update().
		Where(agentstats.AgentIDEQ(ids...)).
		AddTotalMemberCount(1).
		Save(ctx)
	return err
}

// OperatorContext 校验可信厅身份并将厅编码绑定到 Ent 上下文。
func OperatorContext(ctx context.Context) (context.Context, string, error) {
	claims := ctxdata.ClaimsFromCtx(ctx)
	if claims == nil || strings.TrimSpace(claims.OperatorCode) == "" || strings.TrimSpace(claims.OperatorCode) != claims.OperatorCode {
		return ctx, "", fmt.Errorf("missing trusted operator identity")
	}
	return ctxdata.WithOperatorCode(ctx, claims.OperatorCode), claims.OperatorCode, nil
}

