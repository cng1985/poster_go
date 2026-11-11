# 海报工坊

海报工坊是一个按《Go 项目结构与开发标准规范》v3.4 组织的模块化单体。用户可以注册登录，从系统海报模块起步，上传图片，在编辑器里排版，也可以让大模型或本地排版生成海报，并分析自己的作品。

## 启动

需要 Go 1.26 和 Node.js 22。

```bash
cd web && npm ci && npm run build && cd ..
go run ./cmd/server --profile dev
```

浏览器打开 `http://127.0.0.1:8080`。开发时也可以只启动后端，再在 `web/` 执行 `npm run dev`，Vite 会把 `/api` 代理到 8080。

```bash
go run ./cmd/server --version
```

数据默认写到 `./data/app.db`，上传文件在 `./data/uploads`。这两个目录不进入版本库。

## 配置

`configs/application.yaml`、`application-dev.yaml`、`application-pro.yaml` 放在同一目录。启动参数 `--profile` 或 `APP_PROFILE` 只能是 `dev` 或 `pro`；不指定时只加载基础配置。

常用覆盖：

| 变量 | 作用 |
| --- | --- |
| `APP_DATA_DIR` | 数据目录 |
| `APP_SERVER_HOST` / `APP_SERVER_PORT` | 监听地址 |
| `APP_DATABASE_DSN` | MySQL、PostgreSQL、SQL Server 的连接串 |
| `APP_AI_API_KEY` | OpenAI 兼容接口的密钥，不写进 YAML |

普通请求超时是 90 秒，模型超时是 45 秒。同步生成占用同一次 HTTP 请求，所以请求预算高于模型超时。没有设置 `ReadTimeout` 和 `WriteTimeout`，避免长请求被整个服务的写超时切断。

未配置密钥时，开发配置会使用本地排版和分析规则。生产配置 `ai.fallback_local: false`，这时没有密钥会明确返回“未配置大模型”。

## 数据库

默认 SQLite，驱动是纯 Go 的 `github.com/glebarez/sqlite`，`CGO_ENABLED=0` 可以交叉编译。连接上会核对外键、`journal_mode` 和 `busy_timeout`。

MySQL、PostgreSQL、SQL Server 的驱动已接入，但本发行只携带已验证的 SQLite 迁移。选择其他驱动时进程会拒绝启动，而不是假装兼容。网络库的连接池不要沿用 SQLite 的单连接，切换前需要单独的迁移和验证。

## 构建与安装

```bash
pwsh scripts/build-platforms.ps1
```

Windows 入口是 `scripts/build-platforms.cmd`。发布目录在 `release/<os>-<arch>/`，包含二进制、配置模板、Linux 安装脚本和版本说明。

```bash
sudo sh scripts/install-linux.sh
```

安装到 `/opt/poster`，数据在 `/var/lib/poster`，并由 systemd 以 `poster` 用户、`--profile pro` 运行。升级会保留已有配置和数据，并留下 `poster.prev`。

```bash
systemctl status poster
journalctl -u poster
```

## 检查

```bash
go test ./...
```

健康检查：

- `GET /health/live`
- `GET /health/ready`

更多说明见 `docs/`。
