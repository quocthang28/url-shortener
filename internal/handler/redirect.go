package handler

import (
	"errors"
	"net/http"
	"url-shortener/internal/store"

	"github.com/gin-gonic/gin"
)

// Redirect handles GET /:shortCode.
func (h *Handler) Redirect(c *gin.Context) {
	code := c.Param("shortCode")

	originalURL, err := h.store.FindByCode(code)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			respondError(c, http.StatusNotFound, "url doesn't exist for this code")
			return
		default:
			respondError(c, http.StatusInternalServerError, "")
			return
		}
	}

	c.Redirect(http.StatusFound, originalURL)
}
