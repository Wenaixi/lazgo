# 蓝奏云列表文件逆向

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
| task | `5` | 列表任务 |
| folder_id | 文件夹ID | `-1`=根目录 |
| pg | 页码 | 从 1 开始 |
| vei | Base64 | 反爬 (可选) |

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
  "zt": 1,
  "info": 1,
  "text": [
    {
      "icon": "txt",
      "id": "286099123",
      "name_all": "文件.txt",
      "name": "文件",
      "size": "10.7 K",
      "time": "4 分钟前",
      "downs": "0",
      "onof": "0",
      "is_lock": "0"
    },
    {
      "icon": "folder",
      "id": "13489013",
      "name_all": "我的文件夹",
      "name": "我的文件夹",
      "size": "",
      "onof": "1"
    }
  ]
}
```

| 字段 | 说明 |
|------|------|
| zt | `1`=成功 |
| text[].id | 文件/文件夹 ID |
| text[].name_all | 完整名称 |
| text[].size | 文件大小 (文件夹为空) |
| text[].time | 上传时间 |
| text[].onof | `"0"`=文件, `"1"`=文件夹 |
| text[].icon | 图标类型 |

---

## 完整流程

```
POST https://pc.woozooo.com/doupload.php
  data: task=5&folder_id=-1&pg=1
→ JSON: text[] 列表
```

---

## 注意事项

- `folder_id=-1` 是根目录，正数是文件夹 ID
- 文件和文件夹混在同一个列表中，通过 `onof` 区分
- 翻页用 `pg` 参数，从 1 开始
- `vei` 可选，Go 实现中没传也能用

## Go 代码位置

`go/pkg/files/files.go` → `ListFiles()`

---

*创建: 2026-05-12 | 更新: 2026-05-19*
