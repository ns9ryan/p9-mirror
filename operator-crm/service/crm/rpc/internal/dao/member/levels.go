package member

import (
	"context"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/member"
	"oa.98ent.com/p9/operator-crm/ent/memberautolevel"
	"oa.98ent.com/p9/operator-crm/ent/memberfixedlevel"
	"oa.98ent.com/p9/operator-crm/ent/memberfixedlevelassignment"
)

const maxFixedLevels = 100

// withMemberUpdateTx 校验操作员和会员，并在事务内执行会员更新。
func (s *Service) withMemberUpdateTx(ctx context.Context, id int64, update func(*ent.Tx) error) error {
	_, ok := ctxdata.OperatorClaimsFromCtx(ctx)
	if !ok {
		return xerr.Unauthorized("common.unauthorized")
	}
	if id <= 0 {
		return xerr.BadRequest("common.invalidParam")
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return err
	}
	// 会员存在性检查与更新回调共用事务，回调成功后才提交。
	defer tx.Rollback()
	_, err = tx.Member.Query().Where(member.IDEQ(id)).Only(ctx)
	if ent.IsNotFound(err) {
		return xerr.BadRequest("common.invalidParam")
	}
	if err != nil {
		return err
	}
	if err = update(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) UpdateAutoLevel(ctx context.Context, id, levelID int64) error {
	if levelID <= 0 {
		return xerr.BadRequest("common.invalidParam")
	}
	return s.withMemberUpdateTx(ctx, id, func(tx *ent.Tx) error {
		// 确认目标自动层级存在后再更新会员的层级关联。
		_, err := tx.MemberAutoLevel.Query().Where(memberautolevel.IDEQ(levelID)).Only(ctx)
		if ent.IsNotFound(err) {
			return xerr.BadRequest("common.invalidParam")
		}
		if err != nil {
			return err
		}
		return tx.Member.UpdateOneID(id).SetAutoLevelID(levelID).Exec(ctx)
	})
}

func (s *Service) UpdateFixedLevels(ctx context.Context, id int64, ids []int64) error {
	if len(ids) > maxFixedLevels {
		return xerr.BadRequest("common.invalidParam")
	}
	// 先校验并去重，后续按唯一层级 ID 校验目标记录是否齐全。
	unique := make([]int64, 0, len(ids))
	seen := make(map[int64]bool, len(ids))
	for _, levelID := range ids {
		if levelID <= 0 {
			return xerr.BadRequest("common.invalidParam")
		}
		if !seen[levelID] {
			seen[levelID] = true
			unique = append(unique, levelID)
		}
	}
	return s.withMemberUpdateTx(ctx, id, func(tx *ent.Tx) error {
		if len(unique) > 0 {
			count, err := tx.MemberFixedLevel.Query().Where(memberfixedlevel.IDIn(unique...)).Count(ctx)
			if err != nil {
				return err
			}
			if count != len(unique) {
				return xerr.BadRequest("common.invalidParam")
			}
		}
		// 在同一事务中替换全部固定层级分配，空列表表示清空。
		if _, err := tx.MemberFixedLevelAssignment.Delete().Where(memberfixedlevelassignment.MemberIDEQ(id)).Exec(ctx); err != nil {
			return err
		}
		if len(unique) == 0 {
			return nil
		}
		builders := make([]*ent.MemberFixedLevelAssignmentCreate, 0, len(unique))
		for _, levelID := range unique {
			builders = append(builders, tx.MemberFixedLevelAssignment.Create().SetMemberID(id).SetFixedLevelID(levelID))
		}
		return tx.MemberFixedLevelAssignment.CreateBulk(builders...).Exec(ctx)
	})
}
