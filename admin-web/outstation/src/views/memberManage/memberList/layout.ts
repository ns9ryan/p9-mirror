import { useStorage } from '@vueuse/core';
import type { MemberLayout } from './types';

/** 定版后改这里切换默认版本 */
export const DEFAULT_MEMBER_LAYOUT: MemberLayout = 'classic';

const layout = useStorage<MemberLayout>(
  'memberList.layout',
  DEFAULT_MEMBER_LAYOUT
);

/** `?layout=v2` 优先于本地记忆，方便直接分享新版链接 */
export function useMemberLayout(queryValue?: unknown) {
  if (queryValue === 'classic' || queryValue === 'v2') {
    layout.value = queryValue;
  }
  return layout;
}
