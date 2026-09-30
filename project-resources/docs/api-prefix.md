# P9 API 路径前缀约定

P9 HTTP API 对外统一使用 `/api` 作为 API 入口前缀。

业务域和访问端由域名区分，具体后端服务由 API 路径中的服务前缀区分。

## 基本规则

对外访问路径：

```text
/api/{service-prefix}/...
```

go-zero API 服务内部实际路由：

```text
/{service-prefix}/...
```

Nginx 对外接收 `/api/...` 请求后，去掉最前面的 `/api`，再转发到对应 API 服务。

例如：

```text
浏览器请求：
/api/core/login

Nginx 转发：
/core/login

core-api 实际路由：
/core/login
```

`/api` 只属于统一 Web 入口，不需要写入各后端 API 服务自身的 go-zero 路由。

---

## 业务域与服务边界

P9 使用以下原则区分业务域和服务：

```text
Host
→ 当前属于哪个业务域、哪个 Operator

Path
→ 当前调用哪个后端服务
```

例如：

```text
https://platform.example.com/api/core/login
```

表示：

```text
platform.example.com
→ 总网

/api/core
→ core-api
```

而：

```text
https://op001.example.com/api/core/login
```

表示：

```text
op001.example.com
→ OP_001 分站

/api/core
→ 当前分站节点的 core-api
```

总网和分站可以使用相同的 `/api/core` 路径，因为具体业务环境由访问域名区分。

不在 API Path 中重复增加：

```text
/platform
/operator
/admin
```

等用于表达访问端或部署域的前缀。

---

## `/admin` 前缀

P9 后台 API 不再统一使用 `/admin` 前缀。

原有：

```text
/admin/login
/admin/user/info
/admin/operator/list
/admin/game/list
```

逐步调整为按服务划分：

```text
/core/login
/core/user/info

/operator/list

/game/list
```

是否属于：

```text
总网后台
分站后台
代理后台
会员端
```

由访问域名和认证上下文决定，不再通过 `/admin` 表达。

---

## 当前服务前缀

| 服务 | 后端路由前缀 | 对外路由前缀 |
|---|---|---|
| core | `/core` | `/api/core` |
| platform-base | `/base` | `/api/base` |
| operator-base | `/base` | `/api/base` |
| platform-operator | `/operator` | `/api/operator` |
| platform-game | `/game` | `/api/game` |
| operator-game | `/game` | `/api/game` |

同一个服务前缀可以在不同业务域复用。

例如 `platform-base` 和 `operator-base` 都使用：

```text
/base
```

但完整访问地址不同：

```text
总网：
https://platform.example.com/api/base/...

分站：
https://op001.example.com/api/base/...
```

同理，`platform-game` 和 `operator-game` 都使用：

```text
/game
```

通过 Host 区分实际调用的是总网游戏服务还是厅游戏服务。

---

## Core

Core 统一使用：

```text
/core
```

例如：

```text
/core/login
/core/refresh
/core/logout
/core/user/info
/core/user/perm
/core/role/list
```

对外：

```text
/api/core/login
/api/core/refresh
/api/core/logout
/api/core/user/info
/api/core/user/perm
/api/core/role/list
```

总网和分站共用 Core 代码时，保持相同 HTTP Contract。

例如：

```text
总网：
https://platform.example.com/api/core/login

分站：
https://op001.example.com/api/core/login
```

两者由 Host 区分业务环境，不需要分别设计：

```text
/api/platform/core
/api/operator/core
```

---

## Base

基础业务服务统一使用：

```text
/base
```

总网 `platform-base` 例如：

```text
/base/timezone/list
/base/currency/list
/base/region/list
```

对外：

```text
/api/base/timezone/list
/api/base/currency/list
/api/base/region/list
```

分站 `operator-base` 后续业务接口同样使用：

```text
/base/...
```

例如：

```text
/base/operator/info
```

对外：

```text
/api/base/operator/info
```

---

## Platform Operator

总网分站管理服务 `platform-operator` 使用：

```text
/operator
```

例如：

```text
/operator/list
/operator/detail
/operator/profile/...
/operator/domain/...
/operator/language-allocation/...
/operator/region-allocation/...
```

对外：

```text
/api/operator/list
/api/operator/detail
/api/operator/profile/...
/api/operator/domain/...
```

这里的 `/operator` 表示总网的分站管理业务服务，不表示“当前请求属于 Operator 部署域”。

---

## Game

游戏服务统一使用：

```text
/game
```

建议资源继续按层级组织。

例如：

```text
/game/list
/game/get
/game/update

/game/category/list
/game/category/get
/game/category/update

/game/provider/list
/game/provider/get
/game/provider/update

/game/channel/list
/game/channel/get
/game/channel/update
```

对外：

```text
/api/game/list
/api/game/category/list
/api/game/provider/list
/api/game/channel/list
```

总网和分站可以保持一致：

```text
总网：
https://platform.example.com/api/game/list

分站：
https://op001.example.com/api/game/list
```

分别由 Nginx 转发到：

```text
platform-game-api
operator-game-api
```

---

## Nginx 路由

分站后台域名例如：

```text
op001.example.com
```

Nginx 对外统一处理：

```text
/
→ operator-admin-web

/api/core/*
→ core-api

/api/base/*
→ operator-base-api

/api/game/*
→ operator-game-api
```

并根据访问域名注入：

```text
X-Operator-Code
```

例如：

```text
Host: op001.example.com
X-Operator-Code: OP_001
```

前端不传 `operator_code`，也不自行设置可信的 `X-Operator-Code`。

---

## 前端调用

正式部署时，前端统一使用相对路径：

```text
/api/...
```

例如：

```text
/api/core/login
/api/core/user/info
/api/base/operator/info
/api/game/list
```

前端不需要知道：

```text
Node IP
API 端口
operator_code
具体后端容器地址
```

Nginx 负责根据：

```text
Host
+
API 服务前缀
```

完成实际转发。

本地 Vite 开发时，同样请求 `/api/...`，通过 Vite Dev Proxy 将 `/api` 转发到需要联调的 Operator 域名。

---

## 内部接口

服务内部接口不使用对外 `/api` 规则。

例如当前调度中心回调总网：

```text
/internal/dispatch/...
```

继续作为内部接口保留。

不调整为：

```text
/operator/internal/...
```

也不通过：

```text
/api/operator/...
```

对外暴露。

统一原则：

```text
前端业务 API
→ /{service-prefix}/...

服务内部 HTTP 接口
→ /internal/...
```

---

## 健康检查

服务健康检查接口不强制套用业务服务前缀。

例如：

```text
/ping
```

主要用于服务自身检查或内部部署探测。

不要求统一改成：

```text
/core/ping
/base/ping
/game/ping
```

也不需要默认通过 Nginx 对外暴露：

```text
/api/core/ping
/api/base/ping
/api/game/ping
```

---

## 当前规则总结

P9 HTTP 路径统一按照以下职责划分：

```text
域名
→ 区分业务域、访问端和 Operator

/api
→ Nginx 对外统一 API 入口

/core
/base
/operator
/game
→ 区分后端业务服务

/internal
→ 服务内部 HTTP 接口

/ping
→ 服务自身健康检查
```

例如：

```text
https://op001.example.com/api/game/category/list
```

含义为：

```text
op001.example.com
→ OP_001 分站

/api
→ 后端 API

/game
→ operator-game-api

/category/list
→ 游戏分类列表
```