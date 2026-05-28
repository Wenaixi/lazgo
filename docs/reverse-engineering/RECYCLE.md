# 蓝奏云回收站逆向

## 状态: 已完成 (2026-05-19)

---

## 6 种操作一览

| 操作 | action 值 | 参数字段 | 说明 |
|------|----------|---------|------|
| 恢复文件 | `file_restore` | `file_id` | 从回收站恢复 |
| 永久删除文件 | `file_delete_complete` | `file_id` | 不可恢复 |
| 恢复文件夹 | `folder_restore` | `folder_id` | 文件夹恢复 |
| 彻底删除文件夹 | `folder_delete_complete` | `folder_id` | 不可恢复 |
| 恢复全部 | `restore_all` | — | 批量恢复 |
| 清空回收站 | `delete_all` | — | 批量永久删除 |

---

## 统一两步流程

所有 6 种操作遵循相同模式：

### Step 1: GET 确认页

```
GET /mydisk.php?item=recycle&action={action}[&{id_field}={id}]
```

返回含 `<form>` 的 HTML 确认页面，form 内含 `formhash`。

### Step 2: POST 执行

```
POST /mydisk.php?item=recycle
Content-Type: application/x-www-form-urlencoded
```

### POST 参数

**单个文件/文件夹操作** (有 `ref`):

| 字段 | 值 | 说明 |
|------|-----|------|
| action | 操作名 | 如 `file_restore` |
| task | 同 action | |
| file_id 或 folder_id | ID | 目标 ID |
| ref | URL (URL编码) | 来源页面 |
| formhash | hex | 反 CSRF |

**批量操作** (`restore_all` / `delete_all`，无 `ref`):

| 字段 | 值 | 说明 |
|------|-----|------|
| action | 操作名 | |
| task | 同 action | |
| formhash | hex | 反 CSRF |

### POST Headers

```
User-Agent: Chrome UA
Content-Type: application/x-www-form-urlencoded
Origin: https://pc.woozooo.com
Referer: 确认页 URL
Cookie: (4 个必需 cookie)
```

### 响应

成功时返回 `User's Control Panel` HTML 页面，含 JS 重定向到回收站列表页。检测 `成功` 字符串确认操作完成。

---

## 回收站列表

### 端点

```
GET https://pc.woozooo.com/mydisk.php?item=recycle&action=files
```

### 解析方式

从 HTML `<tr>` 行中提取：

| 提取内容 | 正则/来源 |
|---------|----------|
| 文件 ID | `file_restore&file_id=(\d+)` 链接 |
| 文件夹 ID | `folder_restore&folder_id=(\d+)` 链接 |
| 文件名 | `<img .../> 文件名</a>` |
| 日期 | `\d{4}-\d{2}-\d{2}` |

### 重要

`doupload.php?folder_id=-1` **不是**回收站列表（返回的是根目录文件）。必须用 mydisk.php 页面 HTML 解析。

---

## formhash

### 提取

从回收站列表页或确认页的 `<form>` 中提取：

```html
<input type="hidden" name="formhash" value="84a7ea2d" />
```

正则：`name="formhash" value="([a-f0-9]+)"`

### 缓存策略

- 登录会话内 formhash 一致，可缓存复用
- 每次 POST 操作成功后应 `InvalidateFormhash()` 重新获取
- 过期/无效 formhash 会导致操作返回"成功"但实际无效

---

## 完整示例

### 恢复文件

```
1. GET /mydisk.php?item=recycle&action=file_restore&file_id=286941808
   → 确认页面 HTML

2. POST /mydisk.php?item=recycle
   data: action=file_restore&task=file_restore&file_id=286941808
         &ref=https%3A%2F%2Fpc.woozooo.com%2Fmydisk.php%3Fitem%3Drecycle%26action%3Dfiles
         &formhash=84a7ea2d
   → 200 HTML with "成功" → JS 重定向回 recycle&action=files
```

### 清空回收站

```
1. GET /mydisk.php?item=recycle&action=delete_all
   → 确认页面 HTML ("文件夹: 0, 文件: 0")

2. POST /mydisk.php?item=recycle
   data: action=delete_all&task=delete_all&formhash=84a7ea2d
   → 200 HTML with "成功" → JS 重定向回 recycle&action=files
```

---

## 踩坑记录

| 问题 | 表现 | 根因 | 修复 |
|------|------|------|------|
| Referer 校验 | POST 返回"成功"但操作无效 | 服务端校验 Referer 必须匹配确认页 URL（含 `&file_id=X`），不能用通用列表页 | 动态构造 Referer |
| 列表缓存 | 操作后 `ListRecycle` 仍显示旧数据 | 服务端/中间层对 GET 响应有缓存 | URL 加 `_=<nano>` 时间戳破缓存 |
| 连接复用 | 同一 `http.Client` 多次请求返回相同响应 | HTTP keep-alive 导致连接复用 | `ListRecycle` 每次创建独立 `http.Client` |
| 服务端延迟 | 操作成功后立刻查列表仍有旧数据 | 服务端更新回收站有 2-3 秒延迟 | 无需处理，重查即正确 |

## 注意事项

- POST 字段顺序对蓝奏云服务端有影响，必须为 `action→task→id→ref→formhash`
- `ref` 参数必须 URL 编码（`url.Values.Encode()` 自动处理）
- 批量操作 (`restore_all`/`delete_all`) 不含 `ref` 参数
- 文件用 `file_id`，文件夹用 `folder_id`，两个字段不同
- 回收站列表不是 `doupload.php?folder_id=-1`

## Go 代码位置

- `go/pkg/recycle/recycle.go` → 全部 6 个操作 + HTML 列表解析
- `go/pkg/client/client.go` → `FetchFormhash()`, `InvalidateFormhash()`

---

*创建: 2026-05-19*
