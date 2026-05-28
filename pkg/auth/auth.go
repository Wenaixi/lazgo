package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/lazgo/lazgo/internal/utils"
	"github.com/lazgo/lazgo/pkg/client"
)

const (
	loginURL  = "https://accounts.woozooo.com/accounts.php"
	loginPage = loginURL + "?action=login&ref=up.woozooo.com"
)

var saveCookieNames = []string{"phpdisk_info", "ylogin", "uag", "PHPSESSID"}

// Login performs login and returns cookies on success.
func Login(username, password string) (map[string]string, error) {
	httpClient := &http.Client{}
	allCookies := make(map[string]string)

	// Step 1: GET login page
	req, err := http.NewRequest("GET", loginPage, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", client.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求登录页失败: %w", err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("读取登录页失败: %w", err)
	}

	// Extract acw_tc
	for name, val := range utils.CollectSetCookies(resp.Header, "acw_tc") {
		allCookies[name] = val
	}

	// Handle ACW SC V2
	if strings.Contains(string(body), "var arg1='") {
		acw, err := utils.ACWSCV2(string(body))
		if err != nil {
			return nil, fmt.Errorf("ACW 计算失败: %w", err)
		}
		allCookies["acw_sc__v2"] = acw

		// Retry with cookie
		req2, err := http.NewRequest("GET", loginPage, nil)
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %w", err)
		}
		req2.Header.Set("User-Agent", client.UserAgent)
		req2.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req2.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
		req2.Header.Set("Cookie", buildCookie(allCookies))
		resp2, err := httpClient.Do(req2)
		if err == nil {
			resp2.Body.Close()
		}
	}

	// Step 2: POST login
	form := url.Values{}
	form.Set("task", "uselogin")
	form.Set("username", username)
	form.Set("password", password)
	form.Set("ref", "up.woozooo.com")

	req3, err := http.NewRequest("POST", loginURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req3.Header.Set("User-Agent", client.UserAgent)
	req3.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req3.Header.Set("X-Requested-With", "XMLHttpRequest")
	req3.Header.Set("Accept", "application/json, text/javascript, */*; q=0.9")
	req3.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req3.Header.Set("Referer", loginPage)
	req3.Header.Set("Origin", "https://accounts.woozooo.com")
	if len(allCookies) > 0 {
		req3.Header.Set("Cookie", buildCookie(allCookies))
	}

	resp3, err := httpClient.Do(req3)
	if err != nil {
		return nil, fmt.Errorf("登录请求失败: %w", err)
	}
	defer resp3.Body.Close()

	// Extract auth cookies
	for name, val := range utils.CollectSetCookies(resp3.Header, saveCookieNames...) {
		allCookies[name] = val
	}

	body3, err := io.ReadAll(resp3.Body)
	if err != nil {
		return nil, fmt.Errorf("读取登录响应失败: %w", err)
	}

	var result struct {
		Zt   int    `json:"zt"`
		Msgs string `json:"msgs"`
		Info string `json:"info"`
	}
	if err := json.Unmarshal(body3, &result); err != nil {
		return nil, fmt.Errorf("解析登录响应失败: %w", err)
	}

	if result.Zt == 1 {
		return allCookies, nil
	}
	msg := result.Msgs
	if msg == "" {
		msg = result.Info
	}
	if msg == "" {
		msg = "登录失败"
	}
	return nil, fmt.Errorf("%s", msg)
}

func buildCookie(cookies map[string]string) string {
	var parts []string
	for k, v := range cookies {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, "; ")
}
