# 蓝奏云删除文件夹逆向

## 状态: 已完成 (2026-05-19)

---

## 端点

```
URL:  https://pc.woozooo.com/doupload.php
Method: POST
Content-Type: application/x-www-form-urlencoded
```

---

## 请求参数

| 字段 | 值 | 说明 |
|------|-----|------|
| task | `3` | 删除文件夹任务 |
| folder_id | 文件夹ID | 要删除的文件夹 ID |

## 请求 Headers

```
User-Agent: Chrome UA
X-Requested-With: XMLHttpRequest
Referer: https://pc.woozooo.com/mydisk.php?item=files&action=index&u=<uid>
Cookie: (4 个必需 cookie)
```

## 响应

```json
{"zt":1,"info":"删除成功","text":1}
```

- `zt=1`: 成功

---

## 完整流程

```
POST https://pc.woozooo.com/doupload.php
  data: task=3&folder_id=13489846
→ {"zt":1,"info":"删除成功"}
→ 文件夹及其内容进入回收站
```

---

## 注意事项

- 删除文件夹会同时删除其中所有文件（全部进入回收站）
- 与删除文件 (`task=6`) 同 endpoint，参数不同
- 回收站中永久删除文件夹见 `RECYCLE.md` 的 `folder_delete_complete`

## Go 代码位置

`go/pkg/folders/folders.go` → `DeleteFolder()`

---

*创建: 2026-05-12 | 更新: 2026-05-19*
