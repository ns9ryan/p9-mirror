/**
 * 节点管理 - 节点列表页，独立词条命名空间，前缀 nodeList.*
 */
export default {
  'menu.nodeManage': '节点管理',
  'menu.nodeManage.nodeList': '节点列表',
  'menu.nodeManage.tasks': '节点任务',

  'nodeList.title': '节点列表',
  'nodeList.subtitle': '管理部署节点的基础信息、在线状态、运行状态及运维配置。',

  'nodeList.search.keyword': '关键词：',
  'nodeList.search.keywordPlaceholder': '请输入节点编码、节点名称',
  'nodeList.search.status': '节点状态：',
  'nodeList.search.onlineStatus': '在线状态：',

  'nodeList.toolbar.create': '新增节点',

  'nodeList.status.enabled': '启用',
  'nodeList.status.disabled': '停用',
  'nodeList.onlineStatus.online': '在线',
  'nodeList.onlineStatus.offline': '离线',

  'nodeList.column.id': '节点ID',
  'nodeList.column.code': '节点编码',
  'nodeList.column.name': '节点名称',
  'nodeList.column.status': '节点状态',
  'nodeList.column.onlineStatus': '在线状态',
  'nodeList.column.lastActiveAt': '最近活动时间',
  'nodeList.column.assignedOperatorCount': '已分配分站',
  'nodeList.column.remark': '备注',
  'nodeList.column.operations': '操作',
  'nodeList.column.unit': '个',

  'nodeList.action.detail': '详情',
  'nodeList.action.edit': '编辑',
  'nodeList.action.resetSecret': '重置密钥',

  'nodeList.detail.title': '节点详情',
  'nodeList.detail.tab.base': '基本信息',
  'nodeList.detail.tab.tasks': '调度任务',
  'nodeList.detail.task.keyword': '关键词：',
  'nodeList.detail.task.keywordPlaceholder': '请输入任务编号或请求编号',
  'nodeList.detail.task.status': '任务状态：',
  'nodeList.detail.task.type': '任务类型：',
  'nodeList.detail.task.column.taskNo': '任务编号',
  'nodeList.detail.task.column.requestNo': '请求编号',
  'nodeList.detail.task.column.type': '任务类型',
  'nodeList.detail.task.column.status': '任务状态',
  'nodeList.detail.task.column.startedAt': '开始时间',
  'nodeList.detail.task.column.finishedAt': '结束时间',
  'nodeList.detail.task.status.pending': '待执行',
  'nodeList.detail.task.status.running': '执行中',
  'nodeList.detail.task.status.success': '成功',
  'nodeList.detail.task.status.failed': '失败',
  'nodeList.detail.task.type.createOperator': 'CREATE_OPERATOR',
  'nodeList.detail.task.type.updateOperator': 'UPDATE_OPERATOR',
  'nodeList.detail.task.type.deleteOperator': 'DELETE_OPERATOR',

  'nodeList.create.title': '新增节点',
  'nodeList.create.step.base': '基础信息',
  'nodeList.create.step.confirm': '确认信息',
  'nodeList.create.field.name': '节点名称',
  'nodeList.create.field.namePlaceholder': '请输入节点名称，最多100个字符',
  'nodeList.create.field.status': '节点状态',
  'nodeList.create.field.remark': '运维备注',
  'nodeList.create.field.remarkPlaceholder': '请输入运维备注信息...',
  'nodeList.create.action.cancel': '取消',
  'nodeList.create.action.next': '下一步',
  'nodeList.create.action.prev': '上一步',
  'nodeList.create.action.submit': '确认创建',
  'nodeList.create.action.close': '关闭',
  'nodeList.create.action.copyCode': '复制编码',
  'nodeList.create.action.copySecret': '复制密钥',
  'nodeList.create.rule.nameRequired': '请输入节点名称',
  'nodeList.create.rule.statusRequired': '请选择节点状态',
  'nodeList.create.rule.remarkRequired': '请输入运维备注',
  'nodeList.create.result.title': '节点创建成功',
  'nodeList.create.result.secretLabel': '节点认证密钥',
  'nodeList.create.result.warning':
    '请立即复制并保存节点认证密钥。关闭后无法再次查看，如遗失只能重置。',
  'nodeList.create.message.copied': '节点认证密钥已复制',
  'nodeList.create.message.codeCopied': '节点编码已复制',
  'nodeList.create.message.copyFailed': '复制失败，请手动选择后复制',
  'nodeList.create.message.secretMissing':
    '节点已创建，但未返回认证密钥，请联系管理员处理',
  'nodeList.create.message.closeSecretTitle': '尚未复制认证密钥',
  'nodeList.create.message.closeSecretContent':
    '关闭后无法再次查看该密钥，如遗失只能重置。确定关闭吗？',

  'nodeList.edit.title': '编辑节点',
  'nodeList.edit.message.noChange': '没有需要保存的修改',

  'nodeList.reset.title': '密钥已重置',
  'nodeList.reset.okText': '确认重置',
  'nodeList.reset.warning':
    '请立即复制并保存新的节点认证密钥。原密钥已立即失效，当前连接已断开，关闭后无法再次查看。节点端必须更新为新密钥后才能重新连接。',
  'nodeList.reset.message.secretMissing':
    '密钥已重置，但未返回新的认证密钥，请联系管理员处理',
  'nodeList.reset.message.closeTitle': '尚未复制新密钥',
  'nodeList.reset.message.closeContent':
    '关闭后无法再次查看新密钥，原密钥已失效。确定关闭吗？',

  'nodeList.message.resetSecretConfirmTitle': '确认重置密钥',
  'nodeList.message.resetSecretConfirmContent':
    '重置后，节点「{name}」的原认证密钥将立即失效，当前连接会被断开。节点端必须换成新密钥后才能重新连接。是否确认重置？',
};
