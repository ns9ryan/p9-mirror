package cache

import (
	"fmt"

	"oa.98ent.com/p9/common/modules/cache"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

// 进行中活动缓存前缀
const ongoingActivityCachePrefix = "activity:ongoing:"

// OngoingActivityCacheKey 生成指定类型活动的缓存键
func OngoingActivityCacheKey(operatorCode string, activityType types.ActivityType) string {
	return fmt.Sprintf("%s%s:%d", ongoingActivityCachePrefix, operatorCode, int(activityType))
}

// DelOngoingActivitiesCache 删除指定类型活动的缓存
func DelOngoingActivitiesCache(operatorCode string, activityType types.ActivityType) {
	cache.TwoMinuteCache.Cache.Delete(OngoingActivityCacheKey(operatorCode, activityType))
}

// GetOngoingActivitiesCache 获取指定类型活动的缓存，未命中时调用 fn 加载。
func GetOngoingActivitiesCache(operatorCode string, activityType types.ActivityType, fn func(_ map[string]any) (any, error)) ([]*types.Activity, error) {
	key := OngoingActivityCacheKey(operatorCode, activityType)
	raw, err := cache.TwoMinuteCache.GetC(key, fn, nil, false)
	if err != nil {
		return nil, err
	}
	activities, ok := raw.([]*types.Activity)
	if !ok {
		loaded, err := fn(nil)
		if err != nil {
			return nil, err
		}
		activities, ok = loaded.([]*types.Activity)
		if !ok {
			return nil, fmt.Errorf("ongoing activity cache: unexpected type %T", raw)
		}
	}
	return cloneActivities(activities), nil
}

func cloneActivities(src []*types.Activity) []*types.Activity {
	if len(src) == 0 {
		return []*types.Activity{}
	}
	out := make([]*types.Activity, len(src))
	for i, p := range src {
		if p == nil {
			continue
		}
		cp := *p
		out[i] = &cp
	}
	return out
}
