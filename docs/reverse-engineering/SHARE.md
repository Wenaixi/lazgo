# 蓝奏云分享链接获取逆向

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
| task | `22` | 获取分享链接 |
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
{
  "zt": 1,
  "info": {
    "pwd": "5n3o",
    "onof": "1",
    "f_id": "i8fQN3pcg7uj",
    "taoc": "",
    "is_newd": "https://wwaqh.lanzoul.com"
  }
}
```

| 字段 | 说明 |
|------|------|
| zt | `1`=成功 |
| info.is_newd | 分享域名 |
| info.f_id | 分享短码 |
| info.onof | `"1"`=有密码, `"0"`=无密码 |
| info.pwd | 分享密码 (onof="1" 时有效) |
| info.taoc | 短网址 (通常为空) |

## 分享链接构造

```
分享链接 = is_newd + "/" + f_id
示例: https://wwaqh.lanzoul.com/i8fQN3pcg7uj
```

---

## 完整流程

```
POST https://pc.woozooo.com/doupload.php
  data: task=22&file_id=286163599
→ {"zt":1,"info":{"is_newd":"https://wwaqh.lanzoul.com","f_id":"i8fQN3pcg7uj","onof":"1","pwd":"5n3o"}}
→ 分享链接: https://wwaqh.lanzoul.com/i8fQN3pcg7uj (密码: 5n3o)
```

---

## 注意事项

- 只有文件可以创建分享链接，文件夹不行
- 分享链接转直链见 `DIRECT_LINK.md`（无需登录）
- 密码在创建分享时由蓝奏云自动生成，用户无法自定义

## Go 代码位置

`go/pkg/share/share.go` → `GetShareLink()`

---

*创建: 2026-05-13 | 更新: 2026-05-19*
