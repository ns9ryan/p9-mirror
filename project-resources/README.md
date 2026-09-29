# project-resources

P9 项目公共资源目录，用于统一保存数据库设计、开发规范、API 文档、Docker 本地环境及方案验证资料。

## 目录说明

```text
project-resources
├── database/    # 数据库 SQL 和 Ent Schema
├── demos/       # Demo、方案验证和选型示例
├── docker/      # 本地 Docker 基础环境和配置
└── docs/        # 开发规范、API 文档及项目说明
```

## 项目文档

### 开发规范与配置

- [代码规范](docs/code-norm.md)
- [服务端口](docs/service-ports.md)
- [Git SSH 配置](docs/git-ssh.md)

### API 文档

- [Core API 文档](docs/core-api.md)
- [Platform Base API 文档](docs/platform-base-api.md)
- [Platform Operator API 文档](docs/platform-operator-api.md)
- [Node Dispatch API 文档](docs/node-dispatch-api.md)
- [Integration API 文档](docs/integration-api.md)
- [OSS API 文档](docs/oss-api.md)
- [Operator Game API 文档](docs/operator-game-api.md)

## 说明

`project-resources` 用于保存 P9 项目的公共开发资料和参考资源，不承载正式业务服务代码。

`demos/` 中的内容主要用于方案讨论、技术验证和选型参考，不作为正式业务实现依据。