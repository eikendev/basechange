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
	defer handling.Close(resp.Body)

	challenge := strings.ToLower(resp.Header.Get(ChallengeHeader))

	if strings.HasPrefix(challenge, "bearer") {
		return getBearerHeader(challenge, image)
	}

	return "", errors.New("unsupported challenge type from registry")
}

func getBearerHeader(challenge string, image string) (string, error) {
	if strings.Contains(image, ":") {
		image = strings.Split(image, ":")[0]
	}

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

func getAuthURL(challenge string, image string) (*url.URL, error) {
	loweredChallenge := strings.ToLower(challenge)
	raw := strings.TrimPrefix(loweredChallenge, "bearer")

	pairs := strings.Split(raw, ",")
	values := make(map[string]string, len(pairs))

	for _, pair := range pairs {
		trimmed := strings.Trim(pair, " ")
		kv := strings.Split(trimmed, "=")
		key := kv[0]
		val := strings.Trim(kv[1], "\"")
		values[key] = val
	}

	if values["realm"] == "" || values["service"] == "" {
		return nil, fmt.Errorf("malformed challenge header")
	}

	authURL, _ := url.Parse(values["realm"])
	q := authURL.Query()
	q.Add("service", values["service"])

	scopeImage := GetScopeFromImageName(image, values["service"])
	scope := fmt.Sprintf("repository:%s:pull", scopeImage)
	q.Add("scope", scope)

	authURL.RawQuery = q.Encode()

	return authURL, nil
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
