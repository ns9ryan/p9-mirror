import { h } from 'vue';
import type { Router } from 'vue-router';
import NProgress from 'nprogress';
import { Modal, Tag } from '@arco-design/web-vue';
import { getAgentInitStatus } from '@/api/agent';
import type { AgentInitStatus } from '@/api/agent';
import i18n from '@/language';
import { useUserStore } from '@/store';
import { isLogin, isPreviewSession } from '@/utils/auth';
import { isAgentInitReady, markAgentInitReady } from '@/utils/agent-init';
import { AGENT_INIT_ROUTE_NAME } from '../constants';

const SKIP_ROUTE_NAMES = new Set([
  'login',
  'noPermission',
  AGENT_INIT_ROUTE_NAME,
]);

function lineLabel(lineCode: string) {
  const { t } = i18n.global;
  const key = `agentList.line.${String(lineCode || '').toUpperCase()}`;
  const text = t(key);
  return text === key ? lineCode : text;
}

function renderNotice(status: AgentInitStatus | null) {
  const { t } = i18n.global;
  const pendingLines = (status?.lines || []).filter(
    (line) => !line.initialized
  );
  const steps = [
    {
      text: t('agentInit.notice.stepLevels'),
      done: Boolean(status?.levels_ready),
    },
    { text: t('agentInit.notice.stepAgent'), done: false },
  ];
  return h('div', { style: 'line-height: 22px' }, [
    h('p', { style: 'margin: 0 0 12px' }, t('agentInit.notice.desc')),
    pendingLines.length
      ? h(
          'div',
          {
            style:
              'display: flex; flex-wrap: wrap; gap: 6px; align-items: center; margin-bottom: 12px',
          },
          [
            h('span', t('agentInit.notice.pending')),
            ...pendingLines.map((line) =>
              h(Tag, { color: 'orangered', size: 'small' }, () =>
                lineLabel(line.line_code)
              )
            ),
          ]
        )
      : null,
    h(
      'ol',
      { style: 'margin: 0; padding-left: 20px' },
      steps.map((step) =>
        h('li', [
          step.text,
          step.done
            ? h(
                Tag,
                { color: 'green', size: 'small', style: 'margin-left: 6px' },
                () => t('agentInit.notice.done')
              )
            : null,
        ])
      )
    ),
  ]);
}

/** 提示代理未初始化，选择开始初始化返回 true，退出登录返回 false */
function confirmAgentInit(status: AgentInitStatus | null) {
  const { t } = i18n.global;
  NProgress.done();
  return new Promise<boolean>((resolve) => {
    Modal.confirm({
      title: t('agentInit.notice.title'),
      content: () => renderNotice(status),
      okText: t('agentInit.notice.start'),
      cancelText: t('agentInit.logout'),
      closable: false,
      maskClosable: false,
      escToClose: false,
      onOk: () => resolve(true),
      onCancel: () => resolve(false),
    });
  });
}

function wait(ms: number) {
  return new Promise<void>((resolve) => {
    window.setTimeout(resolve, ms);
  });
}

/** 拉取初始化状态；失败时重试，仍失败则抛出 */
async function fetchAgentInitStatus(retries = 1): Promise<AgentInitStatus | null> {
  let lastError: unknown;
  for (let attempt = 0; attempt <= retries; attempt += 1) {
    try {
      const { data } = await getAgentInitStatus();
      return data || null;
    } catch (error) {
      lastError = error;
      if (attempt < retries) {
        await wait(300);
      }
    }
  }
  throw lastError;
}

/** 已开通代理线未全部初始化时，进入业务页前先转到初始化页 */
export default function setupAgentInitGuard(router: Router) {
  router.beforeEach(async (to, from) => {
    if (
      SKIP_ROUTE_NAMES.has(String(to.name)) ||
      !isLogin() ||
      // 预览 token 只能查看，无法提交初始化
      isPreviewSession() ||
      isAgentInitReady()
    ) {
      return true;
    }
    let status: AgentInitStatus | null = null;
    try {
      const data = await fetchAgentInitStatus(1);
      // 空响应视为未知，避免误进初始化页再被踢回
      if (!data) {
        return true;
      }
      if (data.ready) {
        markAgentInitReady();
        return true;
      }
      status = data;
    } catch {
      // 重试后仍失败：状态未知，放行本次导航，下次路由再检查
      return true;
    }
    // 刚登录时先弹窗说明再进入初始化页；刷新或直接访问时已提示过，直接跳转
    if (from.name === 'login' && !(await confirmAgentInit(status))) {
      await useUserStore().logout();
      return false;
    }
    return {
      name: AGENT_INIT_ROUTE_NAME,
      query: { redirect: to.fullPath },
      replace: true,
    };
  });
}
