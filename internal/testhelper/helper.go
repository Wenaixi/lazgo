package testhelper

import (
	"encoding/json"
	"os"

	"github.com/lazgo/lazgo/pkg/client"
	"github.com/lazgo/lazgo/pkg/files"
	"github.com/lazgo/lazgo/pkg/folders"
	"github.com/lazgo/lazgo/pkg/recycle"
)

// TestFolderIDs holds IDs of test folders created during tests.
var TestFolderIDs []int

// TestFileIDs holds IDs of test files created during tests.
var TestFileIDs []int

// LoadTestClient loads cookies from file and creates a test client.
// Returns nil if no cookies - caller should skip test.
func LoadTestClient(t T) *client.Client {
	t.Helper()

	// Try multiple possible cookie file locations
	paths := []string{
		"../../../data/lanzou_cookies.json",
		"../../data/lanzou_cookies.json",
		"../data/lanzou_cookies.json",
		"data/lanzou_cookies.json",
	}

	var cookies map[string]string

	for _, p := range paths {
		if _, statErr := os.Stat(p); statErr == nil {
			if data, readErr := os.ReadFile(p); readErr == nil {
				json.Unmarshal(data, &cookies)
			}
		}
	}

	if len(cookies) == 0 {
		t.Skip("No cookies found, skipping integration test")
		return nil
	}

	c := client.New(cookies)
	if c.UID() == "" {
		t.Skip("Invalid cookies, skipping integration test")
		return nil
	}

	return c
}

// CreateTestFolder creates a temporary test folder.
// The ID is tracked for cleanup.
func CreateTestFolder(t T, c *client.Client, name string) int {
	t.Helper()

	info, err := folders.CreateFolder(c, name, -1, "test")
	if err != nil {
		t.Fatalf("Failed to create test folder: %v", err)
	}

	var folderID int
	for _, ch := range info.ID {
		if ch >= '0' && ch <= '9' {
			folderID = folderID*10 + int(ch-'0')
		}
	}

	TestFolderIDs = append(TestFolderIDs, folderID)
	return folderID
}

// CreateTestFile creates a temporary test file content.
// Writes to temp file and returns path (for manual upload).
func CreateTestFile(t T, content string) string {
	t.Helper()

	tmpFile := "/tmp/lazgo_test_" + string(rune(os.Getpid())) + ".txt"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	return tmpFile
}

// TrackFileID tracks a file ID for cleanup.
func TrackFileID(t T, id int) {
	t.Helper()
	TestFileIDs = append(TestFileIDs, id)
}

// CleanupTestResources deletes all tracked test resources.
func CleanupTestResources(t T, c *client.Client) {
	t.Helper()

	// Soft delete files first
	for _, id := range TestFileIDs {
		files.DeleteFile(c, id)
	}

	// Soft delete folders
	for _, id := range TestFolderIDs {
		folders.DeleteFolder(c, id)
	}

	// Then permanently delete from recycle
	for _, id := range TestFileIDs {
		recycle.DeleteRecycleFile(c, id)
	}
	for _, id := range TestFolderIDs {
		recycle.DeleteRecycleFolder(c, id)
	}

	TestFileIDs = nil
	TestFolderIDs = nil
}

// T is a subset of testing.TB for compatibility.
type T interface {
	Helper()
	Skip(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}