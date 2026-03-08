# Mihomo 订阅管理 (Golang 版本)

![Version](https://img.shields.io/badge/version-v1.0.0-blue.svg)
![Platform](https://img.shields.io/badge/platform-linux%20%7C%20darwin%20%7C%20windows-lightgrey)
![Go](https://img.shields.io/badge/go-1.21+-blue)

Mihomo (Clash Meta) 订阅配置管理工具，支持多订阅、定时更新、软链接管理、自动重试。

## 特性

- 支持多订阅配置
- 定时自动更新 (crontab)
- 软链接管理
- 配置模板合并
- 忽略 SSL 证书错误（可选）
- 网络请求自动重试（可配置重试次数和超时时间）
- 交互式订阅选择（支持序号和别名选择）

## 快速开始

### 下载预编译二进制

从 [Releases](https://github.com/your-repo/mihomo-sync/releases) 下载对应平台的二进制文件。

### 使用 Makefile 编译

```bash
# 编译当前平台
make build

# 编译所有平台
make build-all

# 指定版本号编译
make build VERSION=1.0.0
```

### 使用 Go 直接编译

```bash
go build -o mihomo-sync
```

## 配置说明

### subs.conf

```yaml
# 软链接目标目录
output_dir: "./Mihomo-sync"

# 配置模板路径
template: "./Mihomo-sync/template.yaml"

# 日志配置
log_file: "./Mihomo-sync/update.log"
log_level: "INFO"              # DEBUG, INFO, WARNING, ERROR, CRITICAL
max_log_size: "10MB"           # 日志文件最大大小，支持 B, KB, MB, GB

# 网络配置
retry: 3                       # 下载失败重试次数，默认 3 次
timeout: 30                    # 下载超时时间（秒），默认 30 秒

# 服务重启配置（可选）
restart_command: "systemctl restart mihomo"

subscriptions:
  - alias: "main"
    url: "https://example.com/sub?token=your_token"
    auto_update: true
    interval: 6
    insecure_skip_verify: true  # 忽略证书错误（可选，默认 false）
```

### template.yaml

配置模板文件，包含：
- 端口、DNS、geodata 等基础配置
- 空的 `proxies: []`、`proxy-groups: []`、`rules: []` 占位

脚本会保留模板配置，只更新这三个字段。

## 使用方法

### 交互式选择订阅（推荐）
```bash
./mihomo-sync
# 列出所有订阅，等待输入选择
# 支持序号（1, 2, 3...）或别名（main, backup）
```

### 列出所有订阅
```bash
./mihomo-sync --list
```

### 更新所有订阅
```bash
./mihomo-sync --all
# 输出：configs/main.yaml, configs/backup.yaml...
```

### 更新指定订阅并切换
```bash
./mihomo-sync -s main
# 输出：configs/main.yaml
# 软链接：config.yaml -> configs/main.yaml
```

### 指定软链接目标目录
```bash
./mihomo-sync -s main -t /etc/mihomo
# 创建：/etc/mihomo/config.yaml -> configs/main.yaml
```

### 使用自定义配置文件
```bash
./mihomo-sync -s main -c /path/to/subs.conf
```

### 更新但不设置 crontab
```bash
./mihomo-sync -s main --no-crontab
```

## 参数说明

| 参数 | 说明 |
|------|------|
| `-l, --list` | 列出所有订阅 |
| `-s, --select ALIAS` | 指定订阅别名（只更新该订阅） |
| `-t, --target PATH` | 软链接目标目录 |
| `-c, --config PATH` | 配置文件路径 |
| `-d, --base-dir PATH` | 脚本所在目录 |
| `--all` | 更新所有订阅到 configs 目录 |
| `--no-crontab` | 不设置定时任务 |
| `--template PATH` | 配置模板路径（默认 template.yaml） |
| `-h, --help` | 显示帮助信息 |
| `-v, --version` | 显示版本号 |

## 工作模式

### 模式 1：更新指定订阅
```bash
./mihomo-sync -s main
```
- 只更新 `main` 订阅
- 输出到 `configs/main.yaml`
- 创建/更新软链接：`<output_dir>/config.yaml -> configs/main.yaml`
- 如果配置了 `auto_update: true`，自动设置 crontab

### 模式 2：更新所有订阅
```bash
./mihomo-sync --all
```
- 更新所有订阅
- 输出到 `configs/` 目录：`configs/main.yaml`, `configs/backup.yaml`...
- 不创建软链接

## 定时任务

```bash
# 查看定时任务
crontab -l

# 删除特定订阅的定时任务
crontab -l | grep -v "mihomo-sync.*main" | crontab -
```

## 网络配置

在 `subs.conf` 中配置网络相关参数：

```yaml
# 下载失败重试次数，默认 3 次
retry: 3

# 下载超时时间（秒），默认 30 秒
timeout: 30
```

重试间隔为 5 秒，适合网络不稳定或订阅源响应较慢的场景。
