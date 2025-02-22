// Extracted from https://containrrr.dev/watchtower/

// Package registry provides functions to fetch the latest digest of an image.
package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/eikendev/basechange/internal/handling"
)

// ChallengeHeader is the HTTP Header containing challenge instructions
const ChallengeHeader = "WWW-Authenticate"

type tokenResponse struct {
	Token string
}

// GetToken fetches a token for the registry hosting the provided image
func GetToken(image string) (string, error) {
	var challengeURL url.URL
	var err error

	if challengeURL, err = getChallengeURL(image); err != nil {
		return "", err
	}

	header := http.Header{"Accept": []string{"*/*"}}
	resp, err := retryReq("GET", challengeURL.String(), maxRetries+1, header, http.StatusUnauthorized)
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", errors.New("received nil response")
	}
	defer handling.Close(resp.Body)

	challenge := strings.ToLower(resp.Header.Get(ChallengeHeader))
	if challenge == "" {
		return "", errors.New("empty challenge header")
	}

	if strings.HasPrefix(challenge, "bearer") {
		return getBearerHeader(challenge, image)
	}

	return "", errors.New("unsupported challenge type from registry")
}

func getBearerHeader(challenge string, image string) (string, error) {
	if image == "" {
		return "", errors.New("empty image name")
	}
	parts := strings.Split(image, ":")
	if len(parts) == 0 {
		return "", errors.New("invalid image name format")
	}
	image = parts[0]

	authURL, err := getAuthURL(challenge, image)
	if err != nil {
		return "", err
	}

	resp, err := Req("GET", authURL.String(), http.Header{})
	if err != nil {
		return "", err
	}

	tokenResponse := &tokenResponse{}

	err = json.Unmarshal(resp, tokenResponse)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Bearer %s", tokenResponse.Token), nil
}

func parseChallengeKeyValue(pair string) (string, string, error) {
	trimmed := strings.Trim(pair, " ")
	if trimmed == "" {
		return "", "", errors.New("empty pair")
	}

	kv := strings.Split(trimmed, "=")
	if len(kv) != 2 {
		return "", "", errors.New("invalid key-value format")
	}

	key := strings.TrimSpace(kv[0])
	if key == "" {
		return "", "", errors.New("empty key")
	}

	val := strings.Trim(strings.TrimSpace(kv[1]), "\"")
	return key, val, nil
}

func extractChallengeValues(raw string) (map[string]string, error) {
	pairs := strings.Split(raw, ",")
	if len(pairs) == 0 {
		return nil, errors.New("invalid challenge format")
	}

	values := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		key, val, err := parseChallengeKeyValue(pair)
		if err != nil {
			continue
		}
		values[key] = val
	}

	if values["realm"] == "" || values["service"] == "" {
		return nil, fmt.Errorf("malformed challenge header")
	}

	return values, nil
}

func buildAuthURL(values map[string]string, image string) (*url.URL, error) {
	authURL, err := url.Parse(values["realm"])
	if err != nil {
		return nil, fmt.Errorf("failed to parse realm URL: %w", err)
	}
	if authURL == nil {
		return nil, errors.New("nil URL after parsing")
	}

	q := authURL.Query()
	q.Add("service", values["service"])

	scopeImage := GetScopeFromImageName(image, values["service"])
	scope := fmt.Sprintf("repository:%s:pull", scopeImage)
	q.Add("scope", scope)

	authURL.RawQuery = q.Encode()
	return authURL, nil
}

func getAuthURL(challenge string, image string) (*url.URL, error) {
	loweredChallenge := strings.ToLower(challenge)
	raw := strings.TrimPrefix(loweredChallenge, "bearer")
	if raw == "" {
		return nil, errors.New("empty bearer challenge")
	}

	values, err := extractChallengeValues(raw)
	if err != nil {
		return nil, err
	}

	return buildAuthURL(values, image)
}

func getChallengeURL(image string) (url.URL, error) {
	host, err := getHost(image)
	if err != nil {
		return url.URL{}, err
	}

	URL := url.URL{
		Scheme: "https",
		Host:   host,
		Path:   "/v2/",
	}

	return URL, nil
}
