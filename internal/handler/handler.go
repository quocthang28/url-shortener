package handler

import (
	"github.com/gin-gonic/gin"

	"url-shortener/internal/store"
)

const (
	defaultCodeLength = 6
	maxEncodeRetries  = 5

	// maxBodyBytes caps the request body before decoding
	maxBodyBytes = 4 << 10
)

// Handler holds the dependencies shared by every HTTP handler.
type Handler struct {
	store      store.Store
	baseURL    string
	codeLength int
	maxRetries int
}

// New builds a Handler. baseURL is prepended to short codes in /encode responses.
func New(s store.Store, baseURL string) *Handler {
	return &Handler{
		store:      s,
		baseURL:    baseURL,
		codeLength: defaultCodeLength,
		maxRetries: maxEncodeRetries,
	}
}

type encodeRequest struct {
	URL string `json:"url"`
}

type encodeResponse struct {
	ShortURL string `json:"short_url"`
}

type decodeRequest struct {
	ShortURL string `json:"short_url"`
}

type decodeResponse struct {
	URL string `json:"url"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// respondError writes a JSON error body with the given status.
func respondError(c *gin.Context, status int, msg string) {
	c.JSON(status, errorResponse{Error: msg})
}
