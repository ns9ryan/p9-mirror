// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package deposit_order

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-admin-api/internal/constants"
	agentconv "oa.98ent.com/p9/operator-admin-api/internal/logic/agent"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentdepositorderpb"
)

// depositTypeLabel 存款方式文字映射
var depositTypeLabel = map[int32]string{
	1: "加密货币",
	2: "人工入账",
	3: "第三方代收",
}

// depositStatusLabel 订单状态文字映射
var depositStatusLabel = map[int32]string{
	1:  "待审核",
	2:  "冻结",
	3:  "解冻",
	4:  "通过",
	5:  "拒绝",
	6:  "支付成功",
	7:  "支付失败",
	8:  "支付拒绝",
	11: "待一审",
	12: "待二审",
}

// depositCSVHeader 固定15列表头（中文）
var depositCSVHeader = []string{
	"订单号", "会员账号", "会员层级", "存款时间",
	"存款货币", "存款数量",
	"存款方式", "到账金额", "手续费", "实际到账金额",
	"三方订单号", "审核人", "审核时间", "创建IP", "订单状态",
}

type ExportAgentDepositOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 导出存款订单
func NewExportAgentDepositOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExportAgentDepositOrdersLogic {
	return &ExportAgentDepositOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ExportAgentDepositOrdersLogic) ExportAgentDepositOrders(req *types.ExportAgentDepositOrdersReq) (resp *types.ExportAgentDepositOrdersResp, err error) {
	if l.svcCtx.Config.OssApiBaseURL == "" {
		return nil, xerr.Err(http.StatusServiceUnavailable, "OSS not configured")
	}
	claims := ctxdata.ClaimsFromCtx(l.ctx)
	if claims == nil {
		return nil, xerr.Unauthorized("unauthorized")
	}

	taskID := uuid.New().String()

	// 导出固定使用 DepositOrderModeManage 的状态范围，不接受外部 Statuses。
	// Status 由调用方传入（可选）：在允许集合内有效则单状态过滤，否则兜底用 mode 全量状态。
	allowed := modeAllowedStatuses(constants.DepositOrderModeManage)
	pbReq := &agentdepositorderpb.ExportDepositOrdersReq{
		TimeType:          req.TimeType,
		TimeStart:         req.TimeStart,
		TimeEnd:           req.TimeEnd,
		MemberId:          req.MemberId,
		DepositType:       req.DepositType,
		OrderNo:           req.OrderNo,
		MemberAccount:     req.MemberAccount,
		ThirdPartyOrderNo: req.ThirdPartyOrderNo,
	}
	if req.Status != 0 {
		filtered := filterStatus(req.Status, allowed)
		if filtered == 0 {
			pbReq.Statuses = modeAllowedKeys(allowed)
		} else {
			pbReq.Status = filtered
		}
	} else {
		pbReq.Statuses = modeAllowedKeys(allowed)
	}

	rpcCtx, rpcCancel := context.WithTimeout(l.ctx, constants.DepositExportRPCTimeout)
	defer rpcCancel()
	rpcResp, err := l.svcCtx.AgentDepositOrderRpc.ExportDepositOrders(rpcCtx, pbReq)
	if err != nil {
		return nil, err
	}

	result, err := doDepositExport(l.ctx, l.svcCtx, taskID, claims.OperatorCode, rpcResp.GetList())
	if err != nil {
		return nil, err
	}

	return &types.ExportAgentDepositOrdersResp{
		TaskId:     taskID,
		Status:     "completed",
		Url:        result.URL,
		Key:        result.Key,
		Rows:       rpcResp.GetRows(),
		ExpiresAt:  result.URLExpiresIn * 1000, // 转为毫秒
		Visibility: constants.DepositExportVisibility,
	}, nil
}

// doDepositExport 将 RPC 返回的存款订单列表写成 CSV 并上传 OSS。
// 内部自行查询站点结算货币倍率（严格模式，查询失败直接报错）。
func doDepositExport(ctx context.Context, svcCtx *svc.ServiceContext, taskID, operatorCode string, list []*agentdepositorderpb.ExportDepositOrderItem) (*OssUploadResult, error) {
	// 查询结算货币倍率：取列表第一项的 DepositCurrency 与站点配置核验后查询倍率
	var exportCurrencyCode string
	if len(list) > 0 {
		exportCurrencyCode = list[0].DepositCurrency
	}
	amountFactorI64, factorErr := agentconv.GetSiteCurrencyFactor(ctx, svcCtx, exportCurrencyCode)
	if factorErr != nil {
		return nil, factorErr
	}
	amountFactor := float64(amountFactorI64)

	// 写入 UTF-8 BOM，确保 Excel 正确识别中文
	tmpFile, err := os.CreateTemp("", "deposit-export-*.csv")
	if err != nil {
		return nil, xerr.InternalServerError("export temporary file failed")
	}
	if _, err = tmpFile.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return nil, xerr.InternalServerError("export BOM write failed")
	}
	defer func() {
		tmpFile.Close()
		if e := os.Remove(tmpFile.Name()); e != nil {
			logx.WithContext(ctx).Errorw("deposit export cleanup failed", logx.Field("task_id", taskID))
		}
	}()

	w := csv.NewWriter(tmpFile)
	w.UseCRLF = true
	if err = w.Write(depositCSVHeader); err != nil {
		return nil, xerr.InternalServerError("export CSV header failed")
	}
	for _, item := range list {
		createdAt := ""
		if item.CreatedAt > 0 {
			createdAt = time.Unix(item.CreatedAt, 0).UTC().Format(time.RFC3339)
		}
		finalAuditedAt := ""
		if item.FinalAuditedAt > 0 {
			finalAuditedAt = time.Unix(item.FinalAuditedAt, 0).UTC().Format(time.RFC3339)
		}
		dtLabel, ok := depositTypeLabel[item.DepositType]
		if !ok {
			dtLabel = fmt.Sprintf("%d", item.DepositType)
		}
		stLabel, ok := depositStatusLabel[item.Status]
		if !ok {
			stLabel = fmt.Sprintf("%d", item.Status)
		}
		row := []string{
			item.OrderNo,
			item.MemberAccount,
			item.MemberLevelName,
			createdAt,
			item.DepositCurrency,
			fmt.Sprintf("%.2f", float64(item.DepositAmount)/amountFactor),
			dtLabel,
			fmt.Sprintf("%.2f", float64(item.ReceivedAmount)/amountFactor),
			fmt.Sprintf("%.2f", float64(item.FeeAmount)/amountFactor),
			fmt.Sprintf("%.2f", float64(item.ActualReceivedAmount)/amountFactor),
			item.ThirdPartyOrderNo,
			item.AuditorUsername,
			finalAuditedAt,
			item.CreateIp,
			stLabel,
		}
		if err = w.Write(row); err != nil {
			return nil, xerr.InternalServerError("export CSV write failed")
		}
	}
	w.Flush()
	if err = w.Error(); err != nil {
		return nil, xerr.InternalServerError("export CSV flush failed")
	}

	stat, err := tmpFile.Stat()
	if err != nil {
		return nil, xerr.InternalServerError("export CSV stat failed")
	}
	size := stat.Size()

	ossCtx, ossCancel := context.WithTimeout(ctx, constants.DepositExportOSSTimeout)
	defer ossCancel()

	filename := taskID + ".csv"
	params := ossUploadParams{
		BaseURL:               svcCtx.Config.OssApiBaseURL,
		Filename:              filename,
		Scene:                 constants.DepositExportScene,
		MaxBytes:              constants.DepositExportMaxBytes,
		UploadPath:            constants.DepositExportUploadPath,
		AuthScopeHeader:       constants.DepositExportOSSAuthScopeHeader,
		AuthScope:             constants.DepositExportOSSAuthScope,
		UploadTimeout:         constants.DepositExportUploadTimeout,
		BusinessSuccessCode:   constants.DepositExportOSSBusinessSuccessCode,
		MaxResponseBytes:      constants.DepositExportMaxUploadResponseBytes,
		ResponseOverflowBytes: constants.DepositExportResponseOverflowBytes,
		Visibility:            constants.DepositExportVisibility,
		MaxDownloadTTL:        constants.DepositExportMaxDownloadTTL,
		KeyDateLayout:         constants.DepositExportKeyDateLayout,
		KeyDateGroup:          constants.DepositExportKeyDateGroup,
		OperatorCode:          operatorCode,
	}
	result, err := UploadCSVToOSS(ossCtx, params, tmpFile, size)
	if err != nil {
		return nil, err
	}
	if !ValidOSSResult(result, params) {
		return nil, xerr.Err(http.StatusBadGateway, "invalid private CSV upload result")
	}
	return result, nil
}
