# 蓝奏云获取文件夹信息逆向

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
| task | `47` | 获取文件夹信息 |
| folder_id | 文件夹ID | `-1`=根目录 |

## 请求 Headers

```
User-Agent: Chrome UA
X-Requested-With: XMLHttpRequest
Referer: https://pc.woozooo.com/mydisk.php?item=files&action=index&u=<uid>
Cookie: (4 个必需 cookie)
```

## 响应

```json
{
  "zt": 2,
  "info": [{
    "name": "测试文件夹",
    "folder_des": "",
    "folderid": 13489819,
    "now": 1
  }],
  "text": []
}
```

| 字段 | 说明 |
|------|------|
| zt | `1` 或 `2`=成功 (与常规 `zt=1` 不同) |
| info[].name | 文件夹名 |
| info[].folderid | 文件夹 ID |
| info[].folder_des | 文件夹描述 |

---

## 完整流程

```
POST https://pc.woozooo.com/doupload.php
  data: task=47&folder_id=13489819
→ {"zt":2,"info":[{"name":"xxx","folderid":13489819}]}
```

---

## 注意事项

- `zt=2` 也表示成功（与常规 task 返回 `zt=1` 不同）
- 文件夹信息在 `info` 数组中（单条），文件列表在 `text` 中
- 根目录 `folder_id=-1` 返回信息通常为空
- `folderid` 是 JSON number，需注意科学计数法

## Go 代码位置

`go/pkg/folders/folders.go` → `GetFolderInfo()`

---

*创建: 2026-05-12 | 更新: 2026-05-19*
