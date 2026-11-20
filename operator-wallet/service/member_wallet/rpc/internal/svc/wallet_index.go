package svc

import (
	"fmt"

	"github.com/google/uuid"
)

// walletIndexKey 使用厅和会员全局标识定位钱包
type walletIndexKey struct {
	operatorCode   string    // 所属厅业务编码
	memberGlobalID uuid.UUID // CRM会员全局唯一标识
}

// walletIndexEntry 保存钱包的不可变标识信息
type walletIndexEntry struct {
	walletGlobalID uuid.UUID // 钱包全局唯一标识
	currencyCode   string    // 钱包结算币种编码
}

// CachedWallet 获取已缓存的钱包标识
func (s *ServiceContext) CachedWallet(operatorCode string, memberGlobalID uuid.UUID) (uuid.UUID, string, bool) {
	if s == nil {
		return uuid.Nil, "", false
	}

	key := walletIndexKey{
		operatorCode:   operatorCode,   // 所属厅业务编码
		memberGlobalID: memberGlobalID, // CRM会员全局唯一标识
	}

	s.walletIndexMu.RLock()
	entry, ok := s.walletIndex[key]
	s.walletIndexMu.RUnlock()

	if !ok {
		return uuid.Nil, "", false
	}

	return entry.walletGlobalID, entry.currencyCode, true
}

// CacheWallet 缓存会员与钱包的映射关系
func (s *ServiceContext) CacheWallet(operatorCode string, memberGlobalID uuid.UUID, walletGlobalID uuid.UUID, currencyCode string) error {
	if s == nil {
		return fmt.Errorf("cache wallet: service context is nil")
	}

	key := walletIndexKey{
		operatorCode:   operatorCode,   // 所属厅业务编码
		memberGlobalID: memberGlobalID, // CRM会员全局唯一标识
	}

	entry := walletIndexEntry{
		walletGlobalID: walletGlobalID, // 钱包全局唯一标识
		currencyCode:   currencyCode,   // 钱包结算币种编码
	}

	s.walletIndexMu.Lock()
	defer s.walletIndexMu.Unlock()

	if s.walletIndex == nil {
		s.walletIndex = make(map[walletIndexKey]walletIndexEntry)
	}

	if existing, ok := s.walletIndex[key]; ok {
		if existing != entry {
			return fmt.Errorf("wallet index conflict: operator_code=%s, member_global_id=%s", operatorCode, memberGlobalID)
		}

		return nil
	}

	s.walletIndex[key] = entry
	return nil
}
