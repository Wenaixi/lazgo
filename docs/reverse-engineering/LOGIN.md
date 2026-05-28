# 蓝奏云登录逆向

## 状态: 已完成 (2026-05-19)

---

## 端点

| 用途 | URL | Method |
|------|-----|--------|
| 登录页 | `https://accounts.woozooo.com/accounts.php?action=login&ref=up.woozooo.com` | GET |
| 登录提交 | `https://accounts.woozooo.com/accounts.php` | POST |

---

## 完整流程

```
1. GET 登录页
   → 从 Set-Cookie 提取 acw_tc
   → 检查响应是否包含 var arg1=' (ACW SC V2 反爬)
   
2. 如触发 ACW: 计算 acw_sc__v2
   → 提取 arg1 字符串
   → 按位置映射表 (40位) 重新排列
   → 每 2 字符 hex 与 mask XOR 得到 cookie 值
   
3. 带 cookie 重 GET 登录页 (可选)
   
4. POST 登录
   → 发送 task=uselogin, username, password, ref
   → 从响应 Set-Cookie 提取 phpdisk_info, ylogin, uag, PHPSESSID
```

---

## POST 参数

| 字段 | 值 | 说明 |
|------|-----|------|
| task | `uselogin` | 固定 |
| username | 用户名 | 手机号/邮箱 |
| password | 密码 | 明文 |
| ref | `up.woozooo.com` | 固定 |

## POST Headers

```
User-Agent: Chrome UA
Content-Type: application/x-www-form-urlencoded
X-Requested-With: XMLHttpRequest
Origin: https://accounts.woozooo.com
Referer: 登录页 URL
Cookie: acw_tc + acw_sc__v2 (如触发 ACW)
```

## 登录成功响应

```json
{"zt":1,"msgs":"https://up.woozooo.com/acc.php?t=xxx","usename":""}
```

- `zt=1`: 成功
- `msgs`: 登录后跳转 URL

---

## ACW SC V2 反爬

### 触发条件

登录页响应 HTML 包含 `var arg1=`。

### 算法

`arg1` 40 字符 → 按 40 位置映射表重排 → 每 2 hex 字节与 mask XOR → 结果 40 hex → `acw_sc__v2` cookie

| 常量 | 来源 |
|------|------|
| 位置表 | `internal/utils/acw.go:acwPos` |
| mask | `3000176000856006061501533003690027800375` |

### Go 实现

`go/internal/utils/acw.go` → `ACWSCV2(html string) (string, error)`

---

## Cookie

登录后需要的 4 个核心 cookie：

| Cookie | 来源 | 说明 |
|--------|------|------|
| `phpdisk_info` | Set-Cookie | 主凭证 (最长) |
| `ylogin` | Set-Cookie | 用户 ID |
| `uag` | Set-Cookie | 会话标识 |
| `PHPSESSID` | Set-Cookie | PHP 会话 |

外加 `acw_tc` 和 `acw_sc__v2` (反爬用)。

---

## Go 代码位置

- `go/pkg/auth/auth.go` → `Login()`
- `go/internal/utils/acw.go` → `ACWSCV2()`, `ParseSetCookies()`

---

*创建: 2024-05-12 | 更新: 2026-05-19*
