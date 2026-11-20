# 会员管理接口

接口需要后台管理员 Bearer Token、有效分站运营身份及对应接口权限。会员和目标层级必须属于当前分站。

## 修改自动层级

`POST http://127.0.0.1:38010/member/level/auto/update`

```json
{"member_id":8,"auto_level_id":25}
```

更新会员的 auto_level_id。示例 ID 需按实际数据替换。

## 全量替换固定层级

`POST http://127.0.0.1:38010/member/level/fixed/update`

```json
{"member_id":8,"fixed_level_ids":[2,3]}
```

传空数组清空。最多100项，ID必须为正数，重复项去重。事务内校验目标层级，依赖 RequiredOperatorCodeInterceptor 隔离当前分站，然后删除原 member_fixed_level_assignment 并使用 CreateBulk 批量插入新列表，失败回滚。不显式加行锁；普通事务不保证并发全量替换的串行语义。

两接口成功时业务 data 为空对象。不存在或跨分站的会员、层级返回参数错误。

部署需同时重启 CRM RPC 和 Admin API；Admin API 注册新接口目录后，为普通管理员角色授予相应接口权限。

Nginx入口按项目现有规则加 /admin/operator 前缀；本地直连38010不添加前缀。


## 修改会员状态（冻结、解冻、禁用、恢复）

`POST /member/status/update`

```json
{"member_id":8,"status":2}
```

status：1正常（解冻或恢复），2冻结，3禁用。只修改 member.status，不隐式改变禁言、投注或提款开关。

## 重置认证错误次数

`POST /member/auth/errors/reset`

```json
{"member_id":8,"field":"sms_code_error_count"}
```

field 必须是以下之一，每次只将指定计数清零，其他计数不变：

| field | 含义 |
|---|---|
| sms_code_error_count | 短信验证码错误次数 |
| email_code_error_count | 邮箱验证码错误次数 |
| login_password_error_count | 登录密码错误次数 |
| withdrawal_password_error_count | 提款密码错误次数 |

此接口不修改密码、不发送验证码。member_auth 记录不存在时返回参数错误。

## 修改标签/备注

`POST /member/remark/update`

```json
{"member_id":8,"remark":"重点关注会员"}
```

写入 member_profile.remark，最多200个字符，空字符串清空备注。标签按备注文本存储，不另建标签关系。必须传 remark；member_profile 记录不存在时返回参数错误。

## 禁止发言、投注、提款

| 方法 | 路径 | 修改字段 |
|---|---|---|
| POST | /member/chat/update | member.chat_disabled |
| POST | /member/betting/update | member.betting_disabled |
| POST | /member/withdrawal/update | member.withdrawal_disabled |

三个接口使用相同请求：

```json
{"member_id":8,"disabled":true}
```

true 禁止，false 解除禁止。disabled 必填，不能用字符串代替布尔值。每个接口只修改对应字段。

以上全部接口直连基址为 http://127.0.0.1:38010，统一需要分站后台身份和各自接口权限。会员必须属于当前分站

## 修改会员联系方式与头像资源

新增接口及数据库迁移见 [member-contact-update.md](member-contact-update.md)。
