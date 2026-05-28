# 蓝奏云分享链接转直链逆向

## 状态: 已更新 (2026-05-19)

页面架构已升级，旧版静态 HTML 方案失效。

---

## 新架构 (2023+)

分享页现在是两层结构：

```
GET /<share_id>
  → ACW 验证
  → HTML 主页面 (包含 fid + iframe → fn 页)
  
GET /fn?<encoded_params>
  → HTML 子页面 (包含 wp_sign + ajaxdata + kdns)
  → JS 自动发起 AJAX

POST /ajaxm.php?file=<fid>
  → {"zt":1,"dom":"...","url":"..."}
  → 直链 = dom + "/file/" + url
```

---

## Step 1: GET 分享页 + ACW

```
GET https://wwaqh.lanzoul.com/iTllh3psynhg
```

如响应包含 `var arg1=` → ACW SC V2 挑战，解算后带 cookie 重试。

### 主页面提取

| 提取内容 | 正则 | 说明 |
|---------|------|------|
| 文件 ID | `var fid = (\d+)` | 数字文件ID |
| fn 页面路径 | `src="(/fn\?[^"]+)"` | iframe 中的子页面地址 |
| 是否有密码 | `id="pwd"` | 密码门检测 |

---

## Step 2: GET fn 页面

```
GET https://wwaqh.lanzoul.com/fn?BmAAag9gV...
```

### fn 页面提取

| 提取内容 | 正则 | 说明 |
|---------|------|------|
| wp_sign | `var wp_sign = '([^']+)'` | 预计算签名（服务端生成） |
| ajaxdata | `var ajaxdata = '([^']+)'` | websignkey + signs 共用值 |
| kdns | `var kdns = (\d+)` | kd 参数 |
| file_id | `ajaxm\.php\?file=(\d+)` | 确认文件ID（通常不需要） |

---

## Step 3: POST 获取直链

```
POST https://<domain>/ajaxm.php?file=<fid>
Content-Type: application/x-www-form-urlencoded
```

### 新参数 (2023+)

| 参数 | 值来源 | 说明 |
|------|--------|------|
| action | `downprocess` | 固定 |
| websignkey | ajaxdata | 从 fn 页提取 |
| signs | ajaxdata | 从 fn 页提取 |
| sign | wp_sign | 从 fn 页提取 |
| websign | (空) | 固定 |
| kd | kdns | 从 fn 页提取 |
| ves | 1 | 固定 |
| p | 密码 | 仅当有密码时 |

### 响应

```json
{
  "zt": 1,
  "dom": "https://developer2.lanrar.com",
  "url": "?A2UAPl5vV2YIAVRs...",
  "inf": 0
}
```

### 直链构造

```
direct_link = dom + "/file/" + url
```

---

## 旧架构 (2023前, 已废弃)

旧版直接从主页面提取 `ajaxm.php?file=XXXX` 和 `sign`，POST 参数仅 `action=downprocess&sign=X&kd=1&p=<pwd>`。

新旧关键区别：新增 `websignkey`、`signs`、`websign`、`ves` 参数，sign 改名为 `wp_sign` 且需从 fn 子页面提取。

---

*创建时间: 2026-05-13*
*最后更新: 2026-05-19*
