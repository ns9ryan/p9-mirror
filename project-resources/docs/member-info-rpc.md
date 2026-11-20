# 会员基本信息与验证状态 RPC

当前仅提供 CRM Member 服务的两个 RPC，GetBasicInfo，GetVerificationStatus；调用方需更新生成代码/依赖。

协议：proto/crm.proto、proto/types/member.proto。两个方法统一接收 GetMemberContextRequest，唯一业务参数为 member_id（正整数）。

## 返回字段

### GetBasicInfo → GetMemberBasicInfoResponse

| 字段 | 含义 |
|---|---|
| auto_level | 当前自动层级：id、name、min_deposit、min_water、created_at、updated_at |
| fixed_levels | 通过 member_fixed_level_assignment 查询的全部固定层级，按层级ID升序；每项包含id、name、description、可空withdrawal_limit、created_at、updated_at |
| agent_id | member.agent_id，所属代理 |
| agent_account_id | member.agent_account_id，可空，会员自身代理账号 |
| inviter_member_id | member.inviter_member_id，可空，邀请人会员 |
| register_ip | member.register_ip，可空 |
| register_device_code | member.register_device_code，可空 |

时间戳为Unix毫秒；没有固定层级时为空列表。可空字段使用proto optional，Go中为指针。

### GetVerificationStatus → GetMemberVerificationStatusResponse

| 字段 | 判断方式 |
|---|---|
| has_email | member_auth.email_ciphertext 非NULL且非空 |
| has_phone | member_auth.phone_ciphertext 非NULL且非空 |
| has_withdrawal_password | member_auth.withdrawal_password_hash 非NULL且非空 |
| register_ip | member.register_ip，可空 |
| register_device_code | member.register_device_code，可空 |

以上状态供上游判断权限条件，不返回密文、联系方式或密码哈希。当前按字段是否设置判断，不检查验证码验证记录；若业务要求证明验证码验证成功，需保证联系方式仅在验证成功后写入。会员认证记录不存在时返回错误，不伪装成全部false。

## 本地手动联调

在 operator-crm 仓库根目录使用 PowerShell，先启动配置所需的数据库、Redis、Etcd及相关依赖，再启动CRM（已有实例时需重启加载新代码）：

```powershell
go run ./service/crm/rpc -f ./service/crm/rpc/etc/crm.yaml
```

默认配置监听127.0.0.1:39002，实际地址以所用配置为准。以下使用已安装的grpcurl，并显式加载本地proto，不依赖服务反射。

把OPERATOR_CODE替换成目标会员所属分站编码，1替换成有效运营用户ID，8替换成实际会员ID。PowerShell使用管道传入JSON，-d '@'表示从标准输入读取，避免外部命令传参时丢失双引号。RPC方法名不能单独作为PowerShell命令执行。

第一个接口GetBasicInfo：

```powershell
'{"member_id":8}' | grpcurl -plaintext -emit-defaults -import-path ./service/crm/rpc/proto -proto crm.proto -H 'x-user-id: 1' -H 'x-user-type: operator' -H 'x-operator-code: OP_D4C8CD' -d '@' 127.0.0.1:39002 crm.Member/GetBasicInfo
```

第二个接口GetVerificationStatus：

```powershell
'{"member_id":8}' | grpcurl -plaintext -emit-defaults -import-path ./service/crm/rpc/proto -proto crm.proto -H 'x-user-id: 1' -H 'x-user-type: operator' -H 'x-operator-code: OP_D4C8CD' -d '@' 127.0.0.1:39002 crm.Member/GetVerificationStatus
```

grpcurl输出遵循ProtoJSON：字段通常显示为camelCase，int64通常显示为字符串；false、空列表等默认值可能省略，不表示查询失败。需要显示默认字段时可使用grpcurl的-emit-defaults选项。

线上服务调用须沿用现有可信RPC身份透传机制；Go调用方可使用common/ctxdata.WithClaims透传实际身份，不能仅传member_id省略分站上下文。

## 手动验收场景

1. 查询正常会员：基本信息的三个归属ID、注册IP和设备标识应与member记录一致，自动层级应与auto_level_id一致。
2. 会员关联两个或更多固定层级：fixed_levels应完整返回，与关联表逐项对应且按ID排序；没有关联时为空列表。
3. 分别选择邮箱/手机号/提款密码字段为空和非空的测试会员：验证响应应逐项符合上表，不包含任何密文或哈希。
4. 注册来源或两个可空归属ID为NULL时：对应optional字段应保持未设置。
5. member_id为0、负数、不存在ID，或切换到其他分站编码查询同一会员：应返回错误。
6. 删除身份metadata后调用：应返回错误。

只在测试环境准备数据；这些RPC均为只读。

## 自动检查

```powershell
go test ./service/crm/rpc/...
go test -v ./service/crm/rpc/internal/dao/member -run 'TestContextMemberValidation|TestMemberLevelReplacement'
```

TestContextMemberValidation不依赖数据库，覆盖缺失身份和非法ID。TestMemberLevelReplacement中调用testContextQueries，覆盖基本信息关联加载、注册来源、空固定层级、跨分站和会员不存在；它是DAO集成测试，不等同于端到端RPC联调，也未覆盖全部验证状态组合。

数据库集成测试需显式提供独立测试数据库连接：

```powershell
$env:CRM_TEST_POSTGRES_DSN = 'postgres://TEST_USER:TEST_PASSWORD@127.0.0.1:15432/TEST_DATABASE?sslmode=disable'
go test -v ./service/crm/rpc/internal/dao/member -run TestMemberLevelReplacement -count=1
Remove-Item Env:CRM_TEST_POSTGRES_DSN
```

测试会创建临时schema并在结束后清理；账号需有对应权限，测试数据库需支持citext。未设置连接时显示SKIP，不能视为数据库联调通过。
