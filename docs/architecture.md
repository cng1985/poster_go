# 架构

海报工坊是一个 Go 模块化单体，前端由 Vue 3 构建后嵌入二进制。

## 模块

| 模块 | 能力 | 数据 |
| --- | --- | --- |
| account | 注册、登录、会话、显示名 | `account`、`session` |
| studio/template | 系统海报模块，只读 | `poster_template` |
| studio/asset | 上传图片 | `asset` 与 `uploads/` |
| studio/poster | 用户海报与编辑稿 | `poster` |
| studio/generation | 大模型或本地排版生成 | `poster_generation` |
| studio/analysis | 规则分析，模型只补充一句话 | `poster_analysis` |

叶子模块都是 `domain.go`、`resource.go`、`controller.go`、`module.go`，有数据库的再加上 `entity.go`、`repository.go`、`convert.go`。没有单独的 Service 层。公开业务方法都是一个 Request 或 Search 进，一个 `Response` 或 `PageResponse` 出。

账号不是海报子模块。身份、会话令牌和密码哈希放在 `platform/auth`，账号业务放在 `modules/account`，中间件只依赖 `auth.Resolver`。

## 单租户

Scope 只有 `ActorID`、`SessionID`、`RequestID`。没有租户字段。列表、详情、修改和删除都按登录用户过滤；模板对所有登录用户只读。客户端不能提交所有者。

首版没有角色表。登录用户就是创作者。以后如果要分管理员，应加在认证模块，而不是只靠前端隐藏按钮。

## 海报稿

图层保存在海报的 `document` JSON 里，保存时整稿替换，并用 `version` 做乐观锁。图层没有独立表，因为它们没有单独的生命周期。坐标用浮点数，这不是金额。

生成记录和海报分两次写入。海报保存失败时，生成记录仍然保留，响应会说明作品没写上。删除海报不删除历史分析和生成记录。

模型调用在数据库事务之外。文件先写入磁盘，数据库插入失败再删文件；数据库删除成功后文件删除失败只留下孤儿文件，不回滚已经提交的记录。

## 模型

`platform/ai` 只调用配置里的 OpenAI 兼容 `chat/completions`。用户不能指定供应商地址。密钥只来自 `APP_AI_API_KEY`。

没有密钥且允许本地回退时：

- 生成使用内置版式，响应里的 `provider` 为 `local`。
- 分析只跑对比度、溢出、层次和重叠规则，分数不交给模型改写。

模型失败在允许回退时改为本地排版；不允许回退时返回 503。分析即使模型失败也返回规则结果。

## 已记录的覆盖

| 项 | 决定 | 原因 |
| --- | --- | --- |
| `server.request_timeout` 为 90s | 高于规范示例的 30s | 同步生成要包住 45s 模型超时 |
| 不设置 HTTP `WriteTimeout` | 只设置请求头超时和空闲超时 | 避免上传和生成被服务级写超时切断 |
| 模板名称和海报标题使用转义后的包含匹配 | 规范样板示例是精确匹配 | 画廊需要按关键词找；只在 SQLite 验证了 `ESCAPE` |
| 分页 0 视为缺省 | 负数才算非法 | JSON 无法区分缺字段和 0 |
| 生产配置不把 SQLite 连接池改成 10 | 规范提醒网络库不要继承单连接 | 本发行默认仍是 SQLite，网络库切换时再改池并验证 |
