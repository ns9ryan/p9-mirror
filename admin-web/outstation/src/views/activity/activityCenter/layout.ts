import { useStorage } from '@vueuse/core';
import type { ActivityLayout } from './types';

/** 定版后改这里切换默认版本 */
export const DEFAULT_ACTIVITY_LAYOUT: ActivityLayout = 'classic';

const layout = useStorage<ActivityLayout>(
  'activityCenter.layout',
  DEFAULT_ACTIVITY_LAYOUT
);

/** `?layout=v2` 优先于本地记忆，方便直接分享新版链接 */
export function useActivityLayout(queryValue?: unknown) {
  if (queryValue === 'classic' || queryValue === 'v2') {
    layout.value = queryValue;
  }
  return layout;
}
