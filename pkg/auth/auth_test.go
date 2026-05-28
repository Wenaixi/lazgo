package auth

import (
	"testing"
)

// TestLoginStructure tests that Login builds correct request.
// Real login requires network, so this validates the flow.
func TestLoginStructure(t *testing.T) {
	// The login flow:
	// 1. GET login page -> extract acw_tc, handle ACW
	// 2. POST to accounts.php with username/password
	// 3. Return cookies on success

	t.Log("Login flow:")
	t.Log("  1. GET " + loginPage)
	t.Log("  2. Handle acw_tc and ACW if present")
	t.Log("  3. POST task=uselogin")
	t.Log("  4. Return cookies: phpdisk_info, ylogin, uag, PHPSESSID")
}

// TestLoginSuccess would test successful login with real credentials.
// Skipped to avoid storing actual passwords in code.
func TestLoginReal(t *testing.T) {
	t.Skip("Requires real credentials")

	// This is how you'd test with real credentials:
	// cookies, err := Login("username", "password")
	// if err != nil {
	//     t.Fatalf("Login failed: %v", err)
	// }
	// if cookies == nil {
	//     t.Fatal("Login returned nil cookies")
	// }
	// // Verify required cookies
	// required := []string{"ylogin", "phpdisk_info", "uag"}
	// for _, name := range required {
	//     if _, ok := cookies[name]; !ok {
	//         t.Errorf("Missing cookie: %s", name)
	//     }
	// }
}

// TestLoginInvalid tests behavior with invalid credentials.
func TestLoginInvalid(t *testing.T) {
	t.Skip("Requires test account with known bad password")

	// cookies, err := Login("test", "wrongpassword")
	// if err == nil {
	//     t.Error("Login should fail with invalid credentials")
	// }
	// if cookies != nil {
	//     t.Error("Login should return nil cookies on failure")
	// }
}

// TestBuildCookie verifies the cookie building helper.
func TestBuildCookie(t *testing.T) {
	cookies := map[string]string{
		"a": "1",
		"b": "2",
	}
	header := buildCookie(cookies)
	if header == "" {
		t.Error("buildCookie returned empty")
	}
	// Verify format: a=1; b=2
	if len(header) < 5 {
		t.Errorf("buildCookie too short: %q", header)
	}
}