package stability

import (
	commonutils "bookfinder/utils/common"
	"context"
	"fmt"
	"math"
	"reflect"
	"time"
)

type retryOptions struct {
	delay         time.Duration
	isRetryNeeded func(error, ...any) bool
}
type SetRetryOption func(*retryOptions) retryOptions

func SetDelay(delay time.Duration) SetRetryOption {
	return func(ro *retryOptions) retryOptions {
		ro.delay = delay
		return *ro
	}
}
func CheckIfRetryNeeded(clb func(error, ...any) bool) SetRetryOption {
	return func(ro *retryOptions) retryOptions {
		ro.isRetryNeeded = clb
		return *ro
	}
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

func Retry[T any](function T, retries int, options ...SetRetryOption) T {
	var opts = retryOptions{
		delay: time.Second,
		isRetryNeeded: func(err error, a ...any) bool {
			return err != nil
		},
	}
	for _, setOption := range options {
		setOption(&opts)
	}
	var delayIncrement = 1.5

	v := reflect.MakeFunc(reflect.TypeOf(function), func(in []reflect.Value) []reflect.Value {
		f := reflect.ValueOf(function)
		if f.Type().NumOut() == 0 {
			return f.Call(in)
		}
		errorIndex := f.Type().NumOut() - 1

		for r := 1; ; r++ {
			returnValues := f.Call(in)
			if r >= retries {
				return returnValues
			}
			err := returnValues[errorIndex]
			if e, ok := err.Interface().(error); ok && opts.isRetryNeeded(
				e, commonutils.Map(returnValues, func(value reflect.Value) any {
					return value.Interface()
				})...) {
				<-time.After(opts.delay * time.Duration(math.Pow(delayIncrement, float64(r-1))))
			} else {
				return returnValues
			}
		}
	})
	return v.Interface().(T)
}
