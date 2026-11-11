# PosterGo

海报编辑器。左侧可以管理图层、添加组件、从图片库选图，也可以套用模板。大模型按 OpenAI Chat Completions 协议配置，密钥只保存在本机。

## 运行

在仓库根目录执行：

```bash
go run ./cmd/poster
```

浏览器打开 http://127.0.0.1:8080 。

常用参数：

```bash
go run ./cmd/poster -addr :8080 -data data -web web
```

`data/` 存放海报、上传图片和模型配置，已在 `.gitignore` 中忽略。首次启动会放一张「春季上新」示例海报。

## 编辑器

- **图层**：列表从上到下是从前景到背景。可以拖动手柄排序，也可以隐藏、锁定。
- **组件**：主标题、副标题、正文、矩形、圆形、直线，以及从图片库选图。
- **图片库**：内置背景、纹理、插画和图标。也可以上传 PNG、JPEG、GIF、WEBP、SVG。点击加入画布，或拖到海报上。背景类图片会铺满并放到最底层。
- **模板**：用现成版式替换当前画面。
- 右侧检查器改尺寸、颜色、文字和叠放顺序。画布上可以拖动、缩放、旋转，双击文字直接编辑。
- **导出 PNG** 把当前海报下载成图片。

快捷键：`Ctrl/⌘ S` 保存，`Ctrl/⌘ Z` 撤销，`Ctrl/⌘ Shift Z` 重做，`Ctrl/⌘ D` 复制图层，方向键微调，`Delete` 删除。

## 配置大模型

点顶部「模型」。接口需要兼容 OpenAI 的 `POST /v1/chat/completions`。可以填 API 根地址，也可以填完整的 `/chat/completions` 地址。

| 服务 | 接口地址 | 模型示例 |
| --- | --- | --- |
| OpenAI | `https://api.openai.com/v1` | `gpt-4o-mini` |
| DeepSeek | `https://api.deepseek.com/v1` | `deepseek-chat` |
| Moonshot | `https://api.moonshot.cn/v1` | `moonshot-v1-8k` |
| 通义千问兼容模式 | `https://dashscope.aliyuncs.com/compatible-mode/v1` | `qwen-plus` |
| Ollama | `http://127.0.0.1:11434/v1` | `llama3.2` |
| LM Studio | `http://127.0.0.1:1234/v1` | 本地模型名 |

本地服务可以不填密钥。云端密钥写入 `data/llm.json`（文件权限 `0600`），保存之后接口不再把密钥返回给浏览器。测试连接和拉取模型列表时，留空密钥会继续使用已保存的那一把。所有模型请求都由本服务转发，浏览器不直连供应商。

配置并启用一个模型后，可以用「AI 生成」写文案或按描述重排当前画布。模型只能使用图片库里已有的图片地址。

## 开发

```bash
go test ./...
node --test web/js/geom.test.mjs
```

项目只使用 Go 标准库，前端是原生 ES 模块，由同一个进程提供。
