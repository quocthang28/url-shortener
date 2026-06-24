package handler

import (
	"errors"
	"net/http"
	"url-shortener/internal/shortener"
	"url-shortener/internal/store"

	"github.com/gin-gonic/gin"
)

// Encode handles POST /encode.
func (h *Handler) Encode(c *gin.Context) {
	// request body length should be below max allowed length
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)

	var req encodeRequest

	err := c.ShouldBindJSON(&req)
	if err != nil {
		respondError(c, http.StatusBadRequest, "")
		return
	}

	if err := validateURL(req.URL); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	// check if the URL already has a shortcode, if so return the shortcode instead of create a new one
	code, err := h.store.FindByURL(req.URL)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		respondError(c, http.StatusInternalServerError, "")
		return
	}

	if code != "" {
		c.JSON(http.StatusOK, encodeResponse{ShortURL: h.baseURL + "/" + code})
		return
	}

	for retries := 0; retries < h.maxRetries; retries++ {
		code = shortener.Generate(h.codeLength)

		err := h.store.Save(code, req.URL)
		switch {
		case err == nil:
			c.JSON(http.StatusOK, encodeResponse{ShortURL: h.baseURL + "/" + code})
			return

		case errors.Is(err, store.ErrCodeTaken):
			continue // retries to get another code

		case errors.Is(err, store.ErrURLExists): // idempotency, reuse existing code of that url
			savedCode, err := h.store.FindByURL(req.URL)
			if err != nil {
				respondError(c, http.StatusInternalServerError, "")
				return
			}

			c.JSON(http.StatusOK, encodeResponse{ShortURL: h.baseURL + "/" + savedCode})
			return

		default:
			respondError(c, http.StatusInternalServerError, "")
			return
		}
	}

	// retries exceed max retries, return error
	respondError(c, http.StatusServiceUnavailable, "")
}
