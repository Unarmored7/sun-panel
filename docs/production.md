# 正式部署

2026-10-10，经用户批准，OVH `ovh-ca-32g-01` 上的导航站已从官方闭源镜像切换到本 fork。

2026-10-11 起 ovh-ca-32g-01 改为 Proxmox VE：生产实例在虚拟机 **`ovh-ca-app-01`**（`ssh ovh-ca-app-01`，经 Tailscale 子网路由 10.10.20.11），开发实例在 **`ovh-ca-dev-01`**；目录 `/opt/docker/sun-panel`、`/opt/docker/sun-panel-dev` 不变，由 edge 上的 Caddy 反代 `ovh-ca-app-01:3002`。见 `Code/shell/ovh-ca-32g-01`。包含 Gemini 的 Usagi UI 提交 `ed8c7c6`，以及 lint、旧搜索引擎资源和深色搜索框配色兼容修复。

- 域名：`https://nav.usaginest.com`；沿用宿主机端口 `3002`。
- 生产 Compose：`/opt/docker/sun-panel/docker-compose.yaml`；项目名 `sun-panel-usagi`，容器名 `sun-panel`。
- 镜像使用 `usagi/sun-panel:<提交短 SHA>` 固定标签，不使用开发镜像的浮动标签；关闭此容器的 Watchtower 自动更新。
- 实际数据目录、源码提交、镜像 ID 与验证结果以服务器 `/opt/docker/sun-panel/USAGI_DEPLOYMENT.json` 为准。
- 原账号密码与数据保持原值，重启后需要重新登录。当前有 14 个分组、215 个链接、3 个私密分组；公开访问仅显示 162 个链接。
- 原配置有 1 个上传图标文件缺失，图标地址仍保留，前端以站点简称兜底显示。
- 升级首次访问时将旧浏览器的深色或跟随系统偏好迁移为浅色，默认保持奶油背景；此后在个人设置中主动选择的主题会正常保存。

## 数据与静态资源

生产使用停止写入后建立的完整数据副本，原 `/opt/docker/sun-panel/conf` 保留。新副本分别挂载到 `/app/conf`、`/app/database`、`/app/uploads` 与 `/app/web/custom`；SQLite 和上传路径已适配本分支。favicon 保留；页面使用源码构建的 `index.html`，旧 HTML 注入保存在备份中。

开发实例仍使用 `/opt/docker/sun-panel-dev/src/conf` 和独立开发账号，仅绑定 `127.0.0.1:3003`。不要用开发数据或账号覆盖生产。

## 备份与回滚

首次成功切换前的冷备份位于：

`/opt/docker/sun-panel/backups/20261010T164727Z`

其中保留原 Compose、完整 `conf`、镜像 ID、容器信息以及 `rollback.sh`。旧镜像 `hslr/sun-panel:latest` 已随迁移加载到 ovh-ca-app-01。

迁移时没有带上旧容器 `sun-panel-legacy-20261010T164727Z`，所以 `rollback.sh`（依赖重命名该容器）已不可用。回滚改为用备份中的原 Compose 重新创建：

```sh
ssh ovh-ca-app-01 'cd /opt/docker/sun-panel && docker compose down && cp backups/20261010T164727Z/docker-compose.yaml docker-compose.yaml && docker compose up -d'
```

该回滚恢复切换前的旧站点及原数据；执行前应备份切换后新增的数据。后续更新也应保留当前 Compose 与数据库一致性备份，然后再更新镜像。

## 验证

切换前完成前端类型检查、构建、lint、Go 测试、GitHub CI、12 种布局场景，以及分类、搜索、登录和深色主题检查。正式启动后逐列比对分组、链接、账号及用户配置，验证数据库完整性、公开内容数量和私密分组访问限制；通过正式 HTTPS 域名检查页面。
