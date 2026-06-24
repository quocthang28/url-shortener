package handler

import (
	"strings"
	"testing"
)

func TestValidateURL(t *testing.T) {
	// overLengthURL is one byte past the cap; the cap itself is exercised by the
	// "at length limit" valid case below.
	overLengthURL := "https://example.com/" + strings.Repeat("a", maxURLLength)

	// atLimitURL is exactly maxURLLength bytes long and otherwise valid.
	const prefix = "https://example.com/"
	atLimitURL := prefix + strings.Repeat("a", maxURLLength-len(prefix))

	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		// valid
		{"plain http", "http://example.com", false},
		{"https with path", "https://codesubmit.io/library/react", false},
		{"https with port", "https://example.com:8443/path", false},
		{"surrounding whitespace trimmed", "  https://example.com/x  ", false},
		{"at length limit", atLimitURL, false},

		// length
		{"over length limit", overLengthURL, true},

		// parse failure
		{"unparseable host", "http://[::1", true},

		// scheme
		{"empty string", "", true},
		{"ftp scheme", "ftp://example.com", true},
		{"javascript scheme", "javascript:alert(1)", true},
		{"data scheme", "data:text/html,hi", true},
		{"file scheme", "file:///etc/passwd", true},

		// host
		{"scheme only, no host", "http://", true},

		// embedded credentials
		{"user and password", "http://user:pass@example.com", true},
		{"user only", "https://user@example.com", true},
		{"userinfo phishing form", "http://www.trusted.com@evil.example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURL(tt.raw)
			if tt.wantErr && err == nil {
				t.Errorf("validateURL(%q) = nil, want error", tt.raw)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validateURL(%q) = %v, want nil", tt.raw, err)
			}
		})
	}
}

func TestExtractCode(t *testing.T) {
	tests := []struct {
		name     string
		shortURL string
		wantCode string
		wantOK   bool
	}{
		{"full short url", "http://host/GeAi9K", "GeAi9K", true},
		{"https with port", "https://host:8080/GeAi9K", "GeAi9K", true},
		{"trailing slash", "http://host/GeAi9K/", "GeAi9K", true},
		{"surrounding whitespace", "  http://host/GeAi9K  ", "GeAi9K", true},
		{"query and fragment ignored", "http://host/GeAi9K?a=1#x", "GeAi9K", true},
		{"bare code", "GeAi9K", "GeAi9K", true},

		{"root only", "http://host/", "", false},
		{"no path", "http://host", "", false},
		{"empty string", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, ok := extractCode(tt.shortURL)
			if ok != tt.wantOK {
				t.Errorf("extractCode(%q) ok = %v, want %v", tt.shortURL, ok, tt.wantOK)
			}
			if code != tt.wantCode {
				t.Errorf("extractCode(%q) code = %q, want %q", tt.shortURL, code, tt.wantCode)
			}
		})
	}
}
