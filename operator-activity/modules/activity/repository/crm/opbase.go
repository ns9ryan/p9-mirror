package crm

import (
	"context"
	"fmt"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
	"oa.98ent.com/p9/operator-base/rpc/pb/operatorbaserpc/operatorpb"
)

// LoadOpBaseInfo 加载分站基础信息。同一分站在进程内只请求一次。
func (s *CrmServ) LoadOpBaseInfo(ctx context.Context) (*types.OpBaseInfo, error) {
	code := ctxdata.OperatorCodeFromCtx(ctx)
	if code == "" {
		return nil, fmt.Errorf("缺少分站编码")
	}
	if info := s.cachedOpBase(code); info != nil {
		return info, nil
	}

	resp, err := s.opBaseRpc.GetOperatorInfo(ctx, &operatorpb.GetOperatorInfoRequest{})
	if err != nil {
		return nil, err
	}
	info := &types.OpBaseInfo{
		Currency: resp.SettlementCurrencyCode,
		TimeZone: resp.TimezoneCode,
	}

	s.opBaseMu.Lock()
	if existing := s.opBaseCache[code]; existing != nil {
		s.opBaseMu.Unlock()
		return existing, nil
	}
	s.opBaseCache[code] = info
	s.opBaseMu.Unlock()
	return info, nil
}

func (s *CrmServ) cachedOpBase(code string) *types.OpBaseInfo {
	s.opBaseMu.RLock()
	defer s.opBaseMu.RUnlock()
	return s.opBaseCache[code]
}
