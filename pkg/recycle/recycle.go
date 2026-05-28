package recycle

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lazgo/lazgo/pkg/client"
	"github.com/lazgo/lazgo/pkg/models"
)

var htmlRowRe = regexp.MustCompile(`(?s)<tr[^>]*>.*?</tr>`)

var (
	fidRe      = regexp.MustCompile(`file_restore&file_id=(\d+)`)
	folderIDRe = regexp.MustCompile(`folder_restore&folder_id=(\d+)`)
	nameRe     = regexp.MustCompile(`<img[^>]*/>\s*([^<\s][^<]*)</a>`)
	nameRe2    = regexp.MustCompile(`<a[^>]*>([^<]+)</a>`)
	timeRe     = regexp.MustCompile(`(\d{4}-\d{2}-\d{2})`)
	relTimeRe  = regexp.MustCompile(`(\d+\s*(?:天|小时|分钟|秒)前)`)
)

// ListRecycle lists files and folders in the recycle bin by parsing the HTML page.
func ListRecycle(c *client.Client, page int) ([]models.FileInfo, error) {
	// Cache-busting timestamp to avoid server-side caching
	urlStr := fmt.Sprintf("%s/mydisk.php?item=recycle&action=files&_=%d", client.BaseURL, time.Now().UnixNano())
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", client.UserAgent)
	req.Header.Set("Cookie", c.CookieHeader())
	req.Header.Set("Cache-Control", "no-cache, no-store")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", client.BaseURL+"/mydisk.php?item=files&action=index&u="+c.UID())
	resp, err := c.HTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}
	body := string(raw)
	var items []models.FileInfo
	for _, row := range htmlRowRe.FindAllString(body, -1) {
		if m := fidRe.FindStringSubmatch(row); m != nil {
			items = append(items, models.FileInfo{
				ID:   m[1],
				Name: extractName(row),
				Time: extractTime(row),
			})
		}
		if m := folderIDRe.FindStringSubmatch(row); m != nil {
			items = append(items, models.FileInfo{
				ID:       m[1],
				Name:     extractName(row),
				IsFolder: true,
			})
		}
	}
	return items, nil
}

func extractName(row string) string {
	// Name appears after <img .../> inside the file link <a>
	if m := nameRe.FindStringSubmatch(row); m != nil {
		return strings.TrimSpace(m[1])
	}
	// Fallback: try any <a>text</a>
	if m := nameRe2.FindStringSubmatch(row); m != nil {
		return strings.TrimSpace(m[1])
	}
	return ""
}

func extractTime(row string) string {
	if m := timeRe.FindStringSubmatch(row); m != nil {
		return m[0]
	}
	if m := relTimeRe.FindStringSubmatch(row); m != nil {
		return m[0]
	}
	return ""
}

// recycleAction performs a recycle operation (restore or delete) for files or folders.
func recycleAction(c *client.Client, action, idField string, id int) error {
	formhash, err := c.FetchFormhash()
	if err != nil {
		return err
	}

	idStr := strconv.Itoa(id)
	form := make(map[string]string)
	form["action"] = action
	form["task"] = action
	form[idField] = idStr
	form["ref"] = client.BaseURL + "/mydisk.php?item=recycle&action=files"
	form["formhash"] = formhash

	return postRecycleForm(c, form)
}

// batchAction performs a batch recycle operation (restore_all or delete_all).
func batchAction(c *client.Client, action string) error {
	formhash, err := c.FetchFormhash()
	if err != nil {
		return err
	}

	form := make(map[string]string)
	form["action"] = action
	form["task"] = action
	form["formhash"] = formhash

	return postRecycleForm(c, form)
}

func postRecycleForm(c *client.Client, data map[string]string) error {
	// Build encoded form with fixed field order (matching browser)
	var parts []string
	for _, k := range []string{"action", "task"} {
		if v, ok := data[k]; ok {
			parts = append(parts, url.QueryEscape(k)+"="+url.QueryEscape(v))
		}
	}
	// Remaining fields (file_id/folder_id, ref, formhash)
	for _, k := range []string{"file_id", "folder_id", "ref", "formhash"} {
		if v, ok := data[k]; ok {
			parts = append(parts, url.QueryEscape(k)+"="+url.QueryEscape(v))
		}
	}
	body := strings.Join(parts, "&")

	// Build referer matching the action
	referer := client.BaseURL + "/mydisk.php?item=recycle&action=" + data["action"]
	if v, ok := data["file_id"]; ok {
		referer += "&file_id=" + v
	} else if v, ok := data["folder_id"]; ok {
		referer += "&folder_id=" + v
	}

	req, err := http.NewRequest("POST",
		client.BaseURL+"/mydisk.php?item=recycle",
		strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", client.UserAgent)
	req.Header.Set("Cookie", c.CookieHeader())
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", "https://pc.woozooo.com")
	req.Header.Set("Referer", referer)

	resp, err := c.HTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		return fmt.Errorf("操作失败 (HTTP %d)", resp.StatusCode)
	}
	text := string(respBody)
	if !strings.Contains(text, "成功") && !strings.Contains(text, "zt\":1") {
		return fmt.Errorf("操作失败")
	}
	c.InvalidateFormhash()
	return nil
}

// RestoreFile restores a single file from recycle bin.
func RestoreFile(c *client.Client, fileID int) error {
	return recycleAction(c, "file_restore", "file_id", fileID)
}

// DeleteRecycleFile permanently deletes a single file from recycle bin.
func DeleteRecycleFile(c *client.Client, fileID int) error {
	return recycleAction(c, "file_delete_complete", "file_id", fileID)
}

// RestoreFolder restores a single folder from recycle bin.
func RestoreFolder(c *client.Client, folderID int) error {
	return recycleAction(c, "folder_restore", "folder_id", folderID)
}

// DeleteRecycleFolder permanently deletes a single folder from recycle bin.
func DeleteRecycleFolder(c *client.Client, folderID int) error {
	return recycleAction(c, "folder_delete_complete", "folder_id", folderID)
}

// ClearRecycle clears the entire recycle bin.
func ClearRecycle(c *client.Client) error {
	return batchAction(c, "delete_all")
}

// RestoreAll restores all files from recycle bin.
func RestoreAll(c *client.Client) error {
	return batchAction(c, "restore_all")
}
