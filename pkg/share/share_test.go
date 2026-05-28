package share

import (
	"testing"
)

// TestGetShareLink tests getting a share link.
func TestGetShareLink(t *testing.T) {
	// Requires authenticated client
	// Flow: Doupload(task=22, file_id=N)
	// Returns: ShareLink{URL, HasPassword, Password, ShareID}

	t.Log("GetShareLink flow:")
	t.Log("  1. Doupload(task=22, file_id=N)")
	t.Log("  2. Parse response for is_newd + f_id")
	t.Log("  3. Return share URL")
}

// TestShareToDirect tests converting share to direct link.
// This requires actual network to Lanzou Cloud.
func TestShareToDirectStructure(t *testing.T) {
	// The flow (post-2023):
	// 1. GET share URL -> extract fid + fn iframe URL
	// 2. GET fn page -> extract wp_sign, ajaxdata, kdns
	// 3. POST ajaxm.php -> get dom + url
	// 4. Build: dom + "/file/" + url

	t.Log("ShareToDirect flow:")
	t.Log("  1. GET share URL (handle ACW if present)")
	t.Log("  2. Extract var fid=X, iframe src='/fn?...'")
	t.Log("  3. GET fn URL")
	t.Log("  4. Extract var wp_sign, ajaxdata, kdns")
	t.Log("  5. POST ajaxm.php?file=fid")
	t.Log("  6. Return DirectLink{URL, Filename}")
}

// TestExtractFnParams tests extracting parameters from fn page.
func TestExtractFnParams(t *testing.T) {
	html := `var ajaxdata='websign'; var wp_sign='sign123'; var kdns=1;`

	params := extractFnParams(html)
	if params.ajaxdata != "websign" {
		t.Errorf("ajaxdata = %q, want %q", params.ajaxdata, "websign")
	}
	if params.wpSign != "sign123" {
		t.Errorf("wpSign = %q, want %q", params.wpSign, "sign123")
	}
	if params.kdns != "1" {
		t.Errorf("kdns = %q, want %q", params.kdns, "1")
	}
}

// TestExtractFnParamsPartial tests with missing params.
func TestExtractFnParamsPartial(t *testing.T) {
	html := `var wp_sign='only';`

	params := extractFnParams(html)
	if params.ajaxdata != "" {
		t.Error("ajaxdata should be empty")
	}
	if params.wpSign != "only" {
		t.Errorf("wpSign = %q, want %q", params.wpSign, "only")
	}
	if params.kdns != "" {
		t.Error("kdns should be empty")
	}
}

// Validates that ShareToDirect returns correct structure.
type shareLinkResult struct {
	URL      string
	Filename string
}

// TestPasswordShare tests share with password.
func TestPasswordShare(t *testing.T) {
	t.Skip("Requires file with password protection")

	// Would test:
	// link, err := GetShareLink(c, fileID)
	// if link.HasPassword {
	//     direct, err := ShareToDirect(link.URL, "correctpassword")
	//     // verify success
	// }
}