package registry

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/eikendev/basechange/internal/handling"
)

const (
	maxRetries = 2
	userAgent  = "basechange"
)

var client http.Client

// https://codereview.stackexchange.com/q/173468
func shouldRetry(maxAttempts, attempts int, response *http.Response) (time.Duration, bool) {
	if attempts >= maxAttempts {
		return time.Duration(0), false
	}

	delay := time.Duration(attempts) * time.Second

	if response != nil && response.Header.Get("Retry-After") != "" {
		after, err := strconv.ParseInt(response.Header.Get("Retry-After"), 10, 64)
		if err != nil && after > 0 {
			delay = time.Duration(after) * time.Second
		}
	}

	return delay, true
}

func makeRequest(method, url string, header http.Header) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header = header
	return client.Do(req)
}

func handleResponse(resp *http.Response, expectedCode int) (*http.Response, error) {
	if resp == nil {
		return nil, errors.New("received nil response")
	}

	if resp.StatusCode == expectedCode {
		return resp, nil
	}

	if resp.Body != nil {
		handling.Close(resp.Body)
	}

	return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
}

func handleRetry(attempts, maxAttempts int, method string, resp *http.Response) error {
	delay, retry := shouldRetry(maxAttempts, attempts, resp)
	if !retry {
		return fmt.Errorf("request failed after %d attempts", maxAttempts)
	}

	time.Sleep(delay)
	log.Printf("%s attempt %d/%d", method, attempts+1, maxAttempts)
	return nil
}

// https://codereview.stackexchange.com/q/173468
func retryReq(method, url string, maxAttempts int, header http.Header, expectedCode int) (*http.Response, error) {
	log.Debugf("Requesting %s %s\n", method, url)

	if maxAttempts < 1 {
		return nil, errors.New("maxAttempts must be at least 1")
	}

	header.Set("User-Agent", userAgent)
	attempts := 0

	for attempts < maxAttempts {
		attempts++

		resp, err := makeRequest(method, url, header)
		if err == nil {
			resp, err = handleResponse(resp, expectedCode)
			if err == nil {
				return resp, nil
			}
		}

		if err := handleRetry(attempts, maxAttempts, method, resp); err != nil {
			return nil, err
		}
	}

	return nil, fmt.Errorf("exceeded maximum attempts (%d)", maxAttempts)
}

// Req retrieves the Req HTTP response for a given URL.
func Req(method, uri string, header http.Header) ([]byte, error) {
	resp, err := retryReq(method, uri, maxRetries+1, header, http.StatusOK)
	if err != nil {
		return nil, err
	}
	defer handling.Close(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
