# 蓝奏云删除文件逆向

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
| task | `6` | 删除文件任务 |
| file_id | 文件ID | 数字 ID |

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
- `text=1`: 确认

---

## 完整流程

```
POST https://pc.woozooo.com/doupload.php
  data: task=6&file_id=286123310
→ {"zt":1,"info":"删除成功"}
→ 文件进入回收站
```

---

## 注意事项

- 删除文件进入回收站（可恢复），不是永久删除
- 永久删除见 `RECYCLE.md` 的 `file_delete_complete`
- 同一 endpoint `doupload.php` 处理多种 task

## Go 代码位置

`go/pkg/files/files.go` → `DeleteFile()`

---

*创建: 2026-05-12 | 更新: 2026-05-19*
