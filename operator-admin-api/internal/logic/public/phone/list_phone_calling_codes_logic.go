package phone

import (
	"context"

	"oa.98ent.com/p9/common/utils"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPhoneCallingCodesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 手机区号列表
func NewListPhoneCallingCodesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPhoneCallingCodesLogic {
	return &ListPhoneCallingCodesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListPhoneCallingCodesLogic) ListPhoneCallingCodes() (*types.ListPhoneCallingCodesResp, error) {
	codes := utils.CallingCodes()
	items := make([]types.PhoneCallingCodeItem, 0, len(codes))
	for _, c := range codes {
		items = append(items, types.PhoneCallingCodeItem{
			Region: c.Region,
			Code:   c.Code,
		})
	}
	return &types.ListPhoneCallingCodesResp{Items: items}, nil
}
