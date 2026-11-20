# 会员联系方式及头像资源接口

统一入口：`/admin/operator`；本地直连 Admin API 38010 不加此前缀。需要后台管理员 Bearer Token、有效分站运营身份及接口权限。会员只能属于当前分站。

## 修改联系方式

`POST /member/contact/update`

```json
{"member_id":8,"avatar_id":1,"avatar_frame_id":2,"nickname":"会员昵称","real_name":"张三","phone":"8613800000000","identity_number":"12345678901","email":"member@example.com","phone_verification":{"record_id":100,"verification_code":"012345"},"email_verification":{"record_id":101,"verification_code":"678901"}}
```

除 member_id 外均可选，至少传一个修改字段；省略或 null 保持原值。昵称去首尾空白后不能为空，最多100字符。真实姓名及身份ID/CPF最多100字符，空字符串清空。手机号需含国际区号，去空格及前导 + 后为 7 至 15 位数字。邮箱去首尾空白并转小写，最多 254 字符。修改时先向新的手机号或邮箱申请验证码，并在更新请求中传入对应凭证。空字符串不支持解除绑定。发送方式见 [会员管理接口文档](operator-member-api.md#发送验证码与确认修改)。同一分站的手机号、邮箱不可重复。CPF按身份号码文本存储，不做国家专属校验。

头像/头像框必须传资源接口返回的正整数ID，不能传URL，不能选择尚未配置图片的资源。资料写入 member_profile，手机号/邮箱写入 member_auth；真实姓名、身份ID、手机号、邮箱均使用公共 fieldcrypto AES-GCM 加密，手机号和邮箱另存 HMAC-SHA256 检索摘要。所有数据库修改在同一事务提交；会员、资料或需要更新的认证记录缺失返回参数错误。成功 data 为 `{}`。

## 获取头像与头像框

`GET /member/images`

业务 data：
```json
{"avatars":[{"id":1,"image_url":"https://cdn.example.com/avatar.png"}],"avatar_frames":[{"id":2,"image_url":"https://cdn.example.com/frame.png"}]}
```

Admin API 调 CRM，CRM 调 platform-base 的 DefaultAvatarService.ListAll 和 DefaultAvatarFrameService.ListAll。只返回已配置URL的资源；空列表返回 []。总站不可用时返回错误，不使用硬编码URL。总站图片更新保持ID不变，前端通过本接口获取最新URL。

会员详情响应改为 avatar_id、avatar_frame_id，前端按资源列表解析图片；0表示历史未设置。数据库不再保存URL。新建会员初始ID为0，后台可通过本接口选择资源。

## 配置与部署

CRM新增配置示例（密钥通过部署系统注入，不提交真实值）：
```yaml
PlatformBaseRpc:
  Endpoints:
    - 127.0.0.1:平台RPC端口
FieldCrypto:
  EncryptionKey: "替换为16或24或32字节的加密密钥"
  LookupKey: "替换为非空HMAC密钥"
```

密钥必须与已有会员敏感字段使用的密钥保持一致；不可随机生成替代已有密钥。未配置加密时非空敏感字段更新报错。未配置总站RPC时资源查询和头像修改报错。

先停止CRM，备份 member_profile，执行同目录 member-profile-image-ids.sql；直接新增ID字段并删除旧URL字段，不迁移旧URL，新增ID默认0（未设置）。确认完成后重启CRM与Admin API，给普通管理员授予新增接口权限。CRM启动包含自动删列迁移，必须先执行数据迁移。前后端及RPC协议需要同步更新。
