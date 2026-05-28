# 快速开始

## 前置要求

- Go 1.21+
- 蓝奏云账号

## 构建

```bash
cd go
go build -o lazgo ./cmd/lazgo
```

得到单二进制文件 `lazgo`（或 `lazgo.exe`）。

## 登录

```bash
./lazgo login -u <用户名> -p <密码>
```

成功后将 cookie 保存到 `lanzou_cookies.json`。

```bash
# 可选：自定义 cookie 路径
export LAZGO_COOKIE_FILE=~/my_cookies.json
./lazgo login -u user -p pass
```

## 基本操作

```bash
# 上传文件
./lazgo upload hello.txt

# 查看根目录
./lazgo list

# 获取分享链接
./lazgo share link <file_id>

# 分享链接转直链
./lazgo share direct <share_url>

# 删除文件
./lazgo delete <file_id>

# 回收站
./lazgo recycle list
./lazgo recycle restore <file_id>
```

## 作为 Go 库使用

```go
package main

import (
    "encoding/json"
    "os"

    "github.com/lazgo/lazgo/pkg/client"
    "github.com/lazgo/lazgo/pkg/files"
    "github.com/lazgo/lazgo/pkg/share"
)

func main() {
    // 加载已保存的 cookie
    data, _ := os.ReadFile("lanzou_cookies.json")
    var cookies map[string]string
    json.Unmarshal(data, &cookies)

    c := client.New(cookies)

    // 上传
    info, err := files.Upload(c, "report.pdf", -1)
    if err != nil {
        panic(err)
    }
    println("上传成功:", info.ID)

    // 列表
    items, _ := files.ListFiles(c, -1, 1)
    for _, f := range items {
        println(f.ID, f.Name, f.Size)
    }

    // 分享
    link, _ := share.GetShareLink(c, 123456)
    dl, _ := share.ShareToDirect(link.URL, "")
    println("直链:", dl.URL)
}
```

## 项目结构

```
├── go.mod
├── go/
│   ├── cmd/lazgo/main.go   # CLI 源码
│   ├── internal/utils/       # ACW 解算
│   └── pkg/                  # 可导入库
│       ├── client/           # HTTP 客户端
│       ├── auth/             # 登录
│       ├── files/            # 文件操作
│       ├── folders/          # 文件夹
│       ├── recycle/          # 回收站
│       ├── share/            # 分享
│       └── models/           # 数据结构
├── docs/                     # 文档
└── data/                     # Cookie
```

## 下一步

- [CLI 完整命令参考](CLI.md)
- [API 参考](API.md)
- [逆向工程文档](reverse-engineering/)
