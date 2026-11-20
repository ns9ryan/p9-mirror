import { inject, provide, ref } from 'vue';
import type { InjectionKey, Ref } from 'vue';
import {
  MEMBER_STATUS,
  batchCreateMembers as batchCreateMembersApi,
  createMember as createMemberApi,
  getMemberDetail,
  getMemberImages,
  queryMemberList,
  resetMemberAuthErrors,
  updateMemberAutoLevel,
  updateMemberContact,
  updateMemberFixedLevels,
  updateMemberRemark,
  updateMemberRestriction,
  updateMemberStatus,
  type BatchCreateMembersParams,
  type BatchCreateMembersRes,
  type CreateMemberParams,
  type MemberAuthErrorField,
  type MemberImagesRes,
  type MemberListParams,
  type MemberRestriction,
  type UpdateMemberContactParams,
} from '@/api/member';
import toMemberRecord, { toMemberDetailRecord } from './transform';
import type { AccountStatus, MemberDetailRecord, MemberRecord } from './types';

const STATUS_CODES: Record<AccountStatus, number> = {
  normal: MEMBER_STATUS.NORMAL,
  frozen: MEMBER_STATUS.FROZEN,
  disabled: MEMBER_STATUS.DISABLED,
};

export interface StatusResult {
  /** 实际修改成功的会员数，已是目标状态的不计入 */
  changed: number;
  failed: number;
}

export interface LevelChange {
  /** 不传表示不修改 */
  autoLevelId?: number;
  /** 全量替换，空数组清空；不传表示不修改 */
  fixedLevelIds?: number[];
}

export interface MemberStore {
  /** 当前页会员 */
  members: Ref<MemberRecord[]>;
  /** 符合条件的会员总数 */
  total: Ref<number>;
  loading: Ref<boolean>;
  fetchMembers: (params: MemberListParams) => Promise<void>;
  /** 当前查看的会员详情，未加载或加载失败时为 null */
  detail: Ref<MemberDetailRecord | null>;
  detailLoading: Ref<boolean>;
  /** 传 null 时清空详情 */
  fetchDetail: (id: number | null) => Promise<void>;
  /** 优先取当前页，不在当前页时取已加载的详情 */
  findMember: (id: number | null) => MemberRecord | null;
  /**
   * 接口一次只改一个会员，批量时逐个提交，失败的不弹窗只计数。
   * 已是目标状态的会员跳过
   */
  updateStatus: (ids: number[], status: AccountStatus) => Promise<StatusResult>;
  createMember: (
    value: CreateMemberParams
  ) => Promise<{ id: number; code: string }>;
  /** 同步接口，调用方需负责禁用提交并处理超时 */
  batchCreateMembers: (
    value: BatchCreateMembersParams
  ) => Promise<BatchCreateMembersRes>;
  /** 以下修改成功后都会重新拉取该会员详情，同步当前页和已打开的详情 */
  updateContact: (value: UpdateMemberContactParams) => Promise<void>;
  /** 自动层级和固定层级分两个接口，先改自动层级；任一失败即抛出 */
  updateLevels: (id: number, value: LevelChange) => Promise<void>;
  updateRemark: (id: number, remark: string) => Promise<void>;
  setRestriction: (
    id: number,
    kind: MemberRestriction,
    disabled: boolean
  ) => Promise<void>;
  resetAuthErrors: (id: number, field: MemberAuthErrorField) => Promise<void>;
  /** 头像与头像框资源，未加载或加载失败时为空列表 */
  images: Ref<MemberImagesRes>;
  /** 默认只加载一次；force 时重新拉取最新图片地址 */
  loadImages: (force?: boolean) => Promise<void>;
}

const MEMBER_STORE_KEY: InjectionKey<MemberStore> = Symbol('memberStore');

function createMemberStore(): MemberStore {
  const members = ref<MemberRecord[]>([]);
  const total = ref(0);
  const loading = ref(false);
  let ticket = 0;

  const detail = ref<MemberDetailRecord | null>(null);
  const detailLoading = ref(false);
  let detailTicket = 0;

  async function fetchMembers(params: MemberListParams) {
    ticket += 1;
    const current = ticket;
    loading.value = true;
    try {
      const { data } = await queryMemberList(params);
      if (current !== ticket) return;
      members.value = (data?.list || []).map(toMemberRecord);
      total.value = data?.total || 0;
    } catch {
      if (current !== ticket) return;
      members.value = [];
      total.value = 0;
    } finally {
      if (current === ticket) loading.value = false;
    }
  }

  async function fetchDetail(id: number | null) {
    detailTicket += 1;
    const current = detailTicket;
    if (detail.value?.id !== id) detail.value = null;
    if (id == null) {
      detailLoading.value = false;
      return;
    }
    detailLoading.value = true;
    try {
      const { data } = await getMemberDetail(id);
      if (current !== detailTicket) return;
      detail.value = data ? toMemberDetailRecord(data) : null;
    } catch {
      if (current !== detailTicket) return;
      detail.value = null;
    } finally {
      if (current === detailTicket) detailLoading.value = false;
    }
  }

  function findMember(id: number | null) {
    if (id == null) return null;
    const record = members.value.find((item) => item.id === id);
    if (record) return record;
    return detail.value?.id === id ? detail.value : null;
  }

  function patchStatus(id: number, status: AccountStatus) {
    members.value = members.value.map((item) =>
      item.id === id ? { ...item, status } : item
    );
    if (detail.value?.id === id) {
      detail.value = { ...detail.value, status };
    }
  }

  async function updateStatus(ids: number[], status: AccountStatus) {
    const targets = ids.filter((id) => findMember(id)?.status !== status);
    const silent = ids.length > 1;
    const result: StatusResult = { changed: 0, failed: 0 };
    // 逐个提交，避免一次性并发大量请求
    await targets.reduce(async (previous, id) => {
      await previous;
      try {
        await updateMemberStatus(
          { member_id: id, status: STATUS_CODES[status] },
          silent
        );
        patchStatus(id, status);
        result.changed += 1;
      } catch {
        result.failed += 1;
      }
    }, Promise.resolve());
    return result;
  }

  /** 修改已成功，同步失败只影响展示，不再抛出 */
  async function syncMember(id: number) {
    try {
      const { data } = await getMemberDetail(id, true);
      if (!data) return;
      const record = toMemberRecord(data);
      members.value = members.value.map((item) =>
        item.id === id
          ? { ...record, phone: item.phone, email: item.email }
          : item
      );
      if (detail.value?.id === id) detail.value = toMemberDetailRecord(data);
    } catch {
      // 下次刷新列表或详情时恢复
    }
  }

  async function createMember(value: CreateMemberParams) {
    const { data } = await createMemberApi(value);
    return data;
  }

  async function batchCreateMembers(value: BatchCreateMembersParams) {
    const { data } = await batchCreateMembersApi(value);
    return data;
  }

  async function updateContact(value: UpdateMemberContactParams) {
    await updateMemberContact(value);
    await syncMember(value.member_id);
  }

  async function updateLevels(id: number, value: LevelChange) {
    try {
      if (value.autoLevelId != null) {
        await updateMemberAutoLevel({
          member_id: id,
          auto_level_id: value.autoLevelId,
        });
      }
      if (value.fixedLevelIds) {
        await updateMemberFixedLevels({
          member_id: id,
          fixed_level_ids: value.fixedLevelIds,
        });
      }
    } finally {
      await syncMember(id);
    }
  }

  async function updateRemark(id: number, remark: string) {
    await updateMemberRemark({ member_id: id, remark });
    await syncMember(id);
  }

  async function setRestriction(
    id: number,
    kind: MemberRestriction,
    disabled: boolean
  ) {
    await updateMemberRestriction(kind, { member_id: id, disabled });
    await syncMember(id);
  }

  async function resetAuthErrors(id: number, field: MemberAuthErrorField) {
    await resetMemberAuthErrors({ member_id: id, field });
  }

  const images = ref<MemberImagesRes>({ avatars: [], avatar_frames: [] });
  let imagesLoaded = false;
  let imagesRequest: Promise<void> | null = null;

  function loadImages(force = false) {
    if (imagesRequest) return imagesRequest;
    if (imagesLoaded && !force) return Promise.resolve();
    imagesRequest = (async () => {
      try {
        const { data } = await getMemberImages();
        images.value = {
          avatars: data?.avatars || [],
          avatar_frames: data?.avatar_frames || [],
        };
        imagesLoaded = true;
      } catch {
        // 保留上次结果，下次调用重试
      } finally {
        imagesRequest = null;
      }
    })();
    return imagesRequest;
  }

  return {
    members,
    total,
    loading,
    fetchMembers,
    detail,
    detailLoading,
    fetchDetail,
    findMember,
    updateStatus,
    createMember,
    batchCreateMembers,
    updateContact,
    updateLevels,
    updateRemark,
    setRestriction,
    resetAuthErrors,
    images,
    loadImages,
  };
}

export function provideMemberStore() {
  const store = createMemberStore();
  provide(MEMBER_STORE_KEY, store);
  return store;
}

export function useMemberStore() {
  const store = inject(MEMBER_STORE_KEY);
  if (!store) throw new Error('useMemberStore 需要在会员列表页面内使用');
  return store;
}
