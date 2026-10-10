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

设计常量位于 `src/styles/usagi-tokens.css`，页面布局位于 `src/styles/usagi.css`，跨站链接位于 `src/components/usagi/UsagiFooter.vue`。分组样式解析集中在 `src/utils/groupCardStyle.ts`；不要再修改构建后的 HTML 或用负边距补页脚。

验收要求见 [docs/usagi.md](docs/usagi.md)。迁移检查与已知限制见 [docs/migration.md](docs/migration.md)。验证记录见 [docs/validation.md](docs/validation.md)。本仓库 CI 只验证源码和构建镜像，不自动发布或切换站点。
