package utils

import "time"

// GetTimeZoneLocation 获取时区位置
func GetTimeZoneLocation(timeZone string) *time.Location {
	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		loc = time.UTC
	}
	return loc
}

// PtrTimeInTZLoc 指针时间转对应时区时间
func PtrTimeInTZLoc(t *time.Time, timeZone string) *time.Time {
	if t != nil {
		loc := GetTimeZoneLocation(timeZone)
		zt := t.In(loc)
		return &zt
	}
	return nil
}
