package initlogic

import (
	"context"

	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/memberautolevel"
)

// initMemberAutoLevel 初始化默认自动层级, 已存在时不覆盖
func initMemberAutoLevel(ctx context.Context, tx *ent.Tx) error {
	return tx.MemberAutoLevel.Create().
		SetName("普通会员").  // 自动层级名称
		SetMinDeposit(0). // 晋级要求的累计存款金额
		SetMinWater(0).   // 晋级要求的累计洗码量
		OnConflictColumns(
			memberautolevel.FieldOperatorCode,
			memberautolevel.FieldMinDeposit,
			memberautolevel.FieldMinWater,
		).
		DoNothing().
		Exec(ctx)
}
