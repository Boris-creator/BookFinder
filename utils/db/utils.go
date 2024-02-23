package dbutils

import (
	utils "bookfinder/utils/common"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

type insertOptions[T any, M any] struct {
	driverName     string
	dataSourceName string
	prepare        func(value T) M
	replace        bool
}

type setInsertOption[T any, M any] func(*insertOptions[T, M]) error

type Config[T any, M any] struct{}

func (Config[T, M]) Replace(value bool) setInsertOption[T, M] {
	return func(options *insertOptions[T, M]) error {
		options.replace = value
		return nil
	}
}
func (Config[T, M]) Prepare(value func(value T) M) setInsertOption[T, M] {
	return func(options *insertOptions[T, M]) error {
		options.prepare = value
		return nil
	}
}

func BulkInsert[T any, M any](tableName string, unsavedRows []T, options ...setInsertOption[T, M]) error {
	insertOptions := insertOptions[T, M]{
		driverName:     "sqlite3",
		dataSourceName: "db/bookstore.db",
		replace:        true,
	}
	for _, setOption := range options {
		setOption(&insertOptions)
	}
	var prepare func(T) any = func(t T) any { return t }
	if insertOptions.prepare != nil {
		prepare = func(t T) any {
			return insertOptions.prepare(t)
		}
	}

	db, err := sql.Open(insertOptions.driverName, insertOptions.dataSourceName)
	if err != nil {
		return err
	}
	defer db.Close()

	queryBindings := make([]string, 0, len(unsavedRows))

	typed := reflect.ValueOf(prepare(unsavedRows[0]))
	columnNames := make([]string, 0, typed.NumField())
	utils.LoopFields(typed, func(i int) {
		columnNames = append(columnNames, typed.Type().Field(i).Tag.Get("db"))
	})

	valuesToBind := make([]interface{}, 0, len(unsavedRows)*len(columnNames))
	for _, row := range unsavedRows {
		queryBindings = append(queryBindings, fmt.Sprintf("(%s)", strings.Join(utils.Repeat("?", len(columnNames)), ",")))
		rowValues := reflect.ValueOf(prepare(row))
		values := make([]any, 0, len(columnNames))
		utils.LoopFields(rowValues, func(i int) {
			values = append(values, rowValues.Field(i).Interface())
		})
		valuesToBind = append(valuesToBind, values...)
	}
	statement := "INSERT"
	if insertOptions.replace {
		statement = "REPLACE"
	}
	stmt := fmt.Sprintf("%s INTO %s (%s) VALUES %s",
		statement,
		tableName,
		strings.Join(columnNames, ","),
		strings.Join(queryBindings, ","))
	_, err = db.Exec(stmt, valuesToBind...)

	return err
}
