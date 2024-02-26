package commonutils

import (
	"errors"
	"reflect"
)

func Repeat[T any](v T, times int) []T {
	list := make([]T, 0, times)
	for i := 0; i < times; i++ {
		list = append(list, v)
	}
	return list
}
func Map[T, U any](ts []T, f func(T) U) []U {
	us := make([]U, len(ts))
	for i := range ts {
		us[i] = f(ts[i])
	}
	return us
}

func LoopFields(typed reflect.Value, clb func(fieldIndex int)) error {
	if typed.Kind() != reflect.Struct {
		return errors.New("")
	}
	for i := 0; i < typed.NumField(); i++ {
		if !typed.Type().Field(i).IsExported() {
			continue
		}
		clb(i)
	}
	return nil
}
