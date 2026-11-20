package activityservicelogic

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

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

type UpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLogic {
	return &UpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Update 编辑活动（仅草稿）
func (l *UpdateLogic) Update(in *activitypb.UpdateActivityReq) (*activitypb.UpdateActivityResp, error) {
	claims, ok := ctxdata.OperatorClaimsFromCtx(l.ctx)
	if !ok {
		return nil, xerr.RpcErr(xerr.Unauthorized(i18n.Unauthorized))
	}
	if in.GetId() <= 0 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidID))
	}
	if err := validateUpdateActivityReq(in); err != nil {
		return nil, xerr.RpcErr(err)
	}

	activityModel, err := buildUpdateActivityModel(claims.OperatorCode, claims.UserID, in)
	if err != nil {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

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
	if err := adminServ.UpdateActivity(l.ctx, activityModel, i18nList); err != nil {
		switch {
		case errors.Is(err, admin.ErrActivityNotDraft):
			return nil, xerr.RpcErr(xerr.BadRequest(i18n.NotDraft))
		case ent.IsNotFound(err):
			return nil, xerr.RpcErr(xerr.EntNotFound(i18n.NotFound, err))
		case ent.IsConstraintError(err):
			return nil, xerr.RpcErr(xerr.EntConstraintError(i18n.InvalidParam, err))
		case ent.IsValidationError(err):
			return nil, xerr.RpcErr(xerr.EntValidationError(i18n.InvalidParam, err))
		default:
			return nil, xerr.RpcErr(err)
		}
	}

	return &activitypb.UpdateActivityResp{Id: in.GetId()}, nil
}

func validateUpdateActivityReq(in *activitypb.UpdateActivityReq) error {
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

func buildUpdateActivityModel(operatorCode string, updateUID int64, in *activitypb.UpdateActivityReq) (*repository.ActivityModel, error) {
	openAtSec := in.OpenAtSec
	closeAtSec := in.CloseAtSec
	releaseAtSec := in.ReleaseAtSec
	// 草稿编辑不发布；永久活动不设置开放/结束时间
	releaseAtSec = nil
	if in.GetIsPermanent() {
		openAtSec = nil
		closeAtSec = nil
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

	return &repository.ActivityModel{
		ID:                 in.GetId(),
		OperatorCode:       operatorCode,
		Type:               types.ActivityType(in.GetType()),
		Status:             types.ActivityStatusDraft,
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
		OpenAt:             secToTimePtr(openAtSec),
		CloseAt:            secToTimePtr(closeAtSec),
		ReleaseAt:          secToTimePtr(releaseAtSec),
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
		UpdateUid:          updateUID,
		SortWeight:         int(in.GetSortWeight()),
	}, nil
}
