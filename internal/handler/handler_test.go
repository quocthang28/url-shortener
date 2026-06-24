package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"url-shortener/internal/store"
)

// newTestRouter wires the handlers onto a Gin engine backed by the in-memory
// store — no disk I/O, fast and deterministic.
func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := New(store.NewMemory(), "http://localhost:8080")
	r := gin.New()
	r.POST("/encode", h.Encode)
	r.POST("/decode", h.Decode)
	r.GET("/:shortCode", h.Redirect)
	return r
}

// TestEncodeDecodeRoundTrip encodes a URL then decodes the result back.
func TestEncodeDecodeRoundTrip(t *testing.T) {
	const originalURL = "https://codesubmit.io/library/react"

	r := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/encode",
		bytes.NewBufferString(`{"url":"`+originalURL+`"}`))
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("encode status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	shortURL := decodeShortURL(t, rec)

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/decode",
		bytes.NewBufferString(`{"short_url":"`+shortURL+`"}`))
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("decode status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	gotURL := decodeOriginalURL(t, rec)
	if gotURL != originalURL {
		t.Fatalf("url = %q, want %q", gotURL, originalURL)
	}
}
