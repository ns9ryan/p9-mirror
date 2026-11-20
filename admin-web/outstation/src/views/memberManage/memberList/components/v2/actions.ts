import type { AccountStatus } from '../../types';
import type { OperationCommand } from '../operations/commands';

export type StatusCommand = 'freeze' | 'unfreeze' | 'disable' | 'enable';
export type MemberCommand = 'edit' | OperationCommand | StatusCommand;
export type BatchCommand = 'freeze' | 'disable' | 'enable';

/** 文案字段均为 i18n key */
interface CommandMeta {
  target: AccountStatus;
  label: string;
  confirmTitle: string;
  confirmContent: string;
  success: string;
}

export const STATUS_COMMANDS: Record<StatusCommand, CommandMeta> = {
  freeze: {
    target: 'frozen',
    label: 'memberManage.v2.action.freeze',
    confirmTitle: 'memberManage.v2.confirm.freezeTitle',
    confirmContent: 'memberManage.v2.confirm.freezeContent',
    success: 'memberManage.v2.message.frozen',
  },
  unfreeze: {
    target: 'normal',
    label: 'memberManage.v2.action.unfreeze',
    confirmTitle: 'memberManage.v2.confirm.unfreezeTitle',
    confirmContent: 'memberManage.v2.confirm.unfreezeContent',
    success: 'memberManage.v2.message.unfrozen',
  },
  disable: {
    target: 'disabled',
    label: 'memberManage.v2.action.disable',
    confirmTitle: 'memberManage.confirm.disableTitle',
    confirmContent: 'memberManage.confirm.disableContent',
    success: 'memberManage.message.disabled',
  },
  enable: {
    target: 'normal',
    label: 'memberManage.v2.action.enable',
    confirmTitle: 'memberManage.confirm.enableTitle',
    confirmContent: 'memberManage.confirm.enableContent',
    success: 'memberManage.message.enabled',
  },
};

export const BATCH_COMMANDS: Record<BatchCommand, CommandMeta> = {
  freeze: {
    target: 'frozen',
    label: 'memberManage.v2.toolbar.batchFreeze',
    confirmTitle: 'memberManage.v2.confirm.batchFreezeTitle',
    confirmContent: 'memberManage.v2.confirm.batchFreezeContent',
    success: 'memberManage.v2.message.batchFrozen',
  },
  disable: {
    target: 'disabled',
    label: 'memberManage.toolbar.batchDisable',
    confirmTitle: 'memberManage.confirm.batchDisableTitle',
    confirmContent: 'memberManage.confirm.batchDisableContent',
    success: 'memberManage.message.batchDisabled',
  },
  enable: {
    target: 'normal',
    label: 'memberManage.toolbar.batchEnable',
    confirmTitle: 'memberManage.confirm.batchEnableTitle',
    confirmContent: 'memberManage.confirm.batchEnableContent',
    success: 'memberManage.message.batchEnabled',
  },
};

export const BATCH_COMMAND_KEYS = Object.keys(BATCH_COMMANDS) as BatchCommand[];

export function statusCommandsOf(status: AccountStatus): StatusCommand[] {
  if (status === 'normal') return ['freeze', 'disable'];
  if (status === 'frozen') return ['unfreeze', 'disable'];
  return ['enable'];
}
