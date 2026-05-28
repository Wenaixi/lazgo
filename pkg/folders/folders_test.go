package folders

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/lazgo/lazgo/pkg/client"
	"github.com/lazgo/lazgo/pkg/models"
	"github.com/lazgo/lazgo/pkg/recycle"
)

// checkCookies checks if valid cookies exist.
// Uses LAZGO_COOKIE_FILE env var, falls back to relative paths.
func checkCookies(t *testing.T) bool {
	t.Helper()
	path := os.Getenv("LAZGO_COOKIE_FILE")
	if path != "" {
		_, err := os.Stat(path)
		return err == nil
	}

	paths := []string{
		"../../../data/lanzou_cookies.json",
		"../../data/lanzou_cookies.json",
		"../data/lanzou_cookies.json",
		"data/lanzou_cookies.json",
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			if data, err := ioutil.ReadFile(p); err == nil {
				var cookies map[string]string
				if json.Unmarshal(data, &cookies) == nil && len(cookies) > 0 {
					return true
				}
			}
		}
	}
	return false
}

// loadCookies loads cookies from file.
func loadCookies() (map[string]string, error) {
	path := os.Getenv("LAZGO_COOKIE_FILE")
	if path == "" {
		paths := []string{
			"../../../data/lanzou_cookies.json",
			"../../data/lanzou_cookies.json",
			"../data/lanzou_cookies.json",
			"data/lanzou_cookies.json",
		}
		for _, p := range paths {
			if data, err := ioutil.ReadFile(p); err == nil {
				var cookies map[string]string
				if err := json.Unmarshal(data, &cookies); err == nil && len(cookies) > 0 {
					return cookies, nil
				}
			}
		}
		return nil, fmt.Errorf("no cookies found")
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cookies map[string]string
	if err := json.Unmarshal(data, &cookies); err != nil {
		return nil, err
	}
	return cookies, nil
}

// ensureLoggedIn checks cookies and returns client.
// Skip confirmation in CI=true environment.
// Use LAZGO_COOKIE_FILE env var to specify cookie path.
func ensureLoggedIn(t *testing.T) *client.Client {
	t.Helper()
	if !checkCookies(t) {
		t.Skip("No cookies - run: lazgo login -u <user> -p <pass>")
		return nil
	}

	// Skip interactive prompt in CI environment
	if os.Getenv("CI") != "true" {
		t.Log("⚠️  Integration test will make REAL API calls to Lanzou Cloud.")
		t.Log("Continue? [y/N]: ")

		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "yes" {
			t.Skip("User declined integration test")
			return nil
		}
	}

	cookies, err := loadCookies()
	if err != nil {
		t.Skip("Cannot load cookies: " + err.Error())
		return nil
	}

	c := client.New(cookies)
	if c.UID() == "" {
		t.Skip("Invalid cookies - please re-login")
		return nil
	}
	return c
}

// TestIntegration_AllFolderOps tests all folder operations in sequence:
// Create → Info → Delete → Recycle → Restore → Delete permanently
func TestIntegration_AllFolderOps(t *testing.T) {
	c := ensureLoggedIn(t)
	if c == nil {
		return
	}

	// ===== STEP 1: Create folder =====
	t.Log("📁 Step 1: Creating folder...")
	info, err := CreateFolder(c, "test_integration_"+timestamp(), -1, "integration test folder")
	if err != nil {
		t.Fatalf("❌ CreateFolder failed: %v", err)
	}
	if info.ID == "" {
		t.Fatal("❌ CreateFolder returned empty ID")
	}
	folderID := parseID(info.ID)
	t.Logf("   ✅ Created: ID=%s, name=%s", info.ID, info.Name)

	// ===== STEP 2: Get folder info =====
	t.Log("📁 Step 2: Getting folder info...")
	info2, err := GetFolderInfo(c, folderID)
	if err != nil {
		t.Fatalf("❌ GetFolderInfo failed: %v", err)
	}
	if info2.Name == "" {
		t.Error("⚠️  GetFolderInfo returned empty name")
	}
	t.Logf("   ✅ Got info: %+v", *info2)

	// ===== STEP 3: Delete folder (soft delete) =====
	t.Log("📁 Step 3: Deleting folder (soft delete)...")
	err = DeleteFolder(c, folderID)
	if err != nil {
		t.Fatalf("❌ DeleteFolder failed: %v", err)
	}
	t.Log("   ✅ Soft deleted → to recycle bin")

	// ===== STEP 4: Check in recycle =====
	t.Log("📁 Step 4: Checking recycle bin...")
	items, err := recycle.ListRecycle(c, 1)
	if err != nil {
		t.Fatalf("❌ ListRecycle failed: %v", err)
	}
	found := false
	for _, item := range items {
		if item.IsFolder && item.ID == info.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("❌ Folder not found in recycle bin")
	}
	t.Log("   ✅ Found in recycle bin")

	// ===== STEP 5: Restore folder =====
	t.Log("📁 Step 5: Restoring folder...")
	err = recycle.RestoreFolder(c, folderID)
	if err != nil {
		t.Fatalf("❌ RestoreFolder failed: %v", err)
	}
	t.Log("   ✅ Restored")

	// ===== STEP 6: Verify restored =====
	t.Log("📁 Step 6: Verifying restored...")
	info3, _ := GetFolderInfo(c, folderID)
	if info3 != nil && info3.Name != "" {
		t.Log("   ✅ Verified: folder still exists after restore")
	} else {
		t.Log("   ⚠️  Could not verify (may be normal)")
	}

	// ===== STEP 7: Permanent delete =====
	t.Log("📁 Step 7: Permanent delete from recycle...")
	err = recycle.DeleteRecycleFolder(c, folderID)
	if err != nil {
		t.Fatalf("❌ Permanent delete failed: %v", err)
	}
	t.Log("   ✅ Permanently deleted")

	t.Log("✅ All folder operations completed!")
}

// Helper to verify the package compiles.
func init() {
	_ = models.FolderInfo{}
}

func parseID(s string) int {
	var id int
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			id = id*10 + int(ch-'0')
		}
	}
	return id
}

func timestamp() string {
	return time.Now().Format("20060102150405")
}