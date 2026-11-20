package operator

import (
	"context"
	"fmt"
	"strings"

	"oa.98ent.com/p9/operator-base/rpc/ent"
	"oa.98ent.com/p9/operator-base/rpc/ent/operatorlanguage"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// normalizeLanguageCodes 整理 operator 初始化语言编码
func normalizeLanguageCodes(languageCodes []string) {
	for i := range languageCodes {
		languageCodes[i] = strings.TrimSpace(languageCodes[i])
	}
}

// validateLanguageCodes 校验 operator 初始化语言编码
func validateLanguageCodes(languageCodes []string) error {
	// 初始化时至少需要一种语言
	if len(languageCodes) == 0 {
		return status.Error(codes.InvalidArgument, "language_codes is required")
	}

	seen := make(map[string]struct{}, len(languageCodes))

	for _, languageCode := range languageCodes {
		// 校验语言编码
		if languageCode == "" {
			return status.Error(codes.InvalidArgument, "language_codes contains empty value")
		}

		// 同一个 operator 不能分配重复语言
		if _, exists := seen[languageCode]; exists {
			return status.Error(codes.InvalidArgument, "language_codes contains duplicate value")
		}
		seen[languageCode] = struct{}{}
	}

	return nil
}

// createLanguages 创建 operator 当前有效语言
func createLanguages(ctx context.Context, tx *ent.Tx, operatorID int64, languageCodes []string) error {
	builders := make([]*ent.OperatorLanguageCreate, 0, len(languageCodes))

	for _, languageCode := range languageCodes {
		builders = append(builders,
			tx.OperatorLanguage.
				Create().
				SetOperatorID(operatorID).     // operator 本地主键
				SetLanguageCode(languageCode), // 语言编码
		)
	}

	if err := tx.OperatorLanguage.CreateBulk(builders...).Exec(ctx); err != nil {
		return fmt.Errorf("创建 operator 语言失败: %w", err)
	}

	return nil
}

// SyncLanguages 在调用方事务中同步本站获分配的语言，不是保存前端的启用勾选列表。
// 保留已有语言的启停状态，仅删除撤回项并新增分配项；事务提交或回滚由调用方负责。
func SyncLanguages(ctx context.Context, tx *ent.Tx, operatorID int64, languageCodes []string) error {
	// 读取该站当前分配记录，用于区分保留项与新增项，不调用其他服务 RPC。
	current, err := tx.OperatorLanguage.Query().Where(operatorlanguage.OperatorIDEQ(operatorID)).All(ctx)
	if err != nil {
		return fmt.Errorf("查询 operator 语言失败: %w", err)
	}
	existing := make(map[string]bool, len(current))
	for _, row := range current {
		existing[row.LanguageCode] = true
	}
	// 仅撤回不再分配的语言，不删除重建仍在分配范围内的记录。
	if _, err := tx.OperatorLanguage.Delete().Where(
		operatorlanguage.OperatorIDEQ(operatorID), operatorlanguage.LanguageCodeNotIn(languageCodes...),
	).Exec(ctx); err != nil {
		return fmt.Errorf("删除已撤回语言失败: %w", err)
	}
	// 已有记录保持原状态；新分配的语言才交给现有创建函数，默认启用。
	var added []string
	for _, code := range languageCodes {
		if !existing[code] {
			added = append(added, code)
		}
	}
	if len(added) == 0 {
		return nil
	}
	// 复用本站语言批量创建逻辑，仍使用同一事务客户端写库。
	return createLanguages(ctx, tx, operatorID, added)
}
