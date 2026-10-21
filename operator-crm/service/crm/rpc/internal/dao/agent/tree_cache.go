package agent

import (
	"context"
	"strings"
	"sync"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent/agentuser"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"
)

// agentTreeNode 内存树节点（按 ID 上溯用）。
type agentTreeNode struct {
	ID            int64
	ParentID      int64  // 顶层为 0
	Account       string // 账号
	Level         int32  // 层级 1~7
	AgentLineCode string // 代理线代码
	ShareRate     int32  // 绝对占成 0~100
	LayerRate     int32  // 本层占成：本人=绝对占成；上级=自身绝对占成−链上下一级绝对占成
}

// agentTree 分站代理扁平索引。
type agentTree struct {
	nodes map[int64]*agentTreeNode
}

type agentTreeCache struct {
	mu    sync.RWMutex
	trees map[string]*agentTree // operator_code -> tree
}

func newAgentTreeCache() *agentTreeCache {
	return &agentTreeCache{trees: make(map[string]*agentTree)}
}

type agentTreeRow struct {
	ID            int64
	ParentID      int64
	Account       string
	Level         int32
	AgentLineCode string
	ShareRate     int32
}

func buildAgentTree(rows []agentTreeRow) *agentTree {
	nodes := make(map[int64]*agentTreeNode, len(rows))
	for _, row := range rows {
		nodes[row.ID] = &agentTreeNode{
			ID:            row.ID,
			ParentID:      row.ParentID,
			Account:       row.Account,
			ShareRate:     row.ShareRate,
			Level:         row.Level,
			AgentLineCode: row.AgentLineCode,
		}
	}
	return &agentTree{nodes: nodes}
}

func (c *agentTreeCache) get(op string) (*agentTree, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	t, ok := c.trees[op]
	return t, ok
}

func (c *agentTreeCache) set(op string, t *agentTree) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.trees[op] = t
}

func (c *agentTreeCache) invalidate(op string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.trees, op)
}

func (t *agentTree) ancestorsByID(agentID int64) ([]*agentTreeNode, error) {
	if t == nil || t.nodes == nil {
		return nil, xerr.NotFound(i18n.AgentNotFound)
	}
	start, ok := t.nodes[agentID]
	if !ok {
		return nil, xerr.NotFound(i18n.AgentNotFound)
	}
	limit := int(maxLevel) + 1
	chain := make([]*agentTreeNode, 0, limit)
	seen := make(map[int64]struct{}, limit)
	cur := start
	for cur != nil && len(chain) < limit {
		if _, ok := seen[cur.ID]; ok {
			break
		}
		seen[cur.ID] = struct{}{}
		chain = append(chain, cur)
		if cur.ParentID <= 0 {
			break
		}
		cur = t.nodes[cur.ParentID]
	}
	// 拷贝并填充本层占成，避免改写缓存节点
	out := make([]*agentTreeNode, len(chain))
	for i, n := range chain {
		layerRate := n.ShareRate
		if i > 0 {
			layerRate = n.ShareRate - chain[i-1].ShareRate
		}
		out[i] = &agentTreeNode{
			ID:            n.ID,
			ParentID:      n.ParentID,
			Account:       n.Account,
			ShareRate:     n.ShareRate,
			LayerRate:     layerRate,
			Level:         n.Level,
			AgentLineCode: n.AgentLineCode,
		}
	}
	return out, nil
}

// LoadAgentTree 按当前分站加载（或命中）内存代理树。
func (s *Service) LoadAgentTree(ctx context.Context) (*agentTree, error) {
	ctx, op, err := requireOperator(ctx)
	if err != nil {
		return nil, err
	}
	if t, ok := s.treeCache.get(op); ok {
		return t, nil
	}
	rows, err := s.DB.AgentUser.Query().
		Where(agentuser.OperatorCodeEQ(op)).
		Select(
			agentuser.FieldID, agentuser.FieldParentID,
			agentuser.FieldAccount, agentuser.FieldLevel,
			agentuser.FieldAgentLineCode, agentuser.FieldShareRate,
		).
		All(ctx)
	if err != nil {
		return nil, err
	}
	flat := make([]agentTreeRow, 0, len(rows))
	for _, row := range rows {
		parentID := int64(0)
		if row.ParentID != nil {
			parentID = *row.ParentID
		}
		flat = append(flat, agentTreeRow{
			ID:            row.ID,
			ParentID:      parentID,
			Account:       row.Account,
			Level:         int32(row.Level),
			AgentLineCode: row.AgentLineCode,
			ShareRate:     int32(row.ShareRate),
		})
	}
	t := buildAgentTree(flat)
	s.treeCache.set(op, t)
	return t, nil
}

// AncestorsByID 返回 [自己, 父, ..., 顶] 的 id/parent_id/account/share_rate/layer_rate。
func (s *Service) AncestorsByID(ctx context.Context, agentID int64) ([]*agentTreeNode, error) {
	if agentID <= 0 {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	t, err := s.LoadAgentTree(ctx)
	if err != nil {
		return nil, err
	}
	return t.ancestorsByID(agentID)
}

// AncestorsByIDRPC RPC 适配：按 agent_id 返回祖先链。
func (s *Service) AncestorsByIDRPC(ctx context.Context, in *agentpb.AncestorsByIDReq) (*agentpb.AncestorsByIDResp, error) {
	if in == nil {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	chain, err := s.AncestorsByID(ctx, in.GetAgentId())
	if err != nil {
		return nil, err
	}
	items := make([]*agentpb.AgentAncestorItem, 0, len(chain))
	for _, n := range chain {
		if n == nil {
			continue
		}
		items = append(items, &agentpb.AgentAncestorItem{
			Id:            n.ID,
			ParentId:      n.ParentID,
			Account:       n.Account,
			Level:         n.Level,
			AgentLineCode: n.AgentLineCode,
			ShareRate:     n.ShareRate,
			LayerRate:     n.LayerRate,
		})
	}
	return &agentpb.AncestorsByIDResp{Items: items}, nil
}

func (s *Service) invalidateTree(op string) {
	if s == nil || s.treeCache == nil {
		return
	}
	op = strings.TrimSpace(op)
	if op == "" {
		return
	}
	s.treeCache.invalidate(op)
}
