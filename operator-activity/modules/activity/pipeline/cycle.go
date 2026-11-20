package pipeline

import (
	"time"

	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

// Manager 循环周期管理器
type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

// CurrentRound 计算当前轮次与时间窗
func (m *Manager) CurrentRound(activity *types.Activity, eventTime time.Time) (round int, window types.TimeWindow, ok bool) {
	tz := time.Local
	t := eventTime.In(tz)
	openAt := timeOrZero(activity.OpenAt).In(tz)

	switch activity.CycleMode {
	case types.CycleUnspecified, types.CycleNone:
		return 1, fullWindow(activity), true

	case types.CycleDaily:
		start := startOfDay(t)
		round = daysBetween(openAt, t) + 1
		window = types.TimeWindow{Start: start, End: start.Add(24 * time.Hour)}
		return round, window, m.withinLimit(activity, round)

	case types.CycleWeekly:
		start := startOfWeek(t)
		round = weeksBetween(openAt, t) + 1
		window = types.TimeWindow{Start: start, End: start.Add(7 * 24 * time.Hour)}
		return round, window, m.withinLimit(activity, round)

	case types.CycleMonthly:
		start := startOfMonth(t)
		round = monthsBetween(openAt, t) + 1
		end := start.AddDate(0, 1, 0)
		window = types.TimeWindow{Start: start, End: end}
		return round, window, m.withinLimit(activity, round)

	case types.CycleInfinite:
		round = 1
		return round, fullWindow(activity), true

	default:
		return 1, fullWindow(activity), true
	}
}

func (m *Manager) withinLimit(activity *types.Activity, round int) bool {
	// if activity.CycleMode == types.CycleDaily && round > 30 {
	// 	return false
	// }
	return true
}

func fullWindow(activity *types.Activity) types.TimeWindow {
	end := timeOrZero(activity.CloseAt)
	if activity.IsPermanent {
		end = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
	}
	return types.TimeWindow{Start: timeOrZero(activity.OpenAt), End: end}
}

func timeOrZero(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func startOfWeek(t time.Time) time.Time {
	day := startOfDay(t)
	offset := int(day.Weekday())
	if offset == 0 {
		offset = 7
	}
	return day.AddDate(0, 0, -(offset - 1))
}

func startOfMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
}

func daysBetween(start, end time.Time) int {
	s := startOfDay(start)
	e := startOfDay(end)
	return int(e.Sub(s).Hours() / 24)
}

func weeksBetween(start, end time.Time) int {
	return daysBetween(startOfWeek(start), startOfWeek(end)) / 7
}

func monthsBetween(start, end time.Time) int {
	sy, sm, _ := start.Date()
	ey, em, _ := end.Date()
	return (ey-sy)*12 + int(em-sm)
}
