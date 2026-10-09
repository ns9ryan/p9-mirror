# P9 服务端口

P9 默认服务端口按部署域划分。

同一服务同时提供 API 和 RPC 时，原则上使用相同尾号；独立 API 接入层和业务 RPC 分别分配端口。WebSocket 等其他通信端口按实际需要单独分配。

本文定义的是 P9 默认端口约定。当前开发环境因多个逻辑 Node 部署在同一台物理服务器上，部分 Node 使用偏移端口避免冲突；这些端口仅用于联调，不作为正式端口标准。

## 端口段

| 部署域 | HTTP / WebSocket 端口段 | RPC 端口段 |
|---|---:|---:|
| Platform | `18000-18999` | `19000-19999` |
| Node Dispatch | `28000-28999` | `29000-29999` |
| Operator | `38000-38999` | `39000-39999` |

---

## Platform

| 服务 | API | RPC |
|---|---:|---:|
| core | `18000` | `19000` |
| platform-base | `18001` | `19001` |
| platform-operator | `18002` | `19002` |
| platform-game | `18003` | `19003` |
| platform-message | `18004` | `19004` |
| integration | `18005` | `19005` |
| file | `18006` | - |

说明：

- Platform 服务默认使用 `180xx` 作为 HTTP API 端口
- Platform RPC 服务默认使用对应的 `190xx` 端口
- `file` 当前只分配 HTTP API 端口，RPC 暂不分配，后续有实际需求时再增加
- 同一服务的 API 和 RPC 使用相同尾号，例如 `platform-base` 使用 `18001 / 19001`

---

## Node Dispatch

| 服务 | API | WebSocket | RPC |
|---|---:|---:|---:|
| node-dispatch | `28001` | `28002` | `29001` |

说明：

- `28001` 用于总网后台访问 node-dispatch HTTP API
- `28002` 用于 node-agent 与 node-dispatch 建立 WebSocket 长连接
- `29001` 用于 Platform 内部服务调用 node-dispatch RPC
- node-agent 主动连接 node-dispatch WebSocket，本身当前不分配固定对外监听端口

---

## Operator

每个正式 Operator Node 独立部署时，统一使用以下默认端口。**新增规划的端口不代表服务已部署。**

| 服务 | API | RPC | 状态 |
|---|---:|---:|---|
| core | `38000` | `39000` | 已使用 |
| operator-base | `38001` | `39001` | 已使用 |
| operator-crm / crm-rpc | - | `39002` | 新增规划 |
| operator-game | `38003` | `39003` | 已使用 |
| operator-crm / settlement-rpc | - | `39004` | 新增规划 |
| operator-wallet | - | `39005` | 新增规划 |
| operator-activity | - | `39006` | 新增规划 |
| operator-operation | - | `39007` | 新增规划 |
| operator-risk | - | `39008` | 新增规划 |
| operator-report | - | `39009` | 新增规划 |
| operator-admin-api | `38010` | - | 新增规划 |
| agent-admin-api | `38011` | - | 新增规划 |
| member-h5-api | `38012` | - | 新增规划 |

说明：

- `operator-crm` 共用一个 Go Module 和一套 Ent，内部的 `crm-rpc`、`settlement-rpc` 独立部署
- 三个客户端 API 分别服务厅后台、代理后台和会员 H5，按需调用业务 RPC
- `core`、`operator-base`、`operator-game` 的既有 API 端口保持不变；后续接口调整时再决定迁移范围
- `38002` 当前未分配；`39002` 已规划给 `crm-rpc`
- 不同物理节点可以使用相同的默认端口；同机部署多个逻辑 Node 时需使用不同端口

当前已有部署方式示例（不包含新增规划服务）：

```text
node01
├── core-api              38000
├── core-rpc              39000
├── operator-base-api     38001
├── operator-base-rpc     39001
├── operator-game-api     38003
└── operator-game-rpc     39003

node02
├── core-api              38000
├── core-rpc              39000
├── operator-base-api     38001
├── operator-base-rpc     39001
├── operator-game-api     38003
└── operator-game-rpc     39003
```

---

## 当前单机联调环境

当前开发环境只有一台物理服务器：

```text
192.168.0.15
```

为了在同一台服务器上同时模拟多个 Operator Node，不同逻辑 Node 使用不同端口段避免冲突。

### node01

| 服务 | API | RPC |
|---|---:|---:|
| core | `38000` | `39000` |
| operator-base | `38001` | `39001` |
| operator-game | `38003` | `39003` |

### node02

| 服务 | API | RPC |
|---|---:|---:|
| core | `38200` | `39200` |
| operator-base | `38201` | `39201` |
| operator-game | - | - |

说明：

- node01 当前直接使用 Operator 默认端口
- node02 因与 node01 共用同一台物理服务器，当前使用 `382xx / 392xx` 避免端口冲突
- node02 当前尚未部署 operator-game
- `382xx / 392xx` 属于当前单机联调环境的临时分配，不作为正式 Operator Node 端口标准
- 后续 Operator Node 独立部署到不同服务器后，应统一恢复使用 `380xx / 390xx` 默认端口

---

## Nginx

正式 Operator Node 原则上每个节点部署一个 Nginx。

```text
Operator Node
├── Nginx
├── node-agent
├── core
├── operator-base
└── operator-game
```

Nginx 对外通常监听：

```text
HTTP   80
HTTPS  443
```

`80 / 443` 属于节点统一 Web 入口，不占用 P9 API / RPC 服务端口段。

正式环境中，请求通过域名进入对应节点的 Nginx，再由 Nginx 转发到本节点的 API 服务。

当前单机联调环境中，一个 Nginx 可以临时同时代理 node01、node02 等多个逻辑 Node，并根据访问域名转发到对应的联调端口。
