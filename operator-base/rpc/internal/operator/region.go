package operator

import (
	"context"
	"fmt"
	"strings"

	"oa.98ent.com/p9/operator-base/rpc/ent"
	"oa.98ent.com/p9/operator-base/rpc/ent/operatorregion"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// normalizeRegionCodes 整理 operator 初始化经营地区编码
func normalizeRegionCodes(regionCodes []string) {
	for i := range regionCodes {
		regionCodes[i] = strings.ToUpper(strings.TrimSpace(regionCodes[i]))
	}
}

// validateRegionCodes 校验 operator 初始化经营地区编码
func validateRegionCodes(regionCodes []string) error {
	// 初始化时至少需要一个经营地区
	if len(regionCodes) == 0 {
		return status.Error(codes.InvalidArgument, "region_codes is required")
	}

	seen := make(map[string]struct{}, len(regionCodes))

	for _, regionCode := range regionCodes {
		// 校验经营地区编码
		if regionCode == "" {
			return status.Error(codes.InvalidArgument, "region_codes contains empty value")
		}

		// 同一个 operator 不能分配重复经营地区
		if _, exists := seen[regionCode]; exists {
			return status.Error(codes.InvalidArgument, "region_codes contains duplicate value")
		}
		seen[regionCode] = struct{}{}
	}

	return nil
}

// createRegions 创建 operator 当前有效经营地区
func createRegions(ctx context.Context, tx *ent.Tx, operatorID int64, regionCodes []string) error {
	builders := make([]*ent.OperatorRegionCreate, 0, len(regionCodes))

	for _, regionCode := range regionCodes {
		builders = append(builders,
			tx.OperatorRegion.
				Create().
				SetOperatorID(operatorID). // operator 本地主键
				SetRegionCode(regionCode), // 经营地区编码
		)
	}

	if err := tx.OperatorRegion.CreateBulk(builders...).Exec(ctx); err != nil {
		return fmt.Errorf("创建 operator 经营地区失败: %w", err)
	}

	return nil
}

// SyncRegions 在调用方事务中同步本站获分配的地区，不是保存前端的启用勾选列表。
// 保留已有地区的启停状态，仅删除撤回项并新增分配项；事务提交或回滚由调用方负责。
func SyncRegions(ctx context.Context, tx *ent.Tx, operatorID int64, regionCodes []string) error {
	// 读取该站当前分配记录，用于区分保留项与新增项，不调用其他服务 RPC。
	current, err := tx.OperatorRegion.Query().Where(operatorregion.OperatorIDEQ(operatorID)).All(ctx)
	if err != nil {
		return fmt.Errorf("查询 operator 经营地区失败: %w", err)
	}
	existing := make(map[string]bool, len(current))
	for _, row := range current {
		existing[row.RegionCode] = true
	}
	// 仅撤回不再分配的地区，不删除重建仍在分配范围内的记录。
	if _, err := tx.OperatorRegion.Delete().Where(
		operatorregion.OperatorIDEQ(operatorID), operatorregion.RegionCodeNotIn(regionCodes...),
	).Exec(ctx); err != nil {
		return fmt.Errorf("删除已撤回经营地区失败: %w", err)
	}
	// 已有记录保持原状态；新分配的地区才交给现有创建函数，默认启用。
	var added []string
	for _, code := range regionCodes {
		if !existing[code] {
			added = append(added, code)
		}
	}
	if len(added) == 0 {
		return nil
	}
	// 复用本站地区批量创建逻辑，仍使用同一事务客户端写库。
	return createRegions(ctx, tx, operatorID, added)
}
