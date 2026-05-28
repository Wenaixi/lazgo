package utils

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var acwPos = []int{
	15, 35, 29, 24, 33, 16, 1, 38, 10, 9, 19, 31, 40,
	27, 22, 23, 25, 13, 6, 11, 39, 18, 20, 8, 14, 21,
	32, 26, 2, 30, 7, 4, 17, 5, 3, 28, 34, 37, 12, 36,
}

const acwMask = "3000176000856006061501533003690027800375"

var arg1Re = regexp.MustCompile(`arg1='([^']+)'`)

// ACWSCV2 solves the ACW SC V2 anti-bot challenge and returns the cookie value.
func ACWSCV2(html string) (string, error) {
	match := arg1Re.FindStringSubmatch(html)
	if match == nil {
		return "", fmt.Errorf("未找到 arg1 参数")
	}
	arg1 := match[1]

	var q strings.Builder
	for _, pos := range acwPos {
		q.WriteByte(arg1[pos-1])
	}
	u := q.String()

	var v strings.Builder
	for i := 0; i < len(u); i += 2 {
		var byteU, byteM byte
		if _, err := fmt.Sscanf(u[i:i+2], "%x", &byteU); err != nil {
			return "", fmt.Errorf("ACW hex 解析失败: %w", err)
		}
		if _, err := fmt.Sscanf(acwMask[i:i+2], "%x", &byteM); err != nil {
			return "", fmt.Errorf("ACW mask 解析失败: %w", err)
		}
		v.WriteString(fmt.Sprintf("%02x", byteU^byteM))
	}
	return v.String(), nil
}

// FormatID safely converts a JSON number (float64) or string to a plain integer string.
func FormatID(v interface{}) string {
	switch n := v.(type) {
	case float64:
		return fmt.Sprintf("%.0f", n)
	case string:
		return n
	default:
		return fmt.Sprintf("%v", v)
	}
}

// StrVal extracts a string value from a map, returning def if missing or nil.
func StrVal(m map[string]interface{}, key, def string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return def
}

// ParseSetCookie extracts the name=value pair from a single Set-Cookie header value.
// Unlike Cookie headers, Set-Cookie attributes include Expires dates that contain commas,
// so we only extract the first name=value pair before any semicolon.
func ParseSetCookie(raw string) (string, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ""
	}
	// Extract name=value before the first ;
	if idx := strings.Index(raw, ";"); idx != -1 {
		raw = raw[:idx]
	}
	if idx := strings.Index(raw, "="); idx != -1 {
		return strings.TrimSpace(raw[:idx]), strings.TrimSpace(raw[idx+1:])
	}
	return "", ""
}

// CollectSetCookies iterates all Set-Cookie response headers and extracts matching cookie names.
func CollectSetCookies(header http.Header, names ...string) map[string]string {
	result := make(map[string]string)
	for _, sc := range header["Set-Cookie"] {
		name, value := ParseSetCookie(sc)
		if name == "" {
			continue
		}
		for _, want := range names {
			if name == want {
				result[name] = value
			}
		}
	}
	return result
}
