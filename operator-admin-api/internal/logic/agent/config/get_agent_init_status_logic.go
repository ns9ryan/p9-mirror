package config

import (
	"context"

	agentlogic "oa.98ent.com/p9/operator-admin-api/internal/logic/agent"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

const minInitLevelCount = 3

type GetAgentInitStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAgentInitStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAgentInitStatusLogic {
	return &GetAgentInitStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAgentInitStatusLogic) GetAgentInitStatus() (*types.AgentInitStatusResp, error) {
	codes, err := agentlogic.OpenedLineCodes(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	listed, err := l.svcCtx.AgentRpc.ListOperatorAgents(l.ctx, &agentpb.ListOperatorAgentsReq{Level: 1})
	if err != nil {
		return nil, err
	}
	opened := make(map[string]struct{}, len(codes))
	for _, code := range codes {
		opened[code] = struct{}{}
	}
	hasAgent := make(map[string]struct{}, len(codes))
	for _, group := range listed.GetGroups() {
		if group == nil || len(group.GetAgents()) == 0 {
			continue
		}
		if _, ok := opened[group.GetLineCode()]; !ok {
			continue
		}
		hasAgent[group.GetLineCode()] = struct{}{}
	}

	levels, err := l.svcCtx.AgentRpc.ListAgentLevels(l.ctx, &agentpb.ListAgentLevelsReq{})
	if err != nil {
		return nil, err
	}
	levelCount := int32(len(levels.GetItems()))
	levelsReady := levelCount >= minInitLevelCount

	lines := make([]types.LineInitStatus, 0, len(codes))
	linesReady := true
	for _, code := range codes {
		_, initialized := hasAgent[code]
		if !initialized {
			linesReady = false
		}
		lines = append(lines, types.LineInitStatus{LineCode: code, Initialized: initialized})
	}
	return &types.AgentInitStatusResp{
		Ready:       levelsReady && linesReady,
		LevelsReady: levelsReady,
		LevelCount:  levelCount,
		Lines:       lines,
	}, nil
}
