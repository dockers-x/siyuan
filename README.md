# SiYuan Docker Patch

自动构建定制版 [SiYuan](https://github.com/siyuan-note/siyuan) Docker 镜像,去除了云账号相关内容,去除需要登陆才能使用s3和webdav同步的思源笔记版本。


## Patch

Patches 包含:
- **001-mock-vip-user.patch**: `isPaidUser` 始终返回 true (启用 S3/WebDAV 同步)
- **002-disable-account-ui.patch**: 隐藏账户设置，移除思源官方同步选项
- **003-sync-check.patch**: 同步前检查配置
- **004-remove-share-to-liandi.patch**: 移除分享到链滴社区功能
- **005-reuse-upstream-onnx.patch**: 从同版本的官方 Docker 镜像复用 ONNX Runtime 库及许可文件

## Workflow

- **定时执行**: 每周二和周五 12:00 (UTC) 自动检查上游新版本
- **手动触发**: 在 GitHub Actions 点击 "Run workflow" 即可
- 自动检测最新上游版本
- 如果版本已存在则跳过构建
- 构建依赖 `b3log/siyuan` 发布同版本的 amd64 / arm64 镜像；通过摘要锁定依赖，不回退到 `latest`
- 新镜像发布并验证成功后，工作流会在默认分支更新 `.workflows/last-release.json`，记录版本、UTC 发布时间和本次 Actions 运行链接

## 设置

1. Fork 此仓库
2. 添加以下 Secrets:
   - `DOCKERHUB_USERNAME`: Docker Hub 用户名
   - `DOCKERHUB_TOKEN`: Docker Hub Access Token

## 添加/更新 Patches

1. 在 `patches/` 目录创建 patch 文件
2. 使用数字前缀排序 (如 `005-my-feature.patch`)
3. Patches 按字母顺序应用

## 致谢

参考了 [siyuan-patch](https://github.com/demoshang/siyuan-patch)
