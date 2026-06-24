package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEncode(t *testing.T) {
	overLengthURL := "https://example.com/" + strings.Repeat("a", maxURLLength)

	tests := []struct {
		name       string
		body       string
		wantStatus int
		repeat     bool
	}{
		{
			name:       "valid url returns short url",
			body:       `{"url":"https://codesubmit.io/library/react"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "same url returns same short url",
			body:       `{"url":"https://codesubmit.io/library/react"}`,
			wantStatus: http.StatusOK,
			repeat:     true,
		},
		{
			name:       "missing scheme returns bad request",
			body:       `{"url":"codesubmit.io/library/react"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty body returns bad request",
			body:       "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "over length url returns bad request",
			body:       `{"url":"` + overLengthURL + `"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRouter()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/encode", bytes.NewBufferString(tt.body))
			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}

			if tt.wantStatus != http.StatusOK {
				return
			}

			firstShortURL := decodeShortURL(t, rec)
			assertShortURL(t, firstShortURL)

			if !tt.repeat {
				return
			}

			rec = httptest.NewRecorder()
			req = httptest.NewRequest(http.MethodPost, "/encode", bytes.NewBufferString(tt.body))
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("repeat status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
			}

			secondShortURL := decodeShortURL(t, rec)
			if secondShortURL != firstShortURL {
				t.Fatalf("repeat short_url = %q, want %q", secondShortURL, firstShortURL)
			}
		})
	}
}

func decodeShortURL(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()

	var resp encodeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	return resp.ShortURL
}

func assertShortURL(t *testing.T, shortURL string) {
	t.Helper()

	const prefix = "http://localhost:8080/"
	if !strings.HasPrefix(shortURL, prefix) {
		t.Fatalf("short_url = %q, want prefix %q", shortURL, prefix)
	}

	code := strings.TrimPrefix(shortURL, prefix)
	if len(code) != defaultCodeLength {
		t.Fatalf("short code length = %d, want %d", len(code), defaultCodeLength)
	}
}
