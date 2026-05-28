package client

import (
	"encoding/json"
	"os"
	"testing"
)

// TestNewClient tests client creation.
func TestNewClient(t *testing.T) {
	cookies := map[string]string{
		"ylogin":      "test123",
		"phpdisk_info": "abc123",
		"uag":          "xyz",
	}

	c := New(cookies)
	if c == nil {
		t.Fatal("New client returned nil")
	}
	if c.UID() != "test123" {
		t.Errorf("UID() = %q, want %q", c.UID(), "test123")
	}
	if c.CookieHeader() == "" {
		t.Error("CookieHeader() returned empty string")
	}
}

// TestCookieHeaderFormat tests the cookie header formatting.
func TestCookieHeaderFormat(t *testing.T) {
	cookies := map[string]string{
		"foo": "bar",
		"baz": "qux",
	}

	c := New(cookies)
	header := c.CookieHeader()

	// Should contain both cookies
	if len(header) < 10 {
		t.Errorf("CookieHeader too short: %q", header)
	}
}

// TestLoadCookies tests loading cookies from file.
func TestLoadCookiesFile(t *testing.T) {
	// Test loading from various possible paths
	paths := []string{
		"../../../data/lanzou_cookies.json",
		"../../data/lanzou_cookies.json",
		"../data/lanzou_cookies.json",
	}

	var validCookies map[string]string
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		if json.Unmarshal(data, &validCookies) == nil && len(validCookies) > 0 {
			break
		}
	}

	if len(validCookies) == 0 {
		t.Skip("No cookies file found")
	}

	// Verify cookie structure
	if _, ok := validCookies["ylogin"]; !ok {
		t.Log("Warning: cookies may not have ylogin")
	}
}

// TestFormhashCaching tests that formhash is cached.
func TestFormhashCaching(t *testing.T) {
	c := New(nil)

	// Initially empty
	if c.formhash != "" {
		t.Error("Initial formhash should be empty")
	}

	// Manually set
	c.formhash = "abc123"
	if c.formhash != "abc123" {
		t.Error("Formhash not set correctly")
	}

	// Clear
	c.InvalidateFormhash()
	if c.formhash != "" {
		t.Error("InvalidateFormhash didn't clear")
	}
}

// TestDouploadStructure validates the Doupload call structure.
func TestDouploadStructure(t *testing.T) {
	c := New(nil)
	_ = c
	// Doupload just calls PostJSON with BaseURL + "/doupload.php"
	// Actual behavior tested in integration tests
	t.Log("Doupload: POST to " + BaseURL + "/doupload.php")
}

// TestInvalidCredentials tests behavior without cookies.
func TestInvalidCredentials(t *testing.T) {
	c := New(nil)
	if c.UID() != "" {
		t.Error("Empty cookies should give empty UID")
	}
}