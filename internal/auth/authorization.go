package auth

import (
	"errors"
	"net/http"
	"strings"
)

var ErrNoAuthHeaderIncluded = errors.New("No authorization header included")

func GetApiKeyToken(header http.Header) (string, error) {
	headerAuth := header.Get("Authorization")
	if headerAuth == "" || !strings.HasPrefix(headerAuth, "ApiKey ") {
		return "", ErrNoAuthHeaderIncluded
	}

	apiKey := strings.TrimPrefix(headerAuth, "ApiKey ")
	if len(strings.Fields(apiKey)) != 1 {
		return "", errors.New("Malformed authorization header")
	}

	return apiKey, nil
}
