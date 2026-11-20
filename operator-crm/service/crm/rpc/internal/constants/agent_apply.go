package constants

const (
	AuditStatusPending      int64 = 1 // 待审核
	AuditStatusApproved     int64 = 2 // 已通过
	AuditStatusManualReject int64 = 3 // 人工拒绝
	AuditStatusRiskReject   int64 = 4 // 风控拒绝

	ReviewActionApprove      int32 = 1 // 通过
	ReviewActionManualReject int32 = 2 // 人工拒绝
	ReviewActionRiskReject   int32 = 3 // 风控拒绝

	AccountTypeAgent  int32 = 1 // 代理账号
	AccountTypeMember int32 = 2 // 会员账号
)
