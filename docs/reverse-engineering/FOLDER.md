# 蓝奏云创建文件夹逆向

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
| task | `2` | 创建文件夹任务 |
| parent_id | 父文件夹ID | `0`=根目录 |
| folder_name | 名称 | 文件夹名 |
| folder_description | 描述 | 可选，默认空 |

## 请求 Headers

```
User-Agent: Chrome UA
X-Requested-With: XMLHttpRequest
Referer: https://pc.woozooo.com/mydisk.php?item=files&action=index&u=<uid>
Cookie: (4 个必需 cookie)
```

## 响应

```json
{"zt":1,"info":"创建成功","text":"13489013"}
```

| 字段 | 说明 |
|------|------|
| zt | `1`=成功 |
| text | 新文件夹 ID (字符串) |

---

## 完整流程

```
POST https://pc.woozooo.com/doupload.php
  data: task=2&parent_id=0&folder_name=新文件夹
→ {"zt":1,"text":"13489013"}
```

---

## 注意事项

- `parent_id=0` 在根目录下创建，正数是子文件夹
- 返回的 `text` 是字符串类型的数字 ID
- 创建后文件夹进入回收站时需要用 `folder_id` (与 `file_id` 不同)

## Go 代码位置

`go/pkg/folders/folders.go` → `CreateFolder()`

---

*创建: 2026-05-12 | 更新: 2026-05-19*
