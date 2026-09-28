package operatornodeservicelogic

import (
	nodedispatchoperatornodepb "oa.98ent.com/p9/node-dispatch/rpc/pb/nodedispatchrpc/operatornodepb"
	"oa.98ent.com/p9/platform-operator/rpc/pb/platformoperatorrpc/operatornodepb"
)

// toOperatorNodeInfo 转换分站部署节点信息
func toOperatorNodeInfo(operatorID int64, data *nodedispatchoperatornodepb.OperatorNodeInfo) *operatornodepb.OperatorNodeInfo {
	if data == nil || data.Node == nil {
		return nil
	}

	nodeData := data.Node

	return &operatornodepb.OperatorNodeInfo{
		OperatorId:     operatorID,          // 分站ID
		OperatorCode:   data.OperatorCode,   // 分站全局唯一业务编码
		NodeCode:       nodeData.Code,       // 节点业务编码
		NodeName:       nodeData.Name,       // 节点名称
		NodeStatus:     nodeData.Status,     // 节点状态: 1启用, 2停用
		NodeOnline:     nodeData.Online,     // 节点是否在线
		NodeLastSeenAt: nodeData.LastSeenAt, // 节点最近一次活动时间, Unix毫秒时间戳
		NodeRemark:     nodeData.Remark,     // 节点运维备注
		CreatedAt:      data.CreatedAt,      // 部署节点关系创建时间, Unix毫秒时间戳
		UpdatedAt:      data.UpdatedAt,      // 部署节点关系更新时间, Unix毫秒时间戳
	}
}
