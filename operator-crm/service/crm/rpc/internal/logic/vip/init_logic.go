package viplogic

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-crm/ent/membervipconfig"
	"oa.98ent.com/p9/operator-crm/ent/memberviplevel"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/vippb"
)

// InitLogic 分站VIP初始化逻辑
type InitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewInitLogic 创建分站VIP初始化逻辑
func NewInitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitLogic {
	return &InitLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Init 初始化当前分站VIP基础配置及VIP0等级
func (l *InitLogic) Init(in *vippb.InitRequest) (*vippb.InitResponse, error) {
	// 校验初始化参数
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "初始化参数不能为空")
	}

	// 校验分站编码, 不允许为空, 包含首尾空格或超过64个字符
	operatorCode := in.GetOperatorCode()
	if operatorCode == "" || strings.TrimSpace(operatorCode) != operatorCode || utf8.RuneCountInString(operatorCode) > 64 {
		return nil, status.Error(codes.InvalidArgument, "分站编码无效")
	}

	// 校验请求分站编码与RPC上下文一致
	if ctxdata.OperatorCodeFromCtx(l.ctx) != operatorCode {
		return nil, status.Error(codes.PermissionDenied, "分站编码不匹配")
	}

	// 启动事务
	tx, err := l.svcCtx.DB.Tx(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("开启VIP初始化事务失败: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// 初始化VIP基础配置, 已存在时不覆盖
	err = tx.MemberVipConfig.Create().
		OnConflictColumns(membervipconfig.FieldOperatorCode).
		DoNothing().
		Exec(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("初始化VIP基础配置失败: %w", err)
	}

	// 初始化VIP0等级, 已存在时不覆盖
	err = tx.MemberVipLevel.Create().
		SetLevel(0).            // VIP等级编号, 从0开始连续递增
		SetUpgradeDeposit(0).   // 晋级所需存款金额
		SetUpgradeWater(0).     // 晋级所需洗码量
		SetRetentionDeposit(0). // 保级所需存款金额
		SetRetentionWater(0).   // 保级所需洗码量
		SetUpgradeReward(0).    // 晋级奖励金额
		OnConflictColumns(memberviplevel.FieldOperatorCode, memberviplevel.FieldLevel).
		DoNothing().
		Exec(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("初始化VIP0等级失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交VIP初始化事务失败: %w", err)
	}

	l.Logger.Infow("分站VIP初始化完成", logx.Field("operator_code", operatorCode))

	return &vippb.InitResponse{}, nil
}
