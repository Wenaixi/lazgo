# 蓝奏云上传文件逆向

## 状态: 已完成 (2026-05-19)

---

## 端点

```
URL:  https://pc.woozooo.com/html5up.php
Method: POST
Content-Type: multipart/form-data
```

---

## 请求参数

| 字段 | 值 | 类型 | 说明 |
|------|-----|------|------|
| task | `1` | form | 上传任务 |
| folder_id | 文件夹ID | form | `-1`=根目录 |
| vei | Base64 | form | 反爬 (可选) |
| upload_file | 文件内容 | file | 文件流 |

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
  "info": "上传成功",
  "text": [{
    "icon": "txt",
    "id": "286099123",
    "f_id": "is50X3pb2gud",
    "name_all": "文件名.txt",
    "name": "文件名",
    "size": "10.7 K",
    "time": "0 分钟前",
    "is_newd": "https://wwaqh.lanzoul.com"
  }]
}
```

| 字段 | 说明 |
|------|------|
| zt | `1`=成功 |
| text[].id | 新文件 ID (数字) |
| text[].f_id | 分享短码 |
| text[].name_all | 完整文件名 |
| text[].size | 大小 |
| text[].is_newd | 下载域名 |

---

## 完整流程

```
POST https://pc.woozooo.com/html5up.php
  multipart: upload_file=<file> + task=1 + folder_id=-1
→ {"zt":1,"text":[{"id":"286099123","name_all":"test.txt"}]}
```

---

## 注意事项

- 上传入口是 `html5up.php`，不是 `doupload.php`
- `folder_id=-1` 表示根目录，正数是具体文件夹 ID
- `vei` 参数可选，浏览器有时带有时不带
- 文件通过 multipart 流式上传，不需要 base64 编码

## Go 代码位置

`go/pkg/files/files.go` → `Upload()`

---

*创建: 2026-05-12 | 更新: 2026-05-19*
