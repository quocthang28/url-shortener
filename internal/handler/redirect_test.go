package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRedirect(t *testing.T) {
	const originalURL = "https://codesubmit.io/library/react"

	tests := []struct {
		name         string
		encodeURL    string
		path         string
		wantStatus   int
		wantLocation string
	}{
		{
			name:         "known short code redirects to original url",
			encodeURL:    originalURL,
			wantStatus:   http.StatusFound,
			wantLocation: originalURL,
		},
		{
			name:       "unknown short code returns not found",
			path:       "/unknown",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRouter()

			path := tt.path
			if tt.encodeURL != "" {
				shortURL := encodeShortURL(t, r, tt.encodeURL)
				code, ok := extractCode(shortURL)
				if !ok {
					t.Fatalf("extractCode(%q) ok = false, want true", shortURL)
				}
				path = "/" + code
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, path, nil)
			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}

			if tt.wantLocation == "" {
				return
			}

			if got := rec.Header().Get("Location"); got != tt.wantLocation {
				t.Fatalf("Location = %q, want %q", got, tt.wantLocation)
			}
		})
	}
}
