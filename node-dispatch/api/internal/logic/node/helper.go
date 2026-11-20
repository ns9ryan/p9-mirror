package node

import (
	"oa.98ent.com/p9/node-dispatch/api/internal/types"
	"oa.98ent.com/p9/node-dispatch/rpc/pb/nodedispatchrpc/nodepb"
)

// toNodeInfo 转换节点信息
func toNodeInfo(data *nodepb.NodeInfo) types.NodeInfo {
	return types.NodeInfo{
		Id:         data.Id,         // 节点ID
		Code:       data.Code,       // 节点业务编码
		Name:       data.Name,       // 节点名称
		Status:     data.Status,     // 节点状态: 1启用, 2停用
		Online:     data.Online,     // 节点是否在线
		LastSeenAt: data.LastSeenAt, // 最近一次活动时间, Unix毫秒时间戳
		Remark:     data.Remark,     // 运维备注
		CreatedAt:  data.CreatedAt,  // 创建时间, Unix毫秒时间戳
		UpdatedAt:  data.UpdatedAt,  // 更新时间, Unix毫秒时间戳
	}
}
