# 接口

普通 JSON 接口使用 `POST`。成功和业务错误的 HTTP 状态都是 200，用正文 `code` 区分。分页成功没有 `code` 和 `msg`。

单条成功：

```json
{"code":200,"msg":"success","data":{"id":"1"}}
```

分页成功：

```json
{"data":[],"total":0,"page":1,"page_size":20}
```

错误：

```json
{"code":400,"msg":"标题不能超过 80 个字"}
```

标识使用十进制字符串。创建时不要传 `id`。

## 公开

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/account/register` | 注册并返回会话 |
| POST | `/api/account/login` | 登录 |
| GET | `/health/live` | 存活，HTTP 200 |
| GET | `/health/ready` | 迁移完成且数据库可 ping，失败为 HTTP 503 |

## 登录后

请求头：`Authorization: Bearer <token>`。

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/account/logout` | 吊销当前会话，重复调用仍成功 |
| POST | `/api/account/one` | 当前用户 |
| POST | `/api/account/update` | 修改显示名 |
| POST | `/api/template/search` | 海报模块分页 |
| POST | `/api/template/one` | 模块详情，含整稿 |
| POST | `/api/poster/create` | 新建作品 |
| POST | `/api/poster/update` | 按版本整稿保存 |
| POST | `/api/poster/one` | 作品详情 |
| POST | `/api/poster/search` | 只看自己的作品 |
| POST | `/api/poster/delete` | 硬删除，再次删除返回不存在 |
| POST | `/api/poster/duplicate` | 复制作品 |
| POST | `/api/generation/create` | 生成，`save_poster` 为真时另存作品 |
| POST | `/api/generation/one` | 生成详情 |
| POST | `/api/generation/search` | 生成记录 |
| POST | `/api/analysis/create` | 分析一张自己的海报 |
| POST | `/api/analysis/one` | 分析详情 |
| POST | `/api/analysis/search` | 分析记录 |
| POST | `/api/asset/upload` | `multipart` 字段名 `file` |
| POST | `/api/asset/search` | 自己的素材 |
| POST | `/api/asset/one` | 素材详情 |
| POST | `/api/asset/delete` | 删除素材 |
| GET | `/api/asset/content/:id` | 图片字节，需要登录 |

未知的 `/api` 路径返回 `code=404`。已注册路径用错方法返回 `code=405`。这两种响应的 HTTP 状态仍是 200。缺失的静态资源返回真正的 HTTP 404，页面路径回退到 `index.html`。

## 业务码

| 码 | 含义 |
| --- | --- |
| 400 | 参数错误 |
| 401 | 未登录或会话失效 |
| 403 | 账号不可用 |
| 404 | 不存在或不属于当前用户 |
| 405 | 方法不允许 |
| 413 | 请求体或图片过大 |
| 429 | 登录失败过多，或生成/分析过于频繁 |
| 500 | 内部错误，不返回原始原因 |
| 503 | 模型不可用或处理超时 |
| 10001 | 用户名已存在 |
| 10002 | 用户名或密码错误 |
| 10004 | 海报版本冲突 |

登录失败使用 10002，不会把已有会话清掉。401 才会让前端回到登录页。

## 海报稿

```json
{
  "width": 1080,
  "height": 1440,
  "background": "#112233",
  "layers": [
    {
      "id": "t1",
      "type": "text",
      "name": "标题",
      "x": 80,
      "y": 120,
      "w": 800,
      "h": 180,
      "rotation": 0,
      "opacity": 1,
      "text": "你好",
      "font_size": 72,
      "font_weight": 700,
      "font_family": "serif",
      "color": "#ffffff",
      "align": "left"
    }
  ]
}
```

`type` 可以是 `text`、`rect`、`ellipse`、`image`。图片层的 `asset_id` 必须是当前用户的素材。生成结果不会带外链图片。

生成的 `style`：`editorial`、`sale`、`festive`、`minimal`、`bold`。`size_preset`：`story`、`poster`、`square`、`cover`。

排序白名单：`updated_at`、`created_at`、`title`，方向 `asc` 或 `desc`。
