package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRateLimiterRejectsAfterBurst(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := rateLimitTestRouter(NewRateLimiter(0.001, 2))

	for i := 0; i < 2; i++ {
		rec := performRateLimitedRequest(r, "192.0.2.1:1234")
		if rec.Code != http.StatusNoContent {
			t.Fatalf("request %d status = %d, want %d", i+1, rec.Code, http.StatusNoContent)
		}
	}

	rec := performRateLimitedRequest(r, "192.0.2.1:1234")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("limited status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
}

func TestRateLimiterUsesSeparateBucketsPerIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := rateLimitTestRouter(NewRateLimiter(0.001, 1))

	rec := performRateLimitedRequest(r, "192.0.2.1:1234")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("first IP first request status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	rec = performRateLimitedRequest(r, "192.0.2.1:1234")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("first IP second request status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}

	rec = performRateLimitedRequest(r, "192.0.2.2:1234")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("second IP first request status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func rateLimitTestRouter(rl *RateLimiter) *gin.Engine {
	r := gin.New()
	r.GET("/", rl.Middleware(), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	return r
}

func performRateLimitedRequest(h http.Handler, remoteAddr string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = remoteAddr
	h.ServeHTTP(rec, req)
	return rec
}
