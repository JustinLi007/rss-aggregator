package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetApiKeyToken(header http.Header) (string, error) {
	headerAuth := header.Get("Authorization")
	if !strings.HasPrefix(headerAuth, "ApiKey ") {
		return "", errors.New("Header does not contain Authorization: ApiKey")
	}

	apiKey := strings.TrimPrefix(headerAuth, "ApiKey ")
	if len(strings.Fields(apiKey)) != 1 {
		return "", errors.New("Malformed ApiKey")
	}

	return apiKey, nil
}
