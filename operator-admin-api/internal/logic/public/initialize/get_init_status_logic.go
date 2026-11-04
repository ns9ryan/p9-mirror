// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package initialize

import (
	"context"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	agentlogic "oa.98ent.com/p9/operator-admin-api/internal/logic/agent"

	"github.com/zeromicro/go-zero/core/logx"
)

// 最低初始化等级数量
const minInitLevelCount = 3

type GetInitStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 检测分站代理线与等级是否完成初始化
func NewGetInitStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInitStatusLogic {
	return &GetInitStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetInitStatusLogic) GetInitStatus() (resp *types.InitStatusResp, err error) {
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
	levelCount := int32(len(levels.GetList()))
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
	return &types.InitStatusResp{
		Ready:       levelsReady && linesReady,
		LevelsReady: levelsReady,
		LevelCount:  levelCount,
		Lines:       lines,
	}, nil
}
