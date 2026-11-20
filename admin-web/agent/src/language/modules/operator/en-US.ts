export default {
  'operatorStation.title': 'Station List',
  'operatorStation.subtitle':
    'Manage basic information, creation progress, publish status and lifecycle status of each station.',

  'operatorStation.search.keyword': 'Keyword:',
  'operatorStation.search.keywordPlaceholder':
    'Enter station name or code, up to 100 characters',
  'operatorStation.search.creationStatus': 'Creation Status:',
  'operatorStation.search.publishStatus': 'Publish Status:',
  'operatorStation.search.status': 'Status:',

  'operatorStation.toolbar.create': 'Create Station',

  'operatorStation.tip.title': 'Tip:',
  'operatorStation.tip.content':
    'A station only goes live once the creation wizard is completed and it has been published.',

  'operatorStation.creationStatus.draft': 'Draft',
  'operatorStation.creationStatus.completed': 'Completed',

  'operatorStation.publishStatus.unpublished': 'Unpublished',
  'operatorStation.publishStatus.publishing': 'Publishing',
  'operatorStation.publishStatus.published': 'Published',
  'operatorStation.publishStatus.failed': 'Publish Failed',

  'operatorStation.status.normal': 'Normal',
  'operatorStation.status.paused': 'Paused',
  'operatorStation.status.closed': 'Closed',

  'operatorStation.column.id': 'Station ID',
  'operatorStation.column.code': 'Station Code',
  'operatorStation.column.name': 'Station Name',
  'operatorStation.column.timezone': 'Timezone',
  'operatorStation.column.currency': 'Currency',
  'operatorStation.column.creationStatus': 'Creation Status',
  'operatorStation.column.publishStatus': 'Publish Status',
  'operatorStation.column.status': 'Status',
  'operatorStation.column.publishedAt': 'Last Published At',
  'operatorStation.column.remark': 'Internal Remark',

  'operatorStation.action.detail': 'Detail',
  'operatorStation.action.edit': 'Edit',
  'operatorStation.action.complete': 'Complete',
  'operatorStation.action.publish': 'Publish',
  'operatorStation.action.delete': 'Delete',

  'operatorStation.message.completeConfirmTitle': 'Confirm Completion',
  'operatorStation.message.completeConfirmContent':
    'Are the profile, domain and basic resources of station "{name}" fully configured? Its creation status will change to "Completed".',
  'operatorStation.message.completeSuccess': 'Creation completed',
  'operatorStation.message.publishConfirmTitle': 'Confirm Publish',
  'operatorStation.message.publishConfirmContent':
    'Once published, this station will be live to the public. Continue?',
  'operatorStation.message.publishTodoTip':
    ' (The backend publish flow is still under development. Refresh later to confirm the publish status.)',
  'operatorStation.message.publishSubmitted': 'Publish request submitted',
  'operatorStation.message.deleteNotAllowed':
    'Only unpublished or publish-failed stations can be deleted',
  'operatorStation.message.deleteConfirmContent':
    'This cannot be undone. Delete station "{name}"?',

  'operatorStation.edit.title': 'Edit Station',
  'operatorStation.edit.lockTip':
    'This station is publishing or has been published before, so the timezone and settlement currency can no longer be changed.',
  'operatorStation.edit.noChange': 'Nothing changed',

  'operatorStation.detail.title': 'Station Detail',
  'operatorStation.detail.section.baseInfo': 'Basic Info',
  'operatorStation.detail.section.profile': 'Profile',
  'operatorStation.detail.section.admin': 'Admin Accounts',
  'operatorStation.detail.section.domain': 'Domain Config',
  'operatorStation.detail.section.resource': 'Basic Resource Summary',
  'operatorStation.detail.section.gameResource': 'Game Resource Summary',
  'operatorStation.detail.section.publish': 'Publish Info',
  'operatorStation.detail.profileEmpty': 'No profile yet',
  'operatorStation.detail.resourceAssigned': '{count} assigned',
  'operatorStation.detail.gameResource.category': 'Game Categories',
  'operatorStation.detail.gameResource.provider': 'Game Providers',
  'operatorStation.detail.gameResource.channel': 'Game Channels',
  'operatorStation.detail.gameResource.count': 'Game Count',
  'operatorStation.detail.gameResource.unit': 'games',
  'operatorStation.detail.gameResource.mockTip':
    '(Sample data, will update once the station game resource API is available)',
  'operatorStation.detail.publishResult': 'Last Publish Result',
  'operatorStation.detail.publishNote': 'Current Note',

  'operatorStation.create.title': 'Create Station',
  'operatorStation.create.stepLabel': 'Step {step}: {name}',
  'operatorStation.create.step.baseInfo': 'Basic Info',
  'operatorStation.create.step.profileInfo': 'Profile Info',
  'operatorStation.create.step.adminAccount': 'Admin Account',
  'operatorStation.create.step.domain': 'Domain Setup',
  'operatorStation.create.step.resource': 'Basic Resources',

  'operatorStation.create.action.cancel': 'Cancel',
  'operatorStation.create.action.saveDraft': 'Save Draft',
  'operatorStation.create.action.save': 'Save',
  'operatorStation.create.action.prev': 'Previous',
  'operatorStation.create.action.next': 'Next',
  'operatorStation.create.action.finish': 'Finish',

  'operatorStation.create.cancelConfirm.title': 'Confirm Exit',
  'operatorStation.create.cancelConfirm.content':
    'Station "{name}" has been saved as a draft. You can continue setting it up from the list anytime.',

  'operatorStation.create.baseInfo.title': 'Basic Info',
  'operatorStation.create.baseInfo.desc':
    'Fill in the basic information used to identify and manage the station.',
  'operatorStation.create.baseInfo.name': 'Station Name',
  'operatorStation.create.baseInfo.namePlaceholder':
    'Enter station name, e.g. Thailand Station',
  'operatorStation.create.baseInfo.code': 'Station Code',
  'operatorStation.create.baseInfo.codePlaceholder':
    'Enter station code, e.g. P9-TH001',
  'operatorStation.create.baseInfo.codeTip':
    'For local reference only; the actual business code is generated by the system on creation',
  'operatorStation.create.baseInfo.timezone': 'Timezone',
  'operatorStation.create.baseInfo.timezonePlaceholder': 'Select timezone',
  'operatorStation.create.baseInfo.currency': 'Settlement Currency',
  'operatorStation.create.baseInfo.currencyPlaceholder':
    'Select settlement currency',
  'operatorStation.create.baseInfo.status': 'Status',
  'operatorStation.create.baseInfo.statusPlaceholder': 'Select status',
  'operatorStation.create.baseInfo.remark': 'Remark',
  'operatorStation.create.baseInfo.remarkPlaceholder':
    'Enter remark (optional), e.g. positioning, operation notes...',
  'operatorStation.create.baseInfo.tipTitle': 'Instructions',
  'operatorStation.create.baseInfo.tipContent':
    'This step records the core attributes of the operator, including station name, timezone, settlement currency and status. Make sure the information is accurate for the following domain setup and resource allocation.',

  'operatorStation.create.profileInfo.title': 'Profile Info',
  'operatorStation.create.profileInfo.tip':
    'This information is for internal headquarters filing and communication only and will not be shown publicly.',
  'operatorStation.create.profileInfo.companyName': 'Company Name',
  'operatorStation.create.profileInfo.companyNamePlaceholder':
    'Enter company name',
  'operatorStation.create.profileInfo.contactName': 'Primary Contact Name',
  'operatorStation.create.profileInfo.contactNamePlaceholder':
    'Enter primary contact name',
  'operatorStation.create.profileInfo.contactEmail': 'Primary Contact Email',
  'operatorStation.create.profileInfo.contactEmailPlaceholder':
    'Enter primary contact email',
  'operatorStation.create.profileInfo.remark': 'Internal Profile Remark',
  'operatorStation.create.profileInfo.remarkPlaceholder':
    'Enter remark (optional)',
  'operatorStation.create.profileInfo.statusBar':
    'Current Station: {code} / {name}',
  'operatorStation.create.profileInfo.statusDraft': 'Draft',

  'operatorStation.create.adminAccount.title': 'Create Initial Admin Account',
  'operatorStation.create.adminAccount.desc':
    'Create an initial operations admin account for this station to log in to the station backend.',
  'operatorStation.create.adminAccount.account': 'Login Account',
  'operatorStation.create.adminAccount.accountPlaceholder':
    'Enter login account (letters, digits, underscore)',
  'operatorStation.create.adminAccount.password': 'Login Password',
  'operatorStation.create.adminAccount.passwordPlaceholder':
    'Enter login password, 6-32 characters',
  'operatorStation.create.adminAccount.passwordOptionalPlaceholder':
    'Leave blank to keep current password',
  'operatorStation.create.adminAccount.confirmPassword': 'Confirm Password',
  'operatorStation.create.adminAccount.confirmPasswordPlaceholder':
    'Re-enter login password',
  'operatorStation.create.adminAccount.confirmPasswordOptionalPlaceholder':
    'Re-enter if changing password',
  'operatorStation.create.adminAccount.displayName': 'Display Name',
  'operatorStation.create.adminAccount.displayNamePlaceholder':
    'Enter display name, e.g. Operations Admin',
  'operatorStation.create.adminAccount.status': 'Status',
  'operatorStation.create.adminAccount.tip':
    'The initial admin account created here can later be managed under "Admin Accounts", including password reset, enable/disable, and role permissions.',
  'operatorStation.create.adminAccount.summaryTitle': 'Summary',

  'operatorStation.create.domain.title': 'Domain Setup',
  'operatorStation.create.domain.tipTitle': 'Tip',
  'operatorStation.create.domain.tip':
    'Multiple domain records can be added for the same domain type, but only one record can be enabled at a time.',
  'operatorStation.create.domain.addButton': 'Add Domain',
  'operatorStation.create.domain.column.type': 'Domain Type',
  'operatorStation.create.domain.column.domain': 'Domain',
  'operatorStation.create.domain.column.status': 'Status',
  'operatorStation.create.domain.column.remark': 'Remark',
  'operatorStation.create.domain.column.operations': 'Operations',
  'operatorStation.create.domain.type.admin': 'Station Backend',
  'operatorStation.create.domain.type.agent': 'Agent Backend',
  'operatorStation.create.domain.type.h5': 'Member H5',
  'operatorStation.create.domain.modal.addTitle': 'Add Domain',
  'operatorStation.create.domain.modal.editTitle': 'Edit Domain',
  'operatorStation.create.domain.field.type': 'Domain Type',
  'operatorStation.create.domain.field.domain': 'Domain',
  'operatorStation.create.domain.field.domainPlaceholder':
    'Enter domain, e.g. admin.th.example.com',
  'operatorStation.create.domain.field.status': 'Status',
  'operatorStation.create.domain.field.remark': 'Remark',
  'operatorStation.create.domain.field.remarkPlaceholder':
    'Enter remark (optional)',
  'operatorStation.create.domain.rule.typeRequired':
    'Please select a domain type',
  'operatorStation.create.domain.rule.domainRequired': 'Please enter a domain',
  'operatorStation.create.domain.delete.content': 'Delete domain "{domain}"?',
  'operatorStation.create.domain.empty':
    'No domains yet. Click "Add Domain" to create one.',
  'operatorStation.create.domain.message.editLocalOnly':
    'This domain already exists on the server. There is no update API yet, so only the wizard display was updated.',
  'operatorStation.create.domain.message.deleteLocalOnly':
    'This domain already exists on the server. There is no delete API yet, so it was only removed from the wizard list.',

  'operatorStation.create.resource.title': 'Basic Resources',
  'operatorStation.create.resource.tip':
    'Complete the basic resource setup in this wizard to create the station.',
  'operatorStation.create.resource.tab.language': 'Language',
  'operatorStation.create.resource.tab.region': 'Region',
  'operatorStation.create.resource.tab.proxyLine': 'Agent Sub-line',
  'operatorStation.create.resource.searchPlaceholder': 'Search by name or code',
  'operatorStation.create.resource.filter.all': 'All',
  'operatorStation.create.resource.filter.assigned': 'Assigned',
  'operatorStation.create.resource.filter.unassigned': 'Unassigned',
  'operatorStation.create.resource.column.name': 'Name',
  'operatorStation.create.resource.column.code': 'Code',
  'operatorStation.create.resource.column.status': 'Status',
  'operatorStation.create.resource.column.assigned': 'Assigned',
  'operatorStation.create.resource.column.assignedAt': 'Assigned At',
  'operatorStation.create.resource.summaryTitle': 'Resources Selected',
  'operatorStation.create.resource.summary.language': 'Language',
  'operatorStation.create.resource.summary.region': 'Region',
  'operatorStation.create.resource.summary.proxyLine': 'Agent Sub-line',
  'operatorStation.create.resource.unit': 'items',

  'operatorStation.create.rule.nameRequired': 'Please enter station name',
  'operatorStation.create.rule.timezoneRequired': 'Please select timezone',
  'operatorStation.create.rule.currencyRequired':
    'Please select settlement currency',
  'operatorStation.create.rule.statusRequired': 'Please select status',
  'operatorStation.create.rule.companyNameRequired':
    'Please enter company name',
  'operatorStation.create.rule.contactNameRequired':
    'Please enter primary contact name',
  'operatorStation.create.rule.contactEmailRequired':
    'Please enter primary contact email',
  'operatorStation.create.rule.contactEmailInvalid':
    'Please enter a valid email address',
  'operatorStation.create.rule.accountRequired': 'Please enter login account',
  'operatorStation.create.rule.passwordRequired': 'Please enter login password',
  'operatorStation.create.rule.passwordLength':
    'Password must be 6-32 characters',
  'operatorStation.create.rule.confirmPasswordRequired':
    'Please re-enter password',
  'operatorStation.create.rule.passwordNotMatch': 'Passwords do not match',
  'operatorStation.create.rule.displayNameRequired':
    'Please enter display name',

  'operatorStation.create.message.saveDraftSuccess': 'Draft saved',
  'operatorStation.create.message.createSuccess':
    'Station "{name}" created successfully',
  'operatorStation.create.message.completeFailed':
    'The station was saved as a draft but could not be completed. Please retry "Complete" from the list.',
  'operatorStation.create.message.pendingApiTip':
    'Admin account has no backend API yet. It is stored locally for preview only until the backend is available.',
  'operatorStation.create.message.needBaseInfoFirst':
    'Please complete basic info and create the station draft first',

  /** Station domain management page, independent namespace prefixed operatorDomain.* */
  'operatorDomain.title': 'Domain Management',
  'operatorDomain.subtitle':
    'Manage the station backend, agent backend and member H5 domains of each station.',

  'operatorDomain.search.operator': 'Station:',
  'operatorDomain.search.operatorPlaceholder': 'Select a station',
  'operatorDomain.search.keyword': 'Domain Keyword:',
  'operatorDomain.search.keywordPlaceholder':
    'Enter domain, up to 253 characters',
  'operatorDomain.search.domainType': 'Domain Type:',
  'operatorDomain.search.status': 'Status:',

  'operatorDomain.toolbar.create': 'Add Domain',

  'operatorDomain.type.admin': 'Station Backend',
  'operatorDomain.type.agent': 'Agent Backend',
  'operatorDomain.type.h5': 'Member H5',

  'operatorDomain.column.id': 'Domain ID',
  'operatorDomain.column.operator': 'Station',
  'operatorDomain.column.domainName': 'Domain',
  'operatorDomain.column.domainType': 'Domain Type',
  'operatorDomain.column.status': 'Status',
  'operatorDomain.column.remark': 'Remark',
  'operatorDomain.column.updatedAt': 'Updated At',

  'operatorDomain.action.detail': 'Detail',
  'operatorDomain.action.edit': 'Edit',
  'operatorDomain.action.enable': 'Enable',
  'operatorDomain.action.disable': 'Disable',
  'operatorDomain.action.delete': 'Delete',

  'operatorDomain.modal.createTitle': 'Add Domain',
  'operatorDomain.modal.editTitle': 'Edit Domain',
  'operatorDomain.modal.detailTitle': 'Domain Detail',
  'operatorDomain.field.operator': 'Station',
  'operatorDomain.field.operatorPlaceholder': 'Select a station',
  'operatorDomain.field.domainType': 'Domain Type',
  'operatorDomain.field.domainName': 'Domain',
  'operatorDomain.field.domainNamePlaceholder':
    'Enter domain, e.g. admin.th.example.com',
  'operatorDomain.field.status': 'Status',
  'operatorDomain.field.remark': 'Remark',
  'operatorDomain.field.remarkPlaceholder':
    'Enter remark (optional), up to 1000 characters',

  'operatorDomain.rule.operatorRequired': 'Please select a station',
  'operatorDomain.rule.domainTypeRequired': 'Please select a domain type',
  'operatorDomain.rule.domainNameRequired': 'Please enter a domain',

  'operatorDomain.message.createSuccess': 'Domain created successfully',
  'operatorDomain.message.updateSuccess': 'Domain updated successfully',
  'operatorDomain.message.noChange': 'Nothing changed',
  'operatorDomain.message.statusUpdateSuccess': 'Status updated successfully',
  'operatorDomain.message.deleteConfirmContent':
    'This cannot be undone. Delete domain "{domain}"?',

  /** Station admin account management page, namespace prefixed operatorAdmin.* */
  'operatorAdmin.title': 'Admin Accounts',
  'operatorAdmin.subtitle':
    'Manage station backend admin accounts, status and basic information.',

  'operatorAdmin.search.operator': 'Station:',
  'operatorAdmin.search.operatorPlaceholder': 'Select a station',
  'operatorAdmin.search.keyword': 'Keyword:',
  'operatorAdmin.search.keywordPlaceholder':
    'Enter username or display name, up to 100 characters',
  'operatorAdmin.search.status': 'Status:',

  'operatorAdmin.toolbar.create': 'Add Admin',

  'operatorAdmin.column.id': 'Admin ID',
  'operatorAdmin.column.operator': 'Station',
  'operatorAdmin.column.username': 'Username',
  'operatorAdmin.column.displayName': 'Display Name',
  'operatorAdmin.column.status': 'Status',
  'operatorAdmin.column.updatedAt': 'Updated At',

  'operatorAdmin.action.edit': 'Edit',
  'operatorAdmin.action.resetPassword': 'Reset Password',
  'operatorAdmin.action.enable': 'Enable',
  'operatorAdmin.action.disable': 'Disable',

  'operatorAdmin.modal.createTitle': 'Add Admin',
  'operatorAdmin.modal.editTitle': 'Edit Admin',
  'operatorAdmin.modal.resetPasswordTitle': 'Reset Password',
  'operatorAdmin.modal.resetPasswordTip':
    'Set a new password for admin "{username}". It takes effect immediately.',
  'operatorAdmin.field.operator': 'Station',
  'operatorAdmin.field.operatorPlaceholder': 'Select a station',
  'operatorAdmin.field.username': 'Username',
  'operatorAdmin.field.usernamePlaceholder':
    'Enter username (letters, digits, underscore)',
  'operatorAdmin.field.password': 'Password',
  'operatorAdmin.field.passwordPlaceholder':
    'Enter password, 6-32 characters',
  'operatorAdmin.field.passwordOptionalPlaceholder':
    'Leave blank to keep current password',
  'operatorAdmin.field.confirmPassword': 'Confirm Password',
  'operatorAdmin.field.confirmPasswordPlaceholder': 'Re-enter password',
  'operatorAdmin.field.confirmPasswordOptionalPlaceholder':
    'Re-enter if changing password',
  'operatorAdmin.field.displayName': 'Display Name',
  'operatorAdmin.field.displayNamePlaceholder':
    'Enter display name, e.g. Operations Admin',
  'operatorAdmin.field.status': 'Status',

  'operatorAdmin.rule.operatorRequired': 'Please select a station',
  'operatorAdmin.rule.usernameRequired': 'Please enter username',
  'operatorAdmin.rule.passwordRequired': 'Please enter password',
  'operatorAdmin.rule.passwordLength': 'Password must be 6-32 characters',
  'operatorAdmin.rule.confirmPasswordRequired': 'Please re-enter password',
  'operatorAdmin.rule.passwordNotMatch': 'Passwords do not match',
  'operatorAdmin.rule.displayNameRequired': 'Please enter display name',

  'operatorAdmin.message.createSuccess': 'Admin created successfully',
  'operatorAdmin.message.updateSuccess': 'Admin updated successfully',
  'operatorAdmin.message.statusUpdateSuccess': 'Status updated successfully',
  'operatorAdmin.message.resetPasswordSuccess': 'Password reset successfully',

  /** Station basic resource allocation summary page, namespace prefixed operatorBasicResource.* */
  'operatorBasicResource.title': 'Basic Resource Allocation',
  'operatorBasicResource.subtitle':
    'View allocated language, region and agent sub-line counts for each station.',

  'operatorBasicResource.search.keyword': 'Keyword:',
  'operatorBasicResource.search.keywordPlaceholder':
    'Enter station name or code, up to 100 characters',

  'operatorBasicResource.column.operatorId': 'Station ID',
  'operatorBasicResource.column.operatorCode': 'Station Code',
  'operatorBasicResource.column.operatorName': 'Station Name',
  'operatorBasicResource.column.languageCount': 'Languages',
  'operatorBasicResource.column.regionCount': 'Regions',
  'operatorBasicResource.column.agentLineCount': 'Agent Sub-lines',
  'operatorBasicResource.column.unit': 'items',
  'operatorBasicResource.column.operations': 'Operations',
  'operatorBasicResource.action.configure': 'Configure',

  'operatorBasicResource.modal.title': 'Resource Allocation',
  'operatorBasicResource.modal.selectOperator': 'Station',
  'operatorBasicResource.modal.selectOperatorPlaceholder': 'Select a station',
  'operatorBasicResource.modal.info.id': 'Station ID',
  'operatorBasicResource.modal.info.code': 'Station Code',
  'operatorBasicResource.modal.info.name': 'Station Name',
  'operatorBasicResource.modal.info.timezone': 'Timezone',
  'operatorBasicResource.modal.info.currency': 'Currency',
  'operatorBasicResource.modal.tab.language': 'Language',
  'operatorBasicResource.modal.tab.region': 'Region',
  'operatorBasicResource.modal.tab.proxyLine': 'Agent Sub-line',
  'operatorBasicResource.modal.searchPlaceholder': 'Search by name or code',
  'operatorBasicResource.modal.filter.all': 'All',
  'operatorBasicResource.modal.filter.enabled': 'Enabled',
  'operatorBasicResource.modal.filter.disabled': 'Disabled',
  'operatorBasicResource.modal.column.name': 'Name',
  'operatorBasicResource.modal.column.code': 'Code',
  'operatorBasicResource.modal.column.status': 'Status',
  'operatorBasicResource.modal.column.assigned': 'Assigned',
  'operatorBasicResource.modal.column.assignedAt': 'Assigned At',
  'operatorBasicResource.modal.summary.title': 'Current Allocation',
  'operatorBasicResource.modal.summary.language': 'Language',
  'operatorBasicResource.modal.summary.region': 'Region',
  'operatorBasicResource.modal.summary.proxyLine': 'Agent Sub-line',
  'operatorBasicResource.modal.summary.desc': 'Assigned / Total',
  'operatorBasicResource.modal.action.save': 'Save',
  'operatorBasicResource.modal.action.resetSelection': 'Reset Selection',
  'operatorBasicResource.modal.tip.title': 'Tip:',
  'operatorBasicResource.modal.tip.line1':
    '1. Select a station above before allocating resources.',
  'operatorBasicResource.modal.tip.line2':
    '2. Changes take effect immediately after saving.',
  'operatorBasicResource.modal.message.saveSuccess':
    'Configuration saved successfully',
  'operatorBasicResource.modal.message.selectOperatorFirst':
    'Please select a station first',
  'operatorBasicResource.modal.empty': 'No data',

  /** Station game resource allocation summary page, namespace prefixed operatorGameResource.* */
  'operatorGameResource.title': 'Game Resource Allocation',
  'operatorGameResource.subtitle':
    'View allocated game, category, provider and channel counts for each station.',

  'operatorGameResource.search.opCode': 'Station Code:',
  'operatorGameResource.search.opCodePlaceholder':
    'Enter station business code, up to 64 characters',
  'operatorGameResource.search.status': 'Station Status:',

  'operatorGameResource.column.id': 'ID',
  'operatorGameResource.column.opCode': 'Station Code',
  'operatorGameResource.column.opName': 'Station Name',
  'operatorGameResource.column.gameCount': 'Games',
  'operatorGameResource.column.gameCategoryCount': 'Categories',
  'operatorGameResource.column.gameProviderCount': 'Providers',
  'operatorGameResource.column.gameChannelCount': 'Channels',
  'operatorGameResource.column.updatedAt': 'Updated At',
  'operatorGameResource.column.unit': 'items',
  'operatorGameResource.column.gameUnit': 'games',
  'operatorGameResource.column.operations': 'Operations',
  'operatorGameResource.action.configure': 'Configure',

  'operatorGameResource.modal.title': 'Game Resource Configuration',
  'operatorGameResource.modal.selectOperator': 'Current Station',
  'operatorGameResource.modal.selectOperatorPlaceholder': 'Select a station',
  'operatorGameResource.modal.info.code': 'Station Code',
  'operatorGameResource.modal.info.timezone': 'Timezone',
  'operatorGameResource.modal.info.domain': 'Domain',
  'operatorGameResource.modal.info.status': 'Status',
  'operatorGameResource.modal.tab.category': 'Categories',
  'operatorGameResource.modal.tab.provider': 'Providers',
  'operatorGameResource.modal.tab.channel': 'Channels',
  'operatorGameResource.modal.tab.game': 'Games',
  'operatorGameResource.modal.search.keyword': 'Keyword',
  'operatorGameResource.modal.search.keywordPlaceholder.category':
    'Enter category name or code',
  'operatorGameResource.modal.search.keywordPlaceholder.provider':
    'Enter provider name or code',
  'operatorGameResource.modal.search.keywordPlaceholder.channel':
    'Enter channel name or code',
  'operatorGameResource.modal.search.keywordPlaceholder.game':
    'Enter game name or code',
  'operatorGameResource.modal.search.status': 'Status',
  'operatorGameResource.modal.search.allocated': 'Allocated',
  'operatorGameResource.modal.filter.all': 'All',
  'operatorGameResource.modal.filter.enabled': 'Enabled',
  'operatorGameResource.modal.filter.disabled': 'Disabled',
  'operatorGameResource.modal.filter.allocated': 'Allocated',
  'operatorGameResource.modal.filter.unallocated': 'Unallocated',
  'operatorGameResource.modal.column.name': 'Resource Name',
  'operatorGameResource.modal.column.code': 'Code',
  'operatorGameResource.modal.column.status': 'Status',
  'operatorGameResource.modal.column.allocated': 'Allocated',
  'operatorGameResource.modal.column.allocatedAt': 'Allocated At',
  'operatorGameResource.modal.column.remark': 'Remark',
  'operatorGameResource.modal.column.operations': 'Operations',
  'operatorGameResource.modal.allocated.yes': 'Allocated',
  'operatorGameResource.modal.allocated.no': 'Unallocated',
  'operatorGameResource.modal.action.detail': 'Details',
  'operatorGameResource.modal.action.save': 'Save',
  'operatorGameResource.modal.action.resetSelection': 'Reset Selection',
  'operatorGameResource.modal.action.batchStatus': 'Batch Update Status',
  'operatorGameResource.modal.action.batchEnable': 'Enable Selected',
  'operatorGameResource.modal.action.batchDisable': 'Disable Selected',
  'operatorGameResource.modal.summary.title': 'Selected Resources',
  'operatorGameResource.modal.summary.category': 'Categories',
  'operatorGameResource.modal.summary.provider': 'Providers',
  'operatorGameResource.modal.summary.channel': 'Channels',
  'operatorGameResource.modal.summary.game': 'Games',
  'operatorGameResource.modal.summary.unit': 'items',
  'operatorGameResource.modal.tip.title': 'Tip:',
  'operatorGameResource.modal.tip.content':
    'Select appropriate game resources based on the station strategy. Changes take effect immediately after saving.',
  'operatorGameResource.modal.message.selectOperatorFirst':
    'Please select a station first',
  'operatorGameResource.modal.message.saveSuccess':
    'Configuration saved successfully',
  'operatorGameResource.modal.message.selectItemsFirst':
    'Please select resources first',
  'operatorGameResource.modal.message.batchStatusSuccess':
    'Status updated successfully',
  'operatorGameResource.modal.message.saveEmpty':
    'Please select at least one resource before saving',
  'operatorGameResource.modal.detail.title': 'Resource Details',
  'operatorGameResource.modal.empty': 'No data',
};
