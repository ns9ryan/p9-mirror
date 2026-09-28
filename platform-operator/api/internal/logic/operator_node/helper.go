package operator_node

import (
	"oa.98ent.com/p9/node-dispatch/rpc/pb/nodedispatchrpc/nodepb"
	"oa.98ent.com/p9/platform-operator/api/internal/types"
)

// toOperatorNodeInfo 转换分站部署节点信息
func toOperatorNodeInfo(data *nodepb.NodeInfo, selectedNodeCode string) types.OperatorNodeInfo {
	return types.OperatorNodeInfo{
		NodeCode:       data.Code,                     // 节点业务编码
		NodeName:       data.Name,                     // 节点名称
		NodeStatus:     data.Status,                   // 节点状态: 1启用, 2停用
		NodeOnline:     data.Online,                   // 节点是否在线
		NodeLastSeenAt: data.LastSeenAt,               // 节点最近一次活动时间, Unix毫秒时间戳
		NodeRemark:     data.Remark,                   // 节点运维备注
		Selected:       data.Code == selectedNodeCode, // 是否为当前分站已选择的部署节点
	}
}
