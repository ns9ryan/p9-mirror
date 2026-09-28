package operatornodeservicelogic

import (
	"oa.98ent.com/p9/node-dispatch/rpc/ent"
	"oa.98ent.com/p9/node-dispatch/rpc/pb/nodedispatchrpc/nodepb"
	"oa.98ent.com/p9/node-dispatch/rpc/pb/nodedispatchrpc/operatornodepb"
)

// toOperatorNodeInfo 转换分站部署节点信息
func toOperatorNodeInfo(data *ent.OperatorNode, nodeData *ent.Node, online bool) *operatornodepb.OperatorNodeInfo {
	nodeInfo := &nodepb.NodeInfo{
		Id:        nodeData.ID,                    // 节点ID
		Code:      nodeData.Code,                  // 节点业务编码
		Name:      nodeData.Name,                  // 节点名称
		Status:    nodeData.Status,                // 节点状态: 1启用, 2停用
		Online:    online,                         // 是否在线
		Remark:    nodeData.Remark,                // 运维备注
		CreatedAt: nodeData.CreatedAt.UnixMilli(), // 创建时间, Unix毫秒时间戳
		UpdatedAt: nodeData.UpdatedAt.UnixMilli(), // 更新时间, Unix毫秒时间戳
	}

	// 设置最近一次活动时间
	if nodeData.LastSeenAt != nil {
		nodeInfo.LastSeenAt = new(nodeData.LastSeenAt.UnixMilli())
	}

	return &operatornodepb.OperatorNodeInfo{
		OperatorCode: data.OperatorCode,          // 分站全局唯一业务编码
		Node:         nodeInfo,                   // 部署节点信息
		CreatedAt:    data.CreatedAt.UnixMilli(), // 创建时间, Unix毫秒时间戳
		UpdatedAt:    data.UpdatedAt.UnixMilli(), // 更新时间, Unix毫秒时间戳
	}
}
