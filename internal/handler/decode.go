package handler

import (
	"errors"
	"net/http"
	"url-shortener/internal/store"

	"github.com/gin-gonic/gin"
)

// Decode handles POST /decode.
func (h *Handler) Decode(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)

	var req decodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "")
		return
	}

	shortCode, ok := extractCode(req.ShortURL)
	if !ok {
		respondError(c, http.StatusBadRequest, "")
		return
	}

	originalUrl, err := h.store.FindByCode(shortCode)
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

	c.JSON(http.StatusOK, decodeResponse{URL: originalUrl})
}
