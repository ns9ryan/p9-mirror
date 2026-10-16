# Outstation Admin（分站后台）

P9 分站管理后台，基于与 `headquarter/`（总网）相同的 Vue 3 + Vite + Arco Design Pro 技术栈。

总网通过「分站列表 → 创建分站」开通分站后，运营人员使用本项目登录对应分站后台。

## 本地开发

```bash
cd outstation
pnpm install
pnpm dev
```

默认开发地址：`http://localhost:5174`（与总网 `5173` 错开）。

## 与总网的差异（脚手架阶段）

| 项 | headquarter | outstation |
|----|-------------|------------|
| 产品名 | 总网平台 | 分站平台 |
| 开发端口 | 5173 | 5174 |
| `VITE_I18N_CODE` | `platform` | `operator` |
| 登录 | 账号密码 | 账号密码 + **分站代码** (`operator_code`) |

业务菜单由后端按分站权限下发；当前脚手架保留与总网同源的页面与接口封装，后续按分站职责裁剪（例如不需要「创建分站 / 跨站资源分配」等总网能力）。
