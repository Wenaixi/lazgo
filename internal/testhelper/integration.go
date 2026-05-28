package testhelper

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/lazgo/lazgo/pkg/client"
)

// NeedLogin prompts user for confirmation before running integration tests.
// Returns true if user confirms, false to skip.
func NeedLogin(t T) bool {
	t.Helper()

	fmt.Printf("\n⚠️  Integration test will make REAL API calls to Lanzou Cloud.\n")
	fmt.Printf("   This will use your saved cookies: data/lanzou_cookies.json\n\n")
	fmt.Printf("Continue? [y/N]: ")

	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))

	if response != "y" && response != "yes" {
		t.Skip("User declined integration test")
		return false
	}

	// Verify cookies still work
	cookies, err := loadCookies()
	if err != nil || len(cookies) == 0 {
		t.Skip("No cookies found")
		return false
	}

	c := client.New(cookies)
	if c.UID() == "" {
		t.Skip("Invalid cookies - please re-login")
		return false
	}

	fmt.Printf("✓ Logged in as: %s\n\n", c.UID())
	return true
}

// EnsureLoggedIn checks cookies and returns client, or skips test.
func EnsureLoggedIn(t T) *client.Client {
	t.Helper()

	cookies, err := loadCookies()
	if err != nil || len(cookies) == 0 {
		t.Skip("No cookies - run: lazgo login -u <user> -p <pass>")
		return nil
	}

	c := client.New(cookies)
	if c.UID() == "" {
		t.Skip("Invalid cookies - please re-login")
		return nil
	}

	return c
}

// SimpleCheck is a simple "skip if no cookies" check.
func SimpleCheck(t T) {
	t.Helper()

	cookies, err := loadCookies()
	if err != nil || len(cookies) == 0 {
		t.Skip("No cookies found")
	}
}

func loadCookies() (map[string]string, error) {
	paths := []string{
		"../../../data/lanzou_cookies.json",
		"../../data/lanzou_cookies.json",
		"../data/lanzou_cookies.json",
		"data/lanzou_cookies.json",
	}

	for _, p := range paths {
		if _, statErr := os.Stat(p); statErr == nil {
			if data, readErr := os.ReadFile(p); readErr == nil {
				var cookies map[string]string
				if json.Unmarshal(data, &cookies) == nil && len(cookies) > 0 {
					return cookies, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("no cookies")
}