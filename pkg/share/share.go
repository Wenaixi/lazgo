package share

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/lazgo/lazgo/internal/utils"
	"github.com/lazgo/lazgo/pkg/client"
	"github.com/lazgo/lazgo/pkg/models"
)

// GetShareLink gets the share link for a file.
func GetShareLink(c *client.Client, fileID int) (*models.ShareLink, error) {
	result, err := c.Doupload(map[string]string{
		"task":    "22",
		"file_id": fmt.Sprintf("%d", fileID),
	})
	if err != nil {
		return nil, err
	}

	var info map[string]interface{}
	if err := json.Unmarshal(result.Info, &info); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	isNewd := utils.StrVal(info, "is_newd", "")
	fID := utils.StrVal(info, "f_id", "")
	shareURL := ""
	if isNewd != "" && fID != "" {
		shareURL = isNewd + "/" + fID
	}

	return &models.ShareLink{
		URL:         shareURL,
		HasPassword: utils.StrVal(info, "onof", "0") == "1",
		Password:    utils.StrVal(info, "pwd", ""),
		ShareID:     fID,
	}, nil
}

// fnPageParams holds extracted data from the fn page HTML.
type fnPageParams struct {
	ajaxdata string // websignkey / signs value
	wpSign   string // pre-computed sign
	kdns     string // kd parameter
}

var (
	fidRe       = regexp.MustCompile(`var fid\s*=\s*(\d+)`)
	iframeSrcRe = regexp.MustCompile(`src="(/fn\?[^"]+)"`)
	wpSignRe    = regexp.MustCompile(`var wp_sign\s*=\s*'([^']+)'`)
	ajaxdataRe  = regexp.MustCompile(`var ajaxdata\s*=\s*'([^']+)'`)
	kdnsRe      = regexp.MustCompile(`var kdns\s*=\s*(\d+)`)
	pwdGateRe   = regexp.MustCompile(`id="pwd"|id='pwd'`)
)

// ShareToDirect converts a share link to a direct download link.
//
// New flow (post-2023 update):
//  1. GET share URL → handle ACW
//  2. Extract fid + fn iframe URL from main page
//  3. GET fn page → extract wp_sign, ajaxdata, kdns
//  4. POST ajaxm.php?file=<fid> with sign params
//  5. Build direct link from response dom + url
func ShareToDirect(shareURL string, password string) (*models.DirectLink, error) {
	httpClient := &http.Client{}

	// ---- Step 1: GET share page, handle ACW ----
	resp, err := doGet(httpClient, shareURL)
	if err != nil {
		return nil, fmt.Errorf("访问分享页失败: %w", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return nil, fmt.Errorf("读取分享页失败: %w", err)
	}
	resp.Body.Close()
	html := string(body)

	if strings.Contains(html, "var arg1='") {
		acw, err := utils.ACWSCV2(html)
		if err != nil {
			return nil, fmt.Errorf("ACW 计算失败: %w", err)
		}
		parsed, _ := url.Parse(resp.Request.URL.String())
		req, err := http.NewRequest("GET", shareURL, nil)
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %w", err)
		}
		req.Header.Set("User-Agent", client.UserAgent)
		req.AddCookie(&http.Cookie{Name: "acw_sc__v2", Value: acw, Domain: parsed.Host})
		resp2, err := httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("ACW 验证后访问失败: %w", err)
		}
		b2, err := io.ReadAll(resp2.Body)
	if err != nil {
		resp2.Body.Close()
		return nil, fmt.Errorf("读取ACW验证后页面失败: %w", err)
	}
	resp2.Body.Close()
		html = string(b2)
	}

	currentURL := resp.Request.URL.String()
	parsed, _ := url.Parse(currentURL)
	baseDomain := fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)

	// ---- Step 2: Extract fid + fn iframe URL ----
	m := fidRe.FindStringSubmatch(html)
	if m == nil {
		return nil, fmt.Errorf("无法从页面提取文件ID")
	}
	fid := m[1]

	hasPassword := pwdGateRe.MatchString(html)
	if hasPassword && password == "" {
		return nil, fmt.Errorf("文件需要密码，请提供密码")
	}

	m = iframeSrcRe.FindStringSubmatch(html)
	if m == nil {
		return nil, fmt.Errorf("无法从页面提取fn地址")
	}
	fnPath := m[1]

	// ---- Step 3: GET fn page ----
	fnURL := baseDomain + fnPath
	fnHTML, err := doGetString(httpClient, fnURL)
	if err != nil {
		return nil, fmt.Errorf("获取fn页面失败: %w", err)
	}

	params := extractFnParams(fnHTML)

	// ---- Step 4: POST ajaxm.php ----
	form := url.Values{}
	form.Set("action", "downprocess")
	form.Set("websignkey", params.ajaxdata)
	form.Set("signs", params.ajaxdata)
	form.Set("sign", params.wpSign)
	form.Set("websign", "")
	form.Set("kd", params.kdns)
	form.Set("ves", "1")
	if password != "" {
		form.Set("p", password)
	}

	ajaxURL := fmt.Sprintf("%s/ajaxm.php?file=%s", baseDomain, fid)
	req, err := http.NewRequest("POST", ajaxURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", client.UserAgent)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Referer", fnURL)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp3, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求直链失败: %w", err)
	}
	defer resp3.Body.Close()

	body3, err := io.ReadAll(resp3.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}
	var result struct {
		Zt  int    `json:"zt"`
		Inf interface{} `json:"inf"`
		Dom string `json:"dom"`
		URL string `json:"url"`
	}
	if err := json.Unmarshal(body3, &result); err != nil {
		return nil, fmt.Errorf("JSON 解析失败: %w", err)
	}

	if result.Zt != 1 {
		msg := fmt.Sprintf("%v", result.Inf)
		if msg == "0" || msg == "" {
			msg = "密码错误或获取失败"
		}
		return nil, fmt.Errorf("%s", msg)
	}

	// ---- Step 5: Build direct link ----
	directURL := ""
	if result.Dom != "" && result.URL != "" {
		directURL = result.Dom + "/file/" + result.URL
	}

	filename := ""
	if s, ok := result.Inf.(string); ok {
		filename = s
	}

	return &models.DirectLink{
		URL:      directURL,
		Filename: filename,
		Method:   "ajax_downprocess",
	}, nil
}

func extractFnParams(html string) fnPageParams {
	var p fnPageParams
	if m := ajaxdataRe.FindStringSubmatch(html); m != nil {
		p.ajaxdata = m[1]
	}
	if m := wpSignRe.FindStringSubmatch(html); m != nil {
		p.wpSign = m[1]
	}
	if m := kdnsRe.FindStringSubmatch(html); m != nil {
		p.kdns = m[1]
	}
	return p
}

func doGet(httpClient *http.Client, urlStr string) (*http.Response, error) {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", client.UserAgent)
	return httpClient.Do(req)
}

func doGetString(httpClient *http.Client, urlStr string) (string, error) {
	resp, err := doGet(httpClient, urlStr)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
