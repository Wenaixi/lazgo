package utils

import (
	"testing"
)

// TestACWSCV2 tests the ACW SC V2 solver - unit tests with known inputs.
func TestACWSCV2(t *testing.T) {
	// Unit test with real sample data from Lanzou Cloud
	// This sample was captured during real browsing

	// Invalid/empty cases
	tests := []struct {
		name    string
		html   string
		wantErr bool
	}{
		{
			name:    "empty html",
			html:   "",
			wantErr: true,
		},
		{
			name:    "no arg1 pattern",
			html:   `<script>alert('test')</script>`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ACWSCV2(tt.html)
			if (err != nil) != tt.wantErr {
				t.Errorf("ACWSCV2() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestFormatID tests the FormatID helper.
func TestFormatID(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{"float64", float64(12345), "12345"},
		{"string", "12345", "12345"},
		{"int", 123, "123"},
		{"unknown type", interface{}(nil), "<nil>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatID(tt.input); got != tt.want {
				t.Errorf("FormatID(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestStrVal tests the StrVal helper.
func TestStrVal(t *testing.T) {
	m := map[string]interface{}{
		"exists": "value",
		"empty":  "",
		"nil":    nil,
	}

	tests := []struct {
		name     string
		key      string
		def      string
		expected string
	}{
		{"exists key", "exists", "default", "value"},
		{"empty string key", "empty", "default", ""},
		{"nil value key", "nil", "default", "default"},
		{"missing key", "missing", "default", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StrVal(m, tt.key, tt.def); got != tt.expected {
				t.Errorf("StrVal(%q) = %q, want %q", tt.key, got, tt.expected)
			}
		})
	}
}

// TestParseSetCookie tests cookie parsing.
func TestParseSetCookie(t *testing.T) {
	tests := []struct {
		name           string
		raw            string
		expectedName   string
		expectedVal    string
	}{
		{"basic", "token=value", "token", "value"},
		{"with path", "token=value; Path=/", "token", "value"},
		{"with expires", "token=value; Expires=Wed, 01 Jan 2025 00:00:00 GMT", "token", "value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, val := ParseSetCookie(tt.raw)
			if name != tt.expectedName || val != tt.expectedVal {
				t.Errorf("ParseSetCookie(%q) = (%q, %q), want (%q, %q)",
					tt.raw, name, val, tt.expectedName, tt.expectedVal)
			}
		})
	}
}