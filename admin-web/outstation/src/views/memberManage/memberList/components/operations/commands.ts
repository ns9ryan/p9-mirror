export type OperationCommand = 'level' | 'remark' | 'restriction';

/** 菜单展示顺序；label 为 i18n key */
export const OPERATION_COMMANDS: { key: OperationCommand; label: string }[] = [
  { key: 'level', label: 'memberManage.ops.level.menu' },
  { key: 'remark', label: 'memberManage.ops.remark.menu' },
  { key: 'restriction', label: 'memberManage.ops.restriction.menu' },
];

export function isOperationCommand(value: unknown): value is OperationCommand {
  return OPERATION_COMMANDS.some((item) => item.key === value);
}
