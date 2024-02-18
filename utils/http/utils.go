package httputils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

var maxResponseTimeoutSeconds = 15

func Fetch(fetcher func() (*http.Response, error)) func() (io.ReadCloser, error) {
	wrapper := func() (io.ReadCloser, error) {
		response, err := fetcher()
		if err != nil {
			return nil, err
		}
		if response.StatusCode >= 400 && response.StatusCode < 600 {
			return response.Body, fmt.Errorf("fetch error: %d response status code", response.StatusCode)
		}

		return response.Body, nil
	}

	return wrapper
}

func FetchWithTimeout(fetcher func() (*http.Response, error)) func() (io.ReadCloser, error) {
	wrapper := Fetch(fetcher)
	ctx := context.Background()
	timeout, cancel := context.WithTimeout(ctx, time.Duration(maxResponseTimeoutSeconds)*time.Second)
	return func() (io.ReadCloser, error) {
		chRes := make(chan io.ReadCloser)
		chErr := make(chan error)
		defer cancel()

		go func() {
			res, err := wrapper()
			chRes <- res
			chErr <- err
		}()

		select {
		case res := <-chRes:
			return res, <-chErr
		case <-timeout.Done():
			return nil, fmt.Errorf("fetch error: %s", timeout.Err().Error())
		}
	}
}
