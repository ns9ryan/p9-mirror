package main

import (
	"fmt"

	"oa.98ent.com/p9/common/utils/orderno"
)

func main() {
	// 1. 获取节点ID
	nodeID, ipAddr, err := orderno.GetNodeIDFromLocalIP()
	if err != nil {
		fmt.Println("获取节点ID失败:", err)
		return
	}
	fmt.Println("节点ID:", nodeID, "IP地址:", ipAddr)

	// 2. 生成新的同规则单号
	sf := orderno.NewSnowflake(0)
	fmt.Println("新生成订单号1:", sf.GenerateWithPrefix("PAO"))

	sd := orderno.NewSnowflake(nodeID)
	fmt.Println("新生成订单号2:", sd.GenerateWithPrefix("PAO"))
}
