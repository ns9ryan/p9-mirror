import i18n from '@/language';
import type {
  KnowledgeDocumentStatus,
  KnowledgeDocumentType,
} from '@/api/aichat';

const DOCUMENT_TYPES: {
  labelKey: string;
  value: KnowledgeDocumentType;
  color: string;
}[] = [
  {
    labelKey: 'aichatKnowledge.documentType.tableSchema',
    value: 'TABLE_SCHEMA',
    color: 'arcoblue',
  },
  {
    labelKey: 'aichatKnowledge.documentType.metricDefinition',
    value: 'METRIC_DEFINITION',
    color: 'purple',
  },
  {
    labelKey: 'aichatKnowledge.documentType.sqlExample',
    value: 'SQL_EXAMPLE',
    color: 'cyan',
  },
  {
    labelKey: 'aichatKnowledge.documentType.businessRule',
    value: 'BUSINESS_RULE',
    color: 'orange',
  },
  {
    labelKey: 'aichatKnowledge.documentType.knowledge',
    value: 'KNOWLEDGE',
    color: 'gray',
  },
];

const STATUS_FILTERS: {
  labelKey: string;
  value: KnowledgeDocumentStatus;
}[] = [
  { labelKey: 'aichatKnowledge.status.draft', value: 'DRAFT' },
  { labelKey: 'aichatKnowledge.status.published', value: 'PUBLISHED' },
  { labelKey: 'aichatKnowledge.status.archived', value: 'ARCHIVED' },
];

const STATUS_META: Record<
  KnowledgeDocumentStatus,
  { labelKey: string; color: string }
> = {
  PROCESSING: {
    labelKey: 'aichatKnowledge.status.processing',
    color: 'arcoblue',
  },
  COMPLETED: { labelKey: 'aichatKnowledge.status.completed', color: 'green' },
  FAILED: { labelKey: 'aichatKnowledge.status.failed', color: 'red' },
  DRAFT: { labelKey: 'aichatKnowledge.status.draft', color: 'gray' },
  PUBLISHED: { labelKey: 'aichatKnowledge.status.published', color: 'green' },
  ARCHIVED: { labelKey: 'aichatKnowledge.status.archived', color: 'orangered' },
};

const t = (key: string) => i18n.global.t(key);

/** 需在 computed / 渲染中调用，以便切换语言后文案同步更新 */
export const getDocumentTypeOptions = () =>
  DOCUMENT_TYPES.map(({ labelKey, value, color }) => ({
    label: t(labelKey),
    value,
    color,
  }));

export const getStatusFilterOptions = () =>
  STATUS_FILTERS.map(({ labelKey, value }) => ({ label: t(labelKey), value }));

export const documentTypeMeta = (type?: string) => {
  const item = DOCUMENT_TYPES.find((option) => option.value === type);
  return item
    ? { label: t(item.labelKey), value: item.value, color: item.color }
    : { label: type || '--', value: type, color: 'gray' };
};

export const statusMeta = (status?: string) => {
  const item = STATUS_META[status as KnowledgeDocumentStatus];
  return item
    ? { label: t(item.labelKey), color: item.color }
    : { label: status || '--', color: 'gray' };
};
