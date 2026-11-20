<template>
  <div class="message-content">
    <template v-for="(block, index) in blocks" :key="index">
      <div v-if="block.type === 'code'" class="code-block">
        <div class="code-block-header">
          <span>{{ block.lang || 'code' }}</span>
          <a-link class="code-copy" @click="copy(block.text)">
            <template #icon><icon-copy /></template>
            复制
          </a-link>
        </div>
        <pre><code>{{ block.text }}</code></pre>
      </div>
      <div v-else-if="block.type === 'table'" class="table-block">
        <table>
          <thead>
            <tr>
              <th v-for="(cell, ci) in block.header" :key="ci">{{ cell }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(row, ri) in block.rows" :key="ri">
              <td v-for="(cell, ci) in row" :key="ci">{{ cell }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <p v-else class="text-block">
        <template v-for="(seg, si) in block.segments" :key="si">
          <code v-if="seg.type === 'code'" class="inline-code">{{
            seg.text
          }}</code>
          <strong v-else-if="seg.type === 'bold'">{{ seg.text }}</strong>
          <template v-else>{{ seg.text }}</template>
        </template>
      </p>
    </template>
  </div>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import { Message } from '@arco-design/web-vue';

  interface Segment {
    type: 'text' | 'code' | 'bold';
    text: string;
  }

  type Block =
    | { type: 'code'; lang: string; text: string }
    | { type: 'table'; header: string[]; rows: string[][] }
    | { type: 'text'; segments: Segment[] };

  const props = defineProps<{ content: string }>();

  const TABLE_SEPARATOR = /^\s*\|?\s*:?-{2,}:?\s*(\|\s*:?-{2,}:?\s*)*\|?\s*$/;

  const splitRow = (line: string) =>
    line
      .trim()
      .replace(/^\|/, '')
      .replace(/\|$/, '')
      .split('|')
      .map((cell) => cell.trim());

  const parseInline = (text: string): Segment[] =>
    text
      .split(/(`[^`\n]+`|\*\*[^*\n]+\*\*)/)
      .filter(Boolean)
      .map((part): Segment => {
        if (part.startsWith('`') && part.endsWith('`')) {
          return { type: 'code', text: part.slice(1, -1) };
        }
        if (part.startsWith('**') && part.endsWith('**')) {
          return { type: 'bold', text: part.slice(2, -2) };
        }
        return { type: 'text', text: part };
      });

  const tryFormatJson = (content: string) => {
    const trimmed = content.trim();
    if (!/^[[{]/.test(trimmed)) return null;
    try {
      return JSON.stringify(JSON.parse(trimmed), null, 2);
    } catch {
      return null;
    }
  };

  const parseBlocks = (content: string): Block[] => {
    const json = tryFormatJson(content);
    if (json) return [{ type: 'code', lang: 'json', text: json }];

    const result: Block[] = [];
    const lines = content.replace(/\r\n/g, '\n').split('\n');
    let textBuffer: string[] = [];

    const flushText = () => {
      const text = textBuffer.join('\n').trim();
      if (text) result.push({ type: 'text', segments: parseInline(text) });
      textBuffer = [];
    };

    let i = 0;
    while (i < lines.length) {
      const line = lines[i];
      const fence = line.match(/^\s*```\s*([\w-]*)/);
      if (fence) {
        flushText();
        const codeLines: string[] = [];
        i += 1;
        while (i < lines.length && !/^\s*```/.test(lines[i])) {
          codeLines.push(lines[i]);
          i += 1;
        }
        result.push({
          type: 'code',
          lang: fence[1],
          text: codeLines.join('\n'),
        });
        i += 1;
      } else if (
        line.includes('|') &&
        i + 1 < lines.length &&
        TABLE_SEPARATOR.test(lines[i + 1])
      ) {
        flushText();
        const header = splitRow(line);
        const rows: string[][] = [];
        i += 2;
        while (i < lines.length && lines[i].includes('|') && lines[i].trim()) {
          rows.push(splitRow(lines[i]));
          i += 1;
        }
        result.push({ type: 'table', header, rows });
      } else {
        textBuffer.push(line);
        i += 1;
      }
    }
    flushText();
    return result;
  };

  const blocks = computed(() => parseBlocks(props.content || ''));

  const copy = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      Message.success('已复制');
    } catch {
      Message.error('复制失败，请手动选择复制');
    }
  };
</script>

<style scoped lang="less">
  .message-content {
    color: var(--color-text-1);
    font-size: 14px;
    line-height: 1.75;
    word-break: break-word;
  }

  .text-block {
    margin: 0;
    white-space: pre-wrap;

    & + * {
      margin-top: 8px;
    }
  }

  .inline-code {
    padding: 1px 6px;
    color: rgb(var(--arcoblue-6));
    font-family: Consolas, Monaco, monospace;
    font-size: 13px;
    background: var(--color-fill-2);
    border-radius: 4px;
  }

  .code-block {
    margin: 8px 0;
    overflow: hidden;
    border: 1px solid var(--color-border-2);
    border-radius: 6px;

    &-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 4px 12px;
      color: var(--color-text-3);
      font-size: 12px;
      background: var(--color-fill-2);
    }

    pre {
      margin: 0;
      padding: 10px 12px;
      overflow-x: auto;
      font-family: Consolas, Monaco, monospace;
      font-size: 12px;
      line-height: 1.6;
      white-space: pre;
      background: var(--color-fill-1);
    }
  }

  .code-copy {
    font-size: 12px;
  }

  .table-block {
    margin: 8px 0;
    overflow-x: auto;

    table {
      min-width: 100%;
      font-size: 13px;
      border-collapse: collapse;
    }

    th,
    td {
      padding: 6px 12px;
      white-space: nowrap;
      text-align: left;
      border: 1px solid var(--color-border-2);
    }

    th {
      font-weight: 500;
      background: var(--color-fill-2);
    }

    tbody tr:nth-child(even) {
      background: var(--color-fill-1);
    }
  }
</style>
