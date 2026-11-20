import { useI18n } from 'vue-i18n';
import { Message, Modal } from '@arco-design/web-vue';
import { useActivityStore } from './store';
import type { ActivityRecord } from './types';

const PREFIX = 'discountCenter.activityCenter';

/** 发布、关闭都需二次确认，失败时保留弹窗，错误由全局拦截器提示 */
export default function useActivityActions() {
  const { t } = useI18n();
  const { publish, close } = useActivityStore();

  function confirm(
    kind: 'publish' | 'close',
    record: ActivityRecord,
    run: (id: number) => Promise<void>
  ) {
    Modal.confirm({
      title: t(`${PREFIX}.confirm.${kind}Title`),
      content: t(`${PREFIX}.confirm.${kind}Content`, {
        id: record.id,
        name: record.name || '---',
      }),
      okText: t(`${PREFIX}.confirm.ok`),
      cancelText: t(`${PREFIX}.confirm.cancel`),
      okButtonProps: kind === 'close' ? { status: 'danger' } : undefined,
      onBeforeOk: async () => {
        try {
          await run(record.id);
        } catch {
          return false;
        }
        Message.success(
          t(`${PREFIX}.message.${kind === 'publish' ? 'published' : 'closed'}`)
        );
        return true;
      },
    });
  }

  return {
    confirmPublish: (record: ActivityRecord) =>
      confirm('publish', record, publish),
    confirmClose: (record: ActivityRecord) => confirm('close', record, close),
  };
}
