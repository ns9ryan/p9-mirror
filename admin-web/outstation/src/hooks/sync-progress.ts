import { computed, onBeforeUnmount, ref } from 'vue';
import { getGameSyncCheckpointDetail } from '@/api/gameSyncCheckpoint';

/** 按 sync_scope 拉取并展示同步进度百分数 */
export default function useSyncProgress(syncScope: string) {
  /** 进页默认 100%，未请求接口前不算「接口确认完成」 */
  const syncProgress = ref(100);
  const syncProgressLoading = ref(false);
  /** 当前进度是否来自接口（用于区分默认 100% 与接口返回的 100%） */
  const syncProgressFromApi = ref(false);
  let progressTimer: ReturnType<typeof setTimeout> | null = null;

  /** 仅接口返回的 100% 视为完成：换色且不可点 */
  const isSyncProgressDone = computed(
    () => syncProgressFromApi.value && syncProgress.value >= 100
  );

  const normalizeProgress = (value: unknown) => {
    const num = Number(value);
    if (!Number.isFinite(num)) return 0;
    return Math.min(100, Math.max(0, Math.floor(num)));
  };

  const clearProgressTimer = () => {
    if (progressTimer != null) {
      clearTimeout(progressTimer);
      progressTimer = null;
    }
  };

  const fetchSyncProgress = async () => {
    if (syncProgressLoading.value) return;
    syncProgressLoading.value = true;
    try {
      const { data } = await getGameSyncCheckpointDetail(syncScope);
      syncProgress.value = normalizeProgress(data?.progress);
      syncProgressFromApi.value = true;
    } catch {
      // 错误提示由请求拦截器处理
    } finally {
      syncProgressLoading.value = false;
    }
  };

  /** 确认同步关弹窗后：延迟拉取真实进度 */
  const scheduleFetchSyncProgress = (delayMs = 1000) => {
    clearProgressTimer();
    syncProgressFromApi.value = false;
    syncProgress.value = 100;
    progressTimer = setTimeout(() => {
      progressTimer = null;
      fetchSyncProgress();
    }, delayMs);
  };

  onBeforeUnmount(() => {
    clearProgressTimer();
  });

  return {
    syncProgress,
    syncProgressLoading,
    syncProgressFromApi,
    isSyncProgressDone,
    fetchSyncProgress,
    scheduleFetchSyncProgress,
  };
}
