package httputils

import (
	"bookfinder/utils/stability"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type RequestError struct {
	StatusCode int
}

func (r *RequestError) Error() string {
	return fmt.Sprintf("%d response status code", r.StatusCode)
}

var maxResponseTimeoutSeconds = 15

func Fetch(fetcher func() (*http.Response, error)) func() (io.ReadCloser, error) {
	wrapper := func() (io.ReadCloser, error) {
		response, err := fetcher()
		if err != nil {
			return nil, err
		}
		if response.StatusCode >= 400 && response.StatusCode < 600 {
			return response.Body, fmt.Errorf("fetch error: %w", &RequestError{response.StatusCode})
		}

		return response.Body, nil
	}

	return wrapper
}

func FetchWithTimeout(fetcher func() (*http.Response, error)) func() (io.ReadCloser, error) {
	wrapper := Fetch(fetcher)
	return stability.Timeout[io.ReadCloser](wrapper, time.Duration(maxResponseTimeoutSeconds)*time.Second)
}
func FetchWithRetry(fetcher func() (*http.Response, error)) func() (io.ReadCloser, error) {
	wrapper := Fetch(fetcher)
	return stability.Retry(wrapper, 3, stability.SetDelay(time.Microsecond), stability.CheckIfRetryNeeded(func(err error, args ...any) bool {
		urlErr := new(url.Error)
		if errors.As(err, &urlErr) {
			return true
		}
		statusErr := new(RequestError)
		if errors.As(err, &statusErr) {
			return errors.Unwrap(err).(*RequestError).StatusCode >= 500
		}
		return false
	}))
}
