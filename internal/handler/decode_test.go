package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDecode(t *testing.T) {
	const originalURL = "https://codesubmit.io/library/react"

	tests := []struct {
		name       string
		body       string
		encodeURL  string
		wantStatus int
		wantURL    string
	}{
		{
			name:       "known short url returns original url",
			encodeURL:  originalURL,
			wantStatus: http.StatusOK,
			wantURL:    originalURL,
		},
		{
			name:       "unknown short code returns not found",
			body:       `{"short_url":"http://localhost:8080/unknown"}`,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "malformed short url returns bad request",
			body:       `{"short_url":"http://localhost:8080/"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty body returns bad request",
			body:       "",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRouter()

			body := tt.body
			if tt.encodeURL != "" {
				shortURL := encodeShortURL(t, r, tt.encodeURL)
				body = `{"short_url":"` + shortURL + `"}`
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/decode", bytes.NewBufferString(body))
			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}

			if tt.wantStatus != http.StatusOK {
				return
			}

			gotURL := decodeOriginalURL(t, rec)
			if gotURL != tt.wantURL {
				t.Fatalf("url = %q, want %q", gotURL, tt.wantURL)
			}
		})
	}
}

func encodeShortURL(t *testing.T, h http.Handler, originalURL string) string {
	t.Helper()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/encode", bytes.NewBufferString(`{"url":"`+originalURL+`"}`))
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("encode status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	return decodeShortURL(t, rec)
}

func decodeOriginalURL(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()

	var resp decodeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	return resp.URL
}
