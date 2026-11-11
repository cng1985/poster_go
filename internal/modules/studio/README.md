# 海报业务

父模块只组合子模块，不承载业务实现。

| 子模块 | 职责 | 数据所有权 | 允许依赖 |
| --- | --- | --- | --- |
| template | 系统海报模块，只读 | `poster_template` | shared 文档模型 |
| asset | 用户上传的图片素材 | `asset` 与 `uploads/` 文件 | shared 不依赖；文件存储使用 platform/storage |
| poster | 用户海报作品与编辑稿 | `poster` | asset，用于校验图片图层归属 |
| generation | 按描述生成海报稿，可另存为作品 | `poster_generation` | poster、platform/ai |
| analysis | 检查作品的对比、层次和溢出 | `poster_analysis` | poster、platform/ai |

shared 只放海报文档模型和校验，不是业务模块。子模块不引用父包。

跨模块写入不是一个事务：生成记录先落库，再创建海报。海报创建失败时生成记录保留，响应里说明作品未保存。删除海报不级联删除分析或生成记录。

模型调用发生在数据库事务之外。未配置 `APP_AI_API_KEY` 时，生成使用本地排版，分析只使用规则。
