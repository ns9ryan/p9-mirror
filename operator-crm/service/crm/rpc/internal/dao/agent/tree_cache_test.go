package agent

import (
	"testing"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
)

func TestBuildAgentTreeAncestorsChain(t *testing.T) {
	// A(顶) -> B -> C -> D
	tree := buildAgentTree([]agentTreeRow{
		{ID: 1, ParentID: 0, Account: "A", ShareRate: 100},
		{ID: 2, ParentID: 1, Account: "B", ShareRate: 95},
		{ID: 3, ParentID: 2, Account: "C", ShareRate: 80},
		{ID: 4, ParentID: 3, Account: "D", ShareRate: 75},
	})
	chain, err := tree.ancestorsByID(4)
	if err != nil {
		t.Fatalf("ancestorsByID: %v", err)
	}
	want := []struct {
		id, parentID int64
		account      string
		share, layer int32
	}{
		{4, 3, "D", 75, 75},
		{3, 2, "C", 80, 5},
		{2, 1, "B", 95, 15},
		{1, 0, "A", 100, 5},
	}
	if len(chain) != len(want) {
		t.Fatalf("len=%d want %d", len(chain), len(want))
	}
	for i, w := range want {
		n := chain[i]
		if n.ID != w.id || n.ParentID != w.parentID || n.Account != w.account || n.ShareRate != w.share || n.LayerRate != w.layer {
			t.Fatalf("[%d] got %+v want id=%d parent=%d account=%s share=%d layer=%d",
				i, n, w.id, w.parentID, w.account, w.share, w.layer)
		}
	}
}

func TestAncestorsByIDNotFound(t *testing.T) {
	tree := buildAgentTree([]agentTreeRow{
		{ID: 1, ParentID: 0, Account: "A", ShareRate: 100},
	})
	_, err := tree.ancestorsByID(99)
	if err == nil {
		t.Fatal("expected not found")
	}
	assertApplyKey(t, err, i18n.AgentNotFound)
}

func TestTreeCacheInvalidate(t *testing.T) {
	c := newAgentTreeCache()
	c.set("op01", buildAgentTree([]agentTreeRow{
		{ID: 1, ParentID: 0, Account: "A", ShareRate: 100},
	}))
	if _, ok := c.get("op01"); !ok {
		t.Fatal("expected cache hit")
	}
	c.invalidate("op01")
	if _, ok := c.get("op01"); ok {
		t.Fatal("expected cache miss after invalidate")
	}
}

func TestAncestorsByIDCycleBreak(t *testing.T) {
	// 人为环：1 <-> 2
	tree := buildAgentTree([]agentTreeRow{
		{ID: 1, ParentID: 2, Account: "A", ShareRate: 100},
		{ID: 2, ParentID: 1, Account: "B", ShareRate: 90},
	})
	chain, err := tree.ancestorsByID(1)
	if err != nil {
		t.Fatalf("ancestorsByID: %v", err)
	}
	if len(chain) != 2 {
		t.Fatalf("cycle should stop after 2 nodes, got %d", len(chain))
	}
}
