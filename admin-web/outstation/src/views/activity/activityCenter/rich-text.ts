/** Quill 编辑器产出的标签；其余标签只保留文字，script/style 等整段丢弃 */
const ALLOWED_TAGS = new Set([
  'P',
  'BR',
  'STRONG',
  'B',
  'EM',
  'I',
  'U',
  'S',
  'OL',
  'UL',
  'LI',
  'BLOCKQUOTE',
  'PRE',
  'CODE',
  'H1',
  'H2',
  'H3',
  'SPAN',
]);

const DROPPED_TAGS = new Set([
  'SCRIPT',
  'STYLE',
  'IFRAME',
  'OBJECT',
  'EMBED',
  'TEMPLATE',
]);

/** 只保留文字颜色和背景色，避免 url()、定位等样式 */
const ALLOWED_STYLES = ['color', 'background-color'];

const HTML_PATTERN = /<\/?[a-z][\s\S]*>/i;

export function isRichText(value: string) {
  return HTML_PATTERN.test(value);
}

function cleanStyle(source: HTMLElement, target: HTMLElement) {
  ALLOWED_STYLES.forEach((name) => {
    const value = source.style.getPropertyValue(name);
    if (value && !/url\(|expression/i.test(value)) {
      target.style.setProperty(name, value);
    }
  });
}

function cleanNode(node: Node, doc: Document): Node[] {
  if (node.nodeType === Node.TEXT_NODE) {
    return [doc.createTextNode(node.textContent || '')];
  }
  if (node.nodeType !== Node.ELEMENT_NODE) return [];
  const element = node as HTMLElement;
  if (DROPPED_TAGS.has(element.tagName)) return [];
  const children = Array.from(element.childNodes).flatMap((child) =>
    cleanNode(child, doc)
  );
  if (!ALLOWED_TAGS.has(element.tagName)) return children;
  const clean = doc.createElement(element.tagName.toLowerCase());
  const qlClasses = Array.from(element.classList).filter((name) =>
    name.startsWith('ql-')
  );
  if (qlClasses.length) clean.className = qlClasses.join(' ');
  cleanStyle(element, clean);
  children.forEach((child) => clean.appendChild(child));
  return [clean];
}

/** 白名单过滤富文本，去掉所有事件属性、链接和非白名单标签 */
export function sanitizeRichText(html: string) {
  const doc = new DOMParser().parseFromString(html, 'text/html');
  const container = doc.createElement('div');
  Array.from(doc.body.childNodes)
    .flatMap((node) => cleanNode(node, doc))
    .forEach((node) => container.appendChild(node));
  return container.innerHTML;
}
