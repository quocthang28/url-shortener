package handler

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

const maxURLLength = 2048

var errInvalidURL = errors.New("invalid url")

// validateURL check if the user's URL is valid base on 3 rules: the scheme must be HTTP, the host must not be empty
// and no embed user credentials
func validateURL(raw string) error {
	raw = strings.TrimSpace(raw)

	if len(raw) > maxURLLength {
		return errors.New("url exceeds max allowed bytes")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("scheme must be http")
	}

	if u.Host == "" {
		return fmt.Errorf("host must not be empty")
	}

	if u.User != nil {
		return errors.New("embedded credentials are not allowed")
	}

	host := u.Hostname()

	if host == "" {
		return errors.New("hostname must not be empty")
	}

	if strings.ContainsAny(host, " \t\r\n") {
		return errors.New("hostname contains whitespace")
	}

	return nil
}

// extractCode pulls the short code out of a full short URL submitted to /decode
// (e.g. "http://host/GeAi9K" -> "GeAi9K"). Returns false if it can't.
func extractCode(shortURL string) (string, bool) {
	shortURL = strings.TrimSpace(shortURL)

	u, err := url.Parse(shortURL)
	if err != nil {
		return "", false
	}

	// Take the last non-empty path segment, so a trailing slash
	// ("http://host/GeAi9K/") still yields the code.
	segments := strings.Split(u.Path, "/")
	for i := len(segments) - 1; i >= 0; i-- {
		if segments[i] != "" {
			return segments[i], true
		}
	}

	return "", false
}
