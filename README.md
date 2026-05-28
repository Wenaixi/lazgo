# lazgo

蓝奏云 (Lanzou Cloud) HTTP API 客户端 — Go 实现，纯 HTTP 方案，无需浏览器。

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://pkg.go.dev/github.com/lazgo/lazgo)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 特性

- **单语言 Go**：编译为单二进制文件，零依赖部署
- **纯 HTTP**：无需浏览器，只需 HTTP 请求
- **库与 CLI 兼备**：`go get` 可导入，`go install` 得 CLI
- **反爬处理**：内置 ACW SC V2 反爬虫挑战解决
- **完整功能**：文件/文件夹/回收站/分享链接/直链转换
- **并发友好**：支持高并发上传

## 快速开始

### 安装

```bash
# CLI
go install github.com/lazgo/lazgo/cmd/lazgo@latest

# 库
go get github.com/lazgo/lazgo/pkg/client
```

### CLI 使用

```bash
# 登录
lazgo login -u username -p password

# 上传
lazgo upload report.pdf

# 列出文件
lazgo list

# 获取直链
lazgo share link 123456
lazgo share direct https://wwaqh.lanzoul.com/xxxx
```

### 作为库

```go
package main

import (
    "fmt"
    "github.com/lazgo/lazgo/pkg/client"
    "github.com/lazgo/lazgo/pkg/files"
)

func main() {
    c := client.NewFromFile("lanzou_cookies.json")
    info, err := files.Upload(c, "test.txt", -1)
    if err != nil {
        panic(err)
    }
    fmt.Printf("上传成功: %s\n", info.Name)
}
```

## CLI 命令

```
用法: lazgo [命令]

登录:
  lazgo login -u <用户名> -p <密码>     登录账户

文件:
  lazgo upload <文件> [--folder-id ID]  上传文件
  lazgo delete <ID>                     删除文件
  lazgo list [--folder-id ID] [--page N]  列出文件

文件夹:
  lazgo folder create <名称> [--parent-id ID] [--description DESC]
  lazgo folder delete <ID>
  lazgo folder info [ID]

回收站:
  lazgo recycle list [--page N]
  lazgo recycle restore <ID>
  lazgo recycle delete <ID>
  lazgo recycle restore-folder <ID>
  lazgo recycle delete-folder <ID>
  lazgo recycle clear
  lazgo recycle restore-all

分享:
  lazgo share link <ID>                   获取分享链接
  lazgo share direct <URL> [--password P]   获取直链
```

## 目录结构

```
lazgo/
├── cmd/lazgo/           # CLI 入口
├── internal/utils/     # ACW 解算
└── pkg/                # 可导入库
    ├── client/          # HTTP 客户端
    ├── auth/            # 登录认证
    ├── files/           # 文件操作
    ├── folders/         # 文件夹操作
    ├── recycle/         # 回收站
    ├── share/           # 分享 + 直链
    └── models/          # 数据类型
```

## 文档

- [API 参考](docs/API.md)
- [快速开始](docs/GETTING_STARTED.md)
- [CLI 指南](docs/CLI.md)
- [逆向工程](docs/reverse-engineering/)

## 配置

Cookie 文件默认 `lanzou_cookies.json`，可通过环境变量指定：

```bash
export LAZGO_COOKIE_FILE=/path/to/cookies.json
```

## 贡献

欢迎提交 Issue 和 Pull Request！请参阅 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 许可证

MIT License - 请参阅 [LICENSE](LICENSE) 文件。