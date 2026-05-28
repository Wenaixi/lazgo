# CLI 使用指南

## 安装

```bash
cd go
rm -f lazgo.exe && go build -o lazgo ./cmd/lazgo
```

Cookie 默认读取 `./lanzou_cookies.json`，可通过环境变量覆盖:

```bash
export LAZGO_COOKIE_FILE=/path/to/cookies.json
```

## 命令结构

```
lazgo
├── login                      登录蓝奏云
├── upload <file>              上传文件
├── delete <id>                删除文件
├── list                       列出根目录文件
├── folder
│   ├── create <name>          创建文件夹
│   ├── delete <id>            删除文件夹
│   └── info [id]              查看文件夹信息
├── recycle
│   ├── list                   回收站列表
│   ├── restore <id>           恢复文件
│   ├── delete <id>            永久删除文件
│   ├── restore-folder <id>    恢复文件夹
│   ├── delete-folder <id>     永久删除文件夹
│   ├── clear                  清空回收站
│   └── restore-all            恢复全部
└── share
    ├── link <id>              获取分享链接
    └── direct <url>           分享链接转直链
```

## 命令详解

### login

```bash
lazgo login -u myname -p mypass
```

### upload

```bash
lazgo upload report.pdf
lazgo upload report.pdf --folder-id 12345
```

### delete

```bash
lazgo delete 286123310
```

### list

```bash
lazgo list
lazgo list --folder-id 12345
lazgo list --folder-id 12345 --page 2
```

### folder

```bash
lazgo folder create "我的文件夹"
lazgo folder create "子目录" --parent-id 12345 --description "备注"
lazgo folder delete 12345
lazgo folder info
lazgo folder info 12345
```

### recycle

```bash
lazgo recycle list
lazgo recycle list --page 2
lazgo recycle restore 286274693           # 恢复文件
lazgo recycle restore-folder 13512388     # 恢复文件夹
lazgo recycle delete 286274691            # 永久删除文件
lazgo recycle delete-folder 13512388      # 永久删除文件夹
lazgo recycle clear                       # 清空
lazgo recycle restore-all                 # 恢复全部
```

### share

```bash
lazgo share link 286163599
lazgo share direct https://example.lanzoul.com/abc123
lazgo share direct https://example.lanzoul.com/abc123 -p 5n3o
```

## 完整工作流

```bash
lazgo login -u user -p pass
lazgo upload document.pdf
lazgo list
lazgo share link 286163599
lazgo share direct https://wwaqh.lanzoul.com/xxxx
lazgo delete 286163599
lazgo recycle list
lazgo recycle restore 286163599
```
