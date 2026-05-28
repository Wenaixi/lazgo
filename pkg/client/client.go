package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/lazgo/lazgo/internal/utils"
)

var errNeedACW = errors.New("ACW challenge detected")

const (
	BaseURL   = "https://pc.woozooo.com"
	UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/148.0.0.0 Safari/537.36"
)

var formhashRe = regexp.MustCompile(`name="formhash" value="([a-f0-9]+)"`)

// Client is the HTTP client for Lanzou Cloud operations.
type Client struct {
	httpClient *http.Client
	Cookies    map[string]string
	uid        string
	formhash   string
}

// APIResponse is a generic Lanzou API JSON response.
type APIResponse struct {
	Zt   int             `json:"zt"`
	Info json.RawMessage `json:"info"`
	Text json.RawMessage `json:"text"`
	Msgs string          `json:"msgs"`
	Inf  string          `json:"inf"`
	Dom  string          `json:"dom"`
	URL  string          `json:"url"`
}

// New creates a new Client with the given cookies.
func New(cookies map[string]string) *Client {
	c := &Client{
		httpClient: &http.Client{},
		Cookies:    make(map[string]string),
	}
	if cookies != nil {
		for k, v := range cookies {
			c.Cookies[k] = v
		}
	}
	if u, ok := c.Cookies["ylogin"]; ok {
		c.uid = u
	}
	return c
}

// UID returns the current user ID from cookies.
func (c *Client) UID() string {
	return c.uid
}

// HTTPClient returns the underlying http.Client for custom requests.
func (c *Client) HTTPClient() *http.Client {
	return c.httpClient
}

// CookieHeader returns the Cookie header string.
func (c *Client) CookieHeader() string {
	var parts []string
	for k, v := range c.Cookies {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, "; ")
}

// PostJSON POSTs form data and expects a JSON response. Returns error on non-zt=1/2.
func (c *Client) PostJSON(urlStr string, data map[string]string, referer string) (*APIResponse, error) {
	form := url.Values{}
	for k, v := range data {
		form.Set(k, v)
	}

	req, err := http.NewRequest("POST", urlStr, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", c.CookieHeader())
	if referer != "" {
		req.Header.Set("Referer", referer)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}
	var result APIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("JSON 解析失败: %w (body: %s)", err, string(body[:min(len(body), 200)]))
	}
	if result.Zt != 1 && result.Zt != 2 {
		msg := result.Inf
		if msg == "" {
			msg = result.Msgs
		}
		if msg == "" {
			var infoStr string
			if err := json.Unmarshal(result.Info, &infoStr); err == nil && infoStr != "" && infoStr != "0" {
				msg = infoStr
			}
		}
		if msg == "" || msg == "0" {
			msg = "未知错误"
		}
		return nil, fmt.Errorf("API 错误 (code=%d): %s", result.Zt, msg)
	}
	return &result, nil
}

// GetHTML performs a GET request, handling ACW anti-bot.
func (c *Client) GetHTML(urlStr string, referer string) (*http.Response, error) {
	resp, body, err := doGetRead(c.httpClient, urlStr, referer, c.CookieHeader())
	if err != nil {
		return nil, err
	}

	if strings.Contains(body, "var arg1='") {
		acw, acwErr := utils.ACWSCV2(body)
		if acwErr == nil {
			c.Cookies["acw_sc__v2"] = acw
			resp.Body.Close()
			resp2, doErr := doGet(c.httpClient, urlStr, referer, c.CookieHeader())
			if doErr != nil {
				return nil, doErr
			}
			return resp2, nil
		}
	}
	return resp, nil
}

// PostPage POSTs form data expecting an HTML response.
func (c *Client) PostPage(urlStr string, data map[string]string, referer string) (*http.Response, error) {
	form := url.Values{}
	for k, v := range data {
		form.Set(k, v)
	}

	req, err := http.NewRequest("POST", urlStr, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", c.CookieHeader())
	req.Header.Set("Origin", "https://pc.woozooo.com")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	bodyBytes, readErr := io.ReadAll(resp.Body)
	resp.Body.Close()
	if readErr != nil {
		return nil, fmt.Errorf("读取响应失败: %w", readErr)
	}
	body := string(bodyBytes)

	if strings.Contains(body, "var arg1='") {
		acw, acwErr := utils.ACWSCV2(body)
		if acwErr != nil {
			return nil, fmt.Errorf("ACW 计算失败: %w", acwErr)
		}
		c.Cookies["acw_sc__v2"] = acw
		// Retry with ACW cookie
		req2, err2 := http.NewRequest("POST", urlStr, strings.NewReader(form.Encode()))
		if err2 != nil {
			return nil, fmt.Errorf("创建请求失败: %w", err2)
		}
		req2.Header.Set("User-Agent", UserAgent)
		req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req2.Header.Set("Cookie", c.CookieHeader())
		req2.Header.Set("Origin", "https://pc.woozooo.com")
		if referer != "" {
			req2.Header.Set("Referer", referer)
		}
		return c.httpClient.Do(req2)
	}

	// Rebuild response body so caller can read it
	resp.Body = io.NopCloser(strings.NewReader(body))
	return resp, nil
}

// Doupload calls doupload.php with the given data.
func (c *Client) Doupload(data map[string]string) (*APIResponse, error) {
	return c.PostJSON(
		BaseURL+"/doupload.php",
		data,
		BaseURL+"/mydisk.php?item=files&action=index&u="+c.uid,
	)
}

// FetchFormhash fetches and caches the formhash from the recycle page.
func (c *Client) FetchFormhash() (string, error) {
	if c.formhash != "" {
		return c.formhash, nil
	}
	resp, body, err := doGetRead(c.httpClient,
		BaseURL+"/mydisk.php?item=recycle&action=files",
		BaseURL+"/mydisk.php?item=files&action=index",
		c.CookieHeader(),
	)
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	match := formhashRe.FindStringSubmatch(body)
	if match == nil {
		return "", fmt.Errorf("未找到 formhash")
	}
	c.formhash = match[1]
	return c.formhash, nil
}

// InvalidateFormhash clears the cached formhash.
func (c *Client) InvalidateFormhash() {
	c.formhash = ""
}

// ---- internal helpers ----

func doGet(httpClient *http.Client, urlStr, referer, cookie string) (*http.Response, error) {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Cookie", cookie)
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	return resp, nil
}

func doGetRead(httpClient *http.Client, urlStr, referer, cookie string) (*http.Response, string, error) {
	resp, err := doGet(httpClient, urlStr, referer, cookie)
	if err != nil {
		return nil, "", err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return nil, "", fmt.Errorf("读取响应失败: %w", err)
	}
	return resp, string(body), nil
}
