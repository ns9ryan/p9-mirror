package manage

import (
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"testing"
)

func TestExportRPCRequestIncludesDemoSelection(t *testing.T) {
	for _, includeDemo := range []bool{false, true} {
		request := exportRPCRequest(&types.ExportMembersRequest{IncludeDemo: includeDemo, AgentType: 1})
		if request.IncludeDemo != includeDemo || request.AgentType != 1 {
			t.Fatal("export RPC request lost demo or agent type selection")
		}
	}
}
