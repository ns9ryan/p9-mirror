/**
 * 包网服务 - 通知推送页，独立词条命名空间，前缀 pushNotification.*
 */
export default {
  'menu.packageNetService': '包网服务',
  'menu.packageNetService.pushNotification': '通知推送',

  'pushNotification.search.keywordField.title': '标题',
  'pushNotification.search.keywordField.content': '内容',
  'pushNotification.search.keywordPlaceholder': '请输入关键字',
  'pushNotification.search.type': '通知类型：',
  'pushNotification.search.status': '状态：',

  'pushNotification.toolbar.create': '新增通知',

  'pushNotification.type.official': '官方通知',
  'pushNotification.status.pending': '待发送',
  'pushNotification.status.sent': '已发送',
  'pushNotification.status.failed': '发送失败',
  'pushNotification.validity.permanent': '永久',

  'pushNotification.column.id': 'ID',
  'pushNotification.column.title': '标题',
  'pushNotification.column.content': '内容',
  'pushNotification.column.type': '通知类型',
  'pushNotification.column.targets': '发送对象',
  'pushNotification.column.createdAt': '创建时间',
  'pushNotification.column.updatedAt': '最后操作时间',
  'pushNotification.column.updatedBy': '操作者',
  'pushNotification.column.status': '状态',
  'pushNotification.column.targetsCount': '共 {count} 个分站',

  'pushNotification.action.detail': '详情',

  'pushNotification.detail.title': '通知详情',
  'pushNotification.detail.validity': '有效期',

  'pushNotification.create.title': '新增通知',
  'pushNotification.create.section.base': '基础信息',
  'pushNotification.create.section.targets': '发送对象',
  'pushNotification.create.section.content': '通知内容',
  'pushNotification.create.field.type': '通知类型',
  'pushNotification.create.field.validity': '有效期',
  'pushNotification.create.field.title': '标题',
  'pushNotification.create.field.language': '选择语言',
  'pushNotification.create.field.titlePlaceholder':
    '请输入标题，最多 200 个字符',
  'pushNotification.create.field.content': '内容',
  'pushNotification.create.field.contentPlaceholder':
    '请输入通知内容，最多 5000 个字符',
  'pushNotification.create.action.addLanguage': '添加语言',
  'pushNotification.create.action.submit': '确认发送',
  'pushNotification.create.rule.typeRequired': '请选择通知类型',
  'pushNotification.create.rule.targetsRequired': '请至少选择一个分站',
  'pushNotification.create.rule.entriesRequired': '请至少添加一种语言',
  'pushNotification.create.rule.languageRequired': '第 {index} 组未选择语言',
  'pushNotification.create.rule.entryIncomplete':
    '「{language}」的标题和内容不能为空',
  'pushNotification.create.message.success': '通知已提交',

  'pushNotification.picker.searchPlaceholder': '搜索分站名称或编码',
  'pushNotification.picker.selected': '已选 {selected} / {total}',
  'pushNotification.picker.selectAll': '全选',
  'pushNotification.picker.clear': '清空',
  'pushNotification.picker.allNodes': '全部节点',
  'pushNotification.picker.unassigned': '未分配节点',
  'pushNotification.picker.online': '在线',
  'pushNotification.picker.offline': '离线',
  'pushNotification.picker.disabled': '已停用',
  'pushNotification.picker.selectNode': '全选本节点',
  'pushNotification.picker.emptyNode': '该节点下暂无分站',
  'pushNotification.picker.emptySearch': '没有匹配的分站',
  'pushNotification.picker.onlySelected': '只看已选',
  'pushNotification.picker.loadFailed':
    '分站或节点数据加载失败，请点击刷新重试',
};
