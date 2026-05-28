# lazgo API 参考

## 导入

```go
import (
    "github.com/lazgo/lazgo/pkg/client"
    "github.com/lazgo/lazgo/pkg/files"
    "github.com/lazgo/lazgo/pkg/folders"
    "github.com/lazgo/lazgo/pkg/recycle"
    "github.com/lazgo/lazgo/pkg/share"
    "github.com/lazgo/lazgo/pkg/auth"
    "github.com/lazgo/lazgo/pkg/models"
)
```

## Client

```go
c := client.New(cookies)                       // 从 map 创建
c := auth.Login(user, pass) → cookies          // 登录获取 cookies
```

| 方法 | 说明 |
|------|------|
| `c.UID() string` | 用户 ID |
| `c.CookieHeader() string` | Cookie 请求头 |
| `c.HTTPClient() *http.Client` | 底层 HTTP 客户端 |
| `c.Doupload(data) (*APIResponse, error)` | doupload.php 通用调用 |
| `c.FetchFormhash() (string, error)` | 获取回收站 formhash |

---

## files

```go
info, err := files.Upload(c, "file.txt", -1)
err := files.DeleteFile(c, 123456)
items, err := files.ListFiles(c, -1, 1)
```

---

## folders

```go
info, err := folders.CreateFolder(c, "name", 0, "")
err := folders.DeleteFolder(c, 123456)
info, err := folders.GetFolderInfo(c, 123456)
```

---

## recycle

```go
items, err := recycle.ListRecycle(c, 1)
err := recycle.RestoreFile(c, 123456)
err := recycle.DeleteRecycleFile(c, 123456)
err := recycle.RestoreFolder(c, 123456)
err := recycle.DeleteRecycleFolder(c, 123456)
err := recycle.ClearRecycle(c)
err := recycle.RestoreAll(c)
```

---

## share

```go
link, err := share.GetShareLink(c, 123456)
dl, err := share.ShareToDirect(shareURL, password)  // 无需 Client
```

---

## 数据类型

### FileInfo

```go
type FileInfo struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    Size     string `json:"size"`
    Time     string `json:"time"`
    IsFolder bool   `json:"is_folder"`
}
```

### FolderInfo

```go
type FolderInfo struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
}
```

### ShareLink / DirectLink

```go
type ShareLink struct {
    URL, Password, ShareID string
    HasPassword            bool
}

type DirectLink struct {
    URL, Filename, Method, Warning string
}
```

---

## 使用示例

```go
data, _ := os.ReadFile("lanzou_cookies.json")
var cookies map[string]string
json.Unmarshal(data, &cookies)
c := client.New(cookies)

// 上传
info, _ := files.Upload(c, "report.pdf", -1)

// 列表
items, _ := files.ListFiles(c, -1, 1)
for _, f := range items {
    fmt.Printf("%s: %s\n", f.ID, f.Name)
}

// 分享转直链
link, _ := share.GetShareLink(c, 123456)
dl, _ := share.ShareToDirect(link.URL, "")
fmt.Println(dl.URL)

// 回收站
items, _ = recycle.ListRecycle(c, 1)
recycle.RestoreFile(c, 123456)
recycle.ClearRecycle(c)
```
