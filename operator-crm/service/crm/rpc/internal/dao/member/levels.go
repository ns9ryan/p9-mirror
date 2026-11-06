package member

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	m "oa.98ent.com/p9/operator-crm/ent/member"
	a "oa.98ent.com/p9/operator-crm/ent/memberautolevel"
	f "oa.98ent.com/p9/operator-crm/ent/memberfixedlevel"
	assignment "oa.98ent.com/p9/operator-crm/ent/memberfixedlevelassignment"
)

// updateLevels locks the member so simultaneous replacements cannot merge assignments.
func (s *Service) updateLevels(ctx context.Context, id int64, update func(*ent.Tx, string) error) error {
	claims, ok := ctxdata.OperatorClaimsFromCtx(ctx)
	if !ok {
		return xerr.Unauthorized("common.unauthorized")
	}
	code := claims.OperatorCode
	if id <= 0 {
		return xerr.BadRequest("common.invalidParam")
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Member.Query().Where(m.IDEQ(id), m.OperatorCodeEQ(code)).Where(func(sel *sql.Selector) { sel.ForUpdate() }).Only(ctx)
	if ent.IsNotFound(err) {
		return xerr.BadRequest("common.invalidParam")
	}
	if err != nil {
		return err
	}
	if err = update(tx, code); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) UpdateAutoLevel(ctx context.Context, id, levelID int64) error {
	if levelID <= 0 {
		return xerr.BadRequest("common.invalidParam")
	}
	return s.updateLevels(ctx, id, func(tx *ent.Tx, code string) error {
		_, err := tx.MemberAutoLevel.Query().Where(a.IDEQ(levelID), a.OperatorCodeEQ(code)).Only(ctx)
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
	if len(ids) > 100 {
		return xerr.BadRequest("common.invalidParam")
	}
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
	return s.updateLevels(ctx, id, func(tx *ent.Tx, code string) error {
		if len(unique) > 0 {
			count, err := tx.MemberFixedLevel.Query().Where(f.IDIn(unique...), f.OperatorCodeEQ(code)).Count(ctx)
			if err != nil {
				return err
			}
			if count != len(unique) {
				return xerr.BadRequest("common.invalidParam")
			}
		}
		if _, err := tx.MemberFixedLevelAssignment.Delete().Where(assignment.MemberIDEQ(id)).Exec(ctx); err != nil {
			return err
		}
		for _, levelID := range unique {
			if _, err := tx.MemberFixedLevelAssignment.Create().SetMemberID(id).SetFixedLevelID(levelID).Save(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
