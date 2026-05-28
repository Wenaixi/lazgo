package folders

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/lazgo/lazgo/internal/utils"
	"github.com/lazgo/lazgo/pkg/client"
	"github.com/lazgo/lazgo/pkg/models"
)

// CreateFolder creates a new folder.
func CreateFolder(c *client.Client, name string, parentID int, description string) (*models.FolderInfo, error) {
	result, err := c.Doupload(map[string]string{
		"task":               "2",
		"parent_id":          strconv.Itoa(parentID),
		"folder_name":        name,
		"folder_description": description,
	})
	if err != nil {
		return nil, err
	}

	var textID string
	if err := json.Unmarshal(result.Text, &textID); err != nil {
		return nil, fmt.Errorf("解析文件夹ID失败: %w", err)
	}

	return &models.FolderInfo{
		ID:          textID,
		Name:        name,
		Description: description,
	}, nil
}

// DeleteFolder deletes a folder by ID.
func DeleteFolder(c *client.Client, folderID int) error {
	_, err := c.Doupload(map[string]string{
		"task":      "3",
		"folder_id": strconv.Itoa(folderID),
	})
	return err
}

// GetFolderInfo returns folder information.
func GetFolderInfo(c *client.Client, folderID int) (*models.FolderInfo, error) {
	result, err := c.Doupload(map[string]string{
		"task":      "47",
		"folder_id": strconv.Itoa(folderID),
	})
	if err != nil {
		return nil, err
	}

	var items []map[string]interface{}
	raw := result.Info
	if len(raw) == 0 || string(raw) == "null" {
		raw = result.Text
	}
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("解析文件夹信息失败: %w", err)
	}

	if len(items) == 0 {
		return &models.FolderInfo{}, nil
	}
	f := items[0]
	desc := utils.StrVal(f, "folder_des", "")
	return &models.FolderInfo{
		ID:          utils.FormatID(f["folderid"]),
		Name:        utils.StrVal(f, "name", ""),
		Description: desc,
	}, nil
}
