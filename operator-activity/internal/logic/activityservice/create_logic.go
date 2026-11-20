package activityservicelogic

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"oa.98ent.com/p9/operator-activity/ent"
	"oa.98ent.com/p9/operator-activity/internal/svc"
	"oa.98ent.com/p9/operator-activity/modules/activity/repository"
	"oa.98ent.com/p9/operator-activity/modules/activity/repository/admin"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/activitypb"
	"oa.98ent.com/p9/operator-activity/pkg/i18n"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
)

type CreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLogic {
	return &CreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Create 新增活动
func (l *CreateLogic) Create(in *activitypb.CreateActivityReq) (*activitypb.CreateActivityResp, error) {
	claims, ok := ctxdata.OperatorClaimsFromCtx(l.ctx)
	if !ok {
		return nil, xerr.RpcErr(xerr.Unauthorized(i18n.Unauthorized))
	}

	if err := validateCreateActivityReq(in); err != nil {
		return nil, xerr.RpcErr(err)
	}

	activityModel, err := buildActivityModel(claims.OperatorCode, claims.UserID, in)
	if err != nil {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}
	// 名称、规则多语言处理
	i18nList := make([]admin.ActivityI18nInput, 0, len(in.GetI18N()))
	for _, item := range in.GetI18N() {
		if item == nil {
			continue
		}
		lang := strings.TrimSpace(item.GetLang())
		if lang == "" {
			continue
		}
		i18nList = append(i18nList, admin.ActivityI18nInput{
			Lang:        lang,
			Name:        strings.TrimSpace(item.GetName()),
			Rule:        item.GetRule(),
			Description: item.GetDescription(),
			Image:       strings.TrimSpace(item.GetImage()),
		})
	}

	adminServ := admin.NewService(l.svcCtx.DB)
	id, err := adminServ.CreateActivity(l.ctx, activityModel, i18nList)
	if err != nil {
		switch {
		case ent.IsConstraintError(err):
			return nil, xerr.RpcErr(xerr.EntConstraintError(i18n.InvalidParam, err))
		case ent.IsValidationError(err):
			return nil, xerr.RpcErr(xerr.EntValidationError(i18n.InvalidParam, err))
		default:
			return nil, xerr.RpcErr(err)
		}
	}

	return &activitypb.CreateActivityResp{Id: id}, nil
}

// 验证创建活动请求
func validateCreateActivityReq(in *activitypb.CreateActivityReq) error {
	activityType := types.ActivityType(in.GetType())
	switch activityType {
	case types.ActivityTypeDeposit,
		types.ActivityTypeWagering,
		types.ActivityTypeReferralMember,
		types.ActivityTypeLuckyWheel,
		types.ActivityTypeAssist,
		types.ActivityTypeBlindBox,
		types.ActivityTypeNewbieBonus,
		types.ActivityTypeUniversalAgent,
		types.ActivityTypeRescueFund,
		types.ActivityTypeLuckyBet,
		types.ActivityTypeVIPBonus,
		types.ActivityTypeRedPacketRain,
		types.ActivityTypeImageTextAd,
		types.ActivityTypeCheckIn:
	default:
		return xerr.BadRequest(i18n.InvalidType)
	}
	if in.GetApplyMode() <= 0 {
		return xerr.BadRequest(i18n.InvalidApplyMode)
	}
	if !validSection(in.GetSection()) {
		return xerr.BadRequest(i18n.InvalidSection)
	}
	if activityType == types.ActivityTypeLuckyWheel {
		if in.GetSettings() == nil || in.GetSettings().GetLuckWheel() == nil {
			return xerr.BadRequest(i18n.LuckWheelSettingsRequired)
		}
	}
	return nil
}

// 验证活动所属
func validSection(v int32) bool {
	switch types.ActivitySection(v) {
	case types.SectionLobby, types.SectionElectronic, types.SectionSports:
		return true
	default:
		return false
	}
}

// 构建活动模型 显示列表默认true、侧边栏默认false、状态默认草稿
func buildActivityModel(operatorCode string, createUID int64, in *activitypb.CreateActivityReq) (*repository.ActivityModel, error) {
	status := types.ActivityStatus(in.GetStatus())
	if status == 0 || status == types.ActivityStatusDraft {
		status = types.ActivityStatusDraft
		in.ReleaseAtSec = nil
	}
	// 永久活动不设置开放时间
	if in.IsPermanent {
		in.OpenAtSec = nil
		in.CloseAtSec = nil
	}

	settingsJSON := "{}"
	if in.GetSettings() != nil {
		settings, err := toActivitySettings(in.GetSettings())
		if err != nil {
			return nil, err
		}
		raw, err := json.Marshal(settings)
		if err != nil {
			return nil, err
		}
		settingsJSON = string(raw)
	}

	showInList := true
	if in.ShowInList != nil {
		showInList = in.GetShowInList()
	}
	showInSidebar := false
	if in.ShowInSidebar != nil {
		showInSidebar = in.GetShowInSidebar()
	}

	createUIDValue := in.GetCreateUid()
	if createUIDValue <= 0 {
		createUIDValue = createUID
	}

	return &repository.ActivityModel{
		OperatorCode:       operatorCode,
		Type:               types.ActivityType(in.GetType()),
		Status:             status,
		ApplyMode:          types.ApplyMode(in.GetApplyMode()),
		CycleMode:          types.CycleMode(in.GetCycleMode()),
		AllowedFixedLevel:  in.GetAllowedFixedLevel(),
		AllowedAutoLevel:   in.GetAllowedAutoLevel(),
		AllowedClients:     in.GetAllowedClients(),
		AllowedGameIds:     in.GetAllowedGameIds(),
		AllowedAgentIds:    in.GetAllowedAgentIds(),
		ApplyLimit:         in.GetApplyLimit(),
		ClaimLimit:         in.GetClaimLimit(),
		IsPermanent:        in.GetIsPermanent(),
		OpenAt:             secToTimePtr(in.OpenAtSec),
		CloseAt:            secToTimePtr(in.CloseAtSec),
		ReleaseAt:          secToTimePtr(in.ReleaseAtSec),
		AutoAudit:          in.GetAutoAudit(),
		AutoAuditMaxAmt:    in.GetAutoAuditMaxAmt(),
		WageringMultiplier: int(in.GetWageringMultiplier()),
		WageringBase:       types.WageringBase(in.GetWageringBase()),
		PayoutMode:         types.PayoutMode(in.GetPayoutMode()),
		PayoutTiming:       types.PayoutTiming(in.GetPayoutTiming()),
		Section:            types.ActivitySection(in.GetSection()),
		ClaimExpireSec:     in.GetClaimExpireSec(),
		Settings:           settingsJSON,
		ShowInList:         showInList,
		ShowInSidebar:      showInSidebar,
		CreateUid:          createUIDValue,
		SortWeight:         int(in.GetSortWeight()),
	}, nil
}

// 转换活动设置
func toActivitySettings(in *activitypb.ActivitySettings) (types.ActivitySettings, error) {
	out := types.ActivitySettings{
		DepositRewardType: types.DepositRewardType(in.GetDepositRewardType()),
		BetClaimMode:      types.BetClaimMode(in.GetBetClaimMode()),
		BonusMode:         types.BonusMode(in.GetBonusMode()),
		RewardSchemas:     toRewardSchemas(in.GetRewardSchemas()),
	}
	if in.ReferConfig != nil {
		out.ReferConfig = &types.ReferConfig{
			MinDepositAmount: in.GetReferConfig().GetMinDepositAmount(),
			MinBetAmount:     in.GetReferConfig().GetMinBetAmount(),
			IPLimit:          int(in.GetReferConfig().GetIpLimit()),
			DeviceLimit:      int(in.GetReferConfig().GetDeviceLimit()),
		}
	}
	if in.LuckWheel != nil {
		lw := in.GetLuckWheel()
		out.LuckWheel = &types.LuckWheelConfig{
			WheelTriggerMode: types.WheelTriggerMode(lw.GetWheelTriggerMode()),
			BetSchemas:       toRewardSchemas(lw.GetBetSchemas()),
			DepositSchemas:   toRewardSchemas(lw.GetDepositSchemas()),
		}
	}
	return out, nil
}

// 转换奖励方案
func toRewardSchemas(items []*activitypb.RewardSchema) []types.RewardSchema {
	if len(items) == 0 {
		return nil
	}
	out := make([]types.RewardSchema, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, types.RewardSchema{
			ID:        item.GetId(),
			Threshold: item.GetThreshold(),
			Amount:    item.GetAmount(),
			Rate:      int(item.GetRate()),
			MaxBonus:  item.GetMaxBonus(),
			MinBonus:  item.GetMinBonus(),
			Count:     int(item.GetCount()),
		})
	}
	return out
}

// 转换秒为时间
func secToTimePtr(sec *int64) *time.Time {
	if sec == nil || *sec <= 0 {
		return nil
	}
	t := time.Unix(*sec, 0)
	return &t
}
