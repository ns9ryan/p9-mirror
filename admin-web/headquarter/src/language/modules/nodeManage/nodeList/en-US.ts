export default {
  'menu.nodeManage': 'Node Manage',
  'menu.nodeManage.nodeList': 'Node List',
  'menu.nodeManage.tasks': 'Node Tasks',

  'nodeList.title': 'Node List',
  'nodeList.subtitle':
    'Manage basic information, online status, running status and ops configuration of deployment nodes.',

  'nodeList.search.keyword': 'Keyword:',
  'nodeList.search.keywordPlaceholder': 'Enter node code or node name',
  'nodeList.search.status': 'Node Status:',
  'nodeList.search.onlineStatus': 'Online Status:',

  'nodeList.toolbar.create': 'Create Node',

  'nodeList.status.enabled': 'Enabled',
  'nodeList.status.disabled': 'Disabled',
  'nodeList.onlineStatus.online': 'Online',
  'nodeList.onlineStatus.offline': 'Offline',

  'nodeList.column.id': 'Node ID',
  'nodeList.column.code': 'Node Code',
  'nodeList.column.name': 'Node Name',
  'nodeList.column.status': 'Node Status',
  'nodeList.column.onlineStatus': 'Online Status',
  'nodeList.column.lastActiveAt': 'Last Active At',
  'nodeList.column.assignedOperatorCount': 'Assigned Stations',
  'nodeList.column.remark': 'Remark',
  'nodeList.column.operations': 'Operations',
  'nodeList.column.unit': '',

  'nodeList.action.detail': 'Detail',
  'nodeList.action.edit': 'Edit',
  'nodeList.action.resetSecret': 'Reset Secret',

  'nodeList.detail.title': 'Node Detail',
  'nodeList.detail.tab.base': 'Basic Info',
  'nodeList.detail.tab.tasks': 'Dispatch Tasks',
  'nodeList.detail.task.keyword': 'Keyword:',
  'nodeList.detail.task.keywordPlaceholder':
    'Enter a task number or request number',
  'nodeList.detail.task.status': 'Task Status:',
  'nodeList.detail.task.type': 'Task Type:',
  'nodeList.detail.task.column.taskNo': 'Task No.',
  'nodeList.detail.task.column.requestNo': 'Request No.',
  'nodeList.detail.task.column.type': 'Task Type',
  'nodeList.detail.task.column.status': 'Task Status',
  'nodeList.detail.task.column.startedAt': 'Started At',
  'nodeList.detail.task.column.finishedAt': 'Finished At',
  'nodeList.detail.task.status.pending': 'Pending',
  'nodeList.detail.task.status.running': 'Running',
  'nodeList.detail.task.status.success': 'Success',
  'nodeList.detail.task.status.failed': 'Failed',
  'nodeList.detail.task.type.createOperator': 'CREATE_OPERATOR',
  'nodeList.detail.task.type.updateOperator': 'UPDATE_OPERATOR',
  'nodeList.detail.task.type.deleteOperator': 'DELETE_OPERATOR',

  'nodeList.create.title': 'Create Node',
  'nodeList.create.step.base': 'Basic Info',
  'nodeList.create.step.confirm': 'Confirm',
  'nodeList.create.field.name': 'Node Name',
  'nodeList.create.field.namePlaceholder':
    'Enter a node name, up to 100 characters',
  'nodeList.create.field.status': 'Node Status',
  'nodeList.create.field.remark': 'Ops Remark',
  'nodeList.create.field.remarkPlaceholder': 'Enter an ops remark...',
  'nodeList.create.action.cancel': 'Cancel',
  'nodeList.create.action.next': 'Next',
  'nodeList.create.action.prev': 'Previous',
  'nodeList.create.action.submit': 'Confirm',
  'nodeList.create.action.close': 'Close',
  'nodeList.create.action.copyCode': 'Copy Code',
  'nodeList.create.action.copySecret': 'Copy Secret',
  'nodeList.create.rule.nameRequired': 'Please enter a node name',
  'nodeList.create.rule.statusRequired': 'Please select a node status',
  'nodeList.create.rule.remarkRequired': 'Please enter an ops remark',
  'nodeList.create.result.title': 'Node Created',
  'nodeList.create.result.secretLabel': 'Auth Secret',
  'nodeList.create.result.warning':
    'Copy and save the node authentication secret now. It cannot be viewed again after this dialog is closed. If it is lost, it can only be reset.',
  'nodeList.create.message.copied': 'Authentication secret copied',
  'nodeList.create.message.codeCopied': 'Node code copied',
  'nodeList.create.message.copyFailed':
    'Copy failed. Please select the text and copy it manually',
  'nodeList.create.message.secretMissing':
    'The node was created, but no authentication secret was returned. Please contact an administrator',
  'nodeList.create.message.closeSecretTitle': 'Secret not copied',
  'nodeList.create.message.closeSecretContent':
    'This secret cannot be viewed again after closing. If it is lost, it can only be reset. Close anyway?',

  'nodeList.edit.title': 'Edit Node',
  'nodeList.edit.message.noChange': 'No changes to save',

  'nodeList.reset.title': 'Secret Reset',
  'nodeList.reset.okText': 'Reset',
  'nodeList.reset.warning':
    'Copy and save the new node authentication secret now. The previous secret is already invalid and the current connection has been disconnected. It cannot be viewed again after this dialog is closed. The node must use the new secret before it can reconnect.',
  'nodeList.reset.message.secretMissing':
    'The secret was reset, but no new authentication secret was returned. Please contact an administrator',
  'nodeList.reset.message.closeTitle': 'New secret not copied',
  'nodeList.reset.message.closeContent':
    'The new secret cannot be viewed again after closing, and the previous secret is already invalid. Close anyway?',

  'nodeList.message.resetSecretConfirmTitle': 'Confirm Reset Secret',
  'nodeList.message.resetSecretConfirmContent':
    'After resetting, the previous authentication secret of node "{name}" becomes invalid immediately and the current connection is disconnected. The node must use the new secret before it can reconnect. Continue?',
};
