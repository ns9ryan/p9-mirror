/** 本次会话内代理初始化已就绪，后续导航不再请求 /init/status */
let ready = false;

export function isAgentInitReady() {
  return ready;
}

export function markAgentInitReady() {
  ready = true;
}

export function resetAgentInitReady() {
  ready = false;
}
