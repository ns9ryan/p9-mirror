package agent

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/agentuser"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
)

// LoadAgent 加载代理用户及Profile信息
func (s *Service) LoadAgent(ctx context.Context, id int64) (*ent.AgentUser, error) {
	if id <= 0 {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	ctx, _, err := requireOperator(ctx)
	if err != nil {
		return nil, err
	}
	// 获取代理用户及其Profile
	row, err := s.FecthAgentWithProfileByID(ctx, id)
	if ent.IsNotFound(err) {
		return nil, xerr.NotFound(i18n.AgentNotFound)
	}
	return row, err
}

// FecthAllAgents 获取所有代理用户
func (s *Service) FecthAllAgents(ctx context.Context, fields ...string) ([]*ent.AgentUser, error) {
	return s.DB.AgentUser.Query().
		Select(fields...).
		All(ctx)
}

// FecthAgentByID 根据代理ID获取代理用户
func (s *Service) FecthAgentByID(ctx context.Context, id int64, fields ...string) (*ent.AgentUser, error) {
	return s.DB.AgentUser.Query().
		Select(fields...).
		Where(agentuser.IDEQ(id)).
		Only(ctx)
}

// FecthAgentByAccount 根据代理账号获取代理用户
func (s *Service) FecthAgentByAccount(ctx context.Context, account string, fields ...string) (*ent.AgentUser, error) {
	return s.DB.AgentUser.Query().
		Select(fields...).
		Where(agentuser.AccountEQ(account)).
		Only(ctx)
}

// FecthAgentWithProfileByID 根据代理ID获取代理用户及其Profile
func (s *Service) FecthAgentWithProfileByID(ctx context.Context, id int64) (*ent.AgentUser, error) {
	return s.DB.AgentUser.Query().
		Where(agentuser.IDEQ(id)).
		WithProfile().
		Only(ctx)
}

// FecthAgentWithProfileByAccount 根据代理账号获取代理用户及其Profile
func (s *Service) FecthAgentWithProfileByAccount(ctx context.Context, account string) (*ent.AgentUser, error) {
	return s.DB.AgentUser.Query().
		Where(agentuser.AccountEQ(account)).
		WithProfile().
		Only(ctx)
}

// CountAgentByID 统计代理用户数量
func (s *Service) CountAgentByID(ctx context.Context, ID int64) (int, error) {
	return s.DB.AgentUser.Query().
		Where(agentuser.IDEQ(ID)).
		Count(ctx)
}

// CountAgentByParentID 统计父级代理用户数量
func (s *Service) CountAgentByParentID(ctx context.Context, parentID int64) (int, error) {
	return s.DB.AgentUser.Query().
		Where(agentuser.ParentIDEQ(parentID)).
		Count(ctx)
}

// CheckAgentExistByID 检查代理用户ID是否存在
func (s *Service) CheckAgentExistByID(ctx context.Context, ID int64) (bool, error) {
	return s.DB.AgentUser.Query().
		Where(agentuser.IDEQ(ID)).
		Exist(ctx)
}

// CheckAgentExistByAccount 检查代理账号是否存在
func (s *Service) CheckAgentExistByAccount(ctx context.Context, account string) (bool, error) {
	return s.DB.AgentUser.Query().
		Where(agentuser.AccountEQ(account)).
		Exist(ctx)
}
