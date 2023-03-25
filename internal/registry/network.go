package registry

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/eikendev/basechange/internal/handling"
	log "github.com/sirupsen/logrus"
)

const (
	maxRetries = 2
	UserAgent  = "basechange"
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

// https://codereview.stackexchange.com/q/173468
func RetryReq(method, url string, maxAttempts int, header http.Header, expectedCode int) (*http.Response, error) {
	log.Debugf("Requesting %s %s\n", method, url)

	header.Set("User-Agent", UserAgent)
	attempts := 0

	if maxAttempts < 1 {
		return nil, errors.New("maxAttempts must be at least 1")
	}

	for {
		attempts++

		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header = header

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == expectedCode {
			return resp, nil
		}
		defer handling.Close(resp.Body)

		delay, retry := shouldRetry(maxAttempts, attempts, resp)
		if !retry {
			if err == nil {
				err = errors.New("too many attempts")
			}
			return nil, err
		}

		if attempts < maxAttempts {
			time.Sleep(delay)
			log.Printf("%s attempt %d/%d", method, attempts+1, maxAttempts)
		}
	}
}

// Req retrieves the Req HTTP response for a given URL.
func Req(method, uri string, header http.Header) ([]byte, error) {
	resp, err := RetryReq(method, uri, maxRetries+1, header, http.StatusOK)
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
