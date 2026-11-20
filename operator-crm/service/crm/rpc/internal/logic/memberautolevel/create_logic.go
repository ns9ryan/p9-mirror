package memberautolevellogic

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberautolevelpb"
)

// CreateLogic 自动层级创建逻辑
type CreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewCreateLogic 创建自动层级创建逻辑
func NewCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLogic {
	return &CreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Create 创建会员自动层级
func (l *CreateLogic) Create(in *memberautolevelpb.CreateMemberAutoLevelRequest) (*memberautolevelpb.CreateMemberAutoLevelResponse, error) {
	if in == nil {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	// 校验层级名称和晋级门槛
	name := strings.TrimSpace(in.Name)
	if name == "" || in.MinDeposit < 0 || in.MinWater < 0 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	// 校验晋级条件, 两项门槛必须严格大于当前厅所有已有层级
	conflict, err := l.svcCtx.DAO.MemberAutoLevel.HasThresholdConflict(
		l.ctx,
		in.MinDeposit, // 晋级要求的累计存款金额
		in.MinWater,   // 晋级要求的累计洗码量
	)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	if conflict {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	// 创建会员自动层级
	level, err := l.svcCtx.DAO.MemberAutoLevel.Create(
		l.ctx,
		name,          // 自动层级名称
		in.MinDeposit, // 晋级要求的累计存款金额
		in.MinWater,   // 晋级要求的累计洗码量
	)
	if err != nil {
		switch {
		case ent.IsConstraintError(err):
			// 数据约束冲突，如晋级门槛重复
			return nil, xerr.RpcErr(xerr.EntConstraintError(i18n.InvalidParam, err))

		case ent.IsValidationError(err):
			// 字段校验失败
			return nil, xerr.RpcErr(xerr.EntValidationError(i18n.InvalidParam, err))

		default:
			// 其他错误统一处理
			return nil, xerr.RpcErr(err)
		}
	}

	// 返回层级ID
	return &memberautolevelpb.CreateMemberAutoLevelResponse{
		Id: level.ID, // 层级ID
	}, nil
}
