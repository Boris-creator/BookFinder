package httputils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
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
	return Timeout[io.ReadCloser](wrapper, time.Duration(maxResponseTimeoutSeconds)*time.Second)
}

func Timeout[R any, T any](effector T, delay time.Duration) T {
	ctx := context.Background()
	timeout, cancel := context.WithTimeout(ctx, delay)

	f := reflect.ValueOf(effector)
	if f.Type().NumOut() != 2 {
		panic("function must return value and error")
	}
	isErr := fmt.Sprint(f.Type().Out(1)) == "error"
	if !isErr {
		panic("function must return an error")
	}

	v := reflect.MakeFunc(reflect.TypeOf(effector), func(in []reflect.Value) []reflect.Value {
		chRes := make(chan R)
		chErr := make(chan error)
		defer cancel()

		results := make([]reflect.Value, 0, f.Type().NumOut())

		go func() {
			results = f.Call(in)

			err, ok := results[1].Interface().(error)
			if !ok {
				err = nil
			}
			chRes <- results[0].Interface().(R)
			chErr <- err
		}()

		select {
		case <-chRes:
			return results
		case <-timeout.Done():
			results = append(
				results, reflect.Zero(f.Type().Out(0)), reflect.ValueOf(fmt.Errorf("fetch error: %s", timeout.Err().Error())),
			)
			return results
		}
	})
	return v.Interface().(T)
}

func Retry[T any](function T, retries int, delay time.Duration) T {
	v := reflect.MakeFunc(reflect.TypeOf(function), func(in []reflect.Value) []reflect.Value {
		f := reflect.ValueOf(function)
		if f.Type().NumOut() == 0 {
			return f.Call(in)
		}
		errorIndex := f.Type().NumOut() - 1

		for r := 1; ; r++ {
			returnValues := f.Call(in)
			if r > retries {
				return returnValues
			}
			err := returnValues[errorIndex]
			if e, ok := err.Interface().(error); ok && e != nil {
				<-time.After(delay)
			} else {
				return returnValues
			}
		}
	})
	return v.Interface().(T)
}
