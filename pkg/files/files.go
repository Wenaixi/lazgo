package files

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/lazgo/lazgo/internal/utils"
	"github.com/lazgo/lazgo/pkg/client"
	"github.com/lazgo/lazgo/pkg/models"
)

// Upload uploads a file and returns FileInfo.
func Upload(c *client.Client, filePath string, folderID int) (*models.FileInfo, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法打开文件: %w", err)
	}
	defer f.Close()

	fileName := filepath.Base(filePath)

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	fw, err := w.CreateFormFile("upload_file", fileName)
	if err != nil {
		return nil, fmt.Errorf("创建上传表单失败: %w", err)
	}
	if _, err := io.Copy(fw, f); err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	w.WriteField("task", "1")
	w.WriteField("folder_id", strconv.Itoa(folderID))
	w.Close()

	req, err := http.NewRequest("POST", client.BaseURL+"/html5up.php", &buf)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", client.UserAgent)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Referer", client.BaseURL+"/mydisk.php?item=files&action=index&u="+c.UID())
	req.Header.Set("Cookie", c.CookieHeader())
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.HTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("上传失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}
	var result client.APIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	if result.Zt != 1 {
		return nil, fmt.Errorf("上传失败: %s", result.Inf)
	}

	var items []map[string]interface{}
	if err := json.Unmarshal(result.Text, &items); err != nil {
		return nil, fmt.Errorf("解析文件列表失败: %w", err)
	}
	if len(items) == 0 {
		return &models.FileInfo{Name: fileName}, nil
	}
	item := items[0]
	return &models.FileInfo{
		ID:   utils.FormatID(item["id"]),
		Name: utils.StrVal(item, "name_all", fileName),
		Size: utils.StrVal(item, "size", ""),
	}, nil
}

// DeleteFile deletes a file by ID.
func DeleteFile(c *client.Client, fileID int) error {
	_, err := c.Doupload(map[string]string{
		"task":    "6",
		"file_id": strconv.Itoa(fileID),
	})
	return err
}

// ListFiles lists files in a folder.
func ListFiles(c *client.Client, folderID, page int) ([]models.FileInfo, error) {
	result, err := c.Doupload(map[string]string{
		"task":      "5",
		"folder_id": strconv.Itoa(folderID),
		"pg":        strconv.Itoa(page),
	})
	if err != nil {
		return nil, err
	}

	var items []map[string]interface{}
	if err := json.Unmarshal(result.Text, &items); err != nil {
		return nil, fmt.Errorf("解析文件列表失败: %w", err)
	}

	var files []models.FileInfo
	for _, f := range items {
		files = append(files, models.FileInfo{
			ID:       utils.FormatID(f["id"]),
			Name:     utils.StrVal(f, "name_all", utils.StrVal(f, "name", "")),
			Size:     utils.StrVal(f, "size", ""),
			Time:     utils.StrVal(f, "time", ""),
			IsFolder: utils.StrVal(f, "onof", "0") == "1",
		})
	}
	return files, nil
}
