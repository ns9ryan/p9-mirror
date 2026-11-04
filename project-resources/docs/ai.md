# AI 聊天 API 文档

`mcp-agent` 提供 AI 聊天系统的 HTTP API，支持场景管理、会话管理、消息流式处理。

当前包含：

- 场景管理
- 会话管理
- 消息处理

---

## 一、接口约定

### 服务地址

默认开发端口：

```text
28003
```

实际 Host 和域名以部署环境为准。

### 基础路径

```text
/chat
```

### 请求格式

- GET 接口使用 Query 参数
- POST 接口使用 JSON Body

```http
Content-Type: application/json
```

### 响应格式

统一响应格式：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

- `code`: 错误码，0 表示成功
- `message`: 错误消息或成功提示
- `data`: 返回的数据内容

---

## 二、场景管理接口

前缀：`/chat`

### 1. 获取场景列表

**使用场景**

初始化 AI 聊天时，获取所有可用的聊天场景（例如：数据分析、客服等）。

**请求**

```http
GET /scenarios
```

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "scenarios": [
      {
        "name": "数据分析",
        "description": "用于数据查询和分析",
        "icon": "chart",
        "model_name": "gpt-4",
        "temperature": 0.7,
        "available_tools": ["sql", "knowledge_base"],
        "max_recent_messages": 10
      },
      {
        "name": "客服助手",
        "description": "用于客户服务",
        "icon": "service",
        "model_name": "gpt-3.5-turbo",
        "temperature": 0.5,
        "available_tools": ["faq", "knowledge_base"],
        "max_recent_messages": 20
      }
    ]
  }
}
```

**响应字段说明**

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `name` | string | 场景名称，唯一标识 |
| `description` | string | 场景描述 |
| `icon` | string | 场景图标 |
| `model_name` | string | 使用的 AI 模型名称 |
| `temperature` | float | 模型温度参数（0.0-2.0） |
| `available_tools` | array | 可用工具列表 |
| `max_recent_messages` | int | 最大保留最近消息数 |

---

## 三、会话管理接口

前缀：`/chat`

### 1. 获取会话列表

**使用场景**

用户打开应用时，获取当前用户的所有聊天会话列表，显示在侧边栏。

**请求**

```http
GET /sessions
```

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "sessions": [
      {
        "id": 1,
        "session_code": "sess_abc123",
        "user_id": 123,
        "op_code": "op001",
        "title": "分析用户登录数据",
        "status": "ACTIVE",
        "scene": "数据分析",
        "context_summary": "讨论最近7天的用户登录统计",
        "message_count": 8,
        "created_at": 1698000000000,
        "updated_at": 1698001000000
      },
      {
        "id": 2,
        "session_code": "sess_def456",
        "user_id": 123,
        "op_code": "op001",
        "title": "支付问题咨询",
        "status": "ACTIVE",
        "scene": "客服助手",
        "context_summary": "用户咨询支付流程问题",
        "message_count": 5,
        "created_at": 1697900000000,
        "updated_at": 1697910000000
      }
    ]
  }
}
```

**响应字段说明**

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 会话数据库 ID |
| `session_code` | string | 会话编码（唯一标识） |
| `user_id` | int64 | 所属用户 ID |
| `op_code` | string | 运营商编码 |
| `title` | string | 会话标题 |
| `status` | string | 会话状态：ACTIVE/CLOSED |
| `scene` | string | 场景名称 |
| `context_summary` | string | 对话上下文摘要 |
| `message_count` | int | 会话中的消息数量 |
| `created_at` | int64 | 创建时间（毫秒时间戳） |
| `updated_at` | int64 | 更新时间（毫秒时间戳） |

### 2. 创建会话

**使用场景**

用户点击"新建对话"时，为指定场景创建一个新的聊天会话。

**请求**

```http
POST /sessions
Content-Type: application/json

{
  "title": "我的新聊天",
  "scene": "数据分析"
}
```

**请求体**

| 字段 | 类型 | 必需 | 说明 |
| --- | --- | --- | --- |
| `title` | string | 否 | 会话标题（默认为场景名称） |
| `scene` | string | 是 | 场景名称 |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 3,
    "session_code": "sess_ghi789",
    "user_id": 123,
    "op_code": "op001",
    "title": "我的新聊天",
    "status": "ACTIVE",
    "scene": "数据分析",
    "created_at": 1698100000000,
    "updated_at": 1698100000000
  }
}
```

---

## 四、消息处理接口

前缀：`/chat`

### 1. 获取会话消息

**使用场景**

用户切换或打开聊天会话时，获取该会话的历史消息列表。

**请求**

```http
GET /sessions/messages?session_code=sess_abc123&limit=50&offset=0
```

**查询参数**

| 参数 | 类型 | 必需 | 说明 |
| --- | --- | --- | --- |
| `session_code` | string | 是 | 会话编码 |
| `limit` | int | 否 | 分页限制（默认50） |
| `offset` | int | 否 | 分页偏移（默认0） |
| `role` | string | 否 | 过滤角色：user/assistant/tool/all |
| `order` | string | 否 | 排序顺序：asc/desc（默认desc） |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "messages": [
      {
        "id": "msg_001",
        "sessionCode": "sess_abc123",
        "role": "user",
        "content": "最近7天的登录用户数是多少？",
        "createdAt": "2023-10-22T10:30:00Z",
        "isNeedConfirm": false
      },
      {
        "id": "msg_002",
        "sessionCode": "sess_abc123",
        "role": "assistant",
        "content": "按 user_id 去重统计最近7天的登录会员人数。",
        "createdAt": "2023-10-22T10:30:05Z",
        "isNeedConfirm": true
      },
      {
        "id": "msg_003",
        "sessionCode": "sess_abc123",
        "role": "user",
        "content": "确认",
        "createdAt": "2023-10-22T10:30:10Z",
        "isNeedConfirm": false
      }
    ],
    "total": 3
  }
}
```

**响应字段说明**

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | string | 消息 ID |
| `sessionCode` | string | 所属会话编码 |
| `role` | string | 消息角色：user/assistant |
| `content` | string | 消息内容（对于 assistant 是 answer 字段） |
| `createdAt` | string | 创建时间（ISO 8601 格式） |
| `isNeedConfirm` | boolean | 是否需要用户确认（通常表示 AI 回复中包含 SQL 语句） |

### 2. 发送消息

**使用场景**

用户在聊天界面输入消息时，发送消息到服务器处理，服务器调用 AI 进行分析后直接返回结果。

**请求**

```http
POST /sessions/messages
Content-Type: application/json

{
  "session_code": "sess_abc123",
  "content": "请分析最近10天的登录数据",
  "content_type": "text",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "metadata": "optional_metadata"
}
```

**请求体**

| 字段 | 类型 | 必需 | 说明 |
| --- | --- | --- | --- |
| `session_code` | string | 是 | 会话编码 |
| `content` | string | 是 | 消息内容 |
| `content_type` | string | 否 | 内容类型：text/json/markdown（默认 text） |
| `request_id` | string | 是 | 请求唯一标识，用于幂等性控制（建议使用 UUID v4） |
| `metadata` | string | 否 | 附加元数据（JSON 字符串） |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 123,
    "session_id": 1,
    "sequence": 2,
    "is_need_confirm": true,
    "role": "assistant",
    "content": "按 user_id 去重统计最近10天的登录会员人数。",
    "content_type": "text",
    "request_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "completed",
    "model_name": "gpt-4",
    "input_tokens": 256,
    "output_tokens": 128,
    "total_tokens": 384,
    "cost_usd": 0.0032,
    "latency_ms": 2500,
    "tool_calls": null,
    "tool_results": null,
    "created_at": 1698100005000
  }
}
```

**响应字段说明**

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 消息数据库 ID |
| `session_id` | int64 | 所属会话 ID |
| `sequence` | int32 | 该会话中的消息序号 |
| `is_need_confirm` | boolean | 是否需要用户确认（通常表示 AI 回复中包含可执行的 SQL） |
| `role` | string | 消息角色：user/assistant/tool |
| `content` | string | 消息内容（AI 的最终回答） |
| `content_type` | string | 内容类型：text/json/markdown |
| `request_id` | string | 对应的请求 ID |
| `status` | string | 消息处理状态：processing/completed/failed/cancelled |
| `model_name` | string | 使用的 AI 模型名称 |
| `input_tokens` | int32 | 输入 Token 数量 |
| `output_tokens` | int32 | 输出 Token 数量 |
| `total_tokens` | int32 | 总 Token 数量 |
| `cost_usd` | float64 | 处理成本（美元） |
| `latency_ms` | int32 | 处理耗时（毫秒） |
| `tool_calls` | array | 工具调用列表（可选） |
| `tool_results` | array | 工具执行结果（可选） |
| `created_at` | int64 | 创建时间（毫秒时间戳） |

**确认流程**

当 AI 响应包含需要执行的 SQL 时，`is_need_confirm` 会被设置为 `true`。用户点击确认按钮后：

1. 发送新的消息请求，内容为 "确认"
2. 后端检测确认关键词（确认、同意、可以、ok 等）
3. 从 AI 回复中提取并执行 SQL 语句
4. 返回执行结果作为新的消息

---

## 五、数据流示例

### 场景1：新建会话并发送消息

```
1. GET /scenarios
   ↓ 获取可用场景列表

2. POST /sessions {"scene":"数据分析"}
   ↓ 创建会话
   ← 返回 session_code

3. POST /sessions/messages {"session_code":"sess_abc123","content":"查询登录用户"}
   ↓ 服务器处理请求，调用 AI 分析
   ← 直接返回 Message 对象（包含 is_need_confirm）

4. 如果 is_need_confirm=true，显示确认按钮
   用户点击后，发送新消息 "确认"
   ↓ 后端执行 SQL 并返回结果
```

### 场景2：打开已有会话

```
1. GET /sessions
   ↓ 获取会话列表

2. 点击某个会话
3. GET /sessions/messages?session_code=sess_abc123
   ↓ 加载该会话的历史消息

4. POST /sessions/messages {"session_code":"sess_abc123","content":"新消息"}
   ↓ 发送新消息并处理
   ← 直接返回处理结果
```

### 场景3：确认 AI 建议的 SQL

```
1. 用户看到 is_need_confirm=true 的消息

2. 点击"确认"按钮
3. POST /sessions/messages {"session_code":"sess_abc123","content":"确认"}
   ↓ 后端识别确认指令，提取并执行 SQL
   ← 返回 SQL 执行结果消息

4. 用户看到最新的执行结果
```

---

## 六、常见错误代码

| 代码 | 说明 |
| --- | --- |
| 0 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权（缺少 X-User-ID） |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

## 七、注意事项

1. **request_id 的重要性**: 用于追踪和幂等性控制，建议使用 UUID v4
2. **消息处理耗时**: `latency_ms` 表示 AI 处理时间，可能需要 1-5 秒，客户端应展示加载状态
3. **is_need_confirm 标志**:
   - `true` 表示 AI 响应中包含可执行的 SQL，需要用户确认
   - 用户点击确认后，后端将执行 SQL 并返回结果
4. **确认关键词**: 后端识别以下词汇作为确认：确认、同意、可以、ok、没问题、好的、是的、对的、采用、使用、行
5. **错误处理**: 如果请求失败，HTTP 状态码为非 200，同时返回的 JSON 中 `code` 不为 0，应展示 `message` 中的错误信息
6. **时间戳说明**:
   - `created_at`/`updated_at` 为毫秒级 Unix 时间戳
   - 客户端需要转换为本地时间显示
7. **消息状态流转**:
   - `processing`: 消息正在处理中
   - `completed`: 消息处理成功
   - `failed`: 消息处理失败
   - `cancelled`: 消息已取消
8. **Token 和成本**:
   - `input_tokens`: 输入提示词的 token 数
   - `output_tokens`: AI 生成内容的 token 数
   - `cost_usd`: 本条消息的成本（用于成本统计）

---

## 八、知识库管理接口

前缀：`/knowledge`

### 1. 创建知识库文档

**使用场景**

将文档内容导入到知识库系统，系统会自动分块和生成向量用于相似度搜索。

**请求**

```http
POST /knowledge/documents
Content-Type: application/json

{
  "title": "数据库表结构说明",
  "content": "用户表(user)包含以下字段...",
  "document_type": "TABLE_SCHEMA",
  "source_table": "user",
  "description": "用户表的完整说明文档",
  "tags": ["database", "user"],
  "metadata": "{\"version\": \"1.0\"}"
}
```

**请求体**

| 字段 | 类型 | 必需 | 说明 |
| --- | --- | --- | --- |
| `title` | string | 是 | 文档标题，不超过255字符 |
| `content` | string | 是 | 文档内容，将被自动切分 |
| `document_type` | string | 是 | 文档类型：TABLE_SCHEMA/METRIC_DEFINITION/SQL_EXAMPLE/BUSINESS_RULE/KNOWLEDGE |
| `source_table` | string | 否 | 关联的源表名 |
| `description` | string | 否 | 文档描述 |
| `tags` | array | 否 | 文档标签 |
| `metadata` | string | 否 | 自定义元数据（JSON）|

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "document_id": 1001,
    "document_code": "doc_123_1698000000000",
    "title": "数据库表结构说明",
    "chunk_count": 5,
    "status": "PROCESSING",
    "embed_model": "text-embedding-3-small",
    "embed_dim": 1536,
    "created_at": 1698000000000
  }
}
```

**响应字段说明**

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `document_id` | int64 | 文档唯一标识 |
| `document_code` | string | 文档编码 |
| `title` | string | 文档标题 |
| `chunk_count` | int | 自动分块后的块数 |
| `status` | string | 处理状态：PROCESSING/COMPLETED/FAILED |
| `embed_model` | string | 使用的向量模型 |
| `embed_dim` | int | 向量维度 |
| `created_at` | int64 | 创建时间戳（毫秒） |

### 2. 查询知识库文档列表

**使用场景**

分页查询知识库中的所有文档，支持多条件过滤和排序。

**请求**

```http
GET /knowledge/documents?document_type=TABLE_SCHEMA&keyword=用户&limit=20&offset=0&sort=created_at&order=desc
```

**查询参数**

| 参数 | 类型 | 必需 | 说明 |
| --- | --- | --- | --- |
| `document_type` | string | 否 | 过滤文档类型 |
| `source_table` | string | 否 | 过滤源表 |
| `status` | string | 否 | 过滤状态：DRAFT/PUBLISHED/ARCHIVED |
| `keyword` | string | 否 | 搜索关键词（在标题或内容中搜索） |
| `limit` | int | 否 | 分页限制（默认20，最大100） |
| `offset` | int | 否 | 分页偏移（默认0） |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "documents": [
      {
        "id": 1001,
        "document_code": "doc_123_1698000000000",
        "title": "数据库表结构说明",
        "content": "用户表(user)包含以下字段...",
        "document_type": "TABLE_SCHEMA",
        "source_table": "user",
        "description": "用户表的完整说明文档",
        "content_hash": "sha256_hash_value",
        "version": 1,
        "status": "PUBLISHED",
        "chunk_count": 5,
        "embed_model": "text-embedding-3-small",
        "embed_dim": 1536,
        "created_at": 1698000000000,
        "updated_at": 1698000000000
      }
    ],
    "total": 100,
    "limit": 20,
    "offset": 0
  }
}
```

### 3. 删除知识库文档

**使用场景**

删除不再使用的知识库文档，同时清理关联的向量分块数据。

**请求**

```http
DELETE /knowledge/documents?id=1001
```

**查询参数**

| 参数 | 类型 | 必需 | 说明 |
| --- | --- | --- | --- |
| `id` | string | 是 | 文档ID |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "document_id": 1001,
    "deleted": true,
    "message": "Document deleted successfully"
  }
}
```

**响应字段说明**

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `document_id` | int64 | 被删除的文档ID |
| `deleted` | boolean | 是否删除成功 |
| `message` | string | 删除信息/错误信息 |

### 文档类型说明

| 类型 | 说明 | 示例 |
| --- | --- | --- |
| `TABLE_SCHEMA` | 数据表结构 | 表定义、字段说明、约束条件 |
| `METRIC_DEFINITION` | 指标定义 | KPI、业务指标计算公式 |
| `SQL_EXAMPLE` | SQL示例 | 常用查询、临时计算 SQL |
| `BUSINESS_RULE` | 业务规则 | 业务流程、决策规则、约束 |
| `KNOWLEDGE` | 通用知识 | 其他类型的参考资料 |
