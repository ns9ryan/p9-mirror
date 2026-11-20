package types

// Int64sToApplyLimitTypes 将[]int64转成 []ApplyLimitType
func Int64sToApplyLimitTypes(ids []int64) []ApplyLimitType {
	types := make([]ApplyLimitType, 0, len(ids))
	for _, id := range ids {
		types = append(types, ApplyLimitType(id))
	}
	return types
}

// ApplyLimitTypesToInt64s 将[]ApplyLimitType []int64
func ApplyLimitTypesToInt64s(types []ApplyLimitType) []int64 {
	ids := make([]int64, 0, len(types))
	for _, tp := range types {
		ids = append(ids, int64(tp))
	}
	return ids
}

// Int64sToClaimLimitTypes 将[]int64转成 []ClaimLimitType
func Int64sToClaimLimitTypes(ids []int64) []ClaimLimitType {
	types := make([]ClaimLimitType, 0, len(ids))
	for _, id := range ids {
		types = append(types, ClaimLimitType(id))
	}
	return types
}

// ClaimLimitTypesToInt64s 将[]ClaimLimitType []int64
func ClaimLimitTypesToInt64s(types []ClaimLimitType) []int64 {
	ids := make([]int64, 0, len(types))
	for _, tp := range types {
		ids = append(ids, int64(tp))
	}
	return ids
}
