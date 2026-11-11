# 数据库

迁移在 `migrations/0001_init.sql`。`schema_migration` 记录已执行版本，重复启动不会重放，也不会改写已有模板。

| 表 | 说明 | 删除 |
| --- | --- | --- |
| account | 用户，用户名唯一 | 不提供删除 |
| session | 令牌只存 SHA-256，可吊销 | 登出只写 `revoked_at` |
| poster_template | 系统模块，`code` 唯一 | 不提供删除 |
| poster | 用户作品，`version` 乐观锁 | 硬删除 |
| poster_generation | 生成记录，`poster_id` 可空 | 不随海报删除 |
| poster_analysis | 分析记录 | 不随海报删除 |
| asset | 素材元数据，文件在数据目录 | 硬删除并尝试删除文件 |

`poster.owner_id`、`asset.owner_id`、`session.account_id` 等引用 `account.id`。SQLite 外键在每个连接的 DSN pragma 里打开，启动时会读回确认。

时间以 UTC 存储，接口使用 RFC3339。标识在接口中是字符串，库内是自增整数。

模板种子在启动时按 `code` 做 `ON CONFLICT DO NOTHING`。已存在的模块不会被重置。

## 支持矩阵

| 数据库 | 驱动 | 状态 |
| --- | --- | --- |
| SQLite | glebarez/sqlite，纯 Go | 已用迁移、外键、分页、版本冲突和注册回滚验证 |
| MySQL | gorm.io/driver/mysql | 驱动已接入，无已验证迁移，启动会拒绝 |
| PostgreSQL | gorm.io/driver/postgres | 同上 |
| SQL Server | gorm.io/driver/sqlserver | 同上 |

验证环境：Go 1.26，Linux amd64，`CGO_ENABLED=0`。没有在 Windows 或 arm64 上实际启动过二进制，发布脚本可以交叉编译，但不把未运行的目标写成已验证。
