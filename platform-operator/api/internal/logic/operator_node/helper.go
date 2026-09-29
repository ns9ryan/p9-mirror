package operator_node

import (
	"oa.98ent.com/p9/platform-operator/api/internal/types"
	"oa.98ent.com/p9/platform-operator/rpc/pb/platformoperatorrpc/operatornodepb"
)

// toOperatorNodeInfo 转换分站部署节点信息
func toOperatorNodeInfo(data *operatornodepb.OperatorNodeInfo) *types.OperatorNodeInfo {
	if data == nil {
		return nil
	}

	return &types.OperatorNodeInfo{
		NodeCode:       data.NodeCode,       // 节点业务编码
		NodeName:       data.NodeName,       // 节点名称
		NodeStatus:     data.NodeStatus,     // 节点状态: 1启用, 2停用
		NodeOnline:     data.NodeOnline,     // 节点是否在线
		NodeLastSeenAt: data.NodeLastSeenAt, // 节点最近一次活动时间, Unix毫秒时间戳
		NodeRemark:     data.NodeRemark,     // 节点运维备注
	}
}
