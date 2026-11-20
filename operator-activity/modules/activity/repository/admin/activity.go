package admin

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"

	"oa.98ent.com/p9/common/paging"
	"oa.98ent.com/p9/operator-activity/ent"
	"oa.98ent.com/p9/operator-activity/ent/activity"
	"oa.98ent.com/p9/operator-activity/ent/activityi18n"
	"oa.98ent.com/p9/operator-activity/ent/predicate"
	"oa.98ent.com/p9/operator-activity/modules/activity/repository"
	activitycache "oa.98ent.com/p9/operator-activity/modules/activity/repository/cache"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
	"oa.98ent.com/p9/operator-activity/pkg/database"
)

// ErrActivityNotDraft 非草稿状态不可编辑/发布
var ErrActivityNotDraft = errors.New("activity not draft")

// ActivityI18nInput 活动国际化写入参数
type ActivityI18nInput struct {
	Lang        string
	Name        string
	Rule        string
	Description string
	Image       string
}

// ListActivityParams 活动列表查询参数
type ListActivityParams struct {
	Page     int64
	PageSize int64
	Name     string
	Type     *int32
	Client   *int64
	Status   *int32
	Lang     string
}

// ActivityListRow 活动列表行
type ActivityListRow struct {
	ID                 int64
	Type               types.ActivityType
	Status             types.ActivityStatus
	Name               string
	Image              string
	AllowedClients     []int64
	OpenAt             *time.Time
	CloseAt            *time.Time
	ReleaseAt          *time.Time
	SortWeight         int
	ShowInList         bool
	UpdatedAt          time.Time
	AutoAudit          bool
	AutoAuditMaxAmt    int64
	AllowedFixedLevel  []int64
	AllowedAutoLevel   []int64
	WageringMultiplier int
	AllowedGameIds     []int64
	IsPermanent        bool
	CreatedAt          time.Time
	CreateUid          int64
	UpdateUid          int64
	Section            types.ActivitySection
}

// CreateActivity 事务创建活动及国际化配置
func (s *AdminService) CreateActivity(ctx context.Context, m *repository.ActivityModel, i18nList []ActivityI18nInput) (int64, error) {
	var id int64
	err := database.WithTx(ctx, s.db, func(tx *ent.Tx) error {
		c := tx.Activity.Create().
			SetType(int(m.Type)).
			SetStatus(int(m.Status)).
			SetApplyMode(int(m.ApplyMode)).
			SetCycleMode(int(m.CycleMode)).
			SetIsPermanent(m.IsPermanent).
			SetAllowedFixedLevel(nilToEmpty(m.AllowedFixedLevel)).
			SetAllowedAutoLevel(nilToEmpty(m.AllowedAutoLevel)).
			SetAllowedClients(nilToEmpty(m.AllowedClients)).
			SetAllowedGameIds(nilToEmpty(m.AllowedGameIds)).
			SetAllowedAgentIds(nilToEmpty(m.AllowedAgentIds)).
			SetApplyLimit(nilToEmpty(m.ApplyLimit)).
			SetClaimLimit(nilToEmpty(m.ClaimLimit)).
			SetNillableOpenAt(m.OpenAt).
			SetNillableCloseAt(m.CloseAt).
			SetNillableReleaseAt(m.ReleaseAt).
			SetAutoAudit(m.AutoAudit).
			SetAutoAuditMaxAmt(m.AutoAuditMaxAmt).
			SetWageringMultiplier(m.WageringMultiplier).
			SetWageringBase(int(m.WageringBase)).
			SetPayoutMode(int(m.PayoutMode)).
			SetPayoutTiming(int(m.PayoutTiming)).
			SetSection(int(m.Section)).
			SetClaimExpireSec(m.ClaimExpireSec).
			SetSettings(defaultJSONObject(m.Settings)).
			SetShowInList(m.ShowInList).
			SetShowInSidebar(m.ShowInSidebar).
			SetCreateUID(m.CreateUid).
			SetUpdateUID(m.CreateUid).
			SetSortWeight(m.SortWeight)

		row, err := c.Save(ctx)
		if err != nil {
			return fmt.Errorf("create activity: %w", err)
		}
		id = row.ID

		for _, item := range i18nList {
			if item.Lang == "" {
				continue
			}
			if _, err := tx.ActivityI18n.Create().
				SetActivityID(id).
				SetLang(item.Lang).
				SetName(item.Name).
				SetRule(item.Rule).
				SetDescription(item.Description).
				SetImage(item.Image).
				Save(ctx); err != nil {
				return fmt.Errorf("create activity i18n: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	if m.Status == types.ActivityStatusOngoing {
		activitycache.DelOngoingActivitiesCache(m.OperatorCode, m.Type)
	}
	return id, nil
}

// GetActivityDetail 管理端获取活动详情（含 i18n）
// 注：现有 GetActivity 返回运行时 types.Activity，签名冲突故使用本命名。
func (s *AdminService) GetActivityDetail(ctx context.Context, id int64) (*repository.ActivityModel, []ActivityI18nInput, error) {
	row, err := s.db.Activity.Query().
		Where(activity.IDEQ(id)).
		WithI18ns().
		Only(ctx)
	if err != nil {
		return nil, nil, err
	}

	i18nList := make([]ActivityI18nInput, 0, len(row.Edges.I18ns))
	for _, item := range row.Edges.I18ns {
		if item == nil {
			continue
		}
		i18nList = append(i18nList, ActivityI18nInput{
			Lang:        item.Lang,
			Name:        item.Name,
			Rule:        item.Rule,
			Description: item.Description,
			Image:       item.Image,
		})
	}
	return repository.ActivityFromEnt(row), i18nList, nil
}

// ListActivities 管理端分页查询活动列表
func (s *AdminService) ListActivities(ctx context.Context, params ListActivityParams) ([]ActivityListRow, int64, error) {
	offset, limit, ok := paging.OffsetLimit(params.Page, params.PageSize)
	if !ok {
		return nil, 0, fmt.Errorf("invalid page")
	}
	// 国际化语言，为空则使用系统语言
	lang := strings.TrimSpace(params.Lang)

	q := s.db.Activity.Query()

	name := strings.TrimSpace(params.Name)
	if name != "" {
		q = q.Where(activity.HasI18nsWith(
			activityi18n.LangEQ(lang),
			activityi18n.NameContainsFold(name),
		))
	}
	if params.Type != nil {
		q = q.Where(activity.TypeEQ(int(*params.Type)))
	}
	if params.Status != nil && *params.Status > 0 {
		q = q.Where(activity.StatusEQ(int(*params.Status)))
	}
	if params.Client != nil {
		q = q.Where(allowedClientsContainsOrEmpty(*params.Client))
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []ActivityListRow{}, 0, nil
	}

	rows, err := q.
		WithI18ns(func(iq *ent.ActivityI18nQuery) {
			iq.Where(activityi18n.LangEQ(lang))
		}).
		Order(activity.BySortWeight(sql.OrderDesc()), activity.ByID(sql.OrderDesc())).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	list := make([]ActivityListRow, 0, len(rows))
	for _, row := range rows {
		displayName := ""
		displayImage := ""
		if len(row.Edges.I18ns) > 0 && row.Edges.I18ns[0] != nil {
			displayName = row.Edges.I18ns[0].Name
			displayImage = row.Edges.I18ns[0].Image
		}
		list = append(list, ActivityListRow{
			ID:                 row.ID,
			Type:               types.ActivityType(row.Type),
			Status:             types.ActivityStatus(row.Status),
			Name:               displayName,
			Image:              displayImage,
			AllowedClients:     row.AllowedClients,
			OpenAt:             row.OpenAt,
			CloseAt:            row.CloseAt,
			ReleaseAt:          row.ReleaseAt,
			SortWeight:         row.SortWeight,
			ShowInList:         row.ShowInList,
			UpdatedAt:          row.UpdatedAt,
			AutoAudit:          row.AutoAudit,
			AutoAuditMaxAmt:    row.AutoAuditMaxAmt,
			AllowedFixedLevel:  row.AllowedFixedLevel,
			AllowedAutoLevel:   row.AllowedAutoLevel,
			WageringMultiplier: row.WageringMultiplier,
			AllowedGameIds:     row.AllowedGameIds,
			IsPermanent:        row.IsPermanent,
			CreatedAt:          row.CreatedAt,
			CreateUid:          row.CreateUID,
			UpdateUid:          row.UpdateUID,
			Section:            types.ActivitySection(row.Section),
		})
	}
	return list, int64(total), nil
}

// UpdateActivity 编辑活动：仅草稿可改；主表更新并整批替换 i18n
func (s *AdminService) UpdateActivity(ctx context.Context, m *repository.ActivityModel, i18nList []ActivityI18nInput) error {
	if m == nil || m.ID <= 0 {
		return fmt.Errorf("invalid activity id")
	}

	row, err := s.db.Activity.Get(ctx, m.ID)
	if err != nil {
		return err
	}
	if row.Status != int(types.ActivityStatusDraft) {
		return ErrActivityNotDraft
	}

	return database.WithTx(ctx, s.db, func(tx *ent.Tx) error {
		u := tx.Activity.UpdateOneID(m.ID).
			SetType(int(m.Type)).
			SetStatus(int(types.ActivityStatusDraft)).
			SetApplyMode(int(m.ApplyMode)).
			SetCycleMode(int(m.CycleMode)).
			SetIsPermanent(m.IsPermanent).
			SetAllowedFixedLevel(nilToEmpty(m.AllowedFixedLevel)).
			SetAllowedAutoLevel(nilToEmpty(m.AllowedAutoLevel)).
			SetAllowedClients(nilToEmpty(m.AllowedClients)).
			SetAllowedGameIds(nilToEmpty(m.AllowedGameIds)).
			SetAllowedAgentIds(nilToEmpty(m.AllowedAgentIds)).
			SetApplyLimit(nilToEmpty(m.ApplyLimit)).
			SetClaimLimit(nilToEmpty(m.ClaimLimit)).
			SetNillableOpenAt(m.OpenAt).
			SetNillableCloseAt(m.CloseAt).
			SetNillableReleaseAt(m.ReleaseAt).
			SetAutoAudit(m.AutoAudit).
			SetAutoAuditMaxAmt(m.AutoAuditMaxAmt).
			SetWageringMultiplier(m.WageringMultiplier).
			SetWageringBase(int(m.WageringBase)).
			SetPayoutMode(int(m.PayoutMode)).
			SetPayoutTiming(int(m.PayoutTiming)).
			SetSection(int(m.Section)).
			SetClaimExpireSec(m.ClaimExpireSec).
			SetSettings(defaultJSONObject(m.Settings)).
			SetShowInList(m.ShowInList).
			SetShowInSidebar(m.ShowInSidebar).
			SetUpdateUID(m.UpdateUid).
			SetSortWeight(m.SortWeight).
			AddVersion(1)

		// 清空时间字段：永久活动或显式置空时需要 Clear
		if m.OpenAt == nil {
			u.ClearOpenAt()
		}
		if m.CloseAt == nil {
			u.ClearCloseAt()
		}
		if m.ReleaseAt == nil {
			u.ClearReleaseAt()
		}

		if _, err := u.Save(ctx); err != nil {
			return fmt.Errorf("update activity: %w", err)
		}

		if _, err := tx.ActivityI18n.Delete().
			Where(activityi18n.ActivityIDEQ(m.ID)).
			Exec(ctx); err != nil {
			return fmt.Errorf("delete activity i18n: %w", err)
		}

		for _, item := range i18nList {
			if item.Lang == "" {
				continue
			}
			if _, err := tx.ActivityI18n.Create().
				SetActivityID(m.ID).
				SetLang(item.Lang).
				SetName(item.Name).
				SetRule(item.Rule).
				SetDescription(item.Description).
				SetImage(item.Image).
				Save(ctx); err != nil {
				return fmt.Errorf("create activity i18n: %w", err)
			}
		}
		return nil
	})
}

// CloseActivity 关闭活动；已关闭幂等成功
func (s *AdminService) CloseActivity(ctx context.Context, id int64, operatorCode string) error {
	row, err := s.db.Activity.Get(ctx, id)
	if err != nil {
		return err
	}
	if row.Status == int(types.ActivityStatusClosed) {
		return nil
	}

	wasOngoing := row.Status == int(types.ActivityStatusOngoing)
	activityType := types.ActivityType(row.Type)

	if _, err := s.db.Activity.UpdateOneID(id).
		SetStatus(int(types.ActivityStatusClosed)).
		Save(ctx); err != nil {
		return err
	}

	if wasOngoing {
		activitycache.DelOngoingActivitiesCache(operatorCode, activityType)
	}
	return nil
}

// PublishActivity 发布活动：仅草稿可发布为进行中
func (s *AdminService) PublishActivity(ctx context.Context, id int64, operatorCode string, updateUID int64) error {
	row, err := s.db.Activity.Get(ctx, id)
	if err != nil {
		return err
	}
	if row.Status != int(types.ActivityStatusDraft) {
		return ErrActivityNotDraft
	}

	now := time.Now()
	activityType := types.ActivityType(row.Type)
	if _, err := s.db.Activity.UpdateOneID(id).
		SetStatus(int(types.ActivityStatusOngoing)).
		SetReleaseAt(now).
		SetUpdateUID(updateUID).
		Save(ctx); err != nil {
		return err
	}

	activitycache.DelOngoingActivitiesCache(operatorCode, activityType)
	return nil
}

// allowedClientsContainsOrEmpty 空数组表示不限制，否则 JSON 包含 client
func allowedClientsContainsOrEmpty(clientID int64) predicate.Activity {
	return predicate.Activity(func(s *sql.Selector) {
		col := s.C(activity.FieldAllowedClients)
		s.Where(sql.ExprP(
			fmt.Sprintf("(%s = '[]'::jsonb OR %s @> ?::jsonb)", col, col),
			fmt.Sprintf("[%d]", clientID),
		))
	})
}

func nilToEmpty(v []int64) []int64 {
	if v == nil {
		return []int64{}
	}
	return v
}

func defaultJSONObject(s string) string {
	if s == "" {
		return "{}"
	}
	return s
}
