import { PROMO_HOST } from './options';
import type { AllAgentRecord } from './types';

const inherited = '从往期全民代理活动继承至活动986';

function row(
  partial: Pick<
    AllAgentRecord,
    | 'id'
    | 'subCount'
    | 'subTotal'
    | 'remark'
    | 'registerSource'
    | 'promoCode'
    | 'registeredAt'
    | 'lastLoginAt'
    | 'status'
  >
): AllAgentRecord {
  return {
    parentId: '',
    betLevel1: 0,
    betLevel2: 0,
    betLevel3: 0,
    claimedCommission: 0,
    pendingCommission: 0,
    commission: 0,
    frozenCommission: 0,
    cancelledCommission: 0,
    unclaimedCommission: 0,
    withdrawLimit: 'none',
    promoHost: PROMO_HOST,
    visitUv: 0,
    selfBetBanned: false,
    lineBetBanned: false,
    ...partial,
  };
}

export default function createAllAgentSeed(): AllAgentRecord[] {
  return [
    row({
      id: 51507567,
      subCount: 0,
      subTotal: 0,
      remark: '用户注册自动创建-活动986',
      registerSource: '用户注册',
      promoCode: 'I3Q9OOCEEK',
      registeredAt: '2026-10-07 16:18:45',
      lastLoginAt: '2026-10-07 16:18:45',
      status: 'normal',
    }),
    row({
      id: 81360777,
      subCount: 1,
      subTotal: 1,
      remark: inherited,
      registerSource: '全民代理活动',
      promoCode: '2LDE8MJU32',
      registeredAt: '2026-04-24 10:42:22',
      lastLoginAt: '2026-04-24 10:42:22',
      status: 'normal',
    }),
    row({
      id: 82360798,
      subCount: 1,
      subTotal: 1,
      remark: inherited,
      registerSource: '全民代理活动',
      promoCode: '4GCJMMMQW0',
      registeredAt: '2026-04-24 10:43:09',
      lastLoginAt: '2026-04-24 10:43:09',
      status: 'normal',
    }),
    row({
      id: 71339777,
      subCount: 4,
      subTotal: 8,
      remark: inherited,
      registerSource: '全民代理活动',
      promoCode: 'K7P2N8QW1A',
      registeredAt: '',
      lastLoginAt: '',
      status: 'unset',
    }),
    row({
      id: 79319155,
      subCount: 1,
      subTotal: 1,
      remark: inherited,
      registerSource: '全民代理活动',
      promoCode: 'M4C8R1TZ6B',
      registeredAt: '',
      lastLoginAt: '',
      status: 'unset',
    }),
    row({
      id: 70339756,
      subCount: 3,
      subTotal: 7,
      remark: inherited,
      registerSource: '全民代理活动',
      promoCode: 'Q9H3L5VX2D',
      registeredAt: '',
      lastLoginAt: '',
      status: 'unset',
    }),
    row({
      id: 53362079,
      subCount: 1,
      subTotal: 15,
      remark: inherited,
      registerSource: '全民代理活动',
      promoCode: '1H96DCFGMG',
      registeredAt: '2026-04-27 09:49:50',
      lastLoginAt: '2026-08-23 16:07:23',
      status: 'normal',
    }),
  ];
}
