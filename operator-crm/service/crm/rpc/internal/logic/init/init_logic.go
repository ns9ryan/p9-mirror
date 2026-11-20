package initlogic

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/initpb"
)

type InitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitLogic {
	return &InitLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Init 初始化分站CRM基础数据
func (l *InitLogic) Init(in *initpb.InitRequest) (*initpb.InitResponse, error) {
	// 获取当前分站编码
	operatorCode := ctxdata.OperatorCodeFromCtx(l.ctx)
	if operatorCode == "" {
		return nil, status.Error(codes.Unauthenticated, "缺少分站上下文")
	}

	// 所有初始化操作共用一个事务, 任何一步失败均回滚
	if err := dao.WithTx(l.ctx, l.svcCtx.DB, func(tx *ent.Tx) error {
		// 初始化默认自动层级
		if err := initMemberAutoLevel(l.ctx, tx); err != nil {
			return fmt.Errorf("初始化会员自动层级失败: %w", err)
		}

		// 初始化预设固定层级
		if err := initMemberFixedLevels(l.ctx, tx); err != nil {
			return fmt.Errorf("初始化会员固定层级失败: %w", err)
		}

		// 初始化VIP配置、VIP0及图标组
		if err := initMemberVip(l.ctx, tx); err != nil {
			return fmt.Errorf("初始化会员VIP数据失败: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("CRM初始化失败: %w", err)
	}

	l.Logger.Infow("分站CRM初始化完成", logx.Field("operator_code", operatorCode))

	return &initpb.InitResponse{}, nil
}
