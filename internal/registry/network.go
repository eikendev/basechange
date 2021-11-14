package registry

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	maxRetries = 2
	UserAgent  = "Watchtower (Docker)"
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

		response, err := client.Do(req)
		if err == nil && response.StatusCode == expectedCode {
			return response, nil
		}

		delay, retry := shouldRetry(maxAttempts, attempts, response)
		if !retry {
			return nil, err
		}

		defer response.Body.Close()

		if attempts < maxAttempts {
			time.Sleep(delay)
			log.Printf("%s attempt %d/%d", method, attempts+1, maxAttempts)
		}
	}
}

// Req retrieves the Req HTTP response for a given URL.
func Req(method, uri string, header http.Header) []byte {
	resp, err := RetryReq(method, uri, maxRetries+1, header, http.StatusOK)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return body
}
