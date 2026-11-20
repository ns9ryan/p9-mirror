/**
 * 分站管理（operator 模块）独立词条命名空间，前缀 operatorStation.*，
 * 不与 outstation/stationManage（另一套仍是本地 mock 的分站页面）共用 key，
 * 避免两个页面字段含义不同却共享同一条翻译。
 */
export default {
  'operatorStation.title': '分站列表',
  'operatorStation.subtitle':
    '管理各分站的基础信息、创建进度、发布状态及生命周期状态。',

  'operatorStation.search.keyword': '关键词：',
  'operatorStation.search.keywordPlaceholder':
    '请输入分站名称或编码，最多100字符',
  'operatorStation.search.creationStatus': '创建状态：',
  'operatorStation.search.publishStatus': '发布状态：',
  'operatorStation.search.status': '生命周期状态：',

  'operatorStation.toolbar.create': '创建分站',

  'operatorStation.tip.title': '提示：',
  'operatorStation.tip.content':
    '完成分站创建向导并发布后，分站才会正式对外提供服务。',

  'operatorStation.creationStatus.draft': '草稿',
  'operatorStation.creationStatus.completed': '已完成',

  'operatorStation.publishStatus.unpublished': '未发布',
  'operatorStation.publishStatus.publishing': '发布中',
  'operatorStation.publishStatus.published': '已发布',
  'operatorStation.publishStatus.failed': '发布失败',

  'operatorStation.status.normal': '正常',
  'operatorStation.status.paused': '暂停',
  'operatorStation.status.closed': '关闭',

  'operatorStation.column.id': '分站ID',
  'operatorStation.column.code': '分站编码',
  'operatorStation.column.name': '分站名称',
  'operatorStation.column.timezone': '时区',
  'operatorStation.column.currency': '结算币种',
  'operatorStation.column.creationStatus': '创建状态',
  'operatorStation.column.publishStatus': '发布状态',
  'operatorStation.column.status': '生命周期状态',
  'operatorStation.column.publishedAt': '最近发布时间',
  'operatorStation.column.remark': '总网备注',

  'operatorStation.action.detail': '详情',
  'operatorStation.action.edit': '编辑',
  'operatorStation.action.complete': '完成创建',
  'operatorStation.action.publish': '发布',
  'operatorStation.action.delete': '删除',

  'operatorStation.message.completeConfirmTitle': '确认完成创建',
  'operatorStation.message.completeConfirmContent':
    '确认分站「{name}」的档案、域名与基础资源已配置完成？完成后创建状态将变为「已完成」。',
  'operatorStation.message.completeSuccess': '已完成创建',
  'operatorStation.message.publishConfirmTitle': '确认发布',
  'operatorStation.message.publishConfirmContent':
    '发布后该分站将正式对外提供服务，是否确认发布？',
  'operatorStation.message.publishTodoTip':
    '（后端发布流程仍在开发中，提交后请稍后刷新确认发布状态）',
  'operatorStation.message.publishSubmitted': '发布请求已提交',
  'operatorStation.message.deleteNotAllowed':
    '仅「未发布」或「发布失败」的分站可以删除',
  'operatorStation.message.deleteConfirmContent':
    '删除后不可恢复，是否确认删除分站「{name}」？',

  'operatorStation.edit.title': '编辑分站',
  'operatorStation.edit.lockTip':
    '该分站正在发布或已发布过，时区与结算币种不可再修改。',
  'operatorStation.edit.noChange': '未修改任何内容',

  'operatorStation.detail.title': '分站详情',
  'operatorStation.detail.section.baseInfo': '基础信息',
  'operatorStation.detail.section.profile': '档案信息',
  'operatorStation.detail.section.admin': '管理员账号',
  'operatorStation.detail.section.domain': '域名配置',
  'operatorStation.detail.section.resource': '基础资源摘要',
  'operatorStation.detail.section.gameResource': '游戏资源摘要',
  'operatorStation.detail.section.publish': '发布信息',
  'operatorStation.detail.profileEmpty': '暂无档案信息',
  'operatorStation.detail.resourceAssigned': '已分配 {count} 项',
  'operatorStation.detail.gameResource.category': '游戏分类',
  'operatorStation.detail.gameResource.provider': '游戏厂商',
  'operatorStation.detail.gameResource.channel': '游戏渠道',
  'operatorStation.detail.gameResource.count': '游戏数量',
  'operatorStation.detail.gameResource.unit': '款',
  'operatorStation.detail.gameResource.mockTip':
    '（示例数据，待接入分站游戏资源接口后自动更新）',
  'operatorStation.detail.publishResult': '最近发布结果',
  'operatorStation.detail.publishNote': '当前说明',

  'operatorStation.create.title': '创建分站',
  'operatorStation.create.stepLabel': '第{step}步：{name}',
  'operatorStation.create.step.baseInfo': '基础信息',
  'operatorStation.create.step.profileInfo': '档案信息',
  'operatorStation.create.step.adminAccount': '管理员账号',
  'operatorStation.create.step.domain': '域名配置',
  'operatorStation.create.step.resource': '基础资源',

  'operatorStation.create.action.cancel': '取消',
  'operatorStation.create.action.saveDraft': '保存草稿',
  'operatorStation.create.action.save': '保存',
  'operatorStation.create.action.prev': '上一步',
  'operatorStation.create.action.next': '下一步',
  'operatorStation.create.action.finish': '完成创建',

  'operatorStation.create.cancelConfirm.title': '确认退出创建',
  'operatorStation.create.cancelConfirm.content':
    '分站「{name}」已保存为草稿，退出后可在列表中随时继续完善。',

  'operatorStation.create.baseInfo.title': '基础信息',
  'operatorStation.create.baseInfo.desc':
    '请填写分站的基础信息，用于标识和管理分站。',
  'operatorStation.create.baseInfo.name': '分站名称',
  'operatorStation.create.baseInfo.namePlaceholder':
    '请输入分站名称，如：泰国站',
  'operatorStation.create.baseInfo.code': '分站编码',
  'operatorStation.create.baseInfo.codePlaceholder':
    '请输入分站编码，如：P9-TH001',
  'operatorStation.create.baseInfo.codeTip':
    '此编码仅供本地参考，实际业务编码由系统在创建时自动生成',
  'operatorStation.create.baseInfo.timezone': '时区',
  'operatorStation.create.baseInfo.timezonePlaceholder': '请选择时区',
  'operatorStation.create.baseInfo.currency': '结算币种',
  'operatorStation.create.baseInfo.currencyPlaceholder': '请选择结算币种',
  'operatorStation.create.baseInfo.status': '生命周期状态',
  'operatorStation.create.baseInfo.statusPlaceholder': '请选择状态',
  'operatorStation.create.baseInfo.remark': '备注',
  'operatorStation.create.baseInfo.remarkPlaceholder':
    '请输入备注信息（选填），如分站定位、运营说明等...',
  'operatorStation.create.baseInfo.tipTitle': '填写说明',
  'operatorStation.create.baseInfo.tipContent':
    '本步骤用于记录运营方的核心属性信息，包括分站名称、时区、结算币种及运营状态等。请确保信息准确，便于后续的域名配置和资源分配。',

  'operatorStation.create.profileInfo.title': '档案信息',
  'operatorStation.create.profileInfo.tip':
    '此信息仅用于总网内部档案管理和沟通协调使用，不会对外展示。',
  'operatorStation.create.profileInfo.companyName': '公司名称',
  'operatorStation.create.profileInfo.companyNamePlaceholder': '请输入公司名称',
  'operatorStation.create.profileInfo.contactName': '主要联系人名称',
  'operatorStation.create.profileInfo.contactNamePlaceholder':
    '请输入主要联系人名称',
  'operatorStation.create.profileInfo.contactEmail': '主要联系人邮箱',
  'operatorStation.create.profileInfo.contactEmailPlaceholder':
    '请输入主要联系人邮箱',
  'operatorStation.create.profileInfo.remark': '总网内部档案备注',
  'operatorStation.create.profileInfo.remarkPlaceholder':
    '请输入备注信息（可选）',
  'operatorStation.create.profileInfo.statusBar': '当前分站：{code} / {name}',
  'operatorStation.create.profileInfo.statusDraft': '草稿状态',

  'operatorStation.create.adminAccount.title': '创建初始管理员账号',
  'operatorStation.create.adminAccount.desc':
    '为该分站创建一个初始的操作管理员账号，用于登录分站后台。',
  'operatorStation.create.adminAccount.account': '登录账号',
  'operatorStation.create.adminAccount.accountPlaceholder':
    '请输入登录账号，支持字母、数字、下划线',
  'operatorStation.create.adminAccount.password': '登录密码',
  'operatorStation.create.adminAccount.passwordPlaceholder':
    '请输入登录密码，6-32位',
  'operatorStation.create.adminAccount.passwordOptionalPlaceholder':
    '不修改密码请留空',
  'operatorStation.create.adminAccount.confirmPassword': '确认密码',
  'operatorStation.create.adminAccount.confirmPasswordPlaceholder':
    '请再次输入登录密码',
  'operatorStation.create.adminAccount.confirmPasswordOptionalPlaceholder':
    '修改密码时请再次输入',
  'operatorStation.create.adminAccount.displayName': '显示名称',
  'operatorStation.create.adminAccount.displayNamePlaceholder':
    '请输入显示名称，如：运营管理员',
  'operatorStation.create.adminAccount.status': '状态',
  'operatorStation.create.adminAccount.tip':
    '创建的初始管理员账号，后续可以在「管理员账号」中进行维护，包括修改密码、禁用/启用、角色权限等。',
  'operatorStation.create.adminAccount.summaryTitle': '已填写信息概览',
  'operatorStation.create.domain.title': '域名配置',
  'operatorStation.create.domain.tipTitle': '提示',
  'operatorStation.create.domain.tip':
    '同一域名类型可以添加多条域名记录，但同一时间只能有一条记录处于启用状态。',
  'operatorStation.create.domain.addButton': '新增域名',
  'operatorStation.create.domain.column.type': '域名类型',
  'operatorStation.create.domain.column.domain': '域名',
  'operatorStation.create.domain.column.status': '状态',
  'operatorStation.create.domain.column.remark': '备注',
  'operatorStation.create.domain.column.operations': '操作',
  'operatorStation.create.domain.type.admin': '分站后台',
  'operatorStation.create.domain.type.agent': '代理后台',
  'operatorStation.create.domain.type.h5': '会员H5',
  'operatorStation.create.domain.modal.addTitle': '新增域名',
  'operatorStation.create.domain.modal.editTitle': '编辑域名',
  'operatorStation.create.domain.field.type': '域名类型',
  'operatorStation.create.domain.field.domain': '域名',
  'operatorStation.create.domain.field.domainPlaceholder':
    '请输入域名，如：admin.th.example.com',
  'operatorStation.create.domain.field.status': '状态',
  'operatorStation.create.domain.field.remark': '备注',
  'operatorStation.create.domain.field.remarkPlaceholder': '请输入备注（选填）',
  'operatorStation.create.domain.rule.typeRequired': '请选择域名类型',
  'operatorStation.create.domain.rule.domainRequired': '请输入域名',
  'operatorStation.create.domain.delete.content': '确认删除域名「{domain}」？',
  'operatorStation.create.domain.empty': '暂无域名，点击「新增域名」添加',
  'operatorStation.create.domain.message.editLocalOnly':
    '域名已创建到后端，当前暂无修改接口，本次仅更新向导内显示',
  'operatorStation.create.domain.message.deleteLocalOnly':
    '域名已创建到后端，当前暂无删除接口，本次仅从向导列表移除',

  'operatorStation.create.resource.title': '基础资源',
  'operatorStation.create.resource.tip': '按向导完成基础资源配置，创建分站。',
  'operatorStation.create.resource.tab.language': '语言',
  'operatorStation.create.resource.tab.region': '经营地区',
  'operatorStation.create.resource.tab.proxyLine': '代理子线路',
  'operatorStation.create.resource.searchPlaceholder': '请输入名称或编码搜索',
  'operatorStation.create.resource.filter.all': '全部',
  'operatorStation.create.resource.filter.assigned': '已分配',
  'operatorStation.create.resource.filter.unassigned': '未分配',
  'operatorStation.create.resource.column.name': '名称',
  'operatorStation.create.resource.column.code': '编码',
  'operatorStation.create.resource.column.status': '状态',
  'operatorStation.create.resource.column.assigned': '已分配',
  'operatorStation.create.resource.column.assignedAt': '分配时间',
  'operatorStation.create.resource.summaryTitle': '基础资源已选择',
  'operatorStation.create.resource.summary.language': '语言',
  'operatorStation.create.resource.summary.region': '经营地区',
  'operatorStation.create.resource.summary.proxyLine': '代理子线路',
  'operatorStation.create.resource.unit': '项',

  'operatorStation.create.rule.nameRequired': '请输入分站名称',
  'operatorStation.create.rule.timezoneRequired': '请选择时区',
  'operatorStation.create.rule.currencyRequired': '请选择结算币种',
  'operatorStation.create.rule.statusRequired': '请选择状态',
  'operatorStation.create.rule.companyNameRequired': '请输入公司名称',
  'operatorStation.create.rule.contactNameRequired': '请输入主要联系人名称',
  'operatorStation.create.rule.contactEmailRequired': '请输入主要联系人邮箱',
  'operatorStation.create.rule.contactEmailInvalid': '请输入有效的邮箱地址',
  'operatorStation.create.rule.accountRequired': '请输入登录账号',
  'operatorStation.create.rule.passwordRequired': '请输入登录密码',
  'operatorStation.create.rule.passwordLength': '密码长度需为 6-32 位',
  'operatorStation.create.rule.confirmPasswordRequired': '请再次输入密码',
  'operatorStation.create.rule.passwordNotMatch': '两次输入的密码不一致',
  'operatorStation.create.rule.displayNameRequired': '请输入显示名称',

  'operatorStation.create.message.saveDraftSuccess': '已保存草稿',
  'operatorStation.create.message.createSuccess': '分站「{name}」创建成功',
  'operatorStation.create.message.completeFailed':
    '分站已保存为草稿，但完成创建失败，请在列表中重试「完成创建」',
  'operatorStation.create.message.pendingApiTip':
    '管理员账号当前暂无对应后端接口，仅本地记录展示，等待后端接口开放后再接入。',
  'operatorStation.create.message.needBaseInfoFirst':
    '请先完成基础信息并创建分站草稿',

  /** 分站域名管理页，独立词条命名空间，前缀 operatorDomain.* */
  'operatorDomain.title': '域名管理',
  'operatorDomain.subtitle': '管理各分站的分站后台、代理后台及会员H5域名配置。',

  'operatorDomain.search.operator': '所属分站：',
  'operatorDomain.search.operatorPlaceholder': '请选择分站',
  'operatorDomain.search.keyword': '域名关键字：',
  'operatorDomain.search.keywordPlaceholder': '请输入域名，最多253个字符',
  'operatorDomain.search.domainType': '域名类型：',
  'operatorDomain.search.status': '状态：',

  'operatorDomain.toolbar.create': '新增域名',

  'operatorDomain.type.admin': '分站后台',
  'operatorDomain.type.agent': '代理后台',
  'operatorDomain.type.h5': '会员H5',

  'operatorDomain.column.id': '域名ID',
  'operatorDomain.column.operator': '所属分站',
  'operatorDomain.column.domainName': '域名',
  'operatorDomain.column.domainType': '域名类型',
  'operatorDomain.column.status': '状态',
  'operatorDomain.column.remark': '备注',
  'operatorDomain.column.updatedAt': '更新时间',

  'operatorDomain.action.detail': '详情',
  'operatorDomain.action.edit': '编辑',
  'operatorDomain.action.enable': '启用',
  'operatorDomain.action.disable': '停用',
  'operatorDomain.action.delete': '删除',

  'operatorDomain.modal.createTitle': '新增域名',
  'operatorDomain.modal.editTitle': '编辑域名',
  'operatorDomain.modal.detailTitle': '域名详情',
  'operatorDomain.field.operator': '所属分站',
  'operatorDomain.field.operatorPlaceholder': '请选择分站',
  'operatorDomain.field.domainType': '域名类型',
  'operatorDomain.field.domainName': '域名',
  'operatorDomain.field.domainNamePlaceholder':
    '请输入域名，如：admin.th.example.com',
  'operatorDomain.field.status': '状态',
  'operatorDomain.field.remark': '备注',
  'operatorDomain.field.remarkPlaceholder':
    '请输入备注信息（选填），最多1000个字符',

  'operatorDomain.rule.operatorRequired': '请选择所属分站',
  'operatorDomain.rule.domainTypeRequired': '请选择域名类型',
  'operatorDomain.rule.domainNameRequired': '请输入域名',

  'operatorDomain.message.createSuccess': '域名创建成功',
  'operatorDomain.message.updateSuccess': '域名修改成功',
  'operatorDomain.message.noChange': '未修改任何内容',
  'operatorDomain.message.statusUpdateSuccess': '状态修改成功',
  'operatorDomain.message.deleteConfirmContent':
    '删除后不可恢复，是否确认删除域名「{domain}」？',

  /** 分站管理员账号管理页，独立词条命名空间，前缀 operatorAdmin.* */
  'operatorAdmin.title': '管理员账号',
  'operatorAdmin.subtitle': '管理各分站后台的管理员账号、状态与基本信息。',

  'operatorAdmin.search.operator': '所属分站：',
  'operatorAdmin.search.operatorPlaceholder': '请选择分站',
  'operatorAdmin.search.keyword': '关键词：',
  'operatorAdmin.search.keywordPlaceholder':
    '请输入账号或显示名称，最多100字符',
  'operatorAdmin.search.status': '状态：',

  'operatorAdmin.toolbar.create': '新增管理员',

  'operatorAdmin.column.id': '管理员ID',
  'operatorAdmin.column.operator': '所属分站',
  'operatorAdmin.column.username': '登录账号',
  'operatorAdmin.column.displayName': '显示名称',
  'operatorAdmin.column.status': '状态',
  'operatorAdmin.column.updatedAt': '更新时间',

  'operatorAdmin.action.edit': '编辑',
  'operatorAdmin.action.resetPassword': '重置密码',
  'operatorAdmin.action.enable': '启用',
  'operatorAdmin.action.disable': '停用',

  'operatorAdmin.modal.createTitle': '新增管理员',
  'operatorAdmin.modal.editTitle': '编辑管理员',
  'operatorAdmin.modal.resetPasswordTitle': '重置密码',
  'operatorAdmin.modal.resetPasswordTip':
    '为管理员「{username}」设置新密码，重置后立即生效。',
  'operatorAdmin.field.operator': '所属分站',
  'operatorAdmin.field.operatorPlaceholder': '请选择分站',
  'operatorAdmin.field.username': '登录账号',
  'operatorAdmin.field.usernamePlaceholder':
    '请输入登录账号，支持字母、数字、下划线',
  'operatorAdmin.field.password': '登录密码',
  'operatorAdmin.field.passwordPlaceholder': '请输入登录密码，6-32位',
  'operatorAdmin.field.passwordOptionalPlaceholder': '不修改密码请留空',
  'operatorAdmin.field.confirmPassword': '确认密码',
  'operatorAdmin.field.confirmPasswordPlaceholder': '请再次输入登录密码',
  'operatorAdmin.field.confirmPasswordOptionalPlaceholder':
    '修改密码时请再次输入',
  'operatorAdmin.field.displayName': '显示名称',
  'operatorAdmin.field.displayNamePlaceholder':
    '请输入显示名称，如：运营管理员',
  'operatorAdmin.field.status': '状态',

  'operatorAdmin.rule.operatorRequired': '请选择所属分站',
  'operatorAdmin.rule.usernameRequired': '请输入登录账号',
  'operatorAdmin.rule.passwordRequired': '请输入登录密码',
  'operatorAdmin.rule.passwordLength': '密码长度需为 6-32 位',
  'operatorAdmin.rule.confirmPasswordRequired': '请再次输入密码',
  'operatorAdmin.rule.passwordNotMatch': '两次输入的密码不一致',
  'operatorAdmin.rule.displayNameRequired': '请输入显示名称',

  'operatorAdmin.message.createSuccess': '管理员创建成功',
  'operatorAdmin.message.updateSuccess': '管理员修改成功',
  'operatorAdmin.message.statusUpdateSuccess': '状态修改成功',
  'operatorAdmin.message.resetPasswordSuccess': '密码重置成功',

  /** 分站基础资源分配汇总页，独立词条命名空间，前缀 operatorBasicResource.* */
  'operatorBasicResource.title': '基础资源分配',
  'operatorBasicResource.subtitle':
    '查看各分站已分配的语言、经营地区与代理子线路数量。',

  'operatorBasicResource.search.keyword': '关键词：',
  'operatorBasicResource.search.keywordPlaceholder':
    '请输入分站名称或编码，最多100字符',

  'operatorBasicResource.column.operatorId': '分站ID',
  'operatorBasicResource.column.operatorCode': '分站编码',
  'operatorBasicResource.column.operatorName': '分站名称',
  'operatorBasicResource.column.languageCount': '已分配语言',
  'operatorBasicResource.column.regionCount': '已分配经营地区',
  'operatorBasicResource.column.agentLineCount': '已分配代理子线路',
  'operatorBasicResource.column.unit': '项',
  'operatorBasicResource.column.operations': '操作',
  'operatorBasicResource.action.configure': '配置',

  'operatorBasicResource.modal.title': '资源分配配置',
  'operatorBasicResource.modal.selectOperator': '选择分站',
  'operatorBasicResource.modal.selectOperatorPlaceholder': '请选择分站',
  'operatorBasicResource.modal.info.id': '分站ID',
  'operatorBasicResource.modal.info.code': '分站编码',
  'operatorBasicResource.modal.info.name': '分站名称',
  'operatorBasicResource.modal.info.timezone': '时区',
  'operatorBasicResource.modal.info.currency': '结算币种',
  'operatorBasicResource.modal.tab.language': '语言',
  'operatorBasicResource.modal.tab.region': '经营地区',
  'operatorBasicResource.modal.tab.proxyLine': '代理子线路',
  'operatorBasicResource.modal.searchPlaceholder': '请输入名称、编码搜索',
  'operatorBasicResource.modal.filter.all': '全部',
  'operatorBasicResource.modal.filter.enabled': '启用',
  'operatorBasicResource.modal.filter.disabled': '停用',
  'operatorBasicResource.modal.column.name': '名称',
  'operatorBasicResource.modal.column.code': '编码',
  'operatorBasicResource.modal.column.status': '状态',
  'operatorBasicResource.modal.column.assigned': '已分配',
  'operatorBasicResource.modal.column.assignedAt': '分配时间',
  'operatorBasicResource.modal.summary.title': '当前分配概览',
  'operatorBasicResource.modal.summary.language': '语言',
  'operatorBasicResource.modal.summary.region': '经营地区',
  'operatorBasicResource.modal.summary.proxyLine': '代理子线路',
  'operatorBasicResource.modal.summary.desc': '已分配 / 总数',
  'operatorBasicResource.modal.action.save': '保存配置',
  'operatorBasicResource.modal.action.resetSelection': '重置选择',
  'operatorBasicResource.modal.tip.title': '提示：',
  'operatorBasicResource.modal.tip.line1':
    '1. 请先在上方选择分站，再进行资源分配。',
  'operatorBasicResource.modal.tip.line2': '2. 配置保存后立即生效。',
  'operatorBasicResource.modal.message.saveSuccess': '配置保存成功',
  'operatorBasicResource.modal.message.selectOperatorFirst': '请先选择分站',
  'operatorBasicResource.modal.empty': '暂无数据',

  /** 分站游戏资源分配汇总页，独立词条命名空间，前缀 operatorGameResource.* */
  'operatorGameResource.title': '游戏资源分配',
  'operatorGameResource.subtitle':
    '查看各分站已分配的游戏、分类、厂商与渠道数量。',

  'operatorGameResource.search.opCode': '分站编码：',
  'operatorGameResource.search.opCodePlaceholder':
    '请输入分站业务编码，最多64字符',
  'operatorGameResource.search.status': '分站状态：',

  'operatorGameResource.column.id': 'ID',
  'operatorGameResource.column.opCode': '分站编码',
  'operatorGameResource.column.opName': '分站名称',
  'operatorGameResource.column.gameCount': '游戏数量',
  'operatorGameResource.column.gameCategoryCount': '游戏分类',
  'operatorGameResource.column.gameProviderCount': '游戏厂商',
  'operatorGameResource.column.gameChannelCount': '游戏渠道',
  'operatorGameResource.column.updatedAt': '更新时间',
  'operatorGameResource.column.unit': '项',
  'operatorGameResource.column.gameUnit': '款',
  'operatorGameResource.column.operations': '操作',
  'operatorGameResource.action.configure': '配置',

  'operatorGameResource.modal.title': '游戏资源配置',
  'operatorGameResource.modal.selectOperator': '当前操作分站',
  'operatorGameResource.modal.selectOperatorPlaceholder': '请选择分站',
  'operatorGameResource.modal.info.code': '分站编码',
  'operatorGameResource.modal.info.timezone': '所属时区',
  'operatorGameResource.modal.info.domain': '分站域名',
  'operatorGameResource.modal.info.status': '分站状态',
  'operatorGameResource.modal.tab.category': '游戏分类',
  'operatorGameResource.modal.tab.provider': '游戏厂商',
  'operatorGameResource.modal.tab.channel': '游戏渠道',
  'operatorGameResource.modal.tab.game': '游戏',
  'operatorGameResource.modal.search.keyword': '关键词',
  'operatorGameResource.modal.search.keywordPlaceholder.category':
    '请输入分类名称或编码',
  'operatorGameResource.modal.search.keywordPlaceholder.provider':
    '请输入厂商名称或编码',
  'operatorGameResource.modal.search.keywordPlaceholder.channel':
    '请输入渠道名称或编码',
  'operatorGameResource.modal.search.keywordPlaceholder.game':
    '请输入游戏名称或编码',
  'operatorGameResource.modal.search.status': '状态',
  'operatorGameResource.modal.search.allocated': '是否已分配',
  'operatorGameResource.modal.filter.all': '全部',
  'operatorGameResource.modal.filter.enabled': '启用',
  'operatorGameResource.modal.filter.disabled': '停用',
  'operatorGameResource.modal.filter.allocated': '已分配',
  'operatorGameResource.modal.filter.unallocated': '未分配',
  'operatorGameResource.modal.column.name': '资源名称',
  'operatorGameResource.modal.column.code': '编码',
  'operatorGameResource.modal.column.status': '状态',
  'operatorGameResource.modal.column.allocated': '已分配',
  'operatorGameResource.modal.column.allocatedAt': '分配时间',
  'operatorGameResource.modal.column.remark': '备注',
  'operatorGameResource.modal.column.operations': '操作',
  'operatorGameResource.modal.allocated.yes': '已分配',
  'operatorGameResource.modal.allocated.no': '未分配',
  'operatorGameResource.modal.action.detail': '详情',
  'operatorGameResource.modal.action.save': '保存配置',
  'operatorGameResource.modal.action.resetSelection': '重置选择',
  'operatorGameResource.modal.action.batchStatus': '批量修改状态',
  'operatorGameResource.modal.action.batchEnable': '批量启用',
  'operatorGameResource.modal.action.batchDisable': '批量停用',
  'operatorGameResource.modal.summary.title': '已选游戏资源概览',
  'operatorGameResource.modal.summary.category': '分类',
  'operatorGameResource.modal.summary.provider': '厂商',
  'operatorGameResource.modal.summary.channel': '渠道',
  'operatorGameResource.modal.summary.game': '游戏',
  'operatorGameResource.modal.summary.unit': '项',
  'operatorGameResource.modal.tip.title': '提示：',
  'operatorGameResource.modal.tip.content':
    '请根据分站的运营策略，选择合适的游戏资源进行配置。保存后将立即生效。',
  'operatorGameResource.modal.message.selectOperatorFirst': '请先选择分站',
  'operatorGameResource.modal.message.saveSuccess': '配置保存成功',
  'operatorGameResource.modal.message.selectItemsFirst': '请先勾选需要操作的资源',
  'operatorGameResource.modal.message.batchStatusSuccess': '状态修改成功',
  'operatorGameResource.modal.message.saveEmpty': '请至少选择一项资源后再保存',
  'operatorGameResource.modal.detail.title': '资源详情',
  'operatorGameResource.modal.empty': '暂无数据',
};
