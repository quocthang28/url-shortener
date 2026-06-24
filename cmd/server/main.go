package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"url-shortener/internal/handler"
	"url-shortener/internal/middleware"
	"url-shortener/internal/store"
)

func main() {
	cfg := loadConfig()

	st, err := store.NewSQLite(cfg.dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer st.Close()

	h := handler.New(st, cfg.baseURL)

	// gin.New() instead of gin.Default() so we choose our own middleware.
	// Gin reads GIN_MODE from the environment (set GIN_MODE=release in prod).
	r := gin.New()
	// Trust the Docker bridge network so Caddy's X-Forwarded-For is honored and
	// the rate limiter keys on the real client IP, not Caddy's container IP.
	if err := r.SetTrustedProxies([]string{"172.16.0.0/12"}); err != nil {
		log.Fatalf("trusted proxies: %v", err)
	}
	r.Use(gin.Logger(), gin.Recovery())

	writeLimiter := middleware.NewRateLimiter(2, 5)
	readLimiter := middleware.NewRateLimiter(20, 40)

	r.POST("/encode", writeLimiter.Middleware(), h.Encode)
	r.POST("/decode", readLimiter.Middleware(), h.Decode)
	r.GET("/:shortCode", readLimiter.Middleware(), h.Redirect)

	srv := &http.Server{
		Addr:         ":" + cfg.port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("listening on :%s (base url %s)", cfg.port, cfg.baseURL)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM — in-flight requests drain, store closes cleanly.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
	log.Println("stopped")
}

type config struct {
	port    string
	dbPath  string
	baseURL string
}

func loadConfig() config {
	return config{
		port:    env("PORT", "8080"),
		dbPath:  env("DB_PATH", "./urls.db"),
		baseURL: env("BASE_URL", "http://localhost:8080"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
