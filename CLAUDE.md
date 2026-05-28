# CLAUDE.md

This file provides guidance to Claude Code when working with this repository.

## 项目概述

lazgo — 蓝奏云 (Lanzou Cloud) 纯 HTTP 客户端。Go 实现，无需浏览器。

## 开发原则

**文档与代码同步演进，禁止事后补写。**

| 场景 | 必须更新的文档 |
|------|---------------|
| 新增功能 | `docs/reverse-engineering/<FEATURE>.md` + `docs/API.md` |
| 修复 bug（非拼写错误） | 对应逆向文档补充踩坑记录 |
| API 行为变化（如端点参数变） | 对应逆向文档 + `CLAUDE.md` API 汇总 |
| 反向工程有新发现 | 实时写入逆向文档，不要等到"最后" |
| 重构目录/模块 | `CLAUDE.md` 目录结构 + `README.md` |

**为什么：**
- 逆向知识是最容易流失的资产（服务端一变，只有当时抓包的人知道真相）
- 文档滞后 = 下次逆向同一个功能要重新抓包
- 每个 bug 的 root cause 写进文档，后来者不用再踩

## 目录结构

```
W-NetDisk/
├── CLAUDE.md              # 本文件
├── README.md              # 项目说明
├── go.mod                 # Go module
├── go/
│   ├── cmd/lazgo/        # CLI 入口 (cobra)
│   │   └── main.go
│   ├── internal/utils/    # 内部工具
│   │   └── acw.go         # ACW SC V2 解算
│   └── pkg/               # 可导入库
│       ├── client/        # HTTP 客户端 + 反爬
│       ├── auth/          # 登录认证
│       ├── files/         # 文件操作
│       ├── folders/       # 文件夹操作
│       ├── recycle/       # 回收站
│       ├── share/         # 分享链接 + 直链
│       └── models/        # 数据结构
├── docs/                  # 文档
│   ├── API.md            # API 参考
│   ├── CLI.md            # CLI 使用指南
│   ├── GETTING_STARTED.md # 快速开始
│   └── reverse-engineering/ # 逆向工程文档
└── data/                  # Cookie 凭证 (git 忽略)
    ├── lanzou_cookies.json
    └── lanzou_credentials.json
```

## CLI 命令

```
lazgo login -u <user> -p <pass>
lazgo upload <file> [--folder-id ID]
lazgo delete <id>
lazgo list [--folder-id ID] [--page N]
lazgo folder create <name> [--parent-id ID] [--description D]
lazgo folder delete <id>
lazgo folder info [id]
lazgo recycle list [--page N]
lazgo recycle restore <id>
lazgo recycle delete <id>
lazgo recycle restore-folder <id>
lazgo recycle delete-folder <id>
lazgo recycle clear
lazgo recycle restore-all
lazgo share link <id>
lazgo share direct <url> [--password P]
```

Cookie 文件默认 `./lanzou_cookies.json`，可通过环境变量 `LAZGO_COOKIE_FILE` 指定。

## 构建

```bash
cd go
go build -o lazgo ./cmd/lazgo
```

## 逆向工程流程

### jshook MCP 工具

| 工具 | 用途 |
|------|------|
| `browser_launch` | 启动浏览器 |
| `page_navigate` | 导航到页面 |
| `network_enable` | 开启网络捕获 |
| `network_get_requests` | 获取捕获的请求 |
| `network_get_response_body` | 查看响应体 |
| `get_script_source` | 查看页面 JS 源码 |
| `browser_close` | 关闭浏览器 |

### 标准流程

```
1. browser_launch → 打开浏览器
2. page_navigate → 访问目标页面（手动登录）
3. network_enable → 开启捕获
4. 用户在浏览器操作
5. network_get_requests → 分析 API
6. 在 Go 中实现对应 pkg
7. 更新 docs/reverse-engineering/
```

### 关键发现模式

| 类型 | 发现方式 |
|------|---------|
| JSON API | doupload.php + task=N |
| 表单提交 | mydisk.php + action=X + formhash |
| 分享页 | 双层架构: 主页面 + fn 子页面 |

## 反爬机制

| 类型 | 特征 | 解决方案 |
|------|------|---------|
| ACW SC V2 | 响应含 `var arg1=` | `internal/utils/acw.go` → `acw_sc_v2()` |
| formhash | `name="formhash" value="..."` | 正则提取，每次操作前刷新 |
| 分享页双层 | iframe → fn 页面 | 先取 fn URL，再取 wp_sign |

## Cookie 凭证

必须: `phpdisk_info` + `ylogin` + `uag` + `PHPSESSID`

## API 汇总

### 文件操作 (doupload.php)

| task | 功能 | 参数 |
|------|------|------|
| 1 | 上传文件 | multipart → html5up.php |
| 2 | 创建文件夹 | parent_id, folder_name |
| 3 | 删除文件夹 | folder_id |
| 5 | 列表文件 | folder_id, pg |
| 6 | 删除文件 | file_id |
| 22 | 获取分享链接 | file_id |
| 47 | 获取文件夹信息 | folder_id |

### 回收站 (mydisk.php?item=recycle)

两步流程: GET 确认页 → POST /mydisk.php?item=recycle

| action | 功能 | 参数字段 |
|--------|------|---------|
| file_restore | 恢复文件 | file_id, formhash, ref |
| file_delete_complete | 永久删除文件 | file_id, formhash, ref |
| folder_restore | 恢复文件夹 | folder_id, formhash, ref |
| folder_delete_complete | 彻底删除文件夹 | folder_id, formhash, ref |
| restore_all | 恢复全部 | formhash (无 ref) |
| delete_all | 清空回收站 | formhash (无 ref) |

回收站列表: 解析 `mydisk.php?item=recycle&action=files` HTML 页面（非 doupload.php?folder_id=-1）

### 分享链接转直链 (新架构 2023+)

```
1. GET /<share_id> → ACW → 主页面
   提取: var fid, iframe src="/fn?..."
2. GET /fn?<params> → 子页面
   提取: wp_sign, ajaxdata, kdns
3. POST /ajaxm.php?file=<fid>
   data: action=downprocess&websignkey=X&signs=X&sign=wp_sign&websign=&kd=1&ves=1
   响应: {"zt":1,"dom":"...","url":"..."}
4. 直链 = dom + "/file/" + url
```

## 特殊标识

| 值 | 含义 |
|------|------|
| folder_id=-1 | 根目录（doupload.php）|
| file_id | 数字文件ID |
| f_id | 分享短码 |

## 文档规范

每个功能独立一个 `.md` 文件在 `docs/reverse-engineering/`，统一格式：

```
# 标题
## 状态: 已完成 (日期)
## 端点 → 参数表 → Headers → 响应字段 → 完整流程 → 注意事项 → Go 代码位置
```

**核心纪律：**
- 逆向中每发现一个参数、一个 header、一个坑，**立刻**写入文档
- 代码修复后**立刻**更新文档中的注意事项/流程
- 文档末尾标注 Go 代码位置（`go/pkg/xxx/xxx.go`），保持可追溯
- 禁止"先做完再一次性补文档"
