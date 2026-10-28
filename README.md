# Usagi Navigation · Sun-Panel Fork

以 [Sun-Panel v1.3.0](https://github.com/hslr-s/sun-panel/tree/v1.3.0) 的开源代码为基础，自主维护导航站。源码基线：`31fd128633feed4ac8af4d9fb2c7e92696aad20c`。保留原作者红烧猎人的 MIT 许可与署名；不包含后续闭源版本的代码。原项目介绍见 [README.upstream.md](README.upstream.md)。

## 本分支

- 在 Vue 源码中实现 Usagi Design System：奶油背景、樱粉强调、暖棕页脚、统一字体与卡片交互。
- 原生跨站页脚：导航、博客、网盘、论坛、管理；短页面贴底，长页面自然滚动。
- 兼容现有数据的私密分组、分组卡片样式及新版卡片字段。公开接口按分组可见性过滤，编辑和批量导入检查所属用户。
- 更新前后端依赖、以 `go:embed` 内嵌资源，建立源码验证与容器构建流程。版本：`1.3.0-usagi.1`。

这属于从开源版本独立演进，不能视为后续官方闭源版的完整替代。未重新实现后续版本新增的插件、卡片类型等功能。

## 开发

需要 Node.js 24、pnpm 8.15.9、Go 1.27.2。Go SQLite 驱动需要 C 编译器；容器已包含构建依赖。

```sh
pnpm install --frozen-lockfile
pnpm build
cd service
go test ./...
go run .
```

另开终端运行 `pnpm dev`。Vite 通过 `.env` 中的 `VITE_APP_API_BASE_URL` 代理 API。本机前端构建可用 `SUN_PANEL_WEB_OUT_DIR=/path/outside/sync pnpm build` 避免同步生成目录。

## 独立开发实例

不要将开发实例指向生产数据库。先从只读源建立 SQLite 一致性副本，重置开发账号密码、清空登录令牌和外部服务设置。开发凭据仅写入目标目录的 `dev-credentials.json`（权限 0600），不会输出到日志或进入 Git。

```sh
python3 scripts/create-dev-snapshot.py --source-conf /path/to/production/conf --target-conf ./conf
docker compose -f docker-compose.dev.yml up -d --build
```

开发端口仅绑定 `127.0.0.1:3003`，容器名称 `sun-panel-dev`，不参加 Watchtower 自动更新。通过 SSH 转发访问：

```sh
ssh -N -L 13003:127.0.0.1:3003 YOUR_SSH_HOST
```

浏览器打开 `http://127.0.0.1:13003`。副本保留分组、链接、图标和页面配置，但使用独立凭据；原来的 HTML 页脚在页面渲染时由原生 Usagi 页脚接替。开发配置、数据库、上传文件和凭据均被 Git 与容器构建忽略。

## UI 维护

全站统一顶栏位于 `src/components/usagi/UsagiHeader.vue`，使用与页脚同源的大地暖棕色；左侧站名（乌萨奇导航 · Usagi Nav），右侧时间与登录/管理入口，悬浮按钮只负责回到顶部。跨站五站互联生态导航集中收敛于页脚 `src/components/usagi/UsagiFooter.vue`，彻底消除顶底元素重复。界面默认中文，可在个人设置中切换英文。设计常量位于 `src/styles/usagi-tokens.css`，页面布局位于 `src/styles/usagi.css`。分组样式解析集中在 `src/utils/groupCardStyle.ts`；不要再修改构建后的 HTML 或用负边距补页脚。

应用图标统一由 `src/components/common/ItemIcon/index.vue` 渲染：使用连续圆角方形遮罩，图片铺满且不额外添加白底、边框或内边距，文字、符号和加载失败的简称使用相同外形。非方形图片按中心裁切，旧数据中设置的背景色仍生效；原图片内部自带的白边需要替换素材才能消除。

新增书签时只需粘贴 HTTP(S) 网址：名称、简介和图标全部自动识别，编辑器里只读预览，不再提供手动填写或自定义图标（文字、在线图标、图片地址/上传、背景色）。名称优先取网站声明的站点名（`og:site_name` 等），首页书签用站点名（如 GitHub），具体页面用标题中页面自己的部分（如 `requests · PyPI` → requests），最长 20 字；简介取 meta description，最长 100 字；图标优先 Apple Touch Icon 等大尺寸图标。可点“重新识别”刷新；更换网址会整体重新识别，改回原网址则恢复已保存的内容，打开没有图标的旧书签时自动补识别，网站暂时不可达时不会清空已保存的内容。浏览器端没能识别到的书签（站点暂时不可达、批量导入等）保存后由服务器在后台补全：缺图标时写入图标，名称只是域名时换成识别到的名称，简介为空时补简介；已有的图标、名称、简介不会被覆盖，同一网址 30 分钟内不重复请求。分组管理中的“重新识别图标与简介”会在后台检查全部书签：补全缺失的图标、名称和简介，并把已无法下载的图片图标、依赖在线图标库（api.iconify.design）的旧图标换成网站图标；已有的名称和简介不会被覆盖，图标正常的书签不会去请求其网站。

已移除原版的内网地址与内外网切换（编辑器“内网地址”、悬浮栏切换按钮、右键“打开内网/外网地址”、设置中的切换按钮开关），书签始终打开其网址。数据库中旧的内网地址列保留不动，但接口不再读写。

图标缓存默认存放在 `./conf/icon-cache`（可用 `[base] icon_cache_path` 修改），原图片 URL 保留在数据库中。缓存按图片内容共享，位图最长边 192px、透明度保留，通过 `/icon-cache/` 提供带长期缓存头的文件；首页只提前加载可见区域附近的图标。删除书签、删除分组或更换图标会清理失去引用的文件；共用图标及页面配置引用仍然保留，未保存的识别预览在 24 小时后定期回收。备份时需包含整个 `conf` 目录。

首页合并分组与书签读取，页面配置到齐后一次显示，加载和失败重试有明确状态。编辑器、设置、监控组件按需加载，构建预生成 Brotli/Gzip 静态资源，哈希资源长期缓存，API 响应禁止缓存并压缩较大 JSON。

加载性能：首页的三个接口在入口脚本启动时即发出（`src/main.ts` 的 `prefetch`），与页面代码下载并行；`index.html` 预加载首页代码块。首屏先渲染约 36 张卡片，其余分组在下一帧一次性补齐；拖拽排序组件只在进入排序时加载；分组样式按组解析一次；分组悬浮按钮由 CSS 控制，不再触发整页重渲染；详情卡片标题/简介用 CSS 截断而非逐卡 tooltip 组件。入口脚本移除了 axios（改为 fetch）、crypto-js（本地存储不再“加密”）、moment（本地格式化函数）、未使用的 Markdown/KaTeX/代码高亮样式与依赖，SVG 图标精灵改为独立代码块。站点只保留 Usagi 浅色主题，不再提供深色模式。

验收要求见 [docs/usagi.md](docs/usagi.md)。迁移检查与已知限制见 [docs/migration.md](docs/migration.md)。验证记录见 [docs/validation.md](docs/validation.md)。本仓库 CI 只验证源码和构建镜像，不自动发布或切换站点。
