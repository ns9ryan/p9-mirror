package agent

import (
	"context"
	"net"
	"strings"
	"time"

	"oa.98ent.com/p9/common/utils"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/agentuser"
	"oa.98ent.com/p9/operator-crm/pkg/agentline"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"
)

// AddOperatorAgent 分站按指定线添加代理。
func (s *Service) AddOperatorAgent(ctx context.Context, in *agentpb.AddOperatorAgentReq) (*agentpb.AddAgentResp, error) {
	ctx, _, err := requireOperator(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil || in.Agent == nil {
		return nil, xerr.BadRequest(i18n.AgentFieldsRequired)
	}
	// 如果初始化代理，则必须指定代理子线路编码
	if in.IsInit && !agentline.IsValid(in.Agent.AgentLineCode) {
		return nil, xerr.BadRequest(i18n.AgentLineCodeInvalid)
	}
	parentID := in.ParentId
	if parentID <= 0 {
		// 如果初始化代理，获取root代理ID没有则创建root代理
		if in.IsInit {
			parentID, err = s.resolveDesignatedParent(ctx)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, xerr.BadRequest(i18n.AgentParentRequired)
		}
	}
	return s.createUnderParent(ctx, in.Agent, parentID, registerSourceAdmin, in.IsInit)
}

// AddSubAgent 当前登录代理添加直属下级。
func (s *Service) AddSubAgent(ctx context.Context, in *agentpb.AddSubAgentReq) (*agentpb.AddAgentResp, error) {
	ctx, claims, err := requireAgentClaims(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil || in.Agent == nil {
		return nil, xerr.BadRequest(i18n.AgentFieldsRequired)
	}
	return s.createUnderParent(ctx, in.Agent, claims.UserID, registerSourceAgent, false)
}

// 获取root代理ID没有则创建root代理
func (s *Service) resolveDesignatedParent(ctx context.Context) (int64, error) {
	rootAgent, err := s.DB.AgentUser.
		Query().Select(agentuser.FieldID).
		Where(agentuser.LevelEQ(0)).
		Where(agentuser.ParentIDIsNil()).Only(ctx)

	if err != nil {
		if !ent.IsNotFound(err) {
			return 0, err
		}
		return s.createRootAgent(ctx)
	}
	return rootAgent.ID, nil
}

// createRootAgent 创建系统根代理：parent 为空、level=0、top_id is null。
func (s *Service) createRootAgent(ctx context.Context) (int64, error) {
	// 生成随机密码
	password, err := utils.GenRandomPassword(16)
	if err != nil {
		return 0, err
	}
	hash, err := dao.HashPassword(password)
	if err != nil {
		return 0, err
	}
	// 生成随机盐值
	salt, err := utils.GenShortID()
	if err != nil {
		return 0, err
	}
	// 生成代理编码
	agentCode, err := utils.GenLongID()
	if err != nil {
		return 0, err
	}
	// 生成推广码
	promoCode, err := utils.GenShortID()
	if err != nil {
		return 0, err
	}

	now := time.Now()
	var rootID int64
	err = dao.WithTx(ctx, s.DB, func(tx *ent.Tx) error {
		u, cerr := tx.AgentUser.Create().
			SetAccount("root").
			SetCode(strings.ToUpper(agentCode)).
			SetLevel(0).
			SetAgentType(1).
			SetAgentLineCode("ROOT_LINE").
			SetShareRate(maxShareRate).
			SetStatus(statusNormal).
			SetPayoutRestricted(1).
			SetRegisterSource(registerSourceAdmin).
			SetPromoCode(strings.ToUpper(promoCode)).
			SetLoginPasswordHash(hash).
			SetSalt(salt).
			SetRegisterAt(now).
			SetLoginCount(0).
			Save(ctx)
		if cerr != nil {
			return mapEntErr(cerr)
		}
		if _, cerr = tx.AgentProfile.Create().SetAgentID(u.ID).Save(ctx); cerr != nil {
			return mapEntErr(cerr)
		}
		if _, cerr = tx.AgentStats.Create().SetAgentID(u.ID).Save(ctx); cerr != nil {
			return mapEntErr(cerr)
		}
		if _, cerr = tx.AgentClosure.Create().
			SetAncestorID(u.ID).
			SetDescendantID(u.ID).
			SetDepth(0).
			Save(ctx); cerr != nil {
			return mapEntErr(cerr)
		}
		rootID = u.ID
		return nil
	})
	if err != nil {
		return 0, err
	}
	if _, op, oerr := requireOperator(ctx); oerr == nil {
		s.invalidateTree(op)
	}
	return rootID, nil
}

func (s *Service) createUnderParent(ctx context.Context, fields *agentpb.AgentCreateFields, parentID, registerSource int64, isInit bool) (*agentpb.AddAgentResp, error) {
	account := strings.TrimSpace(fields.Account)
	agentCode := strings.TrimSpace(fields.Code)
	password := fields.Password
	lineCode := strings.TrimSpace(fields.AgentLineCode)
	promoCode := strings.TrimSpace(fields.PromoCode)
	remark := strings.TrimSpace(fields.Remark)
	shareRate := int64(fields.ShareRate)
	agentType := int64(fields.AgentType)
	// 如果代理编码为空，则生成代理编码
	if agentCode == "" {
		agentCodeTemp, err := utils.GenLongID()
		if err != nil {
			return nil, err
		}
		agentCode = strings.ToUpper(agentCodeTemp)
	}
	// 如果推广码为空，则生成推广码
	if promoCode == "" {
		promoCodeTemp, err := utils.GenShortID()
		if err != nil {
			return nil, err
		}
		promoCode = strings.ToUpper(promoCodeTemp)
	}

	if account == "" || password == "" {
		return nil, xerr.BadRequest(i18n.AgentFieldsRequired)
	}
	if agentType < 1 || agentType > 2 {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	if shareRate < 0 || shareRate > maxShareRate {
		return nil, xerr.BadRequest(i18n.AgentShareRateInvalid)
	}
	// 获取上级代理用户信息
	parent, err := s.FecthAgentByID(ctx, parentID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, xerr.BadRequest(i18n.AgentParentNotFound)
		}
		return nil, err
	}
	// 获取管理员设置的最大层级
	opMaxLevel, err := s.GetMaxLevel(ctx)
	if err != nil {
		return nil, err
	}
	// 如果非初始化代理，则使用上级代理的子线路编码
	if !isInit {
		// 上级不能是系统根（level=0）
		if parent.Level < 1 {
			return nil, xerr.BadRequest(i18n.AgentParentLevelInvalid)
		}
		lineCode = parent.AgentLineCode
	}
	if parent.Status != statusNormal {
		return nil, xerr.BadRequest(i18n.AgentParentDisabled)
	}
	// 判断是否超过管理员设置的最大层级
	if parent.Level >= opMaxLevel {
		return nil, xerr.BadRequest(i18n.AgentLevelExceeded)
	}
	// 占层比例不能超过上级代理
	if shareRate > parent.ShareRate {
		return nil, xerr.BadRequest(i18n.AgentShareRateExceedParent)
	}
	// 生成密码哈希
	hash, err := dao.HashPassword(password)
	if err != nil {
		return nil, err
	}
	// 生成随机盐值
	salt, err := utils.GenShortID()
	if err != nil {
		return nil, err
	}

	var registerIP *string
	if ip := utils.NormalizeIP(strings.TrimSpace(fields.RegisterIp)); net.ParseIP(ip) != nil {
		registerIP = &ip
	}
	now := time.Now()
	level := parent.Level + 1
	// 上级为顶层（top_id 为空）时，下级的 top_id 指向该上级；否则继承上级 top_id。
	topID := parent.ID
	if parent.TopID != nil {
		topID = *parent.TopID
	}

	var created *ent.AgentUser
	err = dao.WithTx(ctx, s.DB, func(tx *ent.Tx) error {
		c := tx.AgentUser.Create().
			SetAccount(account).
			SetCode(agentCode).
			SetParentID(parent.ID).
			SetTopID(topID).
			SetLevel(level).
			SetAgentType(agentType).
			SetAgentLineCode(lineCode).
			SetShareRate(shareRate).
			SetStatus(statusNormal).
			SetPayoutRestricted(1).
			SetRegisterSource(registerSource).
			SetPromoCode(promoCode).
			SetLoginPasswordHash(hash).
			SetSalt(salt).
			SetRegisterAt(now).
			SetLoginCount(0).
			SetRemark(remark)
		if registerIP != nil {
			c.SetRegisterIP(*registerIP)
		}
		u, cerr := c.Save(ctx)
		if cerr != nil {
			return mapEntErr(cerr)
		}
		// 插入代理用户信息
		profileTx := tx.AgentProfile.Create().SetAgentID(u.ID)
		if fields.Mobile != nil {
			profileTx.SetMobile(*fields.Mobile)
		}
		if fields.Email != nil {
			profileTx.SetEmail(*fields.Email)
		}
		if fields.Realname != nil {
			profileTx.SetRealname(*fields.Realname)
		}
		if _, cerr = profileTx.Save(ctx); cerr != nil {
			return mapEntErr(cerr)
		}
		// 插入代理统计信息
		if _, cerr = tx.AgentStats.Create().SetAgentID(u.ID).Save(ctx); cerr != nil {
			return mapEntErr(cerr)
		}
		// 插入闭包
		if cerr = insertClosures(ctx, tx, u.ID, parent.ID); cerr != nil {
			return mapEntErr(cerr)
		}
		if cerr = bumpAgentCounts(ctx, tx, parent.ID); cerr != nil {
			return cerr
		}
		created = u
		return nil
	})
	if err != nil {
		// 记录操作日志
		s.RecordOpFail(ctx, &RecordOpParams{
			targetType:    opTargetTypeAgent,
			targetID:      0,
			targetAccount: account,
			action:        opActionCreate,
			before:        "",
			after:         account,
		}, err)
		return nil, err
	}
	// 记录操作日志
	s.RecordOpOK(ctx, &RecordOpParams{
		targetType:    opTargetTypeAgent,
		targetID:      created.ID,
		targetAccount: account,
		action:        opActionCreate,
		before:        "",
		after:         account,
	})
	// 创建代理后，刷新内存树缓存
	if _, op, oerr := requireOperator(ctx); oerr == nil {
		s.invalidateTree(op)
	}
	resp := &agentpb.AddAgentResp{
		Id:        created.ID,
		Account:   created.Account,
		Code:      created.Code,
		ParentId:  parent.ID,
		Level:     int32(created.Level),
		ShareRate: int32(created.ShareRate),
	}
	if created.TopID != nil {
		resp.TopId = *created.TopID
	}
	return resp, nil
}
