package files

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lazgo/lazgo/pkg/client"
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

// TestIntegration_AllFileOps tests all file operations:
// Upload → List → Delete → Recycle → Restore → Delete permanently
func TestIntegration_AllFileOps(t *testing.T) {
	c := ensureLoggedIn(t)
	if c == nil {
		return
	}

	// ===== STEP 1: Create temp file =====
	t.Log("📄 Step 1: Creating temp file...")
	content := []byte("integration test content " + time.Now().Format("20060102150405"))
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "lazgo_test.txt")
	if err := ioutil.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	// Use t.Cleanup for guaranteed cleanup
	t.Cleanup(func() { os.Remove(tmpFile) })
	t.Logf("   ✅ Created: %s", tmpFile)

	// ===== STEP 2: Upload =====
	t.Log("📄 Step 2: Uploading file...")
	info, err := Upload(c, tmpFile, -1)
	if err != nil {
		t.Fatalf("❌ Upload failed: %v", err)
	}
	if info.ID == "" {
		t.Fatal("❌ Upload returned empty ID")
	}
	t.Logf("   ✅ Uploaded: ID=%s, name=%s", info.ID, info.Name)

	// Remember fileID for cleanup
	fileID := parseID(info.ID)

	// ===== STEP 3: List files =====
	t.Log("📄 Step 3: Listing files...")
	files, err := ListFiles(c, -1, 1)
	if err != nil {
		t.Fatalf("❌ ListFiles failed: %v", err)
	}
	t.Logf("   ✅ Listed %d items", len(files))

	// ===== STEP 4: Soft delete =====
	t.Log("📄 Step 4: Deleting file (soft delete)...")
	err = DeleteFile(c, fileID)
	if err != nil {
		t.Fatalf("❌ DeleteFile failed: %v", err)
	}
	t.Log("   ✅ Soft deleted → to recycle bin")

	// ===== STEP 5: Check in recycle =====
	t.Log("📄 Step 5: Checking recycle bin...")
	items, err := recycle.ListRecycle(c, 1)
	if err != nil {
		t.Fatalf("❌ ListRecycle failed: %v", err)
	}
	found := false
	for _, item := range items {
		if !item.IsFolder && item.ID == info.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("❌ File not found in recycle bin")
	}
	t.Log("   ✅ Found in recycle bin")

	// ===== STEP 6: Restore =====
	t.Log("📄 Step 6: Restoring file...")
	err = recycle.RestoreFile(c, fileID)
	if err != nil {
		t.Fatalf("❌ RestoreFile failed: %v", err)
	}
	t.Log("   ✅ Restored")

	// ===== STEP 7: Permanent delete =====
	t.Log("📄 Step 7: Permanent delete from recycle...")
	err = recycle.DeleteRecycleFile(c, fileID)
	if err != nil {
		t.Fatalf("❌ Permanent delete failed: %v", err)
	}
	t.Log("   ✅ Permanently deleted")

	t.Log("✅ All file operations completed!")
}

// parseID extracts numeric ID from string.
func parseID(s string) int {
	var id int
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			id = id*10 + int(ch-'0')
		}
	}
	return id
}